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

type RackRoleService struct {
	repository RackRoleRepository
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewRackRoleService(
	repository RackRoleRepository,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*RackRoleService, error) {
	missing := make([]string, 0, 5)
	if nilInterface(repository) {
		missing = append(missing, "repository")
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
			"RackRole service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &RackRoleService{
		repository: repository, unitOfWork: unitOfWork, recorder: recorder,
		authorizer: authorizer, clock: clock,
	}, nil
}

func (service *RackRoleService) ListRackRoles(
	ctx context.Context,
	principal identity.Principal,
	query ListRackRolesQuery,
) (RackRolePage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return RackRolePage{}, err
	}
	criteria, err := validateListRackRolesQuery(query)
	if err != nil {
		return RackRolePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceRackRole,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return RackRolePage{}, normalizeOperationError("Could not list RackRoles.", err)
	}
	visible := make([]*dcimdomain.RackRole, 0, len(page.Results))
	for _, role := range page.Results {
		if role == nil {
			return RackRolePage{}, shared.NewError(
				shared.ErrorReasonInternal, "RackRole repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, role)
		if authorizeErr == nil {
			visible = append(visible, role)
			continue
		}
		if hasCompleteScope {
			return RackRolePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"RackRole visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return RackRolePage{}, authorizeErr
	}
	if hasCompleteScope {
		return RackRolePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return RackRolePage{Count: count, Results: visible[start:end]}, nil
}

func (service *RackRoleService) GetRackRole(
	ctx context.Context,
	principal identity.Principal,
	query GetRackRoleQuery,
) (*dcimdomain.RackRole, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	role, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get RackRole.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (service *RackRoleService) CreateRackRole(
	ctx context.Context,
	principal identity.Principal,
	command CreateRackRoleCommand,
) (*dcimdomain.RackRole, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var role *dcimdomain.RackRole
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		values, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewRackRole(values, now)
		if domainErr != nil {
			return domainErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Add, candidate); authorizeErr != nil {
			return authorizeErr
		}
		if createErr := service.repository.Create(transactionContext, candidate); createErr != nil {
			return createErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.RackRoleObjectType, candidate.ID(), candidate.Display(),
			changelog.ActionCreate, nil, candidate.Snapshot(), now,
		)
		if changeErr != nil {
			return changeErr
		}
		if recordErr := service.recorder.Record(transactionContext, change); recordErr != nil {
			return recordErr
		}
		role = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create RackRole.", err)
	}
	return role, nil
}

func (service *RackRoleService) ReplaceRackRole(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceRackRoleCommand,
) (*dcimdomain.RackRole, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var role *dcimdomain.RackRole
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Change, loaded); authorizeErr != nil {
			return authorizeErr
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		values, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		if replaceErr := loaded.Replace(values, now); replaceErr != nil {
			return replaceErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if changeErr := service.recordUpdate(transactionContext, principal, loaded, before, now); changeErr != nil {
			return changeErr
		}
		role = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace RackRole.", err)
	}
	return role, nil
}

func (service *RackRoleService) UpdateRackRole(
	ctx context.Context,
	principal identity.Principal,
	command UpdateRackRoleCommand,
) (*dcimdomain.RackRole, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var role *dcimdomain.RackRole
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Change, loaded); authorizeErr != nil {
			return authorizeErr
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		patch, patchErr := command.patch()
		if patchErr != nil {
			return patchErr
		}
		if domainErr := loaded.ApplyPatch(patch, now); domainErr != nil {
			return domainErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if changeErr := service.recordUpdate(transactionContext, principal, loaded, before, now); changeErr != nil {
			return changeErr
		}
		role = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update RackRole.", err)
	}
	return role, nil
}

func (service *RackRoleService) DeleteRackRole(
	ctx context.Context,
	principal identity.Principal,
	command DeleteRackRoleCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		role, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Delete, role); authorizeErr != nil {
			return authorizeErr
		}
		before := role.Snapshot()
		now := service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, role); deleteErr != nil {
			return deleteErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.RackRoleObjectType, role.ID(), role.Display(),
			changelog.ActionDelete, before, nil, now,
		)
		if changeErr != nil {
			return changeErr
		}
		return service.recorder.Record(transactionContext, change)
	})
	return normalizeOperationError("Could not delete RackRole.", err)
}

func (service *RackRoleService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	role *dcimdomain.RackRole,
	before dcimdomain.RackRoleSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.RackRoleObjectType, role.ID(), role.Display(),
		changelog.ActionUpdate, before, role.Snapshot(), now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *RackRoleService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	role *dcimdomain.RackRole,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if role == nil {
		object = nil
	} else {
		object = authz.NewObject(role.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceRackRole,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}
