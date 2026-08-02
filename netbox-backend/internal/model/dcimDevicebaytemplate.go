package model

import (
	"time"
)

type DcimDevicebaytemplate struct {
	Created      *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated  *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID           uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name         string     `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Label        string     `gorm:"column:label;type:varchar(64);not null" json:"label"`
	Description  string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	DeviceTypeID int64      `gorm:"column:device_type_id;type:int8;not null" json:"deviceTypeID"`
}

// TableName table name
func (m *DcimDevicebaytemplate) TableName() string {
	return "dcim_devicebaytemplate"
}

// DcimDevicebaytemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimDevicebaytemplateColumnNames = map[string]bool{
	"created":        true,
	"last_updated":   true,
	"id":             true,
	"name":           true,
	"label":          true,
	"description":    true,
	"device_type_id": true,
}
