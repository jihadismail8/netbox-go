package dcim

import (
	"fmt"
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	InterfaceObjectType           = "dcim.interface"
	InterfaceNameMaxLength        = 64
	InterfaceLabelMaxLength       = 64
	InterfaceDescriptionMaxLength = 200
	InterfaceMTUMin               = uint32(1)
	InterfaceMTUMax               = uint32(65536)
	InterfaceSpeedMax             = uint64(2147483647)
)

// DeviceReference is the immutable relationship projection held by an
// Interface. Name is retained independently from Display because it is a
// declared Interface list filter.
type DeviceReference struct {
	id      shared.ID
	name    DeviceNullable[string]
	display string
}

func NewDeviceReference(
	id shared.ID,
	name DeviceNullable[string],
	display string,
) (DeviceReference, error) {
	display = strings.TrimSpace(display)
	if !id.IsValid() || display == "" {
		return DeviceReference{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid Device reference.",
		)
	}
	name = normalizeDeviceNullableString(name)
	return DeviceReference{id: id, name: name, display: display}, nil
}

func (reference DeviceReference) ID() shared.ID                { return reference.id }
func (reference DeviceReference) Name() DeviceNullable[string] { return reference.name }
func (reference DeviceReference) Display() string              { return reference.display }
func (reference DeviceReference) Valid() bool {
	return reference.id.IsValid() && reference.display != ""
}

type InterfaceDuplex string

const (
	InterfaceDuplexHalf InterfaceDuplex = "half"
	InterfaceDuplexFull InterfaceDuplex = "full"
	InterfaceDuplexAuto InterfaceDuplex = "auto"
)

func ParseInterfaceDuplex(value string) (InterfaceDuplex, bool) {
	duplex := InterfaceDuplex(value)
	switch duplex {
	case "", InterfaceDuplexHalf, InterfaceDuplexFull, InterfaceDuplexAuto:
		return duplex, true
	default:
		return "", false
	}
}

func (duplex InterfaceDuplex) String() string { return string(duplex) }

type InterfaceValues struct {
	Device      DeviceReference
	Name        string
	Label       string
	Type        string
	Enabled     bool
	MgmtOnly    bool
	MTU         DeviceNullable[uint32]
	Speed       DeviceNullable[uint64]
	Duplex      DeviceNullable[string]
	Description string
}

type InterfacePatch struct {
	Device      *DeviceReference
	Name        *string
	Label       *string
	Type        *string
	Enabled     *bool
	MgmtOnly    *bool
	MTU         *DeviceNullable[uint32]
	Speed       *DeviceNullable[uint64]
	Duplex      *DeviceNullable[string]
	Description *string
}

func (patch InterfacePatch) Empty() bool {
	return patch.Device == nil && patch.Name == nil && patch.Label == nil &&
		patch.Type == nil && patch.Enabled == nil && patch.MgmtOnly == nil &&
		patch.MTU == nil && patch.Speed == nil && patch.Duplex == nil &&
		patch.Description == nil
}

type InterfaceState struct {
	ID             shared.ID
	Device         DeviceReference
	Name           string
	Label          string
	Type           string
	Enabled        bool
	MgmtOnly       bool
	MTU            DeviceNullable[uint32]
	Speed          DeviceNullable[uint64]
	Duplex         DeviceNullable[string]
	Description    string
	Created        shared.Timestamp
	LastUpdated    shared.Timestamp
	IPAddressCount uint64
}

type InterfaceSnapshot struct {
	DeviceID    shared.ID
	Name        string
	Label       string
	Type        string
	Enabled     bool
	MgmtOnly    bool
	MTU         *uint32
	Speed       *uint64
	Duplex      *string
	Description string
}

func (InterfaceSnapshot) ObjectType() string { return InterfaceObjectType }
func (snapshot InterfaceSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	if snapshot.MTU != nil {
		value := *snapshot.MTU
		snapshot.MTU = &value
	}
	if snapshot.Speed != nil {
		value := *snapshot.Speed
		snapshot.Speed = &value
	}
	snapshot.Duplex = cloneString(snapshot.Duplex)
	return snapshot
}

type Interface struct {
	id             shared.ID
	device         DeviceReference
	name           string
	label          string
	interfaceType  InterfaceType
	enabled        bool
	mgmtOnly       bool
	mtu            DeviceNullable[uint32]
	speed          DeviceNullable[uint64]
	duplex         DeviceNullable[InterfaceDuplex]
	description    string
	created        shared.Timestamp
	lastUpdated    shared.Timestamp
	ipAddressCount uint64
}

func NewInterface(values InterfaceValues, now shared.Timestamp) (*Interface, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateInterfaceValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return interfaceFromNormalized(normalized, now, now), nil
}

func RestoreInterface(state InterfaceState) (*Interface, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot restore invalid Interface identity or timestamps.",
		)
	}
	normalized, violations := validateInterfaceValues(InterfaceValues{
		Device: state.Device, Name: state.Name, Label: state.Label, Type: state.Type,
		Enabled: state.Enabled, MgmtOnly: state.MgmtOnly, MTU: state.MTU,
		Speed: state.Speed, Duplex: state.Duplex, Description: state.Description,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted Interface violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	networkInterface := interfaceFromNormalized(normalized, state.Created, state.LastUpdated)
	networkInterface.id = state.ID
	networkInterface.ipAddressCount = state.IPAddressCount
	return networkInterface, nil
}

type normalizedInterfaceValues struct {
	device        DeviceReference
	name          string
	label         string
	interfaceType InterfaceType
	enabled       bool
	mgmtOnly      bool
	mtu           DeviceNullable[uint32]
	speed         DeviceNullable[uint64]
	duplex        DeviceNullable[InterfaceDuplex]
	description   string
}

func interfaceFromNormalized(
	values normalizedInterfaceValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *Interface {
	return &Interface{
		device: values.device, name: values.name, label: values.label,
		interfaceType: values.interfaceType, enabled: values.enabled,
		mgmtOnly: values.mgmtOnly, mtu: values.mtu, speed: values.speed,
		duplex: values.duplex, description: values.description,
		created: created, lastUpdated: lastUpdated,
	}
}

func (networkInterface *Interface) AssignID(id shared.ID) error {
	if networkInterface == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid Interface ID.")
	}
	if networkInterface.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned Interface ID.")
	}
	networkInterface.id = id
	return nil
}

func (networkInterface *Interface) Replace(
	values InterfaceValues,
	now shared.Timestamp,
) error {
	if networkInterface == nil || now.IsZero() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot replace Interface with invalid state or time.",
		)
	}
	normalized, violations := validateInterfaceValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	if networkInterface.device.Valid() &&
		normalized.device.ID() != networkInterface.device.ID() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "device", Reason: "immutable",
			Description: "An Interface cannot be moved to another Device.",
		})
	}
	id, created, count := networkInterface.id, networkInterface.created, networkInterface.ipAddressCount
	*networkInterface = *interfaceFromNormalized(normalized, created, now)
	networkInterface.id, networkInterface.ipAddressCount = id, count
	return nil
}

func (networkInterface *Interface) ApplyPatch(
	patch InterfacePatch,
	now shared.Timestamp,
) error {
	if networkInterface == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil Interface.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	values := networkInterface.Values()
	if patch.Device != nil {
		values.Device = *patch.Device
	}
	setString(&values.Name, patch.Name)
	setString(&values.Label, patch.Label)
	setString(&values.Type, patch.Type)
	if patch.Enabled != nil {
		values.Enabled = *patch.Enabled
	}
	if patch.MgmtOnly != nil {
		values.MgmtOnly = *patch.MgmtOnly
	}
	if patch.MTU != nil {
		values.MTU = *patch.MTU
	}
	if patch.Speed != nil {
		values.Speed = *patch.Speed
	}
	if patch.Duplex != nil {
		values.Duplex = *patch.Duplex
	}
	setString(&values.Description, patch.Description)
	return networkInterface.Replace(values, now)
}

func (networkInterface Interface) ID() shared.ID                 { return networkInterface.id }
func (networkInterface Interface) Device() DeviceReference       { return networkInterface.device }
func (networkInterface Interface) Name() string                  { return networkInterface.name }
func (networkInterface Interface) Label() string                 { return networkInterface.label }
func (networkInterface Interface) Type() InterfaceType           { return networkInterface.interfaceType }
func (networkInterface Interface) Enabled() bool                 { return networkInterface.enabled }
func (networkInterface Interface) MgmtOnly() bool                { return networkInterface.mgmtOnly }
func (networkInterface Interface) MTU() DeviceNullable[uint32]   { return networkInterface.mtu }
func (networkInterface Interface) Speed() DeviceNullable[uint64] { return networkInterface.speed }
func (networkInterface Interface) Duplex() DeviceNullable[InterfaceDuplex] {
	return networkInterface.duplex
}
func (networkInterface Interface) Description() string       { return networkInterface.description }
func (networkInterface Interface) Created() shared.Timestamp { return networkInterface.created }
func (networkInterface Interface) LastUpdated() shared.Timestamp {
	return networkInterface.lastUpdated
}
func (networkInterface Interface) IPAddressCount() uint64 {
	return networkInterface.ipAddressCount
}
func (networkInterface Interface) Display() string {
	if networkInterface.label != "" {
		return fmt.Sprintf("%s (%s)", networkInterface.name, networkInterface.label)
	}
	return networkInterface.name
}

func (networkInterface Interface) Values() InterfaceValues {
	var duplex DeviceNullable[string]
	if value, present := networkInterface.duplex.Get(); present {
		duplex = NonNullDeviceValue(value.String())
	}
	return InterfaceValues{
		Device: networkInterface.device, Name: networkInterface.name,
		Label: networkInterface.label, Type: networkInterface.interfaceType.String(),
		Enabled: networkInterface.enabled, MgmtOnly: networkInterface.mgmtOnly,
		MTU: networkInterface.mtu, Speed: networkInterface.speed, Duplex: duplex,
		Description: networkInterface.description,
	}
}

func (networkInterface Interface) State() InterfaceState {
	values := networkInterface.Values()
	return InterfaceState{
		ID: networkInterface.id, Device: values.Device, Name: values.Name,
		Label: values.Label, Type: values.Type, Enabled: values.Enabled,
		MgmtOnly: values.MgmtOnly, MTU: values.MTU, Speed: values.Speed,
		Duplex: values.Duplex, Description: values.Description,
		Created: networkInterface.created, LastUpdated: networkInterface.lastUpdated,
		IPAddressCount: networkInterface.ipAddressCount,
	}
}

func (networkInterface Interface) Snapshot() InterfaceSnapshot {
	var mtu *uint32
	if value, present := networkInterface.mtu.Get(); present {
		mtu = &value
	}
	var speed *uint64
	if value, present := networkInterface.speed.Get(); present {
		speed = &value
	}
	var duplex *string
	if value, present := networkInterface.duplex.Get(); present {
		text := value.String()
		duplex = &text
	}
	return InterfaceSnapshot{
		DeviceID: networkInterface.device.ID(), Name: networkInterface.name,
		Label: networkInterface.label, Type: networkInterface.interfaceType.String(),
		Enabled: networkInterface.enabled, MgmtOnly: networkInterface.mgmtOnly,
		MTU: mtu, Speed: speed, Duplex: duplex,
		Description: networkInterface.description,
	}
}

func validateInterfaceValues(
	values InterfaceValues,
) (normalizedInterfaceValues, []shared.FieldViolation) {
	var violations []shared.FieldViolation
	if !values.Device.Valid() {
		violations = append(violations, relationViolation("device"))
	}
	name := strings.TrimSpace(values.Name)
	label := strings.TrimSpace(values.Label)
	description := strings.TrimSpace(values.Description)
	validateRequiredLength(&violations, "name", name, InterfaceNameMaxLength)
	validateOptionalLength(&violations, "label", label, InterfaceLabelMaxLength)
	validateOptionalLength(
		&violations,
		"description",
		description,
		InterfaceDescriptionMaxLength,
	)
	interfaceType, validType := ParseInterfaceType(strings.TrimSpace(values.Type))
	if !validType {
		violations = append(violations, shared.FieldViolation{
			Field: "type", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if mtu, present := values.MTU.Get(); present &&
		(mtu < InterfaceMTUMin || mtu > InterfaceMTUMax) {
		violations = append(violations, shared.FieldViolation{
			Field: "mtu", Reason: "out_of_range",
			Description: "Ensure this value is between 1 and 65536.",
		})
	}
	if speed, present := values.Speed.Get(); present &&
		(speed == 0 || speed > InterfaceSpeedMax) {
		violations = append(violations, shared.FieldViolation{
			Field: "speed", Reason: "out_of_range",
			Description: "Ensure this value is between 1 and 2147483647.",
		})
	}
	var duplex DeviceNullable[InterfaceDuplex]
	if value, present := values.Duplex.Get(); present {
		parsed, valid := ParseInterfaceDuplex(strings.TrimSpace(value))
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "duplex", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
		} else {
			duplex = NonNullDeviceValue(parsed)
		}
	}
	return normalizedInterfaceValues{
		device: values.Device, name: name, label: label, interfaceType: interfaceType,
		enabled: values.Enabled, mgmtOnly: values.MgmtOnly, mtu: values.MTU,
		speed: values.Speed, duplex: duplex, description: description,
	}, violations
}
