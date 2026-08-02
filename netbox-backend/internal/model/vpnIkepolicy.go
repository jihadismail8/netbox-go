package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnIkepolicy struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Version         int             `gorm:"column:version;type:int2;not null" json:"version"`
	Mode            string          `gorm:"column:mode;type:varchar" json:"mode"`
	PresharedKey    string          `gorm:"column:preshared_key;type:text;not null" json:"presharedKey"`
}

// TableName table name
func (m *VpnIkepolicy) TableName() string {
	return "vpn_ikepolicy"
}

// VpnIkepolicyColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIkepolicyColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"comments":          true,
	"name":              true,
	"version":           true,
	"mode":              true,
	"preshared_key":     true,
}
