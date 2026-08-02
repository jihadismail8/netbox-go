package model

import (
	"gorm.io/datatypes"
)

type ExtrasDashboard struct {
	ID     uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Layout *datatypes.JSON `gorm:"column:layout;type:jsonb;not null" json:"layout"`
	Config *datatypes.JSON `gorm:"column:config;type:jsonb;not null" json:"config"`
	UserID int64           `gorm:"column:user_id;type:int8;not null" json:"userID"`
}

// TableName table name
func (m *ExtrasDashboard) TableName() string {
	return "extras_dashboard"
}

// ExtrasDashboardColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasDashboardColumnNames = map[string]bool{
	"id":      true,
	"layout":  true,
	"config":  true,
	"user_id": true,
}
