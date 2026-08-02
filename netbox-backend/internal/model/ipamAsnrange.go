package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamAsnrange struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug            string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Start           int64           `gorm:"column:start;type:int8;not null" json:"start"`
	End             int64           `gorm:"column:end;type:int8;not null" json:"end"`
	RirID           int64           `gorm:"column:rir_id;type:int8;not null" json:"rirID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
}

// TableName table name
func (m *IpamAsnrange) TableName() string {
	return "ipam_asnrange"
}

// IpamAsnrangeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamAsnrangeColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"name":              true,
	"slug":              true,
	"start":             true,
	"end":               true,
	"rir_id":            true,
	"tenant_id":         true,
}
