package model

type VpnL2VpnImportTargets struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	L2VpnID       int64  `gorm:"column:l2vpn_id;type:int8;not null" json:"l2vpnID"`
	RoutetargetID int64  `gorm:"column:routetarget_id;type:int8;not null" json:"routetargetID"`
}

// VpnL2VpnImportTargetsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnL2VpnImportTargetsColumnNames = map[string]bool{
	"id":             true,
	"l2vpn_id":       true,
	"routetarget_id": true,
}
