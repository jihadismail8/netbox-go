package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimVirtualdevicecontext struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Identifier      int             `gorm:"column:identifier;type:int2" json:"identifier"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	DeviceID        int64           `gorm:"column:device_id;type:int8" json:"deviceID"`
	PrimaryIp4ID    int64           `gorm:"column:primary_ip4_id;type:int8" json:"primaryIp4ID"`
	PrimaryIp6ID    int64           `gorm:"column:primary_ip6_id;type:int8" json:"primaryIp6ID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
}

// TableName table name
func (m *DcimVirtualdevicecontext) TableName() string {
	return "dcim_virtualdevicecontext"
}

// DcimVirtualdevicecontextColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimVirtualdevicecontextColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"name":              true,
	"status":            true,
	"identifier":        true,
	"comments":          true,
	"device_id":         true,
	"primary_ip4_id":    true,
	"primary_ip6_id":    true,
	"tenant_id":         true,
}
