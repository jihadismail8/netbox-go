package model

type UsersGroupPermissions struct {
	ID           uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	GroupID      int64  `gorm:"column:group_id;type:int8;not null" json:"groupID"`
	PermissionID int    `gorm:"column:permission_id;type:int4;not null" json:"permissionID"`
}

// UsersGroupPermissionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersGroupPermissionsColumnNames = map[string]bool{
	"id":            true,
	"group_id":      true,
	"permission_id": true,
}
