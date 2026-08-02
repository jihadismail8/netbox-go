package model

type DcimInterfaceWirelessLans struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	InterfaceID   int64  `gorm:"column:interface_id;type:int8;not null" json:"interfaceID"`
	WirelesslanID int64  `gorm:"column:wirelesslan_id;type:int8;not null" json:"wirelesslanID"`
}

// DcimInterfaceWirelessLansColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimInterfaceWirelessLansColumnNames = map[string]bool{
	"id":             true,
	"interface_id":   true,
	"wirelesslan_id": true,
}
