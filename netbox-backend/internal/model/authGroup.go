package model

type AuthGroup struct {
	ID   uint64 `gorm:"column:id;type:int4;primary_key" json:"id"`
	Name string `gorm:"column:name;type:varchar(150);not null" json:"name"`
}

// TableName table name
func (m *AuthGroup) TableName() string {
	return "auth_group"
}

// AuthGroupColumnNames Whitelist for custom query fields to prevent sql injection attacks
var AuthGroupColumnNames = map[string]bool{
	"id":   true,
	"name": true,
}
