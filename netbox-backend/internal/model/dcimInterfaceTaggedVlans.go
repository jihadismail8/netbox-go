package model

type DcimInterfaceTaggedVlans struct {
	ID          uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	InterfaceID int64  `gorm:"column:interface_id;type:int8;not null" json:"interfaceID"`
	VlanID      int64  `gorm:"column:vlan_id;type:int8;not null" json:"vlanID"`
}

// DcimInterfaceTaggedVlansColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimInterfaceTaggedVlansColumnNames = map[string]bool{
	"id":           true,
	"interface_id": true,
	"vlan_id":      true,
}
