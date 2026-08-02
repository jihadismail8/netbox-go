package ipam

import (
	"context"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type IPAddressService struct {
	repository IPAddressRepository
	vrfs       IPAddressVRFReader
	interfaces IPAddressInterfaceReader
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewIPAddressService(
	repository IPAddressRepository,
	vrfs IPAddressVRFReader,
	interfaces IPAddressInterfaceReader,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*IPAddressService, error) {
	missing := make([]string, 0, 7)
	if nilInterface(repository) {
		missing = append(missing, "repository")
	}
	if nilInterface(vrfs) {
		missing = append(missing, "VRF reader")
	}
	if nilInterface(interfaces) {
		missing = append(missing, "Interface reader")
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
			"IPAddress service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &IPAddressService{
		repository: repository,
		vrfs:       vrfs,
		interfaces: interfaces,
		unitOfWork: unitOfWork,
		recorder:   recorder,
		authorizer: authorizer,
		clock:      clock,
	}, nil
}

func (service *IPAddressService) ListIPAddresses(
	ctx context.Context,
	principal identity.Principal,
	query ListIPAddressesQuery,
) (IPAddressPage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return IPAddressPage{}, err
	}
	criteria, err := validateListIPAddressesQuery(query)
	if err != nil {
		return IPAddressPage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx, principal, authz.View, authz.ResourceIPAddress,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = ipAddressSharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return IPAddressPage{}, normalizeOperationError("Could not list IP addresses.", err)
	}
	visible := make([]*domainipam.IPAddress, 0, len(page.Results))
	for _, address := range page.Results {
		if address == nil {
			return IPAddressPage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"IPAddress repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, address)
		if authorizeErr == nil {
			visible = append(visible, address)
			continue
		}
		if hasCompleteScope {
			return IPAddressPage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"IPAddress visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return IPAddressPage{}, authorizeErr
	}
	if hasCompleteScope {
		return IPAddressPage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return IPAddressPage{Count: count, Results: visible[start:end]}, nil
}

func (service *IPAddressService) GetIPAddress(
	ctx context.Context,
	principal identity.Principal,
	query GetIPAddressQuery,
) (*domainipam.IPAddress, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	address, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get IP address.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, address); err != nil {
		return nil, err
	}
	return address, nil
}

func (service *IPAddressService) CreateIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command CreateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var result *domainipam.IPAddress
	err := service.unitOfWork.WithinTransaction(
		ctx,
		func(transactionContext context.Context) error {
			commandValues, err := command.values()
			if err != nil {
				return err
			}
			values, err := service.resolveValues(
				transactionContext,
				commandValues,
				domainipam.NullInterfaceAssignment(),
			)
			if err != nil {
				return err
			}
			now := service.clock.Now()
			candidate, err := domainipam.NewIPAddress(values, now)
			if err != nil {
				return err
			}
			if err := service.authorize(
				transactionContext, principal, authz.Add, candidate,
			); err != nil {
				return err
			}
			if err := service.enforceUniqueness(transactionContext, candidate); err != nil {
				return err
			}
			if err := service.repository.Create(transactionContext, candidate); err != nil {
				return err
			}
			change, err := changelog.NewChange(
				principal.ID,
				domainipam.IPAddressObjectType,
				candidate.ID(),
				candidate.Display(),
				changelog.ActionCreate,
				nil,
				candidate.Snapshot(),
				now,
			)
			if err != nil {
				return err
			}
			if err := service.recorder.Record(transactionContext, change); err != nil {
				return err
			}
			result, err = service.repository.Get(transactionContext, candidate.ID())
			return err
		},
	)
	if err != nil {
		return nil, normalizeOperationError("Could not create IP address.", err)
	}
	return result, nil
}

func (service *IPAddressService) ReplaceIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceIPAddressCommand,
) (*domainipam.IPAddress, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var result *domainipam.IPAddress
	err := service.unitOfWork.WithinTransaction(
		ctx,
		func(transactionContext context.Context) error {
			loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
			if err != nil {
				return err
			}
			if err := service.authorize(
				transactionContext, principal, authz.Change, loaded,
			); err != nil {
				return err
			}
			commandValues, err := command.values()
			if err != nil {
				return err
			}
			values, err := service.resolveValues(
				transactionContext,
				commandValues,
				domainipam.NullInterfaceAssignment(),
			)
			if err != nil {
				return err
			}
			before := loaded.Snapshot()
			now := service.clock.Now()
			if err := loaded.Replace(values, now); err != nil {
				return err
			}
			if err := service.enforceUniqueness(transactionContext, loaded); err != nil {
				return err
			}
			if err := service.repository.Update(transactionContext, loaded); err != nil {
				return err
			}
			if err := service.recordUpdate(
				transactionContext, principal, loaded, before, now,
			); err != nil {
				return err
			}
			result, err = service.repository.Get(transactionContext, loaded.ID())
			return err
		},
	)
	if err != nil {
		return nil, normalizeOperationError("Could not replace IP address.", err)
	}
	return result, nil
}

func (service *IPAddressService) UpdateIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command UpdateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	return service.updateIPAddress(ctx, principal, command)
}

func (service *IPAddressService) AssignIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command AssignIPAddressCommand,
) (*domainipam.IPAddress, error) {
	if !command.InterfaceID.IsValid() {
		return nil, invalidIPAddressInterfaceChoice(command.InterfaceID)
	}
	return service.updateIPAddress(ctx, principal, UpdateIPAddressCommand{
		ID: command.ID,
		AssignedObjectType: FieldValue(
			domainipam.IPAddressAssignmentType,
		),
		AssignedObjectID: FieldValue(command.InterfaceID.Int64()),
	})
}

func (service *IPAddressService) UnassignIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command UnassignIPAddressCommand,
) (*domainipam.IPAddress, error) {
	return service.updateIPAddress(ctx, principal, UpdateIPAddressCommand{
		ID: command.ID, AssignedObjectType: NullField[string](),
		AssignedObjectID: NullField[int64](),
	})
}

func (service *IPAddressService) updateIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command UpdateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var result *domainipam.IPAddress
	err := service.unitOfWork.WithinTransaction(
		ctx,
		func(transactionContext context.Context) error {
			loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
			if err != nil {
				return err
			}
			if err := service.authorize(
				transactionContext, principal, authz.Change, loaded,
			); err != nil {
				return err
			}
			commandPatch, err := command.patch()
			if err != nil {
				return err
			}
			patch, err := service.resolvePatch(
				transactionContext, commandPatch, loaded.Assignment(),
			)
			if err != nil {
				return err
			}
			before := loaded.Snapshot()
			now := service.clock.Now()
			if err := loaded.ApplyPatch(patch, now); err != nil {
				return err
			}
			if err := service.enforceUniqueness(transactionContext, loaded); err != nil {
				return err
			}
			if err := service.repository.Update(transactionContext, loaded); err != nil {
				return err
			}
			if err := service.recordUpdate(
				transactionContext, principal, loaded, before, now,
			); err != nil {
				return err
			}
			result, err = service.repository.Get(transactionContext, loaded.ID())
			return err
		},
	)
	if err != nil {
		return nil, normalizeOperationError("Could not update IP address.", err)
	}
	return result, nil
}

func (service *IPAddressService) DeleteIPAddress(
	ctx context.Context,
	principal identity.Principal,
	command DeleteIPAddressCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}
	err := service.unitOfWork.WithinTransaction(
		ctx,
		func(transactionContext context.Context) error {
			loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
			if err != nil {
				return err
			}
			if err := service.authorize(
				transactionContext, principal, authz.Delete, loaded,
			); err != nil {
				return err
			}
			before := loaded.Snapshot()
			now := service.clock.Now()
			if err := service.repository.Delete(transactionContext, loaded); err != nil {
				return err
			}
			change, err := changelog.NewChange(
				principal.ID,
				domainipam.IPAddressObjectType,
				loaded.ID(),
				loaded.Display(),
				changelog.ActionDelete,
				before,
				nil,
				now,
			)
			if err != nil {
				return err
			}
			return service.recorder.Record(transactionContext, change)
		},
	)
	return normalizeOperationError("Could not delete IP address.", err)
}

// DeleteAssignedToInterface is deliberately transaction- and recorder-free.
// InterfaceService owns the ambient transaction and records the returned child
// deletions before its parent deletion.
func (service *IPAddressService) DeleteAssignedToInterface(
	ctx context.Context,
	interfaceID shared.ID,
	now shared.Timestamp,
) ([]IPAddressCascadeChange, error) {
	if !interfaceID.IsValid() || now.IsZero() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot cascade IP addresses for an invalid Interface or time.",
		)
	}
	addresses, err := service.repository.ListAssignedToInterfaceForUpdate(
		ctx, interfaceID,
	)
	if err != nil {
		return nil, err
	}
	changes := make([]IPAddressCascadeChange, 0, len(addresses))
	for _, address := range addresses {
		if address == nil || !address.ID().IsValid() {
			return nil, shared.NewError(
				shared.ErrorReasonInternal,
				"IPAddress repository returned an invalid cascade item.",
			)
		}
		before := address.Snapshot()
		if err := service.repository.Delete(ctx, address); err != nil {
			return nil, err
		}
		changes = append(changes, IPAddressCascadeChange{
			ObjectType:     domainipam.IPAddressObjectType,
			ID:             address.ID(),
			Representation: address.Display(),
			Before:         before,
		})
	}
	return changes, nil
}

func (service *IPAddressService) resolveValues(
	ctx context.Context,
	values ipAddressCommandValues,
	currentAssignment domainipam.NullableInterfaceAssignment,
) (domainipam.IPAddressValues, error) {
	vrf, err := service.resolveIPAddressVRF(ctx, values.vrfID)
	if err != nil {
		return domainipam.IPAddressValues{}, err
	}
	assignment, err := service.resolveAssignment(
		ctx,
		currentAssignment,
		values.assignedObjectType,
		values.assignedObjectID,
		false,
	)
	if err != nil {
		return domainipam.IPAddressValues{}, err
	}
	return domainipam.IPAddressValues{
		Address: values.address, VRF: vrf, Status: values.status,
		Role: values.role, DNSName: values.dnsName,
		Description: values.description, Comments: values.comments,
		Assignment: assignment,
	}, nil
}

func (service *IPAddressService) resolvePatch(
	ctx context.Context,
	patch ipAddressCommandPatch,
	currentAssignment domainipam.NullableInterfaceAssignment,
) (domainipam.IPAddressPatch, error) {
	domainPatch := domainipam.IPAddressPatch{
		Address: patch.address, Status: patch.status, Role: patch.role,
		DNSName: patch.dnsName, Description: patch.description,
		Comments: patch.comments,
	}
	if patch.vrfSet {
		vrf, err := service.resolveIPAddressVRF(ctx, patch.vrfID)
		if err != nil {
			return domainipam.IPAddressPatch{}, err
		}
		domainPatch.VRF = &vrf
	}
	if patch.assignmentSet {
		assignment, err := service.resolveAssignment(
			ctx,
			currentAssignment,
			patch.assignedObjectType,
			patch.assignedObjectID,
			true,
		)
		if err != nil {
			return domainipam.IPAddressPatch{}, err
		}
		domainPatch.Assignment = &assignment
	}
	return domainPatch, nil
}

func (service *IPAddressService) resolveIPAddressVRF(
	ctx context.Context,
	id *shared.ID,
) (domainipam.NullableVRFReference, error) {
	if id == nil {
		return domainipam.NullVRFReference(), nil
	}
	vrf, err := service.vrfs.Get(ctx, *id)
	if err != nil {
		if shared.HasReason(err, shared.ErrorReasonNotFound) {
			return domainipam.NullableVRFReference{}, invalidIPAddressVRFChoice(*id)
		}
		return domainipam.NullableVRFReference{}, err
	}
	if vrf == nil {
		return domainipam.NullableVRFReference{}, shared.NewError(
			shared.ErrorReasonInternal, "VRF reader returned an invalid object.",
		)
	}
	reference, err := domainipam.NewVRFReference(
		vrf.ID(), vrf.Name(), vrf.RD(), vrf.EnforceUnique(),
	)
	if err != nil {
		return domainipam.NullableVRFReference{}, err
	}
	return domainipam.NonNullVRFReference(reference), nil
}

func (service *IPAddressService) resolveAssignment(
	ctx context.Context,
	current domainipam.NullableInterfaceAssignment,
	objectType Field[string],
	objectID Field[int64],
	preserveCurrent bool,
) (domainipam.NullableInterfaceAssignment, error) {
	var resolvedType *string
	var resolvedID *shared.ID
	if preserveCurrent {
		if assignment, present := current.Get(); present {
			valueType := domainipam.IPAddressAssignmentType
			valueID := assignment.ID()
			resolvedType, resolvedID = &valueType, &valueID
		}
	}
	switch objectType.State() {
	case FieldNull:
		resolvedType = nil
	case FieldPresent:
		value, _ := objectType.Get()
		value = strings.TrimSpace(value)
		resolvedType = &value
	}
	switch objectID.State() {
	case FieldNull:
		resolvedID = nil
	case FieldPresent:
		value, _ := objectID.Get()
		id := shared.ID(value)
		resolvedID = &id
	}
	if resolvedType == nil && resolvedID == nil {
		return domainipam.NullInterfaceAssignment(), nil
	}
	if resolvedType == nil {
		return domainipam.NullableInterfaceAssignment{},
			shared.NewValidationError(shared.FieldViolation{
				Field: "assigned_object_type", Reason: "required",
				Description: "This field is required when assigned_object_id is set.",
			})
	}
	if *resolvedType != domainipam.IPAddressAssignmentType {
		return domainipam.NullableInterfaceAssignment{},
			shared.NewValidationError(shared.FieldViolation{
				Field: "assigned_object_type", Reason: "invalid_choice",
				Description: "Only dcim.interface assignments are supported.",
			})
	}
	if resolvedID == nil || !resolvedID.IsValid() {
		return domainipam.NullableInterfaceAssignment{},
			shared.NewValidationError(shared.FieldViolation{
				Field: "assigned_object_id", Reason: "required",
				Description: "A valid Interface ID is required.",
			})
	}
	networkInterface, err := service.interfaces.Get(ctx, *resolvedID)
	if err != nil {
		if shared.HasReason(err, shared.ErrorReasonNotFound) {
			return domainipam.NullableInterfaceAssignment{},
				invalidIPAddressInterfaceChoice(*resolvedID)
		}
		return domainipam.NullableInterfaceAssignment{}, err
	}
	if networkInterface == nil {
		return domainipam.NullableInterfaceAssignment{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Interface reader returned an invalid object.",
		)
	}
	assignment, err := domainipam.NewInterfaceAssignment(networkInterface)
	if err != nil {
		return domainipam.NullableInterfaceAssignment{}, err
	}
	return domainipam.NonNullInterfaceAssignment(assignment), nil
}

func (service *IPAddressService) enforceUniqueness(
	ctx context.Context,
	address *domainipam.IPAddress,
) error {
	if address == nil {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot check uniqueness for a nil IPAddress.",
		)
	}
	vrf := address.VRF()
	if reference, present := vrf.Get(); present && !reference.EnforceUnique() {
		return nil
	}
	if err := service.repository.LockUniqueness(
		ctx, vrf, address.Address(),
	); err != nil {
		return err
	}
	duplicates, err := service.repository.FindDuplicates(
		ctx, vrf, address.Address(), address.ID(),
	)
	if err != nil {
		return err
	}
	if len(duplicates) == 0 {
		return nil
	}
	role, rolePresent := address.Role().Get()
	if rolePresent && role.AllowsDuplicateHost() {
		allExceptional := true
		for _, duplicate := range duplicates {
			if duplicate == nil {
				return shared.NewError(
					shared.ErrorReasonInternal,
					"IPAddress repository returned an invalid duplicate.",
				)
			}
			duplicateRole, present := duplicate.Role().Get()
			if !present || !duplicateRole.AllowsDuplicateHost() {
				allExceptional = false
				break
			}
		}
		if allExceptional {
			return nil
		}
	}
	table := "global table"
	if reference, present := vrf.Get(); present {
		table = "VRF " + reference.Display()
	}
	return shared.NewValidationError(shared.FieldViolation{
		Field: "address", Reason: "unique",
		Description: "Duplicate IP address found in " + table + ": " +
			duplicates[0].Display(),
	})
}

func (service *IPAddressService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	address *domainipam.IPAddress,
	before domainipam.IPAddressSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID,
		domainipam.IPAddressObjectType,
		address.ID(),
		address.Display(),
		changelog.ActionUpdate,
		before,
		address.Snapshot(),
		now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *IPAddressService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	address *domainipam.IPAddress,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if address != nil {
		object = authz.NewObject(address.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceIPAddress,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}

func ipAddressSharedIDs(values []int64) []shared.ID {
	ids := make([]shared.ID, len(values))
	for index, value := range values {
		ids[index] = shared.ID(value)
	}
	return ids
}
