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
