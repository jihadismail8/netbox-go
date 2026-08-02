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

type InterfaceTemplateService struct {
	repository  InterfaceTemplateRepository
	deviceTypes InterfaceTemplateDeviceTypeReader
	unitOfWork  transaction.UnitOfWork
	recorder    changelog.Recorder
	authorizer  authz.ResourceAuthorizer
	clock       shared.Clock
}

func NewInterfaceTemplateService(
	repository InterfaceTemplateRepository,
	deviceTypes InterfaceTemplateDeviceTypeReader,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*InterfaceTemplateService, error) {
	missing := make([]string, 0, 6)
	if nilInterface(repository) {
		missing = append(missing, "repository")
	}
	if nilInterface(deviceTypes) {
		missing = append(missing, "device type reader")
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
			"InterfaceTemplate service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &InterfaceTemplateService{
		repository: repository, deviceTypes: deviceTypes, unitOfWork: unitOfWork,
		recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *InterfaceTemplateService) ListInterfaceTemplates(
	ctx context.Context,
	principal identity.Principal,
	query ListInterfaceTemplatesQuery,
) (InterfaceTemplatePage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return InterfaceTemplatePage{}, err
	}
	criteria, err := validateListInterfaceTemplatesQuery(query)
	if err != nil {
		return InterfaceTemplatePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceInterfaceTemplate,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return InterfaceTemplatePage{}, normalizeOperationError(
			"Could not list InterfaceTemplates.", err,
		)
	}
	visible := make([]*dcimdomain.InterfaceTemplate, 0, len(page.Results))
	for _, template := range page.Results {
		if template == nil {
			return InterfaceTemplatePage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"InterfaceTemplate repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, template)
		if authorizeErr == nil {
			visible = append(visible, template)
			continue
		}
		if hasCompleteScope {
			return InterfaceTemplatePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"InterfaceTemplate visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return InterfaceTemplatePage{}, authorizeErr
	}
	if hasCompleteScope {
		return InterfaceTemplatePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return InterfaceTemplatePage{Count: count, Results: visible[start:end]}, nil
}

func (service *InterfaceTemplateService) GetInterfaceTemplate(
	ctx context.Context,
	principal identity.Principal,
	query GetInterfaceTemplateQuery,
) (*dcimdomain.InterfaceTemplate, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	template, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get InterfaceTemplate.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, template); err != nil {
		return nil, err
	}
	return template, nil
}

func (service *InterfaceTemplateService) CreateInterfaceTemplate(
	ctx context.Context,
	principal identity.Principal,
	command CreateInterfaceTemplateCommand,
) (*dcimdomain.InterfaceTemplate, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var template *dcimdomain.InterfaceTemplate
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		commandValues, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		values, valuesErr := service.resolveValues(transactionContext, commandValues)
		if valuesErr != nil {
			return valuesErr
		}
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewInterfaceTemplate(values, now)
		if domainErr != nil {
			return domainErr
		}
		if authorizeErr := service.authorize(
			transactionContext, principal, authz.Add, candidate,
		); authorizeErr != nil {
			return authorizeErr
		}
		if createErr := service.repository.Create(transactionContext, candidate); createErr != nil {
			return createErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.InterfaceTemplateObjectType,
			candidate.ID(), candidate.Display(), changelog.ActionCreate,
			nil, candidate.Snapshot(), now,
		)
		if changeErr != nil {
			return changeErr
		}
		if recordErr := service.recorder.Record(transactionContext, change); recordErr != nil {
			return recordErr
		}
		template = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create InterfaceTemplate.", err)
	}
	return template, nil
}

func (service *InterfaceTemplateService) ReplaceInterfaceTemplate(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceInterfaceTemplateCommand,
) (*dcimdomain.InterfaceTemplate, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var template *dcimdomain.InterfaceTemplate
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(
			transactionContext, principal, authz.Change, loaded,
		); authorizeErr != nil {
			return authorizeErr
		}
		commandValues, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		values, valuesErr := service.resolveValues(transactionContext, commandValues)
		if valuesErr != nil {
			return valuesErr
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if replaceErr := loaded.Replace(values, now); replaceErr != nil {
			return replaceErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if recordErr := service.recordUpdate(
			transactionContext, principal, loaded, before, now,
		); recordErr != nil {
			return recordErr
		}
		template = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace InterfaceTemplate.", err)
	}
	return template, nil
}

func (service *InterfaceTemplateService) UpdateInterfaceTemplate(
	ctx context.Context,
	principal identity.Principal,
	command UpdateInterfaceTemplateCommand,
) (*dcimdomain.InterfaceTemplate, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var template *dcimdomain.InterfaceTemplate
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(
			transactionContext, principal, authz.Change, loaded,
		); authorizeErr != nil {
			return authorizeErr
		}
		commandPatch, patchErr := command.patch()
		if patchErr != nil {
			return patchErr
		}
		patch, patchErr := service.resolvePatch(transactionContext, commandPatch)
		if patchErr != nil {
			return patchErr
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if domainErr := loaded.ApplyPatch(patch, now); domainErr != nil {
			return domainErr
		}
		if updateErr := service.repository.Update(transactionContext, loaded); updateErr != nil {
			return updateErr
		}
		if recordErr := service.recordUpdate(
			transactionContext, principal, loaded, before, now,
		); recordErr != nil {
			return recordErr
		}
		template = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update InterfaceTemplate.", err)
	}
	return template, nil
}

func (service *InterfaceTemplateService) DeleteInterfaceTemplate(
	ctx context.Context,
	principal identity.Principal,
	command DeleteInterfaceTemplateCommand,
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
		if authorizeErr := service.authorize(
			transactionContext, principal, authz.Delete, loaded,
		); authorizeErr != nil {
			return authorizeErr
		}
		before := loaded.Snapshot()
		now := service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, loaded); deleteErr != nil {
			return deleteErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID, dcimdomain.InterfaceTemplateObjectType,
			loaded.ID(), loaded.Display(), changelog.ActionDelete,
			before, nil, now,
		)
		if changeErr != nil {
			return changeErr
		}
		return service.recorder.Record(transactionContext, change)
	})
	return normalizeOperationError("Could not delete InterfaceTemplate.", err)
}

func (service *InterfaceTemplateService) resolveValues(
	ctx context.Context,
	values interfaceTemplateCommandValues,
) (dcimdomain.InterfaceTemplateValues, error) {
	reference, err := service.resolveDeviceType(ctx, values.deviceTypeID)
	if err != nil {
		return dcimdomain.InterfaceTemplateValues{}, err
	}
	return dcimdomain.InterfaceTemplateValues{
		DeviceType: reference, Name: values.name, Label: values.label,
		Type: values.interfaceType, Enabled: values.enabled, MgmtOnly: values.mgmtOnly,
		Description: values.description,
	}, nil
}

func (service *InterfaceTemplateService) resolvePatch(
	ctx context.Context,
	patch interfaceTemplateCommandPatch,
) (dcimdomain.InterfaceTemplatePatch, error) {
	domainPatch := dcimdomain.InterfaceTemplatePatch{
		Name: patch.name, Label: patch.label, Type: patch.interfaceType,
		Enabled: patch.enabled, MgmtOnly: patch.mgmtOnly, Description: patch.description,
	}
	if patch.deviceTypeID != nil {
		reference, err := service.resolveDeviceType(ctx, *patch.deviceTypeID)
		if err != nil {
			return dcimdomain.InterfaceTemplatePatch{}, err
		}
		domainPatch.DeviceType = &reference
	}
	return domainPatch, nil
}

func (service *InterfaceTemplateService) resolveDeviceType(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.DeviceTypeReference, error) {
	deviceType, err := service.deviceTypes.Get(ctx, id)
	if err != nil {
		if shared.HasReason(err, shared.ErrorReasonNotFound) {
			return dcimdomain.DeviceTypeReference{}, shared.NewValidationError(
				shared.FieldViolation{
					Field: "device_type", Reason: "invalid_choice",
					Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) +
						"\" - object does not exist.",
				},
			)
		}
		return dcimdomain.DeviceTypeReference{}, err
	}
	return dcimdomain.NewDeviceTypeReference(
		deviceType.ID(), deviceType.Model(), deviceType.Slug().String(),
	)
}

func (service *InterfaceTemplateService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	template *dcimdomain.InterfaceTemplate,
	before dcimdomain.InterfaceTemplateSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.InterfaceTemplateObjectType,
		template.ID(), template.Display(), changelog.ActionUpdate,
		before, template.Snapshot(), now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *InterfaceTemplateService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	template *dcimdomain.InterfaceTemplate,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	var object *authz.Object
	if template == nil {
		object = nil
	} else {
		object = authz.NewObject(template.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceInterfaceTemplate,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}
