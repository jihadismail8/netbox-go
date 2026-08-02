package model

import (
	"gorm.io/datatypes"
	"time"
)

type VpnIkeproposal struct {
	ID                      uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Created                 *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated             *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData         *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	Description             string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	Comments                string          `gorm:"column:comments;type:text;not null" json:"comments"`
	Name                    string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	AuthenticationMethod    string          `gorm:"column:authentication_method;type:varchar;not null" json:"authenticationMethod"`
	EncryptionAlgorithm     string          `gorm:"column:encryption_algorithm;type:varchar;not null" json:"encryptionAlgorithm"`
	AuthenticationAlgorithm string          `gorm:"column:authentication_algorithm;type:varchar" json:"authenticationAlgorithm"`
	Group                   int             `gorm:"column:group;type:int2;not null" json:"group"`
	SaLifetime              int             `gorm:"column:sa_lifetime;type:int4" json:"saLifetime"`
}

// TableName table name
func (m *VpnIkeproposal) TableName() string {
	return "vpn_ikeproposal"
}

// VpnIkeproposalColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIkeproposalColumnNames = map[string]bool{
	"id":                       true,
	"created":                  true,
	"last_updated":             true,
	"custom_field_data":        true,
	"description":              true,
	"comments":                 true,
	"name":                     true,
	"authentication_method":    true,
	"encryption_algorithm":     true,
	"authentication_algorithm": true,
	"group":                    true,
	"sa_lifetime":              true,
}
