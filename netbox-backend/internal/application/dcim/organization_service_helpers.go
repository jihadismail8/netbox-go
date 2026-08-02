package dcim

import (
	"netbox-go/internal/domain/shared"
)

func sharedIDs(values []int64) []shared.ID {
	ids := make([]shared.ID, len(values))
	for index, value := range values {
		ids[index] = shared.ID(value)
	}
	return ids
}
