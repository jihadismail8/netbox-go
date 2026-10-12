package dcim

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestPostgresInterfaceScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	bootstrapInterfacePresenceIPAM(t, db)
	createdAt := time.Date(2026, time.October, 12, 20, 0, 0, 0, time.UTC)
	deviceID := seedPostgresInterfacePresenceDevice(t, db, createdAt)
	repository := NewInterfaceRepository(db)
	recorder := postgreschangelog.NewRecorder(db)
	serviceAt := func(now time.Time) *applicationdcim.InterfaceService {
		return newPostgresInterfacePresenceService(t, db, now, repository, recorder)
	}
	create := applicationdcim.CreateInterfaceCommand{
		Device: applicationdcim.FieldValue(deviceID), Name: applicationdcim.FieldValue("  Durable Interface  "),
		Type: applicationdcim.FieldValue("1000base-t"),
	}
	beforeCreate := loadPostgresInterfacePresenceState(t, db)
	require.Len(t, beforeCreate.devices, 1)
	require.Empty(t, beforeCreate.interfaces)
	require.Empty(t, beforeCreate.changes)
	networkInterface, err := serviceAt(createdAt).CreateInterface(t.Context(), principal, create)
	require.NoError(t, err)
	state := loadPostgresInterfacePresenceState(t, db)
	expected := dcimrow.InterfaceRow{
		RowMetadata: dcimrow.RowMetadata{ID: networkInterface.ID().Int64(), Created: createdAt, LastUpdated: createdAt},
		DeviceID:    deviceID.Int64(), Name: "Durable Interface", Type: "1000base-t", Enabled: true,
	}
	requirePostgresInterfacePresenceRow(t, state.interfaces[0], expected)
	requirePostgresInterfacePresenceChange(t, state, beforeCreate, nil, expected, "create", createdAt)
	requirePostgresInterfacePresenceReload(t, repository, networkInterface)

	concreteAt := createdAt.Add(time.Minute)
	patched, err := serviceAt(concreteAt).UpdateInterface(t.Context(), principal, applicationdcim.UpdateInterfaceCommand{
		ID: networkInterface.ID(), Device: applicationdcim.FieldValue(deviceID),
		Name: applicationdcim.FieldValue("  Durable Concrete  "), Label: applicationdcim.FieldValue("  WAN  "),
		Type: applicationdcim.FieldValue("10gbase-sr"), Enabled: applicationdcim.FieldValue(false),
		MgmtOnly: applicationdcim.FieldValue(true), MTU: applicationdcim.FieldValue(int64(65536)),
		Speed: applicationdcim.FieldValue(int64(0)), Duplex: applicationdcim.FieldValue(""),
		Description: applicationdcim.FieldValue("  durable description  "),
	})
	require.NoError(t, err)
	concrete := expected
	concrete.LastUpdated, concrete.Name, concrete.Label, concrete.Type = concreteAt, "Durable Concrete", "WAN", "10gbase-sr"
	concrete.Enabled, concrete.MgmtOnly = false, true
	concrete.MTU, concrete.Speed, concrete.Duplex = interfacePresencePointer(int64(65536)), interfacePresencePointer(int64(0)), interfacePresencePointer("")
	concrete.Description = "durable description"
	concreteState := loadPostgresInterfacePresenceState(t, db)
	requirePostgresInterfacePresenceRow(t, concreteState.interfaces[0], concrete)
	requirePostgresInterfacePresenceChange(t, concreteState, state, &expected, concrete, "update", concreteAt)
	requirePostgresInterfacePresenceReload(t, repository, patched)

	putAt := concreteAt.Add(time.Minute)
	replaced, err := serviceAt(putAt).ReplaceInterface(t.Context(), principal, applicationdcim.ReplaceInterfaceCommand{
		ID: networkInterface.ID(), CreateInterfaceCommand: applicationdcim.CreateInterfaceCommand{
			Device: applicationdcim.FieldValue(deviceID), Name: applicationdcim.FieldValue("  Durable Replaced  "),
			Type: applicationdcim.FieldValue("25gbase-sr"),
		},
	})
	require.NoError(t, err)
	put := concrete
	put.LastUpdated, put.Name, put.Type = putAt, "Durable Replaced", "25gbase-sr"
	putState := loadPostgresInterfacePresenceState(t, db)
	requirePostgresInterfacePresenceRow(t, putState.interfaces[0], put)
	requirePostgresInterfacePresenceChange(t, putState, concreteState, &concrete, put, "update", putAt)
	requirePostgresInterfacePresenceReload(t, repository, replaced)

	clearAt := putAt.Add(time.Minute)
	cleared, err := serviceAt(clearAt).UpdateInterface(t.Context(), principal, applicationdcim.UpdateInterfaceCommand{
		ID: networkInterface.ID(), Label: applicationdcim.FieldValue(""), Enabled: applicationdcim.FieldValue(false),
		MgmtOnly: applicationdcim.FieldValue(false), MTU: applicationdcim.NullField[int64](),
		Speed: applicationdcim.NullField[int64](), Duplex: applicationdcim.NullField[string](),
		Description: applicationdcim.FieldValue(""),
	})
	require.NoError(t, err)
	clear := put
	clear.LastUpdated, clear.Label, clear.Description, clear.MgmtOnly = clearAt, "", "", false
	clear.MTU, clear.Speed, clear.Duplex = nil, nil, nil
	clearState := loadPostgresInterfacePresenceState(t, db)
	requirePostgresInterfacePresenceRow(t, clearState.interfaces[0], clear)
	requirePostgresInterfacePresenceChange(t, clearState, putState, &put, clear, "update", clearAt)
	requirePostgresInterfacePresenceReload(t, repository, cleared)

	blankAt := clearAt.Add(time.Minute)
	blank, err := serviceAt(blankAt).UpdateInterface(t.Context(), principal, applicationdcim.UpdateInterfaceCommand{
		ID: networkInterface.ID(), MTU: applicationdcim.FieldValue(int64(1)),
		Speed: applicationdcim.FieldValue(int64(2147483647)), Duplex: applicationdcim.FieldValue(""),
	})
	require.NoError(t, err)
	blankRow := clear
	blankRow.LastUpdated, blankRow.MTU, blankRow.Speed, blankRow.Duplex = blankAt, interfacePresencePointer(int64(1)), interfacePresencePointer(int64(2147483647)), interfacePresencePointer("")
	blankState := loadPostgresInterfacePresenceState(t, db)
	requirePostgresInterfacePresenceRow(t, blankState.interfaces[0], blankRow)
	requirePostgresInterfacePresenceChange(t, blankState, clearState, &clear, blankRow, "update", blankAt)
	requirePostgresInterfacePresenceReload(t, repository, blank)

	for _, test := range []struct {
		name    string
		command applicationdcim.UpdateInterfaceCommand
	}{
		{"negative MTU", applicationdcim.UpdateInterfaceCommand{MTU: applicationdcim.FieldValue(int64(-1))}},
		{"zero MTU", applicationdcim.UpdateInterfaceCommand{MTU: applicationdcim.FieldValue(int64(0))}},
		{"high MTU", applicationdcim.UpdateInterfaceCommand{MTU: applicationdcim.FieldValue(int64(65537))}},
		{"negative speed", applicationdcim.UpdateInterfaceCommand{Speed: applicationdcim.FieldValue(int64(-1))}},
		{"high speed", applicationdcim.UpdateInterfaceCommand{Speed: applicationdcim.FieldValue(int64(2147483648))}},
		{"untrimmed type", applicationdcim.UpdateInterfaceCommand{Type: applicationdcim.FieldValue(" 25gbase-sr ")}},
		{"untrimmed duplex", applicationdcim.UpdateInterfaceCommand{Duplex: applicationdcim.FieldValue(" full ")}},
		{"unknown owner and scalar aggregation", applicationdcim.UpdateInterfaceCommand{Device: applicationdcim.FieldValue(shared.ID(999999)), Enabled: applicationdcim.NullField[bool](), MTU: applicationdcim.FieldValue(int64(-1)), Name: applicationdcim.FieldValue(" ")}},
	} {
		t.Run("validation non-effects/"+test.name, func(t *testing.T) {
			before := loadPostgresInterfacePresenceState(t, db)
			test.command.ID = networkInterface.ID()
			_, err := serviceAt(blankAt.Add(time.Minute)).UpdateInterface(t.Context(), principal, test.command)
			require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
			require.Equal(t, before, loadPostgresInterfacePresenceState(t, db))
		})
	}
	for _, operation := range []string{"create", "replace", "update"} {
		t.Run("required non-effects/"+operation, func(t *testing.T) {
			before := loadPostgresInterfacePresenceState(t, db)
			service := serviceAt(blankAt.Add(time.Minute))
			var operationErr error
			switch operation {
			case "create":
				_, operationErr = service.CreateInterface(t.Context(), principal, applicationdcim.CreateInterfaceCommand{})
			case "replace":
				_, operationErr = service.ReplaceInterface(t.Context(), principal, applicationdcim.ReplaceInterfaceCommand{ID: networkInterface.ID()})
			case "update":
				_, operationErr = service.UpdateInterface(t.Context(), principal, applicationdcim.UpdateInterfaceCommand{ID: networkInterface.ID(), Name: applicationdcim.NullField[string]()})
			}
			require.True(t, shared.HasReason(operationErr, shared.ErrorReasonValidation), operationErr)
			require.Equal(t, before, loadPostgresInterfacePresenceState(t, db))
		})
	}

	for _, failure := range []string{"repository", "recorder"} {
		for _, operation := range []string{"create", "replace", "update"} {
			t.Run("delegate then error rollback/"+failure+"/"+operation, func(t *testing.T) {
				before := loadPostgresInterfacePresenceState(t, db)
				sentinel := errors.New("forced Interface failure after delegate")
				failingRepository := &delegateThenErrorInterfacePresenceRepository{InterfaceRepository: repository, err: sentinel}
				failingRecorder := &delegateThenErrorInterfacePresenceRecorder{delegate: recorder, err: sentinel}
				var activeRepository applicationdcim.InterfaceRepository = repository
				var activeRecorder applicationchangelog.Recorder = recorder
				if failure == "repository" {
					activeRepository = failingRepository
				} else {
					activeRecorder = failingRecorder
				}
				service := newPostgresInterfacePresenceService(t, db, blankAt.Add(2*time.Minute), activeRepository, activeRecorder)
				var operationErr error
				switch operation {
				case "create":
					_, operationErr = service.CreateInterface(t.Context(), principal, applicationdcim.CreateInterfaceCommand{
						Device: applicationdcim.FieldValue(deviceID), Name: applicationdcim.FieldValue("rollback-create"), Type: applicationdcim.FieldValue("1000base-t"),
					})
				case "replace":
					_, operationErr = service.ReplaceInterface(t.Context(), principal, applicationdcim.ReplaceInterfaceCommand{
						ID: networkInterface.ID(), CreateInterfaceCommand: applicationdcim.CreateInterfaceCommand{
							Device: applicationdcim.FieldValue(deviceID), Name: applicationdcim.FieldValue("rollback-replace"), Type: applicationdcim.FieldValue("1000base-t"),
						},
					})
				case "update":
					_, operationErr = service.UpdateInterface(t.Context(), principal, applicationdcim.UpdateInterfaceCommand{
						ID: networkInterface.ID(), Description: applicationdcim.FieldValue("rollback-update"),
					})
				}
				require.ErrorIs(t, operationErr, sentinel)
				if failure == "repository" {
					require.Equal(t, 1, failingRepository.calls)
				} else {
					require.Equal(t, 1, failingRecorder.calls)
				}
				require.Equal(t, before, loadPostgresInterfacePresenceState(t, db), "persisted rows and inserted object changes must both roll back")
			})
		}
	}
}

func bootstrapInterfacePresenceIPAM(t *testing.T, db *gorm.DB) {
	t.Helper()
	var entries []bootstrap.Entry
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	for _, descriptor := range ipamrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Equal(t, []string{"go_ipam_vrfs", "go_ipam_prefixes", "go_ipam_ip_addresses"}, result.Created)
}

func seedPostgresInterfacePresenceDevice(t *testing.T, db *gorm.DB, now time.Time) shared.ID {
	t.Helper()
	manufacturer := seedInterfaceTemplatePresenceManufacturer(t, db, now)
	deviceType := seedInterfaceTemplatePresenceDeviceType(t, db, now, manufacturer, "Interface Presence Type", "interface-presence-type")
	site, role := seedDeviceTypeDeviceDependencies(t, db)
	name := "interface-presence-device"
	device := dcimrow.DeviceRow{
		RowMetadata:  dcimrow.RowMetadata{Created: now, LastUpdated: now},
		DeviceTypeID: deviceType.Int64(), RoleID: role, SiteID: site, Name: &name, Status: "active",
	}
	require.NoError(t, db.Create(&device).Error)
	return shared.ID(device.ID)
}

type postgresInterfacePresenceState struct {
	interfaces []dcimrow.InterfaceRow
	devices    []dcimrow.DeviceRow
	changes    []postgreschangelog.ChangeRow
}

func loadPostgresInterfacePresenceState(t *testing.T, db *gorm.DB) postgresInterfacePresenceState {
	t.Helper()
	var state postgresInterfacePresenceState
	require.NoError(t, db.Order("id").Find(&state.interfaces).Error)
	require.NoError(t, db.Order("id").Find(&state.devices).Error)
	require.NoError(t, db.Order("id").Find(&state.changes).Error)
	var templates, addresses int64
	require.NoError(t, db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&templates).Error)
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).Count(&addresses).Error)
	require.Zero(t, templates, "I11 fixtures never instantiate templates")
	require.Zero(t, addresses, "I11 fixtures never assign IPAddresses")
	return state
}

func requirePostgresInterfacePresenceRow(t *testing.T, actual, expected dcimrow.InterfaceRow) {
	t.Helper()
	require.True(t, actual.Created.Equal(expected.Created))
	require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
	actual.Created, actual.LastUpdated = expected.Created, expected.LastUpdated
	require.Equal(t, expected, actual)
}

func requirePostgresInterfacePresenceChange(t *testing.T, after, before postgresInterfacePresenceState, prior *dcimrow.InterfaceRow, current dcimrow.InterfaceRow, action string, now time.Time) {
	t.Helper()
	require.Equal(t, before.devices, after.devices)
	require.Len(t, after.interfaces, 1)
	require.Len(t, after.changes, len(before.changes)+1)
	require.Equal(t, before.changes, after.changes[:len(before.changes)])
	change := after.changes[len(after.changes)-1]
	require.Positive(t, change.ID)
	require.Equal(t, int64(1), change.ActorID)
	require.Equal(t, action, change.Action)
	require.Equal(t, domaindcim.InterfaceObjectType, change.Kind)
	require.Equal(t, current.ID, change.ObjectID)
	require.True(t, change.OccurredAt.Equal(now))
	if prior == nil {
		require.Empty(t, change.BeforeData)
	} else {
		encoded, err := json.Marshal(prior)
		require.NoError(t, err)
		require.JSONEq(t, string(encoded), string(change.BeforeData))
	}
	encoded, err := json.Marshal(current)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(change.AfterData))
}

func requirePostgresInterfacePresenceReload(t *testing.T, repository *InterfaceRepository, expected *domaindcim.Interface) {
	t.Helper()
	actual, err := repository.Get(t.Context(), expected.ID())
	require.NoError(t, err)
	require.Equal(t, expected.State(), actual.State())
	require.Zero(t, actual.IPAddressCount())
}

type delegateThenErrorInterfacePresenceRepository struct {
	applicationdcim.InterfaceRepository
	err   error
	calls int
}

func (repository *delegateThenErrorInterfacePresenceRepository) Create(ctx context.Context, item *domaindcim.Interface) error {
	repository.calls++
	if err := repository.InterfaceRepository.Create(ctx, item); err != nil {
		return err
	}
	return repository.err
}

func (repository *delegateThenErrorInterfacePresenceRepository) Update(ctx context.Context, item *domaindcim.Interface) error {
	repository.calls++
	if err := repository.InterfaceRepository.Update(ctx, item); err != nil {
		return err
	}
	return repository.err
}

type delegateThenErrorInterfacePresenceRecorder struct {
	delegate applicationchangelog.Recorder
	err      error
	calls    int
}

func (recorder *delegateThenErrorInterfacePresenceRecorder) Record(ctx context.Context, change applicationchangelog.Change) error {
	recorder.calls++
	if err := recorder.delegate.Record(ctx, change); err != nil {
		return err
	}
	return recorder.err
}

type interfacePresenceNoCascade struct{}

func (interfacePresenceNoCascade) DeleteAssignedToInterface(context.Context, shared.ID, shared.Timestamp) ([]applicationipam.IPAddressCascadeChange, error) {
	panic("I11 must not invoke Interface deletion or IPAddress cascade")
}

func newPostgresInterfacePresenceService(t *testing.T, db *gorm.DB, now time.Time, repository applicationdcim.InterfaceRepository, recorder applicationchangelog.Recorder) *applicationdcim.InterfaceService {
	t.Helper()
	service, err := applicationdcim.NewInterfaceService(repository, NewDeviceRepository(db), interfacePresenceNoCascade{}, postgresTransaction.NewUnitOfWork(db), recorder, authz.AllowAll{}, postgresConcurrencyClock{now: shared.NewTimestamp(now)})
	require.NoError(t, err)
	return service
}

func interfacePresencePointer[T any](value T) *T { return &value }
