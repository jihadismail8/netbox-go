package model

import (
	"time"
)

type ExtrasTag struct {
	Name        string     `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug        string     `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Created     *time.Time `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated *time.Time `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID          uint64     `gorm:"column:id;type:int8;primary_key" json:"id"`
	Color       string     `gorm:"column:color;type:varchar(6);not null" json:"color"`
	Description string     `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Weight      int        `gorm:"column:weight;type:int2;not null" json:"weight"`
}

// TableName table name
func (m *ExtrasTag) TableName() string {
	return "extras_tag"
}

// ExtrasTagColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasTagColumnNames = map[string]bool{
	"name":         true,
	"slug":         true,
	"created":      true,
	"last_updated": true,
	"id":           true,
	"color":        true,
	"description":  true,
	"weight":       true,
}
