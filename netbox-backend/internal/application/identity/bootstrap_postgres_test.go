package identity_test

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	application "netbox-go/internal/application/identity"
	"netbox-go/internal/domain/shared"
)

func TestPostgresAdministratorBootstrapIsOneTimeAcrossConcurrentCallers(t *testing.T) {
	baseDSN := strings.TrimSpace(os.Getenv("NETBOX_TEST_POSTGRES_DSN"))
	if baseDSN == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}

	adminDB, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{})
	require.NoError(t, err)
	schema := fmt.Sprintf("identity_bootstrap_%d", time.Now().UnixNano())
	require.NoError(t, adminDB.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() { require.NoError(t, adminDB.Exec("DROP SCHEMA "+schema+" CASCADE").Error) })

	testDB, err := gorm.Open(postgres.Open(identityTestDSN(t, baseDSN, schema)), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, testDB.AutoMigrate(identitypostgres.Models()...))
	service := application.NewService(identitypostgres.NewStore(testDB), application.RealClock{})

	type result struct{ err error }
	start := make(chan struct{})
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for _, username := range []string{"first-admin", "second-admin"} {
		workers.Add(1)
		go func(username string) {
			defer workers.Done()
			<-start
			_, bootstrapErr := service.BootstrapAdministrator(t.Context(), username, "", "Concurrent-Admin-2026!")
			results <- result{err: bootstrapErr}
		}(username)
	}
	close(start)
	workers.Wait()
	close(results)

	succeeded, conflicted := 0, 0
	for outcome := range results {
		if outcome.err == nil {
			succeeded++
			continue
		}
		var appErr *shared.Error
		require.ErrorAs(t, outcome.err, &appErr)
		require.Equal(t, shared.ErrorReasonConflict, appErr.Reason)
		conflicted++
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, conflicted)
	var count int64
	require.NoError(t, testDB.Model(&identitypostgres.UserRow{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func identityTestDSN(t *testing.T, dsn, schema string) string {
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
