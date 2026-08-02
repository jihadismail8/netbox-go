package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	identityapp "netbox-go/internal/application/identity"
)

// TestAdminCLIProcessIdentityAdministration exercises the compiled command
// against an isolated PostgreSQL schema. The owned PostgreSQL CI job supplies
// the DSN; ordinary unit-test runs remain hermetic.
func TestAdminCLIProcessIdentityAdministration(t *testing.T) {
	baseDSN := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if baseDSN == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	adminDB, err := gorm.Open(postgres.Open(baseDSN), &gorm.Config{})
	require.NoError(t, err)
	schema := fmt.Sprintf("netbox_admin_cli_%d", time.Now().UnixNano())
	require.NoError(t, adminDB.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() { require.NoError(t, adminDB.Exec("DROP SCHEMA "+schema+" CASCADE").Error) })
	scopedDSN := adminCLISchemaDSN(t, baseDSN, schema)

	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	binary := filepath.Join(t.TempDir(), "netbox_go_admin")
	buildContext, cancelBuild := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancelBuild()
	build := exec.CommandContext(buildContext, "go", "build", "-o", binary, ".")
	build.Dir = workingDirectory
	build.Env = append(os.Environ(), "GOCACHE=/tmp/go-cache", "GOFLAGS=-buildvcs=false")
	buildOutput, err := build.CombinedOutput()
	require.NoError(t, err, "build admin CLI: %s", buildOutput)
	configPath := filepath.Clean(filepath.Join(workingDirectory, "..", "..", "configs", "netbox_go.yml"))

	bootstrap := runAdminCLIProcess(t, binary, configPath, scopedDSN, "bootstrap", "admin", "admin@example.test", "Initial-Admin-Password-2026!")
	require.Contains(t, bootstrap, `administrator "admin" created`)
	duplicateOutput, duplicateErr := runAdminCLIProcessError(t, binary, configPath, scopedDSN, "bootstrap", "second", "", "Second-Admin-Password-2026!")
	require.Error(t, duplicateErr)
	require.Contains(t, duplicateOutput, "allowed only on an empty identity store")

	db, err := gorm.Open(postgres.Open(scopedDSN), &gorm.Config{})
	require.NoError(t, err)
	service := identityapp.NewService(identitypostgres.NewStore(db), identityapp.RealClock{})
	_, err = service.AuthenticatePassword(t.Context(), "admin", "Initial-Admin-Password-2026!")
	require.NoError(t, err)
	session, err := service.Login(t.Context(), "admin", "Initial-Admin-Password-2026!")
	require.NoError(t, err)

	reset := runAdminCLIProcess(t, binary, configPath, scopedDSN, "reset-password", "admin", "", "Replacement-Admin-Password-2026!")
	require.Contains(t, reset, `password reset for administrator "admin"`)
	_, err = service.AuthenticateSession(t.Context(), session.Secret)
	require.Error(t, err, "the process-level reset must invalidate sessions created by another process")
	_, err = service.AuthenticatePassword(t.Context(), "admin", "Initial-Admin-Password-2026!")
	require.Error(t, err)
	_, err = service.AuthenticatePassword(t.Context(), "admin", "Replacement-Admin-Password-2026!")
	require.NoError(t, err)

	const limitedPassword = "Limited-Viewer-Password-2026!"
	createOutput := runAdminCLIProcessWithInput(t, binary, scopedDSN,
		"Replacement-Admin-Password-2026!\n"+limitedPassword+"\n"+limitedPassword+"\n",
		"create-user", "--config", configPath, "--actor-username", "admin",
		"--username", "limited", "--email", "limited@example.test",
	)
	require.Contains(t, createOutput, `user "limited" created`)
	require.NotContains(t, createOutput, "Replacement-Admin-Password-2026!")
	require.NotContains(t, createOutput, limitedPassword)
	limited, err := service.AuthenticatePassword(t.Context(), "limited", limitedPassword)
	require.NoError(t, err)
	require.False(t, limited.IsSuperuser)
	require.False(t, limited.IsStaff)
	require.Empty(t, limited.Permissions)
	var storedHash string
	require.NoError(t, db.Raw("SELECT password_hash FROM go_identity_users WHERE username = ?", "limited").Scan(&storedHash).Error)
	require.NotEmpty(t, storedHash)
	require.NotContains(t, createOutput, storedHash, "SQL logging must not disclose the persisted credential hash")

	deniedOutput, deniedErr := runAdminCLIProcessWithInputError(t, binary, scopedDSN, limitedPassword+"\n",
		"grant-permission", "--config", configPath, "--actor-username", "limited",
		"--username", "limited", "--permission", "dcim.view_site",
	)
	require.Error(t, deniedErr)
	require.Contains(t, deniedOutput, "administrator authentication failed")
	require.NotContains(t, deniedOutput, limitedPassword)

	grantOutput := runAdminCLIProcessWithInput(t, binary, scopedDSN,
		"Replacement-Admin-Password-2026!\n",
		"grant-permission", "--config", configPath, "--actor-username", "admin",
		"--username", "limited", "--permission", "dcim.view_site",
	)
	require.Contains(t, grantOutput, `permission "dcim.view_site" granted to user "limited"`)
	require.NotContains(t, grantOutput, "Replacement-Admin-Password-2026!")
	limited, err = service.AuthenticatePassword(t.Context(), "limited", limitedPassword)
	require.NoError(t, err)
	require.True(t, limited.Principal().Has("dcim.view_site"))
}

func runAdminCLIProcess(t *testing.T, binary, configPath, dsn, command, username, email, password string) string {
	t.Helper()
	output, err := runAdminCLIProcessError(t, binary, configPath, dsn, command, username, email, password)
	require.NoError(t, err, "admin CLI output: %s", output)
	return output
}

func runAdminCLIProcessError(t *testing.T, binary, configPath, dsn, command, username, email, password string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	arguments := []string{command, "--config", configPath, "--username", username}
	if email != "" {
		arguments = append(arguments, "--email", email)
	}
	process := exec.CommandContext(ctx, binary, arguments...)
	process.Env = append(os.Environ(), "NETBOX_DATABASE_DSN="+dsn)
	process.Stdin = strings.NewReader(password + "\n" + password + "\n")
	output, err := process.CombinedOutput()
	return string(output), err
}

func runAdminCLIProcessWithInput(t *testing.T, binary, dsn, stdin string, arguments ...string) string {
	t.Helper()
	output, err := runAdminCLIProcessWithInputError(t, binary, dsn, stdin, arguments...)
	require.NoError(t, err, "admin CLI output: %s", output)
	return output
}

func runAdminCLIProcessWithInputError(t *testing.T, binary, dsn, stdin string, arguments ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	process := exec.CommandContext(ctx, binary, arguments...)
	process.Env = append(os.Environ(), "NETBOX_DATABASE_DSN="+dsn)
	process.Stdin = strings.NewReader(stdin)
	output, err := process.CombinedOutput()
	return string(output), err
}

func adminCLISchemaDSN(t *testing.T, dsn, schema string) string {
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
