package model

type IpamServiceIpaddresses struct {
	ID          uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ServiceID   int64  `gorm:"column:service_id;type:int8;not null" json:"serviceID"`
	IpaddressID int64  `gorm:"column:ipaddress_id;type:int8;not null" json:"ipaddressID"`
}

// IpamServiceIpaddressesColumnNames Whitelist for custom query fields to prevent sql injection attacks
var IpamServiceIpaddressesColumnNames = map[string]bool{
	"id":           true,
	"service_id":   true,
	"ipaddress_id": true,
}
