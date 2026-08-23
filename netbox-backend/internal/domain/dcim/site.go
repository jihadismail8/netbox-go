package dcim

import (
	"strings"
	"unicode/utf8"

	"netbox-go/internal/domain/shared"
)

const (
	SiteObjectType           = "dcim.site"
	SiteNameMaxLength        = 100
	SiteSlugMaxLength        = 100
	SiteFacilityMaxLength    = 50
	SiteDescriptionMaxLength = 200
)

// SiteValues contains the complete writable state of a Site. Transport field
// presence is resolved by the application layer before these values enter the
// domain.
type SiteValues struct {
	Name        string
	Slug        string
	Status      string
	Facility    string
	Description string
	Comments    string
}

// SitePatch preserves PATCH/FieldMask presence independently of string values.
type SitePatch struct {
	Name        *string
	Slug        *string
	Status      *string
	Facility    *string
	Description *string
	Comments    *string
}

func (patch SitePatch) Empty() bool {
	return patch.Name == nil &&
		patch.Slug == nil &&
		patch.Status == nil &&
		patch.Facility == nil &&
		patch.Description == nil &&
		patch.Comments == nil
}

// SiteState is the persistence mapping boundary for a Site aggregate.
type SiteState struct {
	ID          shared.ID
	Name        string
	Slug        string
	Status      string
	Facility    string
	Description string
	Comments    string
	Created     shared.Timestamp
	LastUpdated shared.Timestamp
	DeviceCount uint64
	PrefixCount uint64
	RackCount   uint64
}

// SiteSnapshot is the typed object-change projection for a Site mutation.
type SiteSnapshot struct {
	Name        string
	Slug        string
	Status      string
	Facility    string
	Description string
	Comments    string
}

func (SiteSnapshot) ObjectType() string { return SiteObjectType }

func (snapshot SiteSnapshot) CloneSnapshot() shared.ObjectSnapshot { return snapshot }

// Site owns validation and mutation of the first DCIM walking-skeleton
// aggregate. Fields stay private so adapters cannot bypass its invariants.
type Site struct {
	id          shared.ID
	name        string
	slug        shared.Slug
	status      SiteStatus
	facility    string
	description string
	comments    string
	created     shared.Timestamp
	lastUpdated shared.Timestamp
	deviceCount uint64
	prefixCount uint64
	rackCount   uint64
}

func NewSite(values SiteValues, now shared.Timestamp) (*Site, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}

	normalized, violations := validateSiteValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}

	return &Site{
		name:        normalized.name,
		slug:        normalized.slug,
		status:      normalized.status,
		facility:    normalized.facility,
		description: normalized.description,
		comments:    normalized.comments,
		created:     now,
		lastUpdated: now,
	}, nil
}

// RestoreSite reconstitutes persisted state through the same invariant checks
// used for mutations.
func RestoreSite(state SiteState) (*Site, error) {
	if !state.ID.IsValid() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a Site with an invalid ID.")
	}
	if state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot restore a Site with a zero timestamp.")
	}

	normalized, violations := validateSiteValues(SiteValues{
		Name:        state.Name,
		Slug:        state.Slug,
		Status:      state.Status,
		Facility:    state.Facility,
		Description: state.Description,
		Comments:    state.Comments,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted Site violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}

	return &Site{
		id:          state.ID,
		name:        normalized.name,
		slug:        normalized.slug,
		status:      normalized.status,
		facility:    normalized.facility,
		description: normalized.description,
		comments:    normalized.comments,
		created:     state.Created,
		lastUpdated: state.LastUpdated,
		deviceCount: state.DeviceCount,
		prefixCount: state.PrefixCount,
		rackCount:   state.RackCount,
	}, nil
}

// AssignID is used by a repository after a successful insert.
func (site *Site) AssignID(id shared.ID) error {
	if site == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid Site ID.")
	}
	if site.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned Site ID.")
	}
	site.id = id
	return nil
}

func (site *Site) Replace(values SiteValues, now shared.Timestamp) error {
	if site == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace a nil Site.")
	}
	if now.IsZero() {
		return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}

	normalized, violations := validateSiteValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}

	site.name = normalized.name
	site.slug = normalized.slug
	site.status = normalized.status
	site.facility = normalized.facility
	site.description = normalized.description
	site.comments = normalized.comments
	site.lastUpdated = now
	return nil
}

func (site *Site) ApplyPatch(patch SitePatch, now shared.Timestamp) error {
	if site == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil Site.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field:       "update_mask",
			Reason:      "required",
			Description: "At least one writable field must be supplied.",
		})
	}

	return site.Replace(site.valuesWithPatch(patch), now)
}

// ValidatePatch checks the state a patch would produce without mutating the
// aggregate. Empty patches are valid previews; ApplyPatch retains ownership of
// the update-mask requirement for public mutation behavior.
func (site *Site) ValidatePatch(patch SitePatch) error {
	if site == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot validate a patch for a nil Site.")
	}
	_, violations := validateSiteValues(site.valuesWithPatch(patch))
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	return nil
}

func (site Site) valuesWithPatch(patch SitePatch) SiteValues {
	values := site.Values()
	setString(&values.Name, patch.Name)
	setString(&values.Slug, patch.Slug)
	setString(&values.Status, patch.Status)
	setString(&values.Facility, patch.Facility)
	setString(&values.Description, patch.Description)
	setString(&values.Comments, patch.Comments)
	return values
}

func (site Site) ID() shared.ID                 { return site.id }
func (site Site) Name() string                  { return site.name }
func (site Site) Slug() shared.Slug             { return site.slug }
func (site Site) Status() SiteStatus            { return site.status }
func (site Site) Facility() string              { return site.facility }
func (site Site) Description() string           { return site.description }
func (site Site) Comments() string              { return site.comments }
func (site Site) Created() shared.Timestamp     { return site.created }
func (site Site) LastUpdated() shared.Timestamp { return site.lastUpdated }
func (site Site) DeviceCount() uint64           { return site.deviceCount }
func (site Site) PrefixCount() uint64           { return site.prefixCount }
func (site Site) RackCount() uint64             { return site.rackCount }
func (site Site) Display() string               { return site.name }
func (site Site) Values() SiteValues            { return siteValues(site) }
func (site Site) State() SiteState              { return siteState(site) }
func (site Site) Snapshot() SiteSnapshot        { return siteSnapshot(site) }

func siteValues(site Site) SiteValues {
	return SiteValues{
		Name:        site.name,
		Slug:        site.slug.String(),
		Status:      site.status.String(),
		Facility:    site.facility,
		Description: site.description,
		Comments:    site.comments,
	}
}

func siteState(site Site) SiteState {
	return SiteState{
		ID:          site.id,
		Name:        site.name,
		Slug:        site.slug.String(),
		Status:      site.status.String(),
		Facility:    site.facility,
		Description: site.description,
		Comments:    site.comments,
		Created:     site.created,
		LastUpdated: site.lastUpdated,
		DeviceCount: site.deviceCount,
		PrefixCount: site.prefixCount,
		RackCount:   site.rackCount,
	}
}

func siteSnapshot(site Site) SiteSnapshot {
	return SiteSnapshot{
		Name:        site.name,
		Slug:        site.slug.String(),
		Status:      site.status.String(),
		Facility:    site.facility,
		Description: site.description,
		Comments:    site.comments,
	}
}

type normalizedSiteValues struct {
	name        string
	slug        shared.Slug
	status      SiteStatus
	facility    string
	description string
	comments    string
}

func validateSiteValues(values SiteValues) (normalizedSiteValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Slug = strings.TrimSpace(values.Slug)
	values.Facility = strings.TrimSpace(values.Facility)
	values.Description = strings.TrimSpace(values.Description)
	values.Comments = strings.TrimSpace(values.Comments)

	var violations []shared.FieldViolation
	validateRequiredLength(&violations, "name", values.Name, SiteNameMaxLength)

	slug, err := shared.ParseSlug(values.Slug, SiteSlugMaxLength)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}

	status, validStatus := ParseSiteStatus(values.Status)
	if values.Status == "" {
		violations = append(violations, shared.FieldViolation{
			Field:       "status",
			Reason:      "blank",
			Description: "This field may not be blank.",
		})
	} else if !validStatus {
		violations = append(violations, shared.FieldViolation{
			Field:       "status",
			Reason:      "invalid_choice",
			Description: values.Status + " is not a valid choice.",
		})
	}

	validateOptionalLength(&violations, "facility", values.Facility, SiteFacilityMaxLength)
	validateOptionalLength(&violations, "description", values.Description, SiteDescriptionMaxLength)

	return normalizedSiteValues{
		name:        values.Name,
		slug:        slug,
		status:      status,
		facility:    values.Facility,
		description: values.Description,
		comments:    values.Comments,
	}, violations
}

func validateRequiredLength(
	violations *[]shared.FieldViolation,
	field string,
	value string,
	maxLength int,
) {
	if value == "" {
		*violations = append(*violations, shared.FieldViolation{
			Field:       field,
			Reason:      "required",
			Description: "This field may not be blank.",
		})
		return
	}
	validateOptionalLength(violations, field, value, maxLength)
}

func validateOptionalLength(
	violations *[]shared.FieldViolation,
	field string,
	value string,
	maxLength int,
) {
	if utf8.RuneCountInString(value) <= maxLength {
		return
	}
	*violations = append(*violations, shared.FieldViolation{
		Field:       field,
		Reason:      "max_length",
		Description: "Ensure this field has no more than the supported number of characters.",
	})
}

func setString(destination *string, source *string) {
	if source != nil {
		*destination = *source
	}
}
