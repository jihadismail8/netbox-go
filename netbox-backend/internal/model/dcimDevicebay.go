package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimDevicebay struct {
	Created           *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData   *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID                uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name              string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label             string          `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description       string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	DeviceID          int64           `gorm:"column:device_id;type:int8;not null" json:"deviceID"`
	InstalledDeviceID int64           `gorm:"column:installed_device_id;type:int8" json:"installedDeviceID"`
	XLocationID       int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRackID           int64           `gorm:"column:_rack_id;type:int8" json:"XRackID"`
	XSiteID           int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
}

// TableName table name
func (m *DcimDevicebay) TableName() string {
	return "dcim_devicebay"
}

// DcimDevicebayColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimDevicebayColumnNames = map[string]bool{
	"created":             true,
	"last_updated":        true,
	"custom_field_data":   true,
	"id":                  true,
	"name":                true,
	"label":               true,
	"description":         true,
	"device_id":           true,
	"installed_device_id": true,
	"_location_id":        true,
	"_rack_id":            true,
	"_site_id":            true,
}
