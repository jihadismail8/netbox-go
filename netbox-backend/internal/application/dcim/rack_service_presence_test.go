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

func TestReplaceRackAppliesPinnedOmissionDefaultsAndPreservesOptionalState(t *testing.T) {
	t.Parallel()

	t.Run("PUT clears only facility and RackType omissions", func(t *testing.T) {
		backend, repository, service := newRackPresenceService(t, nil)
		seedRackPresenceState(t, backend, true)
		before := backend.state.racks[41]

		rack, err := service.ReplaceRack(
			t.Context(), testPrincipal(), appdcim.ReplaceRackCommand{
				ID: 41,
				CreateRackCommand: appdcim.CreateRackCommand{
					Site: appdcim.FieldValue(shared.ID(3)),
					Name: appdcim.FieldValue("  Replacement Rack  "),
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "Replacement Rack", rack.Name())
		assert.True(t, rack.FacilityID().IsNull())
		assert.True(t, rack.RackType().IsNull())
		assert.Equal(t, before.Status, rack.Status().String())
		assert.Equal(t, rackStateRoleID(t, before), rackRoleID(t, rack))
		assert.Equal(t, before.Serial, rack.Serial())
		assert.Equal(t, rackStateNullableString(t, before.AssetTag), rackNullableString(t, rack.AssetTag()))
		assert.Equal(t, rackStateNullableString(t, before.FormFactor), rackNullableFormFactor(t, rack))
		assert.Equal(t, before.Width, rack.Width().Uint32())
		assert.Equal(t, before.UHeight, rack.UHeight())
		assert.Equal(t, before.StartingUnit, rack.StartingUnit())
		assert.Equal(t, before.DescUnits, rack.DescUnits())
		assert.Equal(t, rackStateNullableString(t, before.Airflow), rackNullableAirflow(t, rack))
		assert.Equal(t, before.Description, rack.Description())
		assert.Equal(t, before.Comments, rack.Comments())
		assert.Equal(t, before.Created, rack.Created())
		assert.Equal(t, updatedAt, rack.LastUpdated())
		assert.Equal(t, 1, backend.transactionCalls)
		assert.Equal(t, 1, repository.getForUpdateCalls)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Equal(t, 1, repository.mountedCalls)
		assert.Zero(t, repository.propagationCalls)
		require.Len(t, backend.state.changes, 1)
		assert.Equal(t, rackPresenceSnapshot(t, before), backend.state.changes[0].Before)
		assert.Equal(t, rack.Snapshot(), backend.state.changes[0].After)
	})

	t.Run("RackType wins over conflicting direct physical fields", func(t *testing.T) {
		backend, repository, service := newRackPresenceService(t, nil)
		seedRackPresenceState(t, backend, false)

		rack, err := service.UpdateRack(
			t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
				ID: 41, RackType: appdcim.FieldValue(shared.ID(8)),
				FormFactor: appdcim.FieldValue("wall-cabinet"),
				Width:      appdcim.FieldValue(uint32(19)), UHeight: appdcim.FieldValue(uint32(80)),
				StartingUnit: appdcim.FieldValue(uint32(7)), DescUnits: appdcim.FieldValue(false),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, shared.ID(8), rackTypeID(t, rack))
		assert.Equal(t, "wall-frame", rackNullableFormFactor(t, rack))
		assert.Equal(t, uint32(23), rack.Width().Uint32())
		assert.Equal(t, uint32(24), rack.UHeight())
		assert.Equal(t, uint32(3), rack.StartingUnit())
		assert.True(t, rack.DescUnits())
		assert.Equal(t, 1, repository.mountedCalls)
		assert.Zero(t, repository.propagationCalls)
		require.Len(t, backend.state.changes, 1)
	})

	t.Run("PUT RackType wins over all five conflicting physical fields", func(t *testing.T) {
		backend, repository, service := newRackPresenceService(t, nil)
		seedRackPresenceState(t, backend, false)
		before := backend.state.racks[41]

		rack, err := service.ReplaceRack(
			t.Context(), testPrincipal(), appdcim.ReplaceRackCommand{
				ID: 41,
				CreateRackCommand: appdcim.CreateRackCommand{
					Site: appdcim.FieldValue(shared.ID(3)), Name: appdcim.FieldValue("  PUT Typed Rack  "),
					RackType:   appdcim.FieldValue(shared.ID(8)),
					FormFactor: appdcim.FieldValue("wall-cabinet"),
					Width:      appdcim.FieldValue(uint32(19)), UHeight: appdcim.FieldValue(uint32(80)),
					StartingUnit: appdcim.FieldValue(uint32(7)), DescUnits: appdcim.FieldValue(false),
				},
			},
		)
		require.NoError(t, err)
		assert.Equal(t, "PUT Typed Rack", rack.Name())
		assert.True(t, rack.FacilityID().IsNull())
		assert.Equal(t, shared.ID(8), rackTypeID(t, rack))
		assert.Equal(t, "wall-frame", rackNullableFormFactor(t, rack))
		assert.Equal(t, uint32(23), rack.Width().Uint32())
		assert.Equal(t, uint32(24), rack.UHeight())
		assert.Equal(t, uint32(3), rack.StartingUnit())
		assert.True(t, rack.DescUnits())
		assert.Equal(t, before.Status, rack.Status().String())
		assert.Equal(t, rackStateRoleID(t, before), rackRoleID(t, rack))
		assert.Equal(t, before.Serial, rack.Serial())
		assert.Equal(t, rackStateNullableString(t, before.AssetTag), rackNullableString(t, rack.AssetTag()))
		assert.Equal(t, rackStateNullableString(t, before.Airflow), rackNullableAirflow(t, rack))
		assert.Equal(t, before.Description, rack.Description())
		assert.Equal(t, before.Comments, rack.Comments())
		assert.Equal(t, before.Created, rack.Created())
		assert.Equal(t, updatedAt, rack.LastUpdated())
		assert.Equal(t, 1, backend.transactionCalls)
		assert.Equal(t, 1, repository.getForUpdateCalls)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Equal(t, 1, repository.mountedCalls)
		assert.Zero(t, repository.propagationCalls)
		require.Len(t, backend.state.changes, 1)
		assert.Equal(t, rackPresenceSnapshot(t, before), backend.state.changes[0].Before)
		assert.Equal(t, rack.Snapshot(), backend.state.changes[0].After)
	})

	t.Run("clearing RackType retains explicit physical replacements", func(t *testing.T) {
		backend, repository, service := newRackPresenceService(t, nil)
		seedRackPresenceState(t, backend, true)

		rack, err := service.UpdateRack(
			t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
				ID: 41, RackType: appdcim.NullField[shared.ID](),
				FormFactor: appdcim.FieldValue("4-post-cabinet"),
				Width:      appdcim.FieldValue(uint32(19)), UHeight: appdcim.FieldValue(uint32(48)),
				StartingUnit: appdcim.FieldValue(uint32(2)), DescUnits: appdcim.FieldValue(false),
			},
		)
		require.NoError(t, err)
		assert.True(t, rack.RackType().IsNull())
		assert.Equal(t, "4-post-cabinet", rackNullableFormFactor(t, rack))
		assert.Equal(t, uint32(19), rack.Width().Uint32())
		assert.Equal(t, uint32(48), rack.UHeight())
		assert.Equal(t, uint32(2), rack.StartingUnit())
		assert.False(t, rack.DescUnits())
		assert.Equal(t, 1, repository.mountedCalls)
		assert.Zero(t, repository.propagationCalls)
		require.Len(t, backend.state.changes, 1)
	})
}

func TestRackScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	longFacility := strings.Repeat("f", dcimdomain.RackFacilityIDMaxLength+1)
	longSerial := strings.Repeat("s", dcimdomain.RackSerialMaxLength+1)
	longAsset := strings.Repeat("a", dcimdomain.RackAssetTagMaxLength+1)
	longDescription := strings.Repeat("d", dcimdomain.RackDescriptionMaxLength+1)
	invalid := appdcim.CreateRackCommand{
		Site: appdcim.NullField[shared.ID](), Name: appdcim.NullField[string](),
		FacilityID: appdcim.FieldValue(longFacility), RackType: appdcim.FieldValue(shared.ID(0)),
		Status: appdcim.FieldValue(" active "), Role: appdcim.FieldValue(shared.ID(0)),
		Serial: appdcim.FieldValue(longSerial), AssetTag: appdcim.FieldValue(longAsset),
		FormFactor: appdcim.FieldValue(" wall-frame "), Width: appdcim.FieldValue(uint32(20)),
		UHeight:      appdcim.FieldValue(uint32(0)),
		StartingUnit: appdcim.FieldValue(dcimdomain.RackTypeMaximumStartingUnit + 1),
		DescUnits:    appdcim.NullField[bool](), Airflow: appdcim.FieldValue(" front-to-rear "),
		Description: appdcim.FieldValue(longDescription), Comments: appdcim.NullField[string](),
	}
	expectedFields := []string{
		"site", "name", "facility_id", "rack_type", "status", "role", "serial",
		"asset_tag", "form_factor", "width", "u_height", "starting_unit",
		"desc_units", "airflow", "description", "comments",
	}

	for _, test := range []struct {
		name   string
		seed   bool
		mutate func(*appdcim.RackService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.RackService) error {
				_, err := service.CreateRack(t.Context(), testPrincipal(), invalid)
				return err
			},
		},
		{
			name: "PUT", seed: true,
			mutate: func(service *appdcim.RackService) error {
				_, err := service.ReplaceRack(t.Context(), testPrincipal(), appdcim.ReplaceRackCommand{
					ID: 41, CreateRackCommand: invalid,
				})
				return err
			},
		},
		{
			name: "PATCH", seed: true,
			mutate: func(service *appdcim.RackService) error {
				_, err := service.UpdateRack(t.Context(), testPrincipal(), rackUpdateFromCreate(41, invalid))
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			backend, repository, service := newRackPresenceService(t, nil)
			if test.seed {
				seedRackPresenceState(t, backend, true)
			}
			before := backend.state.clone()

			err := test.mutate(service)
			require.Error(t, err)
			assert.Equal(t, expectedFields, rackPresenceViolationFields(err))
			assert.Equal(t, before, backend.state)
			assert.Equal(t, 1, backend.transactionCalls)
			if test.seed {
				assert.Equal(t, 1, repository.getForUpdateCalls)
			} else {
				assert.Zero(t, repository.getForUpdateCalls)
			}
			assert.Zero(t, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
			assert.Zero(t, repository.mountedCalls)
			assert.Zero(t, repository.propagationCalls)
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("unknown relationships aggregate without mutation", func(t *testing.T) {
		backend, repository, service := newRackPresenceService(t, nil)
		seedRackPresenceState(t, backend, true)
		before := backend.state.clone()
		_, err := service.UpdateRack(t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
			ID: 41, Site: appdcim.FieldValue(shared.ID(404)),
			RackType:    appdcim.FieldValue(shared.ID(405)),
			Role:        appdcim.FieldValue(shared.ID(406)),
			Description: appdcim.FieldValue("valid sibling"),
		})
		require.Error(t, err)
		assert.Equal(t, []string{"site", "rack_type", "role"}, rackPresenceViolationFields(err))
		assert.Equal(t, before, backend.state)
		assert.Zero(t, repository.updateCalls)
		assert.Zero(t, repository.mountedCalls)
	})

	for _, field := range []string{
		"site", "name", "status", "serial", "width", "u_height", "starting_unit",
		"desc_units", "airflow", "description", "comments",
	} {
		field := field
		t.Run("PATCH shared null validation/"+field, func(t *testing.T) {
			backend, repository, service := newRackPresenceService(t, nil)
			seedRackPresenceState(t, backend, true)
			before := backend.state.clone()

			_, err := service.UpdateRack(
				t.Context(), testPrincipal(), rackNullUpdateCommand(41, field),
			)
			require.Error(t, err)
			assert.Equal(t, []string{field}, rackPresenceViolationFields(err))
			assert.Equal(t, before, backend.state)
			assert.Equal(t, 1, backend.transactionCalls)
			assert.Equal(t, 1, repository.getForUpdateCalls)
			assert.Zero(t, repository.updateCalls)
			assert.Zero(t, repository.mountedCalls)
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("repository failure rolls back Rack and object change", func(t *testing.T) {
		writeFailure := errors.New("forced Rack repository failure")
		backend, repository, service := newRackPresenceService(t, nil)
		seedRackPresenceState(t, backend, true)
		repository.updateErr = writeFailure
		before := backend.state.clone()
		_, err := service.UpdateRack(t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
			ID: 41, Description: appdcim.FieldValue("must roll back"),
		})
		require.ErrorIs(t, err, writeFailure)
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Equal(t, 1, repository.mountedCalls)
		assert.Empty(t, backend.state.changes)
	})

	t.Run("recorder failure rolls back Rack and object change", func(t *testing.T) {
		recorderFailure := errors.New("forced Rack recorder failure")
		backend, repository, service := newRackPresenceService(t, recorderFailure)
		seedRackPresenceState(t, backend, true)
		before := backend.state.clone()
		_, err := service.UpdateRack(t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
			ID: 41, Description: appdcim.FieldValue("must roll back"),
		})
		require.ErrorIs(t, err, recorderFailure)
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Equal(t, 1, repository.mountedCalls)
		assert.Empty(t, backend.state.changes)
	})
}

func rackNullUpdateCommand(id shared.ID, field string) appdcim.UpdateRackCommand {
	command := appdcim.UpdateRackCommand{ID: id}
	switch field {
	case "site":
		command.Site = appdcim.NullField[shared.ID]()
	case "name":
		command.Name = appdcim.NullField[string]()
	case "status":
		command.Status = appdcim.NullField[string]()
	case "serial":
		command.Serial = appdcim.NullField[string]()
	case "width":
		command.Width = appdcim.NullField[uint32]()
	case "u_height":
		command.UHeight = appdcim.NullField[uint32]()
	case "starting_unit":
		command.StartingUnit = appdcim.NullField[uint32]()
	case "desc_units":
		command.DescUnits = appdcim.NullField[bool]()
	case "airflow":
		command.Airflow = appdcim.NullField[string]()
	case "description":
		command.Description = appdcim.NullField[string]()
	case "comments":
		command.Comments = appdcim.NullField[string]()
	}
	return command
}

func rackUpdateFromCreate(id shared.ID, command appdcim.CreateRackCommand) appdcim.UpdateRackCommand {
	return appdcim.UpdateRackCommand{
		ID: id, Site: command.Site, Name: command.Name, FacilityID: command.FacilityID,
		RackType: command.RackType, Status: command.Status, Role: command.Role,
		Serial: command.Serial, AssetTag: command.AssetTag, FormFactor: command.FormFactor,
		Width: command.Width, UHeight: command.UHeight, StartingUnit: command.StartingUnit,
		DescUnits: command.DescUnits, Airflow: command.Airflow,
		Description: command.Description, Comments: command.Comments,
	}
}

type rackPresenceTransactionKey struct{}

type rackPresenceState struct {
	nextID  shared.ID
	racks   map[shared.ID]dcimdomain.RackState
	changes []changelog.Change
}

func (state rackPresenceState) clone() rackPresenceState {
	cloned := rackPresenceState{
		nextID:  state.nextID,
		racks:   make(map[shared.ID]dcimdomain.RackState, len(state.racks)),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, rack := range state.racks {
		cloned.racks[id] = rack
	}
	return cloned
}

type rackPresenceBackend struct {
	state            rackPresenceState
	transactionCalls int
}

func newRackPresenceBackend() *rackPresenceBackend {
	return &rackPresenceBackend{state: rackPresenceState{
		nextID: 42, racks: make(map[shared.ID]dcimdomain.RackState),
	}}
}

func (backend *rackPresenceBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, rackPresenceTransactionKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *rackPresenceBackend) stateFor(ctx context.Context) (*rackPresenceState, bool) {
	state, transactional := ctx.Value(rackPresenceTransactionKey{}).(*rackPresenceState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type rackPresenceRepository struct {
	backend           *rackPresenceBackend
	createCalls       int
	updateCalls       int
	getForUpdateCalls int
	mountedCalls      int
	propagationCalls  int
	updateErr         error
}

func (*rackPresenceRepository) List(
	context.Context,
	appdcim.RackListCriteria,
) (appdcim.RackPage, error) {
	return appdcim.RackPage{}, nil
}

func (repository *rackPresenceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.Rack, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.racks[id]
	if !ok {
		return nil, shared.NotFound("Rack", id)
	}
	return dcimdomain.RestoreRack(persisted)
}

func (repository *rackPresenceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.Rack, error) {
	repository.getForUpdateCalls++
	return repository.Get(ctx, id)
}

func (repository *rackPresenceRepository) Create(
	ctx context.Context,
	rack *dcimdomain.Rack,
) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Rack created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := rack.AssignID(id); err != nil {
		return err
	}
	state.racks[id] = rack.State()
	return nil
}

func (repository *rackPresenceRepository) Update(
	ctx context.Context,
	rack *dcimdomain.Rack,
) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Rack updated outside transaction")
	}
	state.racks[rack.ID()] = rack.State()
	return repository.updateErr
}

func (repository *rackPresenceRepository) Delete(
	ctx context.Context,
	rack *dcimdomain.Rack,
) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Rack deleted outside transaction")
	}
	delete(state.racks, rack.ID())
	return nil
}

func (repository *rackPresenceRepository) MountedDevices(
	context.Context,
	shared.ID,
) ([]appdcim.RackDevicePlacement, error) {
	repository.mountedCalls++
	return nil, nil
}

func (repository *rackPresenceRepository) PropagateSiteToDevices(
	context.Context,
	shared.ID,
	shared.ID,
	shared.Timestamp,
) ([]appdcim.RackSitePropagationChange, error) {
	repository.propagationCalls++
	return nil, nil
}

type rackPresenceSiteReader struct {
	sites map[shared.ID]*dcimdomain.Site
}

func (reader rackPresenceSiteReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Site, error) {
	if site := reader.sites[id]; site != nil {
		return site, nil
	}
	return nil, shared.NotFound("Site", id)
}

type rackPresenceTypeReader struct {
	rackTypes map[shared.ID]*dcimdomain.RackType
}

func (reader rackPresenceTypeReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.RackType, error) {
	if rackType := reader.rackTypes[id]; rackType != nil {
		return rackType, nil
	}
	return nil, shared.NotFound("RackType", id)
}

type rackPresenceRoleReader struct {
	roles map[shared.ID]*dcimdomain.RackRole
}

func (reader rackPresenceRoleReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.RackRole, error) {
	if role := reader.roles[id]; role != nil {
		return role, nil
	}
	return nil, shared.NotFound("RackRole", id)
}

type rackPresenceRecorder struct {
	backend *rackPresenceBackend
	err     error
}

func (recorder rackPresenceRecorder) Record(ctx context.Context, change changelog.Change) error {
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("Rack change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

func newRackPresenceService(
	t *testing.T,
	recorderErr error,
) (*rackPresenceBackend, *rackPresenceRepository, *appdcim.RackService) {
	t.Helper()
	site, rackType, role := rackPresenceDependencies(t)
	backend := newRackPresenceBackend()
	repository := &rackPresenceRepository{backend: backend}
	service, err := appdcim.NewRackService(
		repository,
		rackPresenceSiteReader{sites: map[shared.ID]*dcimdomain.Site{site.ID(): site}},
		rackPresenceTypeReader{rackTypes: map[shared.ID]*dcimdomain.RackType{rackType.ID(): rackType}},
		rackPresenceRoleReader{roles: map[shared.ID]*dcimdomain.RackRole{role.ID(): role}},
		backend,
		rackPresenceRecorder{backend: backend, err: recorderErr},
		authz.AllowAll{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return backend, repository, service
}

func rackPresenceDependencies(
	t *testing.T,
) (*dcimdomain.Site, *dcimdomain.RackType, *dcimdomain.RackRole) {
	t.Helper()
	site, err := dcimdomain.NewSite(dcimdomain.SiteValues{
		Name: "Presence Site", Slug: "presence-site", Status: "active",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, site.AssignID(3))
	manufacturer, err := dcimdomain.NewManufacturerReference(5, "Presence Vendor", "presence-vendor")
	require.NoError(t, err)
	rackType, err := dcimdomain.NewRackType(dcimdomain.RackTypeValues{
		Manufacturer: manufacturer, Model: "Presence R24", Slug: "presence-r24",
		FormFactor: "wall-frame", Width: 23, UHeight: 24, StartingUnit: 3, DescUnits: true,
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, rackType.AssignID(8))
	role, err := dcimdomain.NewRackRole(dcimdomain.RackRoleValues{
		Name: "Presence Production", Slug: "presence-production", Color: "00ff00",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, role.AssignID(9))
	return site, rackType, role
}

func seedRackPresenceState(t *testing.T, backend *rackPresenceBackend, withRackType bool) {
	t.Helper()
	site, rackType, role := rackPresenceDependencies(t)
	siteReference, err := dcimdomain.NewSiteReference(site.ID(), site.Name(), site.Slug().String())
	require.NoError(t, err)
	roleReference, err := dcimdomain.NewRackRoleReference(role.ID(), role.Name(), role.Slug().String())
	require.NoError(t, err)
	rackTypeReference, err := dcimdomain.NewRackTypeReference(
		rackType.ID(), rackType.Model(), rackType.Slug().String(), rackType.PhysicalAttributes(),
	)
	require.NoError(t, err)
	rackTypeValue := dcimdomain.NullRackValue[dcimdomain.RackTypeReference]()
	formFactor, width, height, startingUnit, descUnits := "4-post-cabinet", uint32(19), uint32(42), uint32(1), false
	if withRackType {
		rackTypeValue = dcimdomain.NonNullRackValue(rackTypeReference)
		formFactor, width, height, startingUnit, descUnits = "wall-frame", 23, 24, 3, true
	}
	backend.state.racks[41] = dcimdomain.RackState{
		ID: 41, Site: siteReference, Name: "Original Rack",
		FacilityID: dcimdomain.NonNullRackValue("FAC-01"), RackType: rackTypeValue,
		Status: "planned", Role: dcimdomain.NonNullRackValue(roleReference),
		Serial: "SERIAL-01", AssetTag: dcimdomain.NonNullRackValue("ASSET-01"),
		FormFactor: dcimdomain.NonNullRackValue(formFactor),
		Width:      width, UHeight: height, StartingUnit: startingUnit, DescUnits: descUnits,
		Airflow:     dcimdomain.NonNullRackValue("rear-to-front"),
		Description: "Original description", Comments: "Original comments",
		Created: createdAt, LastUpdated: createdAt,
	}
}

func rackPresenceSnapshot(t *testing.T, state dcimdomain.RackState) dcimdomain.RackSnapshot {
	t.Helper()
	rack, err := dcimdomain.RestoreRack(state)
	require.NoError(t, err)
	return rack.Snapshot()
}

func rackPresenceViolationFields(err error) []string {
	violations := shared.ViolationsOf(err)
	fields := make([]string, len(violations))
	for index, violation := range violations {
		fields[index] = violation.Field
	}
	return fields
}

func rackStateNullableString(t *testing.T, value dcimdomain.RackNullable[string]) string {
	t.Helper()
	actual, present := value.Get()
	require.True(t, present)
	return actual
}

func rackNullableString(t *testing.T, value dcimdomain.RackNullable[string]) string {
	t.Helper()
	actual, present := value.Get()
	require.True(t, present)
	return actual
}

func rackStateRoleID(
	t *testing.T,
	state dcimdomain.RackState,
) shared.ID {
	t.Helper()
	role, present := state.Role.Get()
	require.True(t, present)
	return role.ID()
}

func rackRoleID(t *testing.T, rack *dcimdomain.Rack) shared.ID {
	t.Helper()
	role, present := rack.Role().Get()
	require.True(t, present)
	return role.ID()
}

func rackTypeID(t *testing.T, rack *dcimdomain.Rack) shared.ID {
	t.Helper()
	rackType, present := rack.RackType().Get()
	require.True(t, present)
	return rackType.ID()
}

func rackNullableFormFactor(t *testing.T, rack *dcimdomain.Rack) string {
	t.Helper()
	value, present := rack.FormFactor().Get()
	require.True(t, present)
	return value.String()
}

func rackNullableAirflow(t *testing.T, rack *dcimdomain.Rack) string {
	t.Helper()
	value, present := rack.Airflow().Get()
	require.True(t, present)
	return value.String()
}
