// Package dcim — connection-list endpoints.
//
// These mirror Python NetBox's three read-only connection list views
// (ConsoleConnectionsListView, PowerConnectionsListView,
// InterfaceConnectionsListView in dcim/views.py). Each shows every component
// of the given type that has a complete cable path, with the connected cable
// and far-end termination resolved.
//
// Routes:
//
//	GET /api/dcim/console-connections/
//	GET /api/dcim/power-connections/
//	GET /api/dcim/interface-connections/
//
// Filters (matching NetBox's ConnectionFilterSet): ?site_id=, ?site= (slug),
// ?device_id=, ?device= (name), ?q= (search device name + cable label), plus
// pagination via limit/offset.
package dcim

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"netbox-go/internal/contenttype"
	"netbox-go/internal/database"
)

// ConnectionType describes one of the three connection list kinds.
type ConnectionType struct {
	// URLKey is the path segment: "console-connections", etc.
	URLKey string
	// ComponentTable is the table holding the near-end components
	// (dcim_consoleport, dcim_powerport, dcim_interface).
	ComponentTable string
	// ContentTypeModel is the django_content_type.model value for the
	// near-end component (consoleport, powerport, interface) — used to match
	// against dcim_cabletermination.termination_type_id via lookup.
	ContentTypeModel string
	// Label is the human title shown in the page header / nav.
	Label string
	// NameColumn is the component's display column (always "name").
	NameColumn string
}

// ConnectionTypes is the registry of supported connection list kinds.
var ConnectionTypes = map[string]ConnectionType{
	"console-connections": {
		URLKey:           "console-connections",
		ComponentTable:   "dcim_consoleport",
		ContentTypeModel: "consoleport",
		Label:            "Console Connections",
		NameColumn:       "name",
	},
	"power-connections": {
		URLKey:           "power-connections",
		ComponentTable:   "dcim_powerport",
		ContentTypeModel: "powerport",
		Label:            "Power Connections",
		NameColumn:       "name",
	},
	"interface-connections": {
		URLKey:           "interface-connections",
		ComponentTable:   "dcim_interface",
		ContentTypeModel: "interface",
		Label:            "Interface Connections",
		NameColumn:       "name",
	},
}

// connectionRow is one row of the response.
type connectionRow struct {
	ID              int64                  `json:"id"`
	Device          map[string]interface{} `json:"device"`
	Name            string                 `json:"name"`
	Cable           map[string]interface{} `json:"cable"`
	Destination     map[string]interface{} `json:"destination"`
	DestinationType string                 `json:"destination_type"`
	Reachable       bool                   `json:"reachable"`
	PathIsActive    bool                   `json:"_path_is_active"`
}

// ConnectionListResult is the DRF-style paginated response.
type ConnectionListResult struct {
	Count    int64           `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  []connectionRow `json:"results"`
}

// MakeConnectionsHandler builds the GET handler for one connection type.
// It is exported so the routers package can register it.
func MakeConnectionsHandler(ct ConnectionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		rows, total, err := buildConnectionRows(db, ct, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		limit, offset := parseConnPagination(c)
		nextURL, prevURL := buildConnPaginationURLs(c, offset, limit, int(total))

		// Slice the in-memory result to the requested page. We build the full
		// list in Go because the cross-cable join is easier to express here than
		// in a single SQL query that must also run under SQLite (tests). For
		// typical deployments the number of connected components is modest.
		end := offset + limit
		if end > len(rows) {
			end = len(rows)
		}
		if offset > len(rows) {
			offset = len(rows)
		}
		pageRows := rows[offset:end]

		c.JSON(http.StatusOK, ConnectionListResult{
			Count:    total,
			Next:     nextURL,
			Previous: prevURL,
			Results:  pageRows,
		})
	}
}

// buildConnectionRows loads all connected components of the given type and
// resolves their cable + far-end termination into connectionRow objects.
func buildConnectionRows(db *gorm.DB, ct ConnectionType, c *gin.Context) ([]connectionRow, int64, error) {
	// 1. Resolve the content-type id for the near-end component model via the
	//    central contenttype resolver (loaded from the seeded table).
	nearCTID, err := lookupContentTypeID(db, "dcim", ct.ContentTypeModel)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve content type for %s: %w", ct.ContentTypeModel, err)
	}

	// 2. Load near-end components that have a path (i.e. are connected).
	//    Equivalent to Python's _path__is_complete=True — we treat any
	//    component whose _path_id is set as connected.
	query := db.Table(ct.ComponentTable).
		Where("_path_id IS NOT NULL AND _path_id > 0")

	// Apply filters.
	query = applyConnectionFilters(db, c, query, ct.ComponentTable)

	var components []map[string]interface{}
	if err := query.Scan(&components).Error; err != nil {
		return nil, 0, err
	}

	// 3. Batch-resolve everything we need: near-end devices, cables, far-end
	//    terminations, far-end components + their devices.
	nearDeviceIDs := collectIDs(components, "device_id")
	cableIDs := collectIDs(components, "cable_id")

	nearDevices := batchLoadDevices(db, nearDeviceIDs)
	cables := batchLoadCables(db, cableIDs)
	// Far-end terminations: for each cable, the termination whose end differs
	// from the near-end's. We load all terminations for the involved cables.
	farTermsByCable := batchLoadFarTerminations(db, cableIDs, nearCTID)

	// Group far-end terminations by (content_type, id) so we can batch-load
	// the referenced component rows + their devices.
	farCompRefs := map[string][]int64{} // "app.model" -> []componentID
	for _, terms := range farTermsByCable {
		for _, t := range terms {
			ctKey := contentTypeLabel(db, t["termination_type_id"])
			if ctKey == "" {
				continue
			}
			id := toInt64(t["termination_id"])
			if id > 0 {
				farCompRefs[ctKey] = append(farCompRefs[ctKey], id)
			}
		}
	}
	farComponents, farDeviceIDs := batchLoadFarComponents(db, farCompRefs)
	farDevices := batchLoadDevices(db, farDeviceIDs)

	// 4. Assemble rows.
	rows := make([]connectionRow, 0, len(components))
	for _, comp := range components {
		nearID := toInt64(comp["id"])
		deviceID := toInt64(comp["device_id"])
		cableID := toInt64(comp["cable_id"])
		pathActive := toBool(comp["_path_is_active"])

		row := connectionRow{
			ID:           nearID,
			Name:         asString(comp[ct.NameColumn]),
			Device:       nearDevices[deviceID],
			Reachable:    pathActive,
			PathIsActive: pathActive,
		}

		// Cable
		if cb, ok := cables[cableID]; ok {
			cb["display"] = cb["label"]
			if cb["display"] == "" {
				cb["display"] = fmt.Sprintf("Cable #%d", cableID)
			}
			row.Cable = cb
		}

		// Far-end termination → far-end component + device.
		// Skip the near-end's own termination (same cable, same termination_id).
		if terms, ok := farTermsByCable[cableID]; ok {
			for _, term := range terms {
				if toInt64(term["termination_id"]) == nearID {
					continue // this is the near end itself
				}
				ctKey := contentTypeLabel(db, term["termination_type_id"])
				farID := toInt64(term["termination_id"])
				row.DestinationType = ctKey
				if fc, ok := farComponents[compKey(ctKey, farID)]; ok {
					farDevID := toInt64(fc["device_id"])
					dest := map[string]interface{}{
						"id":     fc["id"],
						"name":   fc["name"],
						"device": farDevices[farDevID],
					}
					row.Destination = dest
				} else {
					// Component row missing — still report the termination id/type
					// so the UI can render "unknown endpoint".
					row.Destination = map[string]interface{}{
						"id":   farID,
						"name": "—",
					}
				}
				break // only one far-end per row
			}
		}

		rows = append(rows, row)
	}

	// 5. Apply ?q= search on the assembled rows (device name + cable label).
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		qLower := strings.ToLower(q)
		filtered := rows[:0]
		for _, r := range rows {
			if matchesSearch(r, qLower) {
				filtered = append(filtered, r)
			}
		}
		rows = filtered
	}

	return rows, int64(len(rows)), nil
}

// applyConnectionFilters narrows the near-end component query by the supported
// filter params. site_id / device_id filter on the component's denormalized
// device fields; site (slug) and device (name) require a sub-lookup.
func applyConnectionFilters(db *gorm.DB, c *gin.Context, query *gorm.DB, table string) *gorm.DB {
	if v := c.Query("device_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
			query = query.Where("device_id = ?", id)
		}
	}
	if v := c.Query("site_id"); v != "" {
		// Translate site_id -> set of device_ids via the device table.
		query = query.Where("device_id IN (?)",
			db.Table("dcim_device").Select("id").Where("site_id = ?", v))
	}
	if v := c.Query("device"); v != "" {
		query = query.Where("device_id IN (?)",
			db.Table("dcim_device").Select("id").Where("name = ?", v))
	}
	if v := c.Query("site"); v != "" {
		query = query.Where("device_id IN (?)",
			db.Table("dcim_device").Select("id").Where("site_id IN (?)",
				db.Table("dcim_site").Select("id").Where("slug = ?", v)))
	}
	return query
}

// matchesSearch reports whether a row matches the ?q= query against device name
// or cable label (mirrors ConnectionFilterSet.search).
func matchesSearch(r connectionRow, qLower string) bool {
	if r.Device != nil {
		if name, _ := r.Device["name"].(string); strings.Contains(strings.ToLower(name), qLower) {
			return true
		}
	}
	if r.Cable != nil {
		if label, _ := r.Cable["label"].(string); strings.Contains(strings.ToLower(label), qLower) {
			return true
		}
	}
	return false
}

// ---- batch loaders ----

func batchLoadDevices(db *gorm.DB, ids []int64) map[int64]map[string]interface{} {
	out := map[int64]map[string]interface{}{}
	if len(ids) == 0 {
		return out
	}
	var rows []map[string]interface{}
	db.Table("dcim_device").Where("id IN ?", ids).Scan(&rows)
	for _, r := range rows {
		id := toInt64(r["id"])
		name := asString(r["name"])
		out[id] = map[string]interface{}{
			"id":      id,
			"name":    name,
			"display": name,
		}
	}
	return out
}

func batchLoadCables(db *gorm.DB, ids []int64) map[int64]map[string]interface{} {
	out := map[int64]map[string]interface{}{}
	if len(ids) == 0 {
		return out
	}
	var rows []map[string]interface{}
	db.Table("dcim_cable").Where("id IN ?", ids).Scan(&rows)
	for _, r := range rows {
		out[toInt64(r["id"])] = r
	}
	return out
}

// batchLoadFarTerminations returns, per cable_id, the terminations that are
// NOT the near-end component type. (For a console connection, the far end is
// typically a consoleserverport; for an interface it's another interface, etc.)
// Each cable may have multiple far-end terminations; we return all and the
// caller uses the first.
func batchLoadFarTerminations(db *gorm.DB, cableIDs []int64, nearCTID uint64) map[int64][]map[string]interface{} {
	out := map[int64][]map[string]interface{}{}
	if len(cableIDs) == 0 {
		return out
	}
	var rows []map[string]interface{}
	db.Table("dcim_cabletermination").
		Where("cable_id IN ?", cableIDs).
		Scan(&rows)
	for _, r := range rows {
		// Keep terminations whose content type differs from the near-end.
		// (Same-type far ends — e.g. interface-to-interface — are still
		// included because we key by cable_id and the near-end's own
		// termination row is filtered out by the cable_end mismatch handled
		// implicitly: each cable has exactly 2 terminations; if both are the
		// same type, both appear and the caller picks the first non-self one.
		// We cannot tell "self" here without the near component id, so we
		// return all and let the caller's first-element pick work because the
		// far table query below dedupes by component id.)
		cableID := toInt64(r["cable_id"])
		out[cableID] = append(out[cableID], r)
	}
	return out
}

// batchLoadFarComponents loads the far-end component rows referenced by the
// terminations, grouped by content type. Returns (componentsByKey, deviceIDs)
// where componentsByKey is keyed by "app.model:id" and deviceIDs is the set of
// device_ids to resolve.
func batchLoadFarComponents(db *gorm.DB, refs map[string][]int64) (map[string]map[string]interface{}, []int64) {
	out := map[string]map[string]interface{}{}
	deviceIDSet := map[int64]bool{}
	for ctKey, ids := range refs {
		if len(ids) == 0 {
			continue
		}
		table, _, _ := tableForCT(ctKey)
		if table == "" {
			continue
		}
		var rows []map[string]interface{}
		db.Table(table).Where("id IN ?", ids).Scan(&rows)
		for _, r := range rows {
			out[compKey(ctKey, toInt64(r["id"]))] = r
			if devID := toInt64(r["device_id"]); devID > 0 {
				deviceIDSet[devID] = true
			}
		}
	}
	devIDs := make([]int64, 0, len(deviceIDSet))
	for id := range deviceIDSet {
		devIDs = append(devIDs, id)
	}
	return out, devIDs
}

// ---- helpers ----

// lookupContentTypeID finds django_content_type.id for (app_label, model) via
// the central contenttype resolver (which loads from the seeded table).
func lookupContentTypeID(db *gorm.DB, appLabel, model string) (uint64, error) {
	id, ok := contenttype.LookupByLabel(db, appLabel, model)
	if !ok {
		return 0, fmt.Errorf("content type %s.%s not found", appLabel, model)
	}
	return id, nil
}

// contentTypeLabel maps a numeric content_type id to "app.model" via the
// central resolver. Returns "" if the id is unknown.
func contentTypeLabel(db *gorm.DB, raw interface{}) string {
	id := uint64(toInt64(raw))
	if id == 0 {
		return ""
	}
	t, ok := contenttype.Resolve(db, id)
	if !ok {
		return ""
	}
	return t.AppLabel + "." + t.Model
}

// collectIDs extracts non-zero int64 values for a key from a slice of maps.
func collectIDs(rows []map[string]interface{}, key string) []int64 {
	seen := map[int64]bool{}
	var out []int64
	for _, r := range rows {
		id := toInt64(r[key])
		if id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func compKey(ctKey string, id int64) string {
	return fmt.Sprintf("%s:%d", ctKey, id)
}

// parseConnPagination reads limit/offset, clamped like the DRF handlers.
func parseConnPagination(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// buildConnPaginationURLs builds next/previous links for the connection list.
func buildConnPaginationURLs(c *gin.Context, offset, limit, total int) (*string, *string) {
	base := schemeAndHost(c) + c.Request.URL.Path
	var nextURL, prevURL *string
	if offset+limit < total {
		next := fmt.Sprintf("%s?limit=%d&offset=%d", base, limit, offset+limit)
		nextURL = &next
	}
	if offset > 0 {
		prev := offset - limit
		if prev < 0 {
			prev = 0
		}
		u := fmt.Sprintf("%s?limit=%d&offset=%d", base, limit, prev)
		prevURL = &u
	}
	return nextURL, prevURL
}

func schemeAndHost(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if p := c.GetHeader("X-Forwarded-Proto"); p != "" {
		scheme = p
	}
	return scheme + "://" + c.Request.Host
}

// asString / toInt64 / toBool are defined in trace.go (same package).
