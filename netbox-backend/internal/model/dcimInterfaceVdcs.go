package model

type DcimInterfaceVdcs struct {
	ID                     uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	InterfaceID            int64  `gorm:"column:interface_id;type:int8;not null" json:"interfaceID"`
	VirtualdevicecontextID int64  `gorm:"column:virtualdevicecontext_id;type:int8;not null" json:"virtualdevicecontextID"`
}

// DcimInterfaceVdcsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var DcimInterfaceVdcsColumnNames = map[string]bool{
	"id":                      true,
	"interface_id":            true,
	"virtualdevicecontext_id": true,
}
