package model

type VpnIkepolicyProposals struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	IkepolicyID   int64  `gorm:"column:ikepolicy_id;type:int8;not null" json:"ikepolicyID"`
	IkeproposalID int64  `gorm:"column:ikeproposal_id;type:int8;not null" json:"ikeproposalID"`
}

// VpnIkepolicyProposalsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var VpnIkepolicyProposalsColumnNames = map[string]bool{
	"id":             true,
	"ikepolicy_id":   true,
	"ikeproposal_id": true,
}
