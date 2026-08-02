package model

type ExtrasCustomfieldObjectTypes struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	CustomfieldID int64  `gorm:"column:customfield_id;type:int8;not null" json:"customfieldID"`
	ContenttypeID int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// ExtrasCustomfieldObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasCustomfieldObjectTypesColumnNames = map[string]bool{
	"id":             true,
	"customfield_id": true,
	"contenttype_id": true,
}
