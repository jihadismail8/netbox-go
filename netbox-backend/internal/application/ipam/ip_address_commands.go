package ipam

import (
	"strconv"
	"strings"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type CreateIPAddressCommand struct {
	Address            Field[string]
	VRF                Field[int64]
	Status             Field[string]
	Role               Field[string]
	DNSName            Field[string]
	Description        Field[string]
	Comments           Field[string]
	AssignedObjectType Field[string]
	AssignedObjectID   Field[int64]
}

type ReplaceIPAddressCommand struct {
	ID shared.ID
	CreateIPAddressCommand
}

type UpdateIPAddressCommand struct {
	ID                 shared.ID
	Address            Field[string]
	VRF                Field[int64]
	Status             Field[string]
	Role               Field[string]
	DNSName            Field[string]
	Description        Field[string]
	Comments           Field[string]
	AssignedObjectType Field[string]
	AssignedObjectID   Field[int64]
}

type DeleteIPAddressCommand struct{ ID shared.ID }

type AssignIPAddressCommand struct {
	ID          shared.ID
	InterfaceID shared.ID
}

type UnassignIPAddressCommand struct{ ID shared.ID }

type ipAddressCommandValues struct {
	address            string
	vrfID              *shared.ID
	status             string
	role               domainipam.NullableIPAddressRole
	dnsName            string
	description        string
	comments           string
	assignedObjectType Field[string]
	assignedObjectID   Field[int64]
}

func (command CreateIPAddressCommand) values() (ipAddressCommandValues, error) {
	return fullIPAddressValues(
		command.Address,
		command.VRF,
		command.Status,
		command.Role,
		command.DNSName,
		command.Description,
		command.Comments,
		command.AssignedObjectType,
		command.AssignedObjectID,
	)
}

func fullIPAddressValues(
	address Field[string],
	vrf Field[int64],
	status Field[string],
	role Field[string],
	dnsName Field[string],
	description Field[string],
	comments Field[string],
	assignedObjectType Field[string],
	assignedObjectID Field[int64],
) (ipAddressCommandValues, error) {
	var violations []shared.FieldViolation
	values := ipAddressCommandValues{
		address:            fullIPAddressAddress(&violations, address),
		vrfID:              fullIPAddressVRFID(&violations, vrf),
		status:             fullIPAddressStatus(&violations, status),
		role:               fullIPAddressRole(role),
		dnsName:            fullString(&violations, "dns_name", dnsName, "", false),
		description:        fullString(&violations, "description", description, "", false),
		comments:           fullString(&violations, "comments", comments, "", false),
		assignedObjectType: assignedObjectType,
		assignedObjectID:   assignedObjectID,
	}
	if len(violations) > 0 {
		return ipAddressCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func fullIPAddressAddress(
	violations *[]shared.FieldViolation,
	field Field[string],
) string {
	value := fullString(violations, "address", field, "", true)
	if field.State() == FieldPresent && strings.TrimSpace(value) == "" {
		*violations = append(*violations, blankIPAddressViolation("address"))
	}
	return value
}

func fullIPAddressStatus(
	violations *[]shared.FieldViolation,
	field Field[string],
) string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		if value == "" {
			*violations = append(*violations, blankIPAddressViolation("status"))
		}
		return value
	case FieldNull:
		*violations = append(*violations, blankIPAddressViolation("status"))
	}
	return domainipam.IPAddressStatusActive.String()
}

func fullIPAddressVRFID(
	violations *[]shared.FieldViolation,
	field Field[int64],
) *shared.ID {
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

// A choice field distinguishes omission (database NULL/default) from an
// explicit null, which NetBox normalizes to its blank choice.
func fullIPAddressRole(field Field[string]) domainipam.NullableIPAddressRole {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return domainipam.NonNullIPAddressRole(domainipam.IPAddressRole(value))
	case FieldNull:
		return domainipam.NonNullIPAddressRole("")
	default:
		return domainipam.NullIPAddressRole()
	}
}

type ipAddressCommandPatch struct {
	address            *string
	vrfSet             bool
	vrfID              *shared.ID
	status             *string
	role               *domainipam.NullableIPAddressRole
	dnsName            *string
	description        *string
	comments           *string
	assignedObjectType Field[string]
	assignedObjectID   Field[int64]
	assignmentSet      bool
}

func (command ReplaceIPAddressCommand) patch() (ipAddressCommandPatch, error) {
	return buildIPAddressPatch(
		command.Address,
		command.VRF,
		command.Status,
		command.Role,
		command.DNSName,
		command.Description,
		command.Comments,
		command.AssignedObjectType,
		command.AssignedObjectID,
		true,
	)
}

func (command UpdateIPAddressCommand) patch() (ipAddressCommandPatch, error) {
	return buildIPAddressPatch(
		command.Address,
		command.VRF,
		command.Status,
		command.Role,
		command.DNSName,
		command.Description,
		command.Comments,
		command.AssignedObjectType,
		command.AssignedObjectID,
		false,
	)
}

func buildIPAddressPatch(
	address Field[string],
	vrf Field[int64],
	status Field[string],
	role Field[string],
	dnsName Field[string],
	description Field[string],
	comments Field[string],
	assignedObjectType Field[string],
	assignedObjectID Field[int64],
	requireAddress bool,
) (ipAddressCommandPatch, error) {
	var violations []shared.FieldViolation
	if requireAddress && address.State() == FieldOmitted {
		violations = append(violations, requiredViolation("address"))
	}
	patch := ipAddressCommandPatch{
		address:            patchIPAddressAddress(&violations, address),
		status:             patchIPAddressStatus(&violations, status),
		role:               patchIPAddressRole(role),
		dnsName:            patchString(&violations, "dns_name", dnsName),
		description:        patchString(&violations, "description", description),
		comments:           patchString(&violations, "comments", comments),
		assignedObjectType: assignedObjectType,
		assignedObjectID:   assignedObjectID,
		assignmentSet: assignedObjectType.State() != FieldOmitted ||
			assignedObjectID.State() != FieldOmitted,
	}
	switch vrf.State() {
	case FieldNull:
		patch.vrfSet = true
	case FieldPresent:
		value, _ := vrf.Get()
		id := shared.ID(value)
		patch.vrfSet = true
		patch.vrfID = &id
	}
	if len(violations) > 0 {
		return ipAddressCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return ipAddressCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func patchIPAddressAddress(
	violations *[]shared.FieldViolation,
	field Field[string],
) *string {
	value := patchString(violations, "address", field)
	if field.State() == FieldPresent {
		address, _ := field.Get()
		if strings.TrimSpace(address) == "" {
			*violations = append(*violations, blankIPAddressViolation("address"))
		}
	}
	return value
}

func patchIPAddressStatus(
	violations *[]shared.FieldViolation,
	field Field[string],
) *string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		if value == "" {
			*violations = append(*violations, blankIPAddressViolation("status"))
		}
		return &value
	case FieldNull:
		*violations = append(*violations, blankIPAddressViolation("status"))
	}
	return nil
}

func blankIPAddressViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "blank", Description: "This field may not be blank.",
	}
}

func patchIPAddressRole(
	field Field[string],
) *domainipam.NullableIPAddressRole {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		nullable := domainipam.NonNullIPAddressRole(domainipam.IPAddressRole(value))
		return &nullable
	case FieldNull:
		nullable := domainipam.NonNullIPAddressRole("")
		return &nullable
	default:
		return nil
	}
}

func (patch ipAddressCommandPatch) empty() bool {
	return patch.address == nil && !patch.vrfSet && patch.status == nil &&
		patch.role == nil && patch.dnsName == nil && patch.description == nil &&
		patch.comments == nil && !patch.assignmentSet
}

func invalidIPAddressVRFChoice(id shared.ID) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "vrf", Reason: "invalid_choice",
		Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) + "\" - object does not exist.",
	})
}

func invalidIPAddressInterfaceChoice(id shared.ID) error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "assigned_object_id", Reason: "invalid_choice",
		Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) + "\" - object does not exist.",
	})
}
