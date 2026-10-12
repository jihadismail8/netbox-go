package dcim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()
	fields := []struct {
		name     string
		set      func(*CreateInterfaceCommand, FieldState)
		full     func(interfaceCommandValues) any
		patch    func(interfaceCommandPatch) any
		concrete any
		fallback any
		required bool
		nullable bool
		choice   bool
	}{
		{"device", func(c *CreateInterfaceCommand, s FieldState) { c.Device = interfacePresenceField(s, shared.ID(7)) }, func(v interfaceCommandValues) any { return v.deviceID }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.deviceID) }, shared.ID(7), shared.ID(0), true, false, false},
		{"name", func(c *CreateInterfaceCommand, s FieldState) { c.Name = interfacePresenceField(s, "  Ethernet1  ") }, func(v interfaceCommandValues) any { return v.name }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.name) }, "  Ethernet1  ", "", true, false, false},
		{"label", func(c *CreateInterfaceCommand, s FieldState) { c.Label = interfacePresenceField(s, "") }, func(v interfaceCommandValues) any { return v.label }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.label) }, "", "", false, false, false},
		{"type", func(c *CreateInterfaceCommand, s FieldState) { c.Type = interfacePresenceField(s, "bridge") }, func(v interfaceCommandValues) any { return v.interfaceType }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.interfaceType) }, "bridge", "", true, false, true},
		{"enabled", func(c *CreateInterfaceCommand, s FieldState) { c.Enabled = interfacePresenceField(s, false) }, func(v interfaceCommandValues) any { return v.enabled }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.enabled) }, false, true, false, false, false},
		{"mgmt_only", func(c *CreateInterfaceCommand, s FieldState) { c.MgmtOnly = interfacePresenceField(s, false) }, func(v interfaceCommandValues) any { return v.mgmtOnly }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.mgmtOnly) }, false, false, false, false, false},
		{"mtu", func(c *CreateInterfaceCommand, s FieldState) { c.MTU = interfacePresenceField(s, int64(-1)) }, func(v interfaceCommandValues) any { return v.mtu }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.mtu) }, domain.NonNullDeviceValue(int64(-1)), domain.NullDeviceValue[int64](), false, true, false},
		{"speed", func(c *CreateInterfaceCommand, s FieldState) { c.Speed = interfacePresenceField(s, int64(0)) }, func(v interfaceCommandValues) any { return v.speed }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.speed) }, domain.NonNullDeviceValue(int64(0)), domain.NullDeviceValue[int64](), false, true, false},
		{"duplex", func(c *CreateInterfaceCommand, s FieldState) { c.Duplex = interfacePresenceField(s, "") }, func(v interfaceCommandValues) any { return v.duplex }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.duplex) }, domain.NonNullDeviceValue(""), domain.NullDeviceValue[string](), false, true, false},
		{"description", func(c *CreateInterfaceCommand, s FieldState) { c.Description = interfacePresenceField(s, "") }, func(v interfaceCommandValues) any { return v.description }, func(p interfaceCommandPatch) any { return interfacePresencePointer(p.description) }, "", "", false, false, false},
	}
	states := []struct {
		name  string
		state FieldState
	}{
		{"omitted", FieldOmitted}, {"null", FieldNull}, {"concrete", FieldPresent},
	}
	for _, field := range fields {
		for _, state := range states {
			for _, operation := range []string{"create", "replace", "update"} {
				t.Run(operation+"/"+field.name+"/"+state.name, func(t *testing.T) {
					input := interfacePresenceBaseCommand()
					field.set(&input, state.state)
					wantReason := ""
					if state.state == FieldOmitted && field.required && operation != "update" {
						wantReason = "required"
					}
					if state.state == FieldNull && !field.nullable && !field.choice {
						wantReason = "null"
					}
					var err error
					var actual any
					if operation == "create" {
						var values interfaceCommandValues
						values, err = input.values()
						actual = field.full(values)
					} else {
						var patch interfaceCommandPatch
						if operation == "replace" {
							patch, err = (ReplaceInterfaceCommand{CreateInterfaceCommand: input}).patch()
						} else {
							patch, err = interfacePresenceUpdate(input).patch()
						}
						actual = field.patch(patch)
					}
					if wantReason == "" {
						require.NoError(t, err)
					} else {
						require.Error(t, err)
						violations := shared.ViolationsOf(err)
						require.Len(t, violations, 1)
						assert.Equal(t, field.name, violations[0].Field)
						assert.Equal(t, wantReason, violations[0].Reason)
					}
					var expected any
					switch {
					case state.state == FieldPresent:
						expected = field.concrete
					case operation == "create":
						expected = field.fallback
					case state.state == FieldNull && (field.nullable || field.choice):
						expected = field.fallback
					}
					assert.Equal(t, expected, actual)
				})
			}
		}
	}
	t.Run("null_only_nonnullable_patch_keeps_its_field_error", func(t *testing.T) {
		patch, err := (UpdateInterfaceCommand{Name: NullField[string](), Enabled: NullField[bool](), Description: NullField[string]()}).patch()
		require.Error(t, err)
		assert.True(t, patch.empty())
		violations := shared.ViolationsOf(err)
		require.Len(t, violations, 3)
		assert.Equal(t, []string{"name", "enabled", "description"}, []string{violations[0].Field, violations[1].Field, violations[2].Field})
	})
	t.Run("partial_values_retain_independent_invalid_intent", func(t *testing.T) {
		input := interfacePresenceBaseCommand()
		input.Name = NullField[string]()
		input.Label = FieldValue("pending label")
		input.Type = FieldValue(" invalid ")
		input.MTU = FieldValue(int64(-1))
		input.Speed = FieldValue(int64(-1))
		values, err := input.values()
		require.Error(t, err)
		assert.Equal(t, "pending label", values.label)
		assert.Equal(t, " invalid ", values.interfaceType)
		assert.Equal(t, domain.NonNullDeviceValue(int64(-1)), values.mtu)
		assert.Equal(t, domain.NonNullDeviceValue(int64(-1)), values.speed)
		patch, err := interfacePresenceUpdate(input).patch()
		require.Error(t, err)
		require.NotNil(t, patch.label)
		assert.Equal(t, "pending label", *patch.label)
		require.NotNil(t, patch.mtu)
		assert.Equal(t, values.mtu, *patch.mtu)
	})
	t.Run("merge_is_ordered_and_fail_closed", func(t *testing.T) {
		commandErr := shared.NewValidationError(shared.FieldViolation{Field: "name", Reason: "null", Description: "This field may not be null."})
		domainErr := shared.NewValidationError(
			shared.FieldViolation{Field: "speed", Reason: "out_of_range", Description: "minimum"},
			shared.FieldViolation{Field: "name", Reason: "blank", Description: "blank"},
			shared.FieldViolation{Field: "device", Reason: "invalid_choice", Description: "invalid"},
		)
		merged := shared.ViolationsOf(mergeInterfaceMutationErrors(commandErr, domainErr))
		require.Len(t, merged, 3)
		assert.Equal(t, []string{"device", "name", "speed"}, []string{merged[0].Field, merged[1].Field, merged[2].Field})
		assert.Equal(t, "null", merged[1].Reason)
		generic := shared.NewError(shared.ErrorReasonValidation, "Rejected input.")
		assert.Same(t, generic, mergeInterfaceMutationErrors(generic, domainErr))
		internal := shared.NewError(shared.ErrorReasonInternal, "Dependency failed.")
		assert.Same(t, internal, mergeInterfaceMutationErrors(generic, domainErr, internal))
		assert.NoError(t, mergeInterfaceMutationErrors(nil, nil))
	})
}

func interfacePresenceBaseCommand() CreateInterfaceCommand {
	return CreateInterfaceCommand{Device: FieldValue(shared.ID(7)), Name: FieldValue("Ethernet1"), Type: FieldValue("bridge")}
}

func interfacePresenceUpdate(input CreateInterfaceCommand) UpdateInterfaceCommand {
	return UpdateInterfaceCommand{
		Device: input.Device, Name: input.Name, Label: input.Label, Type: input.Type,
		Enabled: input.Enabled, MgmtOnly: input.MgmtOnly, MTU: input.MTU, Speed: input.Speed,
		Duplex: input.Duplex, Description: input.Description,
	}
}

func interfacePresenceField[T any](state FieldState, value T) Field[T] {
	if state == FieldOmitted {
		return OmittedField[T]()
	}
	if state == FieldNull {
		return NullField[T]()
	}
	return FieldValue(value)
}

func interfacePresencePointer[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
