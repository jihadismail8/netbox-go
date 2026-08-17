package restcontract

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"netbox-go/api/openapi"
	runtimehttp "netbox-go/internal/adapters/rest/netbox/router"
	workflowhttp "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/config"
	"netbox-go/internal/platform/composition"
)

func TestRuntimeRoutesConformToGeneratedOpenAPI(t *testing.T) {
	config.Set(&config.Config{})
	var document struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(openapi.Schema, &document))
	expected := map[string]struct{}{}
	for path, operations := range document.Paths {
		for method := range operations {
			if method == "parameters" || strings.HasPrefix(method, "x-") {
				continue
			}
			expected[strings.ToUpper(method)+" "+strings.TrimSuffix(path, "/")] = struct{}{}
		}
	}
	db, err := gorm.Open(sqlite.Open("file:surface_contract?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	core := composition.NewCore(db)
	router := runtimehttp.New(
		core.Identity,
		core.Sites,
		false,
		nil,
		workflowhttp.WithOrganizationServices(core.Manufacturers, core.RackRoles),
		workflowhttp.WithRackTypeService(core.RackTypes),
		workflowhttp.WithRackService(core.Racks),
		workflowhttp.WithDeviceRoleService(core.DeviceRoles),
		workflowhttp.WithDeviceTypeService(core.DeviceTypes),
		workflowhttp.WithInterfaceTemplateService(core.InterfaceTemplates),
		workflowhttp.WithDeviceService(core.Devices),
		workflowhttp.WithInterfaceService(core.Interfaces),
		workflowhttp.WithVRFService(core.VRFs),
		workflowhttp.WithPrefixService(core.Prefixes),
		workflowhttp.WithIPAddressService(core.IPAddresses),
	)
	actual := map[string]struct{}{}
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/") || strings.TrimSuffix(route.Path, "/") == "/api/schema" {
			continue
		}
		path := strings.ReplaceAll(route.Path, ":id", "{id}")
		actual[route.Method+" "+strings.TrimSuffix(path, "/")] = struct{}{}
	}
	require.Equal(t, expected, actual)
}
