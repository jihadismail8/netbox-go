package dcim

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	"netbox-go/internal/application/transaction"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DeviceRoleService struct {
	repository DeviceRoleRepository
	unitOfWork transaction.UnitOfWork
	recorder   changelog.Recorder
	authorizer authz.ResourceAuthorizer
	clock      shared.Clock
}

func NewDeviceRoleService(
	repository DeviceRoleRepository,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*DeviceRoleService, error) {
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
			"DeviceRole service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &DeviceRoleService{
		repository: repository, unitOfWork: unitOfWork, recorder: recorder,
		authorizer: authorizer, clock: clock,
	}, nil
}

func (service *DeviceRoleService) ListDeviceRoles(
	ctx context.Context,
	principal identity.Principal,
	query ListDeviceRolesQuery,
) (DeviceRolePage, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return DeviceRolePage{}, err
	}
	criteria, err := validateListDeviceRolesQuery(query)
	if err != nil {
		return DeviceRolePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx,
			principal,
			authz.View,
			authz.ResourceDeviceRole,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return DeviceRolePage{}, normalizeOperationError("Could not list DeviceRoles.", err)
	}
	visible := make([]*dcimdomain.DeviceRole, 0, len(page.Results))
	for _, role := range page.Results {
		if role == nil {
			return DeviceRolePage{}, shared.NewError(
				shared.ErrorReasonInternal, "DeviceRole repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, role)
		if authorizeErr == nil {
			visible = append(visible, role)
			continue
		}
		if hasCompleteScope {
			return DeviceRolePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"DeviceRole visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return DeviceRolePage{}, authorizeErr
	}
	if hasCompleteScope {
		return DeviceRolePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return DeviceRolePage{Count: count, Results: visible[start:end]}, nil
}

func (service *DeviceRoleService) GetDeviceRole(
	ctx context.Context,
	principal identity.Principal,
	query GetDeviceRoleQuery,
) (*dcimdomain.DeviceRole, error) {
	if err := service.authorize(ctx, principal, authz.View, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	role, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get DeviceRole.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (service *DeviceRoleService) CreateDeviceRole(
	ctx context.Context,
	principal identity.Principal,
	command CreateDeviceRoleCommand,
) (*dcimdomain.DeviceRole, error) {
	if err := service.authorize(ctx, principal, authz.Add, nil); err != nil {
		return nil, err
	}
	var created *dcimdomain.DeviceRole
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		values, valuesErr := command.values()
		if valuesErr != nil {
			return valuesErr
		}
		hierarchy, hierarchyErr := service.repository.ListHierarchyForUpdate(transactionContext)
		if hierarchyErr != nil {
			return hierarchyErr
		}
		if validationErr := validateDeviceRolePlacement(0, values, hierarchy); validationErr != nil {
			return validationErr
		}
		now := service.clock.Now()
		candidate, domainErr := dcimdomain.NewDeviceRole(values, now)
		if domainErr != nil {
			return domainErr
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Add, candidate); authorizeErr != nil {
			return authorizeErr
		}
		if createErr := service.repository.Create(transactionContext, candidate); createErr != nil {
			return createErr
		}
		if recordErr := service.recordChange(
			transactionContext, principal, candidate, changelog.ActionCreate, nil, candidate.Snapshot(), now,
		); recordErr != nil {
			return recordErr
		}
		projected, getErr := service.repository.Get(transactionContext, candidate.ID())
		if getErr != nil {
			return getErr
		}
		created = projected
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create DeviceRole.", err)
	}
	return created, nil
}

func (service *DeviceRoleService) ReplaceDeviceRole(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceDeviceRoleCommand,
) (*dcimdomain.DeviceRole, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	values, err := command.values()
	if err != nil {
		return nil, err
	}
	return service.mutateDeviceRole(ctx, principal, command.ID, func(role *dcimdomain.DeviceRole, now shared.Timestamp) error {
		return role.Replace(values, now)
	})
}

func (service *DeviceRoleService) UpdateDeviceRole(
	ctx context.Context,
	principal identity.Principal,
	command UpdateDeviceRoleCommand,
) (*dcimdomain.DeviceRole, error) {
	if err := service.authorize(ctx, principal, authz.Change, nil); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	patch, err := command.patch()
	if err != nil {
		return nil, err
	}
	return service.mutateDeviceRole(ctx, principal, command.ID, func(role *dcimdomain.DeviceRole, now shared.Timestamp) error {
		return role.ApplyPatch(patch, now)
	})
}

func (service *DeviceRoleService) mutateDeviceRole(
	ctx context.Context,
	principal identity.Principal,
	id shared.ID,
	mutation func(*dcimdomain.DeviceRole, shared.Timestamp) error,
) (*dcimdomain.DeviceRole, error) {
	var updated *dcimdomain.DeviceRole
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		hierarchy, hierarchyErr := service.repository.ListHierarchyForUpdate(transactionContext)
		if hierarchyErr != nil {
			return hierarchyErr
		}
		role := deviceRoleByID(hierarchy, id)
		if role == nil {
			return shared.NotFound("DeviceRole", id)
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Change, role); authorizeErr != nil {
			return authorizeErr
		}
		before := role.Snapshot()
		now := service.clock.Now()
		if mutationErr := mutation(role, now); mutationErr != nil {
			return mutationErr
		}
		if validationErr := validateDeviceRolePlacement(id, role.Values(), hierarchy); validationErr != nil {
			return validationErr
		}
		if updateErr := service.repository.Update(transactionContext, role); updateErr != nil {
			return updateErr
		}
		if recordErr := service.recordChange(
			transactionContext, principal, role, changelog.ActionUpdate, before, role.Snapshot(), now,
		); recordErr != nil {
			return recordErr
		}
		projected, getErr := service.repository.Get(transactionContext, id)
		if getErr != nil {
			return getErr
		}
		updated = projected
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update DeviceRole.", err)
	}
	return updated, nil
}

func (service *DeviceRoleService) DeleteDeviceRole(
	ctx context.Context,
	principal identity.Principal,
	command DeleteDeviceRoleCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, nil); err != nil {
		return err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return err
	}
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		hierarchy, hierarchyErr := service.repository.ListHierarchyForUpdate(transactionContext)
		if hierarchyErr != nil {
			return hierarchyErr
		}
		root := deviceRoleByID(hierarchy, command.ID)
		if root == nil {
			return shared.NotFound("DeviceRole", command.ID)
		}
		if authorizeErr := service.authorize(transactionContext, principal, authz.Delete, root); authorizeErr != nil {
			return authorizeErr
		}
		postorder, subtreeErr := deviceRoleSubtreePostorder(root, hierarchy)
		if subtreeErr != nil {
			return subtreeErr
		}
		ids := make([]shared.ID, 0, len(postorder))
		for _, role := range postorder {
			ids = append(ids, role.ID())
		}
		dependent, dependentErr := service.repository.FindDeviceUsingRoles(transactionContext, ids)
		if dependentErr != nil {
			return dependentErr
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
		for _, role := range postorder {
			before := role.Snapshot()
			now := service.clock.Now()
			if deleteErr := service.repository.Delete(transactionContext, role); deleteErr != nil {
				return deleteErr
			}
			if recordErr := service.recordChange(
				transactionContext, principal, role, changelog.ActionDelete, before, nil, now,
			); recordErr != nil {
				return recordErr
			}
		}
		return nil
	})
	return normalizeOperationError("Could not delete DeviceRole.", err)
}

func (service *DeviceRoleService) recordChange(
	ctx context.Context,
	principal identity.Principal,
	role *dcimdomain.DeviceRole,
	action changelog.Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.DeviceRoleObjectType, role.ID(), role.Display(),
		action, before, after, now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *DeviceRoleService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	role *dcimdomain.DeviceRole,
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
		authz.ResourceDeviceRole,
		object,
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}

func validateDeviceRolePlacement(
	id shared.ID,
	values dcimdomain.DeviceRoleValues,
	hierarchy []*dcimdomain.DeviceRole,
) error {
	byID := make(map[shared.ID]*dcimdomain.DeviceRole, len(hierarchy))
	for _, role := range hierarchy {
		if role == nil || !role.ID().IsValid() {
			return shared.NewError(shared.ErrorReasonInternal, "DeviceRole hierarchy contains an invalid object.")
		}
		byID[role.ID()] = role
	}
	if parentID, hasParent := values.Parent.Get(); hasParent {
		if byID[parentID] == nil {
			return shared.NewValidationError(shared.FieldViolation{
				Field: "parent", Reason: "does_not_exist", Description: "The related object does not exist.",
			})
		}
		visited := make(map[shared.ID]struct{})
		for current := parentID; current.IsValid(); {
			if current == id {
				return shared.NewValidationError(shared.FieldViolation{
					Field: "parent", Reason: "invalid",
					Description: "Cannot assign self or child device role as parent.",
				})
			}
			if _, duplicate := visited[current]; duplicate {
				return shared.NewError(shared.ErrorReasonInternal, "Persisted DeviceRole hierarchy contains a cycle.")
			}
			visited[current] = struct{}{}
			candidate := byID[current]
			if candidate == nil {
				return shared.NewError(shared.ErrorReasonInternal, "Persisted DeviceRole hierarchy contains a missing parent.")
			}
			next, present := candidate.Parent().Get()
			if !present {
				break
			}
			current = next
		}
	}
	for _, role := range hierarchy {
		if role.ID() == id || !sameDeviceRoleParent(values.Parent, role.Parent()) {
			continue
		}
		if role.Name() == strings.TrimSpace(values.Name) {
			return deviceRoleUniquenessError(values.Parent, "name")
		}
		if role.Slug().String() == strings.TrimSpace(values.Slug) {
			return deviceRoleUniquenessError(values.Parent, "slug")
		}
	}
	return nil
}

func deviceRoleUniquenessError(parent dcimdomain.DeviceRoleParent, field string) error {
	var message string
	if parent.IsRoot() {
		message = "A top-level device role with this " + field + " already exists."
	} else {
		label := map[string]string{"name": "Name", "slug": "Slug"}[field]
		message = "Device role with this Parent and " + label + " already exists."
	}
	return shared.ConflictWithViolations(
		message,
		nil,
		shared.FieldViolation{Field: "non_field_errors", Reason: "unique", Description: message},
	)
}

func sameDeviceRoleParent(left, right dcimdomain.DeviceRoleParent) bool {
	leftID, leftPresent := left.Get()
	rightID, rightPresent := right.Get()
	return leftPresent == rightPresent && (!leftPresent || leftID == rightID)
}

func deviceRoleByID(roles []*dcimdomain.DeviceRole, id shared.ID) *dcimdomain.DeviceRole {
	for _, role := range roles {
		if role != nil && role.ID() == id {
			return role
		}
	}
	return nil
}

func deviceRoleSubtreePostorder(
	root *dcimdomain.DeviceRole,
	hierarchy []*dcimdomain.DeviceRole,
) ([]*dcimdomain.DeviceRole, error) {
	children := make(map[shared.ID][]*dcimdomain.DeviceRole)
	for _, role := range hierarchy {
		if role == nil {
			return nil, shared.NewError(shared.ErrorReasonInternal, "DeviceRole hierarchy contains an invalid object.")
		}
		if parentID, present := role.Parent().Get(); present {
			children[parentID] = append(children[parentID], role)
		}
	}
	for parentID := range children {
		sort.Slice(children[parentID], func(left, right int) bool {
			if children[parentID][left].Name() == children[parentID][right].Name() {
				return children[parentID][left].ID() < children[parentID][right].ID()
			}
			return children[parentID][left].Name() < children[parentID][right].Name()
		})
	}
	visited := make(map[shared.ID]bool)
	active := make(map[shared.ID]bool)
	postorder := make([]*dcimdomain.DeviceRole, 0)
	var visit func(*dcimdomain.DeviceRole) error
	visit = func(role *dcimdomain.DeviceRole) error {
		if active[role.ID()] {
			return shared.NewError(shared.ErrorReasonInternal, "Persisted DeviceRole hierarchy contains a cycle.")
		}
		if visited[role.ID()] {
			return nil
		}
		active[role.ID()] = true
		for _, child := range children[role.ID()] {
			if err := visit(child); err != nil {
				return err
			}
		}
		active[role.ID()] = false
		visited[role.ID()] = true
		postorder = append(postorder, role)
		return nil
	}
	if root == nil {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot traverse a nil DeviceRole.")
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	return postorder, nil
}
