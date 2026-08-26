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

type RackService struct {
	repository RackRepository
	sites      RackSiteReader
	rackTypes  RackTypeReader
	rackRoles  RackRoleReader
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewRackService(
	repository RackRepository,
	sites RackSiteReader,
	rackTypes RackTypeReader,
	rackRoles RackRoleReader,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*RackService, error) {
	missing := make([]string, 0, 8)
	if nilInterface(repository) {
		missing = append(missing, "repository")
	}
	if nilInterface(sites) {
		missing = append(missing, "site reader")
	}
	if nilInterface(rackTypes) {
		missing = append(missing, "rack type reader")
	}
	if nilInterface(rackRoles) {
		missing = append(missing, "rack role reader")
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
			"Rack service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &RackService{
		repository: repository, sites: sites, rackTypes: rackTypes, rackRoles: rackRoles,
		unitOfWork: unitOfWork, recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *RackService) ListRacks(
	ctx context.Context,
	principal identity.Principal,
	query ListRacksQuery,
) (RackPage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return RackPage{}, err
	}
	criteria, err := validateListRacksQuery(query)
	if err != nil {
		return RackPage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(ctx, principal, authz.View, authz.ResourceRack)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return RackPage{}, normalizeOperationError("Could not list Racks.", err)
	}
	visible := make([]*dcimdomain.Rack, 0, len(page.Results))
	for _, rack := range page.Results {
		if rack == nil {
			return RackPage{}, shared.NewError(
				shared.ErrorReasonInternal, "Rack repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, rack)
		if authorizeErr == nil {
			visible = append(visible, rack)
			continue
		}
		if hasCompleteScope {
			return RackPage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"Rack visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return RackPage{}, authorizeErr
	}
	if hasCompleteScope {
		return RackPage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return RackPage{Count: count, Results: visible[start:end]}, nil
}

func (service *RackService) GetRack(
	ctx context.Context,
	principal identity.Principal,
	query GetRackQuery,
) (*dcimdomain.Rack, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	rack, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get Rack.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, rack); err != nil {
		return nil, err
	}
	return rack, nil
}

func (service *RackService) CreateRack(
	ctx context.Context,
	principal identity.Principal,
	command CreateRackCommand,
) (*dcimdomain.Rack, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var rack *dcimdomain.Rack
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		commandValues, commandErr := command.values()
		values, relationshipErr := service.resolveValues(transactionContext, commandValues)
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewRack(values, now)
		if validationErr := mergeRackMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		if ownershipErr := candidate.ApplyRackTypeOwnership(now); ownershipErr != nil {
			return ownershipErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Add, candidate); authorizeErr != nil {
			return authorizeErr
		}
		if createErr := service.repository.Create(transactionContext, candidate); createErr != nil {
			return createErr
		}
		if recordErr := service.record(
			transactionContext, principal, candidate, changelog.ActionCreate, nil, candidate.Snapshot(), now,
		); recordErr != nil {
			return recordErr
		}
		rack = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create Rack.", err)
	}
	return rack, nil
}

func (service *RackService) ReplaceRack(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceRackCommand,
) (*dcimdomain.Rack, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var rack *dcimdomain.Rack
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Change, loaded); authorizeErr != nil {
			return authorizeErr
		}
		commandPatch, commandErr := command.patch()
		patch, relationshipErr := service.resolvePatch(transactionContext, commandPatch)
		domainErr := loaded.ValidatePatch(patch)
		if validationErr := mergeRackMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		before := loaded.Snapshot()
		previousSite := loaded.Site().ID()
		now := service.clock.Now()
		if replaceErr := loaded.ApplyPatch(patch, now); replaceErr != nil {
			return replaceErr
		}
		if ownershipErr := loaded.ApplyRackTypeOwnership(now); ownershipErr != nil {
			return ownershipErr
		}
		if placementErr := service.validateMountedDevices(transactionContext, loaded); placementErr != nil {
			return placementErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if recordErr := service.record(
			transactionContext, principal, loaded, changelog.ActionUpdate, before, loaded.Snapshot(), now,
		); recordErr != nil {
			return recordErr
		}
		if previousSite != loaded.Site().ID() {
			if propagationErr := service.propagateSite(transactionContext, principal, loaded, now); propagationErr != nil {
				return propagationErr
			}
		}
		rack = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace Rack.", err)
	}
	return rack, nil
}

func (service *RackService) UpdateRack(
	ctx context.Context,
	principal identity.Principal,
	command UpdateRackCommand,
) (*dcimdomain.Rack, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var rack *dcimdomain.Rack
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Change, loaded); authorizeErr != nil {
			return authorizeErr
		}
		commandPatch, commandErr := command.patch()
		patch, relationshipErr := service.resolvePatch(transactionContext, commandPatch)
		domainErr := loaded.ValidatePatch(patch)
		if validationErr := mergeRackMutationErrors(
			commandErr, relationshipErr, domainErr,
		); validationErr != nil {
			return validationErr
		}
		before := loaded.Snapshot()
		previousSite := loaded.Site().ID()
		now := service.clock.Now()
		if domainErr := loaded.ApplyPatch(patch, now); domainErr != nil {
			return domainErr
		}
		if ownershipErr := loaded.ApplyRackTypeOwnership(now); ownershipErr != nil {
			return ownershipErr
		}
		if placementErr := service.validateMountedDevices(transactionContext, loaded); placementErr != nil {
			return placementErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if recordErr := service.record(
			transactionContext, principal, loaded, changelog.ActionUpdate, before, loaded.Snapshot(), now,
		); recordErr != nil {
			return recordErr
		}
		if previousSite != loaded.Site().ID() {
			if propagationErr := service.propagateSite(transactionContext, principal, loaded, now); propagationErr != nil {
				return propagationErr
			}
		}
		rack = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update Rack.", err)
	}
	return rack, nil
}

func (service *RackService) DeleteRack(
	ctx context.Context,
	principal identity.Principal,
	command DeleteRackCommand,
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
		before, now := loaded.Snapshot(), service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, loaded); deleteErr != nil {
			return deleteErr
		}
		return service.record(
			transactionContext, principal, loaded, changelog.ActionDelete, before, nil, now,
		)
	})
	return normalizeOperationError("Could not delete Rack.", err)
}

func (service *RackService) resolveValues(
	ctx context.Context,
	values rackCommandValues,
) (dcimdomain.RackValues, error) {
	domainValues := dcimdomain.RackValues{
		Name: values.name, FacilityID: values.facilityID,
		RackType: dcimdomain.NullRackValue[dcimdomain.RackTypeReference](),
		Status:   values.status,
		Role:     dcimdomain.NullRackValue[dcimdomain.RackRoleReference](),
		Serial:   values.serial, AssetTag: values.assetTag, FormFactor: values.formFactor,
		Width: values.width, UHeight: values.uHeight, StartingUnit: values.startingUnit,
		DescUnits: values.descUnits, Airflow: values.airflow,
		Description: values.description, Comments: values.comments,
	}
	var relationshipErrs []error
	if values.siteID.IsValid() {
		site, err := service.resolveSite(ctx, values.siteID)
		if err != nil {
			relationshipErrs = append(relationshipErrs, err)
		} else {
			domainValues.Site = site
		}
	}
	if id, present := values.rackTypeID.Get(); !present || id.IsValid() {
		rackType, err := service.resolveOptionalRackType(ctx, values.rackTypeID)
		if err != nil {
			relationshipErrs = append(relationshipErrs, err)
		} else {
			domainValues.RackType = rackType
		}
	}
	if id, present := values.roleID.Get(); !present || id.IsValid() {
		role, err := service.resolveOptionalRackRole(ctx, values.roleID)
		if err != nil {
			relationshipErrs = append(relationshipErrs, err)
		} else {
			domainValues.Role = role
		}
	}
	return domainValues, mergeRackMutationErrors(relationshipErrs...)
}

func (service *RackService) resolvePatch(
	ctx context.Context,
	patch rackCommandPatch,
) (dcimdomain.RackPatch, error) {
	domainPatch := dcimdomain.RackPatch{
		Name: patch.name, FacilityID: patch.facilityID, Status: patch.status,
		Serial: patch.serial, AssetTag: patch.assetTag, FormFactor: patch.formFactor,
		Width: patch.width, UHeight: patch.uHeight, StartingUnit: patch.startingUnit,
		DescUnits: patch.descUnits, Airflow: patch.airflow,
		Description: patch.description, Comments: patch.comments,
	}
	var relationshipErrs []error
	if patch.siteID != nil && patch.siteID.IsValid() {
		site, err := service.resolveSite(ctx, *patch.siteID)
		if err != nil {
			relationshipErrs = append(relationshipErrs, err)
		} else {
			domainPatch.Site = &site
		}
	}
	if patch.rackTypeID != nil {
		id, present := patch.rackTypeID.Get()
		if !present || id.IsValid() {
			rackType, err := service.resolveOptionalRackType(ctx, *patch.rackTypeID)
			if err != nil {
				relationshipErrs = append(relationshipErrs, err)
			} else {
				domainPatch.RackType = &rackType
			}
		}
	}
	if patch.roleID != nil {
		id, present := patch.roleID.Get()
		if !present || id.IsValid() {
			role, err := service.resolveOptionalRackRole(ctx, *patch.roleID)
			if err != nil {
				relationshipErrs = append(relationshipErrs, err)
			} else {
				domainPatch.Role = &role
			}
		}
	}
	return domainPatch, mergeRackMutationErrors(relationshipErrs...)
}

func (service *RackService) resolveSite(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.SiteReference, error) {
	site, err := service.sites.Get(ctx, id)
	if err != nil {
		return dcimdomain.SiteReference{}, relatedRackError("site", id, err)
	}
	return dcimdomain.NewSiteReference(site.ID(), site.Name(), site.Slug().String())
}

func (service *RackService) resolveOptionalRackType(
	ctx context.Context,
	value dcimdomain.RackNullable[shared.ID],
) (dcimdomain.RackNullable[dcimdomain.RackTypeReference], error) {
	id, present := value.Get()
	if !present {
		return dcimdomain.NullRackValue[dcimdomain.RackTypeReference](), nil
	}
	rackType, err := service.rackTypes.Get(ctx, id)
	if err != nil {
		return dcimdomain.RackNullable[dcimdomain.RackTypeReference]{}, relatedRackError("rack_type", id, err)
	}
	reference, err := dcimdomain.NewRackTypeReference(
		rackType.ID(), rackType.Model(), rackType.Slug().String(), rackType.PhysicalAttributes(),
	)
	if err != nil {
		return dcimdomain.RackNullable[dcimdomain.RackTypeReference]{}, err
	}
	return dcimdomain.NonNullRackValue(reference), nil
}

func (service *RackService) resolveOptionalRackRole(
	ctx context.Context,
	value dcimdomain.RackNullable[shared.ID],
) (dcimdomain.RackNullable[dcimdomain.RackRoleReference], error) {
	id, present := value.Get()
	if !present {
		return dcimdomain.NullRackValue[dcimdomain.RackRoleReference](), nil
	}
	role, err := service.rackRoles.Get(ctx, id)
	if err != nil {
		return dcimdomain.RackNullable[dcimdomain.RackRoleReference]{}, relatedRackError("role", id, err)
	}
	reference, err := dcimdomain.NewRackRoleReference(role.ID(), role.Name(), role.Slug().String())
	if err != nil {
		return dcimdomain.RackNullable[dcimdomain.RackRoleReference]{}, err
	}
	return dcimdomain.NonNullRackValue(reference), nil
}

func relatedRackError(field string, id shared.ID, err error) error {
	if shared.HasReason(err, shared.ErrorReasonNotFound) {
		return shared.NewValidationError(shared.FieldViolation{
			Field: field, Reason: "invalid_choice",
			Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) + "\" - object does not exist.",
		})
	}
	return err
}

func (service *RackService) validateMountedDevices(
	ctx context.Context,
	rack *dcimdomain.Rack,
) error {
	placements, err := service.repository.MountedDevices(ctx, rack.ID())
	if err != nil {
		return err
	}
	if len(placements) == 0 {
		return nil
	}
	minimumPosition := placements[0].PositionHalfUnits
	maximumEnd := placements[0].PositionHalfUnits + int32(placements[0].HeightHalfUnits)
	for _, placement := range placements[1:] {
		if placement.PositionHalfUnits < minimumPosition {
			minimumPosition = placement.PositionHalfUnits
		}
		end := placement.PositionHalfUnits + int32(placement.HeightHalfUnits)
		if end > maximumEnd {
			maximumEnd = end
		}
	}
	start := int32(rack.StartingUnit() * 2)
	end := start + int32(rack.UHeight()*2)
	field := "u_height"
	if _, assigned := rack.RackType().Get(); assigned {
		field = "rack_type"
	}
	if maximumEnd > end {
		minimumHeight := maximumEnd - start
		return shared.NewValidationError(shared.FieldViolation{
			Field: field, Reason: "invalid",
			Description: "Rack must be at least " + formatHalfUnits(minimumHeight) +
				"U tall to house currently installed devices.",
		})
	}
	if minimumPosition < start {
		if field != "rack_type" {
			field = "starting_unit"
		}
		return shared.NewValidationError(shared.FieldViolation{
			Field: field, Reason: "invalid",
			Description: "Rack unit numbering must begin at " + formatHalfUnits(minimumPosition) +
				" or less to house currently installed devices.",
		})
	}
	return nil
}

func formatHalfUnits(halfUnits int32) string {
	if halfUnits%2 == 0 {
		return strconv.FormatInt(int64(halfUnits/2), 10)
	}
	return fmt.Sprintf("%d.5", halfUnits/2)
}

func (service *RackService) propagateSite(
	ctx context.Context,
	principal identity.Principal,
	rack *dcimdomain.Rack,
	now shared.Timestamp,
) error {
	changes, err := service.repository.PropagateSiteToDevices(
		ctx, rack.ID(), rack.Site().ID(), now,
	)
	if err != nil {
		return err
	}
	for _, propagated := range changes {
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.DeviceObjectType, propagated.ID, propagated.Representation,
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

func (service *RackService) record(
	ctx context.Context,
	principal identity.Principal,
	rack *dcimdomain.Rack,
	action changelog.Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.RackObjectType, rack.ID(), rack.Display(),
		action, before, after, now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *RackService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	rack *dcimdomain.Rack,
) error {
	var object *authz.Object
	if rack != nil {
		object = authz.NewObject(rack.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx, principal, action, authz.ResourceRack, object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}
