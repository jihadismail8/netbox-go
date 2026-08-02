// Package routers provides NetBox-compatible (DRF-style) API endpoints.
package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"netbox-go/internal/database"
)

// strPtr returns a pointer to a string.
func strPtr(s string) *string { return &s }

// ModelEndpointConfig defines a DRF-compatible endpoint mapping for a model.
type ModelEndpointConfig struct {
	TableName    string            // e.g. "dcim_site"
	ModelType    interface{}       // pointer to model struct, e.g. &model.DcimSite{}
	ListFields   []string          // fields to select for list view (snake_case)
	SearchCols   []string          // columns searched by ?q=
	FilterCols   map[string]string // query_param -> db_column
	OrderFields  map[string]bool   // whitelist of sortable fields (snake_case)
	NestedFields map[string]NestedRef
	DefaultSort  string                 // e.g. "name ASC"
	Defaults     map[string]interface{} // default values for NOT NULL columns (e.g. custom_field_data: {})

	// RetrieveHandler optionally overrides the generic GET /{base}/:id handler.
	// Set this when a model needs a custom detail response (e.g. racks append
	// devices[] + reservations[]). When nil, the generic makeRetrieveHandler is
	// used. This avoids route-registration collisions: the override is plugged in
	// here rather than registered as a second route on the same path.
	RetrieveHandler gin.HandlerFunc
}

// NestedRef defines how to resolve a foreign-key relation into a nested serializer object.
type NestedRef struct {
	Table      string      // related table name
	IDCol      string      // column on THIS table holding the FK id
	Model      interface{} // pointer to related model struct (for type reference)
	DisplayCol string      // column used for display/label, typically "name"
	SlugCol    string      // column used for slug, typically "slug"
}

// listResult is the DRF-compatible paginated response.
type listResult struct {
	Count    int64         `json:"count"`
	Next     *string       `json:"next"`
	Previous *string       `json:"previous"`
	Results  []interface{} `json:"results"`
}

// netboxModelRegistry maps URL path prefixes to model configs.
var netboxModelRegistry = map[string]ModelEndpointConfig{}

// RegisterModelEndpoint registers a DRF-compatible endpoint for a model.
func RegisterModelEndpoint(pathPrefix string, config ModelEndpointConfig) {
	netboxModelRegistry[strings.TrimSuffix(pathPrefix, "/")] = config
}

// UnregisterModelEndpoint removes a model from the DRF registry so that its
// routes are NOT auto-generated. Use this when a model has bespoke handlers
// (e.g. /api/users/tokens is served by auth_routes.go with user-scoped logic)
// and the generic CRUD registration would otherwise collide with them.
func UnregisterModelEndpoint(pathPrefix string) {
	delete(netboxModelRegistry, strings.TrimSuffix(pathPrefix, "/"))
}

// RegisterNetboxDRFRoutes registers DRF-compatible routes for all registered models.
func RegisterNetboxDRFRoutes(r *gin.Engine) {
	registerDRFRoutesOn(r, netboxModelRegistry)
}

// ---- List Handler ----

func makeListHandler(cfg ModelEndpointConfig, basePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()

		// Build base query — SELECT * to ensure all columns (including FK IDs needed
		// for nested resolution) are available. This matches DRF's default behavior
		// where list endpoints return all model fields.
		query := db.Table(cfg.TableName)

		// Apply search filter (?q=)
		if q := c.Query("q"); q != "" {
			searchFilter := buildSearchFilter(cfg.SearchCols, q)
			if searchFilter != "" {
				query = query.Where(searchFilter)
			}
		}

		// Apply standard filters
		query = applyStandardFilters(c, query, cfg)

		// Count total
		var total int64
		countQuery := db.Table(cfg.TableName)
		if q := c.Query("q"); q != "" {
			searchFilter := buildSearchFilter(cfg.SearchCols, q)
			if searchFilter != "" {
				countQuery = countQuery.Where(searchFilter)
			}
		}
		countQuery = applyStandardFilters(c, countQuery, cfg)
		countQuery.Count(&total)

		// Pagination (limit/offset)
		limit, offset := parsePagination(c)

		// Sorting
		order := parseOrdering(c, cfg)
		query = query.Order(order)

		// Apply pagination
		query = query.Limit(limit).Offset(offset)

		// Execute query — get raw map results
		var results []map[string]interface{}
		if err := query.Find(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
			return
		}

		// Resolve nested relations
		results = resolveNestedFields(results, cfg)

		// Add display field
		for i := range results {
			results[i]["display"] = results[i]["name"]
		}

		// Build next/previous URLs
		nextURL, prevURL := buildPaginationURLs(c, basePath, offset, limit, int(total))

		c.JSON(http.StatusOK, listResult{
			Count:    total,
			Next:     nextURL,
			Previous: prevURL,
			Results:  normalizeListResults(results),
		})
	}
}

// buildSearchFilter builds a case-insensitive search across multiple columns.
// Uses ILIKE on PostgreSQL and LIKE on SQLite (which is case-insensitive for
// ASCII by default), making the code testable with in-memory SQLite.
func buildSearchFilter(cols []string, q string) string {
	if len(cols) == 0 || q == "" {
		return ""
	}
	op := "ILIKE"
	if db := database.GetDB(); db != nil && db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		op = "LIKE"
	}
	conditions := make([]string, 0, len(cols))
	for _, col := range cols {
		conditions = append(conditions, fmt.Sprintf("%s %s '%%%s%%'", col, op, escapeSQL(q)))
	}
	return strings.Join(conditions, " OR ")
}

func escapeSQL(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "--", "")
	return s
}

func applyStandardFilters(c *gin.Context, query *gorm.DB, cfg ModelEndpointConfig) *gorm.DB {
	for param, col := range cfg.FilterCols {
		if col == "" {
			continue // skip special params like "q"
		}
		val := c.Query(param)
		if val == "" {
			continue
		}
		// Handle _in suffix (comma-separated values)
		if strings.HasSuffix(param, "_in") || strings.HasSuffix(param, "__in") {
			ids := strings.Split(val, ",")
			query = query.Where(fmt.Sprintf("%s IN (?)", col), ids)
			continue
		}
		query = query.Where(fmt.Sprintf("%s = ?", col), val)
	}

	// Handle id filter (can be repeated or comma-separated)
	if idStr := c.Query("id"); idStr != "" {
		ids := parseIDList(idStr)
		if len(ids) > 0 {
			query = query.Where("id IN (?)", ids)
		}
	}

	return query
}

func parseIDList(s string) []string {
	if strings.Contains(s, ",") {
		return strings.Split(s, ",")
	}
	return []string{s}
}

func parsePagination(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func parseOrdering(c *gin.Context, cfg ModelEndpointConfig) string {
	ordering := c.Query("ordering")
	if ordering == "" {
		if cfg.DefaultSort != "" {
			return cfg.DefaultSort
		}
		return "id ASC"
	}

	desc := false
	if strings.HasPrefix(ordering, "-") {
		desc = true
		ordering = strings.TrimPrefix(ordering, "-")
	}

	if !cfg.OrderFields[ordering] {
		if cfg.DefaultSort != "" {
			return cfg.DefaultSort
		}
		return "id ASC"
	}

	if desc {
		return ordering + " DESC"
	}
	return ordering + " ASC"
}

func buildPaginationURLs(c *gin.Context, basePath string, offset, limit, total int) (*string, *string) {
	var nextURL, prevURL *string

	baseURL := fmt.Sprintf("%s://%s%s", schemeFromRequest(c), c.Request.Host, basePath)

	if offset+limit < total {
		nextOffset := offset + limit
		next := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, nextOffset)
		nextURL = strPtr(next)
	}
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		prev := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, prevOffset)
		prevURL = strPtr(prev)
	}

	return nextURL, prevURL
}

func schemeFromRequest(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
}

// ---- Create Handler ----

func makeCreateHandler(cfg ModelEndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		// Convert nested objects to FK IDs
		cleanBody := normalizeRequestBody(body, cfg)

		// Apply defaults for NOT NULL columns (e.g. custom_field_data)
		for k, v := range cfg.Defaults {
			if _, exists := cleanBody[k]; !exists {
				cleanBody[k] = v
			}
		}

		db := database.GetDB()

		// Fill remaining NOT NULL columns (without DB defaults) with empty values
		fillNotNullDefaults(db, cfg.TableName, cleanBody)

		// Serialize map/slice values to JSON for SQL compatibility (e.g. custom_field_data)
		for k, v := range cleanBody {
			switch v.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err == nil {
					cleanBody[k] = string(jsonBytes)
				}
			}
		}

		result := db.Table(cfg.TableName).Create(cleanBody)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": result.Error.Error()})
			return
		}

		// Fetch the created record
		var record map[string]interface{}
		if id, ok := cleanBody["id"]; ok {
			db.Table(cfg.TableName).Where("id = ?", id).Scan(&record)
		} else {
			var lastID uint64
			db.Raw(fmt.Sprintf("SELECT MAX(id) FROM %s", cfg.TableName)).Scan(&lastID)
			db.Table(cfg.TableName).Where("id = ?", lastID).Scan(&record)
		}

		// Guard against nil record (can happen with some GORM/SQLite drivers)
		if record == nil {
			c.JSON(http.StatusCreated, cleanBody)
			return
		}

		// Resolve nested fields
		results := resolveNestedFields([]map[string]interface{}{record}, cfg)
		if len(results) > 0 && results[0] != nil {
			results[0]["display"] = results[0]["name"]
			c.JSON(http.StatusCreated, results[0])
			return
		}
		c.JSON(http.StatusCreated, record)
	}
}

// normalizeRequestBody converts nested objects (like {"region": {id: 1}}) to FK columns ({"region_id": 1}).
func normalizeRequestBody(body map[string]interface{}, cfg ModelEndpointConfig) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range body {
		// Skip null values (PATCH semantics)
		if v == nil {
			continue
		}

		// Check if this is a nested field
		if nestedRef, ok := cfg.NestedFields[k]; ok {
			// If it's an object with an id, extract the id
			if obj, isObj := v.(map[string]interface{}); isObj {
				if id, hasID := obj["id"]; hasID {
					result[nestedRef.IDCol] = id
				}
				continue
			}
			// If it's just an ID number, use it directly
			if idNum, isNum := v.(float64); isNum {
				result[nestedRef.IDCol] = int64(idNum)
				continue
			}
		}

		// Skip internal fields
		if k == "display" || k == "_meta" {
			continue
		}

		result[k] = v
	}
	return result
}

// ---- Retrieve Handler ----

func makeRetrieveHandler(cfg ModelEndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()

		var record map[string]interface{}
		if err := db.Table(cfg.TableName).Where("id = ?", id).Scan(&record).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}
		if record == nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}

		// Resolve nested fields
		results := resolveNestedFields([]map[string]interface{}{record}, cfg)
		if len(results) > 0 && results[0] != nil {
			results[0]["display"] = results[0]["name"]
			c.JSON(http.StatusOK, results[0])
			return
		}
		record["display"] = record["name"]
		c.JSON(http.StatusOK, record)
	}
}

// ---- Update Handler (PATCH semantics) ----

func makeUpdateHandler(cfg ModelEndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		cleanBody := normalizeRequestBody(body, cfg)

		db := database.GetDB()
		// Serialize map/slice values to JSON for SQL compatibility
		for k, v := range cleanBody {
			switch v.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err == nil {
					cleanBody[k] = string(jsonBytes)
				}
			}
		}

		result := db.Table(cfg.TableName).Where("id = ?", id).Updates(cleanBody)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": result.Error.Error()})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}

		// Fetch updated record
		var record map[string]interface{}
		db.Table(cfg.TableName).Where("id = ?", id).Scan(&record)
		if record == nil {
			c.JSON(http.StatusOK, cleanBody)
			return
		}

		results := resolveNestedFields([]map[string]interface{}{record}, cfg)
		if len(results) > 0 && results[0] != nil {
			results[0]["display"] = results[0]["name"]
			c.JSON(http.StatusOK, results[0])
			return
		}
		record["display"] = record["name"]
		c.JSON(http.StatusOK, record)
	}
}

// ---- Delete Handler ----

func makeDeleteHandler(cfg ModelEndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db := database.GetDB()
		result := db.Table(cfg.TableName).Where("id = ?", id).Delete(nil)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": result.Error.Error()})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"detail": "Not found."})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

// ---- Bulk Delete Handler ----

func makeBulkDeleteHandler(cfg ModelEndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			PK []uint64 `json:"pk"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if len(body.PK) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "No IDs provided"})
			return
		}

		db := database.GetDB()
		result := db.Table(cfg.TableName).Where("id IN ?", body.PK).Delete(nil)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": result.Error.Error()})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	}
}

// ---- Bulk Update Handler ----

func makeBulkUpdateHandler(cfg ModelEndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawBody, _ := c.GetRawData()
		var rawMap map[string]interface{}
		if err := json.Unmarshal(rawBody, &rawMap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		var pks []uint64
		if pkRaw, ok := rawMap["pk"]; ok {
			if pkSlice, isSlice := pkRaw.([]interface{}); isSlice {
				for _, pk := range pkSlice {
					if id, isNum := pk.(float64); isNum {
						pks = append(pks, uint64(id))
					}
				}
			}
		}
		delete(rawMap, "pk")

		if len(pks) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "No IDs provided"})
			return
		}

		cleanBody := normalizeRequestBody(rawMap, cfg)

		// Serialize map/slice values to JSON for SQL compatibility
		for k, v := range cleanBody {
			switch v.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err == nil {
					cleanBody[k] = string(jsonBytes)
				}
			}
		}

		db := database.GetDB()
		result := db.Table(cfg.TableName).Where("id IN ?", pks).Updates(cleanBody)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": result.Error.Error()})
			return
		}

		// Return updated records
		var records []map[string]interface{}
		db.Table(cfg.TableName).Where("id IN ?", pks).Find(&records)
		results := resolveNestedFields(records, cfg)
		for i := range results {
			results[i]["display"] = results[i]["name"]
		}

		c.JSON(http.StatusOK, listResult{
			Count:   int64(len(results)),
			Results: normalizeListResults(results),
		})
	}
}

// ---- Helper: fillNotNullDefaults ----

// notNullColumn describes a column that is NOT NULL and has no database default.
type notNullColumn struct {
	Name     string
	DataType string
}

// notNullCache caches the NOT NULL column metadata per table to avoid repeated
// information_schema queries on every create request.
var notNullCache = map[string][]notNullColumn{}

// fillNotNullDefaults queries the database schema for NOT NULL columns without
// defaults that are missing from the body, and fills them with type-appropriate
// empty values (empty string for varchar/text, {} for jsonb, etc.). This mirrors
// Django's behavior where CharFields and TextFields default to "" and JSONFields
// default to {}.
func fillNotNullDefaults(db *gorm.DB, tableName string, body map[string]interface{}) {
	cols := getNotNullColumns(db, tableName)
	for _, col := range cols {
		if _, exists := body[col.Name]; exists {
			continue
		}
		body[col.Name] = defaultValueForType(col.DataType)
	}
}

// getNotNullColumns returns the NOT NULL columns (without DB defaults) for a table.
// Results are cached per table.
func getNotNullColumns(db *gorm.DB, tableName string) []notNullColumn {
	if cols, ok := notNullCache[tableName]; ok {
		return cols
	}

	var cols []notNullColumn

	// Query information_schema for PostgreSQL
	// Only applicable when using PostgreSQL (not SQLite in tests)
	if db.Dialector != nil && db.Dialector.Name() == "postgres" {
		rows, err := db.Raw(`
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_name = ?
			  AND is_nullable = 'NO'
			  AND column_default IS NULL
			  AND column_name NOT IN ('id')
		`, tableName).Rows()
		if err != nil {
			notNullCache[tableName] = nil
			return nil
		}
		defer rows.Close()

		for rows.Next() {
			var name, dataType string
			if err := rows.Scan(&name, &dataType); err != nil {
				continue
			}
			cols = append(cols, notNullColumn{Name: name, DataType: dataType})
		}
	}

	notNullCache[tableName] = cols
	return cols
}

// defaultValueForType returns an appropriate empty value for a PostgreSQL data type.
func defaultValueForType(dataType string) interface{} {
	switch dataType {
	case "jsonb", "json":
		return "{}"
	case "boolean":
		return false
	case "integer", "bigint", "smallint", "numeric", "real", "double precision":
		return 0
	default:
		// character varying, text, etc. → empty string
		return ""
	}
}

// ---- Helper: Resolve Nested Relations ----

func resolveNestedFields(records []map[string]interface{}, cfg ModelEndpointConfig) []map[string]interface{} {
	// Collect all unique FK IDs for each nested field
	nestedIDs := make(map[string]map[interface{}]bool) // fieldName -> set of IDs
	for _, record := range records {
		for fieldName, nestedRef := range cfg.NestedFields {
			fkID := record[nestedRef.IDCol]
			if fkID == nil || fkID == 0 {
				continue
			}
			if nestedIDs[fieldName] == nil {
				nestedIDs[fieldName] = make(map[interface{}]bool)
			}
			nestedIDs[fieldName][fkID] = true
		}
	}

	// Fetch nested objects in batch
	nestedCache := make(map[string]map[interface{}]map[string]interface{}) // fieldName -> id -> obj
	db := database.GetDB()
	for fieldName, nestedRef := range cfg.NestedFields {
		ids := nestedIDs[fieldName]
		if len(ids) == 0 {
			continue
		}

		idList := make([]interface{}, 0, len(ids))
		for id := range ids {
			idList = append(idList, id)
		}

		var nestedRecords []map[string]interface{}
		db.Table(nestedRef.Table).Where("id IN ?", idList).Find(&nestedRecords)

		if nestedCache[fieldName] == nil {
			nestedCache[fieldName] = make(map[interface{}]map[string]interface{})
		}
		for _, nr := range nestedRecords {
			id := nr["id"]
			nestedObj := map[string]interface{}{
				"id":   id,
				"name": nr[nestedRef.DisplayCol],
			}
			if nestedRef.SlugCol != "" && nr[nestedRef.SlugCol] != nil {
				nestedObj["slug"] = nr[nestedRef.SlugCol]
			}
			nestedCache[fieldName][id] = nestedObj
		}
	}

	// Assign nested objects to records
	for i := range records {
		// Skip nil records (can happen with certain GORM queries)
		if records[i] == nil {
			continue
		}
		for fieldName, nestedRef := range cfg.NestedFields {
			fkID := records[i][nestedRef.IDCol]
			if fkID == nil || fkID == 0 {
				records[i][fieldName] = nil
				continue
			}
			if cache, ok := nestedCache[fieldName]; ok {
				if obj, found := cache[fkID]; found {
					records[i][fieldName] = obj
				} else {
					records[i][fieldName] = nil
				}
			} else {
				records[i][fieldName] = nil
			}
		}
	}

	return records
}

// containsStr checks if a string exists in a slice.
func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// normalizeListResults converts []map[string]interface{} to []interface{}.
func normalizeListResults(records []map[string]interface{}) []interface{} {
	results := make([]interface{}, len(records))
	for i, r := range records {
		results[i] = r
	}
	return results
}

// ---- Auto-generated model registration support ----

// autoGenConfig is the struct format used by the code generator.
// It is a simplified version of ModelEndpointConfig that the generator
// can easily produce. RegisterAutoModelEndpoint converts it to the full
// ModelEndpointConfig and registers the endpoint.
type autoGenConfig struct {
	path         string
	tableName    string
	modelType    interface{}
	listFields   []string
	searchCols   []string
	defaultSort  string
	filterCols   map[string]string
	orderFields  []string
	nestedFKCols []string
	defaults     map[string]interface{} // default values for NOT NULL columns
}

// fkTableOverrides maps FK base names to their actual PostgreSQL table names
// for cases where the table name doesn't follow the simple module_prefix + singular
// convention (e.g. dcim_site.group_id → dcim_sitegroup, not dcim_site_group).
var fkTableOverrides = map[string]string{
	"group":           "dcim_sitegroup", // Site.group_id → dcim_sitegroup
	"scope":           "",               // Generic FK (polymorphic), resolved by scope_type_id
	"assigned_object": "",               // Generic FK (polymorphic)
	"parent_object":   "",               // Generic FK (polymorphic)
	"changed_object":  "",               // Generic FK (polymorphic)
	"related_object":  "",               // Generic FK (polymorphic)
	"action_object":   "",               // Generic FK (polymorphic)
}

// resolveFKTable resolves a FK base name (e.g. "region", "tenant") to a table name.
// It tries: explicit overrides → module convention (dcim_<name>, ipam_<name>, etc.)
// Returns "" if the table cannot be resolved (generic/polymorphic FKs).
func resolveFKTable(sourceTable, fkBase string) string {
	// 1. Check explicit overrides first
	if tbl, ok := fkTableOverrides[fkBase]; ok {
		return tbl
	}

	// 2. Generic/polymorphic FKs cannot be statically resolved
	if tbl, ok := fkTableOverrides[fkBase]; ok || fkBase == "" {
		return tbl
	}

	// 3. Try: same module prefix as source table
	//    e.g. source=dcim_site, fk=region → dcim_region
	//    e.g. source=dcim_location, fk=parent → dcim_location (self-ref)
	if idx := strings.Index(sourceTable, "_"); idx > 0 {
		module := sourceTable[:idx]
		candidate := module + "_" + fkBase
		// Verify the candidate exists in knownModelTables
		if knownModelTables[candidate] {
			return candidate
		}
		// Try without underscores in the base (e.g. fk="sitegroup" vs "site_group")
		mergedCandidate := module + "_" + strings.ReplaceAll(fkBase, "_", "")
		if knownModelTables[mergedCandidate] {
			return mergedCandidate
		}
		// Self-referential FK: if the FK base is a common self-ref name
		// (parent, master, source, etc.), resolve to the source table itself
		selfRefNames := map[string]bool{
			"parent": true, "master": true, "source": true,
			"primary": true, "backup": true, "peer": true,
		}
		if selfRefNames[fkBase] && knownModelTables[sourceTable] {
			return sourceTable
		}
	}

	// 4. Cross-module FKs — try common target modules
	crossModuleCandidates := []string{
		"dcim_" + fkBase,
		"ipam_" + fkBase,
		"tenancy_" + fkBase,
		"circuits_" + fkBase,
		"virtualization_" + fkBase,
		"extras_" + fkBase,
		"users_" + fkBase,
		"core_" + fkBase,
		"vpn_" + fkBase,
		"wireless_" + fkBase,
	}
	for _, c := range crossModuleCandidates {
		if knownModelTables[c] {
			return c
		}
	}

	return ""
}

// knownModelTables is populated at init() time from all Go model TableName() methods.
// Built dynamically in netbox_drf_autogen.go's init() via RegisterKnownTables().
var knownModelTables = map[string]bool{}

// RegisterKnownTables marks table names as existing, used for FK resolution.
func RegisterKnownTables(tables []string) {
	for _, t := range tables {
		knownModelTables[t] = true
	}
}

// RegisterAutoModelEndpoint converts an autoGenConfig to a full ModelEndpointConfig
// and registers it in the model registry.
func RegisterAutoModelEndpoint(c autoGenConfig) {
	// Convert orderFields []string -> map[string]bool
	orderMap := make(map[string]bool, len(c.orderFields))
	for _, f := range c.orderFields {
		orderMap[f] = true
	}

	// Build nested fields from FK column names, resolving target tables
	nestedFields := make(map[string]NestedRef, len(c.nestedFKCols))
	for _, fkCol := range c.nestedFKCols {
		targetTable := resolveFKTable(c.tableName, fkCol)
		if targetTable == "" {
			// Skip FKs that can't be resolved (polymorphic, generic)
			continue
		}
		nestedFields[fkCol] = NestedRef{
			Table:      targetTable,
			IDCol:      fkCol + "_id",
			Model:      nil,
			DisplayCol: "name",
			SlugCol:    "slug",
		}
	}

	// Ensure id is always in filterCols for detail lookups
	if c.filterCols == nil {
		c.filterCols = map[string]string{}
	}

	// Default NOT NULL columns present on all NetBox models.
	defaults := map[string]interface{}{"custom_field_data": map[string]interface{}{}}

	cfg := ModelEndpointConfig{
		TableName:    c.tableName,
		ModelType:    c.modelType,
		ListFields:   c.listFields,
		SearchCols:   c.searchCols,
		FilterCols:   c.filterCols,
		OrderFields:  orderMap,
		NestedFields: nestedFields,
		DefaultSort:  c.defaultSort,
		Defaults:     defaults,
	}

	RegisterModelEndpoint(c.path, cfg)
}

// registerDRFRoutesOn registers all DRF routes for the given registry onto a
// gin.IRouter. It deduplicates the Gin path strings (avoiding the trailing-slash
// collision panic) and guards against re-registration of identical paths.
func registerDRFRoutesOn(r gin.IRouter, registry map[string]ModelEndpointConfig) {
	seen := make(map[string]bool, len(registry)*8)
	regMethod := func(method, path string, h gin.HandlerFunc) {
		key := method + " " + path
		if seen[key] {
			return
		}
		seen[key] = true
		switch method {
		case http.MethodGet:
			r.GET(path, h)
		case http.MethodPost:
			r.POST(path, h)
		case http.MethodPatch:
			r.PATCH(path, h)
		case http.MethodPut:
			r.PUT(path, h)
		case http.MethodDelete:
			r.DELETE(path, h)
		}
	}

	for pathPrefix, config := range registry {
		base := pathPrefix

		// All routes registered WITHOUT trailing slashes.
		// The trailing-slash-stripping middleware in routers.go normalizes
		// incoming requests before routing:
		//   /api/dcim/sites/  -> /api/dcim/sites
		//   /api/dcim/sites/3/ -> /api/dcim/sites/3
		// This preserves PATCH/PUT request bodies (no 301/307 redirects).

		// GET list
		regMethod(http.MethodGet, base, makeListHandler(config, base))

		// POST create
		regMethod(http.MethodPost, base, makeCreateHandler(config))

		// GET/PATCH/PUT/DELETE detail by ID.
		// Use the per-model RetrieveHandler override if provided (e.g. the rack
		// enrichment handler); otherwise the generic makeRetrieveHandler.
		retrieve := config.RetrieveHandler
		if retrieve == nil {
			retrieve = makeRetrieveHandler(config)
		}
		regMethod(http.MethodGet, base+"/:id", retrieve)
		regMethod(http.MethodPatch, base+"/:id", makeUpdateHandler(config))
		regMethod(http.MethodPut, base+"/:id", makeUpdateHandler(config))
		regMethod(http.MethodDelete, base+"/:id", makeDeleteHandler(config))

		// Bulk operations
		regMethod(http.MethodPost, base+"/bulk_delete", makeBulkDeleteHandler(config))
		regMethod(http.MethodPatch, base+"/bulk_edit", makeBulkUpdateHandler(config))
	}
}

// RegisterNetboxDRFRoutesWithGroup registers DRF-compatible routes on a gin.IRouter
// (either *gin.Engine or *gin.RouterGroup) instead of the root engine.
// This allows auth middleware applied to the group to protect all DRF endpoints.
func RegisterNetboxDRFRoutesWithGroup(r gin.IRouter) {
	registerDRFRoutesOn(r, netboxModelRegistry)
}
