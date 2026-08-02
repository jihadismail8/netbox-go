// Package ipam — HTTP handlers for availability endpoints.
package ipam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"netbox-go/internal/database"
)

// AvailableService is the closure state shared by all IPAM availability
// handlers. Currently stateless; kept as a struct for future caching.
type AvailableService struct{}

// NewAvailableService constructs the service.
func NewAvailableService() *AvailableService { return &AvailableService{} }

// defaultIPLimit is the maximum number of available IPs/prefixes returned when
// ?limit= is omitted. Matches NetBox's default.
const defaultIPLimit = 100

// ---- Available IPs ----

// handleGetAvailableIPs handles GET .../available-ips/ for both prefixes and
// IP ranges. parentKind is "prefix" or "range".
func (s *AvailableService) HandleGetAvailableIPs(parentKind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		limit := parseLimit(c, defaultIPLimit)

		avail, err := s.computeAvailableIPs(database.GetDB(), parentKind, id, limit)
		if err != nil {
			if err == ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		// NetBox returns each IP as {"address": "<ip>/<mask>", "family": 4|6, ...}.
		results := make([]map[string]interface{}, 0, len(avail.addrs))
		for _, a := range avail.addrs {
			results = append(results, map[string]interface{}{
				"family":  familyOf(a),
				"address": fmt.Sprintf("%s/%d", a.String(), avail.maskBits),
			})
		}
		c.JSON(http.StatusOK, results)
	}
}

// handlePostAvailableIPs allocates one or more IPs from the parent.
// Body shape (NetBox-compatible): either an array of {description} objects, or
// a single object. The number of array entries determines how many IPs to claim.
func (s *AvailableService) HandlePostAvailableIPs(parentKind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()

		raws, err := parseIPRequestBody(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		avail, err := s.computeAvailableIPs(db, parentKind, id, len(raws))
		if err != nil {
			if err == ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if len(avail.addrs) < len(raws) {
			c.JSON(http.StatusConflict, gin.H{
				"detail": fmt.Sprintf(
					"Insufficient space available (requested %d, only %d available).",
					len(raws), len(avail.addrs)),
			})
			return
		}

		// Allocate in a transaction.
		created := make([]map[string]interface{}, 0, len(raws))
		txErr := db.Transaction(func(tx *gorm.DB) error {
			for i, raw := range raws {
				addr := avail.addrs[i]
				ip := map[string]interface{}{
					"address":           fmt.Sprintf("%s/%d", addr.String(), avail.maskBits),
					"status":            "active",
					"vrf_id":            avail.vrfID,
					"description":       strOrEmpty(raw["description"]),
					"dns_name":          strOrEmpty(raw["dns_name"]),
					"custom_field_data": "{}",
					"comments":          "",
				}
				if err := tx.Table("ipam_ipaddress").Create(ip).Error; err != nil {
					return err
				}
				created = append(created, ip)
			}
			return nil
		})
		if txErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": txErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

// availableIPs is the computed result for an availability query.
type availableIPs struct {
	addrs    []netip.Addr
	maskBits int
	vrfID    int64
}

// computeAvailableIPs returns up to `limit` free IPs in the parent prefix/range.
func (s *AvailableService) computeAvailableIPs(db *gorm.DB, parentKind, id string, limit int) (*availableIPs, error) {
	if parentKind == "prefix" {
		return s.availableIPsInPrefix(db, id, limit)
	}
	return s.availableIPsInRange(db, id, limit)
}

func (s *AvailableService) availableIPsInPrefix(db *gorm.DB, id string, limit int) (*availableIPs, error) {
	var p map[string]interface{}
	if err := db.Table("ipam_prefix").Where("id = ?", id).Scan(&p).Error; err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	prefixStr, _ := p["prefix"].(string)
	prefix, ok := ParsePrefix(prefixStr)
	if !ok {
		return nil, errStr("invalid prefix %q", prefixStr)
	}
	isPool := toBool(p["is_pool"])
	maskBits := prefix.Bits()

	// Load used child IPs (host address within the prefix).
	// In production (Postgres) we'd use the << operator; for SQLite/tests we
	// load all and filter in Go — both paths are correct.
	var ipRows []map[string]interface{}
	db.Table("ipam_ipaddress").Select("address").Scan(&ipRows)
	var usedAddrs []netip.Addr
	for _, r := range ipRows {
		host, ok := hostAddr(asString(r["address"]))
		if !ok || !prefix.Contains(host) {
			continue
		}
		usedAddrs = append(usedAddrs, host)
	}

	// Load populated IP ranges overlapping the prefix.
	var rangeRows []map[string]interface{}
	db.Table("ipam_iprange").
		Where("mark_populated = ?", true).
		Select("start_address, end_address").
		Scan(&rangeRows)
	var ranges []prefixRange
	for _, r := range rangeRows {
		st, en, ok := rangeBounds(asString(r["start_address"]), asString(r["end_address"]))
		if !ok {
			continue
		}
		if prefix.Contains(st) || prefix.Contains(en) {
			ranges = append(ranges, prefixRange{start: st, end: en})
		}
	}

	skipReserved := shouldSkipReserved(prefix, isPool)
	used := newAddrSet(usedAddrs)
	free := computeFreeAddrs(prefix, used, ranges, skipReserved)
	if limit > 0 && len(free) > limit {
		free = free[:limit]
	}
	return &availableIPs{
		addrs:    free,
		maskBits: maskBits,
		vrfID:    toInt64(p["vrf_id"]),
	}, nil
}

func (s *AvailableService) availableIPsInRange(db *gorm.DB, id string, limit int) (*availableIPs, error) {
	var r map[string]interface{}
	if err := db.Table("ipam_iprange").Where("id = ?", id).Scan(&r).Error; err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrNotFound
	}
	start, end, ok := rangeBounds(asString(r["start_address"]), asString(r["end_address"]))
	if !ok {
		return nil, errStr("invalid iprange bounds")
	}

	// Load used IPs that fall inside [start, end].
	var ipRows []map[string]interface{}
	db.Table("ipam_ipaddress").Select("address").Scan(&ipRows)
	var usedAddrs []netip.Addr
	for _, row := range ipRows {
		host, ok := hostAddr(asString(row["address"]))
		if !ok {
			continue
		}
		if !host.Less(start) && !end.Less(host) {
			usedAddrs = append(usedAddrs, host)
		}
	}

	used := newAddrSet(usedAddrs)
	var free []netip.Addr
	cur := start
	for {
		if !used.has(cur) {
			free = append(free, cur)
		}
		if cur == end {
			break
		}
		next := cur.Next()
		if !next.IsValid() || next.Less(start) {
			break
		}
		cur = next
		if limit > 0 && len(free) >= limit {
			break
		}
	}
	// maskBits for display: borrow from start_address if it carries a mask.
	_, bits, _ := parsePrefixAddr(asString(r["start_address"]))
	if bits == 0 {
		bits = 32
		if start.Is6() {
			bits = 128
		}
	}
	return &availableIPs{
		addrs:    free,
		maskBits: bits,
		vrfID:    toInt64(r["vrf_id"]),
	}, nil
}

// shouldSkipReserved mirrors NetBox: skip network/broadcast for "normal" IPv4
// prefixes (< /31) and the IPv6 subnet-router anycast (< /127), unless is_pool.
func shouldSkipReserved(prefix netip.Prefix, isPool bool) bool {
	if isPool {
		return false
	}
	if prefix.Addr().Is4() {
		return prefix.Bits() < 31
	}
	return prefix.Bits() < 127
}

// ---- Available prefixes ----

func (s *AvailableService) HandleGetAvailablePrefixes() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()

		parent, used, vrfID, ok, err := loadPrefixAndChildren(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}
		free := freePrefixes(parent, used)
		results := make([]map[string]interface{}, 0, len(free))
		for _, f := range free {
			results = append(results, map[string]interface{}{
				"family": familyOf(f.Addr()),
				"prefix": f.String(),
				"vrf_id": vrfID,
			})
		}
		c.JSON(http.StatusOK, results)
	}
}

func (s *AvailableService) HandlePostAvailablePrefixes() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()

		// Body: {"prefix_length": N} or [{...}, ...] with prefix_length each.
		var requests []map[string]interface{}
		body, _ := c.GetRawData()
		trimmed := strings.TrimSpace(string(body))
		if len(trimmed) > 0 && trimmed[0] == '[' {
			if err := json.Unmarshal(body, &requests); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
				return
			}
		} else {
			var single map[string]interface{}
			if err := json.Unmarshal(body, &single); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
				return
			}
			requests = []map[string]interface{}{single}
		}

		parent, used, vrfID, ok, err := loadPrefixAndChildren(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}

		// Allocate each requested prefix greedily, updating `used` as we go.
		created := make([]map[string]interface{}, 0, len(requests))
		txErr := db.Transaction(func(tx *gorm.DB) error {
			for _, req := range requests {
				wantBits := int(toInt64(req["prefix_length"]))
				free := freePrefixes(parent, used)
				pfx, ok := firstFitPrefix(free, wantBits)
				if !ok {
					return errStr("insufficient space for /%d", wantBits)
				}
				row := map[string]interface{}{
					"prefix":            pfx.String(),
					"status":            "active",
					"vrf_id":            vrfID,
					"description":       strOrEmpty(req["description"]),
					"custom_field_data": "{}",
					"comments":          "",
				}
				if err := tx.Table("ipam_prefix").Create(row).Error; err != nil {
					return err
				}
				created = append(created, row)
				used = append(used, pfx) // mark as consumed for the next iteration
			}
			return nil
		})
		if txErr != nil {
			c.JSON(http.StatusConflict, gin.H{"detail": txErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

// loadPrefixAndChildren loads the parent prefix row and all overlapping child
// prefixes (excluding self). Returns ok=false if the parent doesn't exist.
func loadPrefixAndChildren(db *gorm.DB, id string) (netip.Prefix, []netip.Prefix, int64, bool, error) {
	var p map[string]interface{}
	if err := db.Table("ipam_prefix").Where("id = ?", id).Scan(&p).Error; err != nil {
		return netip.Prefix{}, nil, 0, false, err
	}
	if p == nil {
		return netip.Prefix{}, nil, 0, false, nil
	}
	parent, ok := ParsePrefix(asString(p["prefix"]))
	if !ok {
		return netip.Prefix{}, nil, 0, false, errStr("invalid parent prefix")
	}
	var children []map[string]interface{}
	db.Table("ipam_prefix").Select("prefix").Scan(&children)
	var used []netip.Prefix
	for _, c := range children {
		cp, ok := ParsePrefix(asString(c["prefix"]))
		if !ok || cp == parent {
			continue
		}
		if overlapsPrefix(parent, cp) {
			used = append(used, cp)
		}
	}
	return parent, used, toInt64(p["vrf_id"]), true, nil
}

// ---- Available ASNs ----

func (s *AvailableService) HandleGetAvailableASNs() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()
		limit := parseLimit(c, defaultIPLimit)

		start, end, rirID, ok, err := loadASNRange(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}

		used := loadUsedASNs(db, start, end)
		results := []map[string]interface{}{}
		count := 0
		for asn := start; asn <= end; asn++ {
			if used[asn] {
				continue
			}
			results = append(results, map[string]interface{}{
				"asn":    asn,
				"rir_id": rirID,
			})
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
		c.JSON(http.StatusOK, results)
	}
}

func (s *AvailableService) HandlePostAvailableASNs() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()

		// Body: {"quantity": N} or [{...}] with one ASN allocation per entry.
		var requests []map[string]interface{}
		body, _ := c.GetRawData()
		trimmed := strings.TrimSpace(string(body))
		if len(trimmed) > 0 && trimmed[0] == '[' {
			if err := json.Unmarshal(body, &requests); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
				return
			}
		} else {
			var single map[string]interface{}
			if err := json.Unmarshal(body, &single); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
				return
			}
			// Honor {quantity: N} by expanding into N requests.
			if q := toInt64(single["quantity"]); q > 1 {
				requests = make([]map[string]interface{}, 0, int(q))
				for i := int64(0); i < q; i++ {
					requests = append(requests, map[string]interface{}{})
				}
			} else {
				requests = []map[string]interface{}{single}
			}
		}

		start, end, rirID, ok, err := loadASNRange(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}
		used := loadUsedASNs(db, start, end)

		// Collect the next N free ASNs.
		var alloc []int64
		for asn := start; asn <= end && len(alloc) < len(requests); asn++ {
			if !used[asn] {
				alloc = append(alloc, asn)
			}
		}
		if len(alloc) < len(requests) {
			c.JSON(http.StatusConflict, gin.H{
				"detail": fmt.Sprintf("Insufficient ASNs (requested %d, only %d available).",
					len(requests), len(alloc)),
			})
			return
		}

		created := make([]map[string]interface{}, 0, len(alloc))
		txErr := db.Transaction(func(tx *gorm.DB) error {
			for i, asn := range alloc {
				row := map[string]interface{}{
					"asn":               asn,
					"rir_id":            rirID,
					"description":       strOrEmpty(requests[i]["description"]),
					"custom_field_data": "{}",
				}
				if err := tx.Table("ipam_asn").Create(row).Error; err != nil {
					return err
				}
				created = append(created, row)
			}
			return nil
		})
		if txErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": txErr.Error()})
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

// ---- Shared helpers ----

// parseIPRequestBody parses the POST body for available-ips into a slice of
// request descriptors. Accepts a single object or an array of objects.
func parseIPRequestBody(c *gin.Context) ([]map[string]interface{}, error) {
	body, _ := c.GetRawData()
	trimmed := strings.TrimSpace(string(body))
	if len(trimmed) == 0 {
		// No body — allocate a single IP.
		return []map[string]interface{}{{}}, nil
	}
	if trimmed[0] == '[' {
		var raws []map[string]interface{}
		if err := json.Unmarshal(body, &raws); err != nil {
			return nil, fmt.Errorf("invalid request body: %w", err)
		}
		return raws, nil
	}
	var single map[string]interface{}
	if err := json.Unmarshal(body, &single); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}
	// Honor {quantity: N} for single-object allocation.
	if q := toInt64(single["quantity"]); q > 1 {
		out := make([]map[string]interface{}, 0, int(q))
		for i := int64(0); i < q; i++ {
			out = append(out, map[string]interface{}{
				"description": strOrEmpty(single["description"]),
				"dns_name":    strOrEmpty(single["dns_name"]),
			})
		}
		return out, nil
	}
	return []map[string]interface{}{single}, nil
}

func loadASNRange(db *gorm.DB, id string) (start, end, rirID int64, ok bool, err error) {
	var r map[string]interface{}
	if qErr := db.Table("ipam_asnrange").Where("id = ?", id).Scan(&r).Error; qErr != nil {
		return 0, 0, 0, false, qErr
	}
	if r == nil {
		return 0, 0, 0, false, nil
	}
	start = toInt64(r["start"])
	end = toInt64(r["end"])
	rirID = toInt64(r["rir_id"])
	if start == 0 && end == 0 {
		return 0, 0, 0, false, nil
	}
	return start, end, rirID, true, nil
}

func loadUsedASNs(db *gorm.DB, start, end int64) map[int64]bool {
	var rows []map[string]interface{}
	db.Table("ipam_asn").
		Where("asn >= ? AND asn <= ?", start, end).
		Select("asn").
		Scan(&rows)
	out := make(map[int64]bool, len(rows))
	for _, r := range rows {
		out[toInt64(r["asn"])] = true
	}
	return out
}

// ErrNotFound is the sentinel returned when the parent object is missing.
var ErrNotFound = errStr("not found")

// parseLimit reads ?limit=, clamped to [1, 1000]. 0 means "no limit".
func parseLimit(c *gin.Context, def int) int {
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			if n <= 0 {
				return 0
			}
			if n > 1000 {
				return 1000
			}
			return n
		}
	}
	return def
}

// familyOf returns 4 or 6.
func familyOf(a netip.Addr) int {
	if a.Is4() {
		return 4
	}
	return 6
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	}
	return fmt.Sprintf("%v", v)
}

func strOrEmpty(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		i, _ := strconv.ParseInt(string(n), 10, 64)
		return i
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	}
	return 0
}

func toBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case *bool:
		return b != nil && *b
	case int64:
		return b != 0
	case float64:
		return b != 0
	case []byte:
		s := string(b)
		return s == "1" || s == "true" || s == "True"
	case string:
		return b == "1" || b == "true" || b == "True"
	}
	return false
}

// time is imported to keep the import list stable; not otherwise used here.
var _ = time.Now
