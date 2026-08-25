package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	DeviceRoleObjectType           = "dcim.devicerole"
	DeviceRoleNameMaxLength        = 100
	DeviceRoleSlugMaxLength        = 100
	DeviceRoleDescriptionMaxLength = 200
	DeviceRoleDefaultColor         = RackRoleDefaultColor
)

// DeviceRoleParent preserves the difference between a root role and a role
// whose parent has a persisted identity.
type DeviceRoleParent struct {
	id    shared.ID
	valid bool
}

func RootDeviceRoleParent() DeviceRoleParent { return DeviceRoleParent{} }

func NonRootDeviceRoleParent(id shared.ID) DeviceRoleParent {
	return DeviceRoleParent{id: id, valid: true}
}

func (parent DeviceRoleParent) Get() (shared.ID, bool) {
	return parent.id, parent.valid
}

func (parent DeviceRoleParent) IsRoot() bool { return !parent.valid }

type DeviceRoleValues struct {
	Parent      DeviceRoleParent
	Name        string
	Slug        string
	Color       string
	VMRole      bool
	Description string
	Comments    string
}

type DeviceRolePatch struct {
	Parent      *DeviceRoleParent
	Name        *string
	Slug        *string
	Color       *string
	VMRole      *bool
	Description *string
	Comments    *string
}

func (patch DeviceRolePatch) Empty() bool {
	return patch.Parent == nil && patch.Name == nil && patch.Slug == nil && patch.Color == nil &&
		patch.VMRole == nil && patch.Description == nil && patch.Comments == nil
}

type DeviceRoleReference struct {
	ID      shared.ID
	Display string
}

type DeviceRoleState struct {
	ID            shared.ID
	Parent        DeviceRoleParent
	ParentDisplay string
	Name          string
	Slug          string
	Color         string
	VMRole        bool
	Description   string
	Comments      string
	Created       shared.Timestamp
	LastUpdated   shared.Timestamp
	DeviceCount   uint64
	Depth         uint32
}

type DeviceRoleSnapshot struct {
	ParentID    *int64
	Name        string
	Slug        string
	Color       string
	VMRole      bool
	Description string
	Comments    string
}

func (DeviceRoleSnapshot) ObjectType() string { return DeviceRoleObjectType }

func (snapshot DeviceRoleSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	cloned := snapshot
	if snapshot.ParentID != nil {
		value := *snapshot.ParentID
		cloned.ParentID = &value
	}
	return cloned
}

type DeviceRole struct {
	id            shared.ID
	parent        DeviceRoleParent
	parentDisplay string
	name          string
	slug          shared.Slug
	color         RackRoleColor
	vmRole        bool
	description   string
	comments      string
	created       shared.Timestamp
	lastUpdated   shared.Timestamp
	deviceCount   uint64
	depth         uint32
}

func NewDeviceRole(values DeviceRoleValues, now shared.Timestamp) (*DeviceRole, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateDeviceRoleValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return &DeviceRole{
		parent: normalized.parent, name: normalized.name, slug: normalized.slug,
		color: normalized.color, vmRole: normalized.vmRole, description: normalized.description,
		comments: normalized.comments, created: now, lastUpdated: now,
	}, nil
}

func RestoreDeviceRole(state DeviceRoleState) (*DeviceRole, error) {
	if !state.ID.IsValid() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a DeviceRole with an invalid ID.")
	}
	if state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a DeviceRole with a zero timestamp.")
	}
	normalized, violations := validateDeviceRoleValues(DeviceRoleValues{
		Parent: state.Parent, Name: state.Name, Slug: state.Slug, Color: state.Color,
		VMRole: state.VMRole, Description: state.Description, Comments: state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted DeviceRole violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	if parentID, hasParent := normalized.parent.Get(); hasParent {
		if parentID == state.ID || strings.TrimSpace(state.ParentDisplay) == "" || state.Depth == 0 {
			return nil, shared.NewError(
				shared.ErrorReasonInternal,
				"Persisted DeviceRole contains an invalid hierarchy projection.",
			)
		}
	} else if state.Depth != 0 || strings.TrimSpace(state.ParentDisplay) != "" {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Persisted root DeviceRole contains an invalid hierarchy projection.",
		)
	}
	return &DeviceRole{
		id: state.ID, parent: normalized.parent, parentDisplay: strings.TrimSpace(state.ParentDisplay),
		name: normalized.name, slug: normalized.slug, color: normalized.color,
		vmRole: normalized.vmRole, description: normalized.description, comments: normalized.comments,
		created: state.Created, lastUpdated: state.LastUpdated,
		deviceCount: state.DeviceCount, depth: state.Depth,
	}, nil
}

func (role *DeviceRole) AssignID(id shared.ID) error {
	if role == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid DeviceRole ID.")
	}
	if role.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned DeviceRole ID.")
	}
	role.id = id
	return nil
}

func (role *DeviceRole) Replace(values DeviceRoleValues, now shared.Timestamp) error {
	if role == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace a nil DeviceRole.")
	}
	if now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateDeviceRoleValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	role.parent = normalized.parent
	role.parentDisplay = ""
	role.name = normalized.name
	role.slug = normalized.slug
	role.color = normalized.color
	role.vmRole = normalized.vmRole
	role.description = normalized.description
	role.comments = normalized.comments
	role.lastUpdated = now
	role.depth = 0
	return nil
}

func (role *DeviceRole) ApplyPatch(patch DeviceRolePatch, now shared.Timestamp) error {
	if role == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil DeviceRole.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		})
	}
	return role.Replace(role.valuesWithPatch(patch), now)
}

// ValidatePatch checks the state a patch would produce without mutating the
// aggregate. Empty patches are valid previews; ApplyPatch retains ownership of
// the public update-mask requirement.
func (role *DeviceRole) ValidatePatch(patch DeviceRolePatch) error {
	if role == nil {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot validate a patch for a nil DeviceRole.",
		)
	}
	_, violations := validateDeviceRoleValues(role.valuesWithPatch(patch))
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	return nil
}

func (role DeviceRole) valuesWithPatch(patch DeviceRolePatch) DeviceRoleValues {
	values := role.Values()
	if patch.Parent != nil {
		values.Parent = *patch.Parent
	}
	setString(&values.Name, patch.Name)
	setString(&values.Slug, patch.Slug)
	setString(&values.Color, patch.Color)
	if patch.VMRole != nil {
		values.VMRole = *patch.VMRole
	}
	setString(&values.Description, patch.Description)
	setString(&values.Comments, patch.Comments)
	return values
}

func (role DeviceRole) ID() shared.ID                 { return role.id }
func (role DeviceRole) Name() string                  { return role.name }
func (role DeviceRole) Slug() shared.Slug             { return role.slug }
func (role DeviceRole) Color() RackRoleColor          { return role.color }
func (role DeviceRole) VMRole() bool                  { return role.vmRole }
func (role DeviceRole) Description() string           { return role.description }
func (role DeviceRole) Comments() string              { return role.comments }
func (role DeviceRole) Created() shared.Timestamp     { return role.created }
func (role DeviceRole) LastUpdated() shared.Timestamp { return role.lastUpdated }
func (role DeviceRole) DeviceCount() uint64           { return role.deviceCount }
func (role DeviceRole) Depth() uint32                 { return role.depth }
func (role DeviceRole) Display() string               { return role.name }
func (role DeviceRole) Parent() DeviceRoleParent      { return role.parent }

func (role DeviceRole) ParentReference() (DeviceRoleReference, bool) {
	id, present := role.parent.Get()
	if !present {
		return DeviceRoleReference{}, false
	}
	return DeviceRoleReference{ID: id, Display: role.parentDisplay}, true
}

func (role DeviceRole) Values() DeviceRoleValues {
	return DeviceRoleValues{
		Parent: role.parent, Name: role.name, Slug: role.slug.String(), Color: role.color.String(),
		VMRole: role.vmRole, Description: role.description, Comments: role.comments,
	}
}

func (role DeviceRole) State() DeviceRoleState {
	return DeviceRoleState{
		ID: role.id, Parent: role.parent, ParentDisplay: role.parentDisplay,
		Name: role.name, Slug: role.slug.String(), Color: role.color.String(), VMRole: role.vmRole,
		Description: role.description, Comments: role.comments,
		Created: role.created, LastUpdated: role.lastUpdated,
		DeviceCount: role.deviceCount, Depth: role.depth,
	}
}

func (role DeviceRole) Snapshot() DeviceRoleSnapshot {
	var parentID *int64
	if id, present := role.parent.Get(); present {
		value := id.Int64()
		parentID = &value
	}
	return DeviceRoleSnapshot{
		ParentID: parentID, Name: role.name, Slug: role.slug.String(), Color: role.color.String(),
		VMRole: role.vmRole, Description: role.description, Comments: role.comments,
	}
}

type normalizedDeviceRoleValues struct {
	parent      DeviceRoleParent
	name        string
	slug        shared.Slug
	color       RackRoleColor
	vmRole      bool
	description string
	comments    string
}

func validateDeviceRoleValues(values DeviceRoleValues) (normalizedDeviceRoleValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Slug = strings.TrimSpace(values.Slug)
	values.Color = strings.TrimSpace(values.Color)
	values.Description = strings.TrimSpace(values.Description)
	values.Comments = strings.TrimSpace(values.Comments)
	var violations []shared.FieldViolation
	if parentID, present := values.Parent.Get(); present && !parentID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "parent", Reason: "invalid", Description: "A valid object ID is required.",
		})
	}
	validateRequiredLength(&violations, "name", values.Name, DeviceRoleNameMaxLength)
	slug, err := shared.ParseSlug(values.Slug, DeviceRoleSlugMaxLength)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	color, err := ParseRackRoleColor(values.Color)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	validateOptionalLength(&violations, "description", values.Description, DeviceRoleDescriptionMaxLength)
	return normalizedDeviceRoleValues{
		parent: values.Parent, name: values.Name, slug: slug, color: color,
		vmRole: values.VMRole, description: values.Description, comments: values.Comments,
	}, violations
}
