package model

type VpnIpsecpolicyProposals struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	IpsecpolicyID   int64  `gorm:"column:ipsecpolicy_id;type:int8;not null" json:"ipsecpolicyID"`
	IpsecproposalID int64  `gorm:"column:ipsecproposal_id;type:int8;not null" json:"ipsecproposalID"`
}

// VpnIpsecpolicyProposalsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIpsecpolicyProposalsColumnNames = map[string]bool{
	"id":               true,
	"ipsecpolicy_id":   true,
	"ipsecproposal_id": true,
}
