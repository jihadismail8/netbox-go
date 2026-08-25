package dcim_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestCreateInterfaceTemplateAppliesDefaultsAndRecordsTypedChange(t *testing.T) {
	service, repository, recorder, _ := newInterfaceTemplateApplicationService(
		t, authz.AllowAll{},
	)
	template, err := service.CreateInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceTemplateCommand{
			DeviceType: appdcim.FieldValue(shared.ID(9)),
			Name:       appdcim.FieldValue(" Ethernet1 "),
			Type:       appdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)

	assert.Equal(t, shared.ID(41), template.ID())
	assert.Equal(t, "Ethernet1", template.Name())
	assert.Empty(t, template.Label())
	assert.True(t, template.Enabled())
	assert.False(t, template.MgmtOnly())
	assert.Empty(t, template.Description())
	assert.Equal(t, 1, repository.createCalls)
	require.Len(t, recorder.changes, 1)
	assert.Equal(t, dcimdomain.InterfaceTemplateObjectType, recorder.changes[0].ObjectType)
	assert.IsType(t, dcimdomain.InterfaceTemplateSnapshot{}, recorder.changes[0].After)
}

func TestInterfaceTemplatePUTResetsDefaultsAndPATCHPreservesPresence(t *testing.T) {
	service, _, recorder, _ := newInterfaceTemplateApplicationService(t, authz.AllowAll{})
	template, err := service.CreateInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceTemplateCommand{
			DeviceType:  appdcim.FieldValue(shared.ID(9)),
			Name:        appdcim.FieldValue("Ethernet1"),
			Label:       appdcim.FieldValue("Old"),
			Type:        appdcim.FieldValue("1000base-t"),
			Enabled:     appdcim.FieldValue(false),
			MgmtOnly:    appdcim.FieldValue(true),
			Description: appdcim.FieldValue("Old description"),
		},
	)
	require.NoError(t, err)

	template, err = service.UpdateInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.UpdateInterfaceTemplateCommand{
			ID: template.ID(), Label: appdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	assert.Empty(t, template.Label())
	assert.False(t, template.Enabled(), "omitted PATCH fields must be preserved")
	assert.True(t, template.MgmtOnly())
	assert.Equal(t, "Old description", template.Description())

	template, err = service.ReplaceInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.ReplaceInterfaceTemplateCommand{
			ID: template.ID(),
			CreateInterfaceTemplateCommand: appdcim.CreateInterfaceTemplateCommand{
				DeviceType: appdcim.FieldValue(shared.ID(9)),
				Name:       appdcim.FieldValue("Ethernet2"),
				Type:       appdcim.FieldValue("10gbase-sr"),
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Ethernet2", template.Name())
	assert.Empty(t, template.Label())
	assert.False(t, template.Enabled(), "omitted PUT fields must be preserved")
	assert.True(t, template.MgmtOnly())
	assert.Equal(t, "Old description", template.Description())
	assert.Len(t, recorder.changes, 3)
}

func TestInterfaceTemplateRejectsMoveAndUnknownTypeBeforeWrite(t *testing.T) {
	service, repository, _, reader := newInterfaceTemplateApplicationService(
		t, authz.AllowAll{},
	)
	reader.deviceTypes[10] = newInterfaceTemplateDeviceTypeFixture(t, 10, "Switch", "switch")
	template, err := service.CreateInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.CreateInterfaceTemplateCommand{
			DeviceType: appdcim.FieldValue(shared.ID(9)),
			Name:       appdcim.FieldValue("Ethernet1"),
			Type:       appdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)

	_, err = service.UpdateInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.UpdateInterfaceTemplateCommand{
			ID: template.ID(), DeviceType: appdcim.FieldValue(shared.ID(10)),
		},
	)
	require.Error(t, err)
	assert.Equal(t, "device_type", shared.ViolationsOf(err)[0].Field)
	assert.Zero(t, repository.updateCalls)

	_, err = service.UpdateInterfaceTemplate(
		t.Context(), testPrincipal(), appdcim.UpdateInterfaceTemplateCommand{
			ID: template.ID(), Type: appdcim.FieldValue("unknown"),
		},
	)
	require.Error(t, err)
	assert.Equal(t, "type", shared.ViolationsOf(err)[0].Field)
	assert.Zero(t, repository.updateCalls)
}

func TestInterfaceTemplateListPinsFiltersOrderingAndRBAC(t *testing.T) {
	service, repository, _, _ := newInterfaceTemplateApplicationService(
		t, authz.PermissionAuthorizer{},
	)
	viewer := identity.Principal{
		ID: 7, Username: "viewer",
		Permissions: map[string]struct{}{"dcim.view_interfacetemplate": {}},
	}
	enabled := false
	mgmtOnly := true
	_, err := service.ListInterfaceTemplates(
		t.Context(), viewer, appdcim.ListInterfaceTemplatesQuery{
			LimitPresent: true, IDs: []int64{-1, 0, 41},
			DeviceTypeIDs: []int64{-1, 9}, Names: []string{" Ethernet1 ", " Ethernet2 "},
			Types:   []string{"1000base-t", "10gbase-sr"},
			Enabled: &enabled, MgmtOnly: &mgmtOnly,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, appdcim.MaximumInterfaceTemplatePageLimit, repository.criteria.Limit)
	assert.Equal(t, []int64{-1, 0, 41}, repository.criteria.IDs)
	assert.Equal(t, []int64{-1, 9}, repository.criteria.DeviceTypeIDs)
	assert.Equal(t, []string{"Ethernet1", "Ethernet2"}, repository.criteria.Names)
	assert.Equal(t, []dcimdomain.InterfaceType{"1000base-t", "10gbase-sr"}, repository.criteria.Types)
	require.NotNil(t, repository.criteria.Enabled)
	assert.False(t, *repository.criteria.Enabled)
	require.NotNil(t, repository.criteria.MgmtOnly)
	assert.True(t, *repository.criteria.MgmtOnly)
	assert.Equal(t, []appdcim.InterfaceTemplateSort{
		{Field: appdcim.InterfaceTemplateSortDeviceType},
		{Field: appdcim.InterfaceTemplateSortName},
		{Field: appdcim.InterfaceTemplateSortID},
	}, repository.criteria.Ordering)

	_, err = service.CreateInterfaceTemplate(
		t.Context(), viewer, appdcim.CreateInterfaceTemplateCommand{},
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonForbidden))
	assert.Zero(t, repository.createCalls)

	_, err = service.ListInterfaceTemplates(
		t.Context(), viewer,
		appdcim.ListInterfaceTemplatesQuery{Types: []string{"unknown"}},
	)
	require.Error(t, err)
	assert.Equal(t, 1, repository.listCalls)
}

func newInterfaceTemplateApplicationService(
	t *testing.T,
	authorizer authz.ResourceAuthorizer,
) (
	*appdcim.InterfaceTemplateService,
	*interfaceTemplateApplicationRepository,
	*interfaceTemplateApplicationRecorder,
	*interfaceTemplateDeviceTypeReader,
) {
	t.Helper()
	reader := &interfaceTemplateDeviceTypeReader{
		deviceTypes: map[shared.ID]*dcimdomain.DeviceType{
			9: newInterfaceTemplateDeviceTypeFixture(t, 9, "Router", "router"),
		},
	}
	repository := &interfaceTemplateApplicationRepository{
		templates: make(map[shared.ID]dcimdomain.InterfaceTemplateState),
	}
	recorder := &interfaceTemplateApplicationRecorder{}
	service, err := appdcim.NewInterfaceTemplateService(
		repository, reader, interfaceTemplateApplicationUnitOfWork{},
		recorder, authorizer, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service, repository, recorder, reader
}

func newInterfaceTemplateDeviceTypeFixture(
	t *testing.T,
	id shared.ID,
	model string,
	slug string,
) *dcimdomain.DeviceType {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturerReference(1, "Vendor", "vendor")
	require.NoError(t, err)
	deviceType, err := dcimdomain.NewDeviceType(dcimdomain.DeviceTypeValues{
		Manufacturer: manufacturer, Model: model, Slug: slug,
		UHeight: dcimdomain.DeviceTypeDefaultHeight, IsFullDepth: true,
		Airflow: dcimdomain.NullDeviceAirflow(),
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, deviceType.AssignID(id))
	return deviceType
}

type interfaceTemplateDeviceTypeReader struct {
	deviceTypes map[shared.ID]*dcimdomain.DeviceType
}

func (reader *interfaceTemplateDeviceTypeReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	deviceType := reader.deviceTypes[id]
	if deviceType == nil {
		return nil, shared.NotFound("DeviceType", id)
	}
	return deviceType, nil
}

type interfaceTemplateApplicationRepository struct {
	templates   map[shared.ID]dcimdomain.InterfaceTemplateState
	criteria    appdcim.InterfaceTemplateListCriteria
	listCalls   int
	createCalls int
	updateCalls int
	deleteCalls int
}

func (repository *interfaceTemplateApplicationRepository) List(
	_ context.Context,
	criteria appdcim.InterfaceTemplateListCriteria,
) (appdcim.InterfaceTemplatePage, error) {
	repository.listCalls++
	repository.criteria = criteria
	return appdcim.InterfaceTemplatePage{}, nil
}

func (repository *interfaceTemplateApplicationRepository) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.InterfaceTemplate, error) {
	return repository.restore(id)
}

func (repository *interfaceTemplateApplicationRepository) GetForUpdate(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.InterfaceTemplate, error) {
	return repository.restore(id)
}

func (repository *interfaceTemplateApplicationRepository) restore(
	id shared.ID,
) (*dcimdomain.InterfaceTemplate, error) {
	state, present := repository.templates[id]
	if !present {
		return nil, shared.NotFound("InterfaceTemplate", id)
	}
	return dcimdomain.RestoreInterfaceTemplate(state)
}

func (repository *interfaceTemplateApplicationRepository) Create(
	_ context.Context,
	template *dcimdomain.InterfaceTemplate,
) error {
	repository.createCalls++
	if err := template.AssignID(41); err != nil {
		return err
	}
	repository.templates[template.ID()] = template.State()
	return nil
}

func (repository *interfaceTemplateApplicationRepository) Update(
	_ context.Context,
	template *dcimdomain.InterfaceTemplate,
) error {
	repository.updateCalls++
	repository.templates[template.ID()] = template.State()
	return nil
}

func (repository *interfaceTemplateApplicationRepository) Delete(
	_ context.Context,
	template *dcimdomain.InterfaceTemplate,
) error {
	repository.deleteCalls++
	delete(repository.templates, template.ID())
	return nil
}

type interfaceTemplateApplicationUnitOfWork struct{}

func (interfaceTemplateApplicationUnitOfWork) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	return operation(ctx)
}

type interfaceTemplateApplicationRecorder struct {
	changes []changelog.Change
}

func (recorder *interfaceTemplateApplicationRecorder) Record(
	_ context.Context,
	change changelog.Change,
) error {
	recorder.changes = append(recorder.changes, change)
	return nil
}
