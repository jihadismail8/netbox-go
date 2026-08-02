package model

type UsersUserObjectPermissions struct {
	ID                 uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ObjectpermissionID int64  `gorm:"column:objectpermission_id;type:int8;not null" json:"objectpermissionID"`
	UserID             int64  `gorm:"column:user_id;type:int8;not null" json:"userID"`
}

// UsersUserObjectPermissionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersUserObjectPermissionsColumnNames = map[string]bool{
	"id":                  true,
	"objectpermission_id": true,
	"user_id":             true,
}
