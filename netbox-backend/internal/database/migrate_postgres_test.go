package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	"netbox-go/internal/contenttype"
)

// TestAutoMigrateFreshPostgres proves the actual startup registry, rather than
// an adapter-only subset, initializes exactly one clean PostgreSQL namespace
// and leaves its complete semantic catalog unchanged on a second run.
func TestAutoMigrateFreshPostgres(t *testing.T) {
	db, schema := newDatabasePostgresSchema(t, "netbox_bootstrap")
	registry, expectedEntries, expectedTables := startupRegistryContract(t, db)

	if tables := postgresSchemaTables(t, db, schema); len(tables) != 0 {
		t.Fatalf("fresh PostgreSQL schema tables = %v, want none", tables)
	}

	first, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("fresh startup bootstrap: %v", err)
	}
	if !reflect.DeepEqual(first.Created, expectedEntries) {
		t.Fatalf("fresh created order differs from the 198-entry registry")
	}
	if len(first.Existing) != 0 {
		t.Fatalf("fresh existing entries = %v, want none", first.Existing)
	}
	if tables := postgresSchemaTables(t, db, schema); !reflect.DeepEqual(tables, expectedTables) {
		t.Fatalf("fresh physical table set differs from the 198-entry registry")
	}

	before := postgresCatalogFingerprint(t, db, schema, "")
	second, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("idempotent startup bootstrap: %v", err)
	}
	if len(second.Created) != 0 {
		t.Fatalf("second-run created entries = %v, want none", second.Created)
	}
	if !reflect.DeepEqual(second.Existing, expectedEntries) {
		t.Fatalf("second-run existing order differs from the 198-entry registry")
	}
	if tables := postgresSchemaTables(t, db, schema); !reflect.DeepEqual(tables, expectedTables) {
		t.Fatalf("second-run physical table set differs from the 198-entry registry")
	}
	if after := postgresCatalogFingerprint(t, db, schema, ""); !reflect.DeepEqual(after, before) {
		t.Fatal("whole-registry PostgreSQL catalog changed on the second bootstrap run")
	}

	startupDB, startupSchema := newDatabasePostgresSchema(t, "netbox_startup_boundary")
	if tables := postgresSchemaTables(t, startupDB, startupSchema); len(tables) != 0 {
		t.Fatalf("fresh startup-boundary schema tables = %v, want none", tables)
	}
	if err := AutoMigrateWithDB(startupDB); err != nil {
		t.Fatalf("AutoMigrateWithDB startup boundary: %v", err)
	}
	if tables := postgresSchemaTables(t, startupDB, startupSchema); !reflect.DeepEqual(tables, expectedTables) {
		t.Fatal("AutoMigrateWithDB did not create the exact 198-table startup registry")
	}
}

func TestAutoMigratePostgresLeavesCorrectExistingTableUntouched(t *testing.T) {
	db, schema := newDatabasePostgresSchema(t, "netbox_correct_existing")
	entry := startupRegistryEntryForTable(t, db, "django_content_type")
	registry, err := bootstrap.NewRegistry(entry)
	if err != nil {
		t.Fatalf("single-entry registry: %v", err)
	}
	setup, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("create correct existing table: %v", err)
	}
	if !reflect.DeepEqual(setup.Created, []string{entry.Name}) {
		t.Fatalf("setup created entries = %v, want [%s]", setup.Created, entry.Name)
	}

	before := postgresCatalogFingerprint(t, db, schema, "django_content_type")
	result, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("bootstrap correct existing table: %v", err)
	}
	if len(result.Created) != 0 || !reflect.DeepEqual(result.Existing, []string{entry.Name}) {
		t.Fatalf("result = created %v existing %v, want created [] existing [%s]", result.Created, result.Existing, entry.Name)
	}
	if tables := postgresSchemaTables(t, db, schema); !reflect.DeepEqual(tables, []string{"django_content_type"}) {
		t.Fatalf("correct-existing schema tables = %v, want [django_content_type]", tables)
	}
	if after := postgresCatalogFingerprint(t, db, schema, "django_content_type"); !reflect.DeepEqual(after, before) {
		t.Fatal("correct existing registry table catalog changed during bootstrap")
	}
}

func TestAutoMigratePostgresCreatesMissingTableInNonEmptySchema(t *testing.T) {
	db, schema := newDatabasePostgresSchema(t, "netbox_nonempty")
	if err := db.Exec(`
		CREATE TABLE bootstrap_sentinel (
			id bigint PRIMARY KEY,
			payload text NOT NULL,
			CONSTRAINT bootstrap_sentinel_payload_nonempty CHECK (length(payload) > 0)
		)`,
	).Error; err != nil {
		t.Fatalf("create sentinel table: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX bootstrap_sentinel_payload_key ON bootstrap_sentinel (payload)`).Error; err != nil {
		t.Fatalf("create sentinel index: %v", err)
	}
	if err := db.Exec(`INSERT INTO bootstrap_sentinel (id, payload) VALUES (1, 'owned sentinel')`).Error; err != nil {
		t.Fatalf("insert sentinel row: %v", err)
	}
	before := postgresCatalogFingerprint(t, db, schema, "bootstrap_sentinel")

	entry := startupRegistryEntryForTable(t, db, "go_identity_groups")
	registry, err := bootstrap.NewRegistry(entry)
	if err != nil {
		t.Fatalf("single-entry registry: %v", err)
	}
	result, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("bootstrap missing table in non-empty schema: %v", err)
	}
	if !reflect.DeepEqual(result.Created, []string{entry.Name}) || len(result.Existing) != 0 {
		t.Fatalf("result = created %v existing %v, want created [%s] existing []", result.Created, result.Existing, entry.Name)
	}
	if !db.Migrator().HasTable(entry.Model) {
		t.Fatalf("missing registry table %s was not created", entry.Name)
	}
	assertModelColumnsExist(t, db, entry.Model)
	if tables := postgresSchemaTables(t, db, schema); !reflect.DeepEqual(tables, []string{"bootstrap_sentinel", "go_identity_groups"}) {
		t.Fatalf("non-empty schema tables = %v, want [bootstrap_sentinel go_identity_groups]", tables)
	}
	if after := postgresCatalogFingerprint(t, db, schema, "bootstrap_sentinel"); !reflect.DeepEqual(after, before) {
		t.Fatal("unrelated sentinel table catalog changed while creating a missing registry table")
	}
	var payload string
	if err := db.Raw(`SELECT payload FROM bootstrap_sentinel WHERE id = 1`).Scan(&payload).Error; err != nil {
		t.Fatalf("read sentinel row: %v", err)
	}
	if payload != "owned sentinel" {
		t.Fatalf("sentinel payload = %q, want %q", payload, "owned sentinel")
	}
}

func TestContentTypeNaturalKeyUniquePostgres(t *testing.T) {
	db, _ := newDatabasePostgresSchema(t, "netbox_content_type")
	entry := startupRegistryEntryForTable(t, db, "django_content_type")
	registry, err := bootstrap.NewRegistry(entry)
	if err != nil {
		t.Fatalf("single-entry registry: %v", err)
	}
	result, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("bootstrap django_content_type: %v", err)
	}
	if !reflect.DeepEqual(result.Created, []string{entry.Name}) {
		t.Fatalf("created entries = %v, want [%s]", result.Created, entry.Name)
	}

	want := int64(len(contenttype.All()))
	first, err := contenttype.Seed(db)
	if err != nil {
		t.Fatalf("first content-type seed: %v", err)
	}
	second, err := contenttype.Seed(db)
	if err != nil {
		t.Fatalf("second content-type seed: %v", err)
	}
	if first != want || second != want {
		t.Fatalf("content-type counts = first %d second %d, want %d", first, second, want)
	}

	duplicate := contenttype.All()[0]
	if err := db.Exec(
		`INSERT INTO django_content_type (app_label, model) VALUES (?, ?)`,
		duplicate.AppLabel,
		duplicate.Model,
	).Error; err == nil {
		t.Fatalf("duplicate content-type pair %s.%s was accepted", duplicate.AppLabel, duplicate.Model)
	}
	var count int64
	if err := db.Table("django_content_type").Count(&count).Error; err != nil {
		t.Fatalf("count content types after rejected duplicate: %v", err)
	}
	if count != want {
		t.Fatalf("content-type count after rejected duplicate = %d, want %d", count, want)
	}
}

func newDatabasePostgresSchema(t *testing.T, prefix string) (*gorm.DB, string) {
	t.Helper()
	dsn := requiredDatabasePostgresDSN(t)
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatalf("get PostgreSQL admin connection: %v", err)
	}
	t.Cleanup(func() {
		if err := adminSQL.Close(); err != nil {
			t.Errorf("close PostgreSQL admin connection: %v", err)
		}
	})
	schema := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	if err := admin.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		if err := admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})

	db, err := gorm.Open(postgres.Open(postgresTestDSN(t, dsn, schema)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open scoped PostgreSQL: %v", err)
	}
	scopedSQL, err := db.DB()
	if err != nil {
		t.Fatalf("get scoped PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() {
		if err := scopedSQL.Close(); err != nil {
			t.Errorf("close scoped PostgreSQL connection: %v", err)
		}
	})
	return db, schema
}

func requiredDatabasePostgresDSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	if os.Getenv("NETBOX_REQUIRE_TEST_POSTGRES") == "1" {
		t.Fatal("NETBOX_TEST_POSTGRES_DSN is required when NETBOX_REQUIRE_TEST_POSTGRES=1")
	}
	t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	return ""
}

func startupRegistryContract(t *testing.T, db *gorm.DB) (bootstrap.Registry, []string, []string) {
	t.Helper()
	registry, err := modelRegistry()
	if err != nil {
		t.Fatalf("modelRegistry: %v", err)
	}
	entries := registry.Entries()
	if len(entries) != 198 {
		t.Fatalf("startup registry entries = %d, want 198", len(entries))
	}
	names := make([]string, 0, len(entries))
	tables := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name)
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(entry.Model); err != nil {
			t.Fatalf("resolve registry entry %s: %v", entry.Name, err)
		}
		tables = append(tables, statement.Schema.Table)
	}
	sort.Strings(tables)
	return registry, names, tables
}

func startupRegistryEntryForTable(t *testing.T, db *gorm.DB, table string) bootstrap.Entry {
	t.Helper()
	registry, err := modelRegistry()
	if err != nil {
		t.Fatalf("modelRegistry: %v", err)
	}
	for _, entry := range registry.Entries() {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(entry.Model); err != nil {
			t.Fatalf("resolve registry entry %s: %v", entry.Name, err)
		}
		if statement.Schema.Table == table {
			entry.Dependencies = nil
			return entry
		}
	}
	t.Fatalf("startup registry has no table %s", table)
	return bootstrap.Entry{}
}

func assertModelColumnsExist(t *testing.T, db *gorm.DB, model any) {
	t.Helper()
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(model); err != nil {
		t.Fatalf("resolve model columns: %v", err)
	}
	want := make([]string, 0, len(statement.Schema.Fields))
	for _, field := range statement.Schema.Fields {
		if field.DBName != "" {
			want = append(want, field.DBName)
		}
	}
	sort.Strings(want)

	var got []string
	if err := db.Raw(
		`SELECT column_name FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ? ORDER BY column_name`,
		statement.Schema.Table,
	).Scan(&got).Error; err != nil {
		t.Fatalf("read columns for %s: %v", statement.Schema.Table, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("columns for %s = %v, want %v", statement.Schema.Table, got, want)
	}
}

func postgresSchemaTables(t *testing.T, db *gorm.DB, schema string) []string {
	t.Helper()
	var tables []string
	if err := db.Raw(
		`SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = ? ORDER BY tablename`,
		schema,
	).Scan(&tables).Error; err != nil {
		t.Fatalf("read PostgreSQL tables: %v", err)
	}
	return tables
}

// postgresCatalogFingerprint captures the semantic catalog fields bootstrap is
// forbidden to alter on an existing table. An empty table selects the schema.
func postgresCatalogFingerprint(t *testing.T, db *gorm.DB, schema, table string) []string {
	t.Helper()
	type catalogItem struct {
		Item string `gorm:"column:item"`
	}
	var rows []catalogItem
	if err := db.Raw(`
		WITH target AS (
			SELECT c.oid, c.relname, c.relkind, c.relpersistence
			FROM pg_catalog.pg_class AS c
			JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
			WHERE n.nspname = ?
			  AND c.relkind IN ('r', 'p')
			  AND (? = '' OR c.relname = ?)
		), catalog AS (
			SELECT format('table|%s|%s|%s', relname, relkind, relpersistence) AS item
			FROM target
			UNION ALL
			SELECT format(
				'column|%s|%s|%s|%s|%s|%s|%s|%s',
				t.relname,
				a.attnum,
				a.attname,
				pg_catalog.format_type(a.atttypid, a.atttypmod),
				a.attnotnull,
				COALESCE(pg_catalog.pg_get_expr(d.adbin, d.adrelid), ''),
				a.attidentity,
				a.attgenerated
			) AS item
			FROM target AS t
			JOIN pg_catalog.pg_attribute AS a ON a.attrelid = t.oid
			LEFT JOIN pg_catalog.pg_attrdef AS d ON d.adrelid = t.oid AND d.adnum = a.attnum
			WHERE a.attnum > 0 AND NOT a.attisdropped
			UNION ALL
			SELECT format(
				'constraint|%s|%s|%s|%s|%s|%s|%s',
				t.relname,
				c.conname,
				c.contype,
				c.condeferrable,
				c.condeferred,
				c.convalidated,
				pg_catalog.pg_get_constraintdef(c.oid, true)
			) AS item
			FROM target AS t
			JOIN pg_catalog.pg_constraint AS c ON c.conrelid = t.oid
			UNION ALL
			SELECT format(
				'index|%s|%s|%s|%s|%s|%s|%s',
				t.relname,
				i.relname,
				x.indisprimary,
				x.indisunique,
				x.indisvalid,
				x.indisready,
				pg_catalog.pg_get_indexdef(i.oid)
			) AS item
			FROM target AS t
			JOIN pg_catalog.pg_index AS x ON x.indrelid = t.oid
			JOIN pg_catalog.pg_class AS i ON i.oid = x.indexrelid
			UNION ALL
			SELECT format(
				'sequence|%s|%s|%s',
				s.relname,
				t.relname,
				a.attname
			) AS item
			FROM target AS t
			JOIN pg_catalog.pg_depend AS d
			  ON d.refobjid = t.oid
			 AND d.deptype IN ('a', 'i')
			JOIN pg_catalog.pg_class AS s
			  ON s.oid = d.objid
			 AND s.relkind = 'S'
			JOIN pg_catalog.pg_attribute AS a
			  ON a.attrelid = t.oid
			 AND a.attnum = d.refobjsubid
		)
		SELECT item FROM catalog ORDER BY item`,
		schema,
		table,
		table,
	).Scan(&rows).Error; err != nil {
		t.Fatalf("read PostgreSQL catalog fingerprint: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("PostgreSQL catalog fingerprint is empty for schema %s table %q", schema, table)
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Item)
	}
	return items
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
