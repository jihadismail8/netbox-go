package model

type ExtrasEventruleObjectTypes struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	EventruleID   int64  `gorm:"column:eventrule_id;type:int8;not null" json:"eventruleID"`
	ContenttypeID int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// ExtrasEventruleObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasEventruleObjectTypesColumnNames = map[string]bool{
	"id":             true,
	"eventrule_id":   true,
	"contenttype_id": true,
}
