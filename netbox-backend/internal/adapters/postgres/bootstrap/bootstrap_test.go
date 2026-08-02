package bootstrap

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type existingModel struct {
	ID       uint64 `gorm:"primaryKey"`
	Required string `gorm:"not null"`
}

func (existingModel) TableName() string { return "bootstrap_existing" }

type missingModel struct {
	ID   uint64 `gorm:"primaryKey"`
	Name string
}

func (missingModel) TableName() string { return "bootstrap_missing" }

type relationshipParentModel struct {
	ID       uint64 `gorm:"primaryKey"`
	Required string `gorm:"not null"`
}

func (relationshipParentModel) TableName() string { return "bootstrap_relationship_parent" }

type relationshipChildModel struct {
	ID       uint64                   `gorm:"primaryKey"`
	ParentID uint64                   `gorm:"not null;index"`
	Parent   *relationshipParentModel `gorm:"foreignKey:ParentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (relationshipChildModel) TableName() string { return "bootstrap_relationship_child" }

type invalidIndexModel struct {
	ID    uint64 `gorm:"primaryKey"`
	Value string `gorm:"index:idx_bootstrap_invalid_expression,expression:("`
}

func (invalidIndexModel) TableName() string { return "bootstrap_invalid_index" }

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestRunCreatesOnlyMissingTables(t *testing.T) {
	db := openTestDB(t)
	if err := db.Exec(`CREATE TABLE bootstrap_existing (legacy_value INTEGER NOT NULL)`).Error; err != nil {
		t.Fatalf("create mismatched existing table: %v", err)
	}
	before := tableColumns(t, db, "bootstrap_existing")

	registry, err := NewRegistry(
		Entry{Name: "existing", Model: &existingModel{}},
		Entry{Name: "missing", Model: &missingModel{}, Dependencies: []string{"existing"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	result, err := Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !reflect.DeepEqual(result.Existing, []string{"existing"}) {
		t.Fatalf("existing entries = %v, want [existing]", result.Existing)
	}
	if !reflect.DeepEqual(result.Created, []string{"missing"}) {
		t.Fatalf("created entries = %v, want [missing]", result.Created)
	}
	if !db.Migrator().HasTable(&missingModel{}) {
		t.Fatal("missing model table was not created")
	}
	after := tableColumns(t, db, "bootstrap_existing")
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("existing table changed: before=%v after=%v", before, after)
	}
	if db.Migrator().HasColumn(&existingModel{}, "id") || db.Migrator().HasColumn(&existingModel{}, "required") {
		t.Fatal("AutoMigrate altered the mismatched existing table")
	}
}

func TestRunIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	registry, err := NewRegistry(Entry{Name: "missing", Model: &missingModel{}})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	first, err := Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	second, err := Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if !reflect.DeepEqual(first.Created, []string{"missing"}) {
		t.Fatalf("first created = %v, want [missing]", first.Created)
	}
	if !reflect.DeepEqual(second.Existing, []string{"missing"}) || len(second.Created) != 0 {
		t.Fatalf("second result = %+v, want existing-only", second)
	}
}

func TestRunDoesNotAutoMigrateExistingRelationshipDependency(t *testing.T) {
	db := openTestDB(t)
	if err := db.Exec(`CREATE TABLE bootstrap_relationship_parent (id integer PRIMARY KEY)`).Error; err != nil {
		t.Fatalf("create mismatched relationship parent: %v", err)
	}
	before := tableColumns(t, db, "bootstrap_relationship_parent")

	registry, err := NewRegistry(
		Entry{Name: "parent", Model: &relationshipParentModel{}},
		Entry{Name: "child", Model: &relationshipChildModel{}, Dependencies: []string{"parent"}},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	result, err := Run(context.Background(), db, registry)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !reflect.DeepEqual(result.Existing, []string{"parent"}) || !reflect.DeepEqual(result.Created, []string{"child"}) {
		t.Fatalf("result = %+v, want existing parent and created child", result)
	}
	after := tableColumns(t, db, "bootstrap_relationship_parent")
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("relationship dependency changed: before=%v after=%v", before, after)
	}
	if db.Migrator().HasColumn(&relationshipParentModel{}, "required") {
		t.Fatal("AutoMigrate added a column to the existing relationship dependency")
	}
	if !db.Migrator().HasConstraint(&relationshipChildModel{}, "fk_bootstrap_relationship_child_parent") {
		t.Fatal("new child table is missing its declared foreign key")
	}
	if !db.Migrator().HasIndex(&relationshipChildModel{}, "idx_bootstrap_relationship_child_parent_id") {
		t.Fatal("new child table is missing its declared index")
	}
}

func TestRunValidatesBeforeChangingDatabase(t *testing.T) {
	db := openTestDB(t)
	registry, err := NewRegistry(
		Entry{Name: "first", Model: &missingModel{}},
		Entry{Name: "duplicate-table", Model: &missingModel{}},
	)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	_, err = Run(context.Background(), db, registry)
	if err == nil || !strings.Contains(err.Error(), "same table") {
		t.Fatalf("Run error = %v, want duplicate table error", err)
	}
	if db.Migrator().HasTable(&missingModel{}) {
		t.Fatal("Run changed the database before validation completed")
	}
}

func TestRunRollsBackPartiallyCreatedTable(t *testing.T) {
	db := openTestDB(t)
	registry, err := NewRegistry(Entry{Name: "invalid-index", Model: &invalidIndexModel{}})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if _, err := Run(context.Background(), db, registry); err == nil {
		t.Fatal("Run with invalid index returned nil error")
	}
	if db.Migrator().HasTable(&invalidIndexModel{}) {
		t.Fatal("failed table bootstrap was not rolled back")
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	db := openTestDB(t)
	registry, err := NewRegistry(Entry{Name: "model", Model: &missingModel{}})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	var nilContext context.Context
	if _, err := Run(nilContext, db, registry); err == nil {
		t.Fatal("Run with nil context returned nil error")
	}
	if _, err := Run(context.Background(), nil, registry); err == nil {
		t.Fatal("Run with nil database returned nil error")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Run(canceled, db, registry); err == nil {
		t.Fatal("Run with canceled context returned nil error")
	}
}

func tableColumns(t *testing.T, db *gorm.DB, table string) []string {
	t.Helper()
	rows, err := db.Raw("PRAGMA table_info(" + table + ")").Rows()
	if err != nil {
		t.Fatalf("read table columns: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Errorf("close table column rows: %v", err)
		}
	}()

	columns := make([]string, 0)
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan table column: %v", err)
		}
		columns = append(columns, name+":"+columnType)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate table columns: %v", err)
	}
	return columns
}
