package model

type ExtrasNotificationgroupUsers struct {
	ID                  uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	NotificationgroupID int64  `gorm:"column:notificationgroup_id;type:int8;not null" json:"notificationgroupID"`
	UserID              int64  `gorm:"column:user_id;type:int8;not null" json:"userID"`
}

// ExtrasNotificationgroupUsersColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasNotificationgroupUsersColumnNames = map[string]bool{
	"id":                   true,
	"notificationgroup_id": true,
	"user_id":              true,
}
