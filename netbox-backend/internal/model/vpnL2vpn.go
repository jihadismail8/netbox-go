package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnL2Vpn struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug            string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Type            string          `gorm:"column:type;type:varchar(50);not null" json:"type"`
	Identifier      int64           `gorm:"column:identifier;type:int8" json:"identifier"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
}

// TableName table name
func (m *VpnL2Vpn) TableName() string {
	return "vpn_l2vpn"
}

// VpnL2VpnColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnL2VpnColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"name":              true,
	"slug":              true,
	"type":              true,
	"identifier":        true,
	"description":       true,
	"tenant_id":         true,
	"comments":          true,
	"status":            true,
}
