package ipam

import (
	"strings"

	ipamdomain "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultVRFPageLimit uint32 = 50
	MaximumVRFPageLimit uint32 = 1000
)

type ListVRFsQuery struct {
	Limit         uint32
	LimitPresent  bool
	Offset        uint32
	Query         string
	IDs           []int64
	Ordering      []string
	Names         []string
	RDs           []string
	EnforceUnique *bool
}

// EffectiveLimit preserves NetBox's OptionalLimitOffsetPagination contract:
// an omitted limit uses PAGINATE_COUNT while an explicit zero uses the
// configured maximum page size. Non-zero programmatic values are treated as
// present so older callers retain their intended page size.
func (query ListVRFsQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumVRFPageLimit
	}
	return DefaultVRFPageLimit
}

type GetVRFQuery struct {
	ID shared.ID
}

type VRFSortField string

const (
	VRFSortID          VRFSortField = "id"
	VRFSortName        VRFSortField = "name"
	VRFSortRD          VRFSortField = "rd"
	VRFSortCreated     VRFSortField = "created"
	VRFSortLastUpdated VRFSortField = "last_updated"
)

type VRFSort struct {
	Field      VRFSortField
	Descending bool
}

func validateListVRFsQuery(query ListVRFsQuery) (VRFListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumVRFPageLimit {
		violations = append(violations, shared.FieldViolation{
			Field:       "limit",
			Reason:      "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}

	// Filter IDs are signed query values rather than persisted identities.
	// NetBox's MultiValueNumberFilter accepts zero and negative integers; such
	// values simply match no positive database IDs.
	ids := append([]int64(nil), query.IDs...)

	ordering, orderingViolations := parseVRFOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)

	rds := make([]ipamdomain.RouteDistinguisher, 0, len(query.RDs))
	for _, requestedRD := range query.RDs {
		parsed, err := ipamdomain.ParseRouteDistinguisher(requestedRD)
		if err != nil {
			violations = append(violations, shared.ViolationsOf(err)...)
		} else {
			rds = append(rds, parsed)
		}
	}

	if len(violations) > 0 {
		return VRFListCriteria{}, shared.NewValidationError(violations...)
	}
	return VRFListCriteria{
		Limit:         limit,
		Offset:        query.Offset,
		Query:         strings.TrimSpace(query.Query),
		IDs:           ids,
		Ordering:      ordering,
		Names:         trimmedStrings(query.Names),
		RDs:           rds,
		EnforceUnique: query.EnforceUnique,
	}, nil
}

func parseVRFOrdering(values []string) ([]VRFSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []VRFSort{
			{Field: VRFSortName},
			{Field: VRFSortRD},
		}, nil
	}

	ordering := make([]VRFSort, 0, len(values))
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			fieldName := strings.TrimPrefix(item, "-")
			field, valid := parseVRFSortField(fieldName)
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field:       "ordering",
					Reason:      "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, VRFSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseVRFSortField(value string) (VRFSortField, bool) {
	field := VRFSortField(value)
	switch field {
	case VRFSortID, VRFSortName, VRFSortRD, VRFSortCreated, VRFSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}

func trimmedStrings(values []string) []string {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		trimmed = append(trimmed, strings.TrimSpace(value))
	}
	return trimmed
}
