package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
)

type UsersObjectpermission struct {
	ID          uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name        string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Description string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Enabled     *sgorm.Bool     `gorm:"column:enabled;type:bool;not null" json:"enabled"`
	Actions     string          `gorm:"column:actions;type:_varchar;not null" json:"actions"`
	Constraints *datatypes.JSON `gorm:"column:constraints;type:jsonb" json:"constraints"`
}

// TableName table name
func (m *UsersObjectpermission) TableName() string {
	return "users_objectpermission"
}

// UsersObjectpermissionColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersObjectpermissionColumnNames = map[string]bool{
	"id":          true,
	"name":        true,
	"description": true,
	"enabled":     true,
	"actions":     true,
	"constraints": true,
}
