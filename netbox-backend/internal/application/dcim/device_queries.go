package dcim

import (
	"strings"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultDevicePageLimit uint32 = 50
	MaximumDevicePageLimit uint32 = 1000
)

type ListDevicesQuery struct {
	Limit           uint32
	LimitPresent    bool
	Offset          uint32
	Query           string
	IDs             []int64
	Ordering        []string
	SiteIDs         []int64
	SiteSlugs       []string
	RackIDs         []int64
	DeviceTypeIDs   []int64
	DeviceTypeSlugs []string
	RoleIDs         []int64
	RoleSlugs       []string
	Names           []string
	Statuses        []string
}

func (query ListDevicesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumDevicePageLimit
	}
	return DefaultDevicePageLimit
}

type GetDeviceQuery struct{ ID shared.ID }

type DeviceSortField string

const (
	DeviceSortID          DeviceSortField = "id"
	DeviceSortSite        DeviceSortField = "site"
	DeviceSortRack        DeviceSortField = "rack"
	DeviceSortPosition    DeviceSortField = "position"
	DeviceSortName        DeviceSortField = "name"
	DeviceSortStatus      DeviceSortField = "status"
	DeviceSortCreated     DeviceSortField = "created"
	DeviceSortLastUpdated DeviceSortField = "last_updated"
)

type DeviceSort struct {
	Field      DeviceSortField
	Descending bool
}

func validateListDevicesQuery(query ListDevicesQuery) (DeviceListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumDevicePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseDeviceOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	statuses := make([]dcimdomain.DeviceStatus, 0, len(query.Statuses))
	for _, requested := range query.Statuses {
		status, valid := dcimdomain.ParseDeviceStatus(strings.TrimSpace(requested))
		if !valid {
			violations = append(violations, deviceChoiceViolation("status"))
			continue
		}
		statuses = append(statuses, status)
	}
	if len(violations) > 0 {
		return DeviceListCriteria{}, shared.NewValidationError(violations...)
	}
	return DeviceListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		SiteIDs: append([]int64(nil), query.SiteIDs...), SiteSlugs: trimmedStrings(query.SiteSlugs),
		RackIDs:         append([]int64(nil), query.RackIDs...),
		DeviceTypeIDs:   append([]int64(nil), query.DeviceTypeIDs...),
		DeviceTypeSlugs: trimmedStrings(query.DeviceTypeSlugs),
		RoleIDs:         append([]int64(nil), query.RoleIDs...), RoleSlugs: trimmedStrings(query.RoleSlugs),
		Names: trimmedStrings(query.Names), Statuses: statuses,
	}, nil
}

func parseDeviceOrdering(values []string) ([]DeviceSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []DeviceSort{{Field: DeviceSortName}, {Field: DeviceSortID}}, nil
	}
	var result []DeviceSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseDeviceSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			result = append(result, DeviceSort{Field: field, Descending: descending})
		}
	}
	return result, violations
}

func parseDeviceSortField(value string) (DeviceSortField, bool) {
	field := DeviceSortField(value)
	switch field {
	case DeviceSortID, DeviceSortSite, DeviceSortRack, DeviceSortPosition,
		DeviceSortName, DeviceSortStatus, DeviceSortCreated, DeviceSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}
