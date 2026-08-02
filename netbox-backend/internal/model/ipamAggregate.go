package model

import (
	"gorm.io/datatypes"
	"time"
)

type IpamAggregate struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Prefix          string          `gorm:"column:prefix;type:cidr;not null" json:"prefix"`
	DateAdded       *time.Time      `gorm:"column:date_added;type:date" json:"dateAdded"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	RirID           int64           `gorm:"column:rir_id;type:int8;not null" json:"rirID"`
	TenantID        int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
}

// TableName table name
func (m *IpamAggregate) TableName() string {
	return "ipam_aggregate"
}

// IpamAggregateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamAggregateColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"prefix":            true,
	"date_added":        true,
	"description":       true,
	"rir_id":            true,
	"tenant_id":         true,
	"comments":          true,
}
