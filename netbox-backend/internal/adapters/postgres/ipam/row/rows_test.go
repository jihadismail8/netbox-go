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
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
)

func TestIPAMDescriptorsPreserveExactOwnershipAndBootstrapOrder(t *testing.T) {
	descriptors := ipamrow.Descriptors()
	require.Equal(t, []string{
		"go_ipam_vrfs",
		"go_ipam_prefixes",
		"go_ipam_ip_addresses",
	}, ipamDescriptorNames(descriptors))
	require.Equal(t, map[string][]string{
		"go_ipam_vrfs":         nil,
		"go_ipam_prefixes":     {"go_ipam_vrfs"},
		"go_ipam_ip_addresses": {"go_ipam_vrfs", "go_dcim_interfaces"},
	}, ipamDescriptorDependencies(descriptors))
	require.Len(t, ipamrow.Models(), len(descriptors))

	db := openIPAMRowTestDB(t)
	for _, descriptor := range descriptors {
		statement := &gorm.Statement{DB: db}
		require.NoError(t, statement.Parse(descriptor.Model))
		require.Equal(t, descriptor.Name, statement.Schema.Table)
		require.Nil(t, statement.Schema.LookUpField("Data"))
		require.Nil(t, statement.Schema.LookUpField("Kind"))
	}
}

func TestIPAMBootstrapCreatesOnlyMissingTypedTablesWithConstraints(t *testing.T) {
	db := openIPAMRowTestDB(t)
	dcimDescriptors := dcimrow.Descriptors()
	ipamDescriptors := ipamrow.Descriptors()
	registry, err := bootstrap.NewRegistry(combinedEntries(dcimDescriptors, ipamDescriptors)...)
	require.NoError(t, err)

	first, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, first.Created, len(dcimDescriptors)+len(ipamDescriptors))
	require.Empty(t, first.Existing)
	second, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Empty(t, second.Created)
	require.Len(t, second.Existing, len(dcimDescriptors)+len(ipamDescriptors))

	for _, descriptor := range ipamDescriptors {
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
	require.True(t, db.Migrator().HasIndex(&ipamrow.IPAddressRow{}, "idx_go_ip_assignment"))

	assignmentType := "dcim.interface"
	require.Error(t, db.Create(&ipamrow.IPAddressRow{
		RowMetadata: ipamrow.RowMetadata{Created: time.Now(), LastUpdated: time.Now()},
		Address:     "192.0.2.1/24", Status: "active",
		AssignedObjectType: &assignmentType,
	}).Error, "bootstrap must retain the assignment-pair check constraint")
}

func combinedEntries(
	dcimDescriptors []dcimrow.Descriptor,
	ipamDescriptors []ipamrow.Descriptor,
) []bootstrap.Entry {
	entries := make([]bootstrap.Entry, 0, len(dcimDescriptors)+len(ipamDescriptors))
	for _, descriptor := range dcimDescriptors {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
	}
	for _, descriptor := range ipamDescriptors {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
	}
	return entries
}

func ipamDescriptorNames(descriptors []ipamrow.Descriptor) []string {
	names := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		names = append(names, descriptor.Name)
	}
	return names
}

func ipamDescriptorDependencies(descriptors []ipamrow.Descriptor) map[string][]string {
	dependencies := make(map[string][]string, len(descriptors))
	for _, descriptor := range descriptors {
		dependencies[descriptor.Name] = descriptor.Dependencies
	}
	return dependencies
}

func openIPAMRowTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(
		sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared&_foreign_keys=on"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	return db
}
