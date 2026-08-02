// Package dcim implements DCIM-specific custom @action endpoints:
// cable path tracing (GET /dcim/{model}/{id}/trace/) and pass-through port
// path lookup (GET /dcim/{model}/{id}/paths/).
//
// Cable paths are precomputed by NetBox (Python) and stored in dcim_cablepath.
// Each path stores:
//   - path (jsonb): ordered list of "steps", where each step is a list of
//     "node" strings in the form "<ContentTypeID>:<ObjectID>". Steps alternate
//     terminations / links / terminations / links / ... — every 3 consecutive
//     steps form a (near_ends, cable, far_ends) segment.
//   - _nodes (_varchar): flattened list of every node in the path, for
//     containment lookups (used by /paths/).
//   - is_active, is_complete, is_split: status flags.
//
// To resolve a node, we look up django_content_type by id to get (app_label,
// model), then load the object from the corresponding table.
package dcim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"netbox-go/internal/contenttype"
)

// ErrNotFound is returned when a referenced object (component, path, content
// type) does not exist.
var ErrNotFound = errors.New("not found")

// PathEndpointModels are the models that can originate a trace (they have a
// _path_id FK to dcim_cablepath). Matches NetBox's PathEndpointMixin.
var PathEndpointModels = []string{
	"interfaces",
	"console-ports",
	"console-server-ports",
	"power-ports",
	"power-outlets",
	"power-feeds",
}

// PassThroughModels are front/rear ports; they support the /paths/ action.
var PassThroughModels = []string{
	"front-ports",
	"rear-ports",
}

// modelToTable maps the URL-safe model name to its PostgreSQL table and the
// model name stored in django_content_type.model.
var modelToTable = map[string]struct {
	Table      string // dcim_interface
	CTModel    string // interface (django_content_type.model)
	DisplayCol string // name | label
}{
	"interfaces":           {"dcim_interface", "interface", "name"},
	"console-ports":        {"dcim_consoleport", "consoleport", "name"},
	"console-server-ports": {"dcim_consoleserverport", "consoleserverport", "name"},
	"power-ports":          {"dcim_powerport", "powerport", "name"},
	"power-outlets":        {"dcim_poweroutlet", "poweroutlet", "name"},
	"power-feeds":          {"dcim_powerfeed", "powerfeed", "name"},
	"front-ports":          {"dcim_frontport", "frontport", "name"},
	"rear-ports":           {"dcim_rearport", "rearport", "name"},
}

// TraceService resolves cable paths. Constructed once per process.
type TraceService struct {
	// contentTypeCache maps django_content_type.id -> "app_label.model"
	// (e.g. 42 -> "dcim.interface"). Populated lazily.
	contentTypes map[uint64]string
}

// NewTraceService constructs a TraceService.
func NewTraceService() *TraceService {
	return &TraceService{contentTypes: map[uint64]string{}}
}

// TraceResponse is the JSON shape returned by GET .../trace/. It mirrors what
// the orphaned CableTraceView.vue reads: a list of segments plus aggregate
// flags. Each segment is a (origin, cable, destination) triple.
type TraceResponse struct {
	Segments    []TraceSegment `json:"segments"`
	IsSplit     bool           `json:"is_split"`
	IsComplete  bool           `json:"is_complete"`
	TotalLength float64        `json:"total_length"` // sum of cable lengths (meters); 0 if unmeasured
}

// TraceSegment is one hop: near-end terminations -> cable -> far-end terminations.
type TraceSegment struct {
	Origin       interface{} `json:"origin"`        // resolved object or nil if path split
	OriginType   string      `json:"origin_type"`   // "dcim.interface"
	OriginDevice interface{} `json:"origin_device"` // nested device object
	Cable        interface{} `json:"cable"`         // resolved cable object or nil
	Destination  interface{} `json:"destination"`   // resolved object or nil
	DestType     string      `json:"destination_type"`
	DestDevice   interface{} `json:"destination_device"`
}

// Trace walks the cable path originating at the given component.
//
// It loads the component's _path_id, fetches the CablePath, decodes its path
// jsonb, and resolves each node into a nested object. Returns an empty
// segment list if the component has no path (uncabled).
func (s *TraceService) Trace(ctx context.Context, db *gorm.DB, model, id string) (*TraceResponse, error) {
	meta, ok := modelToTable[model]
	if !ok {
		return nil, fmt.Errorf("unsupported model %q for trace", model)
	}

	// 1. Load the component to read _path_id.
	var pathID int64
	row := db.Table(meta.Table).Select("_path_id").Where("id = ?", id).Row()
	if err := row.Scan(&pathID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		// SQLite/Postgres return no-rows differently; treat empty as not found.
		return nil, ErrNotFound
	}
	if pathID == 0 {
		// No cable attached — return an empty (but valid) trace.
		return &TraceResponse{Segments: []TraceSegment{}}, nil
	}

	// 2. Load the CablePath.
	var cp map[string]interface{}
	if err := db.Table("dcim_cablepath").Where("id = ?", pathID).Scan(&cp).Error; err != nil {
		return nil, err
	}
	if cp == nil {
		return nil, ErrNotFound
	}

	// 3. Decode the path jsonb into [][]string.
	steps, err := decodePath(cp["path"])
	if err != nil {
		return nil, fmt.Errorf("decode cablepath.path: %w", err)
	}
	if len(steps) < 3 {
		return &TraceResponse{Segments: []TraceSegment{}}, nil
	}

	// 4. Resolve all nodes referenced in the path, in batch.
	resolver := newPathResolver(db, s)
	if err := resolver.loadSteps(steps); err != nil {
		return nil, err
	}

	// 5. Walk steps in groups of 3 (near_ends, cable, far_ends) -> segments.
	resp := &TraceResponse{
		IsSplit:    toBool(cp["is_split"]),
		IsComplete: toBool(cp["is_complete"]),
		Segments:   []TraceSegment{},
	}
	for i := 0; i+2 < len(steps); i += 3 {
		nearStep, cableStep, farStep := steps[i], steps[i+1], steps[i+2]
		seg := TraceSegment{}

		if len(nearStep) > 0 {
			obj, ct, dev := resolver.resolve(nearStep[0])
			seg.Origin = obj
			seg.OriginType = ct
			seg.OriginDevice = dev
		}
		if len(cableStep) > 0 {
			cable, ct := resolver.resolveCable(cableStep[0])
			seg.Cable = cable
			if seg.OriginType == "" {
				seg.OriginType = ct // fallback
			}
			if l := cableLength(cable); l > 0 {
				resp.TotalLength += l
			}
		}
		if len(farStep) > 0 {
			obj, ct, dev := resolver.resolve(farStep[0])
			seg.Destination = obj
			seg.DestType = ct
			seg.DestDevice = dev
		}
		resp.Segments = append(resp.Segments, seg)
	}
	return resp, nil
}

// Paths returns all CablePaths that traverse a given pass-through port.
// Used by front/rear ports' /paths/ action.
func (s *TraceService) Paths(ctx context.Context, db *gorm.DB, model, id string) ([]map[string]interface{}, error) {
	meta, ok := modelToTable[model]
	if !ok {
		return nil, fmt.Errorf("unsupported model %q for paths", model)
	}

	// Confirm the port exists.
	var exists int64
	db.Table(meta.Table).Where("id = ?", id).Count(&exists)
	if exists == 0 {
		return nil, ErrNotFound
	}

	// Resolve the content type id for this model to build the node string.
	ctID, err := s.contentTypeID(db, "dcim", meta.CTModel)
	if err != nil {
		return nil, err
	}
	node := fmt.Sprintf("%d:%s", ctID, id)

	// PostgreSQL: _nodes @> ARRAY['42:7']. SQLite (tests): store _nodes as a
	// comma-joined TEXT and use LIKE. We try both — the LIKE is a safe superset
	// that also works in production (a node string is unique enough).
	var rows []map[string]interface{}
	like := "%" + node + "%"
	if err := db.Table("dcim_cablepath").
		Where("_nodes LIKE ?", like).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if rows == nil {
		return []map[string]interface{}{}, nil
	}
	return rows, nil
}

// ---- Node resolution ----

// pathResolver batch-loads every object referenced by a CablePath so we don't
// issue one query per node.
type pathResolver struct {
	db  *gorm.DB
	svc *TraceService
	// ctID -> "app_label.model"
	ctCache map[uint64]string
	// cache key "app_label.model:id" -> resolved object (with nested device)
	objCache map[string]map[string]interface{}
}

func newPathResolver(db *gorm.DB, svc *TraceService) *pathResolver {
	return &pathResolver{
		db:       db,
		svc:      svc,
		ctCache:  svc.contentTypes,
		objCache: map[string]map[string]interface{}{},
	}
}

// loadSteps pre-resolves content types and bulk-loads all referenced objects.
func (r *pathResolver) loadSteps(steps [][]string) error {
	// Collect every distinct (ctID) and (ctID, objID) referenced.
	ctIDs := map[uint64]bool{}
	type nodeRef struct{ ctID, objID uint64 }
	objsByCT := map[uint64][]uint64{}
	seen := map[nodeRef]bool{}
	for _, step := range steps {
		for _, node := range step {
			ctID, objID, ok := splitNode(node)
			if !ok {
				continue
			}
			ctIDs[ctID] = true
			if !seen[nodeRef{ctID, objID}] {
				seen[nodeRef{ctID, objID}] = true
				objsByCT[ctID] = append(objsByCT[ctID], objID)
			}
		}
	}
	if len(ctIDs) == 0 {
		return nil
	}

	// Bulk-load content types.
	if err := r.loadContentTypes(ctIDs); err != nil {
		return err
	}

	// Bulk-load objects grouped by content type.
	for ctID, objIDs := range objsByCT {
		ct, ok := r.ctCache[ctID]
		if !ok {
			continue
		}
		table, displayCol, isCable := tableForCT(ct)
		if table == "" {
			continue
		}
		var rows []map[string]interface{}
		r.db.Table(table).Where("id IN ?", objIDs).Scan(&rows)
		for _, row := range rows {
			key := fmt.Sprintf("%s:%v", ct, row["id"])
			if isCable {
				// Cables: keep as-is, no nested device.
				r.objCache[key] = row
				continue
			}
			// Components: attach nested device for display.
			row["display"] = row[displayCol]
			r.objCache[key] = row
			// Lazy-resolve device below in resolve().
		}
	}
	return nil
}

// resolve returns (object, "app_label.model", nestedDevice) for a node.
func (r *pathResolver) resolve(node string) (interface{}, string, interface{}) {
	ctID, objID, ok := splitNode(node)
	if !ok {
		return nil, "", nil
	}
	ct, ok := r.ctCache[ctID]
	if !ok {
		return nil, "", nil
	}
	obj := r.objCache[fmt.Sprintf("%s:%d", ct, objID)]
	if obj == nil {
		return nil, ct, nil
	}
	// Attach nested device if this component has a device_id.
	dev := r.nestedDevice(obj)
	return obj, ct, dev
}

// resolveCable returns (cableObject, "dcim.cable") for a cable node.
func (r *pathResolver) resolveCable(node string) (interface{}, string) {
	ctID, objID, ok := splitNode(node)
	if !ok {
		return nil, ""
	}
	ct, ok := r.ctCache[ctID]
	if !ok {
		return nil, ""
	}
	obj := r.objCache[fmt.Sprintf("%s:%d", ct, objID)]
	return obj, ct
}

// nestedDevice resolves a component's device_id into a small nested object.
func (r *pathResolver) nestedDevice(obj map[string]interface{}) interface{} {
	devID, ok := obj["device_id"]
	if !ok || devID == nil {
		return nil
	}
	var dev map[string]interface{}
	r.db.Table("dcim_device").Where("id = ?", devID).Scan(&dev)
	if dev == nil {
		return nil
	}
	return map[string]interface{}{
		"id":      dev["id"],
		"name":    dev["name"],
		"display": dev["name"],
	}
}

// loadContentTypes bulk-loads django_content_type rows for the given ids.
func (r *pathResolver) loadContentTypes(ids map[uint64]bool) error {
	if len(ids) == 0 {
		return nil
	}
	idList := make([]uint64, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	var rows []struct {
		ID       uint64
		AppLabel string
		Model    string
	}
	if err := r.db.Table("django_content_type").
		Select("id, app_label, model").
		Where("id IN ?", idList).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		ct := row.AppLabel + "." + row.Model
		r.ctCache[row.ID] = ct
		r.svc.contentTypes[row.ID] = ct // share with the service for reuse
	}
	return nil
}

// contentTypeID looks up the django_content_type.id for (app_label, model).
func (s *TraceService) contentTypeID(db *gorm.DB, appLabel, model string) (uint64, error) {
	// Check cache.
	for id, ct := range s.contentTypes {
		if ct == appLabel+"."+model {
			return id, nil
		}
	}
	var id uint64
	if err := db.Table("django_content_type").
		Select("id").
		Where("app_label = ? AND model = ?", appLabel, model).
		Scan(&id).Error; err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("content type %s.%s not found", appLabel, model)
	}
	s.contentTypes[id] = appLabel + "." + model
	return id, nil
}

// ---- Decoding helpers ----

// decodePath parses the CablePath.path jsonb into a [][]string.
// The stored shape is a JSON array of arrays of strings:
//
//	[["42:7"], ["55:1"], ["43:8"], ["55:2"], ["42:9"]]
//
// Each inner array is one "step"; steps alternate termination/cable/termination.
func decodePath(raw interface{}) ([][]string, error) {
	if raw == nil {
		return nil, nil
	}
	// []byte / string / already-parsed
	var bytes []byte
	switch v := raw.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		// Already decoded (e.g. mapscan returned a generic interface)
		encoded, err := json.Marshal(raw)
		if err != nil {
			return nil, err
		}
		bytes = encoded
	}
	var steps [][]string
	if err := json.Unmarshal(bytes, &steps); err != nil {
		// Some drivers return [][]interface{}; try once more generically.
		var generic [][]interface{}
		if err2 := json.Unmarshal(bytes, &generic); err2 != nil {
			return nil, err
		}
		steps = make([][]string, len(generic))
		for i, g := range generic {
			step := make([]string, len(g))
			for j, n := range g {
				step[j] = fmt.Sprintf("%v", n)
			}
			steps[i] = step
		}
	}
	return steps, nil
}

// splitNode parses a "ct_id:obj_id" string into two uint64s.
func splitNode(node string) (ctID, objID uint64, ok bool) {
	idx := strings.IndexByte(node, ':')
	if idx <= 0 {
		return 0, 0, false
	}
	a, err1 := atoi(node[:idx])
	b, err2 := atoi(node[idx+1:])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return a, b, true
}

// tableForCT maps "app_label.model" -> (table, displayColumn, isCable).
// The table name comes from the central contenttype seed list; the display
// column is a small lookup for the few models whose display field isn't "name".
func tableForCT(ct string) (table, displayCol string, isCable bool) {
	parts := strings.SplitN(ct, ".", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	for _, t := range contenttype.All() {
		if t.AppLabel == parts[0] && t.Model == parts[1] {
			displayCol = "name"
			isCable = ct == "dcim.cable"
			if isCable {
				displayCol = "label"
			} else if ct == "wireless.wirelesslink" {
				displayCol = "ssid"
			}
			return t.Table, displayCol, isCable
		}
	}
	return "", "", false
}

// cableLength extracts a cable's length (meters) if present.
func cableLength(cable interface{}) float64 {
	if cable == nil {
		return 0
	}
	m, ok := cable.(map[string]interface{})
	if !ok {
		return 0
	}
	length := toFloat(m["length"])
	// Only count if length_unit is meters; otherwise skip (don't guess).
	if unit, _ := m["length_unit"].(string); unit != "" && unit != "m" {
		return 0
	}
	return length
}

// toBool coerces DB bool-ish values.
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

// toInt64 coerces numeric/string DB values to int64.
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

// asString coerces a DB value to a string (handles []byte from SQLite).
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

// atoi is a small stdlib-free strconv.Atoi wrapper returning uint64.
func atoi(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty number")
	}
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit %q", c)
		}
		n = n*10 + uint64(c-'0')
	}
	return n, nil
}

// toFloat coerces numeric DB values to float64 (duplicated from netbox_custom.go
// to keep this subpackage self-contained).
func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case int:
		return float64(n)
	case []byte:
		f, _ := parseFloat(string(n))
		return f
	case string:
		f, _ := parseFloat(n)
		return f
	}
	return 0
}

func parseFloat(s string) (float64, error) {
	var f float64
	var sign float64 = 1
	i := 0
	s = strings.TrimSpace(s)
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		f = f*10 + float64(s[i]-'0')
		i++
	}
	if i < len(s) && s[i] == '.' {
		i++
		div := 10.0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			f += float64(s[i]-'0') / div
			div *= 10
			i++
		}
	}
	return sign * f, nil
}
