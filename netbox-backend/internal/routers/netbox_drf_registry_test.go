package routers

import (
	"sort"
	"strings"
	"testing"
)

// TestAllModelsRegistered fixes the size of the frozen legacy registry. The 13
// Core Workflow resources have canonical replacements and must not reappear.
func TestAllModelsRegistered(t *testing.T) {
	if len(netboxModelRegistry) != 102 {
		t.Errorf("frozen legacy registry = %d endpoints, want 102", len(netboxModelRegistry))
	}
}

// TestCriticalEndpointsRegistered verifies that the core NetBox endpoints
// (the ones used by every NetBox deployment) are present in the registry.
func TestCriticalEndpointsRegistered(t *testing.T) {
	deferredEndpoints := []string{
		// DCIM
		"/api/dcim/platforms",
		"/api/dcim/regions",
		"/api/dcim/locations",
		"/api/dcim/cables",
		// IPAM
		"/api/ipam/vlans",
		"/api/ipam/vlan-groups",
		"/api/ipam/aggregates",
		"/api/ipam/asns",
		// Circuits
		"/api/circuits/circuits",
		"/api/circuits/providers",
		// Tenancy
		"/api/tenancy/tenants",
		"/api/tenancy/tenant-groups",
		// Virtualization
		"/api/virtualization/clusters",
		"/api/virtualization/virtual-machines",
	}

	for _, ep := range deferredEndpoints {
		t.Run(ep, func(t *testing.T) {
			if _, exists := netboxModelRegistry[ep]; !exists {
				t.Errorf("critical endpoint %q not registered", ep)
			}
		})
	}
}

func TestPromotedEndpointsAreRetiredFromLegacyRegistry(t *testing.T) {
	retired := []string{
		"/api/dcim/sites", "/api/dcim/manufacturers", "/api/dcim/rack-roles",
		"/api/dcim/rack-types", "/api/dcim/racks", "/api/dcim/device-roles",
		"/api/dcim/device-types", "/api/dcim/interface-templates",
		"/api/dcim/devices", "/api/dcim/interfaces", "/api/ipam/vrfs",
		"/api/ipam/prefixes", "/api/ipam/ip-addresses",
	}
	for _, endpoint := range retired {
		if _, exists := netboxModelRegistry[endpoint]; exists {
			t.Errorf("promoted endpoint %q remains in legacy registry", endpoint)
		}
	}
}

// TestEndpointConfigValidity verifies that every registered endpoint has the
// minimum required fields for a functioning DRF-compatible API.
func TestEndpointConfigValidity(t *testing.T) {
	for path, cfg := range netboxModelRegistry {
		t.Run(path, func(t *testing.T) {
			// Every endpoint must have a table name
			if cfg.TableName == "" {
				t.Errorf("endpoint %q has empty TableName", path)
			}

			// Every endpoint must have list fields (at least id)
			if len(cfg.ListFields) == 0 {
				t.Errorf("endpoint %q has no ListFields", path)
			}
			hasID := false
			for _, f := range cfg.ListFields {
				if f == "id" {
					hasID = true
					break
				}
			}
			if !hasID {
				t.Errorf("endpoint %q ListFields missing 'id'", path)
			}

			// Every endpoint must have a default sort
			if cfg.DefaultSort == "" {
				t.Errorf("endpoint %q has empty DefaultSort", path)
			}

			// Every endpoint must have filter columns (at least q)
			if len(cfg.FilterCols) == 0 {
				t.Errorf("endpoint %q has no FilterCols", path)
			}
		})
	}
}

// TestFKTableResolution verifies the FK table resolution logic handles
// common NetBox naming patterns correctly.
func TestFKTableResolution(t *testing.T) {
	// Register known tables for this test (already done by autogen init,
	// but we add a few extra to ensure isolation)
	RegisterKnownTables([]string{
		"dcim_region", "dcim_site", "dcim_sitegroup",
		"tenancy_tenant", "ipam_vlan", "ipam_vrf",
	})

	tests := []struct {
		name        string
		sourceTable string
		fkBase      string
		wantTable   string
	}{
		// Standard same-module FK
		{"site to region", "dcim_site", "region", "dcim_region"},
		{"site to sitegroup override", "dcim_site", "group", "dcim_sitegroup"},
		{"device to site", "dcim_device", "site", "dcim_site"},
		// Cross-module FK
		{"site to tenant", "dcim_site", "tenant", "tenancy_tenant"},
		{"prefix to vrf", "ipam_prefix", "vrf", "ipam_vrf"},
		{"prefix to vlan", "ipam_prefix", "vlan", "ipam_vlan"},
		// Self-referential
		{"region self-ref", "dcim_region", "parent", "dcim_region"},
		// Unresolvable (generic/polymorphic)
		{"generic scope FK", "extras_bookmark", "scope", ""},
		{"generic assigned_object FK", "extras_journalentry", "assigned_object", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveFKTable(tt.sourceTable, tt.fkBase)
			if got != tt.wantTable {
				t.Errorf("resolveFKTable(%q, %q) = %q, want %q",
					tt.sourceTable, tt.fkBase, got, tt.wantTable)
			}
		})
	}
}

// TestKnownModelTablesPopulated verifies that the autogen init() has populated
// the knownModelTables map with all model tables for FK resolution.
func TestKnownModelTablesPopulated(t *testing.T) {
	expectedTables := []string{
		"dcim_region",
		"dcim_sitegroup",
		"dcim_location",
		"ipam_vlan",
		"tenancy_tenant",
		"circuits_circuit",
	}

	for _, table := range expectedTables {
		if !knownModelTables[table] {
			t.Errorf("expected table %q in knownModelTables", table)
		}
	}

	if len(knownModelTables) < 100 {
		t.Errorf("expected at least 100 known model tables, got %d", len(knownModelTables))
	}
}

// TestModuleDistribution verifies that endpoints span all expected NetBox modules.
// This catches issues where a whole module's models fail to generate.
func TestModuleDistribution(t *testing.T) {
	moduleCounts := map[string]int{}
	for path := range netboxModelRegistry {
		// Extract module from path: /api/dcim/sites → dcim
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			module := parts[2] // /api/[module]/resource
			moduleCounts[module]++
		}
	}

	expectedModules := []string{"dcim", "ipam", "circuits", "tenancy", "virtualization"}
	for _, module := range expectedModules {
		if moduleCounts[module] == 0 {
			t.Errorf("expected endpoints for module %q, found none", module)
		}
	}

	t.Logf("Module distribution:")
	modules := make([]string, 0, len(moduleCounts))
	for mod := range moduleCounts {
		modules = append(modules, mod)
	}
	sort.Strings(modules)
	for _, mod := range modules {
		t.Logf("  %s: %d endpoints", mod, moduleCounts[mod])
	}
}

// TestRegistryMatchesKnownEndpointCount is a second guard around init-time
// registration, independent of the per-entry validation above.
func TestRegistryMatchesKnownEndpointCount(t *testing.T) {
	if len(netboxModelRegistry) != 102 {
		t.Errorf("expected 102 frozen legacy endpoints, got %d", len(netboxModelRegistry))
	}
}
