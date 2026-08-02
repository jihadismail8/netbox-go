package model

type IpamVrfExportTargets struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	VrfID         int64  `gorm:"column:vrf_id;type:int8;not null" json:"vrfID"`
	RoutetargetID int64  `gorm:"column:routetarget_id;type:int8;not null" json:"routetargetID"`
}

// IpamVrfExportTargetsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamVrfExportTargetsColumnNames = map[string]bool{
	"id":             true,
	"vrf_id":         true,
	"routetarget_id": true,
}
