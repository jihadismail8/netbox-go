package dcim

import (
	"context"
	"strconv"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type RackTypeService struct {
	repository    RackTypeRepository
	manufacturers RackTypeManufacturerReader
	unitOfWork    transaction.UnitOfWork
	recorder      changelog.Recorder
	authorizer    authz.ResourceAuthorizer
	clock         shared.Clock
}

func NewRackTypeService(
	repository RackTypeRepository,
	manufacturers RackTypeManufacturerReader,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*RackTypeService, error) {
	missing := make([]string, 0, 6)
	if nilInterface(repository) {
		missing = append(missing, "repository")
	}
	if nilInterface(manufacturers) {
		missing = append(missing, "manufacturer reader")
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
			"RackType service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &RackTypeService{
		repository: repository, manufacturers: manufacturers, unitOfWork: unitOfWork,
		recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *RackTypeService) ListRackTypes(
	ctx context.Context,
	principal identity.Principal,
	query ListRackTypesQuery,
) (RackTypePage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return RackTypePage{}, err
	}
	criteria, err := validateListRackTypesQuery(query)
	if err != nil {
		return RackTypePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceRackType,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return RackTypePage{}, normalizeOperationError("Could not list RackTypes.", err)
	}
	visible := make([]*dcimdomain.RackType, 0, len(page.Results))
	for _, rackType := range page.Results {
		if rackType == nil {
			return RackTypePage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"RackType repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, rackType)
		if authorizeErr == nil {
			visible = append(visible, rackType)
			continue
		}
		if hasCompleteScope {
			return RackTypePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"RackType visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return RackTypePage{}, authorizeErr
	}
	if hasCompleteScope {
		return RackTypePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return RackTypePage{Count: count, Results: visible[start:end]}, nil
}

func (service *RackTypeService) GetRackType(
	ctx context.Context,
	principal identity.Principal,
	query GetRackTypeQuery,
) (*dcimdomain.RackType, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	rackType, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get RackType.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, rackType); err != nil {
		return nil, err
	}
	return rackType, nil
}

func (service *RackTypeService) CreateRackType(
	ctx context.Context,
	principal identity.Principal,
	command CreateRackTypeCommand,
) (*dcimdomain.RackType, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var rackType *dcimdomain.RackType
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		commandValues, commandErr := command.values()
		values, relationshipErr := service.resolveValues(transactionContext, commandValues)
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewRackType(values, now)
		if validationErr := mergeRackTypeMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Add, candidate); authorizeErr != nil {
			return authorizeErr
		}
		if createErr := service.repository.Create(transactionContext, candidate); createErr != nil {
			return createErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.RackTypeObjectType, candidate.ID(), candidate.Display(),
			changelog.ActionCreate, nil, candidate.Snapshot(), now,
		)
		if changeErr != nil {
			return changeErr
		}
		if recordErr := service.recorder.Record(transactionContext, change); recordErr != nil {
			return recordErr
		}
		rackType = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create RackType.", err)
	}
	return rackType, nil
}

func (service *RackTypeService) ReplaceRackType(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceRackTypeCommand,
) (*dcimdomain.RackType, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var rackType *dcimdomain.RackType
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
		commandPatch, commandErr := command.patch()
		patch, relationshipErr := service.resolvePatch(transactionContext, commandPatch)
		domainErr := loaded.ValidatePatch(patch)
		if validationErr := mergeRackTypeMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		if replaceErr := loaded.ApplyPatch(patch, now); replaceErr != nil {
			return replaceErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if recordErr := service.recordRackTypeUpdate(transactionContext, principal, loaded, before, now); recordErr != nil {
			return recordErr
		}
		if propagationErr := service.propagate(transactionContext, principal, loaded, now); propagationErr != nil {
			return propagationErr
		}
		rackType = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace RackType.", err)
	}
	return rackType, nil
}

func (service *RackTypeService) UpdateRackType(
	ctx context.Context,
	principal identity.Principal,
	command UpdateRackTypeCommand,
) (*dcimdomain.RackType, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var rackType *dcimdomain.RackType
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
		commandPatch, commandErr := command.patch()
		patch, relationshipErr := service.resolvePatch(transactionContext, commandPatch)
		domainErr := loaded.ValidatePatch(patch)
		if validationErr := mergeRackTypeMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		if domainErr := loaded.ApplyPatch(patch, now); domainErr != nil {
			return domainErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if recordErr := service.recordRackTypeUpdate(transactionContext, principal, loaded, before, now); recordErr != nil {
			return recordErr
		}
		if propagationErr := service.propagate(transactionContext, principal, loaded, now); propagationErr != nil {
			return propagationErr
		}
		rackType = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update RackType.", err)
	}
	return rackType, nil
}

func (service *RackTypeService) DeleteRackType(
	ctx context.Context,
	principal identity.Principal,
	command DeleteRackTypeCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Delete, loaded); authorizeErr != nil {
			return authorizeErr
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, loaded); deleteErr != nil {
			return deleteErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.RackTypeObjectType, loaded.ID(), loaded.Display(),
			changelog.ActionDelete, before, nil, now,
		)
		if changeErr != nil {
			return changeErr
		}
		return service.recorder.Record(transactionContext, change)
	})
	return normalizeOperationError("Could not delete RackType.", err)
}

func (service *RackTypeService) resolveValues(
	ctx context.Context,
	values rackTypeCommandValues,
) (dcimdomain.RackTypeValues, error) {
	domainValues := dcimdomain.RackTypeValues{
		Model: values.model, Slug: values.slug,
		FormFactor: values.formFactor, Width: values.width, UHeight: values.uHeight,
		StartingUnit: values.startingUnit, DescUnits: values.descUnits,
		Description: values.description, Comments: values.comments,
	}
	if !values.manufacturerID.IsValid() {
		return domainValues, nil
	}
	reference, err := service.resolveManufacturer(ctx, values.manufacturerID)
	if err != nil {
		return domainValues, err
	}
	domainValues.Manufacturer = reference
	return domainValues, nil
}

func (service *RackTypeService) resolvePatch(
	ctx context.Context,
	patch rackTypeCommandPatch,
) (dcimdomain.RackTypePatch, error) {
	domainPatch := dcimdomain.RackTypePatch{
		Model: patch.model, Slug: patch.slug, FormFactor: patch.formFactor,
		Width: patch.width, UHeight: patch.uHeight, StartingUnit: patch.startingUnit,
		DescUnits: patch.descUnits, Description: patch.description, Comments: patch.comments,
	}
	if patch.manufacturerID != nil {
		if !patch.manufacturerID.IsValid() {
			return domainPatch, nil
		}
		reference, err := service.resolveManufacturer(ctx, *patch.manufacturerID)
		if err != nil {
			return domainPatch, err
		}
		domainPatch.Manufacturer = &reference
	}
	return domainPatch, nil
}

func (service *RackTypeService) resolveManufacturer(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.ManufacturerReference, error) {
	manufacturer, err := service.manufacturers.Get(ctx, id)
	if err != nil {
		if shared.HasReason(err, shared.ErrorReasonNotFound) {
			return dcimdomain.ManufacturerReference{}, shared.NewValidationError(shared.FieldViolation{
				Field: "manufacturer", Reason: "invalid_choice",
				Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) + "\" - object does not exist.",
			})
		}
		return dcimdomain.ManufacturerReference{}, err
	}
	return dcimdomain.NewManufacturerReference(
		manufacturer.ID(), manufacturer.Name(), manufacturer.Slug().String(),
	)
}

func (service *RackTypeService) recordRackTypeUpdate(
	ctx context.Context,
	principal identity.Principal,
	rackType *dcimdomain.RackType,
	before dcimdomain.RackTypeSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.RackTypeObjectType, rackType.ID(), rackType.Display(),
		changelog.ActionUpdate, before, rackType.Snapshot(), now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *RackTypeService) propagate(
	ctx context.Context,
	principal identity.Principal,
	rackType *dcimdomain.RackType,
	now shared.Timestamp,
) error {
	changes, err := service.repository.PropagateToRacks(
		ctx, rackType.ID(), rackType.PhysicalAttributes(), now,
	)
	if err != nil {
		return err
	}
	for _, propagated := range changes {
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.RackObjectType, propagated.ID, propagated.Representation,
			changelog.ActionUpdate, propagated.Before, propagated.After, now,
		)
		if changeErr != nil {
			return changeErr
		}
		if recordErr := service.recorder.Record(ctx, change); recordErr != nil {
			return recordErr
		}
	}
	return nil
}

func (service *RackTypeService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	rackType *dcimdomain.RackType,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if rackType == nil {
		object = nil
	} else {
		object = authz.NewObject(rackType.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceRackType,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}
