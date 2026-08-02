package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	DefaultManufacturerPageLimit uint32 = 50
	MaximumManufacturerPageLimit uint32 = 1000
	DefaultRackRolePageLimit     uint32 = 50
	MaximumRackRolePageLimit     uint32 = 1000
)

type ListManufacturersQuery struct {
	Limit        uint32
	LimitPresent bool
	Offset       uint32
	Query        string
	IDs          []int64
	Ordering     []string
	Names        []string
	Slugs        []string
}

func (query ListManufacturersQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumManufacturerPageLimit
	}
	return DefaultManufacturerPageLimit
}

type GetManufacturerQuery struct{ ID shared.ID }

type ManufacturerSortField string

const (
	ManufacturerSortID          ManufacturerSortField = "id"
	ManufacturerSortName        ManufacturerSortField = "name"
	ManufacturerSortSlug        ManufacturerSortField = "slug"
	ManufacturerSortCreated     ManufacturerSortField = "created"
	ManufacturerSortLastUpdated ManufacturerSortField = "last_updated"
)

type ManufacturerSort struct {
	Field      ManufacturerSortField
	Descending bool
}

type ListRackRolesQuery struct {
	Limit        uint32
	LimitPresent bool
	Offset       uint32
	Query        string
	IDs          []int64
	Ordering     []string
	Names        []string
	Slugs        []string
}

func (query ListRackRolesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumRackRolePageLimit
	}
	return DefaultRackRolePageLimit
}

type GetRackRoleQuery struct{ ID shared.ID }

type RackRoleSortField string

const (
	RackRoleSortID          RackRoleSortField = "id"
	RackRoleSortName        RackRoleSortField = "name"
	RackRoleSortSlug        RackRoleSortField = "slug"
	RackRoleSortCreated     RackRoleSortField = "created"
	RackRoleSortLastUpdated RackRoleSortField = "last_updated"
)

type RackRoleSort struct {
	Field      RackRoleSortField
	Descending bool
}

func validateListManufacturersQuery(query ListManufacturersQuery) (ManufacturerListCriteria, error) {
	limit, commonViolations := validateOrganizationList(query.EffectiveLimit())
	ordering, orderingViolations := parseManufacturerOrdering(query.Ordering)
	violations := append(commonViolations, orderingViolations...)
	if len(violations) > 0 {
		return ManufacturerListCriteria{}, shared.NewValidationError(violations...)
	}
	return ManufacturerListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		Names: trimmedStrings(query.Names), Slugs: trimmedStrings(query.Slugs),
	}, nil
}

func validateListRackRolesQuery(query ListRackRolesQuery) (RackRoleListCriteria, error) {
	limit, commonViolations := validateOrganizationList(query.EffectiveLimit())
	ordering, orderingViolations := parseRackRoleOrdering(query.Ordering)
	violations := append(commonViolations, orderingViolations...)
	if len(violations) > 0 {
		return RackRoleListCriteria{}, shared.NewValidationError(violations...)
	}
	return RackRoleListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		Names: trimmedStrings(query.Names), Slugs: trimmedStrings(query.Slugs),
	}, nil
}

func validateOrganizationList(limit uint32) (uint32, []shared.FieldViolation) {
	var violations []shared.FieldViolation
	if limit > 1000 {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value", Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	return limit, violations
}

func parseManufacturerOrdering(values []string) ([]ManufacturerSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []ManufacturerSort{{Field: ManufacturerSortName}}, nil
	}
	var ordering []ManufacturerSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, ok := parseManufacturerSortField(strings.TrimPrefix(item, "-"))
			if !ok {
				violations = append(violations, invalidOrganizationOrdering())
				continue
			}
			ordering = append(ordering, ManufacturerSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseManufacturerSortField(value string) (ManufacturerSortField, bool) {
	field := ManufacturerSortField(value)
	switch field {
	case ManufacturerSortID, ManufacturerSortName, ManufacturerSortSlug,
		ManufacturerSortCreated, ManufacturerSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}

func parseRackRoleOrdering(values []string) ([]RackRoleSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []RackRoleSort{{Field: RackRoleSortName}}, nil
	}
	var ordering []RackRoleSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, ok := parseRackRoleSortField(strings.TrimPrefix(item, "-"))
			if !ok {
				violations = append(violations, invalidOrganizationOrdering())
				continue
			}
			ordering = append(ordering, RackRoleSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseRackRoleSortField(value string) (RackRoleSortField, bool) {
	field := RackRoleSortField(value)
	switch field {
	case RackRoleSortID, RackRoleSortName, RackRoleSortSlug,
		RackRoleSortCreated, RackRoleSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}

func invalidOrganizationOrdering() shared.FieldViolation {
	return shared.FieldViolation{
		Field: "ordering", Reason: "invalid_choice", Description: "Select a valid ordering field.",
	}
}
