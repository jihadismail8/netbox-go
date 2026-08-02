package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProvisioningCommands(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		want      commandOptions
	}{
		{
			name: "create non-superuser",
			arguments: []string{
				"create-user", "--config", "/tmp/config.yml", "--actor-username", "admin",
				"--username", "viewer", "--email", "viewer@example.test",
			},
			want: commandOptions{
				command: "create-user", configFile: "/tmp/config.yml", actorUsername: "admin",
				username: "viewer", email: "viewer@example.test",
			},
		},
		{
			name: "grant model permission",
			arguments: []string{
				"grant-permission", "--config", "/tmp/config.yml", "--actor-username", "admin",
				"--username", "viewer", "--permission", "dcim.view_site",
			},
			want: commandOptions{
				command: "grant-permission", configFile: "/tmp/config.yml", actorUsername: "admin",
				username: "viewer", permission: "dcim.view_site",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var diagnostics bytes.Buffer
			got, err := parseCommand(test.arguments, &diagnostics)
			require.NoError(t, err, diagnostics.String())
			require.Equal(t, test.want, got)
		})
	}
}

func TestParseCommandRejectsPasswordsInArgumentsWithoutDisclosingThem(t *testing.T) {
	const secret = "Never-A-Command-Argument-2026!"
	tests := [][]string{
		{
			"create-user", "--actor-username", "admin", "--username", "viewer",
			"--password=" + secret,
		},
		{
			"grant-permission", "--actor-username", "admin", "--username", "viewer",
			"--permission", "dcim.view_site", secret,
		},
	}

	for _, arguments := range tests {
		var diagnostics bytes.Buffer
		_, err := parseCommand(arguments, &diagnostics)
		require.Error(t, err)
		combined := diagnostics.String() + err.Error()
		require.NotContains(t, combined, secret)
	}
}

func TestProvisioningPasswordsShareOneBufferedStdinAndNeverReachPrompts(t *testing.T) {
	const actorPassword = "Administrator-Secret-2026!"
	const userPassword = "Limited-User-Secret-2026!"
	stdin := strings.NewReader(actorPassword + "\n" + userPassword + "\n" + userPassword + "\n")
	reader := bufio.NewReader(stdin)
	var prompts bytes.Buffer

	actor, err := readPassword(reader, stdin, &prompts, "Administrator password: ")
	require.NoError(t, err)
	created, err := readConfirmedPassword(reader, stdin, &prompts, "New password: ", "Confirm new password: ")
	require.NoError(t, err)
	require.Equal(t, actorPassword, actor)
	require.Equal(t, userPassword, created)
	require.Equal(t, "Administrator password: New password: Confirm new password: ", prompts.String())
	require.NotContains(t, prompts.String(), actorPassword)
	require.NotContains(t, prompts.String(), userPassword)
}

func TestParsePermissionCodename(t *testing.T) {
	input, err := parsePermission(" dcim.view_site ")
	require.NoError(t, err)
	require.Equal(t, "dcim", input.AppLabel)
	require.Equal(t, "view", input.Action)
	require.Equal(t, "site", input.Model)

	for _, invalid := range []string{"", "view_site", "dcim.view", "dcim..view_site"} {
		t.Run(strings.ReplaceAll(invalid, ".", "_"), func(t *testing.T) {
			_, err := parsePermission(invalid)
			require.Error(t, err)
		})
	}
}
