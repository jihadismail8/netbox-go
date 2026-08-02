package dcim

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	DeviceObjectType           = "dcim.device"
	DeviceNameMaxLength        = 64
	DeviceSerialMaxLength      = 50
	DeviceAssetTagMaxLength    = 50
	DeviceDescriptionMaxLength = 200
)

// DeviceStatus is the exact value persisted by the pinned Device contract.
type DeviceStatus string

const (
	DeviceStatusOffline         DeviceStatus = "offline"
	DeviceStatusActive          DeviceStatus = "active"
	DeviceStatusPlanned         DeviceStatus = "planned"
	DeviceStatusStaged          DeviceStatus = "staged"
	DeviceStatusFailed          DeviceStatus = "failed"
	DeviceStatusInventory       DeviceStatus = "inventory"
	DeviceStatusDecommissioning DeviceStatus = "decommissioning"
)

func ParseDeviceStatus(value string) (DeviceStatus, bool) {
	status := DeviceStatus(value)
	switch status {
	case DeviceStatusOffline,
		DeviceStatusActive,
		DeviceStatusPlanned,
		DeviceStatusStaged,
		DeviceStatusFailed,
		DeviceStatusInventory,
		DeviceStatusDecommissioning:
		return status, true
	default:
		return "", false
	}
}

func (status DeviceStatus) String() string { return string(status) }

// DeviceFace preserves the pinned blank/front/rear values. A blank face is a
// present empty choice, not a rack position.
type DeviceFace string

const (
	DeviceFaceFront DeviceFace = "front"
	DeviceFaceRear  DeviceFace = "rear"
)

func ParseDeviceFace(value string) (DeviceFace, bool) {
	face := DeviceFace(value)
	switch face {
	case "", DeviceFaceFront, DeviceFaceRear:
		return face, true
	default:
		return "", false
	}
}

func (face DeviceFace) String() string { return string(face) }

// RackPosition stores exact half-units. This avoids floating-point occupancy
// decisions while retaining the public decimal representation.
type RackPosition struct {
	halfUnits uint16
}

func ParseRackPosition(value string) (RackPosition, error) {
	value = strings.TrimSpace(value)
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return RackPosition{}, invalidRackPosition("Position must be a number.")
	}
	doubled := number * 2
	if math.Trunc(doubled) != doubled {
		return RackPosition{}, invalidRackPosition(
			"Position must be in increments of 0.5 rack units.",
		)
	}
	if number < 1 || number > 100.5 {
		return RackPosition{}, invalidRackPosition(
			"Position must be between 1.0 and 100.5 rack units.",
		)
	}
	return RackPosition{halfUnits: uint16(doubled)}, nil
}

func RackPositionFromHalfUnits(halfUnits uint16) (RackPosition, error) {
	if halfUnits < 2 || halfUnits > 201 {
		return RackPosition{}, invalidRackPosition(
			"Position must be between 1.0 and 100.5 rack units.",
		)
	}
	return RackPosition{halfUnits: halfUnits}, nil
}

func invalidRackPosition(description string) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "position", Reason: "invalid", Description: description,
	})
}

func (position RackPosition) HalfUnits() uint16 { return position.halfUnits }
func (position RackPosition) String() string {
	whole := position.halfUnits / 2
	if position.halfUnits%2 == 0 {
		return strconv.FormatUint(uint64(whole), 10)
	}
	return fmt.Sprintf("%d.5", whole)
}

// DeviceNullable preserves NULL independently from a present zero or blank
// value. This is required for nullable names and unique blank asset tags.
type DeviceNullable[T any] struct {
	value T
	valid bool
}

func NullDeviceValue[T any]() DeviceNullable[T] { return DeviceNullable[T]{} }
func NonNullDeviceValue[T any](value T) DeviceNullable[T] {
	return DeviceNullable[T]{value: value, valid: true}
}
func (nullable DeviceNullable[T]) Get() (T, bool) { return nullable.value, nullable.valid }
func (nullable DeviceNullable[T]) IsNull() bool   { return !nullable.valid }

// DeviceTypeInstanceReference carries exactly the DeviceType facts required
// by Device validation and display. It is immutable and never persisted from
// client input without being resolved through a repository.
type DeviceTypeInstanceReference struct {
	id               shared.ID
	model            string
	slug             shared.Slug
	manufacturerName string
	height           DeviceHeight
	fullDepth        bool
	airflow          NullableDeviceAirflow
}

func NewDeviceTypeInstanceReference(
	id shared.ID,
	model string,
	slug string,
	manufacturerName string,
	height DeviceHeight,
	fullDepth bool,
	airflow NullableDeviceAirflow,
) (DeviceTypeInstanceReference, error) {
	model = strings.TrimSpace(model)
	manufacturerName = strings.TrimSpace(manufacturerName)
	parsedSlug, err := shared.ParseSlug(strings.TrimSpace(slug), DeviceTypeSlugMaxLength)
	if !id.IsValid() || model == "" || manufacturerName == "" || err != nil {
		return DeviceTypeInstanceReference{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid DeviceType instance reference.",
		)
	}
	return DeviceTypeInstanceReference{
		id: id, model: model, slug: parsedSlug, manufacturerName: manufacturerName,
		height: height, fullDepth: fullDepth, airflow: airflow,
	}, nil
}

func (reference DeviceTypeInstanceReference) ID() shared.ID     { return reference.id }
func (reference DeviceTypeInstanceReference) Model() string     { return reference.model }
func (reference DeviceTypeInstanceReference) Slug() shared.Slug { return reference.slug }
func (reference DeviceTypeInstanceReference) Display() string   { return reference.model }
func (reference DeviceTypeInstanceReference) FullName() string {
	return reference.manufacturerName + " " + reference.model
}
func (reference DeviceTypeInstanceReference) Height() DeviceHeight {
	return reference.height
}
func (reference DeviceTypeInstanceReference) IsFullDepth() bool {
	return reference.fullDepth
}
func (reference DeviceTypeInstanceReference) Airflow() NullableDeviceAirflow {
	return reference.airflow
}
func (reference DeviceTypeInstanceReference) Valid() bool {
	return reference.id.IsValid() && reference.model != "" &&
		reference.slug.String() != "" && reference.manufacturerName != ""
}

// RackReference is the immutable subset of a Rack needed by a Device.
type RackReference struct {
	id           shared.ID
	display      string
	siteID       shared.ID
	startingUnit uint32
	uHeight      uint32
}

func NewRackReference(
	id shared.ID,
	display string,
	siteID shared.ID,
	startingUnit uint32,
	uHeight uint32,
) (RackReference, error) {
	display = strings.TrimSpace(display)
	if !id.IsValid() || display == "" || !siteID.IsValid() ||
		startingUnit == 0 || uHeight == 0 {
		return RackReference{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid Rack reference.",
		)
	}
	return RackReference{
		id: id, display: display, siteID: siteID,
		startingUnit: startingUnit, uHeight: uHeight,
	}, nil
}

func (reference RackReference) ID() shared.ID        { return reference.id }
func (reference RackReference) Display() string      { return reference.display }
func (reference RackReference) SiteID() shared.ID    { return reference.siteID }
func (reference RackReference) StartingUnit() uint32 { return reference.startingUnit }
func (reference RackReference) UHeight() uint32      { return reference.uHeight }
func (reference RackReference) Valid() bool {
	return reference.id.IsValid() && reference.display != "" &&
		reference.siteID.IsValid() && reference.startingUnit > 0 && reference.uHeight > 0
}

type DeviceValues struct {
	DeviceType  DeviceTypeInstanceReference
	Role        DeviceRoleReference
	Name        DeviceNullable[string]
	Site        SiteReference
	Rack        DeviceNullable[RackReference]
	Position    DeviceNullable[RackPosition]
	Face        string
	Status      string
	Serial      string
	AssetTag    DeviceNullable[string]
	Airflow     NullableDeviceAirflow
	Description string
	Comments    string
}

type DevicePatch struct {
	DeviceType  *DeviceTypeInstanceReference
	Role        *DeviceRoleReference
	Name        *DeviceNullable[string]
	Site        *SiteReference
	Rack        *DeviceNullable[RackReference]
	Position    *DeviceNullable[RackPosition]
	Face        *string
	Status      *string
	Serial      *string
	AssetTag    *DeviceNullable[string]
	Airflow     *NullableDeviceAirflow
	Description *string
	Comments    *string
}

func (patch DevicePatch) Empty() bool {
	return patch.DeviceType == nil && patch.Role == nil && patch.Name == nil &&
		patch.Site == nil && patch.Rack == nil && patch.Position == nil &&
		patch.Face == nil && patch.Status == nil && patch.Serial == nil &&
		patch.AssetTag == nil && patch.Airflow == nil &&
		patch.Description == nil && patch.Comments == nil
}

type DeviceState struct {
	ID             shared.ID
	DeviceType     DeviceTypeInstanceReference
	Role           DeviceRoleReference
	Name           DeviceNullable[string]
	Site           SiteReference
	Rack           DeviceNullable[RackReference]
	Position       DeviceNullable[RackPosition]
	Face           string
	Status         string
	Serial         string
	AssetTag       DeviceNullable[string]
	Airflow        NullableDeviceAirflow
	Description    string
	Comments       string
	Created        shared.Timestamp
	LastUpdated    shared.Timestamp
	InterfaceCount uint64
}

type DeviceSnapshot struct {
	DeviceTypeID shared.ID
	RoleID       shared.ID
	Name         *string
	SiteID       shared.ID
	RackID       *shared.ID
	Position     *string
	Face         string
	Status       string
	Serial       string
	AssetTag     *string
	Airflow      *string
	Description  string
	Comments     string
}

func (DeviceSnapshot) ObjectType() string { return DeviceObjectType }
func (snapshot DeviceSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	snapshot.Name = cloneString(snapshot.Name)
	snapshot.RackID = cloneID(snapshot.RackID)
	snapshot.Position = cloneString(snapshot.Position)
	snapshot.AssetTag = cloneString(snapshot.AssetTag)
	snapshot.Airflow = cloneString(snapshot.Airflow)
	return snapshot
}

type Device struct {
	id             shared.ID
	deviceType     DeviceTypeInstanceReference
	role           DeviceRoleReference
	name           DeviceNullable[string]
	site           SiteReference
	rack           DeviceNullable[RackReference]
	position       DeviceNullable[RackPosition]
	face           DeviceFace
	status         DeviceStatus
	serial         string
	assetTag       DeviceNullable[string]
	airflow        NullableDeviceAirflow
	description    string
	comments       string
	created        shared.Timestamp
	lastUpdated    shared.Timestamp
	interfaceCount uint64
}

func NewDevice(values DeviceValues, now shared.Timestamp) (*Device, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	if airflow, present := values.Airflow.Get(); !present || airflow == "" {
		values.Airflow = values.DeviceType.Airflow()
	}
	normalized, violations := validateDeviceValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return deviceFromNormalized(normalized, now, now), nil
}

func RestoreDevice(state DeviceState) (*Device, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot restore invalid Device identity or timestamps.",
		)
	}
	normalized, violations := validateDeviceValues(DeviceValues{
		DeviceType: state.DeviceType, Role: state.Role, Name: state.Name,
		Site: state.Site, Rack: state.Rack, Position: state.Position, Face: state.Face,
		Status: state.Status, Serial: state.Serial, AssetTag: state.AssetTag,
		Airflow: state.Airflow, Description: state.Description, Comments: state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted Device violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	device := deviceFromNormalized(normalized, state.Created, state.LastUpdated)
	device.id = state.ID
	device.interfaceCount = state.InterfaceCount
	return device, nil
}

type normalizedDeviceValues struct {
	deviceType  DeviceTypeInstanceReference
	role        DeviceRoleReference
	name        DeviceNullable[string]
	site        SiteReference
	rack        DeviceNullable[RackReference]
	position    DeviceNullable[RackPosition]
	face        DeviceFace
	status      DeviceStatus
	serial      string
	assetTag    DeviceNullable[string]
	airflow     NullableDeviceAirflow
	description string
	comments    string
}

func deviceFromNormalized(
	values normalizedDeviceValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *Device {
	return &Device{
		deviceType: values.deviceType, role: values.role, name: values.name,
		site: values.site, rack: values.rack, position: values.position,
		face: values.face, status: values.status, serial: values.serial,
		assetTag: values.assetTag, airflow: values.airflow,
		description: values.description, comments: values.comments,
		created: created, lastUpdated: lastUpdated,
	}
}

func (device *Device) AssignID(id shared.ID) error {
	if device == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid Device ID.")
	}
	if device.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned Device ID.")
	}
	device.id = id
	return nil
}

func (device *Device) Replace(values DeviceValues, now shared.Timestamp) error {
	if device == nil || now.IsZero() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot replace Device with invalid state or time.",
		)
	}
	normalized, violations := validateDeviceValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	id, created, interfaceCount := device.id, device.created, device.interfaceCount
	*device = *deviceFromNormalized(normalized, created, now)
	device.id, device.interfaceCount = id, interfaceCount
	return nil
}

func (device *Device) ApplyPatch(patch DevicePatch, now shared.Timestamp) error {
	if device == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil Device.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	values := device.Values()
	if patch.DeviceType != nil {
		values.DeviceType = *patch.DeviceType
	}
	if patch.Role != nil {
		values.Role = *patch.Role
	}
	if patch.Name != nil {
		values.Name = *patch.Name
	}
	if patch.Site != nil {
		values.Site = *patch.Site
	}
	if patch.Rack != nil {
		values.Rack = *patch.Rack
	}
	if patch.Position != nil {
		values.Position = *patch.Position
	}
	setString(&values.Face, patch.Face)
	setString(&values.Status, patch.Status)
	setString(&values.Serial, patch.Serial)
	if patch.AssetTag != nil {
		values.AssetTag = *patch.AssetTag
	}
	if patch.Airflow != nil {
		values.Airflow = *patch.Airflow
	}
	setString(&values.Description, patch.Description)
	setString(&values.Comments, patch.Comments)
	return device.Replace(values, now)
}

func (device Device) ID() shared.ID                           { return device.id }
func (device Device) DeviceType() DeviceTypeInstanceReference { return device.deviceType }
func (device Device) Role() DeviceRoleReference               { return device.role }
func (device Device) Name() DeviceNullable[string]            { return device.name }
func (device Device) Site() SiteReference                     { return device.site }
func (device Device) Rack() DeviceNullable[RackReference]     { return device.rack }
func (device Device) Position() DeviceNullable[RackPosition]  { return device.position }
func (device Device) Face() DeviceFace                        { return device.face }
func (device Device) Status() DeviceStatus                    { return device.status }
func (device Device) Serial() string                          { return device.serial }
func (device Device) AssetTag() DeviceNullable[string]        { return device.assetTag }
func (device Device) Airflow() NullableDeviceAirflow          { return device.airflow }
func (device Device) Description() string                     { return device.description }
func (device Device) Comments() string                        { return device.comments }
func (device Device) Created() shared.Timestamp               { return device.created }
func (device Device) LastUpdated() shared.Timestamp           { return device.lastUpdated }
func (device Device) InterfaceCount() uint64                  { return device.interfaceCount }

func (device Device) Display() string {
	name, hasName := device.name.Get()
	assetTag, hasAssetTag := device.assetTag.Get()
	if hasName && name != "" {
		if hasAssetTag && assetTag != "" {
			return fmt.Sprintf("%s (%s)", name, assetTag)
		}
		return name
	}
	base := device.deviceType.FullName()
	if hasAssetTag && assetTag != "" {
		return fmt.Sprintf("%s (%s)", base, assetTag)
	}
	if device.id.IsValid() {
		return fmt.Sprintf("%s (%s)", base, device.id)
	}
	return base
}

func (device Device) Values() DeviceValues {
	return DeviceValues{
		DeviceType: device.deviceType, Role: device.role, Name: device.name,
		Site: device.site, Rack: device.rack, Position: device.position,
		Face: device.face.String(), Status: device.status.String(), Serial: device.serial,
		AssetTag: device.assetTag, Airflow: device.airflow,
		Description: device.description, Comments: device.comments,
	}
}

func (device Device) State() DeviceState {
	values := device.Values()
	return DeviceState{
		ID: device.id, DeviceType: values.DeviceType, Role: values.Role,
		Name: values.Name, Site: values.Site, Rack: values.Rack,
		Position: values.Position, Face: values.Face, Status: values.Status,
		Serial: values.Serial, AssetTag: values.AssetTag, Airflow: values.Airflow,
		Description: values.Description, Comments: values.Comments,
		Created: device.created, LastUpdated: device.lastUpdated,
		InterfaceCount: device.interfaceCount,
	}
}

func (device Device) Snapshot() DeviceSnapshot {
	var name *string
	if value, present := device.name.Get(); present {
		name = &value
	}
	var rackID *shared.ID
	if rack, present := device.rack.Get(); present {
		value := rack.ID()
		rackID = &value
	}
	var position *string
	if value, present := device.position.Get(); present {
		text := value.String()
		position = &text
	}
	var assetTag *string
	if value, present := device.assetTag.Get(); present {
		assetTag = &value
	}
	var airflow *string
	if value, present := device.airflow.Get(); present {
		text := value.String()
		airflow = &text
	}
	return DeviceSnapshot{
		DeviceTypeID: device.deviceType.ID(), RoleID: device.role.ID,
		Name: name, SiteID: device.site.ID(), RackID: rackID, Position: position,
		Face: device.face.String(), Status: device.status.String(), Serial: device.serial,
		AssetTag: assetTag, Airflow: airflow,
		Description: device.description, Comments: device.comments,
	}
}

func validateDeviceValues(
	values DeviceValues,
) (normalizedDeviceValues, []shared.FieldViolation) {
	var violations []shared.FieldViolation
	if !values.DeviceType.Valid() {
		violations = append(violations, relationViolation("device_type"))
	}
	if !values.Role.ID.IsValid() || strings.TrimSpace(values.Role.Display) == "" {
		violations = append(violations, relationViolation("role"))
	}
	if !values.Site.Valid() {
		violations = append(violations, relationViolation("site"))
	}

	name := normalizeDeviceNullableString(values.Name)
	if value, present := name.Get(); present {
		validateOptionalLength(&violations, "name", value, DeviceNameMaxLength)
	}
	serial := strings.TrimSpace(values.Serial)
	validateOptionalLength(&violations, "serial", serial, DeviceSerialMaxLength)
	assetTag := normalizeDeviceNullableString(values.AssetTag)
	if value, present := assetTag.Get(); present {
		validateOptionalLength(&violations, "asset_tag", value, DeviceAssetTagMaxLength)
	}
	description := strings.TrimSpace(values.Description)
	validateOptionalLength(&violations, "description", description, DeviceDescriptionMaxLength)
	comments := strings.TrimSpace(values.Comments)

	status, validStatus := ParseDeviceStatus(strings.TrimSpace(values.Status))
	if !validStatus {
		violations = append(violations, shared.FieldViolation{
			Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	face, validFace := ParseDeviceFace(strings.TrimSpace(values.Face))
	if !validFace {
		violations = append(violations, shared.FieldViolation{
			Field: "face", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}

	rack, hasRack := values.Rack.Get()
	position, hasPosition := values.Position.Get()
	if hasRack && !rack.Valid() {
		violations = append(violations, relationViolation("rack"))
	}
	if !hasRack && face != "" {
		violations = append(violations, shared.FieldViolation{
			Field: "face", Reason: "invalid",
			Description: "Cannot select a rack face without assigning a rack.",
		})
	}
	if !hasRack && hasPosition {
		violations = append(violations, shared.FieldViolation{
			Field: "position", Reason: "invalid",
			Description: "Cannot select a rack position without assigning a rack.",
		})
	}
	if hasPosition && face == "" {
		violations = append(violations, shared.FieldViolation{
			Field: "face", Reason: "required",
			Description: "Must specify rack face when defining rack position.",
		})
	}
	if hasRack && rack.Valid() && values.Site.Valid() && rack.SiteID() != values.Site.ID() {
		violations = append(violations, shared.FieldViolation{
			Field: "rack", Reason: "invalid_relationship",
			Description: "The selected rack does not belong to this site.",
		})
	}
	if hasPosition && values.DeviceType.Valid() {
		if values.DeviceType.Height().HalfUnits() == 0 {
			violations = append(violations, shared.FieldViolation{
				Field: "position", Reason: "invalid",
				Description: "A 0U device type cannot be assigned to a rack position.",
			})
		}
		if hasRack && rack.Valid() {
			start := rack.StartingUnit() * 2
			end := start + rack.UHeight()*2
			positionStart := uint32(position.HalfUnits())
			positionEnd := positionStart + uint32(values.DeviceType.Height().HalfUnits())
			if positionStart < start || positionEnd > end {
				violations = append(violations, shared.FieldViolation{
					Field: "position", Reason: "occupied_or_out_of_bounds",
					Description: rackPlacementViolation(
						position,
						values.DeviceType,
					),
				})
			}
		}
	}

	if airflow, present := values.Airflow.Get(); present {
		parsed, valid := ParseDeviceAirflow(strings.TrimSpace(airflow.String()))
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "airflow", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
		} else {
			values.Airflow = NonNullDeviceAirflow(parsed)
		}
	}

	return normalizedDeviceValues{
		deviceType: values.DeviceType, role: values.Role, name: name,
		site: values.Site, rack: values.Rack, position: values.Position,
		face: face, status: status, serial: serial, assetTag: assetTag,
		airflow: values.Airflow, description: description, comments: comments,
	}, violations
}

func normalizeDeviceNullableString(
	value DeviceNullable[string],
) DeviceNullable[string] {
	text, present := value.Get()
	if !present {
		return NullDeviceValue[string]()
	}
	return NonNullDeviceValue(strings.TrimSpace(text))
}

func relationViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "invalid_choice", Description: "Select a valid choice.",
	}
}

func rackPlacementViolation(
	position RackPosition,
	deviceType DeviceTypeInstanceReference,
) string {
	return fmt.Sprintf(
		"U%s is already occupied or does not have sufficient space to accommodate this device type: %s (%sU)",
		position.String(),
		deviceType.Display(),
		deviceType.Height().String(),
	)
}
