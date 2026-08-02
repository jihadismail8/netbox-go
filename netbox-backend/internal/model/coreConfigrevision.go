package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type CoreConfigrevision struct {
	ID      uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created *time.Time      `gorm:"column:created;type:timestamptz;not null" json:"created"`
	Comment string          `gorm:"column:comment;type:varchar(200);not null" json:"comment"`
	Data    *datatypes.JSON `gorm:"column:data;type:jsonb" json:"data"`
	Active  *sgorm.Bool     `gorm:"column:active;type:bool;not null" json:"active"`
}

// TableName table name
func (m *CoreConfigrevision) TableName() string {
	return "core_configrevision"
}

// CoreConfigrevisionColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreConfigrevisionColumnNames = map[string]bool{
	"id":      true,
	"created": true,
	"comment": true,
	"data":    true,
	"active":  true,
}
