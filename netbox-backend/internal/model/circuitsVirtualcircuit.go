package model

import (
	"gorm.io/datatypes"
	"time"
)

type CircuitsVirtualcircuit struct {
	ID                uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created           *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData   *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description       string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments          string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Cid               string          `gorm:"column:cid;type:varchar(100);not null" json:"cid"`
	Status            string          `gorm:"column:status;type:varchar(50);not null" json:"status"`
	ProviderAccountID int64           `gorm:"column:provider_account_id;type:int8" json:"providerAccountID"`
	ProviderNetworkID int64           `gorm:"column:provider_network_id;type:int8;not null" json:"providerNetworkID"`
	TypeID            int64           `gorm:"column:type_id;type:int8;not null" json:"typeID"`
	TenantID          int64           `gorm:"column:tenant_id;type:int8" json:"tenantID"`
}

// TableName table name
func (m *CircuitsVirtualcircuit) TableName() string {
	return "circuits_virtualcircuit"
}

// CircuitsVirtualcircuitColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsVirtualcircuitColumnNames = map[string]bool{
	"id":                  true,
	"created":             true,
	"last_updated":        true,
	"custom_field_data":   true,
	"description":         true,
	"comments":            true,
	"cid":                 true,
	"status":              true,
	"provider_account_id": true,
	"provider_network_id": true,
	"type_id":             true,
	"tenant_id":           true,
}
