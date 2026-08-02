package model

type UsersUserGroups struct {
	ID      uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	UserID  int64  `gorm:"column:user_id;type:int8;not null" json:"userID"`
	GroupID int64  `gorm:"column:group_id;type:int8;not null" json:"groupID"`
}

// UsersUserGroupsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var UsersUserGroupsColumnNames = map[string]bool{
	"id":       true,
	"user_id":  true,
	"group_id": true,
}
