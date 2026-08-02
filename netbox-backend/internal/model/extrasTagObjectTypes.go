package model

type ExtrasTagObjectTypes struct {
	ID            uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	TagID         int64  `gorm:"column:tag_id;type:int8;not null" json:"tagID"`
	ContenttypeID int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// ExtrasTagObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasTagObjectTypesColumnNames = map[string]bool{
	"id":             true,
	"tag_id":         true,
	"contenttype_id": true,
}
