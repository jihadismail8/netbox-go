package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestParseIPAddressListPreservesRepeatedAndAssignmentFilters(t *testing.T) {
	query, err := parseIPAddressList(url.Values{
		"limit": {"0"}, "id": {"-1,0", "11"}, "vrf_id": {"-2", "7"},
		"address": {"192.0.2.1", "2001:db8::1/64"}, "family": {"-4"},
		"parent": {"192.0.2.99/24"}, "status": {"active", "reserved"},
		"assigned": {"true"}, "interface_id": {"-3,9"},
		"device_id": {"-4", "8"}, "ordering": {"vrf,-address"},
	})
	require.NoError(t, err)
	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumIPAddressPageLimit, query.EffectiveLimit())
	assert.Equal(t, []int64{-1, 0, 11}, query.IDs)
	assert.Equal(t, []int64{-2, 7}, query.VRFIDs)
	assert.Equal(t, []string{"192.0.2.1", "2001:db8::1/64"}, query.Addresses)
	assert.Equal(t, int64(-4), *query.Family)
	assert.Equal(t, "192.0.2.99/24", *query.Parent)
	assert.True(t, *query.Assigned)
	assert.Equal(t, []int64{-3, 9}, query.InterfaceIDs)
	assert.Equal(t, []int64{-4, 8}, query.DeviceIDs)
	assert.Equal(t, []string{"vrf", "-address"}, query.Ordering)
}

func TestIPAddressRESTPatchPreservesAssignmentNullPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restIPAddressServiceSpy{address: restIPAddressFixture(t, true)}
	router := gin.New()
	NewIPAddressRESTHandler(spy).Register(
		router, testPrefixPrincipalMiddleware(),
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/ipam/ip-addresses/17/",
		strings.NewReader(
			`{"dns_name":"EDGE.EXAMPLE","assigned_object_type":null,"assigned_object_id":null}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, applicationipam.FieldPresent, spy.updateCommand.DNSName.State())
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.AssignedObjectType.State())
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.AssignedObjectID.State())
	assert.Equal(t, applicationipam.FieldOmitted, spy.updateCommand.Address.State())

	var projected ipAddressDTO
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &projected))
	require.NotNil(t, projected.AssignedObject)
	assert.Equal(t, int64(11), projected.AssignedObject.ID)
	assert.Equal(t, "Ethernet1 (uplink)", projected.AssignedObject.Display)
	assert.Equal(t, "dcim.interface", *projected.AssignedObjectType)
	assert.Equal(t, int64(11), *projected.AssignedObjectID)
}

func restIPAddressFixture(
	t *testing.T,
	assigned bool,
) *domainipam.IPAddress {
	t.Helper()
	stamp := shared.NewTimestamp(
		time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
	)
	assignment := domainipam.NullInterfaceAssignment()
	if assigned {
		device, err := domaindcim.NewDeviceReference(
			7, domaindcim.NonNullDeviceValue("edge-01"), "edge-01",
		)
		require.NoError(t, err)
		networkInterface, err := domaindcim.RestoreInterface(
			domaindcim.InterfaceState{
				ID: 11, Device: device, Name: "Ethernet1", Label: "uplink",
				Type: "1000base-t", Enabled: true, Created: stamp,
				LastUpdated: stamp,
			},
		)
		require.NoError(t, err)
		value, err := domainipam.NewInterfaceAssignment(networkInterface)
		require.NoError(t, err)
		assignment = domainipam.NonNullInterfaceAssignment(value)
	}
	address, err := domainipam.RestoreIPAddress(domainipam.IPAddressState{
		ID: 17, Address: "192.0.2.1/24",
		VRF:        domainipam.NullVRFReference(),
		Status:     domainipam.IPAddressStatusActive.String(),
		Assignment: assignment, Created: stamp, LastUpdated: stamp,
	})
	require.NoError(t, err)
	return address
}

type restIPAddressServiceSpy struct {
	updateCommand applicationipam.UpdateIPAddressCommand
	address       *domainipam.IPAddress
}

func (*restIPAddressServiceSpy) ListIPAddresses(
	context.Context,
	identity.Principal,
	applicationipam.ListIPAddressesQuery,
) (applicationipam.IPAddressPage, error) {
	return applicationipam.IPAddressPage{}, nil
}

func (spy *restIPAddressServiceSpy) GetIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.GetIPAddressQuery,
) (*domainipam.IPAddress, error) {
	return spy.address, nil
}

func (spy *restIPAddressServiceSpy) CreateIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.CreateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	return spy.address, nil
}

func (spy *restIPAddressServiceSpy) ReplaceIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.ReplaceIPAddressCommand,
) (*domainipam.IPAddress, error) {
	return spy.address, nil
}

func (spy *restIPAddressServiceSpy) UpdateIPAddress(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.UpdateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.updateCommand = command
	return spy.address, nil
}

func (*restIPAddressServiceSpy) DeleteIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.DeleteIPAddressCommand,
) error {
	return nil
}
