package model

type CircuitsProviderAsns struct {
	ID         uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ProviderID int64  `gorm:"column:provider_id;type:int8;not null" json:"providerID"`
	AsnID      int64  `gorm:"column:asn_id;type:int8;not null" json:"asnID"`
}

// CircuitsProviderAsnsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var CircuitsProviderAsnsColumnNames = map[string]bool{
	"id":          true,
	"provider_id": true,
	"asn_id":      true,
}
