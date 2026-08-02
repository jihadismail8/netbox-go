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
	"google.golang.org/protobuf/types/known/wrapperspb"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestTypedPrefixListQueryPreservesPresenceSignedIDsAndNetworkFilters(t *testing.T) {
	zero := uint32(0)
	offset := uint32(7)
	vrfID := int64(-9)
	vrfRD := "65000:9"
	prefix := "10.0.0.0/8"
	family := uint32(4)
	statusValue := "reserved"
	within := "10.0.0.0/8"
	withinInclude := "10.1.0.0/16"
	contains := "10.1.2.3"
	query := typedPrefixListQuery(&ipamv1.ListPrefixesRequest{
		Page: &typesv1.PageRequest{
			Limit: &zero, Offset: &offset, Query: prefixStringPointer("tenant"),
			Ordering: []string{"vrf", "-prefix"}, Id: []int64{-1, 0, 11},
		},
		VrfId: &vrfID, VrfRd: &vrfRD, Prefix: &prefix, Family: &family, Status: &statusValue,
		Within: &within, WithinInclude: &withinInclude, Contains: &contains,
	})

	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumPrefixPageLimit, query.EffectiveLimit())
	assert.Equal(t, uint32(7), query.Offset)
	assert.Equal(t, []int64{-1, 0, 11}, query.IDs)
	assert.Equal(t, []int64{-9}, query.VRFIDs)
	assert.Equal(t, []string{"65000:9"}, query.VRFRDs)
	assert.Equal(t, []string{"10.0.0.0/8"}, query.Prefixes)
	assert.Equal(t, int64(4), *query.Family)
	assert.Equal(t, []string{"reserved"}, query.Statuses)
	assert.Equal(t, within, *query.Within)
	assert.Equal(t, withinInclude, *query.WithinInclude)
	assert.Equal(t, contains, *query.Contains)
	assert.Equal(t, []string{"vrf", "-prefix"}, query.Ordering)
}

func prefixStringPointer(value string) *string { return &value }

func TestTypedPrefixUpdateMaskCanClearNullableVRFAndRequiresExplicitValues(t *testing.T) {
	command, err := typedPrefixUpdateCommand(
		17,
		&ipamv1.PrefixInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"vrf"}},
	)
	require.NoError(t, err)
	assert.Equal(t, applicationipam.FieldNull, command.VRF.State())
	assert.Equal(t, applicationipam.FieldOmitted, command.Prefix.State())

	_, err = typedPrefixUpdateCommand(
		17,
		&ipamv1.PrefixInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"prefix"}},
	)
	require.Error(t, err)

	_, err = typedPrefixUpdateCommand(
		17,
		&ipamv1.PrefixInput{Vrf: wrapperspb.Int64(7)},
		&fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
}

func TestTypedPrefixProtoPreservesNullableVRFAndHierarchy(t *testing.T) {
	global := grpcPrefixFixture(t, 1, "2001:db8::/32", domainipam.NullVRFReference(), 4, 2)
	assert.Nil(t, typedPrefixProto(global).VrfId)

	vrf := grpcPrefixVRFReference(t, 7, "Blue", "65000:7", true)
	projected := typedPrefixProto(grpcPrefixFixture(
		t, 17, "10.0.0.0/8", domainipam.NonNullVRFReference(vrf), 3, 1,
	))
	require.NotNil(t, projected.VrfId)
	assert.Equal(t, int64(7), projected.VrfId.Value)
	assert.Equal(t, "/api/ipam/prefixes/17/", projected.Url)
	assert.Equal(t, uint32(4), projected.Family)
	assert.Equal(t, uint64(3), projected.Children)
	assert.Equal(t, uint32(1), projected.Depth)
}

func TestPrefixRPCHandlerUsesTypedServiceForListAndUpdate(t *testing.T) {
	vrf := grpcPrefixVRFReference(t, 7, "Blue", "65000:7", true)
	spy := &grpcPrefixServiceSpy{prefix: grpcPrefixFixture(
		t, 17, "10.0.0.0/8", domainipam.NonNullVRFReference(vrf), 2, 1,
	)}
	handler := NewPrefixRPCHandler(spy)
	ctx := identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-prefix"})

	listResponse, err := handler.ListPrefixes(ctx, &ipamv1.ListPrefixesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, spy.listCalls)
	assert.Equal(t, uint64(1), listResponse.Page.Count)
	require.Len(t, listResponse.Results, 1)
	assert.Equal(t, int64(17), listResponse.Results[0].Id)

	updateResponse, err := handler.UpdatePrefix(ctx, &ipamv1.UpdatePrefixRequest{
		Id: 17, Prefix: &ipamv1.PrefixInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"vrf"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, spy.updateCalls)
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.VRF.State())
	assert.Equal(t, int64(17), updateResponse.Prefix.Id)
}

func TestIPAMServerPrefixDispatchUsesTypedService(t *testing.T) {
	vrfs := &grpcVRFServiceSpy{}
	prefixes := &grpcPrefixServiceSpy{}
	addresses := &grpcIPAddressServiceSpy{}
	server := NewIPAMServer(vrfs, prefixes, addresses)
	ctx := identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-prefix"})

	response, err := server.ListPrefixes(ctx, &ipamv1.ListPrefixesRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(0), response.Page.Count)
	require.Equal(t, 1, prefixes.listCalls)
	require.Zero(t, vrfs.listCalls)
	require.Zero(t, addresses.calls)
}

func TestPrefixRPCHandlerRequiresAuthenticatedPrincipal(t *testing.T) {
	handler := NewPrefixRPCHandler(&grpcPrefixServiceSpy{})
	_, err := handler.ListPrefixes(context.Background(), &ipamv1.ListPrefixesRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func grpcPrefixFixture(
	t *testing.T,
	id shared.ID,
	network string,
	vrf domainipam.NullableVRFReference,
	children uint64,
	depth uint32,
) *domainipam.Prefix {
	t.Helper()
	stamp := shared.NewTimestamp(time.Date(2026, time.July, 22, 15, 0, 0, 0, time.UTC))
	prefix, err := domainipam.RestorePrefix(domainipam.PrefixState{
		ID: id, Prefix: network, VRF: vrf, Status: domainipam.PrefixStatusReserved.String(),
		IsPool: true, MarkUtilized: true, Description: "description", Comments: "comments",
		Created: stamp, LastUpdated: stamp, Children: children, Depth: depth,
	})
	require.NoError(t, err)
	return prefix
}

func grpcPrefixVRFReference(
	t *testing.T,
	id shared.ID,
	name string,
	rdValue string,
	enforceUnique bool,
) domainipam.VRFReference {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	reference, err := domainipam.NewVRFReference(
		id, name, domainipam.NonNullRouteDistinguisher(rd), enforceUnique,
	)
	require.NoError(t, err)
	return reference
}

type grpcPrefixServiceSpy struct {
	listCalls     int
	updateCalls   int
	updateCommand applicationipam.UpdatePrefixCommand
	prefix        *domainipam.Prefix
}

func (spy *grpcPrefixServiceSpy) ListPrefixes(
	context.Context,
	identity.Principal,
	applicationipam.ListPrefixesQuery,
) (applicationipam.PrefixPage, error) {
	spy.listCalls++
	if spy.prefix == nil {
		return applicationipam.PrefixPage{}, nil
	}
	return applicationipam.PrefixPage{Count: 1, Results: []*domainipam.Prefix{spy.prefix}}, nil
}

func (spy *grpcPrefixServiceSpy) GetPrefix(
	context.Context,
	identity.Principal,
	applicationipam.GetPrefixQuery,
) (*domainipam.Prefix, error) {
	return spy.prefix, nil
}

func (spy *grpcPrefixServiceSpy) CreatePrefix(
	context.Context,
	identity.Principal,
	applicationipam.CreatePrefixCommand,
) (*domainipam.Prefix, error) {
	return spy.prefix, nil
}

func (spy *grpcPrefixServiceSpy) ReplacePrefix(
	context.Context,
	identity.Principal,
	applicationipam.ReplacePrefixCommand,
) (*domainipam.Prefix, error) {
	return spy.prefix, nil
}

func (spy *grpcPrefixServiceSpy) UpdatePrefix(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.UpdatePrefixCommand,
) (*domainipam.Prefix, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.prefix, nil
}

func (*grpcPrefixServiceSpy) DeletePrefix(
	context.Context,
	identity.Principal,
	applicationipam.DeletePrefixCommand,
) error {
	return nil
}
