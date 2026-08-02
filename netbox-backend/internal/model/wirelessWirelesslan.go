package model

import (
	"gorm.io/datatypes"
	"time"
)

type WirelessWirelesslan struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Ssid            string          `gorm:"column:ssid;type:varchar(32);not null" json:"ssid"`
	GroupID         int64           `gorm:"column:group_id;type:int8" json:"groupID"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	AuthCipher      string          `gorm:"column:auth_cipher;type:varchar(50)" json:"authCipher"`
	AuthPsk         string          `gorm:"column:auth_psk;type:varchar(64);not null" json:"authPsk"`
	AuthType        string          `gorm:"column:auth_type;type:varchar(50)" json:"authType"`
	VlanID          int64           `gorm:"column:vlan_id;type:int8" json:"vlanID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	XLocationID     int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRegionID       int64           `gorm:"column:_region_id;type:int8" json:"XRegionID"`
	XSiteID         int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
	XSiteGroupID    int64           `gorm:"column:_site_group_id;type:int8" json:"XSiteGroupID"`
	ScopeID         int64           `gorm:"column:scope_id;type:int8" json:"scopeID"`
	ScopeTypeID     int             `gorm:"column:scope_type_id;type:int4" json:"scopeTypeID"`
}

// TableName table name
func (m *WirelessWirelesslan) TableName() string {
	return "wireless_wirelesslan"
}

// WirelessWirelesslanColumnNames Whitelist for custom query fields to prevent sql injection attacks
var WirelessWirelesslanColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"ssid":              true,
	"group_id":          true,
	"description":       true,
	"auth_cipher":       true,
	"auth_psk":          true,
	"auth_type":         true,
	"vlan_id":           true,
	"tenant_id":         true,
	"comments":          true,
	"status":            true,
	"_location_id":      true,
	"_region_id":        true,
	"_site_id":          true,
	"_site_group_id":    true,
	"scope_id":          true,
	"scope_type_id":     true,
}
