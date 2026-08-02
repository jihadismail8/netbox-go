package model

type AuthGroupPermissions struct {
	ID           uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	GroupID      int    `gorm:"column:group_id;type:int4;not null" json:"groupID"`
	PermissionID int    `gorm:"column:permission_id;type:int4;not null" json:"permissionID"`
}

// AuthGroupPermissionsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var AuthGroupPermissionsColumnNames = map[string]bool{
	"id":            true,
	"group_id":      true,
	"permission_id": true,
}
