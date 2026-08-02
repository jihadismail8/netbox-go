package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamRoutetarget struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(21);not null" json:"name"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
}

// TableName table name
func (m *IpamRoutetarget) TableName() string {
	return "ipam_routetarget"
}

// IpamRoutetargetColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamRoutetargetColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"description":       true,
	"tenant_id":         true,
	"comments":          true,
}
