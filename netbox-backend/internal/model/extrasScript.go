package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

type ExtrasScript struct {
	ID           uint64      `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name         string      `gorm:"column:name;type:varchar(79);not null" json:"name"`
	ModuleID     int64       `gorm:"column:module_id;type:int8;not null" json:"moduleID"`
	IsExecutable *sgorm.Bool `gorm:"column:is_executable;type:bool;not null" json:"isExecutable"`
}

// TableName table name
func (m *ExtrasScript) TableName() string {
	return "extras_script"
}

// ExtrasScriptColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasScriptColumnNames = map[string]bool{
	"id":            true,
	"name":          true,
	"module_id":     true,
	"is_executable": true,
}
