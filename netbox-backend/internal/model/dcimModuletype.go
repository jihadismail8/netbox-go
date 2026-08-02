package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"time"
)

type DcimModuletype struct {
	ID              uint64           `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created         *time.Time       `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time       `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON  `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Model           string           `gorm:"column:model;type:varchar(100);not null" json:"model"`
	PartNumber      string           `gorm:"column:part_number;type:varchar(50);not null" json:"partNumber"`
	Comments        string           `gorm:"column:comments;type:text;not null" json:"comments"`
	ManufacturerID  int64            `gorm:"column:manufacturer_id;type:int8;not null" json:"manufacturerID"`
	Weight          *decimal.Decimal `gorm:"column:weight;type:numeric" json:"weight"`
	WeightUnit      string           `gorm:"column:weight_unit;type:varchar(50)" json:"weightUnit"`
	XAbsWeight      int64            `gorm:"column:_abs_weight;type:int8" json:"XAbsWeight"`
	Description     string           `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Airflow         string           `gorm:"column:airflow;type:varchar(50)" json:"airflow"`
	AttributeData   *datatypes.JSON  `gorm:"column:attribute_data;type:jsonb" json:"attributeData"`
	ProfileID       int64            `gorm:"column:profile_id;type:int8" json:"profileID"`
}

// TableName table name
func (m *DcimModuletype) TableName() string {
	return "dcim_moduletype"
}

// DcimModuletypeColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimModuletypeColumnNames = map[string]bool{
	"id":                true,
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"model":             true,
	"part_number":       true,
	"comments":          true,
	"manufacturer_id":   true,
	"weight":            true,
	"weight_unit":       true,
	"_abs_weight":       true,
	"description":       true,
	"airflow":           true,
	"attribute_data":    true,
	"profile_id":        true,
}
