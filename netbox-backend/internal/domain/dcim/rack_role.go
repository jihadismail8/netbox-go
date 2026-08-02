package dcim

import (
	"regexp"
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	RackRoleObjectType           = "dcim.rackrole"
	RackRoleNameMaxLength        = 100
	RackRoleSlugMaxLength        = 100
	RackRoleDescriptionMaxLength = 200
	RackRoleDefaultColor         = "9e9e9e"
)

var rackRoleColorPattern = regexp.MustCompile(`^[0-9a-f]{6}$`)

// RackRoleColor is a validated lowercase RRGGBB color code.
type RackRoleColor string

func ParseRackRoleColor(value string) (RackRoleColor, error) {
	value = strings.TrimSpace(value)
	if !rackRoleColorPattern.MatchString(value) {
		return "", shared.NewValidationError(shared.FieldViolation{
			Field: "color", Reason: "invalid", Description: "Enter a valid hexadecimal RGB color code.",
		})
	}
	return RackRoleColor(value), nil
}

func (color RackRoleColor) String() string { return string(color) }

type RackRoleValues struct {
	Name        string
	Slug        string
	Color       string
	Description string
}

type RackRolePatch struct {
	Name        *string
	Slug        *string
	Color       *string
	Description *string
}

func (patch RackRolePatch) Empty() bool {
	return patch.Name == nil && patch.Slug == nil && patch.Color == nil && patch.Description == nil
}

type RackRoleState struct {
	ID          shared.ID
	Name        string
	Slug        string
	Color       string
	Description string
	Created     shared.Timestamp
	LastUpdated shared.Timestamp
	RackCount   uint64
}

type RackRoleSnapshot struct {
	Name        string
	Slug        string
	Color       string
	Description string
}

func (RackRoleSnapshot) ObjectType() string { return RackRoleObjectType }

func (snapshot RackRoleSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }

type RackRole struct {
	id          shared.ID
	name        string
	slug        shared.Slug
	color       RackRoleColor
	description string
	created     shared.Timestamp
	lastUpdated shared.Timestamp
	rackCount   uint64
}

func NewRackRole(values RackRoleValues, now shared.Timestamp) (*RackRole, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateRackRoleValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return &RackRole{
		name: normalized.name, slug: normalized.slug, color: normalized.color,
		description: normalized.description, created: now, lastUpdated: now,
	}, nil
}

func RestoreRackRole(state RackRoleState) (*RackRole, error) {
	if !state.ID.IsValid() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a RackRole with an invalid ID.")
	}
	if state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a RackRole with a zero timestamp.")
	}
	normalized, violations := validateRackRoleValues(RackRoleValues{
		Name: state.Name, Slug: state.Slug, Color: state.Color, Description: state.Description,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted RackRole violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	return &RackRole{
		id: state.ID, name: normalized.name, slug: normalized.slug, color: normalized.color,
		description: normalized.description, created: state.Created, lastUpdated: state.LastUpdated,
		rackCount: state.RackCount,
	}, nil
}

func (role *RackRole) AssignID(id shared.ID) error {
	if role == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid RackRole ID.")
	}
	if role.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned RackRole ID.")
	}
	role.id = id
	return nil
}

func (role *RackRole) Replace(values RackRoleValues, now shared.Timestamp) error {
	if role == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace a nil RackRole.")
	}
	if now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateRackRoleValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	role.name = normalized.name
	role.slug = normalized.slug
	role.color = normalized.color
	role.description = normalized.description
	role.lastUpdated = now
	return nil
}

func (role *RackRole) ApplyPatch(patch RackRolePatch, now shared.Timestamp) error {
	if role == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil RackRole.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		})
	}
	values := role.Values()
	setString(&values.Name, patch.Name)
	setString(&values.Slug, patch.Slug)
	setString(&values.Color, patch.Color)
	setString(&values.Description, patch.Description)
	return role.Replace(values, now)
}

func (role RackRole) ID() shared.ID                 { return role.id }
func (role RackRole) Name() string                  { return role.name }
func (role RackRole) Slug() shared.Slug             { return role.slug }
func (role RackRole) Color() RackRoleColor          { return role.color }
func (role RackRole) Description() string           { return role.description }
func (role RackRole) Created() shared.Timestamp     { return role.created }
func (role RackRole) LastUpdated() shared.Timestamp { return role.lastUpdated }
func (role RackRole) RackCount() uint64             { return role.rackCount }
func (role RackRole) Display() string               { return role.name }

func (role RackRole) Values() RackRoleValues {
	return RackRoleValues{
		Name: role.name, Slug: role.slug.String(), Color: role.color.String(), Description: role.description,
	}
}

func (role RackRole) State() RackRoleState {
	return RackRoleState{
		ID: role.id, Name: role.name, Slug: role.slug.String(), Color: role.color.String(),
		Description: role.description, Created: role.created, LastUpdated: role.lastUpdated,
		RackCount: role.rackCount,
	}
}

func (role RackRole) Snapshot() RackRoleSnapshot {
	return RackRoleSnapshot{
		Name: role.name, Slug: role.slug.String(), Color: role.color.String(), Description: role.description,
	}
}

type normalizedRackRoleValues struct {
	name        string
	slug        shared.Slug
	color       RackRoleColor
	description string
}

func validateRackRoleValues(values RackRoleValues) (normalizedRackRoleValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Slug = strings.TrimSpace(values.Slug)
	values.Color = strings.TrimSpace(values.Color)
	values.Description = strings.TrimSpace(values.Description)
	var violations []shared.FieldViolation
	validateRequiredLength(&violations, "name", values.Name, RackRoleNameMaxLength)
	validateOptionalLength(&violations, "description", values.Description, RackRoleDescriptionMaxLength)
	slug, err := shared.ParseSlug(values.Slug, RackRoleSlugMaxLength)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	color, err := ParseRackRoleColor(values.Color)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	return normalizedRackRoleValues{
		name: values.Name, slug: slug, color: color, description: values.Description,
	}, violations
}
