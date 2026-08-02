package config

import (
	"fmt"
	"os"
	"strconv"
)

// ApplyEnvironmentOverrides applies the small, documented deployment
// boundary after parsing the generated YAML configuration.
func ApplyEnvironmentOverrides() error {
	cfg := Get()
	if value := os.Getenv("NETBOX_DATABASE_DSN"); value != "" {
		cfg.Database.Postgresql.Dsn = value
	}
	if value := os.Getenv("NETBOX_APP_ENV"); value != "" {
		cfg.App.Env = value
	}
	if value := os.Getenv("NETBOX_HTTP_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("NETBOX_HTTP_PORT: %w", err)
		}
		cfg.HTTP.Port = port
	}
	if value := os.Getenv("NETBOX_GRPC_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("NETBOX_GRPC_PORT: %w", err)
		}
		cfg.Grpc.Port = port
	}
	return nil
}
