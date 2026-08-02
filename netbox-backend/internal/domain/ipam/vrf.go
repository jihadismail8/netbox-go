package ipam

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"netbox-go/internal/domain/shared"
)

const (
	VRFObjectType           = "ipam.vrf"
	VRFNameMaxLength        = 100
	VRFDescriptionMaxLength = 200
)

// VRFValues is the complete writable state after transport presence and
// defaults have been resolved by the application layer.
type VRFValues struct {
	Name          string
	RD            NullableRouteDistinguisher
	EnforceUnique bool
	Description   string
	Comments      string
}

// VRFPatch preserves mutation presence. RD points to a nullable value so the
// domain can distinguish omitted from explicitly null.
type VRFPatch struct {
	Name          *string
	RD            *NullableRouteDistinguisher
	EnforceUnique *bool
	Description   *string
	Comments      *string
}

func (patch VRFPatch) Empty() bool {
	return patch.Name == nil &&
		patch.RD == nil &&
		patch.EnforceUnique == nil &&
		patch.Description == nil &&
		patch.Comments == nil
}

// VRFState is the typed persistence boundary for a VRF aggregate.
type VRFState struct {
	ID             shared.ID
	Name           string
	RD             NullableRouteDistinguisher
	EnforceUnique  bool
	Description    string
	Comments       string
	Created        shared.Timestamp
	LastUpdated    shared.Timestamp
	IPAddressCount uint64
	PrefixCount    uint64
}

// VRFSnapshot is the immutable typed object-change projection.
type VRFSnapshot struct {
	Name          string
	RD            NullableRouteDistinguisher
	EnforceUnique bool
	Description   string
	Comments      string
}

func (VRFSnapshot) ObjectType() string { return VRFObjectType }

func (snapshot VRFSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }

// VRF owns all local invariants. RD uniqueness and protected references are
// cross-aggregate/database invariants enforced by the repository.
type VRF struct {
	id             shared.ID
	name           string
	rd             NullableRouteDistinguisher
	enforceUnique  bool
	description    string
	comments       string
	created        shared.Timestamp
	lastUpdated    shared.Timestamp
	ipAddressCount uint64
	prefixCount    uint64
}

func NewVRF(values VRFValues, now shared.Timestamp) (*VRF, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}

	normalized, violations := validateVRFValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}

	return &VRF{
		name:          normalized.Name,
		rd:            normalized.RD,
		enforceUnique: normalized.EnforceUnique,
		description:   normalized.Description,
		comments:      normalized.Comments,
		created:       now,
		lastUpdated:   now,
	}, nil
}

func RestoreVRF(state VRFState) (*VRF, error) {
	if !state.ID.IsValid() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a VRF with an invalid ID.")
	}
	if state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a VRF with a zero timestamp.")
	}

	normalized, violations := validateVRFValues(VRFValues{
		Name:          state.Name,
		RD:            state.RD,
		EnforceUnique: state.EnforceUnique,
		Description:   state.Description,
		Comments:      state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted VRF violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}

	return &VRF{
		id:             state.ID,
		name:           normalized.Name,
		rd:             normalized.RD,
		enforceUnique:  normalized.EnforceUnique,
		description:    normalized.Description,
		comments:       normalized.Comments,
		created:        state.Created,
		lastUpdated:    state.LastUpdated,
		ipAddressCount: state.IPAddressCount,
		prefixCount:    state.PrefixCount,
	}, nil
}

func (vrf *VRF) AssignID(id shared.ID) error {
	if vrf == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid VRF ID.")
	}
	if vrf.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned VRF ID.")
	}
	vrf.id = id
	return nil
}

func (vrf *VRF) Replace(values VRFValues, now shared.Timestamp) error {
	if vrf == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace a nil VRF.")
	}
	if now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}

	normalized, violations := validateVRFValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	vrf.name = normalized.Name
	vrf.rd = normalized.RD
	vrf.enforceUnique = normalized.EnforceUnique
	vrf.description = normalized.Description
	vrf.comments = normalized.Comments
	vrf.lastUpdated = now
	return nil
}

func (vrf *VRF) ApplyPatch(patch VRFPatch, now shared.Timestamp) error {
	if vrf == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil VRF.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field:       "update_mask",
			Reason:      "required",
			Description: "At least one writable field must be supplied.",
		})
	}

	values := vrf.Values()
	if patch.Name != nil {
		values.Name = *patch.Name
	}
	if patch.RD != nil {
		values.RD = *patch.RD
	}
	if patch.EnforceUnique != nil {
		values.EnforceUnique = *patch.EnforceUnique
	}
	if patch.Description != nil {
		values.Description = *patch.Description
	}
	if patch.Comments != nil {
		values.Comments = *patch.Comments
	}
	return vrf.Replace(values, now)
}

func (vrf VRF) ID() shared.ID                  { return vrf.id }
func (vrf VRF) Name() string                   { return vrf.name }
func (vrf VRF) RD() NullableRouteDistinguisher { return vrf.rd }
func (vrf VRF) EnforceUnique() bool            { return vrf.enforceUnique }
func (vrf VRF) Description() string            { return vrf.description }
func (vrf VRF) Comments() string               { return vrf.comments }
func (vrf VRF) Created() shared.Timestamp      { return vrf.created }
func (vrf VRF) LastUpdated() shared.Timestamp  { return vrf.lastUpdated }
func (vrf VRF) IPAddressCount() uint64         { return vrf.ipAddressCount }
func (vrf VRF) PrefixCount() uint64            { return vrf.prefixCount }
func (vrf VRF) Values() VRFValues              { return vrfValues(vrf) }
func (vrf VRF) State() VRFState                { return vrfState(vrf) }
func (vrf VRF) Snapshot() VRFSnapshot          { return vrfSnapshot(vrf) }
func (vrf VRF) Display() string {
	rd, present := vrf.rd.Get()
	if !present || rd.String() == "" {
		return vrf.name
	}
	return fmt.Sprintf("%s (%s)", vrf.name, rd.String())
}

func vrfValues(vrf VRF) VRFValues {
	return VRFValues{
		Name:          vrf.name,
		RD:            vrf.rd,
		EnforceUnique: vrf.enforceUnique,
		Description:   vrf.description,
		Comments:      vrf.comments,
	}
}

func vrfState(vrf VRF) VRFState {
	return VRFState{
		ID:             vrf.id,
		Name:           vrf.name,
		RD:             vrf.rd,
		EnforceUnique:  vrf.enforceUnique,
		Description:    vrf.description,
		Comments:       vrf.comments,
		Created:        vrf.created,
		LastUpdated:    vrf.lastUpdated,
		IPAddressCount: vrf.ipAddressCount,
		PrefixCount:    vrf.prefixCount,
	}
}

func vrfSnapshot(vrf VRF) VRFSnapshot {
	return VRFSnapshot{
		Name:          vrf.name,
		RD:            vrf.rd,
		EnforceUnique: vrf.enforceUnique,
		Description:   vrf.description,
		Comments:      vrf.comments,
	}
}

func validateVRFValues(values VRFValues) (VRFValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Description = strings.TrimSpace(values.Description)
	values.Comments = strings.TrimSpace(values.Comments)

	var violations []shared.FieldViolation
	if values.Name == "" {
		violations = append(violations, shared.FieldViolation{
			Field:       "name",
			Reason:      "required",
			Description: "This field may not be blank.",
		})
	} else if utf8.RuneCountInString(values.Name) > VRFNameMaxLength {
		violations = append(violations, maxLengthViolation("name", VRFNameMaxLength))
	}
	if utf8.RuneCountInString(values.Description) > VRFDescriptionMaxLength {
		violations = append(violations, maxLengthViolation("description", VRFDescriptionMaxLength))
	}
	return values, violations
}

func maxLengthViolation(field string, maximum int) shared.FieldViolation {
	return shared.FieldViolation{
		Field:       field,
		Reason:      "max_length",
		Description: fmt.Sprintf("Ensure this field has no more than %d characters.", maximum),
	}
}
