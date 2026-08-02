// Package ipam implements IPAM availability custom @action endpoints:
// available IPs, prefixes, and ASNs from their parent objects.
//
// These ports Python NetBox's netaddr.IPSet-based logic to Go stdlib netip.
// Containment math (subtracting used IPs/prefixes/ranges) is done in Go so the
// same code works against both PostgreSQL (production) and SQLite (tests).
//
// Endpoints:
//
//	GET/POST /api/ipam/prefixes/:id/available-ips/
//	GET/POST /api/ipam/ip-ranges/:id/available-ips/
//	GET/POST /api/ipam/prefixes/:id/available-prefixes/
//	GET/POST /api/ipam/asn-ranges/:id/available-asns/
package ipam

import (
	"fmt"
	"net/netip"
	"sort"
)

// ---- IP set helpers ----

// addrSet is an ordered, deduplicated set of netip.Addr, used as the "used"
// side of an availability computation. Contains lookups are binary-search.
type addrSet struct {
	addrs []netip.Addr
}

func newAddrSet(addrs []netip.Addr) *addrSet {
	dedup := make([]netip.Addr, 0, len(addrs))
	seen := make(map[string]bool, len(addrs))
	for _, a := range addrs {
		key := a.String()
		if !seen[key] {
			seen[key] = true
			dedup = append(dedup, a)
		}
	}
	sort.Slice(dedup, func(i, j int) bool {
		return dedup[i].Less(dedup[j])
	})
	return &addrSet{addrs: dedup}
}

// has reports whether the set contains a.
func (s *addrSet) has(a netip.Addr) bool {
	idx := sort.Search(len(s.addrs), func(i int) bool {
		return !s.addrs[i].Less(a)
	})
	return idx < len(s.addrs) && s.addrs[idx] == a
}

// prefixRange describes an inclusive [start, end] interval of consumed
// addresses, derived from an IPRange (mark_populated).
type prefixRange struct {
	start, end netip.Addr
}

// contains reports whether a falls within [r.start, r.end].
func (r prefixRange) contains(a netip.Addr) bool {
	return !a.Less(r.start) && !r.end.Less(a)
}

// computeFreeAddrs computes the addresses available within parent that are not
// in used and not in any consumed range.
//
// skipReserved drops the reserved addresses (network/broadcast for IPv4, the
// subnet-router anycast for IPv6) — matching NetBox's get_available_ips()
// behavior for non-pool prefixes. The caller decides whether to skip based on
// prefix length and is_pool.
func computeFreeAddrs(parent netip.Prefix, used *addrSet, ranges []prefixRange, skipReserved bool) []netip.Addr {
	var out []netip.Addr
	// parent.Masked() ensures we iterate over the network, not from `parent`'s
	// address onward.
	bounds := parent.Masked()
	addr := bounds.Addr()
	for {
		if !addr.IsValid() {
			break
		}
		if !bounds.Contains(addr) {
			break
		}
		if !used.has(addr) && !inAnyRange(ranges, addr) {
			if !skipReserved || !isReserved(bounds, addr) {
				out = append(out, addr)
			}
		}
		next := addr.Next()
		if next == addr || !next.IsValid() {
			break
		}
		addr = next
	}
	return out
}

func inAnyRange(ranges []prefixRange, a netip.Addr) bool {
	for _, r := range ranges {
		if r.contains(a) {
			return true
		}
	}
	return false
}

// isReserved reports whether addr is the network/broadcast of prefix (for IPv4)
// or the subnet-router anycast address — the first address of an IPv6 prefix
// (RFC 4291). NetBox reserves these for "normal" (non-pool, < /31) prefixes.
func isReserved(prefix netip.Prefix, addr netip.Addr) bool {
	masked := prefix.Masked()
	if addr == masked.Addr() {
		// First address: broadcast-equivalent (IPv4) and IPv6 subnet-router anycast.
		return true
	}
	// Last address is only reserved for IPv4.
	if addr.Is4() {
		if bcast, ok := broadcast(masked); ok && addr == bcast {
			return true
		}
	}
	return false
}

// broadcast returns the last address of an IPv4 prefix.
func broadcast(prefix netip.Prefix) (netip.Addr, bool) {
	if !prefix.Addr().Is4() {
		return netip.Addr{}, false
	}
	addr := prefix.Masked().Addr()
	bits := prefix.Bits()
	addr4 := addr.As4()
	// total host bits = 32 - bits; set the low (32 - bits) of the v4 word.
	hostBits := 32 - bits
	for i := 0; i < hostBits; i++ {
		byteIdx := 3 - i/8
		bitIdx := i % 8
		addr4[byteIdx] |= 1 << bitIdx
	}
	return netip.AddrFrom4(addr4), true
}

// ---- Available prefixes: greedy CIDR splitting ----

// freePrefixes returns the list of CIDR prefixes contained within parent that
// are not covered by any of the used child prefixes. The result is sorted and
// minimal (no two returned CIDRs are adjacent and mergeable).
//
// Algorithm: walk the parent's address space; at each candidate prefix, if it
// overlaps a used prefix, recurse into halves; if it is entirely free, emit it.
func freePrefixes(parent netip.Prefix, used []netip.Prefix) []netip.Prefix {
	// Sort used prefixes for deterministic overlap checks.
	sort.Slice(used, func(i, j int) bool {
		ai, aj := used[i].Masked().Addr(), used[j].Masked().Addr()
		if ai == aj {
			return used[i].Bits() < used[j].Bits()
		}
		return ai.Less(aj)
	})

	parent = parent.Masked()
	var result []netip.Prefix
	emit(parent, used, &result)
	sort.Slice(result, func(i, j int) bool {
		ai, aj := result[i].Masked().Addr(), result[j].Masked().Addr()
		if ai == aj {
			return result[i].Bits() < result[j].Bits()
		}
		return ai.Less(aj)
	})
	return result
}

// emit recursively descends: if `cur` is free (no overlap with used), append it;
// otherwise split into two halves and recurse (unless cur is already a single
// address, which can't be split).
func emit(cur netip.Prefix, used []netip.Prefix, out *[]netip.Prefix) {
	cur = cur.Masked()
	overlaps := false
	fullyCovered := false
	for _, u := range used {
		um := u.Masked()
		if overlapsPrefix(cur, um) {
			overlaps = true
			if containsPrefix(um, cur) {
				fullyCovered = true
				break
			}
		}
	}
	if fullyCovered {
		return
	}
	if !overlaps {
		*out = append(*out, cur)
		return
	}
	// Can we split? Only if cur has more than a single address.
	if cur.IsSingleIP() {
		return
	}
	half, err := cur.Addr().Prefix(cur.Bits() + 1)
	if err != nil {
		return
	}
	emit(half, used, out)
	emit(secondHalf(half), used, out)
}

// secondHalf returns the second half of a prefix (e.g. given 10.0.0.0/25
// — itself the first half of a /24 — returns 10.0.0.128/25). It sets the
// "distinguishing bit" of the parent split: when a /N prefix is split into two
// /(N+1) halves, the second half has its (N)th-from-LSB bit set (counting
// from 0). For p=10.0.0.0/25 (bits=25), that bit is bit 25 of the v4 word
// (value 2^7 = 128), yielding 10.0.0.128/25.
func secondHalf(p netip.Prefix) netip.Prefix {
	addr := p.Addr()
	bits := p.Bits()
	if addr.Is4() {
		v4 := addr.As4()
		// Distinguishing bit for a /(bits) prefix within its parent /(bits-1).
		// Bit index from LSB = (32 - bits). For bits=25 → bit 7 (value 128).
		idx := 32 - bits
		byteIdx := 3 - idx/8
		bitIdx := idx % 8
		v4[byteIdx] |= 1 << bitIdx
		newAddr := netip.AddrFrom4(v4)
		pfx, err := netip.ParsePrefix(newAddr.String() + "/" + itoa(bits))
		if err != nil {
			return p
		}
		return pfx.Masked()
	}
	// IPv6
	a16 := addr.As16()
	idx := 128 - bits
	byteIdx := 15 - idx/8
	bitIdx := idx % 8
	a16[byteIdx] |= 1 << bitIdx
	newAddr := netip.AddrFrom16(a16).Unmap()
	pfx, err := netip.ParsePrefix(newAddr.String() + "/" + itoa(bits))
	if err != nil {
		return p
	}
	return pfx.Masked()
}

// itoa converts a non-negative int to its decimal string. Avoids importing
// strconv in this file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// overlapsPrefix reports whether two prefixes share any address.
func overlapsPrefix(a, b netip.Prefix) bool {
	a, b = a.Masked(), b.Masked()
	if a.Addr().Is4() != b.Addr().Is4() {
		return false
	}
	maxBits := a.Bits()
	if b.Bits() > maxBits {
		maxBits = b.Bits()
	}
	ta, err := a.Addr().Prefix(maxBits)
	if err != nil {
		return false
	}
	tb, err := b.Addr().Prefix(maxBits)
	if err != nil {
		return false
	}
	return ta.Masked() == tb.Masked()
}

// containsPrefix reports whether outer fully contains inner (every address of
// inner is in outer).
func containsPrefix(outer, inner netip.Prefix) bool {
	outer, inner = outer.Masked(), inner.Masked()
	if outer.Addr().Is4() != inner.Addr().Is4() {
		return false
	}
	if outer.Bits() > inner.Bits() {
		return false
	}
	return outer.Contains(inner.Addr()) && outer.Contains(lastAddr(inner))
}

// lastAddr returns the highest address in a prefix.
func lastAddr(p netip.Prefix) netip.Addr {
	if p.Addr().Is4() {
		if b, ok := broadcast(p); ok {
			return b
		}
	}
	// IPv6: set all host bits to 1.
	a16 := p.Addr().As16()
	hostBits := 128 - p.Bits()
	for i := 0; i < hostBits; i++ {
		byteIdx := 15 - i/8
		bitIdx := i % 8
		a16[byteIdx] |= 1 << bitIdx
	}
	return netip.AddrFrom16(a16).Unmap()
}

// firstFitPrefix finds the first available prefix of the requested length
// within the free space, or returns ok=false if none fits.
func firstFitPrefix(free []netip.Prefix, wantBits int) (netip.Prefix, bool) {
	for _, f := range free {
		if f.Bits() <= wantBits {
			carved, err := f.Addr().Prefix(wantBits)
			if err == nil {
				return carved.Masked(), true
			}
		}
	}
	return netip.Prefix{}, false
}

// ---- Parsing helpers ----

// parsePrefixAddr splits an IPAddress inet string ("10.0.0.1/24") into its
// host address and the mask length. Returns ok=false on parse failure.
func parsePrefixAddr(s string) (netip.Addr, int, bool) {
	pfx, err := netip.ParsePrefix(s)
	if err == nil {
		return pfx.Addr(), pfx.Bits(), true
	}
	// Some inet values are bare addresses with no mask.
	a, err2 := netip.ParseAddr(s)
	if err2 != nil {
		return netip.Addr{}, 0, false
	}
	if a.Is4() {
		return a, 32, true
	}
	return a, 128, true
}

// hostAddr parses an IPAddress inet string and returns just the host address.
func hostAddr(s string) (netip.Addr, bool) {
	a, _, ok := parsePrefixAddr(s)
	return a, ok
}

// ParsePrefix parses a CIDR string into a netip.Prefix (returns ok=false on error).
func ParsePrefix(s string) (netip.Prefix, bool) {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return netip.Prefix{}, false
	}
	return p.Masked(), true
}

// rangeBounds parses start/end inet strings (e.g. "10.0.0.1/24") into an
// inclusive address range. Returns ok=false if either endpoint is invalid.
func rangeBounds(start, end string) (netip.Addr, netip.Addr, bool) {
	s, ok1 := hostAddr(start)
	e, ok2 := hostAddr(end)
	if !ok1 || !ok2 {
		return netip.Addr{}, netip.Addr{}, false
	}
	if e.Less(s) {
		return netip.Addr{}, netip.Addr{}, false
	}
	return s, e, true
}

// errStr wraps a formatted string as an error.
func errStr(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
