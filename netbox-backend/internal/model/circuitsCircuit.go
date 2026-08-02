package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"time"
)

type CircuitsCircuit struct {
	Created           *time.Time       `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated       *time.Time       `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData   *datatypes.JSON  `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID                uint64           `gorm:"column:id;type:int8;primary_key" json:"id"`
	Cid               string           `gorm:"column:cid;type:varchar(100);not null" json:"cid"`
	Status            string           `gorm:"column:status;type:varchar(50);not null" json:"status"`
	InstallDate       *time.Time       `gorm:"column:install_date;type:date" json:"installDate"`
	CommitRate        int              `gorm:"column:commit_rate;type:int4" json:"commitRate"`
	Description       string           `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments          string           `gorm:"column:comments;type:text;not null" json:"comments"`
	ProviderID        int64            `gorm:"column:provider_id;type:int8;not null" json:"providerID"`
	TenantID          int64            `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	TerminationAID    int64            `gorm:"column:termination_a_id;type:int8" json:"terminationAID"`
	TerminationZID    int64            `gorm:"column:termination_z_id;type:int8" json:"terminationZID"`
	TypeID            int64            `gorm:"column:type_id;type:int8;not null" json:"typeID"`
	TerminationDate   *time.Time       `gorm:"column:termination_date;type:date" json:"terminationDate"`
	ProviderAccountID int64            `gorm:"column:provider_account_id;type:int8" json:"providerAccountID"`
	XAbsDistance      *decimal.Decimal `gorm:"column:_abs_distance;type:numeric" json:"XAbsDistance"`
	Distance          *decimal.Decimal `gorm:"column:distance;type:numeric" json:"distance"`
	DistanceUnit      string           `gorm:"column:distance_unit;type:varchar(50)" json:"distanceUnit"`
}

// TableName table name
func (m *CircuitsCircuit) TableName() string {
	return "circuits_circuit"
}

// CircuitsCircuitColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsCircuitColumnNames = map[string]bool{
	"created":             true,
	"last_updated":        true,
	"custom_field_data":   true,
	"id":                  true,
	"cid":                 true,
	"status":              true,
	"install_date":        true,
	"commit_rate":         true,
	"description":         true,
	"comments":            true,
	"provider_id":         true,
	"tenant_id":           true,
	"termination_a_id":    true,
	"termination_z_id":    true,
	"type_id":             true,
	"termination_date":    true,
	"provider_account_id": true,
	"_abs_distance":       true,
	"distance":            true,
	"distance_unit":       true,
}
