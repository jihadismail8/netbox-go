package authz

import (
	"context"
	"strings"

	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

// ResourceType is the static authorization identity of one managed-object
// type. It replaces table-shaped and map-shaped resource descriptors at typed
// application boundaries.
type ResourceType struct {
	appLabel string
	model    string
}

func newResourceType(appLabel, model string) ResourceType {
	return ResourceType{appLabel: appLabel, model: model}
}

func (resource ResourceType) AppLabel() string { return resource.appLabel }
func (resource ResourceType) Model() string    { return resource.model }
func (resource ResourceType) Valid() bool {
	return resource.appLabel != "" && resource.model != "" &&
		resource.appLabel == strings.ToLower(resource.appLabel) &&
		resource.model == strings.ToLower(resource.model) &&
		!strings.ContainsAny(resource.appLabel+resource.model, "._- ")
}

func (resource ResourceType) Permission(action Action) string {
	if !resource.Valid() {
		return ""
	}
	return resource.appLabel + "." + string(action) + "_" + resource.model
}

var (
	ResourceSite              = newResourceType("dcim", "site")
	ResourceManufacturer      = newResourceType("dcim", "manufacturer")
	ResourceRackRole          = newResourceType("dcim", "rackrole")
	ResourceRackType          = newResourceType("dcim", "racktype")
	ResourceRack              = newResourceType("dcim", "rack")
	ResourceDeviceRole        = newResourceType("dcim", "devicerole")
	ResourceDeviceType        = newResourceType("dcim", "devicetype")
	ResourceInterfaceTemplate = newResourceType("dcim", "interfacetemplate")
	ResourceDevice            = newResourceType("dcim", "device")
	ResourceInterface         = newResourceType("dcim", "interface")
	ResourceVRF               = newResourceType("ipam", "vrf")
	ResourcePrefix            = newResourceType("ipam", "prefix")
	ResourceIPAddress         = newResourceType("ipam", "ipaddress")
)

// Object is the only object-specific fact the persisted permission model
// consumes. Domain aggregates never need to become generic authorization maps.
type Object struct {
	ID int64
}

func NewObject(id int64) *Object {
	if id <= 0 {
		return nil
	}
	return &Object{ID: id}
}

// ResourceAuthorizer is the typed authorization port used by domain-specific
// application services.
type ResourceAuthorizer interface {
	AuthorizeResource(
		context.Context,
		identity.Principal,
		Action,
		ResourceType,
		*Object,
	) error
}

// ResourceListScopeAuthorizer exposes a complete object-visibility set for a
// typed resource. Implementations which cannot express all visibility rules
// as IDs must not implement this optional optimization.
type ResourceListScopeAuthorizer interface {
	ResourceListScope(
		context.Context,
		identity.Principal,
		Action,
		ResourceType,
	) ListScope
}

func (PermissionAuthorizer) AuthorizeResource(
	_ context.Context,
	principal identity.Principal,
	action Action,
	resource ResourceType,
	object *Object,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	permission := resource.Permission(action)
	if permission == "" || !principal.Has(permission) {
		return shared.NewError(
			shared.ErrorReasonForbidden,
			"You do not have permission to perform this action.",
		)
	}
	if object != nil {
		if visible, constrained := principal.ObjectVisibility[permission]; constrained {
			if _, allowed := visible[object.ID]; !allowed {
				return shared.NewError(
					shared.ErrorReasonForbidden,
					"You do not have permission to perform this action.",
				)
			}
		}
	}
	return nil
}

func (PermissionAuthorizer) ResourceListScope(
	_ context.Context,
	principal identity.Principal,
	action Action,
	resource ResourceType,
) ListScope {
	if !resource.Valid() || principal.IsSuperuser {
		return ListScope{}
	}
	visible, constrained := principal.ObjectVisibility[resource.Permission(action)]
	if !constrained {
		return ListScope{}
	}
	ids := make([]int64, 0, len(visible))
	for id := range visible {
		ids = append(ids, id)
	}
	return ListScope{ObjectIDs: ids, Constrained: true}
}

func (AllowAll) AuthorizeResource(
	_ context.Context,
	principal identity.Principal,
	_ Action,
	_ ResourceType,
	_ *Object,
) error {
	if !principal.Authenticated() {
		return shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	return nil
}

func (AllowAll) ResourceListScope(
	_ context.Context,
	_ identity.Principal,
	_ Action,
	_ ResourceType,
) ListScope {
	return ListScope{}
}
