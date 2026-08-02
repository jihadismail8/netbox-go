package model

type UsersGroup struct {
	ID          uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	Name        string `gorm:"column:name;type:varchar(150);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(200);not null" json:"description"`
}

// TableName table name
func (m *UsersGroup) TableName() string {
	return "users_group"
}

// UsersGroupColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersGroupColumnNames = map[string]bool{
	"id":          true,
	"name":        true,
	"description": true,
}
