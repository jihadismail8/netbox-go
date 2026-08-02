package shared

import "strconv"

// ID is the transport- and persistence-neutral identifier used by domain
// objects. A zero ID denotes an object which has not been persisted yet.
type ID int64

// ParseID rejects identifiers which cannot refer to a persisted object.
func ParseID(value int64) (ID, error) {
	if value <= 0 {
		return 0, NewValidationError(FieldViolation{
			Field:       "id",
			Reason:      "invalid",
			Description: "ID must be greater than zero.",
		})
	}

	return ID(value), nil
}

// IsValid reports whether the ID can identify a persisted object.
func (id ID) IsValid() bool {
	return id > 0
}

// Int64 returns the primitive representation used at adapter boundaries.
func (id ID) Int64() int64 {
	return int64(id)
}

func (id ID) String() string {
	return strconv.FormatInt(int64(id), 10)
}
