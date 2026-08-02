package ipam

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	"netbox-go/internal/domain/identity"
	ipamdomain "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type VRFService struct {
	repository VRFRepository
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewVRFService(
	repository VRFRepository,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*VRFService, error) {
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
			"VRF service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &VRFService{
		repository: repository,
		unitOfWork: unitOfWork,
		recorder:   recorder,
		authorizer: authorizer,
		clock:      clock,
	}, nil
}

func (service *VRFService) ListVRFs(
	ctx context.Context,
	principal identity.Principal,
	query ListVRFsQuery,
) (VRFPage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return VRFPage{}, err
	}
	criteria, err := validateListVRFsQuery(query)
	if err != nil {
		return VRFPage{}, err
	}

	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceVRF,
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
		return VRFPage{}, normalizeOperationError("Could not list VRFs.", err)
	}
	visible := make([]*ipamdomain.VRF, 0, len(page.Results))
	for _, vrf := range page.Results {
		if vrf == nil {
			return VRFPage{}, shared.NewError(
				shared.ErrorReasonInternal,
				"VRF repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, vrf)
		if authorizeErr == nil {
			visible = append(visible, vrf)
			continue
		}
		if hasCompleteScope {
			return VRFPage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"VRF visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return VRFPage{}, authorizeErr
	}

	if hasCompleteScope {
		return VRFPage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return VRFPage{Count: count, Results: visible[start:end]}, nil
}

func (service *VRFService) GetVRF(
	ctx context.Context,
	principal identity.Principal,
	query GetVRFQuery,
) (*ipamdomain.VRF, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	vrf, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get VRF.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, vrf); err != nil {
		return nil, err
	}
	return vrf, nil
}

func (service *VRFService) CreateVRF(
	ctx context.Context,
	principal identity.Principal,
	command CreateVRFCommand,
) (*ipamdomain.VRF, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}

	var vrf *ipamdomain.VRF
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		values, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		now := service.clock.Now()
		candidate, domainErr := ipamdomain.NewVRF(values, now)
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
			ipamdomain.VRFObjectType,
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
		vrf = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create VRF.", err)
	}
	return vrf, nil
}

func (service *VRFService) ReplaceVRF(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceVRFCommand,
) (*ipamdomain.VRF, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}

	var vrf *ipamdomain.VRF
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
		vrf = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace VRF.", err)
	}
	return vrf, nil
}

func (service *VRFService) UpdateVRF(
	ctx context.Context,
	principal identity.Principal,
	command UpdateVRFCommand,
) (*ipamdomain.VRF, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}

	var vrf *ipamdomain.VRF
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
		vrf = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update VRF.", err)
	}
	return vrf, nil
}

func (service *VRFService) DeleteVRF(
	ctx context.Context,
	principal identity.Principal,
	command DeleteVRFCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}

	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		vrf, getErr := service.repository.GetForUpdate(transactionContext, command.ID)
		if getErr != nil {
			return getErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Delete, vrf); authorizeErr != nil {
			return authorizeErr
		}
		before := vrf.Snapshot()
		now := service.clock.Now()
		if deleteErr := service.repository.Delete(transactionContext, vrf); deleteErr != nil {
			return deleteErr
		}
		change, changeErr := changelog.NewChange(
			principal.ID,
			ipamdomain.VRFObjectType,
			vrf.ID(),
			vrf.Display(),
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
	return normalizeOperationError("Could not delete VRF.", err)
}

func (service *VRFService) recordUpdate(
	ctx context.Context,
	principal identity.Principal,
	vrf *ipamdomain.VRF,
	before ipamdomain.VRFSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID,
		ipamdomain.VRFObjectType,
		vrf.ID(),
		vrf.Display(),
		changelog.ActionUpdate,
		before,
		vrf.Snapshot(),
		now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *VRFService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	vrf *ipamdomain.VRF,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}

	var object *authz.Object
	if vrf != nil {
		object = authz.NewObject(vrf.ID().Int64())
	}
	if err := service.authorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		authz.ResourceVRF,
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
