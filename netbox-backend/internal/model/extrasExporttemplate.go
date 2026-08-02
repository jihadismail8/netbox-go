package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type ExtrasExporttemplate struct {
	ID                uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name              string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description       string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	TemplateCode      string          `gorm:"column:template_code;type:text;not null" json:"templateCode"`
	MimeType          string          `gorm:"column:mime_type;type:varchar(50);not null" json:"mimeType"`
	FileExtension     string          `gorm:"column:file_extension;type:varchar(15);not null" json:"fileExtension"`
	AsAttachment      *sgorm.Bool     `gorm:"column:as_attachment;type:bool;not null" json:"asAttachment"`
	Created           *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	DataFileID        int64           `gorm:"column:data_file_id;type:int8" json:"dataFileID"`
	DataPath          string          `gorm:"column:data_path;type:varchar(1000);not null" json:"dataPath"`
	DataSourceID      int64           `gorm:"column:data_source_id;type:int8" json:"dataSourceID"`
	AutoSyncEnabled   *sgorm.Bool     `gorm:"column:auto_sync_enabled;type:bool;not null" json:"autoSyncEnabled"`
	DataSynced        *time.Time      `gorm:"column:data_synced;type:timestamptz" json:"dataSynced"`
	FileName          string          `gorm:"column:file_name;type:varchar(200);not null" json:"fileName"`
	EnvironmentParams *datatypes.JSON `gorm:"column:environment_params;type:jsonb" json:"environmentParams"`
}

// TableName table name
func (m *ExtrasExporttemplate) TableName() string {
	return "extras_exporttemplate"
}

// ExtrasExporttemplateColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasExporttemplateColumnNames = map[string]bool{
	"id":                 true,
	"name":               true,
	"description":        true,
	"template_code":      true,
	"mime_type":          true,
	"file_extension":     true,
	"as_attachment":      true,
	"created":            true,
	"last_updated":       true,
	"data_file_id":       true,
	"data_path":          true,
	"data_source_id":     true,
	"auto_sync_enabled":  true,
	"data_synced":        true,
	"file_name":          true,
	"environment_params": true,
}
