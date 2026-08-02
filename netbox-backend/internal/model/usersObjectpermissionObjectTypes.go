package model

type UsersObjectpermissionObjectTypes struct {
	ID                 uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ObjectpermissionID int64  `gorm:"column:objectpermission_id;type:int8;not null" json:"objectpermissionID"`
	ContenttypeID      int    `gorm:"column:contenttype_id;type:int4;not null" json:"contenttypeID"`
}

// UsersObjectpermissionObjectTypesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersObjectpermissionObjectTypesColumnNames = map[string]bool{
	"id":                  true,
	"objectpermission_id": true,
	"contenttype_id":      true,
}
