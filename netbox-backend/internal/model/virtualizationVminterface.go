package model

import (
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"gorm.io/datatypes"
	"time"
)

type VirtualizationVminterface struct {
	Created                 *time.Time      `gorm:"column:created;type:timestamptz" json:"created"`
	LastUpdated             *time.Time      `gorm:"column:last_updated;type:timestamptz" json:"lastUpdated"`
	CustomFieldData         *datatypes.JSON `gorm:"column:custom_field_data;type:jsonb;not null" json:"customFieldData"`
	ID                      uint64          `gorm:"column:id;type:int8;primary_key" json:"id"`
	Enabled                 *sgorm.Bool     `gorm:"column:enabled;type:bool;not null" json:"enabled"`
	Mtu                     int             `gorm:"column:mtu;type:int4" json:"mtu"`
	Mode                    string          `gorm:"column:mode;type:varchar(50)" json:"mode"`
	Name                    string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	XName                   string          `gorm:"column:_name;type:varchar(100);not null" json:"XName"`
	Description             string          `gorm:"column:description;type:varchar(200);not null" json:"description"`
	ParentID                int64           `gorm:"column:parent_id;type:int8" json:"parentID"`
	UntaggedVlanID          int64           `gorm:"column:untagged_vlan_id;type:int8" json:"untaggedVlanID"`
	VirtualMachineID        int64           `gorm:"column:virtual_machine_id;type:int8;not null" json:"virtualMachineID"`
	BridgeID                int64           `gorm:"column:bridge_id;type:int8" json:"bridgeID"`
	VrfID                   int64           `gorm:"column:vrf_id;type:int8" json:"vrfID"`
	VlanTranslationPolicyID int64           `gorm:"column:vlan_translation_policy_id;type:int8" json:"vlanTranslationPolicyID"`
	QinqSvlanID             int64           `gorm:"column:qinq_svlan_id;type:int8" json:"qinqSvlanID"`
	PrimaryMacAddressID     int64           `gorm:"column:primary_mac_address_id;type:int8" json:"primaryMacAddressID"`
}

// TableName table name
func (m *VirtualizationVminterface) TableName() string {
	return "virtualization_vminterface"
}

// VirtualizationVminterfaceColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VirtualizationVminterfaceColumnNames = map[string]bool{
	"created":                    true,
	"last_updated":               true,
	"custom_field_data":          true,
	"id":                         true,
	"enabled":                    true,
	"mtu":                        true,
	"mode":                       true,
	"name":                       true,
	"_name":                      true,
	"description":                true,
	"parent_id":                  true,
	"untagged_vlan_id":           true,
	"virtual_machine_id":         true,
	"bridge_id":                  true,
	"vrf_id":                     true,
	"vlan_translation_policy_id": true,
	"qinq_svlan_id":              true,
	"primary_mac_address_id":     true,
}
