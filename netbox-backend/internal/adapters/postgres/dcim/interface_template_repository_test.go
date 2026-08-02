package dcim

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceTemplateRepositoryMapsFiltersOrderingAndUniqueness(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	deviceTypeRepository := NewDeviceTypeRepository(db)
	repository := NewInterfaceTemplateRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	router := newInterfaceTemplateDeviceTypeFixture(t, manufacturer, "Router", "router")
	switchType := newInterfaceTemplateDeviceTypeFixture(t, manufacturer, "Switch", "switch")
	require.NoError(t, deviceTypeRepository.Create(t.Context(), router))
	require.NoError(t, deviceTypeRepository.Create(t.Context(), switchType))

	fixtures := []*domaindcim.InterfaceTemplate{
		newInterfaceTemplateRepositoryFixture(
			t, switchType, "Ethernet9", "1000base-t", true, false,
		),
		newInterfaceTemplateRepositoryFixture(
			t, router, "Ethernet2", "10gbase-sr", false, true,
		),
		newInterfaceTemplateRepositoryFixture(
			t, router, "Ethernet1", "10gbase-sr", false, true,
		),
	}
	for _, template := range fixtures {
		require.NoError(t, repository.Create(t.Context(), template))
	}

	loaded, err := repository.Get(t.Context(), fixtures[1].ID())
	require.NoError(t, err)
	assert.Equal(t, router.ID(), loaded.DeviceType().ID())
	assert.Equal(t, "Router", loaded.DeviceType().Display())
	assert.Equal(t, "router", loaded.DeviceType().Slug().String())

	enabled := false
	mgmtOnly := true
	page, err := repository.List(t.Context(), applicationdcim.InterfaceTemplateListCriteria{
		Limit: 50, DeviceTypeIDs: []int64{-1, router.ID().Int64()},
		Types:   []domaindcim.InterfaceType{"10gbase-sr"},
		Enabled: &enabled, MgmtOnly: &mgmtOnly, Query: "template",
		Ordering: []applicationdcim.InterfaceTemplateSort{
			{Field: applicationdcim.InterfaceTemplateSortDeviceType},
			{Field: applicationdcim.InterfaceTemplateSortName},
			{Field: applicationdcim.InterfaceTemplateSortID},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, []string{"Ethernet1", "Ethernet2"}, []string{
		page.Results[0].Name(), page.Results[1].Name(),
	})

	duplicate := newInterfaceTemplateRepositoryFixture(
		t, router, "Ethernet1", "other", true, false,
	)
	err = repository.Create(t.Context(), duplicate)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Equal(t, []shared.FieldViolation{{
		Field: "non_field_errors", Reason: "unique_together",
		Description: "The fields device_type, name must make a unique set.",
	}}, shared.ViolationsOf(err))
	assert.False(t, duplicate.ID().IsValid())

	sameNameOtherType := newInterfaceTemplateRepositoryFixture(
		t, switchType, "Ethernet1", "other", true, false,
	)
	require.NoError(t, repository.Create(t.Context(), sameNameOtherType))
}

func TestInterfaceTemplateServiceRollsBackPersistenceWhenAuditFails(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	manufacturerRepository := NewManufacturerRepository(db)
	deviceTypeRepository := NewDeviceTypeRepository(db)
	repository := NewInterfaceTemplateRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	deviceType := newInterfaceTemplateDeviceTypeFixture(t, manufacturer, "Router", "router")
	require.NoError(t, deviceTypeRepository.Create(t.Context(), deviceType))
	template := newInterfaceTemplateRepositoryFixture(
		t, deviceType, "Ethernet1", "1000base-t", true, false,
	)
	require.NoError(t, repository.Create(t.Context(), template))

	sentinel := errors.New("forced interface-template audit failure")
	recorder := &failOnRecorder{
		delegate: postgreschangelog.NewRecorder(db), failAt: 1, err: sentinel,
	}
	service, err := applicationdcim.NewInterfaceTemplateService(
		repository, deviceTypeRepository, postgresTransaction.NewUnitOfWork(db),
		recorder, authz.AllowAll{}, interfaceTemplateRepositoryClock{},
	)
	require.NoError(t, err)
	_, err = service.UpdateInterfaceTemplate(
		t.Context(), rackTypePrincipal(), applicationdcim.UpdateInterfaceTemplateCommand{
			ID: template.ID(), Label: applicationdcim.FieldValue("rolled-back"),
		},
	)
	require.ErrorIs(t, err, sentinel)

	loaded, getErr := repository.Get(t.Context(), template.ID())
	require.NoError(t, getErr)
	assert.Empty(t, loaded.Label())
	assertTableCount(t, db, &postgreschangelog.ChangeRow{}, 0)
}

func TestTypedInterfaceTemplateFeedsExactDeviceInstantiationFields(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	manufacturerRepository := NewManufacturerRepository(db)
	deviceTypeRepository := NewDeviceTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	deviceType := newInterfaceTemplateDeviceTypeFixture(t, manufacturer, "Router", "router")
	require.NoError(t, deviceTypeRepository.Create(t.Context(), deviceType))
	siteID := seedOrganizationSite(t, db)
	role := dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "Router", Slug: "router", Color: "9e9e9e", VMRole: true,
	}
	require.NoError(t, db.Create(&role).Error)

	templateService, err := applicationdcim.NewInterfaceTemplateService(
		NewInterfaceTemplateRepository(db), deviceTypeRepository,
		postgresTransaction.NewUnitOfWork(db), postgreschangelog.NewRecorder(db),
		authz.AllowAll{}, interfaceTemplateRepositoryClock{},
	)
	require.NoError(t, err)
	template, err := templateService.CreateInterfaceTemplate(
		t.Context(), rackTypePrincipal(), applicationdcim.CreateInterfaceTemplateCommand{
			DeviceType: applicationdcim.FieldValue(deviceType.ID()),
			Name:       applicationdcim.FieldValue("Ethernet1"),
			Label:      applicationdcim.FieldValue("WAN"),
			Type:       applicationdcim.FieldValue("10gbase-sr"),
			Enabled:    applicationdcim.FieldValue(false),
			MgmtOnly:   applicationdcim.FieldValue(true),
			Description: applicationdcim.FieldValue(
				"Template descriptions are intentionally not instantiated.",
			),
		},
	)
	require.NoError(t, err)

	graph := deviceWorkflowGraph{
		devices: NewDeviceRepository(db), interfaces: NewInterfaceRepository(db),
		templates: NewInterfaceTemplateRepository(db), deviceTypes: deviceTypeRepository,
		roles: NewDeviceRoleRepository(db), sites: NewSiteRepository(db),
		racks: NewRackRepository(db),
	}
	device, err := newPostgresDeviceService(t, db, graph).CreateDevice(
		t.Context(), rackTypePrincipal(), applicationdcim.CreateDeviceCommand{
			DeviceType: applicationdcim.FieldValue(deviceType.ID()),
			Role:       applicationdcim.FieldValue(shared.ID(role.ID)),
			Site:       applicationdcim.FieldValue(shared.ID(siteID)),
			Name:       applicationdcim.FieldValue("router-01"),
		},
	)
	require.NoError(t, err)

	var row dcimrow.InterfaceRow
	require.NoError(t, db.Where("device_id = ?", device.ID().Int64()).Take(&row).Error)
	assert.Equal(t, template.Name(), row.Name)
	assert.Equal(t, template.Label(), row.Label)
	assert.Equal(t, template.Type().String(), row.Type)
	assert.Equal(t, template.Enabled(), row.Enabled)
	assert.Equal(t, template.MgmtOnly(), row.MgmtOnly)
	assert.Empty(t, row.Description)
}

func newInterfaceTemplateDeviceTypeFixture(
	t *testing.T,
	manufacturer *domaindcim.Manufacturer,
	model string,
	slug string,
) *domaindcim.DeviceType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(
		manufacturer.ID(), manufacturer.Name(), manufacturer.Slug().String(),
	)
	require.NoError(t, err)
	deviceType, err := domaindcim.NewDeviceType(domaindcim.DeviceTypeValues{
		Manufacturer: reference, Model: model, Slug: slug,
		UHeight: domaindcim.DeviceTypeDefaultHeight, IsFullDepth: true,
		Airflow: domaindcim.NullDeviceAirflow(),
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return deviceType
}

func newInterfaceTemplateRepositoryFixture(
	t *testing.T,
	deviceType *domaindcim.DeviceType,
	name string,
	interfaceType string,
	enabled bool,
	mgmtOnly bool,
) *domaindcim.InterfaceTemplate {
	t.Helper()
	reference, err := domaindcim.NewDeviceTypeReference(
		deviceType.ID(), deviceType.Model(), deviceType.Slug().String(),
	)
	require.NoError(t, err)
	template, err := domaindcim.NewInterfaceTemplate(domaindcim.InterfaceTemplateValues{
		DeviceType: reference, Name: name, Type: interfaceType,
		Enabled: enabled, MgmtOnly: mgmtOnly, Description: "Template description",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return template
}

type interfaceTemplateRepositoryClock struct{}

func (interfaceTemplateRepositoryClock) Now() shared.Timestamp {
	return repositoryUpdatedAt
}
