package dcim

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestPostgresRackScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 26, 8, 0, 0, 0, time.UTC)
	_, siteID, rackTypeID, roleID := seedPostgresRackPresenceDependencies(t, db, createdAt)
	service := newPostgresRackPresenceService(t, db, createdAt, postgreschangelog.NewRecorder(db))

	rack, err := service.CreateRack(t.Context(), principal, applicationdcim.CreateRackCommand{
		Site: applicationdcim.FieldValue(siteID), Name: applicationdcim.FieldValue("  Durable Rack  "),
	})
	require.NoError(t, err)
	requirePostgresRackAggregate(
		t, rack, siteID, "Presence Rack Site", "Durable Rack", nil, nil,
		"active", nil, "", nil, nil, 19, 42, 1, false, nil, "", "",
		createdAt, createdAt,
	)
	created := loadPostgresRackPresenceState(t, db, rack.ID())
	requirePostgresRackRow(t, created.row, dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{ID: rack.ID().Int64(), Created: createdAt, LastUpdated: createdAt},
		SiteID:      siteID.Int64(), Name: "Durable Rack", Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	})
	require.Equal(t, int64(1), created.manufacturerCount)
	require.Equal(t, int64(1), created.siteCount)
	require.Equal(t, int64(1), created.rackTypeCount)
	require.Equal(t, int64(1), created.roleCount)
	require.Equal(t, int64(1), created.rackCount)
	require.Zero(t, created.deviceCount, "scalar presence requires zero Devices")
	require.Equal(t, int64(1), created.targetRackChangeCount)
	require.Equal(t, int64(1), created.allRackChangeCount)
	require.Zero(t, created.deviceChangeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	requirePostgresRackReloadedRelationships(t, db, rack.ID(), siteID, nil, nil)

	concreteAt := createdAt.Add(time.Minute)
	service = newPostgresRackPresenceService(t, db, concreteAt, postgreschangelog.NewRecorder(db))
	concrete, err := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{
		ID: rack.ID(), Site: applicationdcim.FieldValue(siteID),
		Name:       applicationdcim.FieldValue("  Durable Concrete Rack  "),
		FacilityID: applicationdcim.FieldValue("  FAC-01  "),
		Status:     applicationdcim.FieldValue("planned"), Role: applicationdcim.FieldValue(roleID),
		Serial:     applicationdcim.FieldValue("  SERIAL-01  "),
		AssetTag:   applicationdcim.FieldValue("  ASSET-I9-01  "),
		FormFactor: applicationdcim.FieldValue("wall-cabinet"),
		Width:      applicationdcim.FieldValue(uint32(23)), UHeight: applicationdcim.FieldValue(uint32(100)),
		StartingUnit: applicationdcim.FieldValue(domaindcim.RackTypeMaximumStartingUnit),
		DescUnits:    applicationdcim.FieldValue(true),
		Airflow:      applicationdcim.FieldValue("rear-to-front"),
		Description:  applicationdcim.FieldValue("  durable description  "),
		Comments:     applicationdcim.FieldValue("  durable comments  "),
	})
	require.NoError(t, err)
	requirePostgresRackAggregate(
		t, concrete, siteID, "Presence Rack Site", "Durable Concrete Rack",
		rackPresenceStringPointer("FAC-01"), nil, "planned", &roleID, "SERIAL-01",
		rackPresenceStringPointer("ASSET-I9-01"), rackPresenceStringPointer("wall-cabinet"),
		23, 100, domaindcim.RackTypeMaximumStartingUnit, true,
		rackPresenceStringPointer("rear-to-front"), "durable description", "durable comments",
		createdAt, concreteAt,
	)
	concreteState := loadPostgresRackPresenceState(t, db, rack.ID())
	requirePostgresRackRow(t, concreteState.row, dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{ID: rack.ID().Int64(), Created: createdAt, LastUpdated: concreteAt},
		SiteID:      siteID.Int64(), Name: "Durable Concrete Rack",
		FacilityID: rackPresenceStringPointer("FAC-01"), Status: "planned", RoleID: rackPresenceInt64Pointer(roleID.Int64()),
		Serial: "SERIAL-01", AssetTag: rackPresenceStringPointer("ASSET-I9-01"),
		FormFactor: rackPresenceStringPointer("wall-cabinet"), Width: 23, UHeight: 100,
		StartingUnit: int64(domaindcim.RackTypeMaximumStartingUnit), DescUnits: true,
		Airflow: rackPresenceStringPointer("rear-to-front"), Description: "durable description", Comments: "durable comments",
	})
	requirePostgresRackMutationRecorded(t, created, concreteState, concreteAt)
	requirePostgresRackReloadedRelationships(t, db, rack.ID(), siteID, nil, &roleID)

	ownedAt := concreteAt.Add(time.Minute)
	service = newPostgresRackPresenceService(t, db, ownedAt, postgreschangelog.NewRecorder(db))
	owned, err := service.ReplaceRack(t.Context(), principal, applicationdcim.ReplaceRackCommand{
		ID: rack.ID(), CreateRackCommand: applicationdcim.CreateRackCommand{
			Site: applicationdcim.FieldValue(siteID), Name: applicationdcim.FieldValue("Durable Concrete Rack"),
			FacilityID: applicationdcim.FieldValue("FAC-01"),
			RackType:   applicationdcim.FieldValue(rackTypeID),
			FormFactor: applicationdcim.FieldValue("4-post-cabinet"),
			Width:      applicationdcim.FieldValue(uint32(19)), UHeight: applicationdcim.FieldValue(uint32(80)),
			StartingUnit: applicationdcim.FieldValue(uint32(7)), DescUnits: applicationdcim.FieldValue(false),
		},
	})
	require.NoError(t, err)
	requirePostgresRackAggregate(
		t, owned, siteID, "Presence Rack Site", "Durable Concrete Rack",
		rackPresenceStringPointer("FAC-01"), &rackTypeID, "planned", &roleID, "SERIAL-01",
		rackPresenceStringPointer("ASSET-I9-01"), rackPresenceStringPointer("wall-frame"),
		23, 24, 3, true, rackPresenceStringPointer("rear-to-front"),
		"durable description", "durable comments", createdAt, ownedAt,
	)
	ownedState := loadPostgresRackPresenceState(t, db, rack.ID())
	requirePostgresRackRow(t, ownedState.row, dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{ID: rack.ID().Int64(), Created: createdAt, LastUpdated: ownedAt},
		SiteID:      siteID.Int64(), Name: "Durable Concrete Rack",
		FacilityID: rackPresenceStringPointer("FAC-01"), RackTypeID: rackPresenceInt64Pointer(rackTypeID.Int64()),
		Status: "planned", RoleID: rackPresenceInt64Pointer(roleID.Int64()), Serial: "SERIAL-01",
		AssetTag: rackPresenceStringPointer("ASSET-I9-01"), FormFactor: rackPresenceStringPointer("wall-frame"),
		Width: 23, UHeight: 24, StartingUnit: 3, DescUnits: true,
		Airflow: rackPresenceStringPointer("rear-to-front"), Description: "durable description", Comments: "durable comments",
	})
	requirePostgresRackMutationRecorded(t, concreteState, ownedState, ownedAt)
	requirePostgresRackReloadedRelationships(t, db, rack.ID(), siteID, &rackTypeID, &roleID)

	replacedAt := ownedAt.Add(time.Minute)
	service = newPostgresRackPresenceService(t, db, replacedAt, postgreschangelog.NewRecorder(db))
	replaced, err := service.ReplaceRack(t.Context(), principal, applicationdcim.ReplaceRackCommand{
		ID: rack.ID(), CreateRackCommand: applicationdcim.CreateRackCommand{
			Site: applicationdcim.FieldValue(siteID), Name: applicationdcim.FieldValue("  Durable PUT Rack  "),
		},
	})
	require.NoError(t, err)
	requirePostgresRackAggregate(
		t, replaced, siteID, "Presence Rack Site", "Durable PUT Rack", nil, nil,
		"planned", &roleID, "SERIAL-01", rackPresenceStringPointer("ASSET-I9-01"),
		rackPresenceStringPointer("wall-frame"), 23, 24, 3, true,
		rackPresenceStringPointer("rear-to-front"), "durable description", "durable comments",
		createdAt, replacedAt,
	)
	replacedState := loadPostgresRackPresenceState(t, db, rack.ID())
	requirePostgresRackRow(t, replacedState.row, dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{ID: rack.ID().Int64(), Created: createdAt, LastUpdated: replacedAt},
		SiteID:      siteID.Int64(), Name: "Durable PUT Rack", Status: "planned",
		RoleID: rackPresenceInt64Pointer(roleID.Int64()), Serial: "SERIAL-01",
		AssetTag: rackPresenceStringPointer("ASSET-I9-01"), FormFactor: rackPresenceStringPointer("wall-frame"),
		Width: 23, UHeight: 24, StartingUnit: 3, DescUnits: true,
		Airflow: rackPresenceStringPointer("rear-to-front"), Description: "durable description", Comments: "durable comments",
	})
	requirePostgresRackMutationRecorded(t, ownedState, replacedState, replacedAt)
	requirePostgresRackReloadedRelationships(t, db, rack.ID(), siteID, nil, &roleID)

	clearedAt := replacedAt.Add(time.Minute)
	service = newPostgresRackPresenceService(t, db, clearedAt, postgreschangelog.NewRecorder(db))
	cleared, err := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{
		ID: rack.ID(), FacilityID: applicationdcim.FieldValue(""), Role: applicationdcim.NullField[shared.ID](),
		Serial: applicationdcim.FieldValue(""), AssetTag: applicationdcim.NullField[string](),
		FormFactor: applicationdcim.FieldValue(""), DescUnits: applicationdcim.FieldValue(false),
		Airflow: applicationdcim.FieldValue(""), Description: applicationdcim.FieldValue(""),
		Comments: applicationdcim.FieldValue(""),
	})
	require.NoError(t, err)
	requirePostgresRackAggregate(
		t, cleared, siteID, "Presence Rack Site", "Durable PUT Rack",
		rackPresenceStringPointer(""), nil, "planned", nil, "", nil,
		rackPresenceStringPointer(""), 23, 24, 3, false, rackPresenceStringPointer(""),
		"", "", createdAt, clearedAt,
	)
	clearedState := loadPostgresRackPresenceState(t, db, rack.ID())
	requirePostgresRackRow(t, clearedState.row, dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{ID: rack.ID().Int64(), Created: createdAt, LastUpdated: clearedAt},
		SiteID:      siteID.Int64(), Name: "Durable PUT Rack", FacilityID: rackPresenceStringPointer(""),
		Status: "planned", FormFactor: rackPresenceStringPointer(""), Width: 23, UHeight: 24,
		StartingUnit: 3, Airflow: rackPresenceStringPointer(""),
	})
	requirePostgresRackMutationRecorded(t, replacedState, clearedState, clearedAt)

	service = newPostgresRackPresenceService(t, db, clearedAt.Add(time.Minute), postgreschangelog.NewRecorder(db))
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresRackPresenceState(t, db, rack.ID())
			operationErr := operation()
			require.Error(t, operationErr)
			require.True(t, shared.HasReason(operationErr, shared.ErrorReasonValidation), operationErr)
			require.Equal(t, before, loadPostgresRackPresenceState(t, db, rack.ID()))
		})
	}
	assertRejected("POST missing required identity", func() error {
		_, operationErr := service.CreateRack(t.Context(), principal, applicationdcim.CreateRackCommand{})
		return operationErr
	})
	assertRejected("POST explicit null airflow", func() error {
		_, operationErr := service.CreateRack(t.Context(), principal, applicationdcim.CreateRackCommand{
			Site: applicationdcim.FieldValue(siteID), Name: applicationdcim.FieldValue("Rejected Airflow"),
			Airflow: applicationdcim.NullField[string](),
		})
		return operationErr
	})
	assertRejected("PUT missing required identity", func() error {
		_, operationErr := service.ReplaceRack(t.Context(), principal, applicationdcim.ReplaceRackCommand{ID: rack.ID()})
		return operationErr
	})
	for _, rejected := range []struct {
		name      string
		operation func() error
	}{
		{"PATCH whitespace status", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), Status: applicationdcim.FieldValue(" active ")})
			return operationErr
		}},
		{"PATCH whitespace form factor", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), FormFactor: applicationdcim.FieldValue(" wall-frame ")})
			return operationErr
		}},
		{"PATCH whitespace airflow", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), Airflow: applicationdcim.FieldValue(" rear-to-front ")})
			return operationErr
		}},
		{"PATCH starting unit overflow", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), StartingUnit: applicationdcim.FieldValue(domaindcim.RackTypeMaximumStartingUnit + 1)})
			return operationErr
		}},
		{"PATCH unknown Site", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), Site: applicationdcim.FieldValue(shared.ID(999997))})
			return operationErr
		}},
		{"PATCH unknown RackType", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), RackType: applicationdcim.FieldValue(shared.ID(999998))})
			return operationErr
		}},
		{"PATCH unknown RackRole", func() error {
			_, operationErr := service.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{ID: rack.ID(), Role: applicationdcim.FieldValue(shared.ID(999999))})
			return operationErr
		}},
	} {
		assertRejected(rejected.name, rejected.operation)
	}

	failureBefore := loadPostgresRackPresenceState(t, db, rack.ID())
	recorderFailure := errors.New("forced PostgreSQL Rack change recording failure")
	failingRecorder := &failingRackPresenceRecorder{err: recorderFailure}
	failingService := newPostgresRackPresenceService(
		t, db, clearedAt.Add(2*time.Minute), failingRecorder,
	)
	_, err = failingService.UpdateRack(t.Context(), principal, applicationdcim.UpdateRackCommand{
		ID: rack.ID(), Description: applicationdcim.FieldValue("must roll back"),
	})
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(t, failureBefore, loadPostgresRackPresenceState(t, db, rack.ID()))

}

type postgresRackPresenceState struct {
	row                   dcimrow.RackRow
	manufacturerCount     int64
	siteCount             int64
	rackTypeCount         int64
	roleCount             int64
	rackCount             int64
	deviceCount           int64
	targetRackChangeCount int64
	allRackChangeCount    int64
	deviceChangeCount     int64
	totalChangeCount      int64
}

func seedPostgresRackPresenceDependencies(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) (shared.ID, shared.ID, shared.ID, shared.ID) {
	t.Helper()
	manufacturer := dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Rack Manufacturer", Slug: "presence-rack-manufacturer",
	}
	require.NoError(t, db.Create(&manufacturer).Error)
	site := dcimrow.SiteRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Rack Site", Slug: "presence-rack-site", Status: "active",
	}
	require.NoError(t, db.Create(&site).Error)
	role := dcimrow.RackRoleRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Rack Role", Slug: "presence-rack-role", Color: "00ff00",
	}
	require.NoError(t, db.Create(&role).Error)
	rackType := dcimrow.RackTypeRow{
		RowMetadata:    dcimrow.RowMetadata{Created: now, LastUpdated: now},
		ManufacturerID: manufacturer.ID, Model: "Presence Rack Type", Slug: "presence-rack-type",
		FormFactor: "wall-frame", Width: 23, UHeight: 24, StartingUnit: 3, DescUnits: true,
	}
	require.NoError(t, db.Create(&rackType).Error)
	return shared.ID(manufacturer.ID), shared.ID(site.ID), shared.ID(rackType.ID), shared.ID(role.ID)
}

func loadPostgresRackPresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresRackPresenceState {
	t.Helper()
	var state postgresRackPresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.ManufacturerRow{}).Count(&state.manufacturerCount).Error)
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&state.siteCount).Error)
	require.NoError(t, db.Model(&dcimrow.RackTypeRow{}).Count(&state.rackTypeCount).Error)
	require.NoError(t, db.Model(&dcimrow.RackRoleRow{}).Count(&state.roleCount).Error)
	require.NoError(t, db.Model(&dcimrow.RackRow{}).Count(&state.rackCount).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.RackObjectType, id.Int64(),
	).Count(&state.targetRackChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.RackObjectType,
	).Count(&state.allRackChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.DeviceObjectType,
	).Count(&state.deviceChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requirePostgresRackMutationRecorded(
	t *testing.T,
	before postgresRackPresenceState,
	after postgresRackPresenceState,
	expectedLastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, before.manufacturerCount, after.manufacturerCount)
	require.Equal(t, before.siteCount, after.siteCount)
	require.Equal(t, before.rackTypeCount, after.rackTypeCount)
	require.Equal(t, before.roleCount, after.roleCount)
	require.Equal(t, before.rackCount, after.rackCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Zero(t, after.deviceCount, "scalar presence requires zero Devices")
	require.Equal(t, before.targetRackChangeCount+1, after.targetRackChangeCount)
	require.Equal(t, before.allRackChangeCount+1, after.allRackChangeCount)
	require.Equal(t, before.deviceChangeCount, after.deviceChangeCount)
	require.Zero(t, after.deviceChangeCount, "scalar presence records no Device changes")
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.True(t, after.row.Created.Equal(before.row.Created))
	require.True(t, after.row.LastUpdated.Equal(expectedLastUpdated))
}

func requirePostgresRackRow(t *testing.T, actual, expected dcimrow.RackRow) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.True(t, actual.Created.Equal(expected.Created))
	require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
	require.Equal(t, expected.SiteID, actual.SiteID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.FacilityID, actual.FacilityID)
	require.Equal(t, expected.RackTypeID, actual.RackTypeID)
	require.Equal(t, expected.Status, actual.Status)
	require.Equal(t, expected.RoleID, actual.RoleID)
	require.Equal(t, expected.Serial, actual.Serial)
	require.Equal(t, expected.AssetTag, actual.AssetTag)
	require.Equal(t, expected.FormFactor, actual.FormFactor)
	require.Equal(t, expected.Width, actual.Width)
	require.Equal(t, expected.UHeight, actual.UHeight)
	require.Equal(t, expected.StartingUnit, actual.StartingUnit)
	require.Equal(t, expected.DescUnits, actual.DescUnits)
	require.Equal(t, expected.Airflow, actual.Airflow)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Comments, actual.Comments)
}

func requirePostgresRackAggregate(
	t *testing.T,
	rack *domaindcim.Rack,
	siteID shared.ID,
	siteDisplay string,
	name string,
	facilityID *string,
	rackTypeID *shared.ID,
	status string,
	roleID *shared.ID,
	serial string,
	assetTag *string,
	formFactor *string,
	width uint32,
	uHeight uint32,
	startingUnit uint32,
	descUnits bool,
	airflow *string,
	description string,
	comments string,
	created time.Time,
	lastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, siteID, rack.Site().ID())
	require.Equal(t, siteDisplay, rack.Site().Display())
	require.Equal(t, name, rack.Name())
	requireRackNullableString(t, rack.FacilityID(), facilityID)
	if rackTypeID == nil {
		require.True(t, rack.RackType().IsNull())
	} else {
		reference, present := rack.RackType().Get()
		require.True(t, present)
		require.Equal(t, *rackTypeID, reference.ID())
		require.Equal(t, "Presence Rack Type", reference.Display())
	}
	require.Equal(t, status, rack.Status().String())
	if roleID == nil {
		require.True(t, rack.Role().IsNull())
	} else {
		reference, present := rack.Role().Get()
		require.True(t, present)
		require.Equal(t, *roleID, reference.ID())
		require.Equal(t, "Presence Rack Role", reference.Display())
	}
	require.Equal(t, serial, rack.Serial())
	requireRackNullableString(t, rack.AssetTag(), assetTag)
	if formFactor == nil {
		require.True(t, rack.FormFactor().IsNull())
	} else {
		value, present := rack.FormFactor().Get()
		require.True(t, present)
		require.Equal(t, *formFactor, value.String())
	}
	require.Equal(t, width, rack.Width().Uint32())
	require.Equal(t, uHeight, rack.UHeight())
	require.Equal(t, startingUnit, rack.StartingUnit())
	require.Equal(t, descUnits, rack.DescUnits())
	if airflow == nil {
		require.True(t, rack.Airflow().IsNull())
	} else {
		value, present := rack.Airflow().Get()
		require.True(t, present)
		require.Equal(t, *airflow, value.String())
	}
	require.Equal(t, description, rack.Description())
	require.Equal(t, comments, rack.Comments())
	require.True(t, rack.Created().Equal(created))
	require.True(t, rack.LastUpdated().Equal(lastUpdated))
	require.Zero(t, rack.DeviceCount())
}

func requireRackNullableString(
	t *testing.T,
	actual domaindcim.RackNullable[string],
	expected *string,
) {
	t.Helper()
	if expected == nil {
		require.True(t, actual.IsNull())
		return
	}
	value, present := actual.Get()
	require.True(t, present)
	require.Equal(t, *expected, value)
}

func requirePostgresRackReloadedRelationships(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
	siteID shared.ID,
	rackTypeID *shared.ID,
	roleID *shared.ID,
) {
	t.Helper()
	reloaded, err := NewRackRepository(db).Get(t.Context(), id)
	require.NoError(t, err)
	require.Equal(t, siteID, reloaded.Site().ID())
	require.Equal(t, "Presence Rack Site", reloaded.Site().Display())
	if rackTypeID == nil {
		require.True(t, reloaded.RackType().IsNull())
	} else {
		reference, present := reloaded.RackType().Get()
		require.True(t, present)
		require.Equal(t, *rackTypeID, reference.ID())
		require.Equal(t, "Presence Rack Type", reference.Display())
	}
	if roleID == nil {
		require.True(t, reloaded.Role().IsNull())
	} else {
		reference, present := reloaded.Role().Get()
		require.True(t, present)
		require.Equal(t, *roleID, reference.ID())
		require.Equal(t, "Presence Rack Role", reference.Display())
	}
}

type failingRackPresenceRecorder struct {
	err   error
	calls int
}

func (recorder *failingRackPresenceRecorder) Record(
	context.Context,
	applicationchangelog.Change,
) error {
	recorder.calls++
	return recorder.err
}

func newPostgresRackPresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	recorder applicationchangelog.Recorder,
) *applicationdcim.RackService {
	t.Helper()
	service, err := applicationdcim.NewRackService(
		NewRackRepository(db), NewSiteRepository(db), NewRackTypeRepository(db),
		NewRackRoleRepository(db), postgresTransaction.NewUnitOfWork(db), recorder,
		authz.AllowAll{}, postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}

func rackPresenceStringPointer(value string) *string { return &value }
func rackPresenceInt64Pointer(value int64) *int64    { return &value }
