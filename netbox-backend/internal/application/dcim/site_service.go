package dcim

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type SiteService struct {
	repository SiteRepository
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewSiteService(
	repository SiteRepository,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*SiteService, error) {
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
			"Site service requires: "+strings.Join(missing, ", ")+".",
		)
	}

	return &SiteService{
		repository: repository,
		unitOfWork: unitOfWork,
		recorder:   recorder,
		authorizer: authorizer,
		clock:      clock,
	}, nil
}

func (service *SiteService) ListSites(
	ctx context.Context,
	principal identity.Principal,
	query ListSitesQuery,
) (SitePage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return SitePage{}, err
	}

	criteria, err := validateListSitesQuery(query)
	if err != nil {
		return SitePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceSite,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = make([]shared.ID, len(scope.ObjectIDs))
		for index, id := range scope.ObjectIDs {
			criteria.VisibleObjectIDs[index] = shared.ID(id)
		}
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return SitePage{}, normalizeOperationError("Could not list Sites.", err)
	}

	visible := make([]*dcimdomain.Site, 0, len(page.Results))
	for _, site := range page.Results {
		if site == nil {
			return SitePage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"Site repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, site)
		if authorizeErr == nil {
			visible = append(visible, site)
			continue
		}
		if hasCompleteScope {
			return SitePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"Site visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return SitePage{}, authorizeErr
	}

	if hasCompleteScope {
		return SitePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return SitePage{Count: count, Results: visible[start:end]}, nil
}

func (service *SiteService) GetSite(
	ctx context.Context,
	principal identity.Principal,
	query GetSiteQuery,
) (*dcimdomain.Site, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	site, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get Site.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, site); err != nil {
		return nil, err
	}
	return site, nil
}

func (service *SiteService) CreateSite(
	ctx context.Context,
	principal identity.Principal,
	command CreateSiteCommand,
) (*dcimdomain.Site, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}

	var site *dcimdomain.Site
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		values, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewSite(values, now)
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
			principal.ID,
			dcimdomain.SiteObjectType,
			candidate.ID(),
			candidate.Display(),
			changelog.ActionCreate,
			nil,
			candidate.Snapshot(),
			now,
		)
		if changeErr != nil {
			return changeErr
		}
		if recordErr := service.recorder.Record(transactionContext, change); recordErr != nil {
			return recordErr
		}
		site = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create Site.", err)
	}
	return site, nil
}

func (service *SiteService) ReplaceSite(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceSiteCommand,
) (*dcimdomain.Site, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}

	var site *dcimdomain.Site
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
		site = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace Site.", err)
	}
	return site, nil
}

func (service *SiteService) UpdateSite(
	ctx context.Context,
	principal identity.Principal,
	command UpdateSiteCommand,
) (*dcimdomain.Site, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}

	var site *dcimdomain.Site
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
		patch, patchBuildErr := command.patch()
		if patchBuildErr != nil {
			return patchBuildErr
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
		site = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update Site.", err)
	}
	return site, nil
}

func (service *SiteService) DeleteSite(
	ctx context.Context,
	principal identity.Principal,
	command DeleteSiteCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}

	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		site, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Delete, site); authorizeErr != nil {
			return authorizeErr
		}

		before := site.Snapshot()
		now := service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, site); deleteErr != nil {
			return deleteErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID,
			dcimdomain.SiteObjectType,
			site.ID(),
			site.Display(),
			changelog.ActionDelete,
			before,
			nil,
			now,
		)
		if changeErr != nil {
			return changeErr
		}
		return service.recorder.Record(transactionContext, change)
	})
	return normalizeOperationError("Could not delete Site.", err)
}

func (service *SiteService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	site *dcimdomain.Site,
	before dcimdomain.SiteSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID,
		dcimdomain.SiteObjectType,
		site.ID(),
		site.Display(),
		changelog.ActionUpdate,
		before,
		site.Snapshot(),
		now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *SiteService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	site *dcimdomain.Site,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}

	var object *authz.Object
	if site != nil {
		object = authz.NewObject(site.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceSite,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}

func normalizeAuthorizationError(err error) error {
	if err == nil {
		return nil
	}
	var sharedError *shared.Error
	if errors.As(err, &sharedError) {
		return err
	}
	return shared.WrapError(shared.ErrorReasonInternal, "Authorization failed.", err)
}

func normalizeOperationError(message string, err error) error {
	if err == nil {
		return nil
	}
	var sharedError *shared.Error
	if errors.As(err, &sharedError) {
		return err
	}
	return shared.WrapError(shared.ErrorReasonInternal, message, err)
}

func validatePersistedID(id shared.ID) error {
	if id.IsValid() {
		return nil
	}
	return shared.NewValidationError(shared.FieldViolation{
		Field:       "id",
		Reason:      "invalid",
		Description: "ID must be greater than zero.",
	})
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
