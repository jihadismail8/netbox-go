package dcim

import (
	"strings"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultSitePageLimit uint32 = 50
	MaximumSitePageLimit uint32 = 1000
)

type ListSitesQuery struct {
	Limit        uint32
	LimitPresent bool
	Offset       uint32
	Query        string
	IDs          []int64
	Ordering     []string
	Names        []string
	Slugs        []string
	Statuses     []string
}

// EffectiveLimit preserves the pinned OptionalLimitOffsetPagination contract:
// an omitted limit uses PAGINATE_COUNT while an explicit zero uses
// MAX_PAGE_SIZE. A non-zero programmatic value is considered present so
// callers which predate LimitPresent retain their intended page size.
func (query ListSitesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumSitePageLimit
	}
	return DefaultSitePageLimit
}

type GetSiteQuery struct {
	ID shared.ID
}

type SiteSortField string

const (
	SiteSortID          SiteSortField = "id"
	SiteSortName        SiteSortField = "name"
	SiteSortSlug        SiteSortField = "slug"
	SiteSortStatus      SiteSortField = "status"
	SiteSortCreated     SiteSortField = "created"
	SiteSortLastUpdated SiteSortField = "last_updated"
)

type SiteSort struct {
	Field      SiteSortField
	Descending bool
}

func validateListSitesQuery(query ListSitesQuery) (SiteListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumSitePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field:       "limit",
			Reason:      "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}

	// List-filter IDs are query values, not persisted aggregate identities.
	// NetBox's MultiValueNumberFilter uses forms.IntegerField without a
	// minimum, so zero and negative values are valid filters which yield no
	// rows when they do not match a database ID.
	ids := append([]int64(nil), query.IDs...)

	ordering, orderingViolations := parseSiteOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)

	statuses := make([]dcimdomain.SiteStatus, 0, len(query.Statuses))
	for _, requestedStatus := range query.Statuses {
		normalizedStatus := strings.TrimSpace(requestedStatus)
		parsedStatus, valid := dcimdomain.ParseSiteStatus(normalizedStatus)
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field:       "status",
				Reason:      "invalid_choice",
				Description: "Select a valid choice.",
			})
		} else {
			statuses = append(statuses, parsedStatus)
		}
	}

	if len(violations) > 0 {
		return SiteListCriteria{}, shared.NewValidationError(violations...)
	}

	return SiteListCriteria{
		Limit:    limit,
		Offset:   query.Offset,
		Query:    strings.TrimSpace(query.Query),
		IDs:      ids,
		Ordering: ordering,
		Names:    trimmedStrings(query.Names),
		Slugs:    trimmedStrings(query.Slugs),
		Statuses: statuses,
	}, nil
}

func parseSiteOrdering(values []string) ([]SiteSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []SiteSort{{Field: SiteSortName}}, nil
	}

	ordering := make([]SiteSort, 0, len(values))
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			fieldName := strings.TrimPrefix(item, "-")
			field, valid := parseSiteSortField(fieldName)
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field:       "ordering",
					Reason:      "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, SiteSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseSiteSortField(value string) (SiteSortField, bool) {
	field := SiteSortField(value)
	switch field {
	case SiteSortID,
		SiteSortName,
		SiteSortSlug,
		SiteSortStatus,
		SiteSortCreated,
		SiteSortLastUpdated:
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
