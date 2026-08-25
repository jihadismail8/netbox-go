package dcim

import (
	"fmt"
	"strings"

	"netbox-go/internal/domain/shared"
)

const (
	InterfaceTemplateObjectType           = "dcim.interfacetemplate"
	InterfaceTemplateNameMaxLength        = 64
	InterfaceTemplateLabelMaxLength       = 64
	InterfaceTemplateTypeMaxLength        = 50
	InterfaceTemplateDescriptionMaxLength = 200
)

// InterfaceType is an exact value from NetBox v4.4.6's
// InterfaceTypeChoices. It is shared by InterfaceTemplate and, eventually,
// the typed Interface aggregate.
type InterfaceType string

func ParseInterfaceType(value string) (InterfaceType, bool) {
	parsed := InterfaceType(value)
	_, valid := interfaceTypeValues[parsed]
	return parsed, valid
}

func (value InterfaceType) String() string { return string(value) }

var interfaceTypeValues = func() map[InterfaceType]struct{} {
	values := strings.Fields(`
virtual bridge lag
100base-fx 100base-lfx 100base-tx 100base-t1
1000base-bx10-d 1000base-bx10-u 1000base-cwdm 1000base-cx 1000base-dwdm 1000base-ex 1000base-sx 1000base-lsx 1000base-lx 1000base-lx10 1000base-t 1000base-tx 1000base-zx
2.5gbase-t 5gbase-t
10gbase-br-d 10gbase-br-u 10gbase-cx4 10gbase-er 10gbase-lr 10gbase-lrm 10gbase-lx4 10gbase-sr 10gbase-t 10gbase-zr
25gbase-cr 25gbase-er 25gbase-lr 25gbase-sr 25gbase-t
40gbase-cr4 40gbase-er4 40gbase-fr4 40gbase-lr4 40gbase-sr4
50gbase-cr 50gbase-er 50gbase-fr 50gbase-lr 50gbase-sr
100gbase-cr1 100gbase-cr2 100gbase-cr4 100gbase-cr10 100gbase-cwdm4 100gbase-dr 100gbase-fr1 100gbase-er4 100gbase-lr1 100gbase-lr4 100gbase-sr1 100gbase-sr1.2 100gbase-sr2 100gbase-sr4 100gbase-sr10 100gbase-zr
200gbase-cr2 200gbase-cr4 200gbase-sr2 200gbase-sr4 200gbase-dr4 200gbase-fr4 200gbase-lr4 200gbase-er4 200gbase-vr2
400gbase-cr4 400gbase-dr4 400gbase-er8 400gbase-fr4 400gbase-fr8 400gbase-lr4 400gbase-lr8 400gbase-sr4 400gbase-sr4_2 400gbase-sr8 400gbase-sr16 400gbase-vr4 400gbase-zr
800gbase-cr8 800gbase-dr8 800gbase-sr8 800gbase-vr8
100base-x-sfp 1000base-x-gbic 1000base-x-sfp 10gbase-x-sfpp 10gbase-x-xfp 10gbase-x-xenpak 10gbase-x-x2 25gbase-x-sfp28 50gbase-x-sfp56 40gbase-x-qsfpp 50gbase-x-sfp28 100gbase-x-cfp 100gbase-x-cfp2 100gbase-x-cfp4 100gbase-x-cxp 100gbase-x-cpak 100gbase-x-dsfp 100gbase-x-sfpdd 100gbase-x-qsfp28 100gbase-x-qsfpdd 200gbase-x-cfp2 200gbase-x-qsfp56 200gbase-x-qsfpdd 400gbase-x-cfp2 400gbase-x-qsfp112 400gbase-x-qsfpdd 400gbase-x-osfp 400gbase-x-osfp-rhs 400gbase-x-cdfp 400gbase-x-cfp8 800gbase-x-qsfpdd 800gbase-x-osfp
1000base-kx 2.5gbase-kx 5gbase-kr 10gbase-kr 10gbase-kx4 25gbase-kr 40gbase-kr4 50gbase-kr 100gbase-kp4 100gbase-kr2 100gbase-kr4
ieee802.11a ieee802.11g ieee802.11n ieee802.11ac ieee802.11ad ieee802.11ax ieee802.11ay ieee802.11be ieee802.15.1 ieee802.15.4 other-wireless
gsm cdma lte 4g 5g
sonet-oc3 sonet-oc12 sonet-oc48 sonet-oc192 sonet-oc768 sonet-oc1920 sonet-oc3840
1gfc-sfp 2gfc-sfp 4gfc-sfp 8gfc-sfpp 16gfc-sfpp 32gfc-sfp28 32gfc-sfpp 64gfc-qsfpp 64gfc-sfpdd 64gfc-sfpp 128gfc-qsfp28
infiniband-sdr infiniband-ddr infiniband-qdr infiniband-fdr10 infiniband-fdr infiniband-edr infiniband-hdr infiniband-ndr infiniband-xdr
t1 e1 t3 e3 xdsl docsis moca
bpon epon 10g-epon gpon xg-pon xgs-pon ng-pon2 25g-pon 50g-pon
cisco-stackwise cisco-stackwise-plus cisco-flexstack cisco-flexstack-plus cisco-stackwise-80 cisco-stackwise-160 cisco-stackwise-320 cisco-stackwise-480 cisco-stackwise-1t juniper-vcp extreme-summitstack extreme-summitstack-128 extreme-summitstack-256 extreme-summitstack-512
other
`)
	set := make(map[InterfaceType]struct{}, len(values))
	for _, value := range values {
		set[InterfaceType(value)] = struct{}{}
	}
	return set
}()

// DeviceTypeReference is the immutable relationship projection held by an
// InterfaceTemplate. Its display is the DeviceType model, matching NetBox's
// nested object reference.
type DeviceTypeReference struct {
	id    shared.ID
	model string
	slug  shared.Slug
}

func NewDeviceTypeReference(id shared.ID, model, slug string) (DeviceTypeReference, error) {
	model = strings.TrimSpace(model)
	if !id.IsValid() || model == "" {
		return DeviceTypeReference{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid DeviceType reference.",
		)
	}
	parsedSlug, err := shared.ParseSlug(strings.TrimSpace(slug), DeviceTypeSlugMaxLength)
	if err != nil {
		return DeviceTypeReference{}, shared.WrapError(
			shared.ErrorReasonInternal,
			"Cannot construct an invalid DeviceType reference.",
			err,
		)
	}
	return DeviceTypeReference{id: id, model: model, slug: parsedSlug}, nil
}

func (reference DeviceTypeReference) ID() shared.ID     { return reference.id }
func (reference DeviceTypeReference) Model() string     { return reference.model }
func (reference DeviceTypeReference) Slug() shared.Slug { return reference.slug }
func (reference DeviceTypeReference) Display() string   { return reference.model }
func (reference DeviceTypeReference) Valid() bool {
	return reference.id.IsValid() && reference.model != "" && reference.slug.String() != ""
}

type InterfaceTemplateValues struct {
	DeviceType  DeviceTypeReference
	Name        string
	Label       string
	Type        string
	Enabled     bool
	MgmtOnly    bool
	Description string
}

type InterfaceTemplatePatch struct {
	DeviceType  *DeviceTypeReference
	Name        *string
	Label       *string
	Type        *string
	Enabled     *bool
	MgmtOnly    *bool
	Description *string
}

func (patch InterfaceTemplatePatch) Empty() bool {
	return patch.DeviceType == nil && patch.Name == nil && patch.Label == nil &&
		patch.Type == nil && patch.Enabled == nil && patch.MgmtOnly == nil &&
		patch.Description == nil
}

type InterfaceTemplateState struct {
	ID          shared.ID
	DeviceType  DeviceTypeReference
	Name        string
	Label       string
	Type        string
	Enabled     bool
	MgmtOnly    bool
	Description string
	Created     shared.Timestamp
	LastUpdated shared.Timestamp
}

// InterfaceTemplateSnapshot is also the canonical DeviceType-cascade audit
// projection. Keep it independent from persistence and transport DTOs.
type InterfaceTemplateSnapshot struct {
	DeviceTypeID shared.ID
	Name         string
	Label        string
	Type         string
	Enabled      bool
	MgmtOnly     bool
	Description  string
}

func (InterfaceTemplateSnapshot) ObjectType() string { return InterfaceTemplateObjectType }
func (snapshot InterfaceTemplateSnapshot) CloneSnapshot() shared.ObjectSnapshot {
	return snapshot
}

type InterfaceTemplate struct {
	id            shared.ID
	deviceType    DeviceTypeReference
	name          string
	label         string
	interfaceType InterfaceType
	enabled       bool
	mgmtOnly      bool
	description   string
	created       shared.Timestamp
	lastUpdated   shared.Timestamp
}

func NewInterfaceTemplate(
	values InterfaceTemplateValues,
	now shared.Timestamp,
) (*InterfaceTemplate, error) {
	if now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
	}
	normalized, violations := validateInterfaceTemplateValues(values)
	if len(violations) > 0 {
		return nil, shared.NewValidationError(violations...)
	}
	return interfaceTemplateFromNormalized(normalized, now, now), nil
}

func RestoreInterfaceTemplate(state InterfaceTemplateState) (*InterfaceTemplate, error) {
	if !state.ID.IsValid() || state.Created.IsZero() || state.LastUpdated.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot restore invalid InterfaceTemplate identity or timestamps.",
		)
	}
	normalized, violations := validateInterfaceTemplateValues(InterfaceTemplateValues{
		DeviceType: state.DeviceType, Name: state.Name, Label: state.Label, Type: state.Type,
		Enabled: state.Enabled, MgmtOnly: state.MgmtOnly, Description: state.Description,
	})
	if len(violations) > 0 {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted InterfaceTemplate violates domain invariants.",
			shared.NewValidationError(violations...),
		)
	}
	template := interfaceTemplateFromNormalized(normalized, state.Created, state.LastUpdated)
	template.id = state.ID
	return template, nil
}

func interfaceTemplateFromNormalized(
	values normalizedInterfaceTemplateValues,
	created shared.Timestamp,
	lastUpdated shared.Timestamp,
) *InterfaceTemplate {
	return &InterfaceTemplate{
		deviceType: values.deviceType, name: values.name, label: values.label,
		interfaceType: values.interfaceType, enabled: values.enabled, mgmtOnly: values.mgmtOnly,
		description: values.description, created: created, lastUpdated: lastUpdated,
	}
}

func (template *InterfaceTemplate) AssignID(id shared.ID) error {
	if template == nil || !id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot assign an invalid InterfaceTemplate ID.")
	}
	if template.id.IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot replace an assigned InterfaceTemplate ID.")
	}
	template.id = id
	return nil
}

func (template *InterfaceTemplate) Replace(
	values InterfaceTemplateValues,
	now shared.Timestamp,
) error {
	if template == nil || now.IsZero() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot replace InterfaceTemplate with invalid state or time.",
		)
	}
	normalized, violations := validateInterfaceTemplateValues(values)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	if template.deviceType.Valid() && template.deviceType.ID() != normalized.deviceType.ID() {
		return immutableInterfaceTemplateDeviceType()
	}
	created, id := template.created, template.id
	*template = *interfaceTemplateFromNormalized(normalized, created, now)
	template.id = id
	return nil
}

func (template *InterfaceTemplate) ApplyPatch(
	patch InterfaceTemplatePatch,
	now shared.Timestamp,
) error {
	if template == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot patch a nil InterfaceTemplate.")
	}
	if patch.Empty() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	if err := template.ValidatePatch(patch); err != nil {
		return err
	}
	return template.Replace(template.valuesWithPatch(patch), now)
}

// ValidatePatch previews a non-mutating replacement. Application services use
// it to merge command, relationship, and aggregate violations before any
// repository mutation. An empty preview is valid; the public ApplyPatch entry
// point owns the update-mask requirement.
func (template *InterfaceTemplate) ValidatePatch(patch InterfaceTemplatePatch) error {
	if template == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot validate a nil InterfaceTemplate.")
	}

	var violations []shared.FieldViolation
	if patch.DeviceType != nil && template.deviceType.Valid() && patch.DeviceType.Valid() &&
		template.deviceType.ID() != patch.DeviceType.ID() {
		violations = append(violations, shared.ViolationsOf(immutableInterfaceTemplateDeviceType())...)
	}
	_, valueViolations := validateInterfaceTemplateValues(template.valuesWithPatch(patch))
	violations = append(violations, valueViolations...)
	if len(violations) > 0 {
		return shared.NewValidationError(violations...)
	}
	return nil
}

func (template InterfaceTemplate) valuesWithPatch(patch InterfaceTemplatePatch) InterfaceTemplateValues {
	values := template.Values()
	if patch.DeviceType != nil {
		values.DeviceType = *patch.DeviceType
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
	setString(&values.Description, patch.Description)
	return values
}

func immutableInterfaceTemplateDeviceType() error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "device_type", Reason: "immutable",
		Description: "An InterfaceTemplate cannot be moved to another DeviceType.",
	})
}

func (template InterfaceTemplate) ID() shared.ID                   { return template.id }
func (template InterfaceTemplate) DeviceType() DeviceTypeReference { return template.deviceType }
func (template InterfaceTemplate) Name() string                    { return template.name }
func (template InterfaceTemplate) Label() string                   { return template.label }
func (template InterfaceTemplate) Type() InterfaceType             { return template.interfaceType }
func (template InterfaceTemplate) Enabled() bool                   { return template.enabled }
func (template InterfaceTemplate) MgmtOnly() bool                  { return template.mgmtOnly }
func (template InterfaceTemplate) Description() string             { return template.description }
func (template InterfaceTemplate) Created() shared.Timestamp       { return template.created }
func (template InterfaceTemplate) LastUpdated() shared.Timestamp   { return template.lastUpdated }
func (template InterfaceTemplate) Display() string {
	if template.label != "" {
		return fmt.Sprintf("%s (%s)", template.name, template.label)
	}
	return template.name
}

func (template InterfaceTemplate) Values() InterfaceTemplateValues {
	return InterfaceTemplateValues{
		DeviceType: template.deviceType, Name: template.name, Label: template.label,
		Type: template.interfaceType.String(), Enabled: template.enabled,
		MgmtOnly: template.mgmtOnly, Description: template.description,
	}
}

func (template InterfaceTemplate) State() InterfaceTemplateState {
	values := template.Values()
	return InterfaceTemplateState{
		ID: template.id, DeviceType: values.DeviceType, Name: values.Name, Label: values.Label,
		Type: values.Type, Enabled: values.Enabled, MgmtOnly: values.MgmtOnly,
		Description: values.Description, Created: template.created, LastUpdated: template.lastUpdated,
	}
}

func (template InterfaceTemplate) Snapshot() InterfaceTemplateSnapshot {
	return InterfaceTemplateSnapshot{
		DeviceTypeID: template.deviceType.ID(), Name: template.name, Label: template.label,
		Type: template.interfaceType.String(), Enabled: template.enabled,
		MgmtOnly: template.mgmtOnly, Description: template.description,
	}
}

type normalizedInterfaceTemplateValues struct {
	deviceType    DeviceTypeReference
	name          string
	label         string
	interfaceType InterfaceType
	enabled       bool
	mgmtOnly      bool
	description   string
}

func validateInterfaceTemplateValues(
	values InterfaceTemplateValues,
) (normalizedInterfaceTemplateValues, []shared.FieldViolation) {
	values.Name = strings.TrimSpace(values.Name)
	values.Label = strings.TrimSpace(values.Label)
	values.Description = strings.TrimSpace(values.Description)
	var violations []shared.FieldViolation
	if !values.DeviceType.Valid() {
		violations = append(violations, shared.FieldViolation{
			Field: "device_type", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	validateRequiredLength(&violations, "name", values.Name, InterfaceTemplateNameMaxLength)
	validateOptionalLength(&violations, "label", values.Label, InterfaceTemplateLabelMaxLength)
	interfaceType, validType := ParseInterfaceType(values.Type)
	if !validType {
		violations = append(violations, shared.FieldViolation{
			Field: "type", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	validateOptionalLength(
		&violations, "description", values.Description, InterfaceTemplateDescriptionMaxLength,
	)
	return normalizedInterfaceTemplateValues{
		deviceType: values.DeviceType, name: values.Name, label: values.Label,
		interfaceType: interfaceType, enabled: values.Enabled, mgmtOnly: values.MgmtOnly,
		description: values.Description,
	}, violations
}
