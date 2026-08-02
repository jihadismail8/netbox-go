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
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
)

func TestIPAMTypedBootstrapPostgres(t *testing.T) {
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
	schema := fmt.Sprintf("ipam_row_test_%d", time.Now().UnixNano())
	require.NoError(t, tx.Exec(`CREATE SCHEMA "`+schema+`"`).Error)
	require.NoError(t, tx.Exec(`SET LOCAL search_path TO "`+schema+`"`).Error)

	dcimDescriptors := dcimrow.Descriptors()
	ipamDescriptors := ipamrow.Descriptors()
	registry, err := bootstrap.NewRegistry(combinedEntries(dcimDescriptors, ipamDescriptors)...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), tx, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, len(dcimDescriptors)+len(ipamDescriptors))
	for _, descriptor := range ipamDescriptors {
		require.True(t, tx.Migrator().HasTable(descriptor.Model), descriptor.Name)
	}

	var assignmentChecks int64
	require.NoError(t, tx.Raw(
		`SELECT count(*) FROM pg_constraint c JOIN pg_class t ON t.oid = c.conrelid JOIN pg_namespace n ON n.oid = t.relnamespace WHERE n.nspname = ? AND t.relname = ? AND c.conname = ?`,
		schema, "go_ipam_ip_addresses", "chk_go_ip_assignment_pair",
	).Scan(&assignmentChecks).Error)
	require.Equal(t, int64(1), assignmentChecks)

	second, err := bootstrap.Run(t.Context(), tx, registry)
	require.NoError(t, err)
	require.Empty(t, second.Created)
	require.Len(t, second.Existing, len(dcimDescriptors)+len(ipamDescriptors))
}
