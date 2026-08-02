package model

type TenancyContactGroups struct {
	ID             uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ContactID      int64  `gorm:"column:contact_id;type:int8;not null" json:"contactID"`
	ContactgroupID int64  `gorm:"column:contactgroup_id;type:int8;not null" json:"contactgroupID"`
}

// TenancyContactGroupsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var TenancyContactGroupsColumnNames = map[string]bool{
	"id":              true,
	"contact_id":      true,
	"contactgroup_id": true,
}
