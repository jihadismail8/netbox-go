// Package changelog owns transport-neutral object-change records.
package changelog

import (
	"fmt"
	"reflect"
	"strings"

	"netbox-go/internal/domain/shared"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Change is staged in the same transaction as its object mutation.
type Change struct {
	ActorID              int64
	ObjectType           string
	ObjectID             shared.ID
	ObjectRepresentation string
	Action               Action
	Before               shared.ObjectSnapshot
	After                shared.ObjectSnapshot
	OccurredAt           shared.Timestamp
}

func NewChange(
	actorID int64,
	objectType string,
	objectID shared.ID,
	representation string,
	action Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
	occurredAt shared.Timestamp,
) (Change, error) {
	objectType = strings.TrimSpace(objectType)
	representation = strings.TrimSpace(representation)
	if actorID <= 0 ||
		objectType == "" ||
		!objectID.IsValid() ||
		representation == "" ||
		occurredAt.IsZero() {
		return Change{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot record an object change with incomplete identity or time data.",
		)
	}
	if !validAction(action) {
		return Change{}, shared.NewError(
			shared.ErrorReasonInternal,
			fmt.Sprintf("Cannot record unsupported object-change action %q.", action),
		)
	}
	if !validSnapshots(action, before, after) {
		return Change{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Object-change snapshots do not match the requested action.",
		)
	}
	if snapshotObjectType(before, after) != objectType {
		return Change{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Object-change snapshot type does not match the recorded object type.",
		)
	}

	return Change{
		ActorID:              actorID,
		ObjectType:           objectType,
		ObjectID:             objectID,
		ObjectRepresentation: representation,
		Action:               action,
		Before:               cloneSnapshot(before),
		After:                cloneSnapshot(after),
		OccurredAt:           occurredAt,
	}, nil
}

func validAction(action Action) bool {
	switch action {
	case ActionCreate, ActionUpdate, ActionDelete:
		return true
	default:
		return false
	}
}

func validSnapshots(
	action Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
) bool {
	switch action {
	case ActionCreate:
		return snapshotIsNil(before) && !snapshotIsNil(after)
	case ActionUpdate:
		return !snapshotIsNil(before) && !snapshotIsNil(after)
	case ActionDelete:
		return !snapshotIsNil(before) && snapshotIsNil(after)
	default:
		return false
	}
}

func cloneSnapshot(snapshot shared.ObjectSnapshot) shared.ObjectSnapshot {
	if snapshotIsNil(snapshot) {
		return nil
	}
	return snapshot.CloneSnapshot()
}

func snapshotObjectType(before, after shared.ObjectSnapshot) string {
	if !snapshotIsNil(before) {
		return before.ObjectType()
	}
	if !snapshotIsNil(after) {
		return after.ObjectType()
	}
	return ""
}

func snapshotIsNil(snapshot shared.ObjectSnapshot) bool {
	if snapshot == nil {
		return true
	}
	reflected := reflect.ValueOf(snapshot)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
