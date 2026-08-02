package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceRESTPreservesNullBlankChoicesFiltersAndWritePresence(t *testing.T) {
	spy := &interfaceServiceSpy{networkInterface: restInterfaceFixture(t)}
	router := interfaceRESTRouter(spy)

	create := interfaceTemplateRESTRequest(
		t, router, http.MethodPost, "/api/dcim/interfaces/",
		`{"device":7,"name":"Ethernet1","type":"10gbase-sr","enabled":false,`+
			`"mtu":null,"speed":2147483647,"duplex":""}`,
	)
	require.Equal(t, http.StatusCreated, create.Code, create.Body.String())
	var payload map[string]any
	require.NoError(t, json.Unmarshal(create.Body.Bytes(), &payload))
	assert.Equal(t, "Ethernet1 (WAN)", payload["display"])
	assert.Equal(t, map[string]any{
		"value": "10gbase-sr", "label": "10GBASE-SR (10GE)",
	}, payload["type"])
	assert.Nil(t, payload["duplex"], "blank nullable choices serialize as null")
	assert.Equal(t, float64(3), payload["count_ipaddresses"])
	assert.Equal(t, map[string]any{
		"id": float64(7), "url": "/api/dcim/devices/7/", "display": "edge-01",
	}, payload["device"])
	assert.Equal(t, applicationdcim.FieldNull, spy.createCommand.MTU.State())
	speed, present := spy.createCommand.Speed.Get()
	require.True(t, present)
	assert.Equal(t, uint64(2147483647), speed)
	duplex, present := spy.createCommand.Duplex.Get()
	require.True(t, present)
	assert.Empty(t, duplex)

	update := interfaceTemplateRESTRequest(
		t, router, http.MethodPatch, "/api/dcim/interfaces/41/",
		`{"mtu":null,"duplex":"auto","mgmt_only":false}`,
	)
	require.Equal(t, http.StatusOK, update.Code, update.Body.String())
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.MTU.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Speed.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Device.State())
	mgmtOnly, present := spy.updateCommand.MgmtOnly.Get()
	require.True(t, present)
	assert.False(t, mgmtOnly)

	list := interfaceTemplateRESTRequest(
		t, router, http.MethodGet,
		"/api/dcim/interfaces/?device_id=-1&device_id=7&device_name=edge-01"+
			"&name=Ethernet1&type=10gbase-sr&enabled=false&mgmt_only=true"+
			"&ordering=-type,name",
		"",
	)
	require.Equal(t, http.StatusOK, list.Code, list.Body.String())
	assert.Equal(t, []int64{-1, 7}, spy.listQuery.DeviceIDs)
	assert.Equal(t, []string{"edge-01"}, spy.listQuery.DeviceNames)
	assert.Equal(t, []string{"Ethernet1"}, spy.listQuery.Names)
	assert.Equal(t, []string{"10gbase-sr"}, spy.listQuery.Types)
	assert.Equal(t, []string{"-type", "name"}, spy.listQuery.Ordering)

	unknown := interfaceTemplateRESTRequest(
		t, router, http.MethodPost, "/api/dcim/interfaces/",
		`{"device":7,"name":"Ethernet1","type":"other","bridge":9}`,
	)
	require.Equal(t, http.StatusBadRequest, unknown.Code, unknown.Body.String())
	assert.Equal(t, 1, spy.createCalls)
}

func interfaceRESTRouter(service InterfaceService) http.Handler {
	engine := newTestRouter()
	NewInterfaceRESTHandler(service).Register(engine, authenticatedRESTMiddleware())
	return engine
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func authenticatedRESTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{
			ID: 1, Username: "interface-rest", IsSuperuser: true,
		})
		c.Next()
	}
}

type interfaceServiceSpy struct {
	listCalls        int
	createCalls      int
	updateCalls      int
	replaceCalls     int
	deleteCalls      int
	listQuery        applicationdcim.ListInterfacesQuery
	createCommand    applicationdcim.CreateInterfaceCommand
	updateCommand    applicationdcim.UpdateInterfaceCommand
	replaceCommand   applicationdcim.ReplaceInterfaceCommand
	deleteCommand    applicationdcim.DeleteInterfaceCommand
	networkInterface *domaindcim.Interface
}

func (spy *interfaceServiceSpy) ListInterfaces(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListInterfacesQuery,
) (applicationdcim.InterfacePage, error) {
	spy.listCalls++
	spy.listQuery = query
	return applicationdcim.InterfacePage{}, nil
}

func (spy *interfaceServiceSpy) GetInterface(
	context.Context,
	identity.Principal,
	applicationdcim.GetInterfaceQuery,
) (*domaindcim.Interface, error) {
	return spy.networkInterface, nil
}

func (spy *interfaceServiceSpy) CreateInterface(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateInterfaceCommand,
) (*domaindcim.Interface, error) {
	spy.createCalls++
	spy.createCommand = command
	return spy.networkInterface, nil
}

func (spy *interfaceServiceSpy) ReplaceInterface(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceInterfaceCommand,
) (*domaindcim.Interface, error) {
	spy.replaceCalls++
	spy.replaceCommand = command
	return spy.networkInterface, nil
}

func (spy *interfaceServiceSpy) UpdateInterface(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateInterfaceCommand,
) (*domaindcim.Interface, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.networkInterface, nil
}

func (spy *interfaceServiceSpy) DeleteInterface(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteInterfaceCommand,
) error {
	spy.deleteCalls++
	spy.deleteCommand = command
	return nil
}

func restInterfaceFixture(t *testing.T) *domaindcim.Interface {
	t.Helper()
	reference, err := domaindcim.NewDeviceReference(
		7, domaindcim.NonNullDeviceValue("edge-01"), "edge-01",
	)
	require.NoError(t, err)
	networkInterface, err := domaindcim.RestoreInterface(domaindcim.InterfaceState{
		ID: 41, Device: reference, Name: "Ethernet1", Label: "WAN",
		Type: "10gbase-sr", Enabled: false, MgmtOnly: true,
		MTU:         domaindcim.NullDeviceValue[uint32](),
		Speed:       domaindcim.NonNullDeviceValue(uint64(2147483647)),
		Duplex:      domaindcim.NonNullDeviceValue(""),
		Description: "Interface description",
		Created: shared.NewTimestamp(
			time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
		),
		LastUpdated: shared.NewTimestamp(
			time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC),
		),
		IPAddressCount: 3,
	})
	require.NoError(t, err)
	return networkInterface
}
