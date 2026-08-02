package dcim

import (
	"strings"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultInterfaceTemplatePageLimit uint32 = 50
	MaximumInterfaceTemplatePageLimit uint32 = 1000
)

type ListInterfaceTemplatesQuery struct {
	Limit         uint32
	LimitPresent  bool
	Offset        uint32
	Query         string
	IDs           []int64
	Ordering      []string
	DeviceTypeIDs []int64
	Names         []string
	Types         []string
	Enabled       *bool
	MgmtOnly      *bool
}

func (query ListInterfaceTemplatesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumInterfaceTemplatePageLimit
	}
	return DefaultInterfaceTemplatePageLimit
}

type GetInterfaceTemplateQuery struct{ ID shared.ID }

type InterfaceTemplateSortField string

const (
	InterfaceTemplateSortID          InterfaceTemplateSortField = "id"
	InterfaceTemplateSortDeviceType  InterfaceTemplateSortField = "device_type"
	InterfaceTemplateSortName        InterfaceTemplateSortField = "name"
	InterfaceTemplateSortType        InterfaceTemplateSortField = "type"
	InterfaceTemplateSortCreated     InterfaceTemplateSortField = "created"
	InterfaceTemplateSortLastUpdated InterfaceTemplateSortField = "last_updated"
)

type InterfaceTemplateSort struct {
	Field      InterfaceTemplateSortField
	Descending bool
}

func validateListInterfaceTemplatesQuery(
	query ListInterfaceTemplatesQuery,
) (InterfaceTemplateListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumInterfaceTemplatePageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseInterfaceTemplateOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	types := make([]dcimdomain.InterfaceType, 0, len(query.Types))
	for _, requested := range query.Types {
		parsed, valid := dcimdomain.ParseInterfaceType(requested)
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "type", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
			continue
		}
		types = append(types, parsed)
	}
	if len(violations) > 0 {
		return InterfaceTemplateListCriteria{}, shared.NewValidationError(violations...)
	}
	return InterfaceTemplateListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		DeviceTypeIDs: append([]int64(nil), query.DeviceTypeIDs...),
		Names:         trimmedStrings(query.Names), Types: types,
		Enabled: cloneBool(query.Enabled), MgmtOnly: cloneBool(query.MgmtOnly),
	}, nil
}

func parseInterfaceTemplateOrdering(
	values []string,
) ([]InterfaceTemplateSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []InterfaceTemplateSort{
			{Field: InterfaceTemplateSortDeviceType},
			{Field: InterfaceTemplateSortName},
			{Field: InterfaceTemplateSortID},
		}, nil
	}
	var ordering []InterfaceTemplateSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseInterfaceTemplateSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, InterfaceTemplateSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseInterfaceTemplateSortField(
	value string,
) (InterfaceTemplateSortField, bool) {
	field := InterfaceTemplateSortField(value)
	switch field {
	case InterfaceTemplateSortID,
		InterfaceTemplateSortDeviceType,
		InterfaceTemplateSortName,
		InterfaceTemplateSortType,
		InterfaceTemplateSortCreated,
		InterfaceTemplateSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
