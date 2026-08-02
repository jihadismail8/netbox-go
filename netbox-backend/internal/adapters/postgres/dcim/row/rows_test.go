package row_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
)

func TestDCIMDescriptorsPreserveExactOwnershipAndBootstrapOrder(t *testing.T) {
	descriptors := dcimrow.Descriptors()
	require.Equal(t, []string{
		"go_dcim_sites",
		"go_dcim_manufacturers",
		"go_dcim_rack_roles",
		"go_dcim_rack_types",
		"go_dcim_racks",
		"go_dcim_device_roles",
		"go_dcim_device_types",
		"go_dcim_interface_templates",
		"go_dcim_devices",
		"go_dcim_interfaces",
	}, descriptorNames(descriptors))
	require.Equal(t, map[string][]string{
		"go_dcim_sites":               nil,
		"go_dcim_manufacturers":       nil,
		"go_dcim_rack_roles":          nil,
		"go_dcim_rack_types":          {"go_dcim_manufacturers"},
		"go_dcim_racks":               {"go_dcim_sites", "go_dcim_rack_roles", "go_dcim_rack_types"},
		"go_dcim_device_roles":        nil,
		"go_dcim_device_types":        {"go_dcim_manufacturers"},
		"go_dcim_interface_templates": {"go_dcim_device_types"},
		"go_dcim_devices":             {"go_dcim_sites", "go_dcim_racks", "go_dcim_device_roles", "go_dcim_device_types"},
		"go_dcim_interfaces":          {"go_dcim_devices"},
	}, descriptorDependencies(descriptors))
	require.Len(t, dcimrow.Models(), len(descriptors))

	db := openDCIMRowTestDB(t)
	for _, descriptor := range descriptors {
		statement := &gorm.Statement{DB: db}
		require.NoError(t, statement.Parse(descriptor.Model))
		require.Equal(t, descriptor.Name, statement.Schema.Table)
		require.Nil(t, statement.Schema.LookUpField("Data"))
		require.Nil(t, statement.Schema.LookUpField("Kind"))
	}
}

func TestDCIMBootstrapCreatesOnlyMissingTypedTablesWithConstraints(t *testing.T) {
	db := openDCIMRowTestDB(t)
	descriptors := dcimrow.Descriptors()
	registry, err := bootstrap.NewRegistry(dcimEntries(descriptors)...)
	require.NoError(t, err)

	first, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, first.Created, len(descriptors))
	require.Empty(t, first.Existing)
	second, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Empty(t, second.Created)
	require.Len(t, second.Existing, len(descriptors))

	for _, descriptor := range descriptors {
		require.True(t, db.Migrator().HasTable(descriptor.Model), descriptor.Name)
		statement := &gorm.Statement{DB: db}
		require.NoError(t, statement.Parse(descriptor.Model))
		for _, relationship := range statement.Schema.Relationships.Relations {
			constraint := relationship.ParseConstraint()
			if constraint == nil || constraint.Schema != statement.Schema {
				continue
			}
			require.True(
				t, db.Migrator().HasConstraint(descriptor.Model, constraint.Name),
				"%s is missing FK %s", descriptor.Name, constraint.Name,
			)
		}
	}
	require.True(t, db.Migrator().HasIndex(&dcimrow.SiteRow{}, "uq_go_site_slug"))
	require.True(t, db.Migrator().HasIndex(&dcimrow.RackTypeRow{}, "uq_go_rack_type_model"))
	require.True(t, db.Migrator().HasIndex(&dcimrow.DeviceTypeRow{}, "uq_go_device_type_slug"))
	require.True(t, db.Migrator().HasIndex(&dcimrow.InterfaceRow{}, "uq_go_interface_name"))
}

func TestDCIMRowsRetainStaticProfileConstraints(t *testing.T) {
	db := openDCIMRowTestDB(t)
	require.NoError(t, db.AutoMigrate(dcimrow.Models()...))
	now := time.Now().UTC()
	metadata := func() dcimrow.RowMetadata {
		return dcimrow.RowMetadata{Created: now, LastUpdated: now}
	}

	siteOne := dcimrow.SiteRow{RowMetadata: metadata(), Name: "One", Slug: "one", Status: "active"}
	siteTwo := dcimrow.SiteRow{RowMetadata: metadata(), Name: "Two", Slug: "two", Status: "active"}
	require.NoError(t, db.Create(&siteOne).Error)
	require.NoError(t, db.Create(&siteTwo).Error)
	require.Error(t, db.Create(&dcimrow.RackRow{
		RowMetadata: metadata(), SiteID: 999, Name: "invalid", Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	}).Error, "site FK must reject an unknown Site")

	rootRole := dcimrow.DeviceRoleRow{
		RowMetadata: metadata(), Name: "Router", Slug: "router", Color: "9e9e9e",
	}
	require.NoError(t, db.Create(&rootRole).Error)
	require.Error(t, db.Create(&dcimrow.DeviceRoleRow{
		RowMetadata: metadata(), Name: "Router", Slug: "router-two", Color: "9e9e9e",
	}).Error, "root DeviceRole names must be unique despite a NULL parent")

	manufacturer := dcimrow.ManufacturerRow{
		RowMetadata: metadata(), Name: "Acme", Slug: "acme",
	}
	require.NoError(t, db.Create(&manufacturer).Error)
	deviceType := dcimrow.DeviceTypeRow{
		RowMetadata: metadata(), ManufacturerID: manufacturer.ID,
		Model: "Edge", Slug: "edge", UHeight: 1, IsFullDepth: true,
	}
	require.NoError(t, db.Create(&deviceType).Error)
	require.Error(t, db.Create(&dcimrow.DeviceTypeRow{
		RowMetadata: metadata(), ManufacturerID: manufacturer.ID,
		Model: "Different", Slug: "edge", UHeight: 1, IsFullDepth: true,
	}).Error, "DeviceType slug must be unique within a manufacturer")

	edge, upperEdge := "edge-1", "EDGE-1"
	require.NoError(t, db.Create(&dcimrow.DeviceRow{
		RowMetadata: metadata(), DeviceTypeID: deviceType.ID, RoleID: rootRole.ID,
		Name: &edge, SiteID: siteOne.ID, Status: "active",
	}).Error)
	crossSite := dcimrow.DeviceRow{
		RowMetadata: metadata(), DeviceTypeID: deviceType.ID, RoleID: rootRole.ID,
		Name: &upperEdge, SiteID: siteOne.ID, Status: "active",
	}
	require.Error(t, db.Create(&crossSite).Error, "Device names must be case-insensitively unique per Site")
	crossSite.ID = 0
	crossSite.SiteID = siteTwo.ID
	require.NoError(t, db.Create(&crossSite).Error, "the same Device name is valid at another Site")
	require.Error(t, db.Delete(&deviceType).Error, "RESTRICT must reject direct deletion of referenced rows")
}

func dcimEntries(descriptors []dcimrow.Descriptor) []bootstrap.Entry {
	entries := make([]bootstrap.Entry, 0, len(descriptors))
	for _, descriptor := range descriptors {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
	}
	return entries
}

func descriptorNames(descriptors []dcimrow.Descriptor) []string {
	names := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		names = append(names, descriptor.Name)
	}
	return names
}

func descriptorDependencies(descriptors []dcimrow.Descriptor) map[string][]string {
	dependencies := make(map[string][]string, len(descriptors))
	for _, descriptor := range descriptors {
		dependencies[descriptor.Name] = descriptor.Dependencies
	}
	return dependencies
}

func openDCIMRowTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(
		sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared&_foreign_keys=on"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	return db
}
