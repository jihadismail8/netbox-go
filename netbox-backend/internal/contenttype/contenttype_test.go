package contenttype

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite DB with the django_content_type
// table and a unique index on (app_label, model) — required for the
// ON CONFLICT clause the seed relies on.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS django_content_type (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_label TEXT NOT NULL,
		model TEXT NOT NULL
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	// Unique index on the natural key so ON CONFLICT works.
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_django_ct_label ON django_content_type (app_label, model)`)

	// Clean slate for test isolation (shared in-memory DB).
	db.Exec("DELETE FROM django_content_type")
	db.Exec("DELETE FROM sqlite_sequence WHERE name = 'django_content_type'")
	InvalidateCache()
	return db
}

func TestSeed_CreatesAllTypes(t *testing.T) {
	db := setupTestDB(t)
	count, err := Seed(db)
	if err != nil {
		t.Fatalf("Seed: %v", err)
	}
	expected := int64(len(allTypes))
	if count != expected {
		t.Errorf("count after seed = %d, want %d", count, expected)
	}
}

func TestSeed_IsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("first Seed: %v", err)
	}
	// Second seed must not duplicate rows.
	count2, err := Seed(db)
	if err != nil {
		t.Fatalf("second Seed: %v", err)
	}
	expected := int64(len(allTypes))
	if count2 != expected {
		t.Errorf("count after second seed = %d, want %d (no duplicates)", count2, expected)
	}
}

func TestSeed_CreatesKnownTypes(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	// Spot-check a representative type from each app.
	cases := []struct{ appLabel, model, table string }{
		{"dcim", "interface", "dcim_interface"},
		{"dcim", "cable", "dcim_cable"},
		{"ipam", "prefix", "ipam_prefix"},
		{"ipam", "ipaddress", "ipam_ipaddress"},
		{"circuits", "circuit", "circuits_circuit"},
		{"extras", "customfield", "extras_customfield"},
		{"users", "user", "users_user"},
		{"core", "objectchange", "core_objectchange"},
		{"wireless", "wirelesslan", "wireless_wirelesslan"},
		{"vpn", "tunnel", "vpn_tunnel"},
	}
	for _, c := range cases {
		var id uint64
		err := db.Table("django_content_type").
			Where("app_label = ? AND model = ?", c.appLabel, c.model).
			Select("id").Scan(&id).Error
		if err != nil || id == 0 {
			t.Errorf("expected content type for %s.%s to exist after seed", c.appLabel, c.model)
			continue
		}
	}
}

func TestSeed_ExcludesJunctionTables(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	// These junction tables have no Python model class, so they must NOT be
	// seeded as content types.
	junctionTables := []string{
		"dcim_interface_tagged_vlans",
		"extras_configcontext_sites",
		"ipam_vrf_import_targets",
	}
	for _, tbl := range junctionTables {
		for _, typ := range allTypes {
			if typ.Table == tbl {
				t.Errorf("junction table %q should not be in the seed list", tbl)
			}
		}
	}
	// Django infrastructure models that Python's contenttype framework DOES
	// create, but we intentionally don't seed (we're matching NetBox business
	// models only; auth/django infra is out of scope for GenericFK targets).
	notSeeded := []struct{ appLabel, model string }{
		{"auth", "group"},
		{"django", "session"},
		{"taggit", "tag"},
	}
	for _, ns := range notSeeded {
		var count int64
		db.Table("django_content_type").
			Where("app_label = ? AND model = ?", ns.appLabel, ns.model).
			Count(&count)
		if count != 0 {
			t.Errorf("infrastructure %s.%s should not be seeded, found %d row(s)", ns.appLabel, ns.model, count)
		}
	}
}

func TestResolve_ByKnownID(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	InvalidateCache() // force resolver to load from DB

	// Look up dcim.interface by label, then resolve back by id.
	id, ok := LookupByLabel(db, "dcim", "interface")
	if !ok {
		t.Fatal("LookupByLabel(dcim.interface) returned false")
	}
	typ, ok := Resolve(db, id)
	if !ok {
		t.Fatalf("Resolve(%d) returned false", id)
	}
	if typ.AppLabel != "dcim" || typ.Model != "interface" {
		t.Errorf("Resolve = %+v, want {dcim interface dcim_interface}", typ)
	}
	if typ.Table != "dcim_interface" {
		t.Errorf("Table = %q, want dcim_interface", typ.Table)
	}
}

func TestResolve_UnknownID(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	InvalidateCache()

	_, ok := Resolve(db, 999999)
	if ok {
		t.Error("Resolve(999999) should return false for unknown id")
	}
}

func TestLookupByLabel_Unknown(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	InvalidateCache()

	_, ok := LookupByLabel(db, "nonexistent", "model")
	if ok {
		t.Error("LookupByLabel for unknown label should return false")
	}
}

func TestResolve_CacheIsReused(t *testing.T) {
	db := setupTestDB(t)
	if _, err := Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	InvalidateCache()

	// First call loads from DB.
	_, _ = Resolve(db, 1)
	// Drop all rows — second call must still resolve (cache hit), proving it
	// didn't re-query.
	db.Exec("DELETE FROM django_content_type")
	id, ok := LookupByLabel(db, "dcim", "interface")
	if !ok {
		t.Error("expected cached lookup to succeed even after DB was cleared")
	}
	if id == 0 {
		t.Error("expected non-zero id from cache")
	}
}

func TestAll_ReturnsStaticList(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("All() returned empty list")
	}
	// Every entry must have all three fields populated.
	for _, t2 := range all {
		if t2.AppLabel == "" || t2.Model == "" || t2.Table == "" {
			t.Errorf("incomplete entry: %+v", t2)
		}
	}
}

func TestSeed_NilDB(t *testing.T) {
	_, err := Seed(nil)
	if err == nil {
		t.Error("Seed(nil) should return an error")
	}
}
