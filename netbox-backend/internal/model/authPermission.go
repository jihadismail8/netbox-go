package model

type AuthPermission struct {
	ID            uint64 `gorm:"column:id;type:int4;primary_key" json:"id"`
	Name          string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	ContentTypeID int    `gorm:"column:content_type_id;type:int4;not null" json:"contentTypeID"`
	Codename      string `gorm:"column:codename;type:varchar(100);not null" json:"codename"`
}

// TableName table name
func (m *AuthPermission) TableName() string {
	return "auth_permission"
}

// AuthPermissionColumnNames Whitelist for custom query fields to prevent sql injection attacks
var AuthPermissionColumnNames = map[string]bool{
	"id":              true,
	"name":            true,
	"content_type_id": true,
	"codename":        true,
}
