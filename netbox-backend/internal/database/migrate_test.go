package database

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
