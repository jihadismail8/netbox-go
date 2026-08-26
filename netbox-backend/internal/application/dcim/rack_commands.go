package dcim

import (
	"sort"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateRackCommand struct {
	Site         Field[shared.ID]
	Name         Field[string]
	FacilityID   Field[string]
	RackType     Field[shared.ID]
	Status       Field[string]
	Role         Field[shared.ID]
	Serial       Field[string]
	AssetTag     Field[string]
	FormFactor   Field[string]
	Width        Field[uint32]
	UHeight      Field[uint32]
	StartingUnit Field[uint32]
	DescUnits    Field[bool]
	Airflow      Field[string]
	Description  Field[string]
	Comments     Field[string]
}

type ReplaceRackCommand struct {
	ID shared.ID
	CreateRackCommand
}

type UpdateRackCommand struct {
	ID           shared.ID
	Site         Field[shared.ID]
	Name         Field[string]
	FacilityID   Field[string]
	RackType     Field[shared.ID]
	Status       Field[string]
	Role         Field[shared.ID]
	Serial       Field[string]
	AssetTag     Field[string]
	FormFactor   Field[string]
	Width        Field[uint32]
	UHeight      Field[uint32]
	StartingUnit Field[uint32]
	DescUnits    Field[bool]
	Airflow      Field[string]
	Description  Field[string]
	Comments     Field[string]
}

type DeleteRackCommand struct{ ID shared.ID }

type rackCommandValues struct {
	siteID       shared.ID
	name         string
	facilityID   dcimdomain.RackNullable[string]
	rackTypeID   dcimdomain.RackNullable[shared.ID]
	status       string
	roleID       dcimdomain.RackNullable[shared.ID]
	serial       string
	assetTag     dcimdomain.RackNullable[string]
	formFactor   dcimdomain.RackNullable[string]
	width        uint32
	uHeight      uint32
	startingUnit uint32
	descUnits    bool
	airflow      dcimdomain.RackNullable[string]
	description  string
	comments     string
}

func (command CreateRackCommand) values() (rackCommandValues, error) {
	var violations []shared.FieldViolation
	values := rackCommandValues{
		siteID:       fullFieldValue(&violations, "site", command.Site, shared.ID(0), true),
		name:         valueForFullMutation(&violations, "name", command.Name, "", true),
		facilityID:   fullRackNullable(command.FacilityID),
		rackTypeID:   fullRackNullable(command.RackType),
		status:       valueForFullMutation(&violations, "status", command.Status, dcimdomain.RackStatusActive.String(), false),
		roleID:       fullRackNullable(command.Role),
		serial:       valueForFullMutation(&violations, "serial", command.Serial, "", false),
		assetTag:     fullRackNullable(command.AssetTag),
		formFactor:   fullRackNullable(command.FormFactor),
		width:        fullFieldValue(&violations, "width", command.Width, dcimdomain.RackDefaultWidth, false),
		uHeight:      fullFieldValue(&violations, "u_height", command.UHeight, dcimdomain.RackDefaultUHeight, false),
		startingUnit: fullFieldValue(&violations, "starting_unit", command.StartingUnit, dcimdomain.RackDefaultStartingUnit, false),
		descUnits:    fullFieldValue(&violations, "desc_units", command.DescUnits, false, false),
		airflow:      fullRackAirflow(&violations, command.Airflow),
		description:  valueForFullMutation(&violations, "description", command.Description, "", false),
		comments:     valueForFullMutation(&violations, "comments", command.Comments, "", false),
	}
	if command.Site.State() == FieldPresent {
		validateRackCommandID(&violations, "site", values.siteID)
	}
	validateNullableRackCommandID(&violations, "rack_type", values.rackTypeID)
	validateNullableRackCommandID(&violations, "role", values.roleID)
	if len(violations) > 0 {
		return values, newRackValidationError(violations)
	}
	return values, nil
}

type rackCommandPatch struct {
	siteID       *shared.ID
	name         *string
	facilityID   *dcimdomain.RackNullable[string]
	rackTypeID   *dcimdomain.RackNullable[shared.ID]
	status       *string
	roleID       *dcimdomain.RackNullable[shared.ID]
	serial       *string
	assetTag     *dcimdomain.RackNullable[string]
	formFactor   *dcimdomain.RackNullable[string]
	width        *uint32
	uHeight      *uint32
	startingUnit *uint32
	descUnits    *bool
	airflow      *dcimdomain.RackNullable[string]
	description  *string
	comments     *string
}

func (patch rackCommandPatch) empty() bool {
	return patch.siteID == nil && patch.name == nil && patch.facilityID == nil &&
		patch.rackTypeID == nil && patch.status == nil && patch.roleID == nil &&
		patch.serial == nil && patch.assetTag == nil && patch.formFactor == nil &&
		patch.width == nil && patch.uHeight == nil && patch.startingUnit == nil &&
		patch.descUnits == nil && patch.airflow == nil && patch.description == nil &&
		patch.comments == nil
}

func (command UpdateRackCommand) patch() (rackCommandPatch, error) {
	return buildRackPatch(CreateRackCommand{
		Site: command.Site, Name: command.Name, FacilityID: command.FacilityID,
		RackType: command.RackType, Status: command.Status, Role: command.Role,
		Serial: command.Serial, AssetTag: command.AssetTag, FormFactor: command.FormFactor,
		Width: command.Width, UHeight: command.UHeight, StartingUnit: command.StartingUnit,
		DescUnits: command.DescUnits, Airflow: command.Airflow,
		Description: command.Description, Comments: command.Comments,
	}, false)
}

func (command ReplaceRackCommand) patch() (rackCommandPatch, error) {
	return buildRackPatch(command.CreateRackCommand, true)
}

func buildRackPatch(command CreateRackCommand, replace bool) (rackCommandPatch, error) {
	var violations []shared.FieldViolation
	if replace {
		requireRackField(&violations, "site", command.Site)
		requireRackField(&violations, "name", command.Name)
	}
	patch := rackCommandPatch{
		siteID:       patchFieldValue(&violations, "site", command.Site),
		name:         patchValue(&violations, "name", command.Name),
		facilityID:   patchRackNullable(command.FacilityID),
		rackTypeID:   patchRackNullable(command.RackType),
		status:       patchValue(&violations, "status", command.Status),
		roleID:       patchRackNullable(command.Role),
		serial:       patchValue(&violations, "serial", command.Serial),
		assetTag:     patchRackNullable(command.AssetTag),
		formFactor:   patchRackNullable(command.FormFactor),
		width:        patchFieldValue(&violations, "width", command.Width),
		uHeight:      patchFieldValue(&violations, "u_height", command.UHeight),
		startingUnit: patchFieldValue(&violations, "starting_unit", command.StartingUnit),
		descUnits:    patchFieldValue(&violations, "desc_units", command.DescUnits),
		airflow:      patchRackAirflow(&violations, command.Airflow),
		description:  patchValue(&violations, "description", command.Description),
		comments:     patchValue(&violations, "comments", command.Comments),
	}
	if replace && command.FacilityID.State() == FieldOmitted {
		value := dcimdomain.NullRackValue[string]()
		patch.facilityID = &value
	}
	if replace && command.RackType.State() == FieldOmitted {
		value := dcimdomain.NullRackValue[shared.ID]()
		patch.rackTypeID = &value
	}
	if patch.siteID != nil {
		validateRackCommandID(&violations, "site", *patch.siteID)
	}
	if patch.rackTypeID != nil {
		validateNullableRackCommandID(&violations, "rack_type", *patch.rackTypeID)
	}
	if patch.roleID != nil {
		validateNullableRackCommandID(&violations, "role", *patch.roleID)
	}
	if len(violations) > 0 {
		return patch, newRackValidationError(violations)
	}
	if patch.empty() {
		return patch, newRackValidationError([]shared.FieldViolation{{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		}})
	}
	return patch, nil
}

func requireRackField[T any](violations *[]shared.FieldViolation, name string, field Field[T]) {
	if field.State() != FieldOmitted {
		return
	}
	*violations = append(*violations, shared.FieldViolation{
		Field: name, Reason: "required", Description: "This field is required.",
	})
}

func fullRackNullable[T any](field Field[T]) dcimdomain.RackNullable[T] {
	if value, present := field.Get(); present {
		return dcimdomain.NonNullRackValue(value)
	}
	return dcimdomain.NullRackValue[T]()
}

func patchRackNullable[T any](field Field[T]) *dcimdomain.RackNullable[T] {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		nullable := dcimdomain.NonNullRackValue(value)
		return &nullable
	case FieldNull:
		nullable := dcimdomain.NullRackValue[T]()
		return &nullable
	default:
		return nil
	}
}

func fullRackAirflow(
	violations *[]shared.FieldViolation,
	field Field[string],
) dcimdomain.RackNullable[string] {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return dcimdomain.NonNullRackValue(value)
	case FieldNull:
		*violations = append(*violations, rackNullViolation("airflow"))
	}
	return dcimdomain.NullRackValue[string]()
}

func patchRackAirflow(
	violations *[]shared.FieldViolation,
	field Field[string],
) *dcimdomain.RackNullable[string] {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		nullable := dcimdomain.NonNullRackValue(value)
		return &nullable
	case FieldNull:
		*violations = append(*violations, rackNullViolation("airflow"))
	}
	return nil
}

func rackNullViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "null", Description: "This field may not be null.",
	}
}

func validateRackCommandID(
	violations *[]shared.FieldViolation,
	field string,
	id shared.ID,
) {
	if id.IsValid() {
		return
	}
	*violations = append(*violations, shared.FieldViolation{
		Field: field, Reason: "invalid_choice", Description: "Select a valid choice.",
	})
}

func validateNullableRackCommandID(
	violations *[]shared.FieldViolation,
	field string,
	value dcimdomain.RackNullable[shared.ID],
) {
	id, present := value.Get()
	if present {
		validateRackCommandID(violations, field, id)
	}
}

var rackValidationFieldOrder = map[string]int{
	"site": 0, "name": 1, "facility_id": 2, "rack_type": 3,
	"status": 4, "role": 5, "serial": 6, "asset_tag": 7,
	"form_factor": 8, "width": 9, "u_height": 10, "starting_unit": 11,
	"desc_units": 12, "airflow": 13, "description": 14, "comments": 15,
	"update_mask": 16,
}

func newRackValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := rackValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := rackValidationFieldOrder[ordered[right].Field]
		switch {
		case leftKnown && rightKnown:
			return leftOrder < rightOrder
		case leftKnown:
			return true
		case rightKnown:
			return false
		default:
			return false
		}
	})
	return shared.NewValidationError(ordered...)
}

func mergeRackMutationErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil && !shared.HasReason(err, shared.ErrorReasonValidation) {
			return err
		}
	}

	seenFields := make(map[string]struct{})
	var merged []shared.FieldViolation
	for _, err := range errs {
		for _, violation := range shared.ViolationsOf(err) {
			if _, duplicate := seenFields[violation.Field]; duplicate {
				continue
			}
			seenFields[violation.Field] = struct{}{}
			merged = append(merged, violation)
		}
	}
	if len(merged) > 0 {
		return newRackValidationError(merged)
	}
	return nil
}
