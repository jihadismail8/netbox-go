package dcim_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceNormalizesValuesAndPreservesNullVersusBlankDuplex(t *testing.T) {
	t.Parallel()

	networkInterface, err := dcim.NewInterface(dcim.InterfaceValues{
		Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
		Name:   " eth0 ", Label: " uplink ", Type: "1000base-t",
		Enabled: true, MgmtOnly: true,
		MTU:         dcim.NonNullDeviceValue(uint32(1500)),
		Speed:       dcim.NonNullDeviceValue(uint64(1_000_000)),
		Duplex:      dcim.NonNullDeviceValue(""),
		Description: " WAN uplink ",
	}, testTime)
	require.NoError(t, err)
	require.NoError(t, networkInterface.AssignID(9))

	assert.Equal(t, "eth0", networkInterface.Name())
	assert.Equal(t, "uplink", networkInterface.Label())
	assert.Equal(t, "eth0 (uplink)", networkInterface.Display())
	assert.Equal(t, dcim.InterfaceType("1000base-t"), networkInterface.Type())
	duplex, present := networkInterface.Duplex().Get()
	require.True(t, present)
	assert.Empty(t, duplex)
	require.NotNil(t, networkInterface.Snapshot().Duplex)
	assert.Empty(t, *networkInterface.Snapshot().Duplex)

	nullDuplex := dcim.NullDeviceValue[string]()
	require.NoError(t, networkInterface.ApplyPatch(
		dcim.InterfacePatch{Duplex: &nullDuplex},
		shared.NewTimestamp(testTime.Add(time.Minute)),
	))
	assert.True(t, networkInterface.Duplex().IsNull())
	assert.Nil(t, networkInterface.Snapshot().Duplex)
}

func TestInterfaceRejectsUnknownTypeAndNumericBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		values    dcim.InterfaceValues
		wantField string
	}{
		{
			name: "unknown_type",
			values: dcim.InterfaceValues{
				Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
				Name:   "eth0", Type: "not-an-interface",
			},
			wantField: "type",
		},
		{
			name: "mtu_too_small",
			values: dcim.InterfaceValues{
				Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
				Name:   "eth0", Type: "1000base-t",
				MTU: dcim.NonNullDeviceValue(uint32(0)),
			},
			wantField: "mtu",
		},
		{
			name: "mtu_too_large",
			values: dcim.InterfaceValues{
				Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
				Name:   "eth0", Type: "1000base-t",
				MTU: dcim.NonNullDeviceValue(uint32(65537)),
			},
			wantField: "mtu",
		},
		{
			name: "speed_too_large",
			values: dcim.InterfaceValues{
				Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
				Name:   "eth0", Type: "1000base-t",
				Speed: dcim.NonNullDeviceValue(uint64(2147483648)),
			},
			wantField: "speed",
		},
		{
			name: "unknown_duplex",
			values: dcim.InterfaceValues{
				Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
				Name:   "eth0", Type: "1000base-t",
				Duplex: dcim.NonNullDeviceValue("sideways"),
			},
			wantField: "duplex",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := dcim.NewInterface(test.values, testTime)
			require.Error(t, err)
			assert.Contains(t, violationFields(shared.ViolationsOf(err)), test.wantField)
		})
	}
}

func TestInterfaceCannotMoveToAnotherDevice(t *testing.T) {
	t.Parallel()

	networkInterface, err := dcim.NewInterface(dcim.InterfaceValues{
		Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
		Name:   "eth0", Type: "1000base-t", Enabled: true,
	}, testTime)
	require.NoError(t, err)
	require.NoError(t, networkInterface.AssignID(9))

	other := interfaceDeviceReference(t, 8, "edge-02", "edge-02")
	err = networkInterface.ApplyPatch(
		dcim.InterfacePatch{Device: &other},
		shared.NewTimestamp(testTime.Add(time.Minute)),
	)
	require.Error(t, err)
	assert.Equal(t, "device", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, shared.ID(7), networkInterface.Device().ID())
	assert.Equal(t, testTime, networkInterface.LastUpdated())
}

func TestInterfacePatchIsAtomicAndPreservesCounter(t *testing.T) {
	t.Parallel()

	networkInterface, err := dcim.RestoreInterface(dcim.InterfaceState{
		ID: 9, Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
		Name: "eth0", Type: "1000base-t", Enabled: true,
		Created: testTime, LastUpdated: testTime, IPAddressCount: 3,
	})
	require.NoError(t, err)

	label := "changed"
	invalidType := "invalid"
	err = networkInterface.ApplyPatch(dcim.InterfacePatch{
		Label: &label, Type: &invalidType,
	}, shared.NewTimestamp(testTime.Add(time.Minute)))
	require.Error(t, err)
	assert.Empty(t, networkInterface.Label())
	assert.Equal(t, dcim.InterfaceType("1000base-t"), networkInterface.Type())
	assert.Equal(t, uint64(3), networkInterface.IPAddressCount())
	assert.Equal(t, testTime, networkInterface.LastUpdated())
}

func TestInterfaceEnforcesTextLengths(t *testing.T) {
	t.Parallel()

	values := dcim.InterfaceValues{
		Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
		Name:   strings.Repeat("x", dcim.InterfaceNameMaxLength+1),
		Type:   "1000base-t",
	}
	_, err := dcim.NewInterface(values, testTime)
	require.Error(t, err)
	assert.Contains(t, violationFields(shared.ViolationsOf(err)), "name")
}

func TestInterfaceScalarNormalizationContract(t *testing.T) {
	t.Parallel()
	values := func(t *testing.T) dcim.InterfaceValues {
		t.Helper()
		return dcim.InterfaceValues{
			Device: interfaceDeviceReference(t, 7, "edge-01", "edge-01"),
			Name:   "  Ethernet1  ", Label: "  uplink  ", Type: "bridge",
			Enabled: false, MgmtOnly: false, Description: "  edge uplink  ",
		}
	}
	t.Run("trim_only_text_fields", func(t *testing.T) {
		networkInterface, err := dcim.NewInterface(values(t), testTime)
		require.NoError(t, err)
		assert.Equal(t, "Ethernet1", networkInterface.Name())
		assert.Equal(t, "uplink", networkInterface.Label())
		assert.Equal(t, "edge uplink", networkInterface.Description())
		assert.False(t, networkInterface.Enabled())
		assert.False(t, networkInterface.MgmtOnly())
	})
	t.Run("speed_zero_is_a_concrete_value", func(t *testing.T) {
		input := values(t)
		input.Speed = dcim.NonNullDeviceValue(uint64(0))
		networkInterface, err := dcim.NewInterface(input, testTime)
		require.NoError(t, err)
		speed, present := networkInterface.Speed().Get()
		require.True(t, present)
		assert.Zero(t, speed)
		require.NotNil(t, networkInterface.Snapshot().Speed)
		assert.Zero(t, *networkInterface.Snapshot().Speed)
	})
	for _, test := range []struct {
		name    string
		change  func(*dcim.InterfaceValues)
		field   string
		message string
	}{
		{"type_is_not_trimmed", func(input *dcim.InterfaceValues) { input.Type = " bridge " }, "type", "Select a valid choice."},
		{"duplex_is_not_trimmed", func(input *dcim.InterfaceValues) { input.Duplex = dcim.NonNullDeviceValue(" full ") }, "duplex", "Select a valid choice."},
		{"blank_type", func(input *dcim.InterfaceValues) { input.Type = "" }, "type", "This field may not be blank."},
		{"mtu_minimum", func(input *dcim.InterfaceValues) { input.MTU = dcim.NonNullDeviceValue(uint32(0)) }, "mtu", "Ensure this value is greater than or equal to 1."},
		{"mtu_maximum", func(input *dcim.InterfaceValues) { input.MTU = dcim.NonNullDeviceValue(uint32(65537)) }, "mtu", "Ensure this value is less than or equal to 65536."},
		{"speed_maximum", func(input *dcim.InterfaceValues) { input.Speed = dcim.NonNullDeviceValue(uint64(2147483648)) }, "speed", "Ensure this value is less than or equal to 2147483647."},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := values(t)
			test.change(&input)
			_, err := dcim.NewInterface(input, testTime)
			require.Error(t, err)
			violations := shared.ViolationsOf(err)
			require.Len(t, violations, 1)
			assert.Equal(t, test.field, violations[0].Field)
			assert.Equal(t, test.message, violations[0].Description)
		})
	}
	t.Run("character_limits", func(t *testing.T) {
		input := values(t)
		input.Name = strings.Repeat("界", dcim.InterfaceNameMaxLength)
		input.Label = strings.Repeat("界", dcim.InterfaceLabelMaxLength)
		input.Description = strings.Repeat("界", dcim.InterfaceDescriptionMaxLength)
		_, err := dcim.NewInterface(input, testTime)
		require.NoError(t, err)
		input.Name += "界"
		input.Label += "界"
		input.Description += "界"
		_, err = dcim.NewInterface(input, testTime)
		require.Error(t, err)
		assert.Equal(t, []string{"name", "label", "description"}, violationFields(shared.ViolationsOf(err)))
	})
	t.Run("duplex_blank_and_null_remain_distinct", func(t *testing.T) {
		input := values(t)
		input.Duplex = dcim.NonNullDeviceValue("")
		networkInterface, err := dcim.NewInterface(input, testTime)
		require.NoError(t, err)
		require.NotNil(t, networkInterface.Snapshot().Duplex)
		assert.Empty(t, *networkInterface.Snapshot().Duplex)
		nullDuplex := dcim.NullDeviceValue[string]()
		require.NoError(t, networkInterface.ApplyPatch(dcim.InterfacePatch{Duplex: &nullDuplex}, shared.NewTimestamp(testTime.Add(time.Minute))))
		assert.True(t, networkInterface.Duplex().IsNull())
		assert.Nil(t, networkInterface.Snapshot().Duplex)
	})
	t.Run("signed_numeric_parsers", func(t *testing.T) {
		for _, test := range []struct {
			name    string
			parse   func(int64) (uint64, error)
			minimum int64
			maximum int64
			inputs  []int64
		}{
			{"mtu", func(value int64) (uint64, error) {
				parsed, err := dcim.ParseInterfaceMTU(value)
				return uint64(parsed), err
			}, 1, 65536, []int64{math.MinInt64, -1, 0, 1, 65536, 65537, math.MaxInt64}},
			{"speed", dcim.ParseInterfaceSpeed, 0, 2147483647, []int64{math.MinInt64, -1, 0, 2147483647, 2147483648, math.MaxInt64}},
		} {
			for _, value := range test.inputs {
				parsed, err := test.parse(value)
				if value >= test.minimum && value <= test.maximum {
					require.NoError(t, err)
					assert.Equal(t, uint64(value), parsed)
					continue
				}
				require.Error(t, err)
				assert.Zero(t, parsed, "invalid signed values must never wrap")
				violations := shared.ViolationsOf(err)
				require.Len(t, violations, 1)
				assert.Equal(t, test.name, violations[0].Field)
				assert.Equal(t, "out_of_range", violations[0].Reason)
				if value < test.minimum {
					assert.Contains(t, violations[0].Description, "greater than or equal")
				} else {
					assert.Contains(t, violations[0].Description, "less than or equal")
				}
			}
		}
	})
	t.Run("preview_and_restore_preserve_valid_zero_state", func(t *testing.T) {
		input := values(t)
		input.Speed = dcim.NonNullDeviceValue(uint64(0))
		networkInterface, err := dcim.NewInterface(input, testTime)
		require.NoError(t, err)
		require.NoError(t, networkInterface.AssignID(42))
		before := networkInterface.State()
		beforeSnapshot := networkInterface.Snapshot()
		restored, err := dcim.RestoreInterface(before)
		require.NoError(t, err)
		assert.Equal(t, before, restored.State())
		assert.Equal(t, beforeSnapshot, restored.Snapshot())
		label, invalidType := "changed", " bridge "
		err = networkInterface.ValidatePatch(dcim.InterfacePatch{Label: &label, Type: &invalidType})
		require.Error(t, err)
		assert.Equal(t, before, networkInterface.State())
		assert.Equal(t, beforeSnapshot, networkInterface.Snapshot())
		require.NoError(t, networkInterface.ValidatePatch(dcim.InterfacePatch{}))
		require.Error(t, networkInterface.ApplyPatch(dcim.InterfacePatch{}, shared.NewTimestamp(testTime.Add(time.Minute))))
		assert.Equal(t, before, networkInterface.State())
		var absent *dcim.Interface
		assert.True(t, shared.HasReason(absent.ValidatePatch(dcim.InterfacePatch{}), shared.ErrorReasonInternal))
	})
}

func interfaceDeviceReference(
	t *testing.T,
	id shared.ID,
	name string,
	display string,
) dcim.DeviceReference {
	t.Helper()
	reference, err := dcim.NewDeviceReference(
		id,
		dcim.NonNullDeviceValue(name),
		display,
	)
	require.NoError(t, err)
	return reference
}
