package dcim_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceScalarValidationLeavesStateUnchanged(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*appdcim.CreateInterfaceCommand)
		fields  []string
		reasons []string
		putOnly bool
	}{
		{"missing device", func(c *appdcim.CreateInterfaceCommand) { c.Device = appdcim.OmittedField[shared.ID]() }, []string{"device"}, []string{"required"}, true},
		{"missing name", func(c *appdcim.CreateInterfaceCommand) { c.Name = appdcim.OmittedField[string]() }, []string{"name"}, []string{"required"}, true},
		{"missing type", func(c *appdcim.CreateInterfaceCommand) { c.Type = appdcim.OmittedField[string]() }, []string{"type"}, []string{"required"}, true},
		{"null device", func(c *appdcim.CreateInterfaceCommand) { c.Device = appdcim.NullField[shared.ID]() }, []string{"device"}, []string{"null"}, false},
		{"null name", func(c *appdcim.CreateInterfaceCommand) { c.Name = appdcim.NullField[string]() }, []string{"name"}, []string{"null"}, false},
		{"null label", func(c *appdcim.CreateInterfaceCommand) { c.Label = appdcim.NullField[string]() }, []string{"label"}, []string{"null"}, false},
		{"null type is blank", func(c *appdcim.CreateInterfaceCommand) { c.Type = appdcim.NullField[string]() }, []string{"type"}, []string{"blank"}, false},
		{"null enabled", func(c *appdcim.CreateInterfaceCommand) { c.Enabled = appdcim.NullField[bool]() }, []string{"enabled"}, []string{"null"}, false},
		{"null mgmt_only", func(c *appdcim.CreateInterfaceCommand) { c.MgmtOnly = appdcim.NullField[bool]() }, []string{"mgmt_only"}, []string{"null"}, false},
		{"null description", func(c *appdcim.CreateInterfaceCommand) { c.Description = appdcim.NullField[string]() }, []string{"description"}, []string{"null"}, false},
		{"zero owner", func(c *appdcim.CreateInterfaceCommand) { c.Device = appdcim.FieldValue(shared.ID(0)) }, []string{"device"}, []string{"invalid_choice"}, false},
		{"negative owner", func(c *appdcim.CreateInterfaceCommand) { c.Device = appdcim.FieldValue(shared.ID(-1)) }, []string{"device"}, []string{"invalid_choice"}, false},
		{"unknown owner", func(c *appdcim.CreateInterfaceCommand) { c.Device = appdcim.FieldValue(shared.ID(999999)) }, []string{"device"}, []string{"invalid_choice"}, false},
		{"blank name", func(c *appdcim.CreateInterfaceCommand) { c.Name = appdcim.FieldValue(" \t ") }, []string{"name"}, []string{"required"}, false},
		{"name character limit", func(c *appdcim.CreateInterfaceCommand) { c.Name = appdcim.FieldValue(strings.Repeat("界", 65)) }, []string{"name"}, []string{"max_length"}, false},
		{"label character limit", func(c *appdcim.CreateInterfaceCommand) { c.Label = appdcim.FieldValue(strings.Repeat("界", 65)) }, []string{"label"}, []string{"max_length"}, false},
		{"description character limit", func(c *appdcim.CreateInterfaceCommand) {
			c.Description = appdcim.FieldValue(strings.Repeat("界", 201))
		}, []string{"description"}, []string{"max_length"}, false},
		{"blank type", func(c *appdcim.CreateInterfaceCommand) { c.Type = appdcim.FieldValue("") }, []string{"type"}, []string{"blank"}, false},
		{"invalid type", func(c *appdcim.CreateInterfaceCommand) { c.Type = appdcim.FieldValue("not-a-type") }, []string{"type"}, []string{"invalid_choice"}, false},
		{"untrimmed type", func(c *appdcim.CreateInterfaceCommand) { c.Type = appdcim.FieldValue(" 1000base-t ") }, []string{"type"}, []string{"invalid_choice"}, false},
		{"negative MTU", func(c *appdcim.CreateInterfaceCommand) { c.MTU = appdcim.FieldValue(int64(-1)) }, []string{"mtu"}, []string{"out_of_range"}, false},
		{"zero MTU", func(c *appdcim.CreateInterfaceCommand) { c.MTU = appdcim.FieldValue(int64(0)) }, []string{"mtu"}, []string{"out_of_range"}, false},
		{"high MTU", func(c *appdcim.CreateInterfaceCommand) { c.MTU = appdcim.FieldValue(int64(65537)) }, []string{"mtu"}, []string{"out_of_range"}, false},
		{"negative speed", func(c *appdcim.CreateInterfaceCommand) { c.Speed = appdcim.FieldValue(int64(-1)) }, []string{"speed"}, []string{"out_of_range"}, false},
		{"high speed", func(c *appdcim.CreateInterfaceCommand) { c.Speed = appdcim.FieldValue(int64(2147483648)) }, []string{"speed"}, []string{"out_of_range"}, false},
		{"invalid duplex", func(c *appdcim.CreateInterfaceCommand) { c.Duplex = appdcim.FieldValue("not-a-duplex") }, []string{"duplex"}, []string{"invalid_choice"}, false},
		{"untrimmed duplex", func(c *appdcim.CreateInterfaceCommand) { c.Duplex = appdcim.FieldValue(" full ") }, []string{"duplex"}, []string{"invalid_choice"}, false},
		{"all layers aggregate in field order", func(c *appdcim.CreateInterfaceCommand) {
			c.Device, c.Name = appdcim.FieldValue(shared.ID(999999)), appdcim.NullField[string]()
			c.Label, c.Type = appdcim.FieldValue(strings.Repeat("界", 65)), appdcim.NullField[string]()
			c.Enabled, c.MgmtOnly = appdcim.NullField[bool](), appdcim.NullField[bool]()
			c.MTU, c.Speed = appdcim.FieldValue(int64(-1)), appdcim.FieldValue(int64(2147483648))
			c.Duplex, c.Description = appdcim.FieldValue(" full "), appdcim.NullField[string]()
		}, []string{"device", "name", "label", "type", "enabled", "mgmt_only", "mtu", "speed", "duplex", "description"}, []string{"invalid_choice", "null", "max_length", "blank", "null", "null", "out_of_range", "out_of_range", "invalid_choice", "null"}, false},
	}
	for _, operation := range []string{"create", "replace", "update"} {
		for _, test := range cases {
			if operation == "update" && test.putOnly {
				continue
			}
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				service, repository, recorder, _, cascade := newInterfaceApplicationService(t, authz.AllowAll{})
				seed := interfacePresenceValidCommand()
				seed.Label, seed.Description = appdcim.FieldValue("old label"), appdcim.FieldValue("old description")
				seed.Enabled, seed.MgmtOnly = appdcim.FieldValue(false), appdcim.FieldValue(true)
				seed.MTU, seed.Speed, seed.Duplex = appdcim.FieldValue(int64(1500)), appdcim.FieldValue(int64(0)), appdcim.FieldValue("")
				created, err := service.CreateInterface(t.Context(), testPrincipal(), seed)
				require.NoError(t, err)
				before, beforeChanges := repository.interfaces[created.ID()], append([]changelog.Change(nil), recorder.changes...)
				command := interfacePresenceValidCommand()
				test.mutate(&command)
				result, err := interfacePresenceMutate(t, service, operation, created.ID(), command)
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, shared.ErrorReasonValidation, shared.ReasonOf(err))
				violations := shared.ViolationsOf(err)
				require.Len(t, violations, len(test.fields))
				for index, field := range test.fields {
					assert.Equal(t, field, violations[index].Field)
					assert.Equal(t, test.reasons[index], violations[index].Reason)
				}
				assert.Equal(t, before, repository.interfaces[created.ID()])
				assert.Equal(t, beforeChanges, recorder.changes)
				assert.Len(t, repository.interfaces, 1)
				assert.Equal(t, 1, repository.createCalls)
				assert.Zero(t, repository.updateCalls)
				assert.Zero(t, repository.deleteCalls)
				assert.Empty(t, cascade.interfaceIDs)
			})
		}
	}
	for _, operation := range []string{"create", "replace", "update"} {
		t.Run(operation+"/selected-owner-not-found-aggregates", func(t *testing.T) {
			service, repository, recorder, devices, cascade := newInterfaceApplicationService(t, authz.AllowAll{})
			created, err := service.CreateInterface(t.Context(), testPrincipal(), interfacePresenceValidCommand())
			require.NoError(t, err)
			service, err = appdcim.NewInterfaceService(
				repository, interfacePresenceDeviceReader{delegate: devices, err: shared.NotFound("Device", 999999)},
				cascade, interfaceApplicationUnitOfWork{}, recorder, authz.AllowAll{}, fixedClock{now: updatedAt},
			)
			require.NoError(t, err)
			before, beforeChanges := repository.interfaces[created.ID()], append([]changelog.Change(nil), recorder.changes...)
			command := interfacePresenceValidCommand()
			command.Device, command.MTU = appdcim.FieldValue(shared.ID(999999)), appdcim.FieldValue(int64(-1))
			result, err := interfacePresenceMutate(t, service, operation, created.ID(), command)
			require.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, shared.ErrorReasonValidation, shared.ReasonOf(err))
			assert.Equal(t, []shared.FieldViolation{
				{Field: "device", Reason: "invalid_choice", Description: "Select a valid choice."},
				{Field: "mtu", Reason: "out_of_range", Description: "Ensure this value is greater than or equal to 1."},
			}, shared.ViolationsOf(err))
			assert.Equal(t, before, repository.interfaces[created.ID()])
			assert.Equal(t, beforeChanges, recorder.changes)
			assert.Equal(t, 1, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
		})
	}
	for _, operation := range []string{"create", "replace", "update"} {
		t.Run(operation+"/reader-infrastructure-failure-is-not-validation", func(t *testing.T) {
			service, repository, recorder, devices, cascade := newInterfaceApplicationService(t, authz.AllowAll{})
			created, err := service.CreateInterface(t.Context(), testPrincipal(), interfacePresenceValidCommand())
			require.NoError(t, err)
			sentinel := errors.New("selected Device reader unavailable")
			service, err = appdcim.NewInterfaceService(repository,
				interfacePresenceDeviceReader{delegate: devices, err: shared.WrapError(shared.ErrorReasonInternal, "Could not read Device.", sentinel)},
				cascade, interfaceApplicationUnitOfWork{}, recorder, authz.AllowAll{}, fixedClock{now: updatedAt})
			require.NoError(t, err)
			before := repository.interfaces[created.ID()]
			command := interfacePresenceValidCommand()
			command.Device, command.MTU = appdcim.FieldValue(shared.ID(999999)), appdcim.FieldValue(int64(-1))
			result, err := interfacePresenceMutate(t, service, operation, created.ID(), command)
			require.ErrorIs(t, err, sentinel)
			assert.Nil(t, result)
			assert.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(err))
			assert.Empty(t, shared.ViolationsOf(err))
			assert.Equal(t, before, repository.interfaces[created.ID()])
			assert.Len(t, recorder.changes, 1)
			assert.Equal(t, 1, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
		})
		for _, denyLoaded := range []bool{false, true} {
			if denyLoaded && operation == "create" {
				continue
			}
			t.Run(operation+"/authorization-before-validation/loaded="+map[bool]string{false: "false", true: "true"}[denyLoaded], func(t *testing.T) {
				service, repository, recorder, devices, cascade := newInterfaceApplicationService(t, authz.AllowAll{})
				created, err := service.CreateInterface(t.Context(), testPrincipal(), interfacePresenceValidCommand())
				require.NoError(t, err)
				var authorizer authz.ResourceAuthorizer = authz.PermissionAuthorizer{}
				if denyLoaded {
					authorizer = interfaceVisibilityAuthorizer{hiddenID: created.ID().Int64()}
				}
				service, err = appdcim.NewInterfaceService(repository,
					interfacePresenceDeviceReader{delegate: devices, err: errors.New("reader must not be reached")},
					cascade, interfaceApplicationUnitOfWork{}, recorder, authorizer, fixedClock{now: updatedAt})
				require.NoError(t, err)
				command := interfacePresenceValidCommand()
				command.Device, command.Name = appdcim.FieldValue(shared.ID(999999)), appdcim.NullField[string]()
				before := repository.interfaces[created.ID()]
				_, err = interfacePresenceMutate(t, service, operation, created.ID(), command)
				require.True(t, shared.HasReason(err, shared.ErrorReasonForbidden), err)
				assert.Empty(t, shared.ViolationsOf(err))
				assert.Equal(t, before, repository.interfaces[created.ID()])
				assert.Len(t, recorder.changes, 1)
				assert.Equal(t, 1, repository.createCalls)
				assert.Zero(t, repository.updateCalls)
			})
		}
	}
	for _, operation := range []string{"replace", "update"} {
		t.Run(operation+"/target-not-found-stays-not-found", func(t *testing.T) {
			service, repository, recorder, _, _ := newInterfaceApplicationService(t, authz.AllowAll{})
			command := interfacePresenceValidCommand()
			command.Device, command.MTU = appdcim.FieldValue(shared.ID(999999)), appdcim.FieldValue(int64(-1))
			_, err := interfacePresenceMutate(t, service, operation, 999999, command)
			require.True(t, shared.HasReason(err, shared.ErrorReasonNotFound), err)
			assert.Empty(t, shared.ViolationsOf(err))
			assert.Empty(t, repository.interfaces)
			assert.Empty(t, recorder.changes)
			assert.Zero(t, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
		})
	}
	t.Run("unauthenticated mutation barriers", func(t *testing.T) {
		service, repository, recorder, _, _ := newInterfaceApplicationService(t, authz.PermissionAuthorizer{})
		_, err := service.CreateInterface(t.Context(), identity.Principal{}, interfacePresenceValidCommand())
		require.True(t, shared.HasReason(err, shared.ErrorReasonUnauthenticated), err)
		_, err = service.ReplaceInterface(t.Context(), identity.Principal{}, appdcim.ReplaceInterfaceCommand{ID: 999999})
		require.True(t, shared.HasReason(err, shared.ErrorReasonUnauthenticated), err)
		_, err = service.UpdateInterface(t.Context(), identity.Principal{}, appdcim.UpdateInterfaceCommand{ID: 999999})
		require.True(t, shared.HasReason(err, shared.ErrorReasonUnauthenticated), err)
		assert.Empty(t, repository.interfaces)
		assert.Empty(t, recorder.changes)
		assert.Zero(t, repository.createCalls)
		assert.Zero(t, repository.updateCalls)
	})
}

func interfacePresenceValidCommand() appdcim.CreateInterfaceCommand {
	return appdcim.CreateInterfaceCommand{
		Device: appdcim.FieldValue(shared.ID(7)), Name: appdcim.FieldValue("Ethernet1"),
		Type: appdcim.FieldValue("1000base-t"),
	}
}

func interfacePresenceMutate(t *testing.T, service *appdcim.InterfaceService, operation string, id shared.ID, command appdcim.CreateInterfaceCommand) (*dcimdomain.Interface, error) {
	t.Helper()
	switch operation {
	case "create":
		return service.CreateInterface(t.Context(), testPrincipal(), command)
	case "replace":
		return service.ReplaceInterface(t.Context(), testPrincipal(), appdcim.ReplaceInterfaceCommand{ID: id, CreateInterfaceCommand: command})
	case "update":
		return service.UpdateInterface(t.Context(), testPrincipal(), appdcim.UpdateInterfaceCommand{
			ID: id, Device: command.Device, Name: command.Name, Label: command.Label, Type: command.Type,
			Enabled: command.Enabled, MgmtOnly: command.MgmtOnly, MTU: command.MTU,
			Speed: command.Speed, Duplex: command.Duplex, Description: command.Description,
		})
	default:
		t.Fatalf("unsupported test mutation %q", operation)
		return nil, nil
	}
}

type interfacePresenceDeviceReader struct {
	delegate appdcim.InterfaceDeviceReader
	err      error
}

func (reader interfacePresenceDeviceReader) GetDeviceReference(ctx context.Context, id shared.ID) (dcimdomain.DeviceReference, error) {
	if id == 999999 {
		return dcimdomain.DeviceReference{}, reader.err
	}
	return reader.delegate.GetDeviceReference(ctx, id)
}

func TestReplaceInterfacePreservesOmittedOptionalState(t *testing.T) {
	service, repository, recorder, _, _ := newInterfaceApplicationService(t, authz.AllowAll{})
	networkInterface, err := service.CreateInterface(t.Context(), testPrincipal(), appdcim.CreateInterfaceCommand{
		Device: appdcim.FieldValue(shared.ID(7)), Name: appdcim.FieldValue("Ethernet1"),
		Label: appdcim.FieldValue("uplink"), Type: appdcim.FieldValue("1000base-t"),
		Enabled: appdcim.FieldValue(false), MgmtOnly: appdcim.FieldValue(true),
		MTU: appdcim.FieldValue(int64(1500)), Speed: appdcim.FieldValue(int64(1_000_000)),
		Duplex: appdcim.FieldValue("full"), Description: appdcim.FieldValue("old description"),
	})
	require.NoError(t, err)
	before := networkInterface.State()
	beforeSnapshot := networkInterface.Snapshot()

	networkInterface, err = service.ReplaceInterface(t.Context(), testPrincipal(), appdcim.ReplaceInterfaceCommand{
		ID: networkInterface.ID(), CreateInterfaceCommand: appdcim.CreateInterfaceCommand{
			Device: appdcim.FieldValue(shared.ID(7)), Name: appdcim.FieldValue("Ethernet2"),
			Type: appdcim.FieldValue("10gbase-sr"),
		},
	})
	require.NoError(t, err)
	expected := before
	expected.Name = "Ethernet2"
	expected.Type = "10gbase-sr"
	assert.Equal(t, expected, networkInterface.State())
	assert.Equal(t, expected, repository.interfaces[networkInterface.ID()])
	assert.Equal(t, 1, repository.updateCalls)
	require.Len(t, recorder.changes, 2)
	assert.Equal(t, changelog.ActionUpdate, recorder.changes[1].Action)
	assert.Equal(t, beforeSnapshot, recorder.changes[1].Before)
	assert.Equal(t, networkInterface.Snapshot(), recorder.changes[1].After)
}
