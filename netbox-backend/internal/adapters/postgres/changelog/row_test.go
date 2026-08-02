package changelog_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
)

func TestChangeRowRetainsPrivateOwnedTableShape(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	statement := &gorm.Statement{DB: db}
	require.NoError(t, statement.Parse(&postgreschangelog.ChangeRow{}))
	require.Equal(t, "go_object_changes", statement.Schema.Table)
	columns := make([]string, 0, len(statement.Schema.Fields))
	for _, field := range statement.Schema.Fields {
		if field.DBName != "" {
			columns = append(columns, field.DBName)
		}
	}
	require.Equal(t, []string{
		"id", "actor_id", "action", "kind", "object_id",
		"before_data", "after_data", "occurred_at",
	}, columns)
}
