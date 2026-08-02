package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestTypedInterfaceTemplateRPCPreservesFiltersPresenceAndProjection(t *testing.T) {
	spy := &grpcInterfaceTemplateServiceSpy{
		template: grpcInterfaceTemplateFixture(t),
	}
	handler := NewInterfaceTemplateRPCHandler(spy)
	ctx := identity.WithPrincipal(
		t.Context(),
		identity.Principal{ID: 1, Username: "interface-template-grpc", IsSuperuser: true},
	)
	disabled := false
	management := true
	limit := uint32(0)
	offset := uint32(3)
	deviceTypeID := int64(7)
	name := "Ethernet1"
	interfaceType := "10gbase-sr"
	query := "uplink"

	list, err := handler.ListInterfaceTemplates(
		ctx,
		&dcimv1.ListInterfaceTemplatesRequest{
			Page: &typesv1.PageRequest{
				Limit: &limit, Offset: &offset, Query: &query,
				Ordering: []string{"-type", "name"}, Id: []int64{-1, 41},
			},
			DeviceTypeId: &deviceTypeID, Name: &name, Type: &interfaceType,
			Enabled: &disabled, MgmtOnly: &management,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, list.Page)
	assert.Equal(t, uint64(0), list.Page.Count)
	assert.True(t, spy.listQuery.LimitPresent)
	assert.Equal(t, uint32(0), spy.listQuery.Limit)
	assert.Equal(t, uint32(3), spy.listQuery.Offset)
	assert.Equal(t, []int64{-1, 41}, spy.listQuery.IDs)
	assert.Equal(t, []int64{7}, spy.listQuery.DeviceTypeIDs)
	assert.Equal(t, []string{"Ethernet1"}, spy.listQuery.Names)
	assert.Equal(t, []string{"10gbase-sr"}, spy.listQuery.Types)
	require.NotNil(t, spy.listQuery.Enabled)
	assert.False(t, *spy.listQuery.Enabled)
	require.NotNil(t, spy.listQuery.MgmtOnly)
	assert.True(t, *spy.listQuery.MgmtOnly)

	create, err := handler.CreateInterfaceTemplate(
		ctx,
		&dcimv1.CreateInterfaceTemplateRequest{
			InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
				DeviceType: &deviceTypeID, Name: &name, Type: &interfaceType,
				Enabled: &disabled,
			},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, create.InterfaceTemplate)
	assert.Equal(t, int64(41), create.InterfaceTemplate.Id)
	assert.Equal(t, "Ethernet1 (WAN)", create.InterfaceTemplate.Display)
	assert.Equal(t, "10gbase-sr", create.InterfaceTemplate.Type)
	require.NotNil(t, create.InterfaceTemplate.DeviceType)
	assert.Equal(t, "Router", create.InterfaceTemplate.DeviceType.Display)
	createdEnabled, present := spy.createCommand.Enabled.Get()
	require.True(t, present)
	assert.False(t, createdEnabled)
	assert.Equal(t, applicationdcim.FieldOmitted, spy.createCommand.MgmtOnly.State())

	label := ""
	otherDeviceTypeID := int64(9)
	update, err := handler.UpdateInterfaceTemplate(
		ctx,
		&dcimv1.UpdateInterfaceTemplateRequest{
			Id: 41,
			InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
				DeviceType: &otherDeviceTypeID, Label: &label,
				Enabled: &disabled, MgmtOnly: &management,
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"label", "enabled"}},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, update.InterfaceTemplate)
	assert.Equal(t, applicationdcim.FieldPresent, spy.updateCommand.Label.State())
	assert.Equal(t, applicationdcim.FieldPresent, spy.updateCommand.Enabled.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.DeviceType.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.MgmtOnly.State())

	_, err = handler.UpdateInterfaceTemplate(
		ctx,
		&dcimv1.UpdateInterfaceTemplateRequest{
			Id: 41, InterfaceTemplate: &dcimv1.InterfaceTemplateInput{Label: &label},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"enabled"}},
		},
	)
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, 1, spy.updateCalls)
}

func grpcInterfaceTemplateFixture(t *testing.T) *domaindcim.InterfaceTemplate {
	t.Helper()
	reference, err := domaindcim.NewDeviceTypeReference(7, "Router", "router")
	require.NoError(t, err)
	template, err := domaindcim.RestoreInterfaceTemplate(
		domaindcim.InterfaceTemplateState{
			ID: 41, DeviceType: reference, Name: "Ethernet1", Label: "WAN",
			Type: "10gbase-sr", Enabled: false, MgmtOnly: true,
			Description: "Template description",
			Created: shared.NewTimestamp(
				time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
			),
			LastUpdated: shared.NewTimestamp(
				time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC),
			),
		},
	)
	require.NoError(t, err)
	return template
}
