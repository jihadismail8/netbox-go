package model

type UsersGroupObjectPermissions struct {
	ID                 uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ObjectpermissionID int64  `gorm:"column:objectpermission_id;type:int8;not null" json:"objectpermissionID"`
	GroupID            int64  `gorm:"column:group_id;type:int8;not null" json:"groupID"`
}

// UsersGroupObjectPermissionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersGroupObjectPermissionsColumnNames = map[string]bool{
	"id":                  true,
	"objectpermission_id": true,
	"group_id":            true,
}
