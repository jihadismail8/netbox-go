package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestTypedVRFListQueryPreservesPresenceAndSignedIDs(t *testing.T) {
	zero := uint32(0)
	offset := uint32(7)
	name := "Blue"
	rd := "65000:7"
	enforceUnique := false
	query := typedVRFListQuery(&ipamv1.ListVRFsRequest{
		Page: &typesv1.PageRequest{
			Limit: &zero, Offset: &offset, Query: vrfStringPointer("tenant"),
			Ordering: []string{"name", "rd"}, Id: []int64{-1, 0, 9},
		},
		Name: &name, Rd: &rd, EnforceUnique: &enforceUnique,
	})

	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumVRFPageLimit, query.EffectiveLimit())
	assert.Equal(t, uint32(7), query.Offset)
	assert.Equal(t, []int64{-1, 0, 9}, query.IDs)
	assert.Equal(t, []string{"Blue"}, query.Names)
	assert.Equal(t, []string{"65000:7"}, query.RDs)
	require.NotNil(t, query.EnforceUnique)
	assert.False(t, *query.EnforceUnique)
}

func vrfStringPointer(value string) *string { return &value }

func TestTypedVRFUpdateMaskCanClearNullableRD(t *testing.T) {
	command, err := typedVRFUpdateCommand(
		7,
		&ipamv1.VRFInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"rd"}},
	)
	require.NoError(t, err)
	assert.Equal(t, applicationipam.FieldNull, command.RD.State())
	assert.Equal(t, applicationipam.FieldOmitted, command.Name.State())

	_, err = typedVRFUpdateCommand(
		7,
		&ipamv1.VRFInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"name"}},
	)
	require.Error(t, err)
}

func TestTypedVRFProtoPreservesNullAndPresentBlankRD(t *testing.T) {
	nullVRF := grpcVRFFixture(t, 1, "Null", domainipam.NullRouteDistinguisher())
	assert.Nil(t, typedVRFProto(nullVRF).Rd)

	blank, err := domainipam.ParseRouteDistinguisher("")
	require.NoError(t, err)
	blankVRF := grpcVRFFixture(t, 2, "Blank", domainipam.NonNullRouteDistinguisher(blank))
	projected := typedVRFProto(blankVRF)
	require.NotNil(t, projected.Rd)
	assert.Empty(t, projected.Rd.Value)
	assert.Equal(t, "/api/ipam/vrfs/2/", projected.Url)
}

func TestVRFRPCHandlerUsesTypedServiceForListAndUpdate(t *testing.T) {
	spy := &grpcVRFServiceSpy{vrf: grpcVRFFixture(
		t, 8, "Blue", requiredNullableRD(t, "65000:8"),
	)}
	handler := NewVRFRPCHandler(spy)
	ctx := identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-vrf"})

	listResponse, err := handler.ListVRFs(ctx, &ipamv1.ListVRFsRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, spy.listCalls)
	assert.Equal(t, uint64(1), listResponse.Page.Count)
	require.Len(t, listResponse.Results, 1)
	assert.Equal(t, int64(8), listResponse.Results[0].Id)

	updateResponse, err := handler.UpdateVRF(ctx, &ipamv1.UpdateVRFRequest{
		Id: 8, Vrf: &ipamv1.VRFInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"rd"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, spy.updateCalls)
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.RD.State())
	assert.Equal(t, int64(8), updateResponse.Vrf.Id)
}

func TestIPAMServerVRFDispatchUsesTypedService(t *testing.T) {
	vrfs := &grpcVRFServiceSpy{}
	prefixes := &grpcPrefixServiceSpy{}
	addresses := &grpcIPAddressServiceSpy{}
	server := NewIPAMServer(vrfs, prefixes, addresses)
	ctx := identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-vrf"})

	response, err := server.ListVRFs(ctx, &ipamv1.ListVRFsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(0), response.Page.Count)
	require.Equal(t, 1, vrfs.listCalls)
	require.Zero(t, prefixes.listCalls)
	require.Zero(t, addresses.calls)
}

func TestVRFRPCHandlerRequiresAuthenticatedPrincipal(t *testing.T) {
	handler := NewVRFRPCHandler(&grpcVRFServiceSpy{})
	_, err := handler.ListVRFs(context.Background(), &ipamv1.ListVRFsRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func grpcVRFFixture(
	t *testing.T,
	id shared.ID,
	name string,
	rd domainipam.NullableRouteDistinguisher,
) *domainipam.VRF {
	t.Helper()
	stamp := shared.NewTimestamp(time.Date(2026, time.July, 22, 13, 0, 0, 0, time.UTC))
	vrf, err := domainipam.RestoreVRF(domainipam.VRFState{
		ID: id, Name: name, RD: rd, EnforceUnique: true,
		Description: "description", Comments: "comments",
		Created: stamp, LastUpdated: stamp, IPAddressCount: 2, PrefixCount: 3,
	})
	require.NoError(t, err)
	return vrf
}

func requiredNullableRD(t *testing.T, value string) domainipam.NullableRouteDistinguisher {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(value)
	require.NoError(t, err)
	return domainipam.NonNullRouteDistinguisher(rd)
}

type grpcVRFServiceSpy struct {
	listCalls     int
	updateCalls   int
	updateCommand applicationipam.UpdateVRFCommand
	vrf           *domainipam.VRF
}

func (spy *grpcVRFServiceSpy) ListVRFs(
	context.Context,
	identity.Principal,
	applicationipam.ListVRFsQuery,
) (applicationipam.VRFPage, error) {
	spy.listCalls++
	if spy.vrf == nil {
		return applicationipam.VRFPage{}, nil
	}
	return applicationipam.VRFPage{Count: 1, Results: []*domainipam.VRF{spy.vrf}}, nil
}

func (spy *grpcVRFServiceSpy) GetVRF(
	context.Context,
	identity.Principal,
	applicationipam.GetVRFQuery,
) (*domainipam.VRF, error) {
	return spy.vrf, nil
}

func (spy *grpcVRFServiceSpy) CreateVRF(
	context.Context,
	identity.Principal,
	applicationipam.CreateVRFCommand,
) (*domainipam.VRF, error) {
	return spy.vrf, nil
}

func (spy *grpcVRFServiceSpy) ReplaceVRF(
	context.Context,
	identity.Principal,
	applicationipam.ReplaceVRFCommand,
) (*domainipam.VRF, error) {
	return spy.vrf, nil
}

func (spy *grpcVRFServiceSpy) UpdateVRF(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.UpdateVRFCommand,
) (*domainipam.VRF, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.vrf, nil
}

func (*grpcVRFServiceSpy) DeleteVRF(
	context.Context,
	identity.Principal,
	applicationipam.DeleteVRFCommand,
) error {
	return nil
}
