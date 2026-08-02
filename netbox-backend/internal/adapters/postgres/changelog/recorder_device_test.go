package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestMarshalDeviceSnapshotUsesPinnedChangePayloadShape(t *testing.T) {
	name, position, assetTag, airflow := "edge-1", "10.5", "asset-1", "front-to-rear"
	rackID := shared.ID(13)
	encoded, err := marshalSnapshot(domaindcim.DeviceSnapshot{
		DeviceTypeID: 7,
		RoleID:       9,
		Name:         &name,
		SiteID:       11,
		RackID:       &rackID,
		Position:     &position,
		Face:         "front",
		Status:       "active",
		Serial:       "serial",
		AssetTag:     &assetTag,
		Airflow:      &airflow,
		Description:  "description",
		Comments:     "comments",
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"device_type": 7,
		"role": 9,
		"name": "edge-1",
		"site": 11,
		"rack": 13,
		"position": "10.5",
		"face": "front",
		"status": "active",
		"serial": "serial",
		"asset_tag": "asset-1",
		"airflow": "front-to-rear",
		"description": "description",
		"comments": "comments"
	}`, string(encoded))
}

func TestMarshalNilDeviceSnapshotPointerRemainsEmpty(t *testing.T) {
	var snapshot *domaindcim.DeviceSnapshot
	encoded, err := marshalSnapshot(snapshot)
	require.NoError(t, err)
	assert.Nil(t, encoded)
}
