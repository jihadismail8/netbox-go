package model

type ExtrasConfigcontextLocations struct {
	ID              uint64 `gorm:"column:id;type:int8;primary_key" json:"id"`
	ConfigcontextID int64  `gorm:"column:configcontext_id;type:int8;not null" json:"configcontextID"`
	LocationID      int64  `gorm:"column:location_id;type:int8;not null" json:"locationID"`
}

// ExtrasConfigcontextLocationsColumnNames Whitelist for custom query fields to prevent sql injection attacks
var ExtrasConfigcontextLocationsColumnNames = map[string]bool{
	"id":               true,
	"configcontext_id": true,
	"location_id":      true,
}
