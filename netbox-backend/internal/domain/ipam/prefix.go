package ipam

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"netbox-go/internal/domain/shared"
)

const (
	PrefixObjectType           = "ipam.prefix"
	PrefixDescriptionMaxLength = 200
)

type PrefixStatus string

const (
	PrefixStatusContainer  PrefixStatus = "container"
	PrefixStatusActive     PrefixStatus = "active"
	PrefixStatusReserved   PrefixStatus = "reserved"
	PrefixStatusDeprecated PrefixStatus = "deprecated"
)

func ParsePrefixStatus(value string) (PrefixStatus, bool) {
	status := PrefixStatus(strings.TrimSpace(value))
	switch status {
	case PrefixStatusContainer, PrefixStatusActive, PrefixStatusReserved, PrefixStatusDeprecated:
		return status, true
	default:
		return "", false
	}
}

func (status PrefixStatus) String() string { return string(status) }

// VRFReference is the immutable relationship projection required by Prefix.
// EnforceUnique is retained because it controls the Prefix uniqueness rule.
type VRFReference struct {
	id            shared.ID
	name          string
	rd            NullableRouteDistinguisher
	enforceUnique bool
}

func NewVRFReference(
	id shared.ID,
	name string,
	rd NullableRouteDistinguisher,
	enforceUnique bool,
) (VRFReference, error) {
	name = strings.TrimSpace(name)
	if !id.IsValid() || name == "" {
		return VRFReference{}, shared.NewError(
			shared.ErrorReasonInternal, "Cannot construct an invalid VRF reference.",
		)
	}
	return VRFReference{id: id, name: name, rd: rd, enforceUnique: enforceUnique}, nil
}

func (reference VRFReference) ID() shared.ID                  { return reference.id }
func (reference VRFReference) Name() string                   { return reference.name }
func (reference VRFReference) RD() NullableRouteDistinguisher { return reference.rd }
func (reference VRFReference) EnforceUnique() bool            { return reference.enforceUnique }
func (reference VRFReference) Display() string {
	rd, present := reference.rd.Get()
	if !present || rd.String() == "" {
		return reference.name
	}
	return fmt.Sprintf("%s (%s)", reference.name, rd.String())
}
func (reference VRFReference) Valid() bool { return reference.id.IsValid() && reference.name != "" }

type NullableVRFReference struct {
	reference VRFReference
	valid     bool
}

func NullVRFReference() NullableVRFReference { return NullableVRFReference{} }

func NonNullVRFReference(reference VRFReference) NullableVRFReference {
	return NullableVRFReference{reference: reference, valid: true}
}

func (nullable NullableVRFReference) Get() (VRFReference, bool) {
	return nullable.reference, nullable.valid
}

func (nullable NullableVRFReference) IsNull() bool { return !nullable.valid }

type PrefixValues struct {
	Prefix       string
	VRF          NullableVRFReference
	Status       string
	IsPool       bool
	MarkUtilized bool
	Description  string
	Comments     string
}

type PrefixPatch struct {
	Prefix       *string
	VRF          *NullableVRFReference
	Status       *string
	IsPool       *bool
	MarkUtilized *bool
	Description  *string
	Comments     *string
}

func (patch PrefixPatch) Empty() bool {
	return patch.Prefix == nil && patch.VRF == nil && patch.Status == nil &&
		patch.IsPool == nil && patch.MarkUtilized == nil && patch.Description == nil &&
		patch.Comments == nil
}

type PrefixState struct {
	ID           shared.ID
	Prefix       string
	VRF          NullableVRFReference
	Status       string
	IsPool       bool
	MarkUtilized bool
	Description  string
	Comments     string
	Created      shared.Timestamp
	LastUpdated  shared.Timestamp
	Children     uint64
	Depth        uint32
}

type PrefixSnapshot struct {
	Prefix       string
	VRF          NullableVRFReference
	Status       string
	IsPool       bool
	MarkUtilized bool
	Description  string
	Comments     string
}

func (PrefixSnapshot) ObjectType() string                            { return PrefixObjectType }
func (snapshot PrefixSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }

type Prefix struct {
	id           shared.ID
	network      PrefixNetwork
	vrf          NullableVRFReference
	status       PrefixStatus
	isPool       bool
	markUtilized bool
	description  string
	comments     string
	created      shared.Timestamp
	lastUpdated  shared.Timestamp
	children     uint64
	depth        uint32
}

func NewPrefix(values PrefixValues, now shared.Timestamp) (*Prefix, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validatePrefixValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return prefixFromNormalized(normalized, now, now), nil
}

func RestorePrefix(state PrefixState) (*Prefix, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore invalid Prefix identity or timestamps.")
	}
	normalized, violations := validatePrefixValues(PrefixValues{
		Prefix: state.Prefix, VRF: state.VRF, Status: state.Status,
		IsPool: state.IsPool, MarkUtilized: state.MarkUtilized,
		Description: state.Description, Comments: state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Persisted Prefix violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	prefix := prefixFromNormalized(normalized, state.Created, state.LastUpdated)
	prefix.id = state.ID
	prefix.children = state.Children
	prefix.depth = state.Depth
	return prefix, nil
}

func prefixFromNormalized(
	values normalizedPrefixValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *Prefix {
	return &Prefix{
		network: values.network, vrf: values.vrf, status: values.status,
		isPool: values.isPool, markUtilized: values.markUtilized,
		description: values.description, comments: values.comments,
		created: created, lastUpdated: lastUpdated,
	}
}

func (prefix *Prefix) AssignID(id shared.ID) error {
	if prefix == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid Prefix ID.")
	}
	if prefix.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned Prefix ID.")
	}
	prefix.id = id
	return nil
}

func (prefix *Prefix) Replace(values PrefixValues, now shared.Timestamp) error {
	if prefix == nil || now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace a nil Prefix or use a zero timestamp.")
	}
	normalized, violations := validatePrefixValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	prefix.network = normalized.network
	prefix.vrf = normalized.vrf
	prefix.status = normalized.status
	prefix.isPool = normalized.isPool
	prefix.markUtilized = normalized.markUtilized
	prefix.description = normalized.description
	prefix.comments = normalized.comments
	prefix.lastUpdated = now
	return nil
}

func (prefix *Prefix) ApplyPatch(patch PrefixPatch, now shared.Timestamp) error {
	if prefix == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil Prefix.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	values := prefix.Values()
	if patch.Prefix != nil {
		values.Prefix = *patch.Prefix
	}
	if patch.VRF != nil {
		values.VRF = *patch.VRF
	}
	if patch.Status != nil {
		values.Status = *patch.Status
	}
	if patch.IsPool != nil {
		values.IsPool = *patch.IsPool
	}
	if patch.MarkUtilized != nil {
		values.MarkUtilized = *patch.MarkUtilized
	}
	if patch.Description != nil {
		values.Description = *patch.Description
	}
	if patch.Comments != nil {
		values.Comments = *patch.Comments
	}
	return prefix.Replace(values, now)
}

func (prefix Prefix) ID() shared.ID                 { return prefix.id }
func (prefix Prefix) Network() PrefixNetwork        { return prefix.network }
func (prefix Prefix) VRF() NullableVRFReference     { return prefix.vrf }
func (prefix Prefix) Status() PrefixStatus          { return prefix.status }
func (prefix Prefix) IsPool() bool                  { return prefix.isPool }
func (prefix Prefix) MarkUtilized() bool            { return prefix.markUtilized }
func (prefix Prefix) Description() string           { return prefix.description }
func (prefix Prefix) Comments() string              { return prefix.comments }
func (prefix Prefix) Created() shared.Timestamp     { return prefix.created }
func (prefix Prefix) LastUpdated() shared.Timestamp { return prefix.lastUpdated }
func (prefix Prefix) Family() uint32                { return prefix.network.Family() }
func (prefix Prefix) Children() uint64              { return prefix.children }
func (prefix Prefix) Depth() uint32                 { return prefix.depth }
func (prefix Prefix) Display() string               { return prefix.network.String() }

func (prefix Prefix) Values() PrefixValues {
	return PrefixValues{
		Prefix: prefix.network.String(), VRF: prefix.vrf, Status: prefix.status.String(),
		IsPool: prefix.isPool, MarkUtilized: prefix.markUtilized,
		Description: prefix.description, Comments: prefix.comments,
	}
}

func (prefix Prefix) State() PrefixState {
	return PrefixState{
		ID: prefix.id, Prefix: prefix.network.String(), VRF: prefix.vrf,
		Status: prefix.status.String(), IsPool: prefix.isPool, MarkUtilized: prefix.markUtilized,
		Description: prefix.description, Comments: prefix.comments,
		Created: prefix.created, LastUpdated: prefix.lastUpdated,
		Children: prefix.children, Depth: prefix.depth,
	}
}

func (prefix Prefix) Snapshot() PrefixSnapshot {
	return PrefixSnapshot{
		Prefix: prefix.network.String(), VRF: prefix.vrf, Status: prefix.status.String(),
		IsPool: prefix.isPool, MarkUtilized: prefix.markUtilized,
		Description: prefix.description, Comments: prefix.comments,
	}
}

type normalizedPrefixValues struct {
	network      PrefixNetwork
	vrf          NullableVRFReference
	status       PrefixStatus
	isPool       bool
	markUtilized bool
	description  string
	comments     string
}

func validatePrefixValues(values PrefixValues) (normalizedPrefixValues, []shared.FieldViolation) {
	var violations []shared.FieldViolation
	network, err := ParsePrefixNetwork(values.Prefix)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	if reference, present := values.VRF.Get(); present && !reference.Valid() {
		violations = append(violations, shared.FieldViolation{
			Field: "vrf", Reason: "invalid", Description: "Invalid VRF reference.",
		})
	}
	status, valid := ParsePrefixStatus(values.Status)
	if !valid {
		violations = append(violations, shared.FieldViolation{
			Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	description := strings.TrimSpace(values.Description)
	comments := strings.TrimSpace(values.Comments)
	if utf8.RuneCountInString(description) > PrefixDescriptionMaxLength {
		violations = append(violations, shared.FieldViolation{
			Field: "description", Reason: "max_length",
			Description: fmt.Sprintf("Ensure this field has no more than %d characters.", PrefixDescriptionMaxLength),
		})
	}
	return normalizedPrefixValues{
		network: network, vrf: values.VRF, status: status,
		isPool: values.IsPool, markUtilized: values.MarkUtilized,
		description: description, comments: comments,
	}, violations
}
