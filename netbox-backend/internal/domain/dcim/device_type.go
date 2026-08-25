package dcim

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"netbox-go/internal/domain/shared"
)

const (
	DeviceTypeObjectType           = "dcim.devicetype"
	DeviceTypeModelMaxLength       = 100
	DeviceTypeSlugMaxLength        = 100
	DeviceTypePartNumberMaxLength  = 50
	DeviceTypeDescriptionMaxLength = 200
	DeviceTypeDefaultHeight        = "1"
)

// DeviceHeight stores exact half-units so validation and rack occupancy never
// depend on binary floating-point comparisons.
type DeviceHeight struct {
	halfUnits uint16
}

func ParseDeviceHeight(value string) (DeviceHeight, error) {
	parsed, ok := parseDeviceHeightDecimal(value)
	if !ok {
		return DeviceHeight{}, deviceHeightViolation("Must be a non-negative number.")
	}
	if description := parsed.precisionViolation(); description != "" {
		return DeviceHeight{}, deviceHeightViolation(description)
	}
	if parsed.negative && parsed.digits != "0" {
		return DeviceHeight{}, deviceHeightViolation("Must be a non-negative number.")
	}
	halfUnits, ok := parsed.halfUnits()
	if !ok {
		return DeviceHeight{}, deviceHeightViolation("Device height must be a multiple of 0.5U.")
	}
	return DeviceHeightFromHalfUnits(halfUnits)
}

type deviceHeightDecimal struct {
	digits   string
	exponent int
	negative bool
}

func parseDeviceHeightDecimal(value string) (deviceHeightDecimal, bool) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > 1000 {
		return deviceHeightDecimal{}, false
	}
	value = strings.ReplaceAll(value, "_", "")
	if value == "" {
		return deviceHeightDecimal{}, false
	}

	negative := false
	switch value[0] {
	case '+':
		value = value[1:]
	case '-':
		negative = true
		value = value[1:]
	}
	if value == "" {
		return deviceHeightDecimal{}, false
	}

	coefficient := value
	exponent := 0
	if exponentIndex := strings.IndexAny(value, "eE"); exponentIndex >= 0 {
		coefficient = value[:exponentIndex]
		exponentText := value[exponentIndex+1:]
		if strings.ContainsAny(exponentText, "eE") {
			return deviceHeightDecimal{}, false
		}
		var ok bool
		exponent, ok = parseDeviceHeightExponent(exponentText)
		if !ok {
			return deviceHeightDecimal{}, false
		}
	}

	digits, decimalPlaces, ok := parseDeviceHeightCoefficient(coefficient)
	if !ok {
		return deviceHeightDecimal{}, false
	}
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		digits = "0"
	}
	return deviceHeightDecimal{
		digits: digits, exponent: exponent - decimalPlaces, negative: negative,
	}, true
}

func parseDeviceHeightExponent(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	sign := 1
	switch value[0] {
	case '+':
		value = value[1:]
	case '-':
		sign = -1
		value = value[1:]
	}
	if value == "" {
		return 0, false
	}

	exponent := 0
	for _, character := range value {
		digit, ok := deviceHeightDecimalDigit(character)
		if !ok {
			return 0, false
		}
		if exponent < 1000 {
			exponent = exponent*10 + int(digit)
			if exponent > 1000 {
				exponent = 1000
			}
		}
	}
	return sign * exponent, true
}

func parseDeviceHeightCoefficient(value string) (string, int, bool) {
	digits := make([]byte, 0, len(value))
	decimalPlaces := 0
	decimalPointSeen := false
	for _, character := range value {
		if character == '.' {
			if decimalPointSeen {
				return "", 0, false
			}
			decimalPointSeen = true
			continue
		}
		digit, ok := deviceHeightDecimalDigit(character)
		if !ok {
			return "", 0, false
		}
		digits = append(digits, '0'+digit)
		if decimalPointSeen {
			decimalPlaces++
		}
	}
	if len(digits) == 0 {
		return "", 0, false
	}
	return string(digits), decimalPlaces, true
}

func deviceHeightDecimalDigit(character rune) (byte, bool) {
	for _, decimalRange := range unicode.Digit.R16 {
		if character < rune(decimalRange.Lo) {
			break
		}
		if character <= rune(decimalRange.Hi) &&
			(character-rune(decimalRange.Lo))%rune(decimalRange.Stride) == 0 {
			index := (character - rune(decimalRange.Lo)) / rune(decimalRange.Stride)
			return byte(index % 10), true
		}
	}
	for _, decimalRange := range unicode.Digit.R32 {
		if character < rune(decimalRange.Lo) {
			break
		}
		if character <= rune(decimalRange.Hi) &&
			(character-rune(decimalRange.Lo))%rune(decimalRange.Stride) == 0 {
			index := (character - rune(decimalRange.Lo)) / rune(decimalRange.Stride)
			return byte(index % 10), true
		}
	}
	return 0, false
}

func (value deviceHeightDecimal) precisionViolation() string {
	totalDigits := len(value.digits)
	var wholeDigits int
	decimalPlaces := 0
	if value.exponent >= 0 {
		totalDigits += value.exponent
		wholeDigits = totalDigits
	} else if totalDigits > -value.exponent {
		wholeDigits = totalDigits + value.exponent
		decimalPlaces = -value.exponent
	} else {
		totalDigits = -value.exponent
		wholeDigits = 0
		decimalPlaces = totalDigits
	}

	switch {
	case totalDigits > 4:
		return "Ensure that there are no more than 4 digits in total."
	case decimalPlaces > 1:
		return "Ensure that there are no more than 1 decimal places."
	case wholeDigits > 3:
		return "Ensure that there are no more than 3 digits before the decimal point."
	default:
		return ""
	}
}

func (value deviceHeightDecimal) halfUnits() (uint16, bool) {
	coefficient, err := strconv.ParseUint(value.digits, 10, 16)
	if err != nil {
		return 0, false
	}
	if value.exponent == -1 {
		if coefficient%5 != 0 {
			return 0, false
		}
		return uint16(coefficient / 5), true
	}
	for range value.exponent {
		coefficient *= 10
	}
	return uint16(coefficient * 2), true
}

func DeviceHeightFromHalfUnits(halfUnits uint16) (DeviceHeight, error) {
	if halfUnits > 1999 {
		return DeviceHeight{}, deviceHeightViolation("Device height must not exceed 999.5U.")
	}
	return DeviceHeight{halfUnits: halfUnits}, nil
}

func deviceHeightViolation(description string) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "u_height", Reason: "invalid", Description: description,
	})
}

func (height DeviceHeight) HalfUnits() uint16 { return height.halfUnits }
func (height DeviceHeight) Float64() float64  { return float64(height.halfUnits) / 2 }
func (height DeviceHeight) String() string {
	whole := height.halfUnits / 2
	if height.halfUnits%2 == 0 {
		return strconv.FormatUint(uint64(whole), 10)
	}
	return fmt.Sprintf("%d.5", whole)
}

// DeviceAirflow is the exact pinned DeviceAirflowChoices value. The blank
// value is valid and distinct in storage from a null value.
type DeviceAirflow string

const (
	DeviceAirflowFrontToRear DeviceAirflow = "front-to-rear"
	DeviceAirflowRearToFront DeviceAirflow = "rear-to-front"
	DeviceAirflowLeftToRight DeviceAirflow = "left-to-right"
	DeviceAirflowRightToLeft DeviceAirflow = "right-to-left"
	DeviceAirflowSideToRear  DeviceAirflow = "side-to-rear"
	DeviceAirflowRearToSide  DeviceAirflow = "rear-to-side"
	DeviceAirflowBottomToTop DeviceAirflow = "bottom-to-top"
	DeviceAirflowTopToBottom DeviceAirflow = "top-to-bottom"
	DeviceAirflowPassive     DeviceAirflow = "passive"
	DeviceAirflowMixed       DeviceAirflow = "mixed"
)

func ParseDeviceAirflow(value string) (DeviceAirflow, bool) {
	airflow := DeviceAirflow(value)
	switch airflow {
	case "",
		DeviceAirflowFrontToRear,
		DeviceAirflowRearToFront,
		DeviceAirflowLeftToRight,
		DeviceAirflowRightToLeft,
		DeviceAirflowSideToRear,
		DeviceAirflowRearToSide,
		DeviceAirflowBottomToTop,
		DeviceAirflowTopToBottom,
		DeviceAirflowPassive,
		DeviceAirflowMixed:
		return airflow, true
	default:
		return "", false
	}
}

func (airflow DeviceAirflow) String() string { return string(airflow) }

type NullableDeviceAirflow struct {
	value DeviceAirflow
	valid bool
}

func NullDeviceAirflow() NullableDeviceAirflow { return NullableDeviceAirflow{} }
func NonNullDeviceAirflow(value DeviceAirflow) NullableDeviceAirflow {
	return NullableDeviceAirflow{value: value, valid: true}
}
func (nullable NullableDeviceAirflow) Get() (DeviceAirflow, bool) {
	return nullable.value, nullable.valid
}
func (nullable NullableDeviceAirflow) IsNull() bool { return !nullable.valid }

type DeviceTypeValues struct {
	Manufacturer           ManufacturerReference
	Model                  string
	Slug                   string
	PartNumber             string
	UHeight                string
	ExcludeFromUtilization bool
	IsFullDepth            bool
	Airflow                NullableDeviceAirflow
	Description            string
	Comments               string
}

type DeviceTypePatch struct {
	Manufacturer           *ManufacturerReference
	Model                  *string
	Slug                   *string
	PartNumber             *string
	UHeight                *string
	ExcludeFromUtilization *bool
	IsFullDepth            *bool
	Airflow                *NullableDeviceAirflow
	Description            *string
	Comments               *string
}

func (patch DeviceTypePatch) Empty() bool {
	return patch.Manufacturer == nil && patch.Model == nil && patch.Slug == nil &&
		patch.PartNumber == nil && patch.UHeight == nil &&
		patch.ExcludeFromUtilization == nil && patch.IsFullDepth == nil &&
		patch.Airflow == nil && patch.Description == nil && patch.Comments == nil
}

type DeviceTypeState struct {
	ID                     shared.ID
	Manufacturer           ManufacturerReference
	Model                  string
	Slug                   string
	PartNumber             string
	UHeight                string
	ExcludeFromUtilization bool
	IsFullDepth            bool
	Airflow                NullableDeviceAirflow
	Description            string
	Comments               string
	Created                shared.Timestamp
	LastUpdated            shared.Timestamp
	DeviceCount            uint64
	InterfaceTemplateCount uint64
}

type DeviceTypeSnapshot struct {
	ManufacturerID         shared.ID
	Model                  string
	Slug                   string
	PartNumber             string
	UHeight                string
	ExcludeFromUtilization bool
	IsFullDepth            bool
	Airflow                *string
	Description            string
	Comments               string
}

func (DeviceTypeSnapshot) ObjectType() string { return DeviceTypeObjectType }
func (snapshot DeviceTypeSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	snapshot.Airflow = cloneString(snapshot.Airflow)
	return snapshot
}

type DeviceType struct {
	id                     shared.ID
	manufacturer           ManufacturerReference
	model                  string
	slug                   shared.Slug
	partNumber             string
	uHeight                DeviceHeight
	excludeFromUtilization bool
	isFullDepth            bool
	airflow                NullableDeviceAirflow
	description            string
	comments               string
	created                shared.Timestamp
	lastUpdated            shared.Timestamp
	deviceCount            uint64
	interfaceTemplateCount uint64
}

func NewDeviceType(values DeviceTypeValues, now shared.Timestamp) (*DeviceType, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateDeviceTypeValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return deviceTypeFromNormalized(normalized, now, now), nil
}

func RestoreDeviceType(state DeviceTypeState) (*DeviceType, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot restore invalid DeviceType identity or timestamps.",
		)
	}
	normalized, violations := validateDeviceTypeValues(DeviceTypeValues{
		Manufacturer: state.Manufacturer, Model: state.Model, Slug: state.Slug,
		PartNumber: state.PartNumber, UHeight: state.UHeight,
		ExcludeFromUtilization: state.ExcludeFromUtilization,
		IsFullDepth:            state.IsFullDepth, Airflow: state.Airflow,
		Description: state.Description, Comments: state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted DeviceType violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	deviceType := deviceTypeFromNormalized(normalized, state.Created, state.LastUpdated)
	deviceType.id = state.ID
	deviceType.deviceCount = state.DeviceCount
	deviceType.interfaceTemplateCount = state.InterfaceTemplateCount
	return deviceType, nil
}

func deviceTypeFromNormalized(
	values normalizedDeviceTypeValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *DeviceType {
	return &DeviceType{
		manufacturer: values.manufacturer, model: values.model, slug: values.slug,
		partNumber: values.partNumber, uHeight: values.uHeight,
		excludeFromUtilization: values.excludeFromUtilization,
		isFullDepth:            values.isFullDepth, airflow: values.airflow,
		description: values.description, comments: values.comments,
		created: created, lastUpdated: lastUpdated,
	}
}

func (deviceType *DeviceType) AssignID(id shared.ID) error {
	if deviceType == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid DeviceType ID.")
	}
	if deviceType.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned DeviceType ID.")
	}
	deviceType.id = id
	return nil
}

func (deviceType *DeviceType) Replace(values DeviceTypeValues, now shared.Timestamp) error {
	if deviceType == nil || now.IsZero() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot replace DeviceType with invalid state or time.",
		)
	}
	normalized, violations := validateDeviceTypeValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	created, id := deviceType.created, deviceType.id
	deviceCount, templateCount := deviceType.deviceCount, deviceType.interfaceTemplateCount
	*deviceType = *deviceTypeFromNormalized(normalized, created, now)
	deviceType.id = id
	deviceType.deviceCount = deviceCount
	deviceType.interfaceTemplateCount = templateCount
	return nil
}

func (deviceType *DeviceType) ApplyPatch(patch DeviceTypePatch, now shared.Timestamp) error {
	if deviceType == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil DeviceType.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return deviceType.Replace(deviceType.valuesWithPatch(patch), now)
}

// ValidatePatch checks the state a patch would produce without mutating the
// aggregate. Empty patches are valid previews; ApplyPatch retains ownership of
// the public update-mask requirement.
func (deviceType *DeviceType) ValidatePatch(patch DeviceTypePatch) error {
	if deviceType == nil {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot validate a patch for a nil DeviceType.",
		)
	}
	_, violations := validateDeviceTypeValues(deviceType.valuesWithPatch(patch))
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	return nil
}

func (deviceType DeviceType) valuesWithPatch(patch DeviceTypePatch) DeviceTypeValues {
	values := deviceType.Values()
	if patch.Manufacturer != nil {
		values.Manufacturer = *patch.Manufacturer
	}
	setString(&values.Model, patch.Model)
	setString(&values.Slug, patch.Slug)
	setString(&values.PartNumber, patch.PartNumber)
	setString(&values.UHeight, patch.UHeight)
	if patch.ExcludeFromUtilization != nil {
		values.ExcludeFromUtilization = *patch.ExcludeFromUtilization
	}
	if patch.IsFullDepth != nil {
		values.IsFullDepth = *patch.IsFullDepth
	}
	if patch.Airflow != nil {
		values.Airflow = *patch.Airflow
	}
	setString(&values.Description, patch.Description)
	setString(&values.Comments, patch.Comments)
	return values
}

func (deviceType DeviceType) ID() shared.ID                       { return deviceType.id }
func (deviceType DeviceType) Manufacturer() ManufacturerReference { return deviceType.manufacturer }
func (deviceType DeviceType) Model() string                       { return deviceType.model }
func (deviceType DeviceType) Slug() shared.Slug                   { return deviceType.slug }
func (deviceType DeviceType) PartNumber() string                  { return deviceType.partNumber }
func (deviceType DeviceType) UHeight() DeviceHeight               { return deviceType.uHeight }
func (deviceType DeviceType) ExcludeFromUtilization() bool {
	return deviceType.excludeFromUtilization
}
func (deviceType DeviceType) IsFullDepth() bool              { return deviceType.isFullDepth }
func (deviceType DeviceType) Airflow() NullableDeviceAirflow { return deviceType.airflow }
func (deviceType DeviceType) Description() string            { return deviceType.description }
func (deviceType DeviceType) Comments() string               { return deviceType.comments }
func (deviceType DeviceType) Created() shared.Timestamp      { return deviceType.created }
func (deviceType DeviceType) LastUpdated() shared.Timestamp  { return deviceType.lastUpdated }
func (deviceType DeviceType) DeviceCount() uint64            { return deviceType.deviceCount }
func (deviceType DeviceType) InterfaceTemplateCount() uint64 {
	return deviceType.interfaceTemplateCount
}
func (deviceType DeviceType) Display() string { return deviceType.model }
func (deviceType DeviceType) FullName() string {
	return deviceType.manufacturer.Display() + " " + deviceType.model
}

func (deviceType DeviceType) Values() DeviceTypeValues {
	return DeviceTypeValues{
		Manufacturer: deviceType.manufacturer, Model: deviceType.model,
		Slug: deviceType.slug.String(), PartNumber: deviceType.partNumber,
		UHeight:                deviceType.uHeight.String(),
		ExcludeFromUtilization: deviceType.excludeFromUtilization,
		IsFullDepth:            deviceType.isFullDepth, Airflow: deviceType.airflow,
		Description: deviceType.description, Comments: deviceType.comments,
	}
}

func (deviceType DeviceType) State() DeviceTypeState {
	values := deviceType.Values()
	return DeviceTypeState{
		ID: deviceType.id, Manufacturer: values.Manufacturer, Model: values.Model,
		Slug: values.Slug, PartNumber: values.PartNumber, UHeight: values.UHeight,
		ExcludeFromUtilization: values.ExcludeFromUtilization,
		IsFullDepth:            values.IsFullDepth, Airflow: values.Airflow,
		Description: values.Description, Comments: values.Comments,
		Created: deviceType.created, LastUpdated: deviceType.lastUpdated,
		DeviceCount:            deviceType.deviceCount,
		InterfaceTemplateCount: deviceType.interfaceTemplateCount,
	}
}

func (deviceType DeviceType) Snapshot() DeviceTypeSnapshot {
	var airflow *string
	if value, present := deviceType.airflow.Get(); present {
		text := value.String()
		airflow = &text
	}
	return DeviceTypeSnapshot{
		ManufacturerID: deviceType.manufacturer.ID(), Model: deviceType.model,
		Slug: deviceType.slug.String(), PartNumber: deviceType.partNumber,
		UHeight:                deviceType.uHeight.String(),
		ExcludeFromUtilization: deviceType.excludeFromUtilization,
		IsFullDepth:            deviceType.isFullDepth, Airflow: airflow,
		Description: deviceType.description, Comments: deviceType.comments,
	}
}

type normalizedDeviceTypeValues struct {
	manufacturer           ManufacturerReference
	model                  string
	slug                   shared.Slug
	partNumber             string
	uHeight                DeviceHeight
	excludeFromUtilization bool
	isFullDepth            bool
	airflow                NullableDeviceAirflow
	description            string
	comments               string
}

func validateDeviceTypeValues(
	values DeviceTypeValues,
) (normalizedDeviceTypeValues, []shared.FieldViolation) {
	values.Model = strings.TrimSpace(values.Model)
	values.Slug = strings.TrimSpace(values.Slug)
	values.PartNumber = strings.TrimSpace(values.PartNumber)
	values.UHeight = strings.TrimSpace(values.UHeight)
	values.Description = strings.TrimSpace(values.Description)
	values.Comments = strings.TrimSpace(values.Comments)

	var violations []shared.FieldViolation
	if !values.Manufacturer.Valid() {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	validateRequiredLength(&violations, "model", values.Model, DeviceTypeModelMaxLength)
	slug, err := shared.ParseSlug(values.Slug, DeviceTypeSlugMaxLength)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	validateOptionalLength(&violations, "part_number", values.PartNumber, DeviceTypePartNumberMaxLength)
	height, err := ParseDeviceHeight(values.UHeight)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	if airflow, present := values.Airflow.Get(); present {
		if _, valid := ParseDeviceAirflow(airflow.String()); !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "airflow", Reason: "invalid_choice",
				Description: "Unsupported airflow direction.",
			})
		}
	}
	validateOptionalLength(&violations, "description", values.Description, DeviceTypeDescriptionMaxLength)
	return normalizedDeviceTypeValues{
		manufacturer: values.Manufacturer, model: values.Model, slug: slug,
		partNumber: values.PartNumber, uHeight: height,
		excludeFromUtilization: values.ExcludeFromUtilization,
		isFullDepth:            values.IsFullDepth, airflow: values.Airflow,
		description: values.Description, comments: values.Comments,
	}, violations
}
