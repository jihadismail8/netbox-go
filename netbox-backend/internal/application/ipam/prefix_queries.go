package ipam

import (
	"strings"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultPrefixPageLimit uint32 = 50
	MaximumPrefixPageLimit uint32 = 1000
)

type ListPrefixesQuery struct {
	Limit         uint32
	LimitPresent  bool
	Offset        uint32
	Query         string
	IDs           []int64
	Ordering      []string
	VRFIDs        []int64
	VRFRDs        []string
	Prefixes      []string
	Family        *int64
	Statuses      []string
	Within        *string
	WithinInclude *string
	Contains      *string
}

func (query ListPrefixesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumPrefixPageLimit
	}
	return DefaultPrefixPageLimit
}

type GetPrefixQuery struct{ ID shared.ID }

type PrefixSortField string

const (
	PrefixSortID          PrefixSortField = "id"
	PrefixSortVRF         PrefixSortField = "vrf"
	PrefixSortPrefix      PrefixSortField = "prefix"
	PrefixSortStatus      PrefixSortField = "status"
	PrefixSortCreated     PrefixSortField = "created"
	PrefixSortLastUpdated PrefixSortField = "last_updated"
)

type PrefixSort struct {
	Field      PrefixSortField
	Descending bool
}

type PrefixNetworkFilter struct {
	Network      domainipam.PrefixNetwork
	ExplicitMask bool
	Valid        bool
}

func validateListPrefixesQuery(query ListPrefixesQuery) (PrefixListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumPrefixPageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parsePrefixOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	statuses := make([]domainipam.PrefixStatus, 0, len(query.Statuses))
	for _, raw := range query.Statuses {
		status, valid := domainipam.ParsePrefixStatus(raw)
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
			continue
		}
		statuses = append(statuses, status)
	}
	prefixes := make([]domainipam.PrefixNetwork, 0, len(query.Prefixes))
	for _, raw := range query.Prefixes {
		if network, _, err := domainipam.ParsePrefixFilter(raw); err == nil {
			prefixes = append(prefixes, network)
		}
	}
	if len(violations) > 0 {
		return PrefixListCriteria{}, shared.NewValidationError(violations...)
	}
	return PrefixListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		VRFIDs: append([]int64(nil), query.VRFIDs...), VRFRDs: trimmedStrings(query.VRFRDs),
		Prefixes: prefixes, PrefixesPresent: len(query.Prefixes) > 0,
		Family: cloneInt64(query.Family), Statuses: statuses,
		Within:        parsePrefixNetworkFilter(query.Within),
		WithinInclude: parsePrefixNetworkFilter(query.WithinInclude),
		Contains:      parsePrefixNetworkFilter(query.Contains),
	}, nil
}

func parsePrefixNetworkFilter(value *string) *PrefixNetworkFilter {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	network, explicit, err := domainipam.ParsePrefixFilter(*value)
	return &PrefixNetworkFilter{Network: network, ExplicitMask: explicit, Valid: err == nil}
}

func parsePrefixOrdering(values []string) ([]PrefixSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []PrefixSort{{Field: PrefixSortVRF}, {Field: PrefixSortPrefix}}, nil
	}
	var ordering []PrefixSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parsePrefixSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, PrefixSort{Field: field, Descending: descending})
		}
	}
	return ordering, violations
}

func parsePrefixSortField(value string) (PrefixSortField, bool) {
	field := PrefixSortField(value)
	switch field {
	case PrefixSortID, PrefixSortVRF, PrefixSortPrefix, PrefixSortStatus, PrefixSortCreated, PrefixSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
