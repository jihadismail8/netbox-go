package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	RackTypeObjectType           = "dcim.racktype"
	RackObjectType               = "dcim.rack"
	RackTypeModelMaxLength       = 100
	RackTypeSlugMaxLength        = 100
	RackTypeDescriptionMaxLength = 200
	RackTypeDefaultWidth         = uint32(19)
	RackTypeDefaultUHeight       = uint32(42)
	RackTypeDefaultStartingUnit  = uint32(1)
	RackTypeMaximumUHeight       = uint32(100)
	RackTypeMaximumStartingUnit  = uint32(32767)
)

// RackFormFactor is the exact value stored by the pinned RackFormFactorChoices.
type RackFormFactor string

const (
	RackFormFactorTwoPostFrame        RackFormFactor = "2-post-frame"
	RackFormFactorFourPostFrame       RackFormFactor = "4-post-frame"
	RackFormFactorFourPostCabinet     RackFormFactor = "4-post-cabinet"
	RackFormFactorWallFrame           RackFormFactor = "wall-frame"
	RackFormFactorWallFrameVertical   RackFormFactor = "wall-frame-vertical"
	RackFormFactorWallCabinet         RackFormFactor = "wall-cabinet"
	RackFormFactorWallCabinetVertical RackFormFactor = "wall-cabinet-vertical"
)

func ParseRackFormFactor(value string) (RackFormFactor, bool) {
	factor := RackFormFactor(value)
	switch factor {
	case RackFormFactorTwoPostFrame,
		RackFormFactorFourPostFrame,
		RackFormFactorFourPostCabinet,
		RackFormFactorWallFrame,
		RackFormFactorWallFrameVertical,
		RackFormFactorWallCabinet,
		RackFormFactorWallCabinetVertical:
		return factor, true
	default:
		return "", false
	}
}

func (factor RackFormFactor) String() string { return string(factor) }

// RackWidth is restricted to the four pinned RackWidthChoices values.
type RackWidth uint32

const (
	RackWidth10 RackWidth = 10
	RackWidth19 RackWidth = 19
	RackWidth21 RackWidth = 21
	RackWidth23 RackWidth = 23
)

func ParseRackWidth(value uint32) (RackWidth, bool) {
	width := RackWidth(value)
	switch width {
	case RackWidth10, RackWidth19, RackWidth21, RackWidth23:
		return width, true
	default:
		return 0, false
	}
}

func (width RackWidth) Uint32() uint32 { return uint32(width) }

// ManufacturerReference is a validated immutable projection used by typed
// aggregates which belong to a Manufacturer.
type ManufacturerReference struct {
	id   shared.ID
	name string
	slug shared.Slug
}

func NewManufacturerReference(id shared.ID, name, slug string) (ManufacturerReference, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if !id.IsValid() || name == "" {
		return ManufacturerReference{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid Manufacturer reference.",
		)
	}
	parsedSlug, err := shared.ParseSlug(slug, ManufacturerSlugMaxLength)
	if err != nil {
		return ManufacturerReference{}, shared.WrapError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid Manufacturer reference.",
			err,
		)
	}
	return ManufacturerReference{id: id, name: name, slug: parsedSlug}, nil
}

func (reference ManufacturerReference) ID() shared.ID     { return reference.id }
func (reference ManufacturerReference) Name() string      { return reference.name }
func (reference ManufacturerReference) Slug() shared.Slug { return reference.slug }
func (reference ManufacturerReference) Display() string   { return reference.name }
func (reference ManufacturerReference) Valid() bool {
	return reference.id.IsValid() && reference.name != "" && reference.slug.String() != ""
}

// RackTypeValues contains the complete writable first-profile state.
type RackTypeValues struct {
	Manufacturer ManufacturerReference
	Model        string
	Slug         string
	FormFactor   string
	Width        uint32
	UHeight      uint32
	StartingUnit uint32
	DescUnits    bool
	Description  string
	Comments     string
}

type RackTypePatch struct {
	Manufacturer *ManufacturerReference
	Model        *string
	Slug         *string
	FormFactor   *string
	Width        *uint32
	UHeight      *uint32
	StartingUnit *uint32
	DescUnits    *bool
	Description  *string
	Comments     *string
}

func (patch RackTypePatch) Empty() bool {
	return patch.Manufacturer == nil && patch.Model == nil && patch.Slug == nil &&
		patch.FormFactor == nil && patch.Width == nil && patch.UHeight == nil &&
		patch.StartingUnit == nil && patch.DescUnits == nil && patch.Description == nil &&
		patch.Comments == nil
}

type RackTypeState struct {
	ID           shared.ID
	Manufacturer ManufacturerReference
	Model        string
	Slug         string
	FormFactor   string
	Width        uint32
	UHeight      uint32
	StartingUnit uint32
	DescUnits    bool
	Description  string
	Comments     string
	Created      shared.Timestamp
	LastUpdated  shared.Timestamp
}

type RackTypeSnapshot struct {
	ManufacturerID shared.ID
	Model          string
	Slug           string
	FormFactor     string
	Width          uint32
	UHeight        uint32
	StartingUnit   uint32
	DescUnits      bool
	Description    string
	Comments       string
}

func (RackTypeSnapshot) ObjectType() string                            { return RackTypeObjectType }
func (snapshot RackTypeSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }

// RackPhysicalAttributes are copied to every Rack referencing a saved
// RackType, mirroring RackType.save() in the pinned source.
type RackPhysicalAttributes struct {
	FormFactor   RackFormFactor
	Width        RackWidth
	UHeight      uint32
	StartingUnit uint32
	DescUnits    bool
}

type RackType struct {
	id           shared.ID
	manufacturer ManufacturerReference
	model        string
	slug         shared.Slug
	formFactor   RackFormFactor
	width        RackWidth
	uHeight      uint32
	startingUnit uint32
	descUnits    bool
	description  string
	comments     string
	created      shared.Timestamp
	lastUpdated  shared.Timestamp
}

func NewRackType(values RackTypeValues, now shared.Timestamp) (*RackType, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateRackTypeValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return rackTypeFromNormalized(normalized, now, now), nil
}

func RestoreRackType(state RackTypeState) (*RackType, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore invalid RackType identity or timestamps.")
	}
	normalized, violations := validateRackTypeValues(RackTypeValues{
		Manufacturer: state.Manufacturer, Model: state.Model, Slug: state.Slug,
		FormFactor: state.FormFactor, Width: state.Width, UHeight: state.UHeight,
		StartingUnit: state.StartingUnit, DescUnits: state.DescUnits,
		Description: state.Description, Comments: state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted RackType violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	rackType := rackTypeFromNormalized(normalized, state.Created, state.LastUpdated)
	rackType.id = state.ID
	return rackType, nil
}

func rackTypeFromNormalized(
	values normalizedRackTypeValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *RackType {
	return &RackType{
		manufacturer: values.manufacturer, model: values.model, slug: values.slug,
		formFactor: values.formFactor, width: values.width, uHeight: values.uHeight,
		startingUnit: values.startingUnit, descUnits: values.descUnits,
		description: values.description, comments: values.comments,
		created: created, lastUpdated: lastUpdated,
	}
}

func (rackType *RackType) AssignID(id shared.ID) error {
	if rackType == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid RackType ID.")
	}
	if rackType.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned RackType ID.")
	}
	rackType.id = id
	return nil
}

func (rackType *RackType) Replace(values RackTypeValues, now shared.Timestamp) error {
	if rackType == nil || now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace RackType with invalid state or time.")
	}
	normalized, violations := validateRackTypeValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	created := rackType.created
	id := rackType.id
	*rackType = *rackTypeFromNormalized(normalized, created, now)
	rackType.id = id
	return nil
}

func (rackType *RackType) ApplyPatch(patch RackTypePatch, now shared.Timestamp) error {
	if rackType == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil RackType.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		})
	}
	return rackType.Replace(rackType.valuesWithPatch(patch), now)
}

// ValidatePatch checks the state a patch would produce without mutating the
// aggregate. Empty patches are valid previews; ApplyPatch retains ownership of
// the public update-mask requirement.
func (rackType *RackType) ValidatePatch(patch RackTypePatch) error {
	if rackType == nil {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot validate a patch for a nil RackType.",
		)
	}
	_, violations := validateRackTypeValues(rackType.valuesWithPatch(patch))
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	return nil
}

func (rackType RackType) valuesWithPatch(patch RackTypePatch) RackTypeValues {
	values := rackType.Values()
	if patch.Manufacturer != nil {
		values.Manufacturer = *patch.Manufacturer
	}
	setString(&values.Model, patch.Model)
	setString(&values.Slug, patch.Slug)
	setString(&values.FormFactor, patch.FormFactor)
	setUint32(&values.Width, patch.Width)
	setUint32(&values.UHeight, patch.UHeight)
	setUint32(&values.StartingUnit, patch.StartingUnit)
	if patch.DescUnits != nil {
		values.DescUnits = *patch.DescUnits
	}
	setString(&values.Description, patch.Description)
	setString(&values.Comments, patch.Comments)
	return values
}

func (rackType RackType) ID() shared.ID                       { return rackType.id }
func (rackType RackType) Manufacturer() ManufacturerReference { return rackType.manufacturer }
func (rackType RackType) Model() string                       { return rackType.model }
func (rackType RackType) Slug() shared.Slug                   { return rackType.slug }
func (rackType RackType) FormFactor() RackFormFactor          { return rackType.formFactor }
func (rackType RackType) Width() RackWidth                    { return rackType.width }
func (rackType RackType) UHeight() uint32                     { return rackType.uHeight }
func (rackType RackType) StartingUnit() uint32                { return rackType.startingUnit }
func (rackType RackType) DescUnits() bool                     { return rackType.descUnits }
func (rackType RackType) Description() string                 { return rackType.description }
func (rackType RackType) Comments() string                    { return rackType.comments }
func (rackType RackType) Created() shared.Timestamp           { return rackType.created }
func (rackType RackType) LastUpdated() shared.Timestamp       { return rackType.lastUpdated }
func (rackType RackType) Display() string                     { return rackType.model }
func (rackType RackType) FullName() string {
	return rackType.manufacturer.Display() + " " + rackType.model
}
func (rackType RackType) PhysicalAttributes() RackPhysicalAttributes {
	return RackPhysicalAttributes{
		FormFactor: rackType.formFactor, Width: rackType.width, UHeight: rackType.uHeight,
		StartingUnit: rackType.startingUnit, DescUnits: rackType.descUnits,
	}
}

func (rackType RackType) Values() RackTypeValues {
	return RackTypeValues{
		Manufacturer: rackType.manufacturer, Model: rackType.model, Slug: rackType.slug.String(),
		FormFactor: rackType.formFactor.String(), Width: rackType.width.Uint32(), UHeight: rackType.uHeight,
		StartingUnit: rackType.startingUnit, DescUnits: rackType.descUnits,
		Description: rackType.description, Comments: rackType.comments,
	}
}

func (rackType RackType) State() RackTypeState {
	values := rackType.Values()
	return RackTypeState{
		ID: rackType.id, Manufacturer: values.Manufacturer, Model: values.Model, Slug: values.Slug,
		FormFactor: values.FormFactor, Width: values.Width, UHeight: values.UHeight,
		StartingUnit: values.StartingUnit, DescUnits: values.DescUnits,
		Description: values.Description, Comments: values.Comments,
		Created: rackType.created, LastUpdated: rackType.lastUpdated,
	}
}

func (rackType RackType) Snapshot() RackTypeSnapshot {
	return RackTypeSnapshot{
		ManufacturerID: rackType.manufacturer.ID(), Model: rackType.model, Slug: rackType.slug.String(),
		FormFactor: rackType.formFactor.String(), Width: rackType.width.Uint32(), UHeight: rackType.uHeight,
		StartingUnit: rackType.startingUnit, DescUnits: rackType.descUnits,
		Description: rackType.description, Comments: rackType.comments,
	}
}

type normalizedRackTypeValues struct {
	manufacturer ManufacturerReference
	model        string
	slug         shared.Slug
	formFactor   RackFormFactor
	width        RackWidth
	uHeight      uint32
	startingUnit uint32
	descUnits    bool
	description  string
	comments     string
}

func validateRackTypeValues(values RackTypeValues) (normalizedRackTypeValues, []shared.FieldViolation) {
	values.Model = strings.TrimSpace(values.Model)
	values.Slug = strings.TrimSpace(values.Slug)
	values.Description = strings.TrimSpace(values.Description)
	values.Comments = strings.TrimSpace(values.Comments)
	var violations []shared.FieldViolation
	if !values.Manufacturer.Valid() {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	validateRequiredLength(&violations, "model", values.Model, RackTypeModelMaxLength)
	validateOptionalLength(&violations, "description", values.Description, RackTypeDescriptionMaxLength)
	slug, err := shared.ParseSlug(values.Slug, RackTypeSlugMaxLength)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	formFactor, validFactor := ParseRackFormFactor(values.FormFactor)
	if values.FormFactor == "" {
		violations = append(violations, shared.FieldViolation{
			Field: "form_factor", Reason: "required", Description: "This field may not be blank.",
		})
	} else if !validFactor {
		violations = append(violations, shared.FieldViolation{
			Field: "form_factor", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
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
			Field: "starting_unit", Reason: "range", Description: "Ensure this value is greater than or equal to 1.",
		})
	}
	return normalizedRackTypeValues{
		manufacturer: values.Manufacturer, model: values.Model, slug: slug,
		formFactor: formFactor, width: width, uHeight: values.UHeight,
		startingUnit: values.StartingUnit, descUnits: values.DescUnits,
		description: values.Description, comments: values.Comments,
	}, violations
}

func setUint32(destination *uint32, source *uint32) {
	if source != nil {
		*destination = *source
	}
}
