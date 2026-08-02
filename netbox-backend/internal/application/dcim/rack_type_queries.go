package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	DefaultRackTypePageLimit uint32 = 50
	MaximumRackTypePageLimit uint32 = 1000
)

type ListRackTypesQuery struct {
	Limit             uint32
	LimitPresent      bool
	Offset            uint32
	Query             string
	IDs               []int64
	Ordering          []string
	ManufacturerIDs   []int64
	ManufacturerSlugs []string
	Models            []string
	Slugs             []string
}

func (query ListRackTypesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumRackTypePageLimit
	}
	return DefaultRackTypePageLimit
}

type GetRackTypeQuery struct{ ID shared.ID }

type RackTypeSortField string

const (
	RackTypeSortID           RackTypeSortField = "id"
	RackTypeSortManufacturer RackTypeSortField = "manufacturer"
	RackTypeSortModel        RackTypeSortField = "model"
	RackTypeSortSlug         RackTypeSortField = "slug"
	RackTypeSortUHeight      RackTypeSortField = "u_height"
	RackTypeSortCreated      RackTypeSortField = "created"
	RackTypeSortLastUpdated  RackTypeSortField = "last_updated"
)

type RackTypeSort struct {
	Field      RackTypeSortField
	Descending bool
}

func validateListRackTypesQuery(query ListRackTypesQuery) (RackTypeListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumRackTypePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value", Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseRackTypeOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	if len(violations) > 0 {
		return RackTypeListCriteria{}, shared.NewValidationError(violations...)
	}
	return RackTypeListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		ManufacturerIDs:   append([]int64(nil), query.ManufacturerIDs...),
		ManufacturerSlugs: trimmedStrings(query.ManufacturerSlugs),
		Models:            trimmedStrings(query.Models), Slugs: trimmedStrings(query.Slugs),
	}, nil
}

func parseRackTypeOrdering(values []string) ([]RackTypeSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []RackTypeSort{
			{Field: RackTypeSortManufacturer},
			{Field: RackTypeSortModel},
			{Field: RackTypeSortID},
		}, nil
	}
	var ordering []RackTypeSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseRackTypeSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice", Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, RackTypeSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseRackTypeSortField(value string) (RackTypeSortField, bool) {
	field := RackTypeSortField(value)
	switch field {
	case RackTypeSortID, RackTypeSortManufacturer, RackTypeSortModel, RackTypeSortSlug,
		RackTypeSortUHeight, RackTypeSortCreated, RackTypeSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}
