package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type ExtrasConfigcontextprofile struct {
	ID              uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	DataPath        string          `gorm:"column:data_path;type:varchar(1000);not null" json:"dataPath"`
	AutoSyncEnabled *sgorm.Bool     `gorm:"column:auto_sync_enabled;type:bool;not null" json:"autoSyncEnabled"`
	DataSynced      *time.Time      `gorm:"column:data_synced;type:timestamptz" json:"dataSynced"`
	Comments        string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name            string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description     string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Schema          *datatypes.JSON `gorm:"column:schema;type:jsonb" json:"schema"`
	DataFileID      int64           `gorm:"column:data_file_id;type:int8" json:"dataFileID"`
	DataSourceID    int64           `gorm:"column:data_source_id;type:int8" json:"dataSourceID"`
}

// TableName table name
func (m *ExtrasConfigcontextprofile) TableName() string {
	return "extras_configcontextprofile"
}

// ExtrasConfigcontextprofileColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextprofileColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"data_path":         true,
	"auto_sync_enabled": true,
	"data_synced":       true,
	"comments":          true,
	"name":              true,
	"description":       true,
	"schema":            true,
	"data_file_id":      true,
	"data_source_id":    true,
}
