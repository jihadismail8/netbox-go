package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	RackNameMaxLength        = 100
	RackFacilityIDMaxLength  = 50
	RackSerialMaxLength      = 50
	RackAssetTagMaxLength    = 50
	RackDescriptionMaxLength = 200
	RackDefaultWidth         = uint32(19)
	RackDefaultUHeight       = uint32(42)
	RackDefaultStartingUnit  = uint32(1)
)

type RackStatus string

const (
	RackStatusReserved   RackStatus = "reserved"
	RackStatusAvailable  RackStatus = "available"
	RackStatusPlanned    RackStatus = "planned"
	RackStatusActive     RackStatus = "active"
	RackStatusDeprecated RackStatus = "deprecated"
)

func ParseRackStatus(value string) (RackStatus, bool) {
	status := RackStatus(value)
	switch status {
	case RackStatusReserved, RackStatusAvailable, RackStatusPlanned, RackStatusActive, RackStatusDeprecated:
		return status, true
	default:
		return "", false
	}
}

func (status RackStatus) String() string { return string(status) }

type RackAirflow string

const (
	RackAirflowFrontToRear RackAirflow = "front-to-rear"
	RackAirflowRearToFront RackAirflow = "rear-to-front"
)

func ParseRackAirflow(value string) (RackAirflow, bool) {
	airflow := RackAirflow(value)
	switch airflow {
	case "", RackAirflowFrontToRear, RackAirflowRearToFront:
		return airflow, true
	default:
		return "", false
	}
}

func (airflow RackAirflow) String() string { return string(airflow) }

// RackNullable preserves the database distinction between NULL and a present
// zero value. This matters for blank asset tags and blank choice values.
type RackNullable[T any] struct {
	value T
	valid bool
}

func NullRackValue[T any]() RackNullable[T] { return RackNullable[T]{} }
func NonNullRackValue[T any](value T) RackNullable[T] {
	return RackNullable[T]{value: value, valid: true}
}
func (nullable RackNullable[T]) Get() (T, bool) { return nullable.value, nullable.valid }
func (nullable RackNullable[T]) IsNull() bool   { return !nullable.valid }

type SiteReference struct {
	id   shared.ID
	name string
	slug shared.Slug
}

func NewSiteReference(id shared.ID, name, slug string) (SiteReference, error) {
	name = strings.TrimSpace(name)
	parsedSlug, err := shared.ParseSlug(strings.TrimSpace(slug), SiteSlugMaxLength)
	if !id.IsValid() || name == "" || err != nil {
		return SiteReference{}, shared.NewError(
			shared.ErrorReasonInternal, "Cannot construct an invalid Site reference.",
		)
	}
	return SiteReference{id: id, name: name, slug: parsedSlug}, nil
}

func (reference SiteReference) ID() shared.ID     { return reference.id }
func (reference SiteReference) Name() string      { return reference.name }
func (reference SiteReference) Slug() shared.Slug { return reference.slug }
func (reference SiteReference) Display() string   { return reference.name }
func (reference SiteReference) Valid() bool {
	return reference.id.IsValid() && reference.name != "" && reference.slug.String() != ""
}

type RackRoleReference struct {
	id   shared.ID
	name string
	slug shared.Slug
}

func NewRackRoleReference(id shared.ID, name, slug string) (RackRoleReference, error) {
	name = strings.TrimSpace(name)
	parsedSlug, err := shared.ParseSlug(strings.TrimSpace(slug), RackRoleSlugMaxLength)
	if !id.IsValid() || name == "" || err != nil {
		return RackRoleReference{}, shared.NewError(
			shared.ErrorReasonInternal, "Cannot construct an invalid RackRole reference.",
		)
	}
	return RackRoleReference{id: id, name: name, slug: parsedSlug}, nil
}

func (reference RackRoleReference) ID() shared.ID     { return reference.id }
func (reference RackRoleReference) Name() string      { return reference.name }
func (reference RackRoleReference) Slug() shared.Slug { return reference.slug }
func (reference RackRoleReference) Display() string   { return reference.name }
func (reference RackRoleReference) Valid() bool {
	return reference.id.IsValid() && reference.name != "" && reference.slug.String() != ""
}

type RackTypeReference struct {
	id         shared.ID
	model      string
	slug       shared.Slug
	attributes RackPhysicalAttributes
}

func NewRackTypeReference(
	id shared.ID,
	model string,
	slug string,
	attributes RackPhysicalAttributes,
) (RackTypeReference, error) {
	model = strings.TrimSpace(model)
	parsedSlug, err := shared.ParseSlug(strings.TrimSpace(slug), RackTypeSlugMaxLength)
	if !id.IsValid() || model == "" || err != nil ||
		attributes.FormFactor == "" || attributes.Width == 0 ||
		attributes.UHeight == 0 || attributes.StartingUnit == 0 {
		return RackTypeReference{}, shared.NewError(
			shared.ErrorReasonInternal, "Cannot construct an invalid RackType reference.",
		)
	}
	return RackTypeReference{id: id, model: model, slug: parsedSlug, attributes: attributes}, nil
}

func (reference RackTypeReference) ID() shared.ID     { return reference.id }
func (reference RackTypeReference) Model() string     { return reference.model }
func (reference RackTypeReference) Slug() shared.Slug { return reference.slug }
func (reference RackTypeReference) Display() string   { return reference.model }
func (reference RackTypeReference) PhysicalAttributes() RackPhysicalAttributes {
	return reference.attributes
}
func (reference RackTypeReference) Valid() bool {
	return reference.id.IsValid() && reference.model != "" && reference.slug.String() != ""
}

type RackValues struct {
	Site         SiteReference
	Name         string
	FacilityID   RackNullable[string]
	RackType     RackNullable[RackTypeReference]
	Status       string
	Role         RackNullable[RackRoleReference]
	Serial       string
	AssetTag     RackNullable[string]
	FormFactor   RackNullable[string]
	Width        uint32
	UHeight      uint32
	StartingUnit uint32
	DescUnits    bool
	Airflow      RackNullable[string]
	Description  string
	Comments     string
}

type RackPatch struct {
	Site         *SiteReference
	Name         *string
	FacilityID   *RackNullable[string]
	RackType     *RackNullable[RackTypeReference]
	Status       *string
	Role         *RackNullable[RackRoleReference]
	Serial       *string
	AssetTag     *RackNullable[string]
	FormFactor   *RackNullable[string]
	Width        *uint32
	UHeight      *uint32
	StartingUnit *uint32
	DescUnits    *bool
	Airflow      *RackNullable[string]
	Description  *string
	Comments     *string
}

func (patch RackPatch) Empty() bool {
	return patch.Site == nil && patch.Name == nil && patch.FacilityID == nil &&
		patch.RackType == nil && patch.Status == nil && patch.Role == nil &&
		patch.Serial == nil && patch.AssetTag == nil && patch.FormFactor == nil &&
		patch.Width == nil && patch.UHeight == nil && patch.StartingUnit == nil &&
		patch.DescUnits == nil && patch.Airflow == nil && patch.Description == nil &&
		patch.Comments == nil
}

type RackState struct {
	ID           shared.ID
	Site         SiteReference
	Name         string
	FacilityID   RackNullable[string]
	RackType     RackNullable[RackTypeReference]
	Status       string
	Role         RackNullable[RackRoleReference]
	Serial       string
	AssetTag     RackNullable[string]
	FormFactor   RackNullable[string]
	Width        uint32
	UHeight      uint32
	StartingUnit uint32
	DescUnits    bool
	Airflow      RackNullable[string]
	Description  string
	Comments     string
	Created      shared.Timestamp
	LastUpdated  shared.Timestamp
	DeviceCount  uint64
}

type Rack struct {
	id           shared.ID
	site         SiteReference
	name         string
	facilityID   RackNullable[string]
	rackType     RackNullable[RackTypeReference]
	status       RackStatus
	role         RackNullable[RackRoleReference]
	serial       string
	assetTag     RackNullable[string]
	formFactor   RackNullable[RackFormFactor]
	width        RackWidth
	uHeight      uint32
	startingUnit uint32
	descUnits    bool
	airflow      RackNullable[RackAirflow]
	description  string
	comments     string
	created      shared.Timestamp
	lastUpdated  shared.Timestamp
	deviceCount  uint64
}

func NewRack(values RackValues, now shared.Timestamp) (*Rack, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateRackValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return rackFromNormalized(normalized, now, now), nil
}

func RestoreRack(state RackState) (*Rack, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore invalid Rack identity or timestamps.")
	}
	normalized, violations := validateRackValues(RackValues{
		Site: state.Site, Name: state.Name, FacilityID: state.FacilityID,
		RackType: state.RackType, Status: state.Status, Role: state.Role,
		Serial: state.Serial, AssetTag: state.AssetTag, FormFactor: state.FormFactor,
		Width: state.Width, UHeight: state.UHeight, StartingUnit: state.StartingUnit,
		DescUnits: state.DescUnits, Airflow: state.Airflow,
		Description: state.Description, Comments: state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted Rack violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	rack := rackFromNormalized(normalized, state.Created, state.LastUpdated)
	rack.id = state.ID
	rack.deviceCount = state.DeviceCount
	return rack, nil
}

func rackFromNormalized(
	values normalizedRackValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *Rack {
	return &Rack{
		site: values.site, name: values.name, facilityID: values.facilityID,
		rackType: values.rackType, status: values.status, role: values.role,
		serial: values.serial, assetTag: values.assetTag, formFactor: values.formFactor,
		width: values.width, uHeight: values.uHeight, startingUnit: values.startingUnit,
		descUnits: values.descUnits, airflow: values.airflow,
		description: values.description, comments: values.comments,
		created: created, lastUpdated: lastUpdated,
	}
}

func (rack *Rack) AssignID(id shared.ID) error {
	if rack == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid Rack ID.")
	}
	if rack.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned Rack ID.")
	}
	rack.id = id
	return nil
}

func (rack *Rack) Replace(values RackValues, now shared.Timestamp) error {
	if rack == nil || now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace Rack with invalid state or time.")
	}
	normalized, violations := validateRackValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	id, created, deviceCount := rack.id, rack.created, rack.deviceCount
	*rack = *rackFromNormalized(normalized, created, now)
	rack.id, rack.deviceCount = id, deviceCount
	return nil
}

func (rack *Rack) ApplyPatch(patch RackPatch, now shared.Timestamp) error {
	if rack == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil Rack.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		})
	}
	values := rack.Values()
	if patch.Site != nil {
		values.Site = *patch.Site
	}
	setString(&values.Name, patch.Name)
	if patch.FacilityID != nil {
		values.FacilityID = *patch.FacilityID
	}
	if patch.RackType != nil {
		values.RackType = *patch.RackType
	}
	setString(&values.Status, patch.Status)
	if patch.Role != nil {
		values.Role = *patch.Role
	}
	setString(&values.Serial, patch.Serial)
	if patch.AssetTag != nil {
		values.AssetTag = *patch.AssetTag
	}
	if patch.FormFactor != nil {
		values.FormFactor = *patch.FormFactor
	}
	setUint32(&values.Width, patch.Width)
	setUint32(&values.UHeight, patch.UHeight)
	setUint32(&values.StartingUnit, patch.StartingUnit)
	if patch.DescUnits != nil {
		values.DescUnits = *patch.DescUnits
	}
	if patch.Airflow != nil {
		values.Airflow = *patch.Airflow
	}
	setString(&values.Description, patch.Description)
	setString(&values.Comments, patch.Comments)
	return rack.Replace(values, now)
}

func (rack *Rack) ApplyRackTypeOwnership(now shared.Timestamp) error {
	if rack == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot apply RackType ownership to a nil Rack.")
	}
	reference, present := rack.rackType.Get()
	if !present {
		return nil
	}
	attributes := reference.PhysicalAttributes()
	factor := attributes.FormFactor.String()
	return rack.ApplyPatch(RackPatch{
		FormFactor:   rackNullablePointer(NonNullRackValue(factor)),
		Width:        uint32Pointer(attributes.Width.Uint32()),
		UHeight:      uint32Pointer(attributes.UHeight),
		StartingUnit: uint32Pointer(attributes.StartingUnit),
		DescUnits:    boolPointer(attributes.DescUnits),
	}, now)
}

func (rack Rack) ID() shared.ID                             { return rack.id }
func (rack Rack) Site() SiteReference                       { return rack.site }
func (rack Rack) Name() string                              { return rack.name }
func (rack Rack) FacilityID() RackNullable[string]          { return rack.facilityID }
func (rack Rack) RackType() RackNullable[RackTypeReference] { return rack.rackType }
func (rack Rack) Status() RackStatus                        { return rack.status }
func (rack Rack) Role() RackNullable[RackRoleReference]     { return rack.role }
func (rack Rack) Serial() string                            { return rack.serial }
func (rack Rack) AssetTag() RackNullable[string]            { return rack.assetTag }
func (rack Rack) FormFactor() RackNullable[RackFormFactor]  { return rack.formFactor }
func (rack Rack) Width() RackWidth                          { return rack.width }
func (rack Rack) UHeight() uint32                           { return rack.uHeight }
func (rack Rack) StartingUnit() uint32                      { return rack.startingUnit }
func (rack Rack) DescUnits() bool                           { return rack.descUnits }
func (rack Rack) Airflow() RackNullable[RackAirflow]        { return rack.airflow }
func (rack Rack) Description() string                       { return rack.description }
func (rack Rack) Comments() string                          { return rack.comments }
func (rack Rack) Created() shared.Timestamp                 { return rack.created }
func (rack Rack) LastUpdated() shared.Timestamp             { return rack.lastUpdated }
func (rack Rack) DeviceCount() uint64                       { return rack.deviceCount }
func (rack Rack) Display() string {
	if facilityID, present := rack.facilityID.Get(); present && facilityID != "" {
		return rack.name + " (" + facilityID + ")"
	}
	return rack.name
}

func (rack Rack) Values() RackValues {
	formFactor := NullRackValue[string]()
	if value, present := rack.formFactor.Get(); present {
		formFactor = NonNullRackValue(value.String())
	}
	airflow := NullRackValue[string]()
	if value, present := rack.airflow.Get(); present {
		airflow = NonNullRackValue(value.String())
	}
	return RackValues{
		Site: rack.site, Name: rack.name, FacilityID: rack.facilityID,
		RackType: rack.rackType, Status: rack.status.String(), Role: rack.role,
		Serial: rack.serial, AssetTag: rack.assetTag, FormFactor: formFactor,
		Width: rack.width.Uint32(), UHeight: rack.uHeight, StartingUnit: rack.startingUnit,
		DescUnits: rack.descUnits, Airflow: airflow,
		Description: rack.description, Comments: rack.comments,
	}
}

func (rack Rack) State() RackState {
	values := rack.Values()
	return RackState{
		ID: rack.id, Site: values.Site, Name: values.Name, FacilityID: values.FacilityID,
		RackType: values.RackType, Status: values.Status, Role: values.Role,
		Serial: values.Serial, AssetTag: values.AssetTag, FormFactor: values.FormFactor,
		Width: values.Width, UHeight: values.UHeight, StartingUnit: values.StartingUnit,
		DescUnits: values.DescUnits, Airflow: values.Airflow,
		Description: values.Description, Comments: values.Comments,
		Created: rack.created, LastUpdated: rack.lastUpdated, DeviceCount: rack.deviceCount,
	}
}

func (rack Rack) Snapshot() RackSnapshot {
	state := rack.State()
	return RackSnapshot{
		SiteID: state.Site.ID(), Name: state.Name,
		FacilityID: nullableStringPointer(state.FacilityID),
		RackTypeID: nullableRackTypeID(state.RackType), Status: state.Status,
		RoleID: nullableRackRoleID(state.Role), Serial: state.Serial,
		AssetTag:   nullableStringPointer(state.AssetTag),
		FormFactor: nullableStringPointer(state.FormFactor),
		Width:      state.Width, UHeight: state.UHeight, StartingUnit: state.StartingUnit,
		DescUnits: state.DescUnits, Airflow: nullableStringPointer(state.Airflow),
		Description: state.Description, Comments: state.Comments,
	}
}

type normalizedRackValues struct {
	site         SiteReference
	name         string
	facilityID   RackNullable[string]
	rackType     RackNullable[RackTypeReference]
	status       RackStatus
	role         RackNullable[RackRoleReference]
	serial       string
	assetTag     RackNullable[string]
	formFactor   RackNullable[RackFormFactor]
	width        RackWidth
	uHeight      uint32
	startingUnit uint32
	descUnits    bool
	airflow      RackNullable[RackAirflow]
	description  string
	comments     string
}

func validateRackValues(values RackValues) (normalizedRackValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Status = strings.TrimSpace(values.Status)
	values.Serial = strings.TrimSpace(values.Serial)
	values.Description = strings.TrimSpace(values.Description)
	values.Comments = strings.TrimSpace(values.Comments)
	var violations []shared.FieldViolation
	if !values.Site.Valid() {
		violations = append(violations, invalidRackReference("site"))
	}
	validateRequiredLength(&violations, "name", values.Name, RackNameMaxLength)
	validateOptionalLength(&violations, "serial", values.Serial, RackSerialMaxLength)
	validateOptionalLength(&violations, "description", values.Description, RackDescriptionMaxLength)
	facilityID := normalizeNullableRackString(&violations, "facility_id", values.FacilityID, RackFacilityIDMaxLength)
	assetTag := normalizeNullableRackString(&violations, "asset_tag", values.AssetTag, RackAssetTagMaxLength)
	if reference, present := values.RackType.Get(); present && !reference.Valid() {
		violations = append(violations, invalidRackReference("rack_type"))
	}
	if reference, present := values.Role.Get(); present && !reference.Valid() {
		violations = append(violations, invalidRackReference("role"))
	}
	status, validStatus := ParseRackStatus(values.Status)
	if !validStatus {
		violations = append(violations, shared.FieldViolation{
			Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	formFactor := NullRackValue[RackFormFactor]()
	if value, present := values.FormFactor.Get(); present {
		value = strings.TrimSpace(value)
		if value == "" {
			formFactor = NonNullRackValue(RackFormFactor(""))
		} else if parsed, valid := ParseRackFormFactor(value); valid {
			formFactor = NonNullRackValue(parsed)
		} else {
			violations = append(violations, shared.FieldViolation{
				Field: "form_factor", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
		}
	}
	width, validWidth := ParseRackWidth(values.Width)
	if !validWidth {
		violations = append(violations, shared.FieldViolation{
			Field: "width", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if values.UHeight < 1 || values.UHeight > RackTypeMaximumUHeight {
		violations = append(violations, shared.FieldViolation{
			Field: "u_height", Reason: "range", Description: "Ensure this value is between 1 and 100.",
		})
	}
	if values.StartingUnit < 1 || values.StartingUnit > RackTypeMaximumStartingUnit {
		violations = append(violations, shared.FieldViolation{
			Field: "starting_unit", Reason: "range",
			Description: "Ensure this value is greater than or equal to 1.",
		})
	}
	airflow := NullRackValue[RackAirflow]()
	if value, present := values.Airflow.Get(); present {
		value = strings.TrimSpace(value)
		if parsed, valid := ParseRackAirflow(value); valid {
			airflow = NonNullRackValue(parsed)
		} else {
			violations = append(violations, shared.FieldViolation{
				Field: "airflow", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
		}
	}
	return normalizedRackValues{
		site: values.Site, name: values.Name, facilityID: facilityID, rackType: values.RackType,
		status: status, role: values.Role, serial: values.Serial, assetTag: assetTag,
		formFactor: formFactor, width: width, uHeight: values.UHeight,
		startingUnit: values.StartingUnit, descUnits: values.DescUnits, airflow: airflow,
		description: values.Description, comments: values.Comments,
	}, violations
}

func invalidRackReference(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "invalid_choice", Description: "Select a valid choice.",
	}
}

func normalizeNullableRackString(
	violations *[]shared.FieldViolation,
	field string,
	value RackNullable[string],
	maxLength int,
) RackNullable[string] {
	text, present := value.Get()
	if !present {
		return NullRackValue[string]()
	}
	text = strings.TrimSpace(text)
	validateOptionalLength(violations, field, text, maxLength)
	return NonNullRackValue(text)
}

func nullableStringPointer(value RackNullable[string]) *string {
	text, present := value.Get()
	if !present {
		return nil
	}
	return &text
}

func nullableRackTypeID(value RackNullable[RackTypeReference]) *shared.ID {
	reference, present := value.Get()
	if !present {
		return nil
	}
	id := reference.ID()
	return &id
}

func nullableRackRoleID(value RackNullable[RackRoleReference]) *shared.ID {
	reference, present := value.Get()
	if !present {
		return nil
	}
	id := reference.ID()
	return &id
}

func rackNullablePointer[T any](value RackNullable[T]) *RackNullable[T] { return &value }
func uint32Pointer(value uint32) *uint32                                { return &value }
func boolPointer(value bool) *bool                                      { return &value }
