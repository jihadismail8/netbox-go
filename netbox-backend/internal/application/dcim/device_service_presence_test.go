package dcim_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestReplaceDeviceAppliesPinnedOmissionDefaultsAndPreservesOptionalState(t *testing.T) {
	t.Parallel()

	harness := newDevicePresenceHarness(t, nil)
	seedDevicePresenceState(t, harness.backend, harness.dependencies)
	before := harness.backend.state.devices[41]

	device, err := harness.service.ReplaceDevice(
		t.Context(), testPrincipal(), appdcim.ReplaceDeviceCommand{
			ID: 41,
			CreateDeviceCommand: appdcim.CreateDeviceCommand{
				DeviceType: appdcim.FieldValue(shared.ID(20)),
				Role:       appdcim.FieldValue(shared.ID(11)),
				Site:       appdcim.FieldValue(shared.ID(12)),
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, shared.ID(20), device.DeviceType().ID())
	assert.Equal(t, shared.ID(11), device.Role().ID)
	assert.Equal(t, shared.ID(12), device.Site().ID())
	assert.True(t, device.Rack().IsNull())
	assert.True(t, device.Position().IsNull())
	assert.Empty(t, device.Face())
	assert.Equal(t, devicePresenceNullableString(t, before.Name), devicePresenceNullableString(t, device.Name()))
	assert.Equal(t, before.Status, device.Status().String())
	assert.Equal(t, before.Serial, device.Serial())
	assert.Equal(t, devicePresenceNullableString(t, before.AssetTag), devicePresenceNullableString(t, device.AssetTag()))
	assert.Equal(t, dcimdomain.DeviceAirflowRearToFront, devicePresenceAirflow(t, device.Airflow()))
	assert.NotEqual(t, dcimdomain.DeviceAirflowPassive, devicePresenceAirflow(t, device.Airflow()))
	assert.Equal(t, before.Description, device.Description())
	assert.Equal(t, before.Comments, device.Comments())
	assert.Equal(t, before.Created, device.Created())
	assert.Equal(t, updatedAt, device.LastUpdated())
	assert.Equal(t, 1, harness.backend.transactionCalls)
	assert.Equal(t, 1, harness.repository.getForUpdateCalls)
	assert.Equal(t, 1, harness.repository.updateCalls)
	assert.Zero(t, harness.repository.occupantCalls)
	assert.Zero(t, harness.templates.listCalls)
	assert.Zero(t, harness.interfaces.createCalls)
	assert.Equal(t, 1, harness.recorder.calls)
	require.Len(t, harness.backend.state.changes, 1)
	assert.Equal(t, devicePresenceSnapshot(t, before), harness.backend.state.changes[0].Before)
	assert.Equal(t, device.Snapshot(), harness.backend.state.changes[0].After)
}

func TestDeviceScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	invalid := appdcim.CreateDeviceCommand{
		DeviceType: appdcim.NullField[shared.ID](),
		Role:       appdcim.FieldValue(shared.ID(404)),
		Name:       appdcim.FieldValue(strings.Repeat("n", dcimdomain.DeviceNameMaxLength+1)),
		Site:       appdcim.FieldValue(shared.ID(405)),
		Rack:       appdcim.FieldValue(shared.ID(0)),
		Position:   appdcim.NullField[string](),
		Face:       appdcim.FieldValue(" front "),
		Status:     appdcim.FieldValue(" active "),
		Serial:     appdcim.FieldValue(strings.Repeat("s", dcimdomain.DeviceSerialMaxLength+1)),
		AssetTag:   appdcim.FieldValue(strings.Repeat("a", dcimdomain.DeviceAssetTagMaxLength+1)),
		Airflow:    appdcim.FieldValue(" front-to-rear "),
		Description: appdcim.FieldValue(
			strings.Repeat("d", dcimdomain.DeviceDescriptionMaxLength+1),
		),
		Comments: appdcim.NullField[string](),
	}
	expectedFields := []string{
		"device_type", "role", "name", "site", "rack", "face", "status",
		"serial", "asset_tag", "airflow", "description", "comments",
	}

	for _, test := range []struct {
		name   string
		seed   bool
		mutate func(*appdcim.DeviceService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.DeviceService) error {
				_, err := service.CreateDevice(t.Context(), testPrincipal(), invalid)
				return err
			},
		},
		{
			name: "PUT", seed: true,
			mutate: func(service *appdcim.DeviceService) error {
				_, err := service.ReplaceDevice(t.Context(), testPrincipal(), appdcim.ReplaceDeviceCommand{
					ID: 41, CreateDeviceCommand: invalid,
				})
				return err
			},
		},
		{
			name: "PATCH", seed: true,
			mutate: func(service *appdcim.DeviceService) error {
				_, err := service.UpdateDevice(
					t.Context(), testPrincipal(), deviceUpdateFromCreate(41, invalid),
				)
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			harness := newDevicePresenceHarness(t, nil)
			if test.seed {
				seedDevicePresenceState(t, harness.backend, harness.dependencies)
			}
			before := harness.backend.state.clone()

			err := test.mutate(harness.service)
			require.Error(t, err)
			assert.Equal(t, expectedFields, devicePresenceViolationFields(err))
			assert.Equal(t, before, harness.backend.state)
			assert.Equal(t, 1, harness.backend.transactionCalls)
			if test.seed {
				assert.Equal(t, 1, harness.repository.getForUpdateCalls)
				require.NotNil(t, harness.repository.lastLoaded)
				assert.Equal(t, before.devices[41], harness.repository.lastLoaded.State())
			} else {
				assert.Zero(t, harness.repository.getForUpdateCalls)
			}
			assert.Zero(t, harness.repository.createCalls)
			assert.Zero(t, harness.repository.updateCalls)
			assert.Zero(t, harness.repository.occupantCalls)
			assert.Zero(t, harness.templates.listCalls)
			assert.Zero(t, harness.interfaces.createCalls)
			assert.Zero(t, harness.recorder.calls)
			assert.Empty(t, harness.backend.state.changes)
		})
	}

	t.Run("unknown relationships aggregate without mutation", func(t *testing.T) {
		harness := newDevicePresenceHarness(t, nil)
		seedDevicePresenceState(t, harness.backend, harness.dependencies)
		before := harness.backend.state.clone()

		_, err := harness.service.UpdateDevice(
			t.Context(), testPrincipal(), appdcim.UpdateDeviceCommand{
				ID: 41, DeviceType: appdcim.FieldValue(shared.ID(404)),
				Role: appdcim.FieldValue(shared.ID(405)), Site: appdcim.FieldValue(shared.ID(406)),
				Rack: appdcim.FieldValue(shared.ID(407)), Description: appdcim.FieldValue("valid sibling"),
			},
		)
		require.Error(t, err)
		assert.Equal(t, []string{"device_type", "role", "site", "rack"}, devicePresenceViolationFields(err))
		assert.Equal(t, before, harness.backend.state)
		require.NotNil(t, harness.repository.lastLoaded)
		assert.Equal(t, before.devices[41], harness.repository.lastLoaded.State())
		assert.Zero(t, harness.repository.updateCalls)
		assert.Zero(t, harness.repository.occupantCalls)
		assert.Zero(t, harness.recorder.calls)
	})

	t.Run("repository failure rolls back Device and object change", func(t *testing.T) {
		writeFailure := errors.New("forced Device repository failure")
		harness := newDevicePresenceHarness(t, nil)
		seedDevicePresenceState(t, harness.backend, harness.dependencies)
		harness.repository.updateErr = writeFailure
		before := harness.backend.state.clone()

		_, err := harness.service.UpdateDevice(
			t.Context(), testPrincipal(), appdcim.UpdateDeviceCommand{
				ID: 41, Description: appdcim.FieldValue("must roll back"),
			},
		)
		require.ErrorIs(t, err, writeFailure)
		assert.Equal(t, before, harness.backend.state)
		assert.Equal(t, 1, harness.repository.updateCalls)
		assert.Equal(t, 1, harness.repository.occupantCalls)
		assert.Zero(t, harness.recorder.calls)
		assert.Empty(t, harness.backend.state.changes)
	})

	t.Run("recorder failure rolls back Device and object change", func(t *testing.T) {
		recorderFailure := errors.New("forced Device recorder failure")
		harness := newDevicePresenceHarness(t, recorderFailure)
		seedDevicePresenceState(t, harness.backend, harness.dependencies)
		before := harness.backend.state.clone()

		_, err := harness.service.UpdateDevice(
			t.Context(), testPrincipal(), appdcim.UpdateDeviceCommand{
				ID: 41, Description: appdcim.FieldValue("must roll back"),
			},
		)
		require.ErrorIs(t, err, recorderFailure)
		assert.Equal(t, before, harness.backend.state)
		assert.Equal(t, 1, harness.repository.updateCalls)
		assert.Equal(t, 1, harness.repository.occupantCalls)
		assert.Equal(t, 1, harness.recorder.calls)
		assert.Empty(t, harness.backend.state.changes)
	})
}

func deviceUpdateFromCreate(
	id shared.ID,
	command appdcim.CreateDeviceCommand,
) appdcim.UpdateDeviceCommand {
	return appdcim.UpdateDeviceCommand{
		ID: id, DeviceType: command.DeviceType, Role: command.Role, Name: command.Name,
		Site: command.Site, Rack: command.Rack, Position: command.Position,
		Face: command.Face, Status: command.Status, Serial: command.Serial,
		AssetTag: command.AssetTag, Airflow: command.Airflow,
		Description: command.Description, Comments: command.Comments,
	}
}

type devicePresenceTransactionKey struct{}

type devicePresenceState struct {
	nextID  shared.ID
	devices map[shared.ID]dcimdomain.DeviceState
	changes []changelog.Change
}

func (state devicePresenceState) clone() devicePresenceState {
	cloned := devicePresenceState{
		nextID: state.nextID, devices: make(map[shared.ID]dcimdomain.DeviceState, len(state.devices)),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, device := range state.devices {
		cloned.devices[id] = device
	}
	return cloned
}

type devicePresenceBackend struct {
	state            devicePresenceState
	transactionCalls int
}

func newDevicePresenceBackend() *devicePresenceBackend {
	return &devicePresenceBackend{state: devicePresenceState{
		nextID: 42, devices: make(map[shared.ID]dcimdomain.DeviceState),
	}}
}

func (backend *devicePresenceBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, devicePresenceTransactionKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *devicePresenceBackend) stateFor(ctx context.Context) (*devicePresenceState, bool) {
	state, transactional := ctx.Value(devicePresenceTransactionKey{}).(*devicePresenceState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type devicePresenceRepository struct {
	backend           *devicePresenceBackend
	createCalls       int
	updateCalls       int
	getForUpdateCalls int
	occupantCalls     int
	updateErr         error
	lastLoaded        *dcimdomain.Device
}

func (*devicePresenceRepository) List(
	context.Context,
	appdcim.DeviceListCriteria,
) (appdcim.DevicePage, error) {
	return appdcim.DevicePage{}, nil
}

func (repository *devicePresenceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.Device, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.devices[id]
	if !ok {
		return nil, shared.NotFound("Device", id)
	}
	return dcimdomain.RestoreDevice(persisted)
}

func (repository *devicePresenceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.Device, error) {
	repository.getForUpdateCalls++
	device, err := repository.Get(ctx, id)
	repository.lastLoaded = device
	return device, err
}

func (repository *devicePresenceRepository) GetDeviceReference(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.DeviceReference, error) {
	device, err := repository.Get(ctx, id)
	if err != nil {
		return dcimdomain.DeviceReference{}, err
	}
	return dcimdomain.NewDeviceReference(device.ID(), device.Name(), device.Display())
}

func (repository *devicePresenceRepository) Create(
	ctx context.Context,
	device *dcimdomain.Device,
) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Device created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := device.AssignID(id); err != nil {
		return err
	}
	state.devices[id] = device.State()
	return nil
}

func (repository *devicePresenceRepository) Update(
	ctx context.Context,
	device *dcimdomain.Device,
) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Device updated outside transaction")
	}
	state.devices[device.ID()] = device.State()
	return repository.updateErr
}

func (repository *devicePresenceRepository) Delete(
	ctx context.Context,
	device *dcimdomain.Device,
) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Device deleted outside transaction")
	}
	delete(state.devices, device.ID())
	return nil
}

func (repository *devicePresenceRepository) ListRackOccupantsForUpdate(
	context.Context,
	shared.ID,
	shared.ID,
) ([]appdcim.DeviceRackOccupant, error) {
	repository.occupantCalls++
	return nil, nil
}

type devicePresenceDeviceTypeReader struct {
	deviceTypes map[shared.ID]*dcimdomain.DeviceType
}

func (reader devicePresenceDeviceTypeReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	if deviceType := reader.deviceTypes[id]; deviceType != nil {
		return deviceType, nil
	}
	return nil, shared.NotFound("DeviceType", id)
}

type devicePresenceRoleReader struct {
	roles map[shared.ID]*dcimdomain.DeviceRole
}

func (reader devicePresenceRoleReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.DeviceRole, error) {
	if role := reader.roles[id]; role != nil {
		return role, nil
	}
	return nil, shared.NotFound("DeviceRole", id)
}

type devicePresenceSiteReader struct {
	sites map[shared.ID]*dcimdomain.Site
}

func (reader devicePresenceSiteReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Site, error) {
	if site := reader.sites[id]; site != nil {
		return site, nil
	}
	return nil, shared.NotFound("Site", id)
}

type devicePresenceRackReader struct {
	racks map[shared.ID]*dcimdomain.Rack
}

func (reader devicePresenceRackReader) GetForUpdate(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Rack, error) {
	if rack := reader.racks[id]; rack != nil {
		return rack, nil
	}
	return nil, shared.NotFound("Rack", id)
}

type devicePresenceTemplateReader struct{ listCalls int }

func (reader *devicePresenceTemplateReader) List(
	context.Context,
	appdcim.InterfaceTemplateListCriteria,
) (appdcim.InterfaceTemplatePage, error) {
	reader.listCalls++
	return appdcim.InterfaceTemplatePage{}, nil
}

type devicePresenceInterfaceCreator struct{ createCalls int }

func (creator *devicePresenceInterfaceCreator) Create(
	context.Context,
	*dcimdomain.Interface,
) error {
	creator.createCalls++
	return nil
}

type devicePresenceInterfaceCascade struct{}

func (devicePresenceInterfaceCascade) DeleteForDevice(
	context.Context,
	shared.ID,
	shared.Timestamp,
) ([]appdcim.InterfaceCascadeChange, error) {
	return nil, nil
}

type devicePresenceRecorder struct {
	backend *devicePresenceBackend
	err     error
	calls   int
}

func (recorder *devicePresenceRecorder) Record(
	ctx context.Context,
	change changelog.Change,
) error {
	recorder.calls++
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Device change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

type devicePresenceDependencies struct {
	deviceTypes map[shared.ID]*dcimdomain.DeviceType
	roles       map[shared.ID]*dcimdomain.DeviceRole
	sites       map[shared.ID]*dcimdomain.Site
	racks       map[shared.ID]*dcimdomain.Rack
}

type devicePresenceHarness struct {
	backend      *devicePresenceBackend
	repository   *devicePresenceRepository
	templates    *devicePresenceTemplateReader
	interfaces   *devicePresenceInterfaceCreator
	recorder     *devicePresenceRecorder
	dependencies devicePresenceDependencies
	service      *appdcim.DeviceService
}

func newDevicePresenceHarness(t *testing.T, recorderErr error) devicePresenceHarness {
	t.Helper()
	dependencies := newDevicePresenceDependencies(t)
	backend := newDevicePresenceBackend()
	repository := &devicePresenceRepository{backend: backend}
	templates := &devicePresenceTemplateReader{}
	interfaces := &devicePresenceInterfaceCreator{}
	recorder := &devicePresenceRecorder{backend: backend, err: recorderErr}
	service, err := appdcim.NewDeviceService(
		repository,
		devicePresenceDeviceTypeReader{deviceTypes: dependencies.deviceTypes},
		devicePresenceRoleReader{roles: dependencies.roles},
		devicePresenceSiteReader{sites: dependencies.sites},
		devicePresenceRackReader{racks: dependencies.racks},
		templates, interfaces, devicePresenceInterfaceCascade{}, backend, recorder,
		authz.AllowAll{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return devicePresenceHarness{
		backend: backend, repository: repository, templates: templates, interfaces: interfaces,
		recorder: recorder, dependencies: dependencies, service: service,
	}
}

func newDevicePresenceDependencies(t *testing.T) devicePresenceDependencies {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturerReference(5, "Presence Vendor", "presence-vendor")
	require.NoError(t, err)
	frontToRear, err := dcimdomain.NewDeviceType(dcimdomain.DeviceTypeValues{
		Manufacturer: manufacturer, Model: "Presence Router A", Slug: "presence-router-a",
		UHeight: "1", IsFullDepth: true,
		Airflow: dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflowFrontToRear),
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, frontToRear.AssignID(10))
	passive, err := dcimdomain.NewDeviceType(dcimdomain.DeviceTypeValues{
		Manufacturer: manufacturer, Model: "Presence Router B", Slug: "presence-router-b",
		UHeight: "1", IsFullDepth: true,
		Airflow: dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflowPassive),
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, passive.AssignID(20))
	role, err := dcimdomain.NewDeviceRole(dcimdomain.DeviceRoleValues{
		Parent: dcimdomain.RootDeviceRoleParent(), Name: "Presence Core",
		Slug: "presence-core", Color: "00ff00",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, role.AssignID(11))
	site, err := dcimdomain.NewSite(dcimdomain.SiteValues{
		Name: "Presence Site", Slug: "presence-site", Status: "active",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, site.AssignID(12))
	siteReference, err := dcimdomain.NewSiteReference(site.ID(), site.Name(), site.Slug().String())
	require.NoError(t, err)
	rack, err := dcimdomain.NewRack(dcimdomain.RackValues{
		Site: siteReference, Name: "Presence Rack",
		FacilityID: dcimdomain.NullRackValue[string](),
		RackType:   dcimdomain.NullRackValue[dcimdomain.RackTypeReference](),
		Status:     "active", Role: dcimdomain.NullRackValue[dcimdomain.RackRoleReference](),
		AssetTag:   dcimdomain.NullRackValue[string](),
		FormFactor: dcimdomain.NullRackValue[string](),
		Width:      19, UHeight: 42, StartingUnit: 1,
		Airflow: dcimdomain.NullRackValue[string](),
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, rack.AssignID(13))
	return devicePresenceDependencies{
		deviceTypes: map[shared.ID]*dcimdomain.DeviceType{10: frontToRear, 20: passive},
		roles:       map[shared.ID]*dcimdomain.DeviceRole{11: role},
		sites:       map[shared.ID]*dcimdomain.Site{12: site},
		racks:       map[shared.ID]*dcimdomain.Rack{13: rack},
	}
}

func seedDevicePresenceState(
	t *testing.T,
	backend *devicePresenceBackend,
	dependencies devicePresenceDependencies,
) {
	t.Helper()
	deviceType := dependencies.deviceTypes[10]
	role := dependencies.roles[11]
	site := dependencies.sites[12]
	rack := dependencies.racks[13]
	deviceTypeReference, err := dcimdomain.NewDeviceTypeInstanceReference(
		deviceType.ID(), deviceType.Model(), deviceType.Slug().String(),
		deviceType.Manufacturer().Name(), deviceType.UHeight(), deviceType.IsFullDepth(),
		deviceType.Airflow(),
	)
	require.NoError(t, err)
	siteReference, err := dcimdomain.NewSiteReference(site.ID(), site.Name(), site.Slug().String())
	require.NoError(t, err)
	rackReference, err := dcimdomain.NewRackReference(
		rack.ID(), rack.Display(), rack.Site().ID(), rack.StartingUnit(), rack.UHeight(),
	)
	require.NoError(t, err)
	position, err := dcimdomain.ParseRackPosition("10")
	require.NoError(t, err)
	backend.state.devices[41] = dcimdomain.DeviceState{
		ID: 41, DeviceType: deviceTypeReference,
		Role: dcimdomain.DeviceRoleReference{ID: role.ID(), Display: role.Display()},
		Name: dcimdomain.NonNullDeviceValue("Original Device"), Site: siteReference,
		Rack:     dcimdomain.NonNullDeviceValue(rackReference),
		Position: dcimdomain.NonNullDeviceValue(position), Face: "front",
		Status: "planned", Serial: "SERIAL-01",
		AssetTag:    dcimdomain.NonNullDeviceValue("ASSET-01"),
		Airflow:     dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflowRearToFront),
		Description: "Original description", Comments: "Original comments",
		Created: createdAt, LastUpdated: createdAt,
	}
}

func devicePresenceSnapshot(
	t *testing.T,
	state dcimdomain.DeviceState,
) dcimdomain.DeviceSnapshot {
	t.Helper()
	device, err := dcimdomain.RestoreDevice(state)
	require.NoError(t, err)
	return device.Snapshot()
}

func devicePresenceViolationFields(err error) []string {
	violations := shared.ViolationsOf(err)
	fields := make([]string, len(violations))
	for index, violation := range violations {
		fields[index] = violation.Field
	}
	return fields
}

func devicePresenceNullableString(
	t *testing.T,
	value dcimdomain.DeviceNullable[string],
) string {
	t.Helper()
	actual, present := value.Get()
	require.True(t, present)
	return actual
}

func devicePresenceAirflow(
	t *testing.T,
	value dcimdomain.NullableDeviceAirflow,
) dcimdomain.DeviceAirflow {
	t.Helper()
	actual, present := value.Get()
	require.True(t, present)
	return actual
}
