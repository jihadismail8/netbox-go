package model

type ExtrasConfigcontextRoles struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	DeviceroleID    int64  `gorm:"column:devicerole_id;type:int8;not null" json:"deviceroleID"`
}

// ExtrasConfigcontextRolesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextRolesColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"devicerole_id":    true,
}
