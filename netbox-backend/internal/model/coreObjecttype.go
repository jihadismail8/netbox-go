package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
)

type CoreObjecttype struct {
	ContenttypePtrID int         `gorm:"column:contenttype_ptr_id;type:int4;primary_key" json:"contenttypePtrID"`
	Public           *sgorm.Bool `gorm:"column:public;type:bool;not null" json:"public"`
	Features         string      `gorm:"column:features;type:_varchar;not null" json:"features"`
}

// TableName table name
func (m *CoreObjecttype) TableName() string {
	return "core_objecttype"
}

// CoreObjecttypeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CoreObjecttypeColumnNames = map[string]bool{
	"contenttype_ptr_id": true,
	"public":             true,
	"features":           true,
}
