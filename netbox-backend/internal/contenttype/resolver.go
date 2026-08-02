// Package contenttype — database seeding and runtime resolution.
package contenttype

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ctRow mirrors the django_content_type table shape for inserts/queries.
type ctRow struct {
	ID       uint64 `gorm:"column:id;primaryKey"`
	AppLabel string `gorm:"column:app_label"`
	Model    string `gorm:"column:model"`
}

func (ctRow) TableName() string { return "django_content_type" }

// Seed populates django_content_type with every concrete NetBox model,
// idempotently. Safe to call on every startup. The (app_label, model) pair is
// the natural key; duplicate inserts are ignored.
//
// Returns the count of rows currently in the table after seeding (useful for
// logging) and any error.
func Seed(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("nil db handle")
	}

	// Build rows from the static seed list. We let the DB allocate the numeric
	// id (serial/identity) so IDs are stable across restarts and reflect the
	// order rows were first inserted — matching Django's behavior.
	rows := make([]ctRow, 0, len(allTypes))
	for _, t := range allTypes {
		rows = append(rows, ctRow{AppLabel: t.AppLabel, Model: t.Model})
	}

	// Clause.OnConflict with DoNothing works on both PostgreSQL (ON CONFLICT)
	// and SQLite (INSERT OR IGNORE via the same abstraction). The conflict
	// target is the (app_label, model) natural key. We rely on a unique index
	// existing on those columns — Django creates one; for SQLite tests the
	// caller is responsible for the schema. If no unique constraint exists,
	// the DoNothing falls back to no-op on conflict by primary key (id=0),
	// which is harmless since we don't set ids.
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&rows).Error; err != nil {
		// Fallback: some SQLite versions don't support the upsert form. Try
		// a plain row-by-row insert, skipping duplicates.
		if err := seedFallback(db); err != nil {
			return 0, err
		}
	}

	var count int64
	if err := db.Table("django_content_type").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count django content types: %w", err)
	}
	return count, nil
}

// seedFallback inserts each row individually with duplicate-safe SQL.
// Used when the bulk upsert isn't supported by the dialect.
func seedFallback(db *gorm.DB) error {
	for _, t := range allTypes {
		// Use a raw INSERT with the dialect-appropriate ignore clause.
		if err := db.Exec(
			dialectIgnoreInsert(db),
			t.AppLabel, t.Model,
		).Error; err != nil {
			return fmt.Errorf("seed django content type %s.%s: %w", t.AppLabel, t.Model, err)
		}
	}
	return nil
}

// dialectIgnoreInsert returns a parameterized INSERT that ignores duplicates,
// chosen by DB dialect.
func dialectIgnoreInsert(db *gorm.DB) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "INSERT OR IGNORE INTO django_content_type (app_label, model) VALUES (?, ?)"
	default: // postgres
		return "INSERT INTO django_content_type (app_label, model) VALUES (?, ?) ON CONFLICT (app_label, model) DO NOTHING"
	}
}

// ─── Resolver ───────────────────────────────────────────────────────────────

// Cache is the in-memory resolver, loaded once from django_content_type.
// Lookups are O(1) and allocation-free after the initial load.
type Cache struct {
	mu      sync.RWMutex
	byID    map[uint64]Type   // content_type.id → Type
	byLabel map[string]uint64 // "app_label.model" → id
	loaded  bool
}

var globalCache = &Cache{
	byLabel: map[string]uint64{},
	byID:    map[uint64]Type{},
}

// Load populates the cache from the database. Safe to call multiple times;
// only the first call hits the DB (subsequent calls are no-ops unless
// Invalidate is called first).
func (c *Cache) Load(db *gorm.DB) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return nil
	}
	return c.loadLocked(db)
}

func (c *Cache) loadLocked(db *gorm.DB) error {
	var rows []ctRow
	if err := db.Table("django_content_type").Find(&rows).Error; err != nil {
		return err
	}
	// Merge the DB rows with the static table-name map (the DB only stores
	// app_label + model; the table name comes from our static seed list).
	tableByLabel := map[string]string{}
	for _, t := range allTypes {
		tableByLabel[t.AppLabel+"."+t.Model] = t.Table
	}
	for _, r := range rows {
		label := r.AppLabel + "." + r.Model
		t := Type{AppLabel: r.AppLabel, Model: r.Model}
		if tbl, ok := tableByLabel[label]; ok {
			t.Table = tbl
		}
		c.byID[r.ID] = t
		c.byLabel[label] = r.ID
	}
	c.loaded = true
	return nil
}

// Invalidate clears the cache, forcing the next Load to re-read from the DB.
// Useful in tests that re-seed between cases.
func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byID = map[uint64]Type{}
	c.byLabel = map[string]uint64{}
	c.loaded = false
}

// Resolve returns the Type for a content_type id. The second return is false
// if the id is unknown. Loads the cache on first use.
func Resolve(db *gorm.DB, id uint64) (Type, bool) {
	if err := globalCache.Load(db); err != nil {
		return Type{}, false
	}
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	t, ok := globalCache.byID[id]
	return t, ok
}

// LookupByLabel returns the content_type id for an "app_label.model" string
// (e.g. "dcim.interface"). Returns false if not found.
func LookupByLabel(db *gorm.DB, appLabel, model string) (uint64, bool) {
	if err := globalCache.Load(db); err != nil {
		return 0, false
	}
	globalCache.mu.RLock()
	defer globalCache.mu.RUnlock()
	id, ok := globalCache.byLabel[appLabel+"."+model]
	return id, ok
}

// InvalidateCache clears the global cache (for tests).
func InvalidateCache() { globalCache.Invalidate() }
