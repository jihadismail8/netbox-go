package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnIpsecproposal struct {
	ID                      uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created                 *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated             *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData         *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description             string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments                string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name                    string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	EncryptionAlgorithm     string          `gorm:"column:encryption_algorithm;type:varchar" json:"encryptionAlgorithm"`
	AuthenticationAlgorithm string          `gorm:"column:authentication_algorithm;type:varchar" json:"authenticationAlgorithm"`
	SaLifetimeSeconds       int             `gorm:"column:sa_lifetime_seconds;type:int4" json:"saLifetimeSeconds"`
	SaLifetimeData          int             `gorm:"column:sa_lifetime_data;type:int4" json:"saLifetimeData"`
}

// TableName table name
func (m *VpnIpsecproposal) TableName() string {
	return "vpn_ipsecproposal"
}

// VpnIpsecproposalColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIpsecproposalColumnNames = map[string]bool{
	"id":                       true,
	"created":                  true,
	"last_updated":             true,
	"custom_field_data":        true,
	"description":              true,
	"comments":                 true,
	"name":                     true,
	"encryption_algorithm":     true,
	"authentication_algorithm": true,
	"sa_lifetime_seconds":      true,
	"sa_lifetime_data":         true,
}
