package ipam

import (
	"strconv"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type CreatePrefixCommand struct {
	Prefix       Field[string]
	VRF          Field[int64]
	Status       Field[string]
	IsPool       Field[bool]
	MarkUtilized Field[bool]
	Description  Field[string]
	Comments     Field[string]
}

type ReplacePrefixCommand struct {
	ID shared.ID
	CreatePrefixCommand
}

type UpdatePrefixCommand struct {
	ID           shared.ID
	Prefix       Field[string]
	VRF          Field[int64]
	Status       Field[string]
	IsPool       Field[bool]
	MarkUtilized Field[bool]
	Description  Field[string]
	Comments     Field[string]
}

type DeletePrefixCommand struct{ ID shared.ID }

type prefixCommandValues struct {
	prefix       string
	vrfID        *shared.ID
	status       string
	isPool       bool
	markUtilized bool
	description  string
	comments     string
}

func (command CreatePrefixCommand) values() (prefixCommandValues, error) {
	return fullPrefixValues(
		command.Prefix, command.VRF, command.Status, command.IsPool,
		command.MarkUtilized, command.Description, command.Comments,
	)
}

func (command ReplacePrefixCommand) values() (prefixCommandValues, error) {
	return command.CreatePrefixCommand.values()
}

func fullPrefixValues(
	prefix Field[string],
	vrf Field[int64],
	status Field[string],
	isPool Field[bool],
	markUtilized Field[bool],
	description Field[string],
	comments Field[string],
) (prefixCommandValues, error) {
	var violations []shared.FieldViolation
	values := prefixCommandValues{
		prefix:       fullString(&violations, "prefix", prefix, "", true),
		vrfID:        fullPrefixVRFID(&violations, vrf),
		status:       fullString(&violations, "status", status, domainipam.PrefixStatusActive.String(), false),
		isPool:       fullBool(&violations, "is_pool", isPool, false),
		markUtilized: fullBool(&violations, "mark_utilized", markUtilized, false),
		description:  fullString(&violations, "description", description, "", false),
		comments:     fullString(&violations, "comments", comments, "", false),
	}
	if len(violations) > 0 {
		return prefixCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func fullPrefixVRFID(violations *[]shared.FieldViolation, field Field[int64]) *shared.ID {
	switch field.State() {
	case FieldOmitted, FieldNull:
		return nil
	case FieldPresent:
		value, _ := field.Get()
		id := shared.ID(value)
		return &id
	default:
		*violations = append(*violations, shared.FieldViolation{
			Field: "vrf", Reason: "invalid", Description: "A valid object ID is required.",
		})
		return nil
	}
}

type prefixCommandPatch struct {
	prefix       *string
	vrfSet       bool
	vrfID        *shared.ID
	status       *string
	isPool       *bool
	markUtilized *bool
	description  *string
	comments     *string
}

func (command UpdatePrefixCommand) patch() (prefixCommandPatch, error) {
	var violations []shared.FieldViolation
	patch := prefixCommandPatch{
		prefix:       patchString(&violations, "prefix", command.Prefix),
		status:       patchString(&violations, "status", command.Status),
		isPool:       patchBool(&violations, "is_pool", command.IsPool),
		markUtilized: patchBool(&violations, "mark_utilized", command.MarkUtilized),
		description:  patchString(&violations, "description", command.Description),
		comments:     patchString(&violations, "comments", command.Comments),
	}
	switch command.VRF.State() {
	case FieldNull:
		patch.vrfSet = true
	case FieldPresent:
		value, _ := command.VRF.Get()
		id := shared.ID(value)
		patch.vrfSet = true
		patch.vrfID = &id
	}
	if len(violations) > 0 {
		return prefixCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return prefixCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func (patch prefixCommandPatch) empty() bool {
	return patch.prefix == nil && !patch.vrfSet && patch.status == nil &&
		patch.isPool == nil && patch.markUtilized == nil && patch.description == nil &&
		patch.comments == nil
}

func invalidVRFChoice(id shared.ID) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "vrf", Reason: "invalid_choice",
		Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) + "\" - object does not exist.",
	})
}
