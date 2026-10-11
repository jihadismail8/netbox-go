package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type startupRegistryRow struct {
	Name         string   `json:"name"`
	Table        string   `json:"table"`
	Dependencies []string `json:"dependencies"`
}

func TestModelRegistryPreservesExactStartupOrder(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "startup_registry.json"))
	if err != nil {
		t.Fatalf("read startup registry fixture: %v", err)
	}
	var want []startupRegistryRow
	if err := json.Unmarshal(fixture, &want); err != nil {
		t.Fatalf("decode startup registry fixture: %v", err)
	}
	if len(want) != 198 {
		t.Fatalf("startup registry fixture entries = %d, want 198", len(want))
	}

	registry, err := modelRegistry()
	if err != nil {
		t.Fatalf("modelRegistry: %v", err)
	}
	entries := registry.Entries()
	if len(entries) != len(want) {
		t.Fatalf("registry entries = %d, want %d", len(entries), len(want))
	}

	db, err := gorm.Open(sqlite.Open("file:model_registry_order?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open schema resolver: %v", err)
	}
	for index, entry := range entries {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(entry.Model); err != nil {
			t.Fatalf("resolve model %q: %v", entry.Name, err)
		}
		got := startupRegistryRow{
			Name:         entry.Name,
			Table:        statement.Schema.Table,
			Dependencies: append([]string{}, entry.Dependencies...),
		}
		if got.Name != want[index].Name || got.Table != want[index].Table ||
			!slices.Equal(got.Dependencies, want[index].Dependencies) {
			t.Errorf("registry entry %d = %#v, want %#v", index, got, want[index])
		}
	}
}

func TestModelRegistryCoversEveryDeclaredModel(t *testing.T) {
	models := getAllModels()
	registry, err := modelRegistry()
	if err != nil {
		t.Fatalf("modelRegistry: %v", err)
	}
	if got, want := len(registry.Entries()), len(models)+22; got != want {
		t.Fatalf("registry entries = %d, want %d", got, want)
	}
	// The original registry contained 189 generated rows. The 13 promoted
	// Core Workflow resources now have typed Go-owned tables and must never be
	// reintroduced into the frozen legacy bootstrap set.
	if len(models) != 176 {
		t.Fatalf("declared frozen legacy models = %d, want 176", len(models))
	}
	db, err := gorm.Open(sqlite.Open("file:model_registry_names?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open schema resolver: %v", err)
	}
	registered := make(map[string]struct{}, len(models))
	for _, candidate := range models {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(candidate); err != nil {
			t.Fatalf("resolve legacy model %T: %v", candidate, err)
		}
		registered[statement.Schema.Table] = struct{}{}
	}
	retired := []string{
		"dcim_site", "dcim_manufacturer", "dcim_rackrole", "dcim_racktype",
		"dcim_rack", "dcim_devicerole", "dcim_devicetype",
		"dcim_interfacetemplate", "dcim_device", "dcim_interface",
		"ipam_vrf", "ipam_prefix", "ipam_ipaddress",
	}
	for _, table := range retired {
		if _, exists := registered[table]; exists {
			t.Errorf("promoted table %s remains in frozen legacy bootstrap", table)
		}
	}
}
