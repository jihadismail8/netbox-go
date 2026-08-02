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

type PrefixService struct {
	repository PrefixRepository
	vrfs       PrefixVRFReader
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewPrefixService(
	repository PrefixRepository,
	vrfs PrefixVRFReader,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*PrefixService, error) {
	missing := make([]string, 0, 6)
	if nilInterface(repository) {
		missing = append(missing, "repository")
	}
	if nilInterface(vrfs) {
		missing = append(missing, "VRF reader")
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
			"Prefix service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &PrefixService{
		repository: repository, vrfs: vrfs, unitOfWork: unitOfWork,
		recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *PrefixService) ListPrefixes(
	ctx context.Context,
	principal identity.Principal,
	query ListPrefixesQuery,
) (PrefixPage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return PrefixPage{}, err
	}
	criteria, err := validateListPrefixesQuery(query)
	if err != nil {
		return PrefixPage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourcePrefix,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = prefixSharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return PrefixPage{}, normalizeOperationError("Could not list Prefixes.", err)
	}
	visible := make([]*domainipam.Prefix, 0, len(page.Results))
	for _, prefix := range page.Results {
		if prefix == nil {
			return PrefixPage{}, shared.NewError(
				shared.ErrorReasonInternal, "Prefix repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, prefix)
		if authorizeErr == nil {
			visible = append(visible, prefix)
			continue
		}
		if hasCompleteScope {
			return PrefixPage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"Prefix visibility scope admitted an unauthorized object.", authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return PrefixPage{}, authorizeErr
	}
	if hasCompleteScope {
		return PrefixPage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return PrefixPage{Count: count, Results: visible[start:end]}, nil
}

func (service *PrefixService) GetPrefix(
	ctx context.Context,
	principal identity.Principal,
	query GetPrefixQuery,
) (*domainipam.Prefix, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	prefix, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get Prefix.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, prefix); err != nil {
		return nil, err
	}
	return prefix, nil
}

func (service *PrefixService) CreatePrefix(
	ctx context.Context,
	principal identity.Principal,
	command CreatePrefixCommand,
) (*domainipam.Prefix, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var result *domainipam.Prefix
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
		candidate, err := domainipam.NewPrefix(values, now)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Add, candidate); err != nil {
			return err
		}
		if err := service.enforceUniqueness(transactionContext, candidate); err != nil {
			return err
		}
		if err := service.repository.Create(transactionContext, candidate); err != nil {
			return err
		}
		change, err := changelog.NewChange(
			principal.ID, domainipam.PrefixObjectType, candidate.ID(), candidate.Display(),
			changelog.ActionCreate, nil, candidate.Snapshot(), now,
		)
		if err != nil {
			return err
		}
		if err := service.recorder.Record(transactionContext, change); err != nil {
			return err
		}
		result, err = service.repository.Get(transactionContext, candidate.ID())
		return err
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create Prefix.", err)
	}
	return result, nil
}

func (service *PrefixService) ReplacePrefix(
	ctx context.Context,
	principal identity.Principal,
	command ReplacePrefixCommand,
) (*domainipam.Prefix, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var result *domainipam.Prefix
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Change, loaded); err != nil {
			return err
		}
		commandValues, err := command.values()
		if err != nil {
			return err
		}
		values, err := service.resolveValues(transactionContext, commandValues)
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
		if err := service.recordUpdate(transactionContext, principal, loaded, before, now); err != nil {
			return err
		}
		result, err = service.repository.Get(transactionContext, loaded.ID())
		return err
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace Prefix.", err)
	}
	return result, nil
}

func (service *PrefixService) UpdatePrefix(
	ctx context.Context,
	principal identity.Principal,
	command UpdatePrefixCommand,
) (*domainipam.Prefix, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var result *domainipam.Prefix
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Change, loaded); err != nil {
			return err
		}
		commandPatch, err := command.patch()
		if err != nil {
			return err
		}
		patch, err := service.resolvePatch(transactionContext, commandPatch)
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
		if err := service.recordUpdate(transactionContext, principal, loaded, before, now); err != nil {
			return err
		}
		result, err = service.repository.Get(transactionContext, loaded.ID())
		return err
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update Prefix.", err)
	}
	return result, nil
}

func (service *PrefixService) DeletePrefix(
	ctx context.Context,
	principal identity.Principal,
	command DeletePrefixCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
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
		if err := service.authorize(transactionContext, principal, authz.Delete, loaded); err != nil {
			return err
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if err := service.repository.Delete(transactionContext, loaded); err != nil {
			return err
		}
		change, err := changelog.NewChange(
			principal.ID, domainipam.PrefixObjectType, loaded.ID(), loaded.Display(),
			changelog.ActionDelete, before, nil, now,
		)
		if err != nil {
			return err
		}
		return service.recorder.Record(transactionContext, change)
	})
	return normalizeOperationError("Could not delete Prefix.", err)
}

func (service *PrefixService) resolveValues(
	ctx context.Context,
	values prefixCommandValues,
) (domainipam.PrefixValues, error) {
	vrf, err := service.resolveVRF(ctx, values.vrfID)
	if err != nil {
		return domainipam.PrefixValues{}, err
	}
	return domainipam.PrefixValues{
		Prefix: values.prefix, VRF: vrf, Status: values.status,
		IsPool: values.isPool, MarkUtilized: values.markUtilized,
		Description: values.description, Comments: values.comments,
	}, nil
}

func (service *PrefixService) resolvePatch(
	ctx context.Context,
	patch prefixCommandPatch,
) (domainipam.PrefixPatch, error) {
	domainPatch := domainipam.PrefixPatch{
		Prefix: patch.prefix, Status: patch.status, IsPool: patch.isPool,
		MarkUtilized: patch.markUtilized, Description: patch.description, Comments: patch.comments,
	}
	if patch.vrfSet {
		vrf, err := service.resolveVRF(ctx, patch.vrfID)
		if err != nil {
			return domainipam.PrefixPatch{}, err
		}
		domainPatch.VRF = &vrf
	}
	return domainPatch, nil
}

func (service *PrefixService) resolveVRF(
	ctx context.Context,
	id *shared.ID,
) (domainipam.NullableVRFReference, error) {
	if id == nil {
		return domainipam.NullVRFReference(), nil
	}
	vrf, err := service.vrfs.Get(ctx, *id)
	if err != nil {
		if shared.HasReason(err, shared.ErrorReasonNotFound) {
			return domainipam.NullableVRFReference{}, invalidVRFChoice(*id)
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

func (service *PrefixService) enforceUniqueness(ctx context.Context, prefix *domainipam.Prefix) error {
	if prefix == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot check uniqueness for a nil Prefix.")
	}
	vrf := prefix.VRF()
	if reference, present := vrf.Get(); present && !reference.EnforceUnique() {
		return nil
	}
	if err := service.repository.LockUniqueness(ctx, vrf, prefix.Network()); err != nil {
		return err
	}
	duplicate, err := service.repository.FindDuplicate(ctx, vrf, prefix.Network(), prefix.ID())
	if err != nil {
		return err
	}
	if duplicate == nil {
		return nil
	}
	table := "global table"
	if reference, present := vrf.Get(); present {
		table = "VRF " + reference.Display()
	}
	return shared.NewValidationError(shared.FieldViolation{
		Field: "prefix", Reason: "unique",
		Description: "Duplicate prefix found in " + table + ": " + duplicate.Display(),
	})
}

func (service *PrefixService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	prefix *domainipam.Prefix,
	before domainipam.PrefixSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, domainipam.PrefixObjectType, prefix.ID(), prefix.Display(),
		changelog.ActionUpdate, before, prefix.Snapshot(), now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *PrefixService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	prefix *domainipam.Prefix,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated, "Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if prefix != nil {
		object = authz.NewObject(prefix.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourcePrefix,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}

func prefixSharedIDs(values []int64) []shared.ID {
	ids := make([]shared.ID, len(values))
	for index, value := range values {
		ids[index] = shared.ID(value)
	}
	return ids
}
