package dcim

import (
	"strings"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultRackPageLimit uint32 = 50
	MaximumRackPageLimit uint32 = 1000
)

type ListRacksQuery struct {
	Limit         uint32
	LimitPresent  bool
	Offset        uint32
	Query         string
	IDs           []int64
	Ordering      []string
	SiteIDs       []int64
	SiteSlugs     []string
	Names         []string
	Statuses      []string
	RoleIDs       []int64
	RoleSlugs     []string
	RackTypeIDs   []int64
	RackTypeSlugs []string
}

func (query ListRacksQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumRackPageLimit
	}
	return DefaultRackPageLimit
}

type GetRackQuery struct{ ID shared.ID }

type RackSortField string

const (
	RackSortID          RackSortField = "id"
	RackSortSite        RackSortField = "site"
	RackSortName        RackSortField = "name"
	RackSortFacilityID  RackSortField = "facility_id"
	RackSortStatus      RackSortField = "status"
	RackSortUHeight     RackSortField = "u_height"
	RackSortCreated     RackSortField = "created"
	RackSortLastUpdated RackSortField = "last_updated"
)

type RackSort struct {
	Field      RackSortField
	Descending bool
}

func validateListRacksQuery(query ListRacksQuery) (RackListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumRackPageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseRackOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	statuses := make([]dcimdomain.RackStatus, 0, len(query.Statuses))
	for _, requested := range query.Statuses {
		status, valid := dcimdomain.ParseRackStatus(strings.TrimSpace(requested))
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
			continue
		}
		statuses = append(statuses, status)
	}
	if len(violations) > 0 {
		return RackListCriteria{}, shared.NewValidationError(violations...)
	}
	return RackListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		SiteIDs: append([]int64(nil), query.SiteIDs...), SiteSlugs: trimmedStrings(query.SiteSlugs),
		Names: trimmedStrings(query.Names), Statuses: statuses,
		RoleIDs: append([]int64(nil), query.RoleIDs...), RoleSlugs: trimmedStrings(query.RoleSlugs),
		RackTypeIDs:   append([]int64(nil), query.RackTypeIDs...),
		RackTypeSlugs: trimmedStrings(query.RackTypeSlugs),
	}, nil
}

func parseRackOrdering(values []string) ([]RackSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []RackSort{
			{Field: RackSortSite}, {Field: RackSortName}, {Field: RackSortID},
		}, nil
	}
	var ordering []RackSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseRackSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, RackSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parseRackSortField(value string) (RackSortField, bool) {
	field := RackSortField(value)
	switch field {
	case RackSortID, RackSortSite, RackSortName, RackSortFacilityID, RackSortStatus,
		RackSortUHeight, RackSortCreated, RackSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}
