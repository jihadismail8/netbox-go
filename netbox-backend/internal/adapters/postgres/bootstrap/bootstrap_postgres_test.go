package bootstrap

import (
	"context"
	"fmt"
	"os"
	"reflect"
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
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}

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
		_ = transaction.Rollback().Error
	})

	schema := fmt.Sprintf("bootstrap_test_%d", time.Now().UnixNano())
	if err := transaction.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	if err := transaction.Exec(`SET LOCAL search_path TO "` + schema + `"`).Error; err != nil {
		t.Fatalf("set isolated search path: %v", err)
	}
	if err := transaction.Exec(
		`CREATE TABLE bootstrap_postgres_existing (legacy_value bigint NOT NULL)`,
	).Error; err != nil {
		t.Fatalf("create mismatched existing table: %v", err)
	}
	before := postgresTableColumns(t, transaction, schema, "bootstrap_postgres_existing")

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
	after := postgresTableColumns(t, transaction, schema, "bootstrap_postgres_existing")
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("existing PostgreSQL table changed: before=%v after=%v", before, after)
	}
	if !transaction.Migrator().HasTable(&postgresMissingModel{}) {
		t.Fatal("missing PostgreSQL table was not created")
	}
}

func postgresTableColumns(t *testing.T, db *gorm.DB, schema, table string) []string {
	t.Helper()
	type column struct {
		Name     string `gorm:"column:column_name"`
		DataType string `gorm:"column:data_type"`
	}
	var columns []column
	if err := db.Raw(
		`SELECT column_name, data_type
		 FROM information_schema.columns
		 WHERE table_schema = ? AND table_name = ?
		 ORDER BY ordinal_position`,
		schema,
		table,
	).Scan(&columns).Error; err != nil {
		t.Fatalf("read PostgreSQL table columns: %v", err)
	}
	result := make([]string, 0, len(columns))
	for _, column := range columns {
		result = append(result, column.Name+":"+column.DataType)
	}
	return result
}
