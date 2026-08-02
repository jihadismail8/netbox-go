package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"time"
)

type ExtrasTableconfig struct {
	ID           uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created      *time.Time  `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated  *time.Time  `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	Table        string      `gorm:"column:table;type:varchar(100);not null" json:"table"`
	Name         string      `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description  string      `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Weight       int         `gorm:"column:weight;type:int2;not null" json:"weight"`
	Enabled      *sgorm.Bool `gorm:"column:enabled;type:bool;not null" json:"enabled"`
	Shared       *sgorm.Bool `gorm:"column:shared;type:bool;not null" json:"shared"`
	Columns      string      `gorm:"column:columns;type:_varchar;not null" json:"columns"`
	Ordering     string      `gorm:"column:ordering;type:_varchar" json:"ordering"`
	ObjectTypeID int         `gorm:"column:object_type_id;type:int4;not null" json:"objectTypeID"`
	UserID       int64       `gorm:"column:user_id;type:int8" json:"userID"`
}

// TableName table name
func (m *ExtrasTableconfig) TableName() string {
	return "extras_tableconfig"
}

// ExtrasTableconfigColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasTableconfigColumnNames = map[string]bool{
	"id":             true,
	"created":        true,
	"last_updated":   true,
	"table":          true,
	"name":           true,
	"description":    true,
	"weight":         true,
	"enabled":        true,
	"shared":         true,
	"columns":        true,
	"ordering":       true,
	"object_type_id": true,
	"user_id":        true,
}
