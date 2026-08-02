package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnIpsecpolicy struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	PfsGroup        int             `gorm:"column:pfs_group;type:int2" json:"pfsGroup"`
}

// TableName table name
func (m *VpnIpsecpolicy) TableName() string {
	return "vpn_ipsecpolicy"
}

// VpnIpsecpolicyColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIpsecpolicyColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"description":       true,
	"comments":          true,
	"name":              true,
	"pfs_group":         true,
}
