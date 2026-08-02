package dcim_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceTemplatePinsChoicesDisplayAndSnapshot(t *testing.T) {
	reference := interfaceTemplateDeviceTypeReference(t, 7, "Router", "router")
	now := shared.NewTimestamp(time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC))
	template, err := dcim.NewInterfaceTemplate(dcim.InterfaceTemplateValues{
		DeviceType: reference, Name: " Ethernet1 ", Label: " Uplink ",
		Type: "1000base-t", Enabled: true, MgmtOnly: true,
		Description: " Template-only description ",
	}, now)
	require.NoError(t, err)

	assert.Equal(t, "Ethernet1 (Uplink)", template.Display())
	assert.Equal(t, dcim.InterfaceType("1000base-t"), template.Type())
	assert.Equal(t, dcim.InterfaceTemplateSnapshot{
		DeviceTypeID: 7, Name: "Ethernet1", Label: "Uplink", Type: "1000base-t",
		Enabled: true, MgmtOnly: true, Description: "Template-only description",
	}, template.Snapshot())

	_, err = dcim.NewInterfaceTemplate(dcim.InterfaceTemplateValues{
		DeviceType: reference, Name: "Ethernet2", Type: "not-a-netbox-interface",
	}, now)
	require.Error(t, err)
	assert.Equal(t, "type", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "invalid_choice", shared.ViolationsOf(err)[0].Reason)
}

func TestInterfaceTemplateCannotMoveBetweenDeviceTypes(t *testing.T) {
	now := shared.NewTimestamp(time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC))
	later := shared.NewTimestamp(time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC))
	original := interfaceTemplateDeviceTypeReference(t, 7, "Router", "router")
	other := interfaceTemplateDeviceTypeReference(t, 8, "Switch", "switch")
	template, err := dcim.NewInterfaceTemplate(dcim.InterfaceTemplateValues{
		DeviceType: original, Name: "Ethernet1", Type: "1000base-t",
	}, now)
	require.NoError(t, err)

	err = template.ApplyPatch(dcim.InterfaceTemplatePatch{DeviceType: &other}, later)
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "device_type", Reason: "immutable",
		Description: "An InterfaceTemplate cannot be moved to another DeviceType.",
	}}, shared.ViolationsOf(err))
	assert.Equal(t, original.ID(), template.DeviceType().ID())
	assert.Equal(t, now, template.LastUpdated())
}

func interfaceTemplateDeviceTypeReference(
	t *testing.T,
	id shared.ID,
	model string,
	slug string,
) dcim.DeviceTypeReference {
	t.Helper()
	reference, err := dcim.NewDeviceTypeReference(id, model, slug)
	require.NoError(t, err)
	return reference
}
