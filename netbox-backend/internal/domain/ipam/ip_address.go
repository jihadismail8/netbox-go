package ipam

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"
	"unicode/utf8"

	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	IPAddressObjectType           = "ipam.ipaddress"
	IPAddressDescriptionMaxLength = 200
	IPAddressDNSNameMaxLength     = 255
	IPAddressAssignmentType       = "dcim.interface"
)

// HostAddress preserves both host bits and prefix length. Unlike Prefix, an
// IPAddress must never silently mask its host address.
type HostAddress struct {
	prefix netip.Prefix
}

func ParseHostAddress(value string) (HostAddress, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
	if err != nil {
		return HostAddress{}, shared.NewValidationError(shared.FieldViolation{
			Field: "address", Reason: "invalid",
			Description: "Enter a valid IPv4 or IPv6 address with mask.",
		})
	}
	if prefix.Bits() == 0 {
		return HostAddress{}, shared.NewValidationError(shared.FieldViolation{
			Field: "address", Reason: "invalid",
			Description: "Cannot create IP address with /0 mask.",
		})
	}
	return HostAddress{prefix: prefix}, nil
}

func (address HostAddress) Prefix() netip.Prefix { return address.prefix }
func (address HostAddress) Host() netip.Addr     { return address.prefix.Addr() }
func (address HostAddress) String() string       { return address.prefix.String() }
func (address HostAddress) Family() uint32 {
	if address.prefix.Addr().Is4() {
		return 4
	}
	return 6
}

func (address HostAddress) Compare(other HostAddress) int {
	if compared := address.Host().Compare(other.Host()); compared != 0 {
		return compared
	}
	switch {
	case address.prefix.Bits() < other.prefix.Bits():
		return -1
	case address.prefix.Bits() > other.prefix.Bits():
		return 1
	default:
		return 0
	}
}

type IPAddressStatus string

const (
	IPAddressStatusActive     IPAddressStatus = "active"
	IPAddressStatusReserved   IPAddressStatus = "reserved"
	IPAddressStatusDeprecated IPAddressStatus = "deprecated"
	IPAddressStatusDHCP       IPAddressStatus = "dhcp"
	IPAddressStatusSLAAC      IPAddressStatus = "slaac"
)

func ParseIPAddressStatus(value string) (IPAddressStatus, bool) {
	status := IPAddressStatus(strings.TrimSpace(value))
	switch status {
	case IPAddressStatusActive, IPAddressStatusReserved, IPAddressStatusDeprecated,
		IPAddressStatusDHCP, IPAddressStatusSLAAC:
		return status, true
	default:
		return "", false
	}
}

func (status IPAddressStatus) String() string { return string(status) }

type IPAddressRole string

const (
	IPAddressRoleLoopback  IPAddressRole = "loopback"
	IPAddressRoleSecondary IPAddressRole = "secondary"
	IPAddressRoleAnycast   IPAddressRole = "anycast"
	IPAddressRoleVIP       IPAddressRole = "vip"
	IPAddressRoleVRRP      IPAddressRole = "vrrp"
	IPAddressRoleHSRP      IPAddressRole = "hsrp"
	IPAddressRoleGLBP      IPAddressRole = "glbp"
	IPAddressRoleCARP      IPAddressRole = "carp"
)

func ParseIPAddressRole(value string) (IPAddressRole, bool) {
	role := IPAddressRole(strings.TrimSpace(value))
	switch role {
	case "", IPAddressRoleLoopback, IPAddressRoleSecondary, IPAddressRoleAnycast,
		IPAddressRoleVIP, IPAddressRoleVRRP, IPAddressRoleHSRP, IPAddressRoleGLBP,
		IPAddressRoleCARP:
		return role, true
	default:
		return "", false
	}
}

func (role IPAddressRole) String() string { return string(role) }

func (role IPAddressRole) AllowsDuplicateHost() bool {
	switch role {
	case IPAddressRoleAnycast, IPAddressRoleVIP, IPAddressRoleVRRP,
		IPAddressRoleHSRP, IPAddressRoleGLBP, IPAddressRoleCARP:
		return true
	default:
		return false
	}
}

type NullableIPAddressRole struct {
	value IPAddressRole
	valid bool
}

func NullIPAddressRole() NullableIPAddressRole { return NullableIPAddressRole{} }
func NonNullIPAddressRole(value IPAddressRole) NullableIPAddressRole {
	return NullableIPAddressRole{value: value, valid: true}
}
func (nullable NullableIPAddressRole) Get() (IPAddressRole, bool) {
	return nullable.value, nullable.valid
}
func (nullable NullableIPAddressRole) IsNull() bool { return !nullable.valid }

// InterfaceAssignment retains the typed Interface and its DeviceReference.
// It is copied into IPAddress state so no generic content-type map crosses the
// domain boundary.
type InterfaceAssignment struct {
	networkInterface domaindcim.Interface
}

func NewInterfaceAssignment(
	networkInterface *domaindcim.Interface,
) (InterfaceAssignment, error) {
	if networkInterface == nil || !networkInterface.ID().IsValid() ||
		!networkInterface.Device().Valid() || strings.TrimSpace(networkInterface.Display()) == "" {
		return InterfaceAssignment{}, shared.NewError(
			shared.ErrorReasonInternal, "Cannot construct an invalid Interface assignment.",
		)
	}
	return InterfaceAssignment{networkInterface: *networkInterface}, nil
}

func (assignment InterfaceAssignment) Interface() domaindcim.Interface {
	return assignment.networkInterface
}
func (assignment InterfaceAssignment) ID() shared.ID {
	return assignment.networkInterface.ID()
}
func (assignment InterfaceAssignment) Display() string {
	return assignment.networkInterface.Display()
}
func (assignment InterfaceAssignment) Device() domaindcim.DeviceReference {
	return assignment.networkInterface.Device()
}
func (assignment InterfaceAssignment) Valid() bool {
	return assignment.networkInterface.ID().IsValid() &&
		assignment.networkInterface.Device().Valid() &&
		strings.TrimSpace(assignment.networkInterface.Display()) != ""
}

type NullableInterfaceAssignment struct {
	assignment InterfaceAssignment
	valid      bool
}

func NullInterfaceAssignment() NullableInterfaceAssignment {
	return NullableInterfaceAssignment{}
}
func NonNullInterfaceAssignment(
	assignment InterfaceAssignment,
) NullableInterfaceAssignment {
	return NullableInterfaceAssignment{assignment: assignment, valid: true}
}
func (nullable NullableInterfaceAssignment) Get() (InterfaceAssignment, bool) {
	return nullable.assignment, nullable.valid
}
func (nullable NullableInterfaceAssignment) IsNull() bool { return !nullable.valid }

type IPAddressValues struct {
	Address     string
	VRF         NullableVRFReference
	Status      string
	Role        NullableIPAddressRole
	DNSName     string
	Description string
	Comments    string
	Assignment  NullableInterfaceAssignment
}

type IPAddressPatch struct {
	Address     *string
	VRF         *NullableVRFReference
	Status      *string
	Role        *NullableIPAddressRole
	DNSName     *string
	Description *string
	Comments    *string
	Assignment  *NullableInterfaceAssignment
}

func (patch IPAddressPatch) Empty() bool {
	return patch.Address == nil && patch.VRF == nil && patch.Status == nil &&
		patch.Role == nil && patch.DNSName == nil && patch.Description == nil &&
		patch.Comments == nil && patch.Assignment == nil
}

type IPAddressState struct {
	ID          shared.ID
	Address     string
	VRF         NullableVRFReference
	Status      string
	Role        NullableIPAddressRole
	DNSName     string
	Description string
	Comments    string
	Assignment  NullableInterfaceAssignment
	Created     shared.Timestamp
	LastUpdated shared.Timestamp
}

type IPAddressSnapshot struct {
	Address            string
	VRFID              *shared.ID
	Status             string
	Role               *string
	DNSName            string
	Description        string
	Comments           string
	AssignedObjectType *string
	AssignedObjectID   *shared.ID
}

func (IPAddressSnapshot) ObjectType() string { return IPAddressObjectType }
func (snapshot IPAddressSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	snapshot.VRFID = cloneSharedID(snapshot.VRFID)
	snapshot.Role = cloneIPAddressString(snapshot.Role)
	snapshot.AssignedObjectType = cloneIPAddressString(snapshot.AssignedObjectType)
	snapshot.AssignedObjectID = cloneSharedID(snapshot.AssignedObjectID)
	return snapshot
}

type IPAddress struct {
	id          shared.ID
	address     HostAddress
	vrf         NullableVRFReference
	status      IPAddressStatus
	role        NullableIPAddressRole
	dnsName     string
	description string
	comments    string
	assignment  NullableInterfaceAssignment
	created     shared.Timestamp
	lastUpdated shared.Timestamp
}

func NewIPAddress(values IPAddressValues, now shared.Timestamp) (*IPAddress, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateIPAddressValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return ipAddressFromNormalized(normalized, now, now), nil
}

func RestoreIPAddress(state IPAddressState) (*IPAddress, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Cannot restore invalid IPAddress identity or timestamps.",
		)
	}
	normalized, violations := validateIPAddressValues(IPAddressValues{
		Address: state.Address, VRF: state.VRF, Status: state.Status, Role: state.Role,
		DNSName: state.DNSName, Description: state.Description, Comments: state.Comments,
		Assignment: state.Assignment,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Persisted IPAddress violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	address := ipAddressFromNormalized(normalized, state.Created, state.LastUpdated)
	address.id = state.ID
	return address, nil
}

type normalizedIPAddressValues struct {
	address     HostAddress
	vrf         NullableVRFReference
	status      IPAddressStatus
	role        NullableIPAddressRole
	dnsName     string
	description string
	comments    string
	assignment  NullableInterfaceAssignment
}

func ipAddressFromNormalized(
	values normalizedIPAddressValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *IPAddress {
	return &IPAddress{
		address: values.address, vrf: values.vrf, status: values.status,
		role: values.role, dnsName: values.dnsName, description: values.description,
		comments: values.comments, assignment: values.assignment,
		created: created, lastUpdated: lastUpdated,
	}
}

func (address *IPAddress) AssignID(id shared.ID) error {
	if address == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid IPAddress ID.")
	}
	if address.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned IPAddress ID.")
	}
	address.id = id
	return nil
}

func (address *IPAddress) Replace(values IPAddressValues, now shared.Timestamp) error {
	if address == nil || now.IsZero() {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot replace IPAddress with invalid state or time.",
		)
	}
	normalized, violations := validateIPAddressValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	id, created := address.id, address.created
	*address = *ipAddressFromNormalized(normalized, created, now)
	address.id = id
	return nil
}

func (address *IPAddress) ApplyPatch(patch IPAddressPatch, now shared.Timestamp) error {
	if address == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil IPAddress.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	values := address.Values()
	if patch.Address != nil {
		values.Address = *patch.Address
	}
	if patch.VRF != nil {
		values.VRF = *patch.VRF
	}
	if patch.Status != nil {
		values.Status = *patch.Status
	}
	if patch.Role != nil {
		values.Role = *patch.Role
	}
	if patch.DNSName != nil {
		values.DNSName = *patch.DNSName
	}
	if patch.Description != nil {
		values.Description = *patch.Description
	}
	if patch.Comments != nil {
		values.Comments = *patch.Comments
	}
	if patch.Assignment != nil {
		values.Assignment = *patch.Assignment
	}
	return address.Replace(values, now)
}

func (address IPAddress) ID() shared.ID                           { return address.id }
func (address IPAddress) Address() HostAddress                    { return address.address }
func (address IPAddress) VRF() NullableVRFReference               { return address.vrf }
func (address IPAddress) Status() IPAddressStatus                 { return address.status }
func (address IPAddress) Role() NullableIPAddressRole             { return address.role }
func (address IPAddress) DNSName() string                         { return address.dnsName }
func (address IPAddress) Description() string                     { return address.description }
func (address IPAddress) Comments() string                        { return address.comments }
func (address IPAddress) Assignment() NullableInterfaceAssignment { return address.assignment }
func (address IPAddress) Created() shared.Timestamp               { return address.created }
func (address IPAddress) LastUpdated() shared.Timestamp           { return address.lastUpdated }
func (address IPAddress) Family() uint32                          { return address.address.Family() }
func (address IPAddress) Display() string                         { return address.address.String() }

func (address IPAddress) Values() IPAddressValues {
	return IPAddressValues{
		Address: address.address.String(), VRF: address.vrf, Status: address.status.String(),
		Role: address.role, DNSName: address.dnsName, Description: address.description,
		Comments: address.comments, Assignment: address.assignment,
	}
}

func (address IPAddress) State() IPAddressState {
	values := address.Values()
	return IPAddressState{
		ID: address.id, Address: values.Address, VRF: values.VRF, Status: values.Status,
		Role: values.Role, DNSName: values.DNSName, Description: values.Description,
		Comments: values.Comments, Assignment: values.Assignment,
		Created: address.created, LastUpdated: address.lastUpdated,
	}
}

func (address IPAddress) Snapshot() IPAddressSnapshot {
	var vrfID *shared.ID
	if reference, present := address.vrf.Get(); present {
		value := reference.ID()
		vrfID = &value
	}
	var role *string
	if value, present := address.role.Get(); present {
		text := value.String()
		role = &text
	}
	var assignmentType *string
	var assignmentID *shared.ID
	if value, present := address.assignment.Get(); present {
		kind := IPAddressAssignmentType
		id := value.ID()
		assignmentType, assignmentID = &kind, &id
	}
	return IPAddressSnapshot{
		Address: address.address.String(), VRFID: vrfID, Status: address.status.String(),
		Role: role, DNSName: address.dnsName, Description: address.description,
		Comments: address.comments, AssignedObjectType: assignmentType,
		AssignedObjectID: assignmentID,
	}
}

var ipAddressDNSNamePattern = regexp.MustCompile(
	`^([0-9A-Za-z_-]+|\*)(\.[0-9A-Za-z_-]+)*\.?$`,
)

func validateIPAddressValues(
	values IPAddressValues,
) (normalizedIPAddressValues, []shared.FieldViolation) {
	var violations []shared.FieldViolation
	hostAddress, err := ParseHostAddress(values.Address)
	if err != nil {
		violations = append(violations, shared.ViolationsOf(err)...)
	}
	if reference, present := values.VRF.Get(); present && !reference.Valid() {
		violations = append(violations, shared.FieldViolation{
			Field: "vrf", Reason: "invalid", Description: "Invalid VRF reference.",
		})
	}
	status, validStatus := ParseIPAddressStatus(values.Status)
	if !validStatus {
		violations = append(violations, shared.FieldViolation{
			Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	} else if status == IPAddressStatusSLAAC &&
		hostAddress.Prefix().IsValid() && !hostAddress.Prefix().Addr().Is6() {
		violations = append(violations, shared.FieldViolation{
			Field: "status", Reason: "invalid_choice",
			Description: "Only IPv6 addresses can be assigned SLAAC status",
		})
	}
	role := values.Role
	if value, present := role.Get(); present {
		parsed, valid := ParseIPAddressRole(value.String())
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "role", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
		} else {
			role = NonNullIPAddressRole(parsed)
		}
	}
	dnsName := strings.TrimSpace(values.DNSName)
	if utf8.RuneCountInString(dnsName) > IPAddressDNSNameMaxLength {
		violations = append(violations, maxLengthViolation("dns_name", IPAddressDNSNameMaxLength))
	} else if dnsName != "" && !ipAddressDNSNamePattern.MatchString(dnsName) {
		violations = append(violations, shared.FieldViolation{
			Field: "dns_name", Reason: "invalid",
			Description: "Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names.",
		})
	}
	description := strings.TrimSpace(values.Description)
	comments := strings.TrimSpace(values.Comments)
	if utf8.RuneCountInString(description) > IPAddressDescriptionMaxLength {
		violations = append(
			violations, maxLengthViolation("description", IPAddressDescriptionMaxLength),
		)
	}
	if assignment, present := values.Assignment.Get(); present {
		if !assignment.Valid() {
			violations = append(violations, shared.FieldViolation{
				Field: "assigned_object_id", Reason: "invalid_choice",
				Description: "The Interface does not exist.",
			})
		} else if reason := unusableInterfaceAddressReason(hostAddress); reason != "" {
			violations = append(violations, shared.FieldViolation{
				Field: "non_field_errors", Reason: "invalid", Description: reason,
			})
		}
	}
	return normalizedIPAddressValues{
		address: hostAddress, vrf: values.VRF, status: status, role: role,
		dnsName: strings.ToLower(dnsName), description: description, comments: comments,
		assignment: values.Assignment,
	}, violations
}

func unusableInterfaceAddressReason(address HostAddress) string {
	prefix := address.Prefix()
	if !prefix.IsValid() {
		return ""
	}
	if prefix.Addr().Is6() {
		if prefix.Bits() < 127 && prefix.Addr() == prefix.Masked().Addr() {
			return fmt.Sprintf(
				"%s is a network ID, which may not be assigned to an interface.",
				prefix.Addr(),
			)
		}
		return ""
	}
	if !prefix.Addr().Is4() || prefix.Bits() >= 31 {
		return ""
	}
	addressBytes := prefix.Addr().As4()
	networkBytes := prefix.Masked().Addr().As4()
	hostBits := uint32(32 - prefix.Bits())
	mask := uint32(1<<hostBits) - 1
	value := uint32(addressBytes[0])<<24 | uint32(addressBytes[1])<<16 |
		uint32(addressBytes[2])<<8 | uint32(addressBytes[3])
	network := uint32(networkBytes[0])<<24 | uint32(networkBytes[1])<<16 |
		uint32(networkBytes[2])<<8 | uint32(networkBytes[3])
	switch value {
	case network:
		return fmt.Sprintf(
			"%s is a network ID, which may not be assigned to an interface.",
			prefix.Addr(),
		)
	case network | mask:
		return fmt.Sprintf(
			"%s is a broadcast address, which may not be assigned to an interface.",
			prefix.Addr(),
		)
	default:
		return ""
	}
}

func cloneSharedID(value *shared.ID) *shared.ID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneIPAddressString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
