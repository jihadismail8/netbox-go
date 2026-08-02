package model

type ExtrasConfigcontextDeviceTypes struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	DevicetypeID    int64  `gorm:"column:devicetype_id;type:int8;not null" json:"devicetypeID"`
}

// ExtrasConfigcontextDeviceTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextDeviceTypesColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"devicetype_id":    true,
}
