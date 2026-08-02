package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
)

type DcimCablepath struct {
	ID         uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	XNodes     string          `gorm:"column:_nodes;type:_varchar;not null" json:"XNodes"`
	IsActive   *sgorm.Bool     `gorm:"column:is_active;type:bool;not null" json:"isActive"`
	IsSplit    *sgorm.Bool     `gorm:"column:is_split;type:bool;not null" json:"isSplit"`
	Path       *datatypes.JSON `gorm:"column:path;type:jsonb;not null" json:"path"`
	IsComplete *sgorm.Bool     `gorm:"column:is_complete;type:bool;not null" json:"isComplete"`
}

// TableName table name
func (m *DcimCablepath) TableName() string {
	return "dcim_cablepath"
}

// DcimCablepathColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimCablepathColumnNames = map[string]bool{
	"id":          true,
	"_nodes":      true,
	"is_active":   true,
	"is_split":    true,
	"path":        true,
	"is_complete": true,
}
