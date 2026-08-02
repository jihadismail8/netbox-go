package ipam

import (
	"context"

	"netbox-go/internal/domain/shared"
)

// IPAddressCascadeChange is the typed audit projection returned when an
// Interface deletion removes its GenericRelation-assigned IP addresses. The
// caller records these child deletions before the Interface parent deletion.
type IPAddressCascadeChange struct {
	ObjectType     string
	ID             shared.ID
	Representation string
	Before         shared.ObjectSnapshot
}

// InterfaceIPAddressCascade is the narrow cross-module port consumed by the
// DCIM Interface service. The caller owns the surrounding transaction.
type InterfaceIPAddressCascade interface {
	DeleteAssignedToInterface(
		context.Context,
		shared.ID,
		shared.Timestamp,
	) ([]IPAddressCascadeChange, error)
}
