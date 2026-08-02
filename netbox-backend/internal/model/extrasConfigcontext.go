package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type ExtrasConfigcontext struct {
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Weight          int             `gorm:"column:weight;type:int2;not null" json:"weight"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	IsActive        *sgorm.Bool     `gorm:"column:is_active;type:bool;not null" json:"isActive"`
	Data            *datatypes.JSON `gorm:"column:data;type:jsonb;not null" json:"data"`
	DataFileID      int64           `gorm:"column:data_file_id;type:int8" json:"dataFileID"`
	DataPath        string          `gorm:"column:data_path;type:varchar(1000);not null" json:"dataPath"`
	DataSourceID    int64           `gorm:"column:data_source_id;type:int8" json:"dataSourceID"`
	AutoSyncEnabled *sgorm.Bool     `gorm:"column:auto_sync_enabled;type:bool;not null" json:"autoSyncEnabled"`
	DataSynced      *time.Time      `gorm:"column:data_synced;type:timestamptz" json:"dataSynced"`
	ProfileID       int64           `gorm:"column:profile_id;type:int8" json:"profileID"`
}

// TableName table name
func (m *ExtrasConfigcontext) TableName() string {
	return "extras_configcontext"
}

// ExtrasConfigcontextColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"id":                true,
	"name":              true,
	"weight":            true,
	"description":       true,
	"is_active":         true,
	"data":              true,
	"data_file_id":      true,
	"data_path":         true,
	"data_source_id":    true,
	"auto_sync_enabled": true,
	"data_synced":       true,
	"profile_id":        true,
}
