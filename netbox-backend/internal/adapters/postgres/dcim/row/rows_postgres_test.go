package row_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
)

func TestDCIMTypedBootstrapPostgres(t *testing.T) {
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { _ = tx.Rollback().Error })
	schema := fmt.Sprintf("dcim_row_test_%d", time.Now().UnixNano())
	require.NoError(t, tx.Exec(`CREATE SCHEMA "`+schema+`"`).Error)
	require.NoError(t, tx.Exec(`SET LOCAL search_path TO "`+schema+`"`).Error)

	descriptors := dcimrow.Descriptors()
	registry, err := bootstrap.NewRegistry(dcimEntries(descriptors)...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), tx, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, len(descriptors))
	for _, descriptor := range descriptors {
		require.True(t, tx.Migrator().HasTable(descriptor.Model), descriptor.Name)
	}

	var dataType string
	require.NoError(t, tx.Raw(
		`SELECT data_type FROM information_schema.columns WHERE table_schema = ? AND table_name = ? AND column_name = ?`,
		schema, "go_dcim_device_types", "u_height",
	).Scan(&dataType).Error)
	require.Equal(t, "numeric", dataType)

	second, err := bootstrap.Run(t.Context(), tx, registry)
	require.NoError(t, err)
	require.Empty(t, second.Created)
	require.Len(t, second.Existing, len(descriptors))
}
