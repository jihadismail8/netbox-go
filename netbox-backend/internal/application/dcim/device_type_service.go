package dcim

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DeviceTypeService struct {
	repository    DeviceTypeRepository
	manufacturers DeviceTypeManufacturerReader
	unitOfWork    transaction.UnitOfWork
	recorder      changelog.Recorder
	authorizer    authz.ResourceAuthorizer
	clock         shared.Clock
}

func NewDeviceTypeService(
	repository DeviceTypeRepository,
	manufacturers DeviceTypeManufacturerReader,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*DeviceTypeService, error) {
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
			"DeviceType service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &DeviceTypeService{
		repository: repository, manufacturers: manufacturers, unitOfWork: unitOfWork,
		recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *DeviceTypeService) ListDeviceTypes(
	ctx context.Context,
	principal identity.Principal,
	query ListDeviceTypesQuery,
) (DeviceTypePage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return DeviceTypePage{}, err
	}
	criteria, err := validateListDeviceTypesQuery(query)
	if err != nil {
		return DeviceTypePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceDeviceType,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return DeviceTypePage{}, normalizeOperationError("Could not list DeviceTypes.", err)
	}
	visible := make([]*dcimdomain.DeviceType, 0, len(page.Results))
	for _, deviceType := range page.Results {
		if deviceType == nil {
			return DeviceTypePage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"DeviceType repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, deviceType)
		if authorizeErr == nil {
			visible = append(visible, deviceType)
			continue
		}
		if hasCompleteScope {
			return DeviceTypePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"DeviceType visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return DeviceTypePage{}, authorizeErr
	}
	if hasCompleteScope {
		return DeviceTypePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return DeviceTypePage{Count: count, Results: visible[start:end]}, nil
}

func (service *DeviceTypeService) GetDeviceType(
	ctx context.Context,
	principal identity.Principal,
	query GetDeviceTypeQuery,
) (*dcimdomain.DeviceType, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	deviceType, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get DeviceType.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, deviceType); err != nil {
		return nil, err
	}
	return deviceType, nil
}

func (service *DeviceTypeService) CreateDeviceType(
	ctx context.Context,
	principal identity.Principal,
	command CreateDeviceTypeCommand,
) (*dcimdomain.DeviceType, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var result *dcimdomain.DeviceType
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		commandValues, commandErr := command.values()
		values, relationshipErr := service.resolveValues(transactionContext, commandValues)
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewDeviceType(values, now)
		if validationErr := mergeDeviceTypeMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		if err := service.authorize(transactionContext, principal, authz.Add, candidate); err != nil {
			return err
		}
		if err := service.repository.Create(transactionContext, candidate); err != nil {
			return err
		}
		if err := service.recordDeviceTypeChange(
			transactionContext, principal, candidate, changelog.ActionCreate, nil,
			candidate.Snapshot(), now,
		); err != nil {
			return err
		}
		result = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create DeviceType.", err)
	}
	return result, nil
}

func (service *DeviceTypeService) ReplaceDeviceType(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceDeviceTypeCommand,
) (*dcimdomain.DeviceType, error) {
	return service.mutateDeviceType(
		ctx, principal, command.ID,
		func(
			transactionContext context.Context,
			loaded *dcimdomain.DeviceType,
			now shared.Timestamp,
		) error {
			commandPatch, commandErr := command.patch()
			patch, relationshipErr := service.resolvePatch(transactionContext, commandPatch)
			domainErr := loaded.ValidatePatch(patch)
			if validationErr := mergeDeviceTypeMutationErrors(
				commandErr, relationshipErr, domainErr,
			); validationErr != nil {
				return validationErr
			}
			return loaded.ApplyPatch(patch, now)
		},
		"replace",
	)
}

func (service *DeviceTypeService) UpdateDeviceType(
	ctx context.Context,
	principal identity.Principal,
	command UpdateDeviceTypeCommand,
) (*dcimdomain.DeviceType, error) {
	return service.mutateDeviceType(
		ctx, principal, command.ID,
		func(
			transactionContext context.Context,
			loaded *dcimdomain.DeviceType,
			now shared.Timestamp,
		) error {
			commandPatch, commandErr := command.patch()
			patch, relationshipErr := service.resolvePatch(transactionContext, commandPatch)
			domainErr := loaded.ValidatePatch(patch)
			if validationErr := mergeDeviceTypeMutationErrors(
				commandErr, relationshipErr, domainErr,
			); validationErr != nil {
				return validationErr
			}
			return loaded.ApplyPatch(patch, now)
		},
		"update",
	)
}

func (service *DeviceTypeService) mutateDeviceType(
	ctx context.Context,
	principal identity.Principal,
	id shared.ID,
	mutation func(context.Context, *dcimdomain.DeviceType, shared.Timestamp) error,
	operation string,
) (*dcimdomain.DeviceType, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(id); err != nil {
		return nil, err
	}
	var result *dcimdomain.DeviceType
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, id)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Change, loaded); err != nil {
			return err
		}
		before := loaded.Snapshot()
		originalHeight := loaded.UHeight()
		now := service.clock.Now()
		if err := mutation(transactionContext, loaded, now); err != nil {
			return err
		}
		if err := service.validateHeightTransition(
			transactionContext, loaded, originalHeight,
		); err != nil {
			return err
		}
		if err := service.repository.Update(transactionContext, loaded); err != nil {
			return err
		}
		if err := service.recordDeviceTypeChange(
			transactionContext, principal, loaded, changelog.ActionUpdate,
			before, loaded.Snapshot(), now,
		); err != nil {
			return err
		}
		result = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not "+operation+" DeviceType.", err)
	}
	return result, nil
}

func (service *DeviceTypeService) DeleteDeviceType(
	ctx context.Context,
	principal identity.Principal,
	command DeleteDeviceTypeCommand,
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
		dependent, err := service.repository.FindDeviceUsingDeviceType(
			transactionContext, command.ID,
		)
		if err != nil {
			return err
		}
		if dependent != nil {
			return shared.NewError(
				shared.ErrorReasonProtected,
				fmt.Sprintf(
					"Unable to delete object. 1 dependent objects were found: %s (%s)",
					dependent.Display, dependent.ID,
				),
			)
		}
		templates, err := service.repository.ListInterfaceTemplatesForUpdate(
			transactionContext, command.ID,
		)
		if err != nil {
			return err
		}
		for _, template := range templates {
			now := service.clock.Now()
			if err := service.repository.DeleteInterfaceTemplate(
				transactionContext, template.ID,
			); err != nil {
				return err
			}
			change, err := changelog.NewChange(
				principal.ID, dcimdomain.InterfaceTemplateObjectType,
				template.ID, template.Representation, changelog.ActionDelete,
				template.Snapshot, nil, now,
			)
			if err != nil {
				return err
			}
			if err := service.recorder.Record(transactionContext, change); err != nil {
				return err
			}
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if err := service.repository.Delete(transactionContext, loaded); err != nil {
			return err
		}
		return service.recordDeviceTypeChange(
			transactionContext, principal, loaded, changelog.ActionDelete,
			before, nil, now,
		)
	})
	return normalizeOperationError("Could not delete DeviceType.", err)
}

func (service *DeviceTypeService) resolveValues(
	ctx context.Context,
	values deviceTypeCommandValues,
) (dcimdomain.DeviceTypeValues, error) {
	domainValues := dcimdomain.DeviceTypeValues{
		Model: values.model, Slug: values.slug,
		PartNumber: values.partNumber, UHeight: values.uHeight,
		ExcludeFromUtilization: values.excludeFromUtilization,
		IsFullDepth:            values.isFullDepth, Airflow: values.airflow,
		Description: values.description, Comments: values.comments,
	}
	if !values.manufacturerID.IsValid() {
		return domainValues, nil
	}
	manufacturer, err := service.resolveManufacturer(ctx, values.manufacturerID)
	if err != nil {
		return domainValues, err
	}
	domainValues.Manufacturer = manufacturer
	return domainValues, nil
}

func (service *DeviceTypeService) resolvePatch(
	ctx context.Context,
	patch deviceTypeCommandPatch,
) (dcimdomain.DeviceTypePatch, error) {
	domainPatch := dcimdomain.DeviceTypePatch{
		Model: patch.model, Slug: patch.slug, PartNumber: patch.partNumber,
		UHeight:                patch.uHeight,
		ExcludeFromUtilization: patch.excludeFromUtilization,
		IsFullDepth:            patch.isFullDepth, Airflow: patch.airflow,
		Description: patch.description, Comments: patch.comments,
	}
	if patch.manufacturerID != nil {
		if !patch.manufacturerID.IsValid() {
			return domainPatch, nil
		}
		manufacturer, err := service.resolveManufacturer(ctx, *patch.manufacturerID)
		if err != nil {
			return domainPatch, err
		}
		domainPatch.Manufacturer = &manufacturer
	}
	return domainPatch, nil
}

func (service *DeviceTypeService) resolveManufacturer(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.ManufacturerReference, error) {
	manufacturer, err := service.manufacturers.Get(ctx, id)
	if err != nil {
		if shared.HasReason(err, shared.ErrorReasonNotFound) {
			return dcimdomain.ManufacturerReference{}, shared.NewValidationError(
				shared.FieldViolation{
					Field: "manufacturer", Reason: "invalid_choice",
					Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) +
						"\" - object does not exist.",
				},
			)
		}
		return dcimdomain.ManufacturerReference{}, err
	}
	return dcimdomain.NewManufacturerReference(
		manufacturer.ID(), manufacturer.Name(), manufacturer.Slug().String(),
	)
}

func (service *DeviceTypeService) validateHeightTransition(
	ctx context.Context,
	deviceType *dcimdomain.DeviceType,
	original dcimdomain.DeviceHeight,
) error {
	updated := deviceType.UHeight()
	originalHalfUnits := uint32(original.HalfUnits())
	updatedHalfUnits := uint32(updated.HalfUnits())
	if updatedHalfUnits <= originalHalfUnits &&
		(originalHalfUnits == 0 || updatedHalfUnits != 0) {
		return nil
	}
	placements, err := service.repository.ListPositionedDevicesForUpdate(ctx)
	if err != nil {
		return err
	}
	for _, placement := range placements {
		if placement.DeviceTypeID != deviceType.ID() {
			continue
		}
		if updatedHalfUnits == 0 {
			return shared.NewError(
				shared.ErrorReasonProtected,
				"A DeviceType cannot become 0U while an instance is positioned.",
			)
		}
		startingUnit := placement.RackStartingUnit
		if startingUnit == 0 {
			startingUnit = 1
		}
		rackEnd := uint64(startingUnit+placement.RackUHeight) * 2
		if uint64(placement.PositionHalfUnits)+uint64(updatedHalfUnits) > rackEnd {
			return shared.NewError(
				shared.ErrorReasonProtected,
				"The new DeviceType height does not fit a positioned Device.",
			)
		}
		for _, other := range placements {
			if other.ID == placement.ID || other.RackID != placement.RackID {
				continue
			}
			facesConflict := deviceType.IsFullDepth() ||
				other.StoredFullDepth ||
				other.Face == placement.Face
			overlaps := placement.PositionHalfUnits+updatedHalfUnits >
				other.PositionHalfUnits &&
				other.PositionHalfUnits+other.StoredHeightHalfUnits >
					placement.PositionHalfUnits
			if facesConflict && overlaps {
				return shared.NewError(
					shared.ErrorReasonProtected,
					"The new DeviceType height overlaps another positioned Device.",
				)
			}
		}
	}
	return nil
}

func (service *DeviceTypeService) recordDeviceTypeChange(
	ctx context.Context,
	principal identity.Principal,
	deviceType *dcimdomain.DeviceType,
	action changelog.Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.DeviceTypeObjectType, deviceType.ID(),
		deviceType.Display(), action, before, after, now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *DeviceTypeService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	deviceType *dcimdomain.DeviceType,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if deviceType == nil {
		object = nil
	} else {
		object = authz.NewObject(deviceType.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceDeviceType,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}
