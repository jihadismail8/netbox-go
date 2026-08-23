package ipam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestIPAddressScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	defaults, err := (CreateIPAddressCommand{
		Address: FieldValue("192.0.2.17"),
	}).values()
	require.NoError(t, err)
	assert.Equal(t, domainipam.IPAddressStatusActive.String(), defaults.status)
	assert.True(t, defaults.role.IsNull())
	assert.Empty(t, defaults.dnsName)
	assert.Empty(t, defaults.description)
	assert.Empty(t, defaults.comments)

	_, err = (CreateIPAddressCommand{}).values()
	assertIPAddressCommandViolation(
		t, err, "address", "This field is required.",
	)
	_, err = (ReplaceIPAddressCommand{ID: 1}).patch()
	assertIPAddressCommandViolation(
		t, err, "address", "This field is required.",
	)

	_, err = (CreateIPAddressCommand{
		Address: FieldValue(""), Status: FieldValue(""),
	}).values()
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 2)
	assert.Equal(t, shared.FieldViolation{
		Field: "address", Reason: "blank",
		Description: "This field may not be blank.",
	}, violations[0])
	assert.Equal(t, shared.FieldViolation{
		Field: "status", Reason: "blank",
		Description: "This field may not be blank.",
	}, violations[1])

	for _, test := range []struct {
		name  string
		field Field[string]
	}{
		{name: "null", field: NullField[string]()},
		{name: "blank", field: FieldValue("")},
	} {
		test := test
		t.Run("create status "+test.name, func(t *testing.T) {
			_, err := (CreateIPAddressCommand{
				Address: FieldValue("192.0.2.17"), Status: test.field,
			}).values()
			assertIPAddressCommandViolation(
				t, err, "status", "This field may not be blank.",
			)
		})
		t.Run("patch status "+test.name, func(t *testing.T) {
			_, err := (UpdateIPAddressCommand{
				ID: 1, Status: test.field,
			}).patch()
			assertIPAddressCommandViolation(
				t, err, "status", "This field may not be blank.",
			)
		})
	}

	for _, test := range []struct {
		name  string
		field Field[string]
	}{
		{name: "null", field: NullField[string]()},
		{name: "blank", field: FieldValue("")},
	} {
		test := test
		t.Run("role "+test.name, func(t *testing.T) {
			values, err := (CreateIPAddressCommand{
				Address: FieldValue("192.0.2.17"), Role: test.field,
			}).values()
			require.NoError(t, err)
			role, present := values.role.Get()
			require.True(t, present)
			assert.Equal(t, domainipam.IPAddressRole(""), role)
		})
	}

	for _, test := range []struct {
		name    string
		command CreateIPAddressCommand
		field   string
	}{
		{name: "address", field: "address", command: CreateIPAddressCommand{
			Address: NullField[string](),
		}},
		{name: "dns_name", field: "dns_name", command: CreateIPAddressCommand{
			Address: FieldValue("192.0.2.17"), DNSName: NullField[string](),
		}},
		{name: "description", field: "description", command: CreateIPAddressCommand{
			Address: FieldValue("192.0.2.17"), Description: NullField[string](),
		}},
		{name: "comments", field: "comments", command: CreateIPAddressCommand{
			Address: FieldValue("192.0.2.17"), Comments: NullField[string](),
		}},
	} {
		test := test
		t.Run(test.name+" null", func(t *testing.T) {
			_, err := test.command.values()
			assertIPAddressCommandViolation(
				t, err, test.field, "This field may not be null.",
			)
		})
	}

	replacePatch, err := (ReplaceIPAddressCommand{
		ID: 1,
		CreateIPAddressCommand: CreateIPAddressCommand{
			Address: FieldValue("198.51.100.8"),
		},
	}).patch()
	require.NoError(t, err)
	require.NotNil(t, replacePatch.address)
	assert.Nil(t, replacePatch.status)
	assert.Nil(t, replacePatch.role)
	assert.Nil(t, replacePatch.dnsName)
	assert.Nil(t, replacePatch.description)
	assert.Nil(t, replacePatch.comments)
	assert.False(t, replacePatch.vrfSet)
	assert.False(t, replacePatch.assignmentSet)

	rolePatch, err := (UpdateIPAddressCommand{
		ID: 1, Role: NullField[string](),
	}).patch()
	require.NoError(t, err)
	require.NotNil(t, rolePatch.role)
	role, present := rolePatch.role.Get()
	require.True(t, present)
	assert.Equal(t, domainipam.IPAddressRole(""), role)

	criteria, err := validateListIPAddressesQuery(ListIPAddressesQuery{
		Statuses: []string{" active "},
	})
	require.NoError(t, err, "the scalar-write slice must not change filter parsing")
	assert.Equal(t, []domainipam.IPAddressStatus{
		domainipam.IPAddressStatusActive,
	}, criteria.Statuses)
}

func assertIPAddressCommandViolation(
	t *testing.T,
	err error,
	field string,
	description string,
) {
	t.Helper()
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 1)
	assert.Equal(t, field, violations[0].Field)
	assert.Equal(t, description, violations[0].Description)
}
