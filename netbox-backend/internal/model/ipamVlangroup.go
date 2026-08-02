package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamVlangroup struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug            string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	ScopeID         int64           `gorm:"column:scope_id;type:int8" json:"scopeID"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	ScopeTypeID     int             `gorm:"column:scope_type_id;type:int4" json:"scopeTypeID"`
	VidRanges       string          `gorm:"column:vid_ranges;type:_int4range;not null" json:"vidRanges"`
	XTotalVlanIds   int64           `gorm:"column:_total_vlan_ids;type:int8;not null" json:"XTotalVlanIds"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
}

// TableName table name
func (m *IpamVlangroup) TableName() string {
	return "ipam_vlangroup"
}

// IpamVlangroupColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamVlangroupColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"slug":              true,
	"scope_id":          true,
	"description":       true,
	"scope_type_id":     true,
	"vid_ranges":        true,
	"_total_vlan_ids":   true,
	"tenant_id":         true,
}
