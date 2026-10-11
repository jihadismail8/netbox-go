package bootstrap

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type postgresExistingModel struct {
	ID       int64 `gorm:"primaryKey"`
	Required string
}

func (postgresExistingModel) TableName() string { return "bootstrap_postgres_existing" }

type postgresMissingModel struct {
	ID int64 `gorm:"primaryKey"`
}

func (postgresMissingModel) TableName() string { return "bootstrap_postgres_missing" }

func TestRunPostgresLeavesMismatchedExistingTableUntouched(t *testing.T) {
	dsn := requiredBootstrapPostgresDSN(t)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	transaction := db.Begin()
	if transaction.Error != nil {
		t.Fatalf("begin PostgreSQL transaction: %v", transaction.Error)
	}
	t.Cleanup(func() {
		if err := transaction.Rollback().Error; err != nil {
			t.Errorf("roll back isolated PostgreSQL transaction: %v", err)
		}
	})

	schema := fmt.Sprintf("bootstrap_test_%d", time.Now().UnixNano())
	if err := transaction.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	if err := transaction.Exec(`SET LOCAL search_path TO "` + schema + `"`).Error; err != nil {
		t.Fatalf("set isolated search path: %v", err)
	}
	if err := transaction.Exec(
		`CREATE TABLE bootstrap_postgres_parent (id bigint PRIMARY KEY)`,
	).Error; err != nil {
		t.Fatalf("create malformed-table dependency: %v", err)
	}
	if err := transaction.Exec(
		`CREATE TABLE bootstrap_postgres_existing (
			legacy_value bigint NOT NULL,
			parent_id bigint,
			CONSTRAINT bootstrap_postgres_existing_legacy_positive CHECK (legacy_value > 0),
			CONSTRAINT bootstrap_postgres_existing_legacy_key UNIQUE (legacy_value),
			CONSTRAINT bootstrap_postgres_existing_parent_fk
				FOREIGN KEY (parent_id) REFERENCES bootstrap_postgres_parent (id)
				ON UPDATE CASCADE ON DELETE SET NULL
		)`,
	).Error; err != nil {
		t.Fatalf("create mismatched existing table: %v", err)
	}
	if err := transaction.Exec(
		`CREATE INDEX bootstrap_postgres_existing_parent_partial
		 ON bootstrap_postgres_existing (parent_id)
		 WHERE parent_id IS NOT NULL`,
	).Error; err != nil {
		t.Fatalf("create mismatched existing-table index: %v", err)
	}
	before := postgresTableCatalogFingerprint(t, transaction, schema, "bootstrap_postgres_existing")
	requirePostgresCatalogItem(t, before, "column|bootstrap_postgres_existing|1|legacy_value|bigint|t")
	requirePostgresCatalogItem(t, before, "constraint|bootstrap_postgres_existing|bootstrap_postgres_existing_legacy_positive|c|")
	requirePostgresCatalogItem(t, before, "constraint|bootstrap_postgres_existing|bootstrap_postgres_existing_parent_fk|f|")
	requirePostgresCatalogItem(t, before, "ON UPDATE CASCADE ON DELETE SET NULL")
	requirePostgresCatalogItem(t, before, "index|bootstrap_postgres_existing|bootstrap_postgres_existing_parent_partial|")

	registry, err := NewRegistry(
		Entry{Name: "existing", Model: &postgresExistingModel{}},
		Entry{Name: "missing", Model: &postgresMissingModel{}, Dependencies: []string{"existing"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	result, err := Run(context.Background(), transaction, registry)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !reflect.DeepEqual(result.Existing, []string{"existing"}) {
		t.Fatalf("existing entries = %v, want [existing]", result.Existing)
	}
	if !reflect.DeepEqual(result.Created, []string{"missing"}) {
		t.Fatalf("created entries = %v, want [missing]", result.Created)
	}
	after := postgresTableCatalogFingerprint(t, transaction, schema, "bootstrap_postgres_existing")
	if !reflect.DeepEqual(after, before) {
		t.Fatal("mismatched existing PostgreSQL columns, constraints, indexes, or foreign-key semantics changed")
	}
	if !transaction.Migrator().HasTable(&postgresMissingModel{}) {
		t.Fatal("missing PostgreSQL table was not created")
	}
}

func requiredBootstrapPostgresDSN(t *testing.T) string {
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

func postgresTableCatalogFingerprint(t *testing.T, db *gorm.DB, schema, table string) []string {
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
			WHERE n.nspname = ? AND c.relname = ? AND c.relkind IN ('r', 'p')
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
		)
		SELECT item FROM catalog ORDER BY item`,
		schema,
		table,
	).Scan(&rows).Error; err != nil {
		t.Fatalf("read PostgreSQL table catalog: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("PostgreSQL catalog fingerprint is empty for %s.%s", schema, table)
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Item)
	}
	return items
}

func requirePostgresCatalogItem(t *testing.T, items []string, substring string) {
	t.Helper()
	for _, item := range items {
		if strings.Contains(item, substring) {
			return
		}
	}
	t.Fatalf("PostgreSQL catalog fingerprint has no item containing %q", substring)
}
