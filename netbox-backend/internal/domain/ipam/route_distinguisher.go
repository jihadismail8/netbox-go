package ipam

import (
	"strings"
	"unicode/utf8"

	"netbox-go/internal/domain/shared"
)

const VRFRouteDistinguisherMaxLength = 21

// RouteDistinguisher is the immutable, non-null value of VRF.rd. The pinned
// NetBox model deliberately accepts any string up to 21 characters; it does
// not impose a stricter RFC-shape validator. Nullability is represented by
// NullableRouteDistinguisher rather than by overloading an empty string.
type RouteDistinguisher struct {
	value string
}

func ParseRouteDistinguisher(value string) (RouteDistinguisher, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > VRFRouteDistinguisherMaxLength {
		return RouteDistinguisher{}, shared.NewValidationError(shared.FieldViolation{
			Field:       "rd",
			Reason:      "max_length",
			Description: "Ensure this field has no more than 21 characters.",
		})
	}
	return RouteDistinguisher{value: value}, nil
}

func (rd RouteDistinguisher) String() string { return rd.value }

// NullableRouteDistinguisher preserves the difference between SQL/public
// null and a present value, including a deliberately present blank string.
// Its zero value is null.
type NullableRouteDistinguisher struct {
	value RouteDistinguisher
	valid bool
}

func NullRouteDistinguisher() NullableRouteDistinguisher {
	return NullableRouteDistinguisher{}
}

func NonNullRouteDistinguisher(value RouteDistinguisher) NullableRouteDistinguisher {
	return NullableRouteDistinguisher{value: value, valid: true}
}

func (nullable NullableRouteDistinguisher) Get() (RouteDistinguisher, bool) {
	return nullable.value, nullable.valid
}

func (nullable NullableRouteDistinguisher) IsNull() bool { return !nullable.valid }
