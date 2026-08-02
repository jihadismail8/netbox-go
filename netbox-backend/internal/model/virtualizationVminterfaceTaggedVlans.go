package model

type VirtualizationVminterfaceTaggedVlans struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	VminterfaceID int64  `gorm:"column:vminterface_id;type:int8;not null" json:"vminterfaceID"`
	VlanID        int64  `gorm:"column:vlan_id;type:int8;not null" json:"vlanID"`
}

// VirtualizationVminterfaceTaggedVlansColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VirtualizationVminterfaceTaggedVlansColumnNames = map[string]bool{
	"id":             true,
	"vminterface_id": true,
	"vlan_id":        true,
}
