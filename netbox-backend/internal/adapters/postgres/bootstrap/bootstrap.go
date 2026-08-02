package bootstrap

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Result reports which registry entries were created and which already
// existed. Names are returned in registry order.
type Result struct {
	Created  []string
	Existing []string
}

// Run creates missing tables in registry order. Existing tables are never
// passed to GORM AutoMigrate, so bootstrap cannot add, alter, or remove their
// columns, indexes, or constraints.
func Run(ctx context.Context, db *gorm.DB, registry Registry) (Result, error) {
	if ctx == nil {
		return Result{}, fmt.Errorf("bootstrap context is nil")
	}
	if db == nil {
		return Result{}, fmt.Errorf("bootstrap database is nil")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("bootstrap context: %w", err)
	}

	entries := registry.Entries()
	if err := validateTables(db, entries); err != nil {
		return Result{}, err
	}

	result := Result{
		Created:  make([]string, 0, len(entries)),
		Existing: make([]string, 0, len(entries)),
	}
	for _, entry := range entries {
		operationDB := db.WithContext(ctx)
		created, err := ensureTable(operationDB, entry)
		if err != nil {
			return result, err
		}
		if !created {
			result.Existing = append(result.Existing, entry.Name)
			continue
		}
		result.Created = append(result.Created, entry.Name)
	}

	return result, nil
}

func ensureTable(db *gorm.DB, entry Entry) (bool, error) {
	created := false
	err := db.Transaction(func(operationDB *gorm.DB) error {
		// Recheck inside the DDL transaction. Creation plus its declared FKs and
		// indexes must commit together: a failed constraint cannot leave a table
		// which a later missing-table-only run would incorrectly skip.
		if operationDB.Migrator().HasTable(entry.Model) {
			return nil
		}
		// Relationships are disabled for this AutoMigrate call so GORM cannot
		// recursively auto-migrate (and therefore alter) an existing dependency.
		// Foreign keys declared by this newly-created model are added immediately
		// afterwards; both operations remain confined to the missing table.
		migrationDB := operationDB.Session(&gorm.Session{})
		migrationDB.IgnoreRelationshipsWhenMigrating = true
		if err := migrationDB.AutoMigrate(entry.Model); err != nil {
			return fmt.Errorf("bootstrap table %q: %w", entry.Name, err)
		}
		if err := createForeignKeys(operationDB, entry); err != nil {
			return err
		}
		if err := createIndexes(operationDB, entry); err != nil {
			return err
		}
		created = true
		return nil
	})
	return created, err
}

func createIndexes(db *gorm.DB, entry Entry) error {
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(entry.Model); err != nil {
		return fmt.Errorf("bootstrap table %q: resolve indexes: %w", entry.Name, err)
	}
	// SQLite rebuilds a table to add a foreign key after creation, which drops
	// indexes created by the relationship-free AutoMigrate. Reassert every
	// declared index after the FK pass. PostgreSQL simply observes that each
	// index already exists.
	for _, index := range statement.Schema.ParseIndexes() {
		if db.Migrator().HasIndex(entry.Model, index.Name) {
			continue
		}
		if err := db.Migrator().CreateIndex(entry.Model, index.Name); err != nil {
			return fmt.Errorf(
				"bootstrap table %q: create index %q: %w",
				entry.Name,
				index.Name,
				err,
			)
		}
	}
	return nil
}

func createForeignKeys(db *gorm.DB, entry Entry) error {
	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(entry.Model); err != nil {
		return fmt.Errorf("bootstrap table %q: resolve foreign keys: %w", entry.Name, err)
	}
	for _, relationship := range statement.Schema.Relationships.Relations {
		if relationship.Field.IgnoreMigration {
			continue
		}
		constraint := relationship.ParseConstraint()
		if constraint == nil || constraint.Schema != statement.Schema {
			continue
		}
		if err := db.Migrator().CreateConstraint(entry.Model, constraint.Name); err != nil {
			return fmt.Errorf(
				"bootstrap table %q: create foreign key %q: %w",
				entry.Name,
				constraint.Name,
				err,
			)
		}
	}
	return nil
}

// validateTables resolves every model before changing the database. This
// catches invalid models and duplicate physical table mappings without
// leaving a partially bootstrapped schema behind.
func validateTables(db *gorm.DB, entries []Entry) error {
	tables := make(map[string]string, len(entries))
	for _, entry := range entries {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(entry.Model); err != nil {
			return fmt.Errorf("bootstrap registry entry %q: resolve table: %w", entry.Name, err)
		}
		table := statement.Schema.Table
		if existing, duplicate := tables[table]; duplicate {
			return fmt.Errorf(
				"bootstrap registry entries %q and %q map to the same table %q",
				existing,
				entry.Name,
				table,
			)
		}
		tables[table] = entry.Name
	}
	return nil
}
