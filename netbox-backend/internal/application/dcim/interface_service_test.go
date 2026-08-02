package dcim_test

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	appipam "netbox-go/internal/application/ipam"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestCreateInterfaceAppliesDefaultsAndPreservesNullableFields(t *testing.T) {
	service, repository, recorder, _, _ := newInterfaceApplicationService(
		t, authz.AllowAll{},
	)
	networkInterface, err := service.CreateInterface(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceCommand{
			Device: appdcim.FieldValue(shared.ID(7)),
			Name:   appdcim.FieldValue(" Ethernet1 "),
			Type:   appdcim.FieldValue("1000base-t"),
			Duplex: appdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)

	assert.Equal(t, shared.ID(41), networkInterface.ID())
	assert.Equal(t, "Ethernet1", networkInterface.Name())
	assert.Empty(t, networkInterface.Label())
	assert.True(t, networkInterface.Enabled())
	assert.False(t, networkInterface.MgmtOnly())
	assert.True(t, networkInterface.MTU().IsNull())
	assert.True(t, networkInterface.Speed().IsNull())
	duplex, present := networkInterface.Duplex().Get()
	require.True(t, present)
	assert.Empty(t, duplex)
	assert.Equal(t, 1, repository.createCalls)
	require.Len(t, recorder.changes, 1)
	assert.Equal(t, dcimdomain.InterfaceObjectType, recorder.changes[0].ObjectType)
	assert.IsType(t, dcimdomain.InterfaceSnapshot{}, recorder.changes[0].After)
}

func TestInterfacePUTPATCHAndNullSemantics(t *testing.T) {
	service, _, _, _, _ := newInterfaceApplicationService(t, authz.AllowAll{})
	networkInterface, err := service.CreateInterface(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceCommand{
			Device:      appdcim.FieldValue(shared.ID(7)),
			Name:        appdcim.FieldValue("Ethernet1"),
			Type:        appdcim.FieldValue("1000base-t"),
			Enabled:     appdcim.FieldValue(false),
			MgmtOnly:    appdcim.FieldValue(true),
			MTU:         appdcim.FieldValue(uint32(1500)),
			Speed:       appdcim.FieldValue(uint64(1_000_000)),
			Duplex:      appdcim.FieldValue("full"),
			Description: appdcim.FieldValue("old"),
		},
	)
	require.NoError(t, err)

	networkInterface, err = service.UpdateInterface(
		t.Context(), testPrincipal(), appdcim.UpdateInterfaceCommand{
			ID: networkInterface.ID(), MTU: appdcim.NullField[uint32](),
			Duplex: appdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	assert.True(t, networkInterface.MTU().IsNull())
	duplex, present := networkInterface.Duplex().Get()
	require.True(t, present)
	assert.Empty(t, duplex)
	assert.False(t, networkInterface.Enabled(), "omitted PATCH fields must be preserved")
	assert.True(t, networkInterface.MgmtOnly())

	networkInterface, err = service.ReplaceInterface(
		t.Context(), testPrincipal(), appdcim.ReplaceInterfaceCommand{
			ID: networkInterface.ID(),
			CreateInterfaceCommand: appdcim.CreateInterfaceCommand{
				Device: appdcim.FieldValue(shared.ID(7)),
				Name:   appdcim.FieldValue("Ethernet2"),
				Type:   appdcim.FieldValue("10gbase-sr"),
			},
		},
	)
	require.NoError(t, err)
	assert.True(t, networkInterface.Enabled())
	assert.False(t, networkInterface.MgmtOnly())
	assert.True(t, networkInterface.MTU().IsNull())
	assert.True(t, networkInterface.Speed().IsNull())
	assert.True(t, networkInterface.Duplex().IsNull())
	assert.Empty(t, networkInterface.Description())
}

func TestInterfaceRejectsMoveAndPinsListContract(t *testing.T) {
	service, repository, _, devices, _ := newInterfaceApplicationService(
		t, authz.PermissionAuthorizer{},
	)
	devices.references[8] = interfaceApplicationDeviceReference(t, 8, "edge-02")
	editor := identity.Principal{
		ID: 7, Username: "editor",
		Permissions: map[string]struct{}{
			"dcim.add_interface": {}, "dcim.change_interface": {},
			"dcim.view_interface": {},
		},
	}
	networkInterface, err := service.CreateInterface(
		t.Context(), editor, appdcim.CreateInterfaceCommand{
			Device: appdcim.FieldValue(shared.ID(7)),
			Name:   appdcim.FieldValue("Ethernet1"),
			Type:   appdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)
	_, err = service.UpdateInterface(
		t.Context(), editor, appdcim.UpdateInterfaceCommand{
			ID: networkInterface.ID(), Device: appdcim.FieldValue(shared.ID(8)),
		},
	)
	require.Error(t, err)
	assert.Equal(t, "device", shared.ViolationsOf(err)[0].Field)
	assert.Zero(t, repository.updateCalls)

	enabled := false
	mgmtOnly := true
	_, err = service.ListInterfaces(
		t.Context(), editor, appdcim.ListInterfacesQuery{
			LimitPresent: true, IDs: []int64{-1, 0, 41},
			DeviceIDs: []int64{-1, 7}, DeviceNames: []string{" edge-01 "},
			Names: []string{" Ethernet1 "}, Types: []string{"1000base-t"},
			Enabled: &enabled, MgmtOnly: &mgmtOnly,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, appdcim.MaximumInterfacePageLimit, repository.criteria.Limit)
	assert.Equal(t, []int64{-1, 7}, repository.criteria.DeviceIDs)
	assert.Equal(t, []string{"edge-01"}, repository.criteria.DeviceNames)
	assert.Equal(t, []string{"Ethernet1"}, repository.criteria.Names)
	assert.Equal(t, []dcimdomain.InterfaceType{"1000base-t"}, repository.criteria.Types)
	assert.Equal(t, []appdcim.InterfaceSort{
		{Field: appdcim.InterfaceSortDevice},
		{Field: appdcim.InterfaceSortName},
		{Field: appdcim.InterfaceSortID},
	}, repository.criteria.Ordering)
}

func TestDeleteInterfaceCascadesIPAddressAndRecordsChildBeforeParent(t *testing.T) {
	service, repository, recorder, _, cascade := newInterfaceApplicationService(
		t, authz.AllowAll{},
	)
	networkInterface, err := service.CreateInterface(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceCommand{
			Device: appdcim.FieldValue(shared.ID(7)),
			Name:   appdcim.FieldValue("Ethernet1"),
			Type:   appdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)
	recorder.changes = nil
	cascade.changes = []appipam.IPAddressCascadeChange{{
		ObjectType: "ipam.ipaddress", ID: 91, Representation: "192.0.2.1/24",
		Before: interfaceCascadeSnapshot{},
	}}

	err = service.DeleteInterface(
		t.Context(), testPrincipal(),
		appdcim.DeleteInterfaceCommand{ID: networkInterface.ID()},
	)
	require.NoError(t, err)
	assert.Equal(t, []shared.ID{networkInterface.ID()}, cascade.interfaceIDs)
	assert.Equal(t, 1, repository.deleteCalls)
	require.Len(t, recorder.changes, 2)
	assert.Equal(t, "ipam.ipaddress", recorder.changes[0].ObjectType)
	assert.Equal(t, shared.ID(91), recorder.changes[0].ObjectID)
	assert.Equal(t, dcimdomain.InterfaceObjectType, recorder.changes[1].ObjectType)
	assert.Equal(t, networkInterface.ID(), recorder.changes[1].ObjectID)
}

func TestInterfaceDeviceCascadeReturnsIPAddressBeforeInterfaceWithoutRecording(t *testing.T) {
	service, _, recorder, _, cascade := newInterfaceApplicationService(
		t, authz.AllowAll{},
	)
	networkInterface, err := service.CreateInterface(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceCommand{
			Device: appdcim.FieldValue(shared.ID(7)),
			Name:   appdcim.FieldValue("Ethernet1"),
			Type:   appdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)
	recorder.changes = nil
	cascade.changes = []appipam.IPAddressCascadeChange{{
		ObjectType: "ipam.ipaddress", ID: 91, Representation: "192.0.2.1/24",
		Before: interfaceCascadeSnapshot{},
	}}

	changes, err := service.DeleteForDevice(t.Context(), 7, updatedAt)
	require.NoError(t, err)
	require.Len(t, changes, 2)
	assert.Equal(t, "ipam.ipaddress", changes[0].ObjectType)
	assert.Equal(t, shared.ID(91), changes[0].ID)
	assert.Equal(t, dcimdomain.InterfaceObjectType, changes[1].ObjectType)
	assert.Equal(t, networkInterface.ID(), changes[1].ID)
	assert.Empty(t, recorder.changes, "Device owns cascade audit recording")
}

func TestInterfaceListAppliesVisibilityBeforePaginationWithoutCompleteScope(t *testing.T) {
	service, repository, _, _, _ := newInterfaceApplicationService(
		t, interfaceVisibilityAuthorizer{hiddenID: 42},
	)
	networkInterface, err := service.CreateInterface(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceCommand{
			Device: appdcim.FieldValue(shared.ID(7)),
			Name:   appdcim.FieldValue("Ethernet1"),
			Type:   appdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)
	for id, name := range map[shared.ID]string{42: "Ethernet2", 43: "Ethernet3"} {
		state := networkInterface.State()
		state.ID = id
		state.Name = name
		repository.interfaces[id] = state
	}

	page, err := service.ListInterfaces(
		t.Context(), testPrincipal(),
		appdcim.ListInterfacesQuery{Limit: 1, Offset: 1},
	)
	require.NoError(t, err)
	assert.True(t, repository.criteria.DeferPagination)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(43), page.Results[0].ID())
}

func newInterfaceApplicationService(
	t *testing.T,
	authorizer authz.ResourceAuthorizer,
) (
	*appdcim.InterfaceService,
	*interfaceApplicationRepository,
	*interfaceApplicationRecorder,
	*interfaceApplicationDeviceReader,
	*interfaceApplicationIPCascade,
) {
	t.Helper()
	repository := &interfaceApplicationRepository{
		interfaces: make(map[shared.ID]dcimdomain.InterfaceState),
	}
	devices := &interfaceApplicationDeviceReader{
		references: map[shared.ID]dcimdomain.DeviceReference{
			7: interfaceApplicationDeviceReference(t, 7, "edge-01"),
		},
	}
	recorder := &interfaceApplicationRecorder{}
	cascade := &interfaceApplicationIPCascade{}
	service, err := appdcim.NewInterfaceService(
		repository, devices, cascade, interfaceApplicationUnitOfWork{},
		recorder, authorizer, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service, repository, recorder, devices, cascade
}

func interfaceApplicationDeviceReference(
	t *testing.T,
	id shared.ID,
	name string,
) dcimdomain.DeviceReference {
	t.Helper()
	reference, err := dcimdomain.NewDeviceReference(
		id, dcimdomain.NonNullDeviceValue(name), name,
	)
	require.NoError(t, err)
	return reference
}

type interfaceApplicationDeviceReader struct {
	references map[shared.ID]dcimdomain.DeviceReference
}

func (reader *interfaceApplicationDeviceReader) GetDeviceReference(
	_ context.Context,
	id shared.ID,
) (dcimdomain.DeviceReference, error) {
	reference, present := reader.references[id]
	if !present {
		return dcimdomain.DeviceReference{}, shared.NewValidationError(
			shared.FieldViolation{
				Field: "device", Reason: "invalid_choice",
				Description: "Select a valid choice.",
			},
		)
	}
	return reference, nil
}

type interfaceApplicationRepository struct {
	interfaces  map[shared.ID]dcimdomain.InterfaceState
	criteria    appdcim.InterfaceListCriteria
	listCalls   int
	createCalls int
	updateCalls int
	deleteCalls int
}

func (repository *interfaceApplicationRepository) List(
	_ context.Context,
	criteria appdcim.InterfaceListCriteria,
) (appdcim.InterfacePage, error) {
	repository.listCalls++
	repository.criteria = criteria
	ids := make([]int, 0, len(repository.interfaces))
	for id := range repository.interfaces {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	results := make([]*dcimdomain.Interface, 0, len(ids))
	for _, rawID := range ids {
		networkInterface, err := repository.restore(shared.ID(rawID))
		if err != nil {
			return appdcim.InterfacePage{}, err
		}
		results = append(results, networkInterface)
	}
	count := uint64(len(results))
	if !criteria.DeferPagination {
		start := min(int(criteria.Offset), len(results))
		end := min(start+int(criteria.Limit), len(results))
		results = results[start:end]
	}
	return appdcim.InterfacePage{Count: count, Results: results}, nil
}

func (repository *interfaceApplicationRepository) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Interface, error) {
	return repository.restore(id)
}

func (repository *interfaceApplicationRepository) GetForUpdate(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Interface, error) {
	return repository.restore(id)
}

func (repository *interfaceApplicationRepository) ListForDeviceForUpdate(
	_ context.Context,
	deviceID shared.ID,
) ([]*dcimdomain.Interface, error) {
	var results []*dcimdomain.Interface
	for id, state := range repository.interfaces {
		if state.Device.ID() != deviceID {
			continue
		}
		networkInterface, err := repository.restore(id)
		if err != nil {
			return nil, err
		}
		results = append(results, networkInterface)
	}
	return results, nil
}

func (repository *interfaceApplicationRepository) restore(
	id shared.ID,
) (*dcimdomain.Interface, error) {
	state, present := repository.interfaces[id]
	if !present {
		return nil, shared.NotFound("Interface", id)
	}
	return dcimdomain.RestoreInterface(state)
}

func (repository *interfaceApplicationRepository) Create(
	_ context.Context,
	networkInterface *dcimdomain.Interface,
) error {
	repository.createCalls++
	if err := networkInterface.AssignID(41); err != nil {
		return err
	}
	repository.interfaces[networkInterface.ID()] = networkInterface.State()
	return nil
}

func (repository *interfaceApplicationRepository) Update(
	_ context.Context,
	networkInterface *dcimdomain.Interface,
) error {
	repository.updateCalls++
	repository.interfaces[networkInterface.ID()] = networkInterface.State()
	return nil
}

func (repository *interfaceApplicationRepository) Delete(
	_ context.Context,
	networkInterface *dcimdomain.Interface,
) error {
	repository.deleteCalls++
	delete(repository.interfaces, networkInterface.ID())
	return nil
}

type interfaceApplicationIPCascade struct {
	changes      []appipam.IPAddressCascadeChange
	interfaceIDs []shared.ID
}

func (cascade *interfaceApplicationIPCascade) DeleteAssignedToInterface(
	_ context.Context,
	interfaceID shared.ID,
	_ shared.Timestamp,
) ([]appipam.IPAddressCascadeChange, error) {
	cascade.interfaceIDs = append(cascade.interfaceIDs, interfaceID)
	return append([]appipam.IPAddressCascadeChange(nil), cascade.changes...), nil
}

type interfaceApplicationUnitOfWork struct{}

func (interfaceApplicationUnitOfWork) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	return operation(ctx)
}

type interfaceApplicationRecorder struct {
	changes []changelog.Change
}

func (recorder *interfaceApplicationRecorder) Record(
	_ context.Context,
	change changelog.Change,
) error {
	recorder.changes = append(recorder.changes, change)
	return nil
}

type interfaceCascadeSnapshot struct{}

func (interfaceCascadeSnapshot) ObjectType() string { return "ipam.ipaddress" }
func (snapshot interfaceCascadeSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	return snapshot
}

type interfaceVisibilityAuthorizer struct{ hiddenID int64 }

func (authorizer interfaceVisibilityAuthorizer) AuthorizeResource(
	_ context.Context,
	principal identity.Principal,
	_ authz.Action,
	_ authz.ResourceType,
	object *authz.Object,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	if object != nil && object.ID == authorizer.hiddenID {
		return shared.NewError(
			shared.ErrorReasonForbidden,
			"You do not have permission to perform this action.",
		)
	}
	return nil
}
