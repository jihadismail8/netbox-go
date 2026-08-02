// Package routers — HTTP integration tests for custom @action endpoints.
// Uses an in-memory SQLite database, mirroring netbox_drf_integration_test.go.
package routers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/database"
	"netbox-go/internal/routers/dcim"
	"netbox-go/internal/routers/ipam"
)

// setupCustomTestDB creates an in-memory SQLite DB with the tables needed for
// the custom endpoint tests, and injects it into the database package.
func setupCustomTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	ddls := []string{
		`CREATE TABLE IF NOT EXISTS django_content_type (
			id INTEGER PRIMARY KEY,
			app_label TEXT NOT NULL,
			model TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_rack (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			facility_id TEXT DEFAULT '',
			status TEXT DEFAULT 'active',
			serial TEXT DEFAULT '',
			asset_tag TEXT DEFAULT '',
			form_factor TEXT DEFAULT '',
			width INTEGER DEFAULT 19,
			u_height INTEGER DEFAULT 42,
			desc_units INTEGER DEFAULT 0,
			outer_width INTEGER DEFAULT 0,
			outer_depth INTEGER DEFAULT 0,
			outer_unit TEXT DEFAULT '',
			comments TEXT DEFAULT '',
			location_id INTEGER DEFAULT NULL,
			role_id INTEGER DEFAULT NULL,
			site_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			weight INTEGER DEFAULT 0,
			max_weight INTEGER DEFAULT 0,
			weight_unit TEXT DEFAULT '',
			mounting_depth INTEGER DEFAULT 0,
			description TEXT DEFAULT '',
			starting_unit INTEGER DEFAULT 1,
			rack_type_id INTEGER DEFAULT NULL,
			airflow TEXT DEFAULT '',
			outer_height INTEGER DEFAULT 0,
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_devicetype (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			manufacturer_id INTEGER DEFAULT NULL,
			model TEXT NOT NULL,
			slug TEXT DEFAULT '',
			part_number TEXT DEFAULT '',
			u_height REAL DEFAULT 1,
			is_full_depth INTEGER DEFAULT 1,
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_device (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			device_type_id INTEGER NOT NULL,
			device_role_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			platform_id INTEGER DEFAULT NULL,
			site_id INTEGER DEFAULT NULL,
			location_id INTEGER DEFAULT NULL,
			rack_id INTEGER DEFAULT NULL,
			position REAL DEFAULT NULL,
			face TEXT DEFAULT '',
			status TEXT DEFAULT 'active',
			serial TEXT DEFAULT '',
			asset_tag TEXT DEFAULT '',
			cluster_id INTEGER DEFAULT NULL,
			virtual_chassis_id INTEGER DEFAULT NULL,
			vc_position INTEGER DEFAULT NULL,
			vc_priority INTEGER DEFAULT NULL,
			primary_ip4_id INTEGER DEFAULT NULL,
			primary_ip6_id INTEGER DEFAULT NULL,
			airflow TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_rackreservation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rack_id INTEGER NOT NULL,
			units TEXT DEFAULT '',
			user_id INTEGER DEFAULT 0,
			tenant_id INTEGER DEFAULT NULL,
			description TEXT DEFAULT '',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			status TEXT DEFAULT 'active',
			comments TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS core_datasource (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			source_url TEXT NOT NULL,
			status TEXT DEFAULT 'new',
			enabled INTEGER DEFAULT 1,
			ignore_rules TEXT DEFAULT '',
			parameters TEXT DEFAULT '{}',
			last_synced TEXT DEFAULT NULL,
			sync_interval INTEGER DEFAULT 0,
			description TEXT DEFAULT '',
			comments TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		// IPAM tables
		`CREATE TABLE IF NOT EXISTS ipam_prefix (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			prefix TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			is_pool INTEGER DEFAULT 0,
			mark_utilized INTEGER DEFAULT 0,
			description TEXT DEFAULT '',
			role_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			vlan_id INTEGER DEFAULT NULL,
			vrf_id INTEGER DEFAULT NULL,
			_scope_id INTEGER DEFAULT NULL,
			_depth INTEGER DEFAULT 0,
			_children INTEGER DEFAULT 0,
			scope_id INTEGER DEFAULT NULL,
			scope_type_id INTEGER DEFAULT NULL,
			comments TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS ipam_iprange (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			start_address TEXT NOT NULL,
			end_address TEXT NOT NULL,
			size INTEGER DEFAULT 0,
			status TEXT DEFAULT 'active',
			mark_utilized INTEGER DEFAULT 0,
			mark_populated INTEGER DEFAULT 0,
			description TEXT DEFAULT '',
			role_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			vrf_id INTEGER DEFAULT NULL,
			comments TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS ipam_ipaddress (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			address TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			role TEXT DEFAULT '',
			dns_name TEXT DEFAULT '',
			description TEXT DEFAULT '',
			assigned_object_id INTEGER DEFAULT NULL,
			assigned_object_type_id INTEGER DEFAULT NULL,
			nat_inside_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			vrf_id INTEGER DEFAULT NULL,
			comments TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS ipam_asnrange (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			start INTEGER NOT NULL,
			"end" INTEGER NOT NULL,
			rir_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			description TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS ipam_asn (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asn INTEGER NOT NULL,
			rir_id INTEGER DEFAULT NULL,
			tenant_id INTEGER DEFAULT NULL,
			description TEXT DEFAULT '',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		// Cable trace tables
		`CREATE TABLE IF NOT EXISTS dcim_interface (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			device_id INTEGER DEFAULT NULL,
			_path_id INTEGER DEFAULT 0,
			cable_id INTEGER DEFAULT 0,
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_cable (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			label TEXT DEFAULT '',
			status TEXT DEFAULT 'connected',
			type TEXT DEFAULT '',
			color TEXT DEFAULT '',
			length REAL DEFAULT 0,
			length_unit TEXT DEFAULT 'm',
			custom_field_data TEXT DEFAULT '{}',
			created TEXT DEFAULT '',
			last_updated TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_cablepath (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT DEFAULT '[]',
			_nodes TEXT DEFAULT '',
			is_active INTEGER DEFAULT 1,
			is_complete INTEGER DEFAULT 1,
			is_split INTEGER DEFAULT 0
		)`,
	}
	for _, ddl := range ddls {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}

	database.SetDB(db)

	// Wipe + reset sequences for test isolation.
	for _, tbl := range []string{
		"django_content_type", "dcim_rack", "dcim_devicetype", "dcim_device",
		"dcim_rackreservation", "core_datasource", "ipam_prefix", "ipam_iprange",
		"ipam_ipaddress", "ipam_asnrange", "ipam_asn",
		"dcim_interface", "dcim_cable", "dcim_cablepath",
	} {
		db.Exec("DELETE FROM " + tbl)
		db.Exec("DELETE FROM sqlite_sequence WHERE name = '" + tbl + "'")
	}
	return db
}

// doReq issues a JSON request and returns the recorder.
func doReq(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// =====================================================================
// Group A3: GET /api/status/
// =====================================================================

func TestStatus_ReturnsVersionAndHostname(t *testing.T) {
	setupCustomTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/status", makeStatusHandler())

	w := doReq(r, http.MethodGet, "/api/status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["netbox-version"] == nil {
		t.Error("expected netbox-version field")
	}
	if resp["go-version"] == nil {
		t.Error("expected go-version field")
	}
	if resp["hostname"] == nil {
		t.Error("expected hostname field")
	}
	apps, ok := resp["installed-apps"].([]interface{})
	if !ok || len(apps) == 0 {
		t.Error("expected non-empty installed-apps array")
	}
}

// =====================================================================
// Group A2: POST /api/core/data-sources/:id/sync/
// =====================================================================

func TestDataSourceSync_Returns202(t *testing.T) {
	db := setupCustomTestDB(t)
	// Seed a data source.
	db.Table("core_datasource").Create(map[string]interface{}{
		"name": "ds1", "type": "local", "source_url": "file:///tmp",
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/core/data-sources/:id/sync", makeDataSourceSyncHandler())

	w := doReq(r, http.MethodPost, "/api/core/data-sources/1/sync", nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	// Verify status was flipped to "syncing".
	var status string
	db.Table("core_datasource").Where("id = 1").Select("status").Row().Scan(&status)
	if status != "syncing" {
		t.Errorf("expected status='syncing' in DB, got %q", status)
	}
}

func TestDataSourceSync_NotFound(t *testing.T) {
	setupCustomTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/core/data-sources/:id/sync", makeDataSourceSyncHandler())

	w := doReq(r, http.MethodPost, "/api/core/data-sources/9999/sync", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// =====================================================================
// Group B: IPAM availability (HTTP)
// =====================================================================

func TestAvailableIPs_PrefixBasic(t *testing.T) {
	db := setupCustomTestDB(t)
	// /29 has 6 usable IPs (after network/broadcast).
	db.Table("ipam_prefix").Create(map[string]interface{}{
		"prefix": "10.0.0.0/29", "status": "active", "is_pool": 0,
	})
	// Mark one IP used.
	db.Table("ipam_ipaddress").Create(map[string]interface{}{
		"address": "10.0.0.1/29", "status": "active",
	})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.GET("/api/ipam/prefixes/:id/available-ips", svc.HandleGetAvailableIPs("prefix"))

	w := doReq(r, http.MethodGet, "/api/ipam/prefixes/1/available-ips", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	// /29 → .1..6 usable; .1 used → 5 free.
	if len(resp) != 5 {
		t.Errorf("expected 5 available IPs, got %d (%+v)", len(resp), resp)
	}
}

func TestAvailableIPs_NotFound(t *testing.T) {
	setupCustomTestDB(t)
	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.GET("/api/ipam/prefixes/:id/available-ips", svc.HandleGetAvailableIPs("prefix"))

	w := doReq(r, http.MethodGet, "/api/ipam/prefixes/9999/available-ips", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAvailableIPs_PostAllocates(t *testing.T) {
	db := setupCustomTestDB(t)
	db.Table("ipam_prefix").Create(map[string]interface{}{
		"prefix": "10.0.0.0/29", "status": "active", "is_pool": 0,
	})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.POST("/api/ipam/prefixes/:id/available-ips", svc.HandlePostAvailableIPs("prefix"))

	// Request 2 IPs.
	body := []map[string]interface{}{
		{"description": "first"},
		{"description": "second"},
	}
	w := doReq(r, http.MethodPost, "/api/ipam/prefixes/1/available-ips", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	if len(created) != 2 {
		t.Fatalf("expected 2 created, got %d", len(created))
	}
	// Verify the IPs now exist in the DB.
	var n int64
	db.Table("ipam_ipaddress").Count(&n)
	if n != 2 {
		t.Errorf("expected 2 IP rows in DB, got %d", n)
	}
}

func TestAvailableIPs_PostInsufficientSpace(t *testing.T) {
	db := setupCustomTestDB(t)
	// /30 has 2 usable IPs.
	db.Table("ipam_prefix").Create(map[string]interface{}{
		"prefix": "10.0.0.0/30", "status": "active", "is_pool": 0,
	})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.POST("/api/ipam/prefixes/:id/available-ips", svc.HandlePostAvailableIPs("prefix"))

	// Request 5 IPs — only 2 available.
	body := make([]map[string]interface{}, 5)
	w := doReq(r, http.MethodPost, "/api/ipam/prefixes/1/available-ips", body)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAvailableIPs_IPRange(t *testing.T) {
	db := setupCustomTestDB(t)
	db.Table("ipam_iprange").Create(map[string]interface{}{
		"start_address": "10.0.0.10/32", "end_address": "10.0.0.12/32",
		"status": "active",
	})
	db.Table("ipam_ipaddress").Create(map[string]interface{}{
		"address": "10.0.0.11/32", "status": "active",
	})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.GET("/api/ipam/ip-ranges/:id/available-ips", svc.HandleGetAvailableIPs("range"))

	w := doReq(r, http.MethodGet, "/api/ipam/ip-ranges/1/available-ips", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	// range .10..12 minus .11 = .10 and .12 = 2 free.
	if len(resp) != 2 {
		t.Errorf("expected 2 free IPs in range, got %d", len(resp))
	}
}

func TestAvailablePrefixes_Basic(t *testing.T) {
	db := setupCustomTestDB(t)
	// /24 with a /25 child → upper /25 free.
	db.Table("ipam_prefix").Create(map[string]interface{}{
		"prefix": "10.0.0.0/24", "status": "active",
	})
	db.Table("ipam_prefix").Create(map[string]interface{}{
		"prefix": "10.0.0.0/25", "status": "active",
	})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.GET("/api/ipam/prefixes/:id/available-prefixes", svc.HandleGetAvailablePrefixes())

	w := doReq(r, http.MethodGet, "/api/ipam/prefixes/1/available-prefixes", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 free prefix, got %d", len(resp))
	}
	if resp[0]["prefix"] != "10.0.0.128/25" {
		t.Errorf("expected 10.0.0.128/25, got %v", resp[0]["prefix"])
	}
}

func TestAvailableASNs_Basic(t *testing.T) {
	db := setupCustomTestDB(t)
	db.Table("ipam_asnrange").Create(map[string]interface{}{
		"name": "r1", "slug": "r1", "start": 65000, "end": 65005, "rir_id": 1,
	})
	// Use 65001 and 65003.
	db.Table("ipam_asn").Create(map[string]interface{}{"asn": 65001, "rir_id": 1})
	db.Table("ipam_asn").Create(map[string]interface{}{"asn": 65003, "rir_id": 1})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.GET("/api/ipam/asn-ranges/:id/available-asns", svc.HandleGetAvailableASNs())

	w := doReq(r, http.MethodGet, "/api/ipam/asn-ranges/1/available-asns", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	// range 65000-65005 minus {65001, 65003} = 4 free.
	if len(resp) != 4 {
		t.Errorf("expected 4 free ASNs, got %d (%+v)", len(resp), resp)
	}
}

func TestAvailableASNs_PostAllocates(t *testing.T) {
	db := setupCustomTestDB(t)
	db.Table("ipam_asnrange").Create(map[string]interface{}{
		"name": "r1", "slug": "r1", "start": 65000, "end": 65005, "rir_id": 1,
	})

	gin.SetMode(gin.TestMode)
	svc := ipam.NewAvailableService()
	r := gin.New()
	r.POST("/api/ipam/asn-ranges/:id/available-asns", svc.HandlePostAvailableASNs())

	body := map[string]interface{}{"quantity": 3}
	w := doReq(r, http.MethodPost, "/api/ipam/asn-ranges/1/available-asns", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	if len(created) != 3 {
		t.Fatalf("expected 3 created ASNs, got %d", len(created))
	}
	// Verify they took the first 3 free (65000, 65001, 65002).
	var n int64
	db.Table("ipam_asn").Count(&n)
	if n != 3 {
		t.Errorf("expected 3 ASN rows in DB, got %d", n)
	}
}

// =====================================================================
// Group C: Cable trace (HTTP)
// =====================================================================

// seedTraceTopology builds a single-hop path: interface1 --cable1--> interface2.
// Content type IDs: 1 = dcim.interface, 2 = dcim.cable.
func seedTraceTopology(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Content types.
	db.Table("django_content_type").Create(map[string]interface{}{"id": 1, "app_label": "dcim", "model": "interface"})
	db.Table("django_content_type").Create(map[string]interface{}{"id": 2, "app_label": "dcim", "model": "cable"})
	// Devices (for nested device resolution).
	db.Table("dcim_device").Create(map[string]interface{}{"id": 10, "name": "dev-A"})
	db.Table("dcim_device").Create(map[string]interface{}{"id": 20, "name": "dev-B"})
	// Interfaces.
	db.Table("dcim_interface").Create(map[string]interface{}{
		"id": 100, "name": "eth0", "device_id": 10, "_path_id": 500,
	})
	db.Table("dcim_interface").Create(map[string]interface{}{
		"id": 200, "name": "eth1", "device_id": 20, "_path_id": 0,
	})
	// Cable.
	db.Table("dcim_cable").Create(map[string]interface{}{
		"id": 300, "label": "patch-A-B", "length": 2, "length_unit": "m",
		"color": "aa0000", "status": "connected",
	})
	// CablePath: [[1:100],[2:300],[1:200]]
	path := `[["1:100"],["2:300"],["1:200"]]`
	db.Table("dcim_cablepath").Create(map[string]interface{}{
		"id": 500, "path": path, "_nodes": "1:100,2:300,1:200",
		"is_active": 1, "is_complete": 1, "is_split": 0,
	})
}

func TestTrace_SingleHop(t *testing.T) {
	db := setupCustomTestDB(t)
	seedTraceTopology(t, db)

	gin.SetMode(gin.TestMode)
	svc := dcim.NewTraceService()
	r := gin.New()
	r.GET("/api/dcim/interfaces/:id/trace", func(c *gin.Context) {
		result, err := svc.Trace(c.Request.Context(), database.GetDB(), "interfaces", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	w := doReq(r, http.MethodGet, "/api/dcim/interfaces/100/trace", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	segs, ok := resp["segments"].([]interface{})
	if !ok {
		t.Fatalf("expected segments array, got %T", resp["segments"])
	}
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	seg := segs[0].(map[string]interface{})
	if seg["origin_type"] != "dcim.interface" {
		t.Errorf("origin_type = %v, want dcim.interface", seg["origin_type"])
	}
	if seg["destination_type"] != "dcim.interface" {
		t.Errorf("destination_type = %v, want dcim.interface", seg["destination_type"])
	}
	if resp["total_length"].(float64) != 2 {
		t.Errorf("total_length = %v, want 2", resp["total_length"])
	}
}

func TestTrace_NoPath(t *testing.T) {
	db := setupCustomTestDB(t)
	seedTraceTopology(t, db)

	gin.SetMode(gin.TestMode)
	svc := dcim.NewTraceService()
	r := gin.New()
	r.GET("/api/dcim/interfaces/:id/trace", func(c *gin.Context) {
		result, err := svc.Trace(c.Request.Context(), database.GetDB(), "interfaces", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// interface 200 has _path_id = 0 (no cable) → empty segments.
	w := doReq(r, http.MethodGet, "/api/dcim/interfaces/200/trace", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	segs, _ := resp["segments"].([]interface{})
	if len(segs) != 0 {
		t.Errorf("expected 0 segments for uncabled interface, got %d", len(segs))
	}
}

func TestTrace_NotFound(t *testing.T) {
	setupCustomTestDB(t)
	gin.SetMode(gin.TestMode)
	svc := dcim.NewTraceService()
	r := gin.New()
	r.GET("/api/dcim/interfaces/:id/trace", func(c *gin.Context) {
		_, err := svc.Trace(c.Request.Context(), database.GetDB(), "interfaces", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nil)
	})

	w := doReq(r, http.MethodGet, "/api/dcim/interfaces/9999/trace", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestPaths_ReturnsTraversingPaths(t *testing.T) {
	db := setupCustomTestDB(t)
	seedTraceTopology(t, db)

	gin.SetMode(gin.TestMode)
	svc := dcim.NewTraceService()
	r := gin.New()
	r.GET("/api/dcim/front-ports/:id/paths", func(c *gin.Context) {
		// Reuse the seeded interface content type (id=1) to simulate a
		// front-port node in the path.
		result, err := svc.Paths(c.Request.Context(), database.GetDB(), "interfaces", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// interface 100 is node "1:100" which is in _nodes → path 500 should match.
	w := doReq(r, http.MethodGet, "/api/dcim/front-ports/100/paths", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 traversing path, got %d", len(resp))
	}
}

// Guard against unused import if the build flags exclude sql.DB.
var _ = sql.ErrNoRows
var _ = fmt.Sprintf
var _ = strings.TrimSpace
