package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"time"
)

type DcimCable struct {
	Created         *time.Time       `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated     *time.Time       `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData *datatypes.JSON  `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID              uint64           `gorm:"column:id;type:int8;primary_key" json:"id"`
	Type            string           `gorm:"column:type;type:varchar(50)" json:"type"`
	Status          string           `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Label           string           `gorm:"column:label;type:varchar(100);not null" json:"label"`
	Color           string           `gorm:"column:color;type:varchar(6);not null" json:"color"`
	Length          *decimal.Decimal `gorm:"column:length;type:numeric" json:"length"`
	LengthUnit      string           `gorm:"column:length_unit;type:varchar(50)" json:"lengthUnit"`
	XAbsLength      *decimal.Decimal `gorm:"column:_abs_length;type:numeric" json:"XAbsLength"`
	TenantID        int64            `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments        string           `gorm:"column:comments;type:text;not null" json:"comments"`
	Description     string           `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *DcimCable) TableName() string {
	return "dcim_cable"
}

// DcimCableColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimCableColumnNames = map[string]bool{
	"created":           true,
	"last_updated":      true,
	"custom_field_data": true,
	"id":                true,
	"type":              true,
	"status":            true,
	"label":             true,
	"color":             true,
	"length":            true,
	"length_unit":       true,
	"_abs_length":       true,
	"tenant_id":         true,
	"comments":          true,
	"description":       true,
}
