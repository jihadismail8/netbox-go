package dcim_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	unpositionedValues := func() dcim.DeviceValues {
		values := validDeviceValues(t)
		values.Rack = dcim.NullDeviceValue[dcim.RackReference]()
		values.Position = dcim.NullDeviceValue[dcim.RackPosition]()
		values.Face = ""
		return values
	}

	t.Run("trims text while preserving nullable blank values", func(t *testing.T) {
		values := unpositionedValues()
		values.Name = dcim.NonNullDeviceValue("  edge-02  ")
		values.Serial = "  SN-02  "
		values.AssetTag = dcim.NonNullDeviceValue("   ")
		values.Description = "  description  "
		values.Comments = "  comments  "

		device, err := dcim.NewDevice(values, testTime)
		require.NoError(t, err)
		name, present := device.Name().Get()
		require.True(t, present)
		assert.Equal(t, "edge-02", name)
		assert.Equal(t, "SN-02", device.Serial())
		assetTag, present := device.AssetTag().Get()
		require.True(t, present)
		assert.Empty(t, assetTag)
		assert.Equal(t, "description", device.Description())
		assert.Equal(t, "comments", device.Comments())
	})

	for _, test := range []struct {
		name   string
		field  string
		mutate func(*dcim.DeviceValues)
	}{
		{name: "status", field: "status", mutate: func(values *dcim.DeviceValues) {
			values.Status = " active "
		}},
		{name: "face", field: "face", mutate: func(values *dcim.DeviceValues) {
			values.Face = " front "
		}},
		{name: "airflow", field: "airflow", mutate: func(values *dcim.DeviceValues) {
			values.Airflow = dcim.NonNullDeviceAirflow(" front-to-rear ")
		}},
	} {
		test := test
		t.Run("choice whitespace rejected/"+test.name, func(t *testing.T) {
			values := unpositionedValues()
			test.mutate(&values)
			_, err := dcim.NewDevice(values, testTime)
			require.Error(t, err)
			require.Contains(t, shared.ViolationsOf(err), shared.FieldViolation{
				Field: test.field, Reason: "invalid_choice", Description: "Select a valid choice.",
			})
		})
	}

	for _, test := range []struct {
		name   string
		field  string
		limit  int
		mutate func(*dcim.DeviceValues, string)
	}{
		{name: "name", field: "name", limit: dcim.DeviceNameMaxLength, mutate: func(values *dcim.DeviceValues, value string) {
			values.Name = dcim.NonNullDeviceValue(value)
		}},
		{name: "serial", field: "serial", limit: dcim.DeviceSerialMaxLength, mutate: func(values *dcim.DeviceValues, value string) {
			values.Serial = value
		}},
		{name: "asset tag", field: "asset_tag", limit: dcim.DeviceAssetTagMaxLength, mutate: func(values *dcim.DeviceValues, value string) {
			values.AssetTag = dcim.NonNullDeviceValue(value)
		}},
		{name: "description", field: "description", limit: dcim.DeviceDescriptionMaxLength, mutate: func(values *dcim.DeviceValues, value string) {
			values.Description = value
		}},
	} {
		test := test
		t.Run("rune length enforced/"+test.name, func(t *testing.T) {
			values := unpositionedValues()
			test.mutate(&values, strings.Repeat("é", test.limit+1))
			_, err := dcim.NewDevice(values, testTime)
			require.Error(t, err)
			require.Contains(t, shared.ViolationsOf(err), shared.FieldViolation{
				Field: test.field, Reason: "max_length",
				Description: "Ensure this field has no more than the supported number of characters.",
			})
		})
	}

	for _, test := range []struct {
		name    string
		request dcim.NullableDeviceAirflow
		want    dcim.DeviceAirflow
	}{
		{name: "null inherits", request: dcim.NullDeviceAirflow(), want: dcim.DeviceAirflowFrontToRear},
		{name: "blank inherits", request: dcim.NonNullDeviceAirflow(""), want: dcim.DeviceAirflowFrontToRear},
		{name: "concrete wins", request: dcim.NonNullDeviceAirflow(dcim.DeviceAirflowRearToFront), want: dcim.DeviceAirflowRearToFront},
	} {
		test := test
		t.Run(test.name+" DeviceType airflow", func(t *testing.T) {
			values := unpositionedValues()
			values.Airflow = test.request
			device, err := dcim.NewDevice(values, testTime)
			require.NoError(t, err)
			airflow, present := device.Airflow().Get()
			require.True(t, present)
			assert.Equal(t, test.want, airflow)
		})
	}

	t.Run("patch preview is atomic and blank airflow does not re-inherit", func(t *testing.T) {
		device, err := dcim.NewDevice(unpositionedValues(), testTime)
		require.NoError(t, err)
		before := device.State()
		blank := dcim.NonNullDeviceAirflow("")
		patch := dcim.DevicePatch{Airflow: &blank}
		require.NoError(t, device.ValidatePatch(patch))
		assert.Equal(t, before, device.State())
		require.NoError(t, device.ApplyPatch(
			patch, shared.NewTimestamp(testTime.Add(time.Minute)),
		))
		airflow, present := device.Airflow().Get()
		require.True(t, present)
		assert.Empty(t, airflow)
	})
}

func TestRackPositionUsesExactHalfUnits(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input     string
		halfUnits uint16
		rendered  string
	}{
		{input: "1", halfUnits: 2, rendered: "1"},
		{input: "1.0", halfUnits: 2, rendered: "1"},
		{input: "1.5", halfUnits: 3, rendered: "1.5"},
		{input: "1.50", halfUnits: 3, rendered: "1.5"},
		{input: "100.5", halfUnits: 201, rendered: "100.5"},
	} {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			position, err := dcim.ParseRackPosition(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.halfUnits, position.HalfUnits())
			assert.Equal(t, test.rendered, position.String())
		})
	}

	for _, invalid := range []string{"", "0.5", "1.1", "100.6", "101", "-1", "NaN"} {
		invalid := invalid
		t.Run("invalid_"+invalid, func(t *testing.T) {
			t.Parallel()
			_, err := dcim.ParseRackPosition(invalid)
			require.Error(t, err)
			require.NotEmpty(t, shared.ViolationsOf(err))
			assert.Equal(t, "position", shared.ViolationsOf(err)[0].Field)
		})
	}
}

func TestDeviceCreationInheritsAirflowAndPreservesNullableFields(t *testing.T) {
	t.Parallel()

	device, err := dcim.NewDevice(dcim.DeviceValues{
		DeviceType: deviceTypeInstanceReference(
			t, 10, "Router 9000", "router-9000", "Acme", "1.5", true,
			dcim.NonNullDeviceAirflow(dcim.DeviceAirflowFrontToRear),
		),
		Role: deviceRoleReference(11, "Core Router"),
		Name: dcim.NonNullDeviceValue(" edge-01 "),
		Site: siteReference(t, 12, "Primary", "primary"),
		Rack: dcim.NonNullDeviceValue(
			rackReference(t, 13, "Rack A", 12, 1, 42),
		),
		Position: dcim.NonNullDeviceValue(mustRackPosition(t, "10.5")),
		Face:     "front", Status: "active", Serial: " SN-1 ",
		AssetTag:    dcim.NonNullDeviceValue(" ASSET-1 "),
		Airflow:     dcim.NullDeviceAirflow(),
		Description: " Edge router ", Comments: " Notes ",
	}, testTime)
	require.NoError(t, err)
	require.NoError(t, device.AssignID(20))

	name, hasName := device.Name().Get()
	require.True(t, hasName)
	assert.Equal(t, "edge-01", name)
	assert.Equal(t, "edge-01 (ASSET-1)", device.Display())
	assert.Equal(t, "SN-1", device.Serial())
	assert.Equal(t, "Edge router", device.Description())
	airflow, hasAirflow := device.Airflow().Get()
	require.True(t, hasAirflow)
	assert.Equal(t, dcim.DeviceAirflowFrontToRear, airflow)

	snapshot := device.Snapshot()
	require.NotNil(t, snapshot.Name)
	assert.Equal(t, "edge-01", *snapshot.Name)
	require.NotNil(t, snapshot.RackID)
	assert.Equal(t, shared.ID(13), *snapshot.RackID)
	require.NotNil(t, snapshot.Position)
	assert.Equal(t, "10.5", *snapshot.Position)
	require.NotNil(t, snapshot.AssetTag)
	assert.Equal(t, "ASSET-1", *snapshot.AssetTag)
}

func TestDeviceCreationDistinguishesNullAndBlankNameAndAssetTag(t *testing.T) {
	t.Parallel()

	values := validDeviceValues(t)
	values.Name = dcim.NullDeviceValue[string]()
	values.AssetTag = dcim.NonNullDeviceValue("")
	device, err := dcim.NewDevice(values, testTime)
	require.NoError(t, err)
	require.NoError(t, device.AssignID(44))

	assert.True(t, device.Name().IsNull())
	assetTag, present := device.AssetTag().Get()
	require.True(t, present)
	assert.Empty(t, assetTag)
	assert.Nil(t, device.Snapshot().Name)
	require.NotNil(t, device.Snapshot().AssetTag)
	assert.Equal(t, "Acme Router 9000 (44)", device.Display())
}

func TestDeviceValidatesRackRelationshipsAndPlacement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mutate    func(*dcim.DeviceValues)
		wantField string
	}{
		{
			name: "face_without_rack",
			mutate: func(values *dcim.DeviceValues) {
				values.Rack = dcim.NullDeviceValue[dcim.RackReference]()
			},
			wantField: "face",
		},
		{
			name: "position_without_face",
			mutate: func(values *dcim.DeviceValues) {
				values.Face = ""
			},
			wantField: "face",
		},
		{
			name: "rack_in_another_site",
			mutate: func(values *dcim.DeviceValues) {
				values.Rack = dcim.NonNullDeviceValue(
					rackReference(t, 13, "Rack A", 999, 1, 42),
				)
			},
			wantField: "rack",
		},
		{
			name: "outside_rack_bounds",
			mutate: func(values *dcim.DeviceValues) {
				values.Position = dcim.NonNullDeviceValue(mustRackPosition(t, "42"))
				values.DeviceType = deviceTypeInstanceReference(
					t, 10, "Router 9000", "router-9000", "Acme", "1.5", true,
					dcim.NullDeviceAirflow(),
				)
			},
			wantField: "position",
		},
		{
			name: "zero_u_positioned",
			mutate: func(values *dcim.DeviceValues) {
				values.DeviceType = deviceTypeInstanceReference(
					t, 10, "PDU", "pdu", "Acme", "0", false,
					dcim.NullDeviceAirflow(),
				)
			},
			wantField: "position",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			values := validDeviceValues(t)
			test.mutate(&values)
			_, err := dcim.NewDevice(values, testTime)
			require.Error(t, err)
			violations := shared.ViolationsOf(err)
			require.NotEmpty(t, violations)
			assert.Contains(t, violationFields(violations), test.wantField)
		})
	}
}

func TestDevicePatchIsAtomicAndDoesNotReinheritAirflow(t *testing.T) {
	t.Parallel()

	values := validDeviceValues(t)
	device, err := dcim.NewDevice(values, testTime)
	require.NoError(t, err)
	require.NoError(t, device.AssignID(55))

	name := dcim.NonNullDeviceValue("changed")
	badStatus := "unsupported"
	err = device.ApplyPatch(dcim.DevicePatch{
		Name: &name, Status: &badStatus,
	}, shared.NewTimestamp(testTime.Add(time.Minute)))
	require.Error(t, err)
	currentName, present := device.Name().Get()
	require.True(t, present)
	assert.Equal(t, "edge-01", currentName)
	assert.Equal(t, dcim.DeviceStatusActive, device.Status())
	assert.Equal(t, testTime, device.LastUpdated())

	blank := dcim.NonNullDeviceAirflow("")
	require.NoError(t, device.ApplyPatch(
		dcim.DevicePatch{Airflow: &blank},
		shared.NewTimestamp(testTime.Add(2*time.Minute)),
	))
	airflow, present := device.Airflow().Get()
	require.True(t, present)
	assert.Empty(t, airflow)
}

func validDeviceValues(t *testing.T) dcim.DeviceValues {
	t.Helper()
	return dcim.DeviceValues{
		DeviceType: deviceTypeInstanceReference(
			t, 10, "Router 9000", "router-9000", "Acme", "1", true,
			dcim.NonNullDeviceAirflow(dcim.DeviceAirflowFrontToRear),
		),
		Role: deviceRoleReference(11, "Core Router"),
		Name: dcim.NonNullDeviceValue("edge-01"),
		Site: siteReference(t, 12, "Primary", "primary"),
		Rack: dcim.NonNullDeviceValue(
			rackReference(t, 13, "Rack A", 12, 1, 42),
		),
		Position: dcim.NonNullDeviceValue(mustRackPosition(t, "10")),
		Face:     "front", Status: "active",
		Airflow: dcim.NullDeviceAirflow(),
	}
}

func deviceTypeInstanceReference(
	t *testing.T,
	id shared.ID,
	model string,
	slug string,
	manufacturer string,
	height string,
	fullDepth bool,
	airflow dcim.NullableDeviceAirflow,
) dcim.DeviceTypeInstanceReference {
	t.Helper()
	parsedHeight, err := dcim.ParseDeviceHeight(height)
	require.NoError(t, err)
	reference, err := dcim.NewDeviceTypeInstanceReference(
		id, model, slug, manufacturer, parsedHeight, fullDepth, airflow,
	)
	require.NoError(t, err)
	return reference
}

func deviceRoleReference(id shared.ID, display string) dcim.DeviceRoleReference {
	return dcim.DeviceRoleReference{ID: id, Display: display}
}

func siteReference(
	t *testing.T,
	id shared.ID,
	name string,
	slug string,
) dcim.SiteReference {
	t.Helper()
	reference, err := dcim.NewSiteReference(id, name, slug)
	require.NoError(t, err)
	return reference
}

func rackReference(
	t *testing.T,
	id shared.ID,
	display string,
	siteID shared.ID,
	startingUnit uint32,
	uHeight uint32,
) dcim.RackReference {
	t.Helper()
	reference, err := dcim.NewRackReference(id, display, siteID, startingUnit, uHeight)
	require.NoError(t, err)
	return reference
}

func mustRackPosition(t *testing.T, value string) dcim.RackPosition {
	t.Helper()
	position, err := dcim.ParseRackPosition(value)
	require.NoError(t, err)
	return position
}

func violationFields(violations []shared.FieldViolation) []string {
	fields := make([]string, 0, len(violations))
	for _, violation := range violations {
		fields = append(fields, violation.Field)
	}
	return fields
}
