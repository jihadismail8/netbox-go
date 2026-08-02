// Package changelog persists typed application change records in the existing
// append-only workflow audit table.
package changelog

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationchangelog "netbox-go/internal/application/changelog"
	domaindcim "netbox-go/internal/domain/dcim"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type Recorder struct {
	db *gorm.DB
}

var _ applicationchangelog.Recorder = (*Recorder)(nil)

func NewRecorder(db *gorm.DB) *Recorder {
	if db == nil {
		panic("postgres change recorder requires a database")
	}
	return &Recorder{db: db}
}

func (recorder *Recorder) Record(ctx context.Context, change applicationchangelog.Change) error {
	before, err := marshalSnapshot(change.Before)
	if err != nil {
		return err
	}
	after, err := marshalSnapshot(change.After)
	if err != nil {
		return err
	}

	row := ChangeRow{
		ActorID:    change.ActorID,
		Action:     string(change.Action),
		Kind:       change.ObjectType,
		ObjectID:   change.ObjectID.Int64(),
		BeforeData: before,
		AfterData:  after,
		OccurredAt: change.OccurredAt.Time,
	}
	if err := recorder.database(ctx).Create(&row).Error; err != nil {
		return shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not record object change.",
			err,
		)
	}
	return nil
}

func (recorder *Recorder) database(ctx context.Context) *gorm.DB {
	db := recorder.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type siteSnapshotJSON struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Facility    string `json:"facility"`
	Description string `json:"description"`
	Comments    string `json:"comments"`
}

type manufacturerSnapshotJSON struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type rackRoleSnapshotJSON struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type rackTypeSnapshotJSON struct {
	ManufacturerID int64  `json:"manufacturer"`
	Model          string `json:"model"`
	Slug           string `json:"slug"`
	FormFactor     string `json:"form_factor"`
	Width          uint32 `json:"width"`
	UHeight        uint32 `json:"u_height"`
	StartingUnit   uint32 `json:"starting_unit"`
	DescUnits      bool   `json:"desc_units"`
	Description    string `json:"description"`
	Comments       string `json:"comments"`
}

type deviceTypeSnapshotJSON struct {
	ManufacturerID         int64       `json:"manufacturer"`
	Model                  string      `json:"model"`
	Slug                   string      `json:"slug"`
	PartNumber             string      `json:"part_number"`
	UHeight                json.Number `json:"u_height"`
	ExcludeFromUtilization bool        `json:"exclude_from_utilization"`
	IsFullDepth            bool        `json:"is_full_depth"`
	Airflow                *string     `json:"airflow"`
	Description            string      `json:"description"`
	Comments               string      `json:"comments"`
}

type interfaceTemplateSnapshotJSON struct {
	DeviceTypeID int64  `json:"device_type"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	MgmtOnly     bool   `json:"mgmt_only"`
	Description  string `json:"description"`
}

type interfaceSnapshotJSON struct {
	DeviceID    int64   `json:"device"`
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	Type        string  `json:"type"`
	Enabled     bool    `json:"enabled"`
	MgmtOnly    bool    `json:"mgmt_only"`
	MTU         *uint32 `json:"mtu"`
	Speed       *uint64 `json:"speed"`
	Duplex      *string `json:"duplex"`
	Description string  `json:"description"`
}

type deviceSnapshotJSON struct {
	DeviceTypeID int64   `json:"device_type"`
	RoleID       int64   `json:"role"`
	Name         *string `json:"name"`
	SiteID       int64   `json:"site"`
	RackID       *int64  `json:"rack"`
	Position     *string `json:"position"`
	Face         string  `json:"face"`
	Status       string  `json:"status"`
	Serial       string  `json:"serial"`
	AssetTag     *string `json:"asset_tag"`
	Airflow      *string `json:"airflow"`
	Description  string  `json:"description"`
	Comments     string  `json:"comments"`
}

type rackSnapshotJSON struct {
	SiteID       int64   `json:"site"`
	Name         string  `json:"name"`
	FacilityID   *string `json:"facility_id"`
	RackTypeID   *int64  `json:"rack_type"`
	Status       string  `json:"status"`
	RoleID       *int64  `json:"role"`
	Serial       string  `json:"serial"`
	AssetTag     *string `json:"asset_tag"`
	FormFactor   *string `json:"form_factor"`
	Width        uint32  `json:"width"`
	UHeight      uint32  `json:"u_height"`
	StartingUnit uint32  `json:"starting_unit"`
	DescUnits    bool    `json:"desc_units"`
	Airflow      *string `json:"airflow"`
	Description  string  `json:"description"`
	Comments     string  `json:"comments"`
}

type deviceRoleSnapshotJSON struct {
	ParentID    *int64 `json:"parent"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Color       string `json:"color"`
	VMRole      bool   `json:"vm_role"`
	Description string `json:"description"`
	Comments    string `json:"comments"`
}

type vrfSnapshotJSON struct {
	Name          string  `json:"name"`
	RD            *string `json:"rd"`
	EnforceUnique bool    `json:"enforce_unique"`
	Description   string  `json:"description"`
	Comments      string  `json:"comments"`
}

type prefixSnapshotJSON struct {
	Prefix       string `json:"prefix"`
	VRFID        *int64 `json:"vrf"`
	Status       string `json:"status"`
	IsPool       bool   `json:"is_pool"`
	MarkUtilized bool   `json:"mark_utilized"`
	Description  string `json:"description"`
	Comments     string `json:"comments"`
}

type ipAddressSnapshotJSON struct {
	Address            string  `json:"address"`
	VRFID              *int64  `json:"vrf"`
	Status             string  `json:"status"`
	Role               *string `json:"role"`
	DNSName            string  `json:"dns_name"`
	Description        string  `json:"description"`
	Comments           string  `json:"comments"`
	AssignedObjectType *string `json:"assigned_object_type"`
	AssignedObjectID   *int64  `json:"assigned_object_id"`
}

func marshalSnapshot(snapshot shared.ObjectSnapshot) (datatypes.JSON, error) {
	if snapshot == nil {
		return nil, nil
	}

	var payload any
	switch typed := snapshot.(type) {
	case domaindcim.SiteSnapshot:
		payload = siteSnapshotPayload(typed)
	case *domaindcim.SiteSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = siteSnapshotPayload(*typed)
	case domaindcim.ManufacturerSnapshot:
		payload = manufacturerSnapshotPayload(typed)
	case *domaindcim.ManufacturerSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = manufacturerSnapshotPayload(*typed)
	case domaindcim.RackRoleSnapshot:
		payload = rackRoleSnapshotPayload(typed)
	case *domaindcim.RackRoleSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = rackRoleSnapshotPayload(*typed)
	case domaindcim.RackTypeSnapshot:
		payload = rackTypeSnapshotPayload(typed)
	case *domaindcim.RackTypeSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = rackTypeSnapshotPayload(*typed)
	case domaindcim.DeviceTypeSnapshot:
		payload = deviceTypeSnapshotPayload(typed)
	case *domaindcim.DeviceTypeSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = deviceTypeSnapshotPayload(*typed)
	case domaindcim.InterfaceTemplateSnapshot:
		payload = interfaceTemplateSnapshotPayload(typed)
	case *domaindcim.InterfaceTemplateSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = interfaceTemplateSnapshotPayload(*typed)
	case domaindcim.InterfaceSnapshot:
		payload = interfaceSnapshotPayload(typed)
	case *domaindcim.InterfaceSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = interfaceSnapshotPayload(*typed)
	case domaindcim.DeviceSnapshot:
		payload = deviceSnapshotPayload(typed)
	case *domaindcim.DeviceSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = deviceSnapshotPayload(*typed)
	case domaindcim.RackSnapshot:
		payload = rackSnapshotPayload(typed)
	case *domaindcim.RackSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = rackSnapshotPayload(*typed)
	case domaindcim.DeviceRoleSnapshot:
		payload = deviceRoleSnapshotPayload(typed)
	case *domaindcim.DeviceRoleSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = deviceRoleSnapshotPayload(*typed)
	case domainipam.VRFSnapshot:
		payload = vrfSnapshotPayload(typed)
	case *domainipam.VRFSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = vrfSnapshotPayload(*typed)
	case domainipam.PrefixSnapshot:
		payload = prefixSnapshotPayload(typed)
	case *domainipam.PrefixSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = prefixSnapshotPayload(*typed)
	case domainipam.IPAddressSnapshot:
		payload = ipAddressSnapshotPayload(typed)
	case *domainipam.IPAddressSnapshot:
		if typed == nil {
			return nil, nil
		}
		payload = ipAddressSnapshotPayload(*typed)
	default:
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			fmt.Sprintf("Unsupported object-change snapshot type %T.", snapshot),
		)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not serialize object-change snapshot.",
			err,
		)
	}
	return datatypes.JSON(encoded), nil
}

func siteSnapshotPayload(snapshot domaindcim.SiteSnapshot) siteSnapshotJSON {
	return siteSnapshotJSON{
		Name:        snapshot.Name,
		Slug:        snapshot.Slug,
		Status:      snapshot.Status,
		Facility:    snapshot.Facility,
		Description: snapshot.Description,
		Comments:    snapshot.Comments,
	}
}

func manufacturerSnapshotPayload(snapshot domaindcim.ManufacturerSnapshot) manufacturerSnapshotJSON {
	return manufacturerSnapshotJSON{
		Name:        snapshot.Name,
		Slug:        snapshot.Slug,
		Description: snapshot.Description,
	}
}

func rackRoleSnapshotPayload(snapshot domaindcim.RackRoleSnapshot) rackRoleSnapshotJSON {
	return rackRoleSnapshotJSON{
		Name:        snapshot.Name,
		Slug:        snapshot.Slug,
		Color:       snapshot.Color,
		Description: snapshot.Description,
	}
}

func rackTypeSnapshotPayload(snapshot domaindcim.RackTypeSnapshot) rackTypeSnapshotJSON {
	return rackTypeSnapshotJSON{
		ManufacturerID: snapshot.ManufacturerID.Int64(), Model: snapshot.Model, Slug: snapshot.Slug,
		FormFactor: snapshot.FormFactor, Width: snapshot.Width, UHeight: snapshot.UHeight,
		StartingUnit: snapshot.StartingUnit, DescUnits: snapshot.DescUnits,
		Description: snapshot.Description, Comments: snapshot.Comments,
	}
}

func deviceTypeSnapshotPayload(snapshot domaindcim.DeviceTypeSnapshot) deviceTypeSnapshotJSON {
	return deviceTypeSnapshotJSON{
		ManufacturerID: snapshot.ManufacturerID.Int64(), Model: snapshot.Model, Slug: snapshot.Slug,
		PartNumber: snapshot.PartNumber, UHeight: json.Number(snapshot.UHeight),
		ExcludeFromUtilization: snapshot.ExcludeFromUtilization, IsFullDepth: snapshot.IsFullDepth,
		Airflow:     cloneStringPointer(snapshot.Airflow),
		Description: snapshot.Description, Comments: snapshot.Comments,
	}
}

func interfaceTemplateSnapshotPayload(
	snapshot domaindcim.InterfaceTemplateSnapshot,
) interfaceTemplateSnapshotJSON {
	return interfaceTemplateSnapshotJSON{
		DeviceTypeID: snapshot.DeviceTypeID.Int64(),
		Name:         snapshot.Name,
		Label:        snapshot.Label,
		Type:         snapshot.Type,
		Enabled:      snapshot.Enabled,
		MgmtOnly:     snapshot.MgmtOnly,
		Description:  snapshot.Description,
	}
}

func interfaceSnapshotPayload(
	snapshot domaindcim.InterfaceSnapshot,
) interfaceSnapshotJSON {
	return interfaceSnapshotJSON{
		DeviceID: snapshot.DeviceID.Int64(), Name: snapshot.Name,
		Label: snapshot.Label, Type: snapshot.Type, Enabled: snapshot.Enabled,
		MgmtOnly: snapshot.MgmtOnly, MTU: cloneUint32Pointer(snapshot.MTU),
		Speed:       cloneUint64Pointer(snapshot.Speed),
		Duplex:      cloneStringPointer(snapshot.Duplex),
		Description: snapshot.Description,
	}
}

func deviceSnapshotPayload(snapshot domaindcim.DeviceSnapshot) deviceSnapshotJSON {
	return deviceSnapshotJSON{
		DeviceTypeID: snapshot.DeviceTypeID.Int64(),
		RoleID:       snapshot.RoleID.Int64(),
		Name:         cloneStringPointer(snapshot.Name),
		SiteID:       snapshot.SiteID.Int64(),
		RackID:       snapshotJSONID(snapshot.RackID),
		Position:     cloneStringPointer(snapshot.Position),
		Face:         snapshot.Face,
		Status:       snapshot.Status,
		Serial:       snapshot.Serial,
		AssetTag:     cloneStringPointer(snapshot.AssetTag),
		Airflow:      cloneStringPointer(snapshot.Airflow),
		Description:  snapshot.Description,
		Comments:     snapshot.Comments,
	}
}

func rackSnapshotPayload(snapshot domaindcim.RackSnapshot) rackSnapshotJSON {
	return rackSnapshotJSON{
		SiteID: snapshot.SiteID.Int64(), Name: snapshot.Name,
		FacilityID: cloneStringPointer(snapshot.FacilityID), RackTypeID: snapshotJSONID(snapshot.RackTypeID),
		Status: snapshot.Status, RoleID: snapshotJSONID(snapshot.RoleID), Serial: snapshot.Serial,
		AssetTag: cloneStringPointer(snapshot.AssetTag), FormFactor: cloneStringPointer(snapshot.FormFactor),
		Width: snapshot.Width, UHeight: snapshot.UHeight, StartingUnit: snapshot.StartingUnit,
		DescUnits: snapshot.DescUnits, Airflow: cloneStringPointer(snapshot.Airflow),
		Description: snapshot.Description, Comments: snapshot.Comments,
	}
}

func deviceRoleSnapshotPayload(snapshot domaindcim.DeviceRoleSnapshot) deviceRoleSnapshotJSON {
	return deviceRoleSnapshotJSON{
		ParentID: cloneInt64Pointer(snapshot.ParentID), Name: snapshot.Name, Slug: snapshot.Slug,
		Color: snapshot.Color, VMRole: snapshot.VMRole,
		Description: snapshot.Description, Comments: snapshot.Comments,
	}
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneUint32Pointer(value *uint32) *uint32 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneUint64Pointer(value *uint64) *uint64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func snapshotJSONID(value *shared.ID) *int64 {
	if value == nil {
		return nil
	}
	primitive := value.Int64()
	return &primitive
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func vrfSnapshotPayload(snapshot domainipam.VRFSnapshot) vrfSnapshotJSON {
	var rdValue *string
	if rd, present := snapshot.RD.Get(); present {
		value := rd.String()
		rdValue = &value
	}
	return vrfSnapshotJSON{
		Name:          snapshot.Name,
		RD:            rdValue,
		EnforceUnique: snapshot.EnforceUnique,
		Description:   snapshot.Description,
		Comments:      snapshot.Comments,
	}
}

func prefixSnapshotPayload(snapshot domainipam.PrefixSnapshot) prefixSnapshotJSON {
	var vrfID *int64
	if reference, present := snapshot.VRF.Get(); present {
		value := reference.ID().Int64()
		vrfID = &value
	}
	return prefixSnapshotJSON{
		Prefix: snapshot.Prefix, VRFID: vrfID, Status: snapshot.Status,
		IsPool: snapshot.IsPool, MarkUtilized: snapshot.MarkUtilized,
		Description: snapshot.Description, Comments: snapshot.Comments,
	}
}

func ipAddressSnapshotPayload(
	snapshot domainipam.IPAddressSnapshot,
) ipAddressSnapshotJSON {
	return ipAddressSnapshotJSON{
		Address: snapshot.Address, VRFID: snapshotJSONID(snapshot.VRFID),
		Status: snapshot.Status, Role: cloneStringPointer(snapshot.Role),
		DNSName: snapshot.DNSName, Description: snapshot.Description,
		Comments:           snapshot.Comments,
		AssignedObjectType: cloneStringPointer(snapshot.AssignedObjectType),
		AssignedObjectID:   snapshotJSONID(snapshot.AssignedObjectID),
	}
}
