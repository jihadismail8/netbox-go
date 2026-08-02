package dcim

import "netbox-go/internal/domain/shared"

// RackSnapshot is the typed first-profile change projection used when a
// RackType save propagates inherited physical attributes to referencing Racks.
type RackSnapshot struct {
	SiteID       shared.ID
	Name         string
	FacilityID   *string
	RackTypeID   *shared.ID
	Status       string
	RoleID       *shared.ID
	Serial       string
	AssetTag     *string
	FormFactor   *string
	Width        uint32
	UHeight      uint32
	StartingUnit uint32
	DescUnits    bool
	Airflow      *string
	Description  string
	Comments     string
}

func (RackSnapshot) ObjectType() string { return RackObjectType }

func (snapshot RackSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	snapshot.FacilityID = cloneString(snapshot.FacilityID)
	snapshot.RackTypeID = cloneID(snapshot.RackTypeID)
	snapshot.RoleID = cloneID(snapshot.RoleID)
	snapshot.AssetTag = cloneString(snapshot.AssetTag)
	snapshot.FormFactor = cloneString(snapshot.FormFactor)
	snapshot.Airflow = cloneString(snapshot.Airflow)
	return snapshot
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneID(value *shared.ID) *shared.ID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
