package dcim

import (
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	DefaultDeviceTypePageLimit uint32 = 50
	MaximumDeviceTypePageLimit uint32 = 1000
)

type ListDeviceTypesQuery struct {
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

func (query ListDeviceTypesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumDeviceTypePageLimit
	}
	return DefaultDeviceTypePageLimit
}

type GetDeviceTypeQuery struct{ ID shared.ID }

type DeviceTypeSortField string

const (
	DeviceTypeSortID           DeviceTypeSortField = "id"
	DeviceTypeSortManufacturer DeviceTypeSortField = "manufacturer"
	DeviceTypeSortModel        DeviceTypeSortField = "model"
	DeviceTypeSortSlug         DeviceTypeSortField = "slug"
	DeviceTypeSortUHeight      DeviceTypeSortField = "u_height"
	DeviceTypeSortCreated      DeviceTypeSortField = "created"
	DeviceTypeSortLastUpdated  DeviceTypeSortField = "last_updated"
)

type DeviceTypeSort struct {
	Field      DeviceTypeSortField
	Descending bool
}

func validateListDeviceTypesQuery(
	query ListDeviceTypesQuery,
) (DeviceTypeListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumDeviceTypePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseDeviceTypeOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	if len(violations) > 0 {
		return DeviceTypeListCriteria{}, shared.NewValidationError(violations...)
	}
	return DeviceTypeListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		ManufacturerIDs:   append([]int64(nil), query.ManufacturerIDs...),
		ManufacturerSlugs: trimmedStrings(query.ManufacturerSlugs),
		Models:            trimmedStrings(query.Models),
		Slugs:             trimmedStrings(query.Slugs),
	}, nil
}

func parseDeviceTypeOrdering(
	values []string,
) ([]DeviceTypeSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []DeviceTypeSort{
			{Field: DeviceTypeSortManufacturer},
			{Field: DeviceTypeSortModel},
			{Field: DeviceTypeSortID},
		}, nil
	}
	var ordering []DeviceTypeSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseDeviceTypeSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, DeviceTypeSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseDeviceTypeSortField(value string) (DeviceTypeSortField, bool) {
	field := DeviceTypeSortField(value)
	switch field {
	case DeviceTypeSortID, DeviceTypeSortManufacturer, DeviceTypeSortModel,
		DeviceTypeSortSlug, DeviceTypeSortUHeight, DeviceTypeSortCreated,
		DeviceTypeSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}
