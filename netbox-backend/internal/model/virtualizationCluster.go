package model

import (
	"gorm.io/datatypes"
	"time"
)

type VirtualizationCluster struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	GroupID         int64           `gorm:"column:group_id;type:int8" json:"groupID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	TypeID          int64           `gorm:"column:type_id;type:int8;not null" json:"typeID"`
	Status          string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	ScopeID         int64           `gorm:"column:scope_id;type:int8" json:"scopeID"`
	ScopeTypeID     int             `gorm:"column:scope_type_id;type:int4" json:"scopeTypeID"`
	XLocationID     int64           `gorm:"column:_location_id;type:int8" json:"XLocationID"`
	XRegionID       int64           `gorm:"column:_region_id;type:int8" json:"XRegionID"`
	XSiteID         int64           `gorm:"column:_site_id;type:int8" json:"XSiteID"`
	XSiteGroupID    int64           `gorm:"column:_site_group_id;type:int8" json:"XSiteGroupID"`
}

// TableName table name
func (m *VirtualizationCluster) TableName() string {
	return "virtualization_cluster"
}

// VirtualizationClusterColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VirtualizationClusterColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"comments":          true,
	"group_id":          true,
	"tenant_id":         true,
	"type_id":           true,
	"status":            true,
	"description":       true,
	"scope_id":          true,
	"scope_type_id":     true,
	"_location_id":      true,
	"_region_id":        true,
	"_site_id":          true,
	"_site_group_id":    true,
}
