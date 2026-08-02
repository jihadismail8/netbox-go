package model

import (
	"gorm.io/datatypes"
	"time"
)

type TenancyTenant struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug            string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	GroupID         int64           `gorm:"column:group_id;type:int8" json:"groupID"`
}

// TableName table name
func (m *TenancyTenant) TableName() string {
	return "tenancy_tenant"
}

// TenancyTenantColumnNames Whitelist for custom query fields to prevent sql injection attacks
var TenancyTenantColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"name":              true,
	"slug":              true,
	"description":       true,
	"comments":          true,
	"group_id":          true,
}
