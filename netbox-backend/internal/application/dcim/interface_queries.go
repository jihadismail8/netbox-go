package dcim

import (
	"strings"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultInterfacePageLimit uint32 = 50
	MaximumInterfacePageLimit uint32 = 1000
)

type ListInterfacesQuery struct {
	Limit        uint32
	LimitPresent bool
	Offset       uint32
	Query        string
	IDs          []int64
	Ordering     []string
	DeviceIDs    []int64
	DeviceNames  []string
	Names        []string
	Types        []string
	Enabled      *bool
	MgmtOnly     *bool
}

func (query ListInterfacesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumInterfacePageLimit
	}
	return DefaultInterfacePageLimit
}

type GetInterfaceQuery struct{ ID shared.ID }

type InterfaceSortField string

const (
	InterfaceSortID          InterfaceSortField = "id"
	InterfaceSortDevice      InterfaceSortField = "device"
	InterfaceSortName        InterfaceSortField = "name"
	InterfaceSortType        InterfaceSortField = "type"
	InterfaceSortCreated     InterfaceSortField = "created"
	InterfaceSortLastUpdated InterfaceSortField = "last_updated"
)

type InterfaceSort struct {
	Field      InterfaceSortField
	Descending bool
}

func validateListInterfacesQuery(query ListInterfacesQuery) (InterfaceListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumInterfacePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseInterfaceOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	types := make([]dcimdomain.InterfaceType, 0, len(query.Types))
	for _, requested := range query.Types {
		parsed, valid := dcimdomain.ParseInterfaceType(strings.TrimSpace(requested))
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "type", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
			continue
		}
		types = append(types, parsed)
	}
	if len(violations) > 0 {
		return InterfaceListCriteria{}, shared.NewValidationError(violations...)
	}
	return InterfaceListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		DeviceIDs:   append([]int64(nil), query.DeviceIDs...),
		DeviceNames: trimmedStrings(query.DeviceNames),
		Names:       trimmedStrings(query.Names), Types: types,
		Enabled: cloneBool(query.Enabled), MgmtOnly: cloneBool(query.MgmtOnly),
	}, nil
}

func parseInterfaceOrdering(values []string) ([]InterfaceSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []InterfaceSort{
			{Field: InterfaceSortDevice},
			{Field: InterfaceSortName},
			{Field: InterfaceSortID},
		}, nil
	}
	var ordering []InterfaceSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseInterfaceSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, InterfaceSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseInterfaceSortField(value string) (InterfaceSortField, bool) {
	field := InterfaceSortField(value)
	switch field {
	case InterfaceSortID, InterfaceSortDevice, InterfaceSortName,
		InterfaceSortType, InterfaceSortCreated, InterfaceSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}
