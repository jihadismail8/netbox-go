package dcim

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/contenttype"
	"netbox-go/internal/database"
)

// seedCTs holds the resolved content-type IDs for the current test run,
// populated by seedInterfaceTopology. Lets later seed helpers reference the
// right termination_type_id without hardcoding numbers.
var seedCTs map[string]uint64

// setupConnectionsTestDB builds an in-memory SQLite DB with the tables the
// connection-list handler joins across, and injects it into the database pkg.
func setupConnectionsTestDB(t *testing.T) *gorm.DB {
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
		`CREATE TABLE IF NOT EXISTS dcim_site (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_device (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			site_id INTEGER DEFAULT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_interface (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			device_id INTEGER DEFAULT NULL,
			_path_id INTEGER DEFAULT 0,
			_path_is_active INTEGER DEFAULT 1,
			cable_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_consoleport (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			device_id INTEGER DEFAULT NULL,
			_path_id INTEGER DEFAULT 0,
			_path_is_active INTEGER DEFAULT 1,
			cable_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_consoleserverport (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			device_id INTEGER DEFAULT NULL,
			_path_id INTEGER DEFAULT 0,
			cable_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_powerport (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			device_id INTEGER DEFAULT NULL,
			_path_id INTEGER DEFAULT 0,
			_path_is_active INTEGER DEFAULT 0,
			cable_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_cable (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			label TEXT DEFAULT '',
			status TEXT DEFAULT 'connected',
			length REAL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS dcim_cabletermination (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cable_id INTEGER NOT NULL,
			cable_end TEXT NOT NULL,
			termination_id INTEGER NOT NULL,
			termination_type_id INTEGER NOT NULL,
			_device_id INTEGER DEFAULT NULL,
			_site_id INTEGER DEFAULT NULL
		)`,
	}
	for _, ddl := range ddls {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}

	// Reset the contenttype resolver cache so each test re-populates from the
	// freshly-seeded django_content_type table.
	contenttype.InvalidateCache()

	database.SetDB(db)
	for _, tbl := range []string{
		"django_content_type", "dcim_site", "dcim_device", "dcim_interface",
		"dcim_consoleport", "dcim_consoleserverport", "dcim_powerport",
		"dcim_cable", "dcim_cabletermination",
	} {
		db.Exec("DELETE FROM " + tbl)
		db.Exec("DELETE FROM sqlite_sequence WHERE name = '" + tbl + "'")
	}
	return db
}

// seedInterfaceTopology wires two devices, two interfaces, a cable, and the
// two cable-termination rows linking them. interface 100 on dev-A is the
// near end (has _path_id set); interface 200 on dev-B is the far end.
//
// Content types are seeded via contenttype.Seed (the real startup path) and
// their IDs resolved through the resolver, so the termination_type_id values
// reference whatever IDs the DB actually allocated.
func seedInterfaceTopology(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Seed all content types (idempotent) and reset the resolver cache so it
	// picks up the freshly-seeded rows.
	if _, err := contenttype.Seed(db); err != nil {
		t.Fatalf("seed content types: %v", err)
	}
	contenttype.InvalidateCache()

	// Resolve the content-type IDs we need for the topology.
	interfaceCT, _ := contenttype.LookupByLabel(db, "dcim", "interface")
	cableCT, _ := contenttype.LookupByLabel(db, "dcim", "cable")
	consoleportCT, _ := contenttype.LookupByLabel(db, "dcim", "consoleport")
	consoleserverportCT, _ := contenttype.LookupByLabel(db, "dcim", "consoleserverport")
	powerportCT, _ := contenttype.LookupByLabel(db, "dcim", "powerport")
	poweroutletCT, _ := contenttype.LookupByLabel(db, "dcim", "poweroutlet")

	// Stash them on the test struct via package vars for later seed functions.
	seedCTs = map[string]uint64{
		"interface":         interfaceCT,
		"cable":             cableCT,
		"consoleport":       consoleportCT,
		"consoleserverport": consoleserverportCT,
		"powerport":         powerportCT,
		"poweroutlet":       poweroutletCT,
	}

	// Sites + devices.
	db.Table("dcim_site").Create(map[string]interface{}{"name": "Site A", "slug": "site-a"})
	db.Table("dcim_site").Create(map[string]interface{}{"name": "Site B", "slug": "site-b"})
	db.Table("dcim_device").Create(map[string]interface{}{"id": 10, "name": "dev-A", "site_id": 1})
	db.Table("dcim_device").Create(map[string]interface{}{"id": 20, "name": "dev-B", "site_id": 2})
	// Interfaces: 100 (dev-A, connected, active), 200 (dev-B, far end).
	db.Table("dcim_interface").Create(map[string]interface{}{
		"id": 100, "name": "eth0", "device_id": 10, "_path_id": 500,
		"_path_is_active": 1, "cable_id": 8,
	})
	db.Table("dcim_interface").Create(map[string]interface{}{
		"id": 200, "name": "eth0", "device_id": 20, "_path_id": 0,
		"cable_id": 8,
	})
	// Cable 8 "patch-A-B".
	db.Table("dcim_cable").Create(map[string]interface{}{
		"id": 8, "label": "patch-A-B", "status": "connected", "length": 2,
	})
	// Two terminations of cable 8: end A = interface 100 (dev-A),
	// end B = interface 200 (dev-B).
	db.Table("dcim_cabletermination").Create(map[string]interface{}{
		"cable_id": 8, "cable_end": "A", "termination_id": 100,
		"termination_type_id": interfaceCT, "_device_id": 10, "_site_id": 1,
	})
	db.Table("dcim_cabletermination").Create(map[string]interface{}{
		"cable_id": 8, "cable_end": "B", "termination_id": 200,
		"termination_type_id": interfaceCT, "_device_id": 20, "_site_id": 2,
	})
}

func doConnReq(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestConnections_InterfaceList(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Only the connected interface (100) should appear; interface 200 has no
	// _path_id and is excluded.
	results := resp["results"].([]interface{})
	if len(results) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(results))
	}
	row := results[0].(map[string]interface{})
	if row["name"] != "eth0" {
		t.Errorf("name = %v, want eth0", row["name"])
	}
	device := row["device"].(map[string]interface{})
	if device["name"] != "dev-A" {
		t.Errorf("device.name = %v, want dev-A", device["name"])
	}
	cable := row["cable"].(map[string]interface{})
	if cable["label"] != "patch-A-B" {
		t.Errorf("cable.label = %v, want patch-A-B", cable["label"])
	}
	dest := row["destination"].(map[string]interface{})
	if dest["name"] != "eth0" {
		t.Errorf("destination.name = %v, want eth0", dest["name"])
	}
	destDevice := dest["device"].(map[string]interface{})
	if destDevice["name"] != "dev-B" {
		t.Errorf("destination.device.name = %v, want dev-B", destDevice["name"])
	}
	if row["destination_type"] != "dcim.interface" {
		t.Errorf("destination_type = %v, want dcim.interface", row["destination_type"])
	}
	if row["reachable"] != true {
		t.Errorf("reachable = %v, want true", row["reachable"])
	}
}

func TestConnections_FilterByDeviceID(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	// dev-A (id=10) has the connected interface → 1 result.
	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?device_id=10")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("count = %v, want 1", resp["count"])
	}

	// dev-B (id=20) has the far-end interface (no _path_id) → 0 results.
	w = doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?device_id=20")
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 0 {
		t.Errorf("count for dev-B = %v, want 0 (no connected interfaces)", resp["count"])
	}
}

func TestConnections_FilterByDeviceName(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?device=dev-A")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("count for device=dev-A = %v, want 1", resp["count"])
	}
}

func TestConnections_FilterBySiteID(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	// site 1 (Site A) hosts dev-A (the connected side).
	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?site_id=1")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("count for site_id=1 = %v, want 1", resp["count"])
	}

	// site 2 (Site B) hosts dev-B (no connected interface) → 0.
	w = doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?site_id=2")
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 0 {
		t.Errorf("count for site_id=2 = %v, want 0", resp["count"])
	}
}

func TestConnections_Search(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	// Search by cable label fragment.
	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?q=patch")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("count for q=patch = %v, want 1", resp["count"])
	}

	// Search by device name.
	w = doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?q=dev-a")
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("count for q=dev-a = %v, want 1", resp["count"])
	}

	// No match.
	w = doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?q=nope")
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 0 {
		t.Errorf("count for q=nope = %v, want 0", resp["count"])
	}
}

func TestConnections_UnconnectedExcluded(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)
	// Add an uncabled interface (no _path_id).
	db.Table("dcim_interface").Create(map[string]interface{}{
		"id": 300, "name": "eth1", "device_id": 10, "_path_id": 0, "cable_id": 0,
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	// Only interface 100 is connected; 300 is excluded.
	if resp["count"].(float64) != 1 {
		t.Errorf("count = %v, want 1 (uncabled interface should be excluded)", resp["count"])
	}
}

func TestConnections_Pagination(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/interface-connections", MakeConnectionsHandler(ConnectionTypes["interface-connections"]))

	w := doConnReq(r, http.MethodGet, "/api/dcim/interface-connections?limit=1&offset=0")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("count = %v, want 1", resp["count"])
	}
	// With 1 total and limit=1 offset=0, next should be nil (no more pages).
	if resp["next"] != nil {
		t.Errorf("next = %v, want nil (single result fits in one page)", resp["next"])
	}
}

func TestConnections_ConsoleAndPower(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)
	// Add a console-port connection: consoleport 110 (dev-A) <-> consoleserverport 210 (dev-B).
	db.Table("dcim_consoleport").Create(map[string]interface{}{
		"id": 110, "name": "Console", "device_id": 10, "_path_id": 510,
		"_path_is_active": 1, "cable_id": 9,
	})
	db.Table("dcim_consoleserverport").Create(map[string]interface{}{
		"id": 210, "name": "Port1", "device_id": 20, "_path_id": 0, "cable_id": 9,
	})
	db.Table("dcim_cable").Create(map[string]interface{}{
		"id": 9, "label": "console-cable", "status": "connected",
	})
	db.Table("dcim_cabletermination").Create(map[string]interface{}{
		"cable_id": 9, "cable_end": "A", "termination_id": 110,
		"termination_type_id": seedCTs["consoleport"], "_device_id": 10,
	})
	db.Table("dcim_cabletermination").Create(map[string]interface{}{
		"cable_id": 9, "cable_end": "B", "termination_id": 210,
		"termination_type_id": seedCTs["consoleserverport"], "_device_id": 20,
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/console-connections", MakeConnectionsHandler(ConnectionTypes["console-connections"]))

	w := doConnReq(r, http.MethodGet, "/api/dcim/console-connections")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	results := resp["results"].([]interface{})
	if len(results) != 1 {
		t.Fatalf("expected 1 console connection, got %d", len(results))
	}
	row := results[0].(map[string]interface{})
	if row["name"] != "Console" {
		t.Errorf("name = %v, want Console", row["name"])
	}
	if row["destination_type"] != "dcim.consoleserverport" {
		t.Errorf("destination_type = %v, want dcim.consoleserverport", row["destination_type"])
	}

	// Power connections: none seeded → empty.
	r2 := gin.New()
	r2.GET("/api/dcim/power-connections", MakeConnectionsHandler(ConnectionTypes["power-connections"]))
	w = doConnReq(r2, http.MethodGet, "/api/dcim/power-connections")
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 0 {
		t.Errorf("power connections count = %v, want 0", resp["count"])
	}
}

func TestConnections_NotReachableFlag(t *testing.T) {
	db := setupConnectionsTestDB(t)
	seedInterfaceTopology(t, db)
	// Add a power-port connection that is NOT active (_path_is_active=0).
	db.Table("dcim_powerport").Create(map[string]interface{}{
		"id": 120, "name": "PSU0", "device_id": 10, "_path_id": 520,
		"_path_is_active": 0, "cable_id": 10,
	})
	db.Table("dcim_cable").Create(map[string]interface{}{
		"id": 10, "label": "power-cable", "status": "decommissioned",
	})
	db.Table("dcim_cabletermination").Create(map[string]interface{}{
		"cable_id": 10, "cable_end": "A", "termination_id": 120,
		"termination_type_id": seedCTs["powerport"], "_device_id": 10,
	})
	db.Table("dcim_cabletermination").Create(map[string]interface{}{
		"cable_id": 10, "cable_end": "B", "termination_id": 999,
		"termination_type_id": seedCTs["poweroutlet"], "_device_id": 20, // far end, nonexistent row
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dcim/power-connections", MakeConnectionsHandler(ConnectionTypes["power-connections"]))

	w := doConnReq(r, http.MethodGet, "/api/dcim/power-connections")
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	results := resp["results"].([]interface{})
	if len(results) != 1 {
		t.Fatalf("expected 1 power connection, got %d", len(results))
	}
	row := results[0].(map[string]interface{})
	if row["reachable"] != false {
		t.Errorf("reachable = %v, want false (path inactive)", row["reachable"])
	}
}
