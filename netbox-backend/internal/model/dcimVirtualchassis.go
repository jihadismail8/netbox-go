package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimVirtualchassis struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Domain          string          `gorm:"column:domain;type:varchar(30);not null" json:"domain"`
	MasterID        int64           `gorm:"column:master_id;type:int8" json:"masterID"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	MemberCount     int64           `gorm:"column:member_count;type:int8;not null" json:"memberCount"`
}

// TableName table name
func (m *DcimVirtualchassis) TableName() string {
	return "dcim_virtualchassis"
}

// DcimVirtualchassisColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimVirtualchassisColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"domain":            true,
	"master_id":         true,
	"comments":          true,
	"description":       true,
	"member_count":      true,
}
