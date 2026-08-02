package model

import (
	"gorm.io/datatypes"
	"time"
)

type DcimPlatform struct {
	Created          *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated      *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData  *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID               uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name             string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Slug             string          `gorm:"column:slug;type:varchar(100);not null" json:"slug"`
	Description      string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	ManufacturerID   int64           `gorm:"column:manufacturer_id;type:int8" json:"manufacturerID"`
	ConfigTemplateID int64           `gorm:"column:config_template_id;type:int8" json:"configTemplateID"`
	ParentID         int64           `gorm:"column:parent_id;type:int8" json:"parentID"`
	Level            int             `gorm:"column:level;type:int4;not null" json:"level"`
	Lft              int             `gorm:"column:lft;type:int4;not null" json:"lft"`
	Rght             int             `gorm:"column:rght;type:int4;not null" json:"rght"`
	TreeID           int             `gorm:"column:tree_id;type:int4;not null" json:"treeID"`
	Comments         string          `gorm:"column:comments;type:text;not null" json:"comments"`
}

// TableName table name
func (m *DcimPlatform) TableName() string {
	return "dcim_platform"
}

// DcimPlatformColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimPlatformColumnNames = map[string]bool{
	"created":            true,
	"last_updated":       true,
	"custom_field_data":  true,
	"id":                 true,
	"name":               true,
	"slug":               true,
	"description":        true,
	"manufacturer_id":    true,
	"config_template_id": true,
	"parent_id":          true,
	"level":              true,
	"lft":                true,
	"rght":               true,
	"tree_id":            true,
	"comments":           true,
}
