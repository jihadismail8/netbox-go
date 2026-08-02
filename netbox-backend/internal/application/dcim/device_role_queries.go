package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	DefaultDeviceRolePageLimit uint32 = 50
	MaximumDeviceRolePageLimit uint32 = 1000
)

type ListDeviceRolesQuery struct {
	Limit        uint32
	LimitPresent bool
	Offset       uint32
	Query        string
	IDs          []int64
	Ordering     []string
	Names        []string
	Slugs        []string
}

func (query ListDeviceRolesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumDeviceRolePageLimit
	}
	return DefaultDeviceRolePageLimit
}

type GetDeviceRoleQuery struct{ ID shared.ID }

type DeviceRoleSortField string

const (
	DeviceRoleSortID          DeviceRoleSortField = "id"
	DeviceRoleSortName        DeviceRoleSortField = "name"
	DeviceRoleSortSlug        DeviceRoleSortField = "slug"
	DeviceRoleSortCreated     DeviceRoleSortField = "created"
	DeviceRoleSortLastUpdated DeviceRoleSortField = "last_updated"
)

type DeviceRoleSort struct {
	Field      DeviceRoleSortField
	Descending bool
}

func validateListDeviceRolesQuery(query ListDeviceRolesQuery) (DeviceRoleListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumDeviceRolePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, defaultTreeOrder, orderingViolations := parseDeviceRoleOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	if len(violations) > 0 {
		return DeviceRoleListCriteria{}, shared.NewValidationError(violations...)
	}
	return DeviceRoleListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		DefaultTreeOrder: defaultTreeOrder,
		Names:            trimmedStrings(query.Names), Slugs: trimmedStrings(query.Slugs),
	}, nil
}

func parseDeviceRoleOrdering(values []string) ([]DeviceRoleSort, bool, []shared.FieldViolation) {
	if len(values) == 0 {
		return nil, true, nil
	}
	var ordering []DeviceRoleSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseDeviceRoleSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, DeviceRoleSort{Field: field, Descending: descending})
		}
	}
	return ordering, false, violations
}

func parseDeviceRoleSortField(value string) (DeviceRoleSortField, bool) {
	field := DeviceRoleSortField(value)
	switch field {
	case DeviceRoleSortID, DeviceRoleSortName, DeviceRoleSortSlug,
		DeviceRoleSortCreated, DeviceRoleSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}
