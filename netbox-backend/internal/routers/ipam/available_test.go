package ipam

import (
	"net/netip"
	"reflect"
	"testing"
)

// addrsToStrings is a test helper.
func addrsToStrings(addrs []netip.Addr) []string {
	out := make([]string, len(addrs))
	for i, a := range addrs {
		out[i] = a.String()
	}
	return out
}

func prefixesToStrings(pfx []netip.Prefix) []string {
	out := make([]string, len(pfx))
	for i, p := range pfx {
		out[i] = p.String()
	}
	return out
}

// ---- addrSet ----

func TestAddrSet_Has(t *testing.T) {
	addrs := []netip.Addr{mustAddr("10.0.0.1"), mustAddr("10.0.0.3"), mustAddr("10.0.0.2")}
	s := newAddrSet(addrs)

	cases := map[string]bool{
		"10.0.0.1": true,
		"10.0.0.2": true,
		"10.0.0.3": true,
		"10.0.0.4": false,
		"9.9.9.9":  false,
	}
	for a, want := range cases {
		if got := s.has(mustAddr(a)); got != want {
			t.Errorf("has(%s) = %v, want %v", a, got, want)
		}
	}
}

func TestAddrSet_DeduplicatesAndSorts(t *testing.T) {
	s := newAddrSet([]netip.Addr{
		mustAddr("10.0.0.3"), mustAddr("10.0.0.1"),
		mustAddr("10.0.0.1"), mustAddr("10.0.0.2"),
	})
	want := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	got := addrsToStrings(s.addrs)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---- computeFreeAddrs ----

func TestComputeFreeAddrs_FullRangeUnused(t *testing.T) {
	// /30 has 4 addresses; with skipReserved we drop network (0) + broadcast (3).
	p := mustPrefix("10.0.0.0/30")
	used := newAddrSet(nil)
	free := computeFreeAddrs(p, used, nil, true)
	got := addrsToStrings(free)
	want := []string{"10.0.0.1", "10.0.0.2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeFreeAddrs_SkipUsedIPs(t *testing.T) {
	p := mustPrefix("10.0.0.0/30")
	used := newAddrSet([]netip.Addr{mustAddr("10.0.0.1")})
	free := computeFreeAddrs(p, used, nil, true)
	got := addrsToStrings(free)
	want := []string{"10.0.0.2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeFreeAddrs_PoolKeepsReserved(t *testing.T) {
	// is_pool = true → skipReserved = false → all 4 addresses returned.
	p := mustPrefix("10.0.0.0/30")
	used := newAddrSet(nil)
	free := computeFreeAddrs(p, used, nil, false)
	got := addrsToStrings(free)
	want := []string{"10.0.0.0", "10.0.0.1", "10.0.0.2", "10.0.0.3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeFreeAddrs_RangeConsumesAddresses(t *testing.T) {
	// A populated IPRange covering .1-.2 removes those addresses.
	p := mustPrefix("10.0.0.0/29")
	used := newAddrSet(nil)
	ranges := []prefixRange{{start: mustAddr("10.0.0.2"), end: mustAddr("10.0.0.4")}}
	free := computeFreeAddrs(p, used, ranges, true)
	got := addrsToStrings(free)
	// /29 = .0..7; skipReserved drops .0 (network) and .7 (broadcast);
	// range removes .2,.3,.4 → free: .1, .5, .6.
	want := []string{"10.0.0.1", "10.0.0.5", "10.0.0.6"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestComputeFreeAddrs_IPv6SkipsAnycast(t *testing.T) {
	// IPv6 /125 has 8 addresses; we drop the subnet-router anycast (first).
	p := mustPrefix("2001:db8::/125")
	used := newAddrSet(nil)
	free := computeFreeAddrs(p, used, nil, true)
	if len(free) != 7 {
		t.Errorf("expected 7 free addresses for IPv6 /125, got %d (%v)", len(free), addrsToStrings(free))
	}
	if free[0] == mustAddr("2001:db8::") {
		t.Error("first IPv6 address (subnet-router anycast) should be reserved")
	}
}

// ---- freePrefixes ----

func TestFreePrefixes_NoChildren(t *testing.T) {
	parent := mustPrefix("10.0.0.0/24")
	free := freePrefixes(parent, nil)
	// The whole /24 is free → one result.
	if len(free) != 1 || free[0].String() != "10.0.0.0/24" {
		t.Errorf("expected single free /24, got %v", prefixesToStrings(free))
	}
}

func TestFreePrefixes_ChildConsumesAll(t *testing.T) {
	parent := mustPrefix("10.0.0.0/24")
	child := mustPrefix("10.0.0.0/24")
	free := freePrefixes(parent, []netip.Prefix{child})
	if len(free) != 0 {
		t.Errorf("expected no free space when child == parent, got %v", prefixesToStrings(free))
	}
}

func TestFreePrefixes_PartialChildSplits(t *testing.T) {
	// /24 split: child .0/25 covers the lower half → free is the upper /25.
	parent := mustPrefix("10.0.0.0/24")
	child := mustPrefix("10.0.0.0/25")
	free := freePrefixes(parent, []netip.Prefix{child})
	got := prefixesToStrings(free)
	want := []string{"10.0.0.128/25"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFreePrefixes_TwoChildrenLeaveMiddle(t *testing.T) {
	// /24 with two /25 children → entirely consumed.
	parent := mustPrefix("10.0.0.0/24")
	free := freePrefixes(parent, []netip.Prefix{
		mustPrefix("10.0.0.0/25"),
		mustPrefix("10.0.0.128/25"),
	})
	if len(free) != 0 {
		t.Errorf("expected no free space, got %v", prefixesToStrings(free))
	}
}

func TestFreePrefixes_SmallChildInBigParent(t *testing.T) {
	// /24 with a /30 child leaves many small free CIDRs that merge into larger
	// ones. The result should be three: the block before .0/30 is impossible
	// (it's at the start), so free space = upper 3/4 of the /24 minus a /30.
	parent := mustPrefix("10.0.0.0/24")
	child := mustPrefix("10.0.0.0/30")
	free := freePrefixes(parent, []netip.Prefix{child})
	// The /30 sits at the very bottom of the /24, so free space begins at
	// .4 and the largest free block is 10.0.0.4/30, 10.0.0.8/29, ..., 10.0.0.128/25.
	// We just assert the first free prefix is 10.0.0.4/30 and that the
	// total adds up to 256 - 4 = 252 addresses.
	if len(free) == 0 {
		t.Fatal("expected free space")
	}
	if free[0].String() != "10.0.0.4/30" {
		t.Errorf("first free block = %s, want 10.0.0.4/30", free[0])
	}
	total := 0
	for _, f := range free {
		total += 1 << uint(32-f.Bits())
	}
	if total != 252 {
		t.Errorf("total free addresses = %d, want 252", total)
	}
}

// ---- firstFitPrefix ----

func TestFirstFitPrefix_FirstFit(t *testing.T) {
	free := []netip.Prefix{mustPrefix("10.0.0.0/30")}
	pfx, ok := firstFitPrefix(free, 30)
	if !ok || pfx.String() != "10.0.0.0/30" {
		t.Errorf("firstFitPrefix(/30) = %v ok=%v, want 10.0.0.0/30", pfx, ok)
	}
}

func TestFirstFitPrefix_NeedSmallerFromBigger(t *testing.T) {
	free := []netip.Prefix{mustPrefix("10.0.0.0/24")}
	pfx, ok := firstFitPrefix(free, 26)
	if !ok || pfx.String() != "10.0.0.0/26" {
		t.Errorf("firstFitPrefix(/26 from /24) = %v ok=%v, want 10.0.0.0/26", pfx, ok)
	}
}

func TestFirstFitPrefix_NoneFits(t *testing.T) {
	free := []netip.Prefix{mustPrefix("10.0.0.0/30")}
	_, ok := firstFitPrefix(free, 24)
	if ok {
		t.Error("expected no fit (/24 from /30)")
	}
}

// ---- containment & overlap helpers ----

func TestContainsPrefix(t *testing.T) {
	cases := []struct {
		outer, inner string
		want         bool
	}{
		{"10.0.0.0/24", "10.0.0.0/25", true},
		{"10.0.0.0/24", "10.0.0.128/25", true},
		{"10.0.0.0/24", "10.0.0.0/24", true},
		{"10.0.0.0/25", "10.0.0.128/25", false},
		{"10.0.0.0/24", "10.0.1.0/24", false},
	}
	for _, c := range cases {
		got := containsPrefix(mustPrefix(c.outer), mustPrefix(c.inner))
		if got != c.want {
			t.Errorf("containsPrefix(%s, %s) = %v, want %v", c.outer, c.inner, got, c.want)
		}
	}
}

func TestOverlapsPrefix(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"10.0.0.0/24", "10.0.0.0/25", true},
		{"10.0.0.0/25", "10.0.0.128/25", false},
		{"10.0.0.0/24", "10.0.1.0/24", false},
		{"10.0.0.0/30", "10.0.0.0/30", true},
	}
	for _, c := range cases {
		got := overlapsPrefix(mustPrefix(c.a), mustPrefix(c.b))
		if got != c.want {
			t.Errorf("overlapsPrefix(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// ---- parsing ----

func TestParsePrefixAddr(t *testing.T) {
	cases := []struct {
		in   string
		addr string
		bits int
		ok   bool
	}{
		{"10.0.0.1/24", "10.0.0.1", 24, true},
		{"10.0.0.5", "10.0.0.5", 32, true},
		{"2001:db8::1/64", "2001:db8::1", 64, true},
		{"2001:db8::abc", "2001:db8::abc", 128, true},
		{"garbage", "", 0, false},
	}
	for _, c := range cases {
		a, b, ok := parsePrefixAddr(c.in)
		if ok != c.ok {
			t.Errorf("parsePrefixAddr(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok {
			if a.String() != c.addr || b != c.bits {
				t.Errorf("parsePrefixAddr(%q) = (%s, %d), want (%s, %d)", c.in, a, b, c.addr, c.bits)
			}
		}
	}
}

func TestRangeBounds(t *testing.T) {
	s, e, ok := rangeBounds("10.0.0.1/24", "10.0.0.10/24")
	if !ok {
		t.Fatal("expected ok")
	}
	if s.String() != "10.0.0.1" || e.String() != "10.0.0.10" {
		t.Errorf("rangeBounds = (%s, %s), want (.1, .10)", s, e)
	}

	// start > end → invalid
	_, _, ok = rangeBounds("10.0.0.10", "10.0.0.1")
	if ok {
		t.Error("expected rangeBounds to reject start > end")
	}
}

// ---- shouldSkipReserved ----

func TestShouldSkipReserved(t *testing.T) {
	cases := []struct {
		prefix string
		pool   bool
		want   bool
	}{
		{"10.0.0.0/24", false, true},  // normal v4 → skip
		{"10.0.0.0/31", false, false}, // v4 /31 → don't skip
		{"10.0.0.0/30", false, true},  // v4 /30 → skip
		{"10.0.0.0/24", true, false},  // pool → never skip
		{"2001:db8::/64", false, true},
		{"2001:db8::/127", false, false},
	}
	for _, c := range cases {
		got := shouldSkipReserved(mustPrefix(c.prefix), c.pool)
		if got != c.want {
			t.Errorf("shouldSkipReserved(%s, pool=%v) = %v, want %v", c.prefix, c.pool, got, c.want)
		}
	}
}

// ---- helpers ----

func mustAddr(s string) netip.Addr {
	a, err := netip.ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return a
}

func mustPrefix(s string) netip.Prefix {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return p.Masked()
}
