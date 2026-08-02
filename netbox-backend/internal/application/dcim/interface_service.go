package dcim

import (
	"context"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type InterfaceService struct {
	repository  InterfaceRepository
	devices     InterfaceDeviceReader
	ipAddresses InterfaceIPAddressCascade
	unitOfWork  transaction.UnitOfWork
	recorder    changelog.Recorder
	authorizer  authz.ResourceAuthorizer
	clock       shared.Clock
}

func NewInterfaceService(
	repository InterfaceRepository,
	devices InterfaceDeviceReader,
	ipAddresses InterfaceIPAddressCascade,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*InterfaceService, error) {
	missing := make([]string, 0, 7)
	if nilInterface(repository) {
		missing = append(missing, "repository")
	}
	if nilInterface(devices) {
		missing = append(missing, "device reader")
	}
	if nilInterface(ipAddresses) {
		missing = append(missing, "IP address cascade")
	}
	if nilInterface(unitOfWork) {
		missing = append(missing, "unit of work")
	}
	if nilInterface(recorder) {
		missing = append(missing, "change recorder")
	}
	if nilInterface(authorizer) {
		missing = append(missing, "authorizer")
	}
	if nilInterface(clock) {
		missing = append(missing, "clock")
	}
	if len(missing) > 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Interface service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &InterfaceService{
		repository: repository, devices: devices, ipAddresses: ipAddresses,
		unitOfWork: unitOfWork, recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *InterfaceService) ListInterfaces(
	ctx context.Context,
	principal identity.Principal,
	query ListInterfacesQuery,
) (InterfacePage, error) {
	if err := service.authorize(ctx, principal, authz.View, 0); err != nil {
		return InterfacePage{}, err
	}
	criteria, err := validateListInterfacesQuery(query)
	if err != nil {
		return InterfacePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx, principal, authz.View, authz.ResourceInterface,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return InterfacePage{}, normalizeOperationError("Could not list Interfaces.", err)
	}
	visible := make([]*dcimdomain.Interface, 0, len(page.Results))
	for _, networkInterface := range page.Results {
		if networkInterface == nil {
			return InterfacePage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"Interface repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(
			ctx, principal, authz.View, networkInterface.ID(),
		)
		if authorizeErr == nil {
			visible = append(visible, networkInterface)
			continue
		}
		if hasCompleteScope {
			return InterfacePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"Interface visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return InterfacePage{}, authorizeErr
	}
	if hasCompleteScope {
		return InterfacePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return InterfacePage{Count: count, Results: visible[start:end]}, nil
}

func (service *InterfaceService) GetInterface(
	ctx context.Context,
	principal identity.Principal,
	query GetInterfaceQuery,
) (*dcimdomain.Interface, error) {
	if err := service.authorize(ctx, principal, authz.View, 0); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	networkInterface, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get Interface.", err)
	}
	if err := service.authorize(
		ctx, principal, authz.View, networkInterface.ID(),
	); err != nil {
		return nil, err
	}
	return networkInterface, nil
}

func (service *InterfaceService) CreateInterface(
	ctx context.Context,
	principal identity.Principal,
	command CreateInterfaceCommand,
) (*dcimdomain.Interface, error) {
	if err := service.authorize(ctx, principal, authz.Add, 0); err != nil {
		return nil, err
	}
	var result *dcimdomain.Interface
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		commandValues, err := command.values()
		if err != nil {
			return err
		}
		values, err := service.resolveValues(transactionContext, commandValues)
		if err != nil {
			return err
		}
		now := service.clock.Now()
		candidate, err := dcimdomain.NewInterface(values, now)
		if err != nil {
			return err
		}
		if err := service.repository.Create(transactionContext, candidate); err != nil {
			return err
		}
		if err := service.recordInterfaceChange(
			transactionContext, principal, candidate, changelog.ActionCreate,
			nil, candidate.Snapshot(), now,
		); err != nil {
			return err
		}
		result = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create Interface.", err)
	}
	return result, nil
}

func (service *InterfaceService) ReplaceInterface(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceInterfaceCommand,
) (*dcimdomain.Interface, error) {
	return service.mutateInterface(
		ctx, principal, command.ID,
		func(
			transactionContext context.Context,
			loaded *dcimdomain.Interface,
			now shared.Timestamp,
		) error {
			commandValues, err := command.values()
			if err != nil {
				return err
			}
			values, err := service.resolveValues(transactionContext, commandValues)
			if err != nil {
				return err
			}
			return loaded.Replace(values, now)
		},
		"replace",
	)
}

func (service *InterfaceService) UpdateInterface(
	ctx context.Context,
	principal identity.Principal,
	command UpdateInterfaceCommand,
) (*dcimdomain.Interface, error) {
	return service.mutateInterface(
		ctx, principal, command.ID,
		func(
			transactionContext context.Context,
			loaded *dcimdomain.Interface,
			now shared.Timestamp,
		) error {
			commandPatch, err := command.patch()
			if err != nil {
				return err
			}
			patch, err := service.resolvePatch(transactionContext, commandPatch)
			if err != nil {
				return err
			}
			return loaded.ApplyPatch(patch, now)
		},
		"update",
	)
}

func (service *InterfaceService) mutateInterface(
	ctx context.Context,
	principal identity.Principal,
	id shared.ID,
	mutation func(context.Context, *dcimdomain.Interface, shared.Timestamp) error,
	operation string,
) (*dcimdomain.Interface, error) {
	if err := service.authorize(ctx, principal, authz.Change, 0); err != nil {
		return nil, err
	}
	if err := validatePersistedID(id); err != nil {
		return nil, err
	}
	var result *dcimdomain.Interface
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, id)
		if err != nil {
			return err
		}
		if err := service.authorize(
			transactionContext, principal, authz.Change, loaded.ID(),
		); err != nil {
			return err
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if err := mutation(transactionContext, loaded, now); err != nil {
			return err
		}
		if err := service.repository.Update(transactionContext, loaded); err != nil {
			return err
		}
		if err := service.recordInterfaceChange(
			transactionContext, principal, loaded, changelog.ActionUpdate,
			before, loaded.Snapshot(), now,
		); err != nil {
			return err
		}
		result = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not "+operation+" Interface.", err)
	}
	return result, nil
}

func (service *InterfaceService) DeleteInterface(
	ctx context.Context,
	principal identity.Principal,
	command DeleteInterfaceCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, 0); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
		if err != nil {
			return err
		}
		if err := service.authorize(
			transactionContext, principal, authz.Delete, loaded.ID(),
		); err != nil {
			return err
		}
		now := service.clock.Now()
		children, err := service.ipAddresses.DeleteAssignedToInterface(
			transactionContext, loaded.ID(), now,
		)
		if err != nil {
			return err
		}
		before := loaded.Snapshot()
		if err := service.repository.Delete(transactionContext, loaded); err != nil {
			return err
		}
		for _, child := range children {
			change, err := changelog.NewChange(
				principal.ID, child.Before.ObjectType(), child.ID,
				child.Representation, changelog.ActionDelete, child.Before, nil, now,
			)
			if err != nil {
				return err
			}
			if err := service.recorder.Record(transactionContext, change); err != nil {
				return err
			}
		}
		return service.recordInterfaceChange(
			transactionContext, principal, loaded, changelog.ActionDelete,
			before, nil, now,
		)
	})
	return normalizeOperationError("Could not delete Interface.", err)
}

// DeleteForDevice deletes all Interfaces for a parent Device under the
// caller's ambient transaction. It deliberately does not authorize or record:
// Device deletion owns those policies and consumes the ordered typed changes.
func (service *InterfaceService) DeleteForDevice(
	ctx context.Context,
	deviceID shared.ID,
	now shared.Timestamp,
) ([]InterfaceCascadeChange, error) {
	if err := validatePersistedID(deviceID); err != nil {
		return nil, err
	}
	if now.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Cannot cascade Interfaces with a zero timestamp.",
		)
	}
	interfaces, err := service.repository.ListForDeviceForUpdate(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	changes := make([]InterfaceCascadeChange, 0, len(interfaces))
	for _, networkInterface := range interfaces {
		if networkInterface == nil {
			return nil, shared.NewError(
				shared.ErrorReasonInternal,
				"Interface repository returned an invalid cascade item.",
			)
		}
		children, err := service.ipAddresses.DeleteAssignedToInterface(
			ctx, networkInterface.ID(), now,
		)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			changes = append(changes, InterfaceCascadeChange{
				ObjectType: child.Before.ObjectType(), ID: child.ID,
				Representation: child.Representation, Before: child.Before,
			})
		}
		before := networkInterface.Snapshot()
		if err := service.repository.Delete(ctx, networkInterface); err != nil {
			return nil, err
		}
		changes = append(changes, InterfaceCascadeChange{
			ObjectType: dcimdomain.InterfaceObjectType, ID: networkInterface.ID(),
			Representation: networkInterface.Display(), Before: before,
		})
	}
	return changes, nil
}

func (service *InterfaceService) resolveValues(
	ctx context.Context,
	values interfaceCommandValues,
) (dcimdomain.InterfaceValues, error) {
	device, err := service.devices.GetDeviceReference(ctx, values.deviceID)
	if err != nil {
		return dcimdomain.InterfaceValues{}, err
	}
	return dcimdomain.InterfaceValues{
		Device: device, Name: values.name, Label: values.label,
		Type: values.interfaceType, Enabled: values.enabled, MgmtOnly: values.mgmtOnly,
		MTU: values.mtu, Speed: values.speed, Duplex: values.duplex,
		Description: values.description,
	}, nil
}

func (service *InterfaceService) resolvePatch(
	ctx context.Context,
	patch interfaceCommandPatch,
) (dcimdomain.InterfacePatch, error) {
	resolved := dcimdomain.InterfacePatch{
		Name: patch.name, Label: patch.label, Type: patch.interfaceType,
		Enabled: patch.enabled, MgmtOnly: patch.mgmtOnly, MTU: patch.mtu,
		Speed: patch.speed, Duplex: patch.duplex, Description: patch.description,
	}
	if patch.deviceID != nil {
		device, err := service.devices.GetDeviceReference(ctx, *patch.deviceID)
		if err != nil {
			return dcimdomain.InterfacePatch{}, err
		}
		resolved.Device = &device
	}
	return resolved, nil
}

func (service *InterfaceService) recordInterfaceChange(
	ctx context.Context,
	principal identity.Principal,
	networkInterface *dcimdomain.Interface,
	action changelog.Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.InterfaceObjectType, networkInterface.ID(),
		networkInterface.Display(), action, before, after, now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *InterfaceService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	id shared.ID,
) error {
	return service.authorizer.AuthorizeResource(
		ctx, principal, action, authz.ResourceInterface, authz.NewObject(id.Int64()),
	)
}

var _ DeviceInterfaceCascade = (*InterfaceService)(nil)
