package model

type ExtrasNotificationgroupGroups struct {
	ID                  uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	NotificationgroupID int64  `gorm:"column:notificationgroup_id;type:int8;not null" json:"notificationgroupID"`
	GroupID             int64  `gorm:"column:group_id;type:int8;not null" json:"groupID"`
}

// ExtrasNotificationgroupGroupsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasNotificationgroupGroupsColumnNames = map[string]bool{
	"id":                   true,
	"notificationgroup_id": true,
	"group_id":             true,
}
