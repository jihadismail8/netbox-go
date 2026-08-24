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

type ManufacturerService struct {
	repository ManufacturerRepository
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewManufacturerService(
	repository ManufacturerRepository,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*ManufacturerService, error) {
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
			"Manufacturer service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &ManufacturerService{
		repository: repository, unitOfWork: unitOfWork, recorder: recorder,
		authorizer: authorizer, clock: clock,
	}, nil
}

func (service *ManufacturerService) ListManufacturers(
	ctx context.Context,
	principal identity.Principal,
	query ListManufacturersQuery,
) (ManufacturerPage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return ManufacturerPage{}, err
	}
	criteria, err := validateListManufacturersQuery(query)
	if err != nil {
		return ManufacturerPage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceManufacturer,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return ManufacturerPage{}, normalizeOperationError("Could not list Manufacturers.", err)
	}
	visible := make([]*dcimdomain.Manufacturer, 0, len(page.Results))
	for _, manufacturer := range page.Results {
		if manufacturer == nil {
			return ManufacturerPage{}, shared.NewError(
				shared.ErrorReasonInternal, "Manufacturer repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, manufacturer)
		if authorizeErr == nil {
			visible = append(visible, manufacturer)
			continue
		}
		if hasCompleteScope {
			return ManufacturerPage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"Manufacturer visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return ManufacturerPage{}, authorizeErr
	}
	if hasCompleteScope {
		return ManufacturerPage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return ManufacturerPage{Count: count, Results: visible[start:end]}, nil
}

func (service *ManufacturerService) GetManufacturer(
	ctx context.Context,
	principal identity.Principal,
	query GetManufacturerQuery,
) (*dcimdomain.Manufacturer, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	manufacturer, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get Manufacturer.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, manufacturer); err != nil {
		return nil, err
	}
	return manufacturer, nil
}

func (service *ManufacturerService) CreateManufacturer(
	ctx context.Context,
	principal identity.Principal,
	command CreateManufacturerCommand,
) (*dcimdomain.Manufacturer, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var manufacturer *dcimdomain.Manufacturer
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		values, commandErr := command.values()
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewManufacturer(values, now)
		if validationErr := mergeManufacturerMutationErrors(commandErr, domainErr); validationErr != nil {
			return validationErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Add, candidate); authorizeErr != nil {
			return authorizeErr
		}
		if createErr := service.repository.Create(transactionContext, candidate); createErr != nil {
			return createErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.ManufacturerObjectType, candidate.ID(), candidate.Display(),
			changelog.ActionCreate, nil, candidate.Snapshot(), now,
		)
		if changeErr != nil {
			return changeErr
		}
		if recordErr := service.recorder.Record(transactionContext, change); recordErr != nil {
			return recordErr
		}
		manufacturer = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create Manufacturer.", err)
	}
	return manufacturer, nil
}

func (service *ManufacturerService) ReplaceManufacturer(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceManufacturerCommand,
) (*dcimdomain.Manufacturer, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var manufacturer *dcimdomain.Manufacturer
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
		if now.IsZero() {
			return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
		}
		patch, commandErr := command.patch()
		domainErr := loaded.ValidatePatch(patch)
		if validationErr := mergeManufacturerMutationErrors(commandErr, domainErr); validationErr != nil {
			return validationErr
		}
		if replaceErr := loaded.ApplyPatch(patch, now); replaceErr != nil {
			return replaceErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if changeErr := service.recordUpdate(transactionContext, principal, loaded, before, now); changeErr != nil {
			return changeErr
		}
		manufacturer = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace Manufacturer.", err)
	}
	return manufacturer, nil
}

func (service *ManufacturerService) UpdateManufacturer(
	ctx context.Context,
	principal identity.Principal,
	command UpdateManufacturerCommand,
) (*dcimdomain.Manufacturer, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var manufacturer *dcimdomain.Manufacturer
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
		if now.IsZero() {
			return shared.NewError(shared.ErrorReasonInternal, "Clock returned a zero timestamp.")
		}
		patch, commandErr := command.patch()
		domainErr := loaded.ValidatePatch(patch)
		if validationErr := mergeManufacturerMutationErrors(commandErr, domainErr); validationErr != nil {
			return validationErr
		}
		if patchErr := loaded.ApplyPatch(patch, now); patchErr != nil {
			return patchErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if changeErr := service.recordUpdate(transactionContext, principal, loaded, before, now); changeErr != nil {
			return changeErr
		}
		manufacturer = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update Manufacturer.", err)
	}
	return manufacturer, nil
}

func (service *ManufacturerService) DeleteManufacturer(
	ctx context.Context,
	principal identity.Principal,
	command DeleteManufacturerCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		manufacturer, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Delete, manufacturer); authorizeErr != nil {
			return authorizeErr
		}
		before := manufacturer.Snapshot()
		now := service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, manufacturer); deleteErr != nil {
			return deleteErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.ManufacturerObjectType, manufacturer.ID(), manufacturer.Display(),
			changelog.ActionDelete, before, nil, now,
		)
		if changeErr != nil {
			return changeErr
		}
		return service.recorder.Record(transactionContext, change)
	})
	return normalizeOperationError("Could not delete Manufacturer.", err)
}

func (service *ManufacturerService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	manufacturer *dcimdomain.Manufacturer,
	before dcimdomain.ManufacturerSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.ManufacturerObjectType, manufacturer.ID(), manufacturer.Display(),
		changelog.ActionUpdate, before, manufacturer.Snapshot(), now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *ManufacturerService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	manufacturer *dcimdomain.Manufacturer,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if manufacturer == nil {
		object = nil
	} else {
		object = authz.NewObject(manufacturer.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceManufacturer,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}
