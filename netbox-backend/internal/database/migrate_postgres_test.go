package database

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestAutoMigrateFreshPostgres proves the actual startup registry, rather than
// an adapter-only subset, can initialize a clean PostgreSQL namespace and run
// again without changing existing tables. It is enabled by the owned V3 job.
func TestAutoMigrateFreshPostgres(t *testing.T) {
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	schema := fmt.Sprintf("netbox_bootstrap_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() { _ = admin.Exec("DROP SCHEMA " + schema + " CASCADE").Error })

	scopedDSN := postgresTestDSN(t, dsn, schema)
	db, err := gorm.Open(postgres.Open(scopedDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("open scoped PostgreSQL: %v", err)
	}
	if err := AutoMigrateWithDB(db); err != nil {
		t.Fatalf("fresh startup bootstrap: %v", err)
	}
	if err := AutoMigrateWithDB(db); err != nil {
		t.Fatalf("idempotent startup bootstrap: %v", err)
	}
	for _, table := range []string{"go_identity_users", "go_identity_groups", "go_identity_permission_grants", "go_identity_group_memberships", "go_identity_user_permission_grants", "go_identity_group_permission_grants", "go_dcim_sites", "go_ipam_ip_addresses", "go_object_changes"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("startup registry did not create %s", table)
		}
	}
}

func postgresTestDSN(t *testing.T, dsn, schema string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err == nil && (parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	if strings.Contains(dsn, "=") {
		return strings.TrimSpace(dsn) + " search_path=" + schema
	}
	t.Fatalf("NETBOX_TEST_POSTGRES_DSN must be a postgres URL or keyword DSN")
	return ""
}
