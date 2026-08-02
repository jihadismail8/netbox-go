package model

import (
	"gorm.io/datatypes"
	"time"
)

type CircuitsVirtualcircuittype struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug            string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Color           string          `gorm:"column:color;type:varchar(6);not null" json:"color"`
}

// TableName table name
func (m *CircuitsVirtualcircuittype) TableName() string {
	return "circuits_virtualcircuittype"
}

// CircuitsVirtualcircuittypeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsVirtualcircuittypeColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"name":              true,
	"slug":              true,
	"description":       true,
	"color":             true,
}
