package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"time"
)

type WirelessWirelesslink struct {
	ID                  uint64           `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created             *time.Time       `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated         *time.Time       `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData     *datatypes.JSON  `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Ssid                string           `gorm:"column:ssid;type:varchar(32);not null" json:"ssid"`
	Status              string           `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Description         string           `gorm:"column:description;type:varchar(200);not null" json:"description"`
	AuthCipher          string           `gorm:"column:auth_cipher;type:varchar(50)" json:"authCipher"`
	AuthPsk             string           `gorm:"column:auth_psk;type:varchar(64);not null" json:"authPsk"`
	AuthType            string           `gorm:"column:auth_type;type:varchar(50)" json:"authType"`
	XInterfaceADeviceID int64            `gorm:"column:_interface_a_device_id;type:int8" json:"XInterfaceADeviceID"`
	XInterfaceBDeviceID int64            `gorm:"column:_interface_b_device_id;type:int8" json:"XInterfaceBDeviceID"`
	InterfaceAID        int64            `gorm:"column:interface_a_id;type:int8;not null" json:"interfaceAID"`
	InterfaceBID        int64            `gorm:"column:interface_b_id;type:int8;not null" json:"interfaceBID"`
	TenantID            int64            `gorm:"column:tenant_id;type:int8" json:"tenantID"`
	Comments            string           `gorm:"column:comments;type:text;not null" json:"comments"`
	XAbsDistance        *decimal.Decimal `gorm:"column:_abs_distance;type:numeric" json:"XAbsDistance"`
	Distance            *decimal.Decimal `gorm:"column:distance;type:numeric" json:"distance"`
	DistanceUnit        string           `gorm:"column:distance_unit;type:varchar(50)" json:"distanceUnit"`
}

// TableName table name
func (m *WirelessWirelesslink) TableName() string {
	return "wireless_wirelesslink"
}

// WirelessWirelesslinkColumnNames Whitelist for custom query fields to prevent sql injection attacks
var WirelessWirelesslinkColumnNames = map[string]bool{
	"id":                     true,
	"created":                true,
	"last_updated":           true,
	"custom_field_data":      true,
	"ssid":                   true,
	"status":                 true,
	"description":            true,
	"auth_cipher":            true,
	"auth_psk":               true,
	"auth_type":              true,
	"_interface_a_device_id": true,
	"_interface_b_device_id": true,
	"interface_a_id":         true,
	"interface_b_id":         true,
	"tenant_id":              true,
	"comments":               true,
	"_abs_distance":          true,
	"distance":               true,
	"distance_unit":          true,
}
