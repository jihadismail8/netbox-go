package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	ManufacturerObjectType           = "dcim.manufacturer"
	ManufacturerNameMaxLength        = 100
	ManufacturerSlugMaxLength        = 100
	ManufacturerDescriptionMaxLength = 200
)

// ManufacturerValues is the complete writable state of a Manufacturer.
type ManufacturerValues struct {
	Name        string
	Slug        string
	Description string
}

// ManufacturerPatch preserves field presence for partial mutations.
type ManufacturerPatch struct {
	Name        *string
	Slug        *string
	Description *string
}

func (patch ManufacturerPatch) Empty() bool {
	return patch.Name == nil && patch.Slug == nil && patch.Description == nil
}

// ManufacturerState is the typed persistence boundary for the aggregate.
type ManufacturerState struct {
	ID              shared.ID
	Name            string
	Slug            string
	Description     string
	Created         shared.Timestamp
	LastUpdated     shared.Timestamp
	DeviceTypeCount uint64
}

// ManufacturerSnapshot is the immutable object-change projection.
type ManufacturerSnapshot struct {
	Name        string
	Slug        string
	Description string
}

func (ManufacturerSnapshot) ObjectType() string { return ManufacturerObjectType }

func (snapshot ManufacturerSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }

// Manufacturer owns the invariants of the first-profile manufacturer model.
type Manufacturer struct {
	id              shared.ID
	name            string
	slug            shared.Slug
	description     string
	created         shared.Timestamp
	lastUpdated     shared.Timestamp
	deviceTypeCount uint64
}

func NewManufacturer(values ManufacturerValues, now shared.Timestamp) (*Manufacturer, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}

	normalized, violations := validateManufacturerValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return &Manufacturer{
		name:        normalized.name,
		slug:        normalized.slug,
		description: normalized.description,
		created:     now,
		lastUpdated: now,
	}, nil
}

func RestoreManufacturer(state ManufacturerState) (*Manufacturer, error) {
	if !state.ID.IsValid() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a Manufacturer with an invalid ID.")
	}
	if state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a Manufacturer with a zero timestamp.")
	}

	normalized, violations := validateManufacturerValues(ManufacturerValues{
		Name: state.Name, Slug: state.Slug, Description: state.Description,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted Manufacturer violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	return &Manufacturer{
		id:              state.ID,
		name:            normalized.name,
		slug:            normalized.slug,
		description:     normalized.description,
		created:         state.Created,
		lastUpdated:     state.LastUpdated,
		deviceTypeCount: state.DeviceTypeCount,
	}, nil
}

func (manufacturer *Manufacturer) AssignID(id shared.ID) error {
	if manufacturer == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid Manufacturer ID.")
	}
	if manufacturer.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned Manufacturer ID.")
	}
	manufacturer.id = id
	return nil
}

func (manufacturer *Manufacturer) Replace(values ManufacturerValues, now shared.Timestamp) error {
	if manufacturer == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace a nil Manufacturer.")
	}
	if now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateManufacturerValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	manufacturer.name = normalized.name
	manufacturer.slug = normalized.slug
	manufacturer.description = normalized.description
	manufacturer.lastUpdated = now
	return nil
}

func (manufacturer *Manufacturer) ApplyPatch(patch ManufacturerPatch, now shared.Timestamp) error {
	if manufacturer == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil Manufacturer.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		})
	}
	values := manufacturer.Values()
	setString(&values.Name, patch.Name)
	setString(&values.Slug, patch.Slug)
	setString(&values.Description, patch.Description)
	return manufacturer.Replace(values, now)
}

func (manufacturer Manufacturer) ID() shared.ID                 { return manufacturer.id }
func (manufacturer Manufacturer) Name() string                  { return manufacturer.name }
func (manufacturer Manufacturer) Slug() shared.Slug             { return manufacturer.slug }
func (manufacturer Manufacturer) Description() string           { return manufacturer.description }
func (manufacturer Manufacturer) Created() shared.Timestamp     { return manufacturer.created }
func (manufacturer Manufacturer) LastUpdated() shared.Timestamp { return manufacturer.lastUpdated }
func (manufacturer Manufacturer) DeviceTypeCount() uint64       { return manufacturer.deviceTypeCount }
func (manufacturer Manufacturer) Display() string               { return manufacturer.name }

func (manufacturer Manufacturer) Values() ManufacturerValues {
	return ManufacturerValues{
		Name: manufacturer.name, Slug: manufacturer.slug.String(), Description: manufacturer.description,
	}
}

func (manufacturer Manufacturer) State() ManufacturerState {
	return ManufacturerState{
		ID: manufacturer.id, Name: manufacturer.name, Slug: manufacturer.slug.String(),
		Description: manufacturer.description, Created: manufacturer.created,
		LastUpdated: manufacturer.lastUpdated, DeviceTypeCount: manufacturer.deviceTypeCount,
	}
}

func (manufacturer Manufacturer) Snapshot() ManufacturerSnapshot {
	return ManufacturerSnapshot{
		Name: manufacturer.name, Slug: manufacturer.slug.String(), Description: manufacturer.description,
	}
}

type normalizedManufacturerValues struct {
	name        string
	slug        shared.Slug
	description string
}

func validateManufacturerValues(values ManufacturerValues) (normalizedManufacturerValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Slug = strings.TrimSpace(values.Slug)
	values.Description = strings.TrimSpace(values.Description)

	var violations []shared.FieldViolation
	validateRequiredLength(&violations, "name", values.Name, ManufacturerNameMaxLength)
	validateOptionalLength(&violations, "description", values.Description, ManufacturerDescriptionMaxLength)
	slug, err := shared.ParseSlug(values.Slug, ManufacturerSlugMaxLength)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	return normalizedManufacturerValues{
		name: values.Name, slug: slug, description: values.Description,
	}, violations
}
