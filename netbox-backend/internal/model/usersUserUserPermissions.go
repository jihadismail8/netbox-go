package model

type UsersUserUserPermissions struct {
	ID           uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	UserID       int64  `gorm:"column:user_id;type:int8;not null" json:"userID"`
	PermissionID int    `gorm:"column:permission_id;type:int4;not null" json:"permissionID"`
}

// UsersUserUserPermissionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersUserUserPermissionsColumnNames = map[string]bool{
	"id":            true,
	"user_id":       true,
	"permission_id": true,
}
