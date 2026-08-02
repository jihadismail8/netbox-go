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

type DeviceService struct {
	repository       DeviceRepository
	deviceTypes      DeviceTypeReader
	roles            DeviceRoleReader
	sites            DeviceSiteReader
	racks            DeviceRackReader
	templates        DeviceInterfaceTemplateReader
	interfaces       DeviceInterfaceCreator
	interfaceCascade DeviceInterfaceCascade
	unitOfWork       transaction.UnitOfWork
	recorder         changelog.Recorder
	authorizer       authz.ResourceAuthorizer
	clock            shared.Clock
}

func NewDeviceService(
	repository DeviceRepository,
	deviceTypes DeviceTypeReader,
	roles DeviceRoleReader,
	sites DeviceSiteReader,
	racks DeviceRackReader,
	templates DeviceInterfaceTemplateReader,
	interfaces DeviceInterfaceCreator,
	interfaceCascade DeviceInterfaceCascade,
	unitOfWork transaction.UnitOfWork,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
	clock shared.Clock,
) (*DeviceService, error) {
	missing := make([]string, 0, 12)
	for name, dependency := range map[string]any{
		"repository": repository, "device type reader": deviceTypes,
		"device role reader": roles, "site reader": sites, "rack reader": racks,
		"interface template reader": templates, "interface creator": interfaces,
		"interface cascade": interfaceCascade, "unit of work": unitOfWork,
		"change recorder": recorder, "authorizer": authorizer, "clock": clock,
	} {
		if nilInterface(dependency) {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Device service requires: "+strings.Join(missing, ", ")+".",
		)
	}
	return &DeviceService{
		repository: repository, deviceTypes: deviceTypes, roles: roles, sites: sites,
		racks: racks, templates: templates, interfaces: interfaces,
		interfaceCascade: interfaceCascade, unitOfWork: unitOfWork,
		recorder: recorder, authorizer: authorizer, clock: clock,
	}, nil
}

func (service *DeviceService) ListDevices(
	ctx context.Context,
	principal identity.Principal,
	query ListDevicesQuery,
) (DevicePage, error) {
	if err := service.authorize(ctx, principal, authz.View, 0); err != nil {
		return DevicePage{}, err
	}
	criteria, err := validateListDevicesQuery(query)
	if err != nil {
		return DevicePage{}, err
	}
	scopeAuthorizer, hasCompleteScope := service.authorizer.(authz.ResourceListScopeAuthorizer)
	if hasCompleteScope {
		scope := scopeAuthorizer.ResourceListScope(
			ctx, principal, authz.View, authz.ResourceDevice,
		)
		criteria.VisibilityConstrained = scope.Constrained
		criteria.VisibleObjectIDs = sharedIDs(scope.ObjectIDs)
	} else {
		criteria.DeferPagination = true
	}
	page, err := service.repository.List(ctx, criteria)
	if err != nil {
		return DevicePage{}, normalizeOperationError("Could not list Devices.", err)
	}
	visible := make([]*dcimdomain.Device, 0, len(page.Results))
	for _, device := range page.Results {
		if device == nil {
			return DevicePage{}, shared.NewError(
				shared.ErrorReasonInternal, "Device repository returned an invalid list item.",
			)
		}
		authorizeErr := service.authorize(ctx, principal, authz.View, device.ID())
		if authorizeErr == nil {
			visible = append(visible, device)
			continue
		}
		if hasCompleteScope {
			return DevicePage{}, shared.WrapError(
				shared.ErrorReasonInternal,
				"Device visibility scope admitted an unauthorized object.",
				authorizeErr,
			)
		}
		if shared.HasReason(authorizeErr, shared.ErrorReasonForbidden) {
			continue
		}
		return DevicePage{}, authorizeErr
	}
	if hasCompleteScope {
		return DevicePage{Count: page.Count, Results: visible}, nil
	}
	count := uint64(len(visible))
	start := min(int(criteria.Offset), len(visible))
	end := min(start+int(criteria.Limit), len(visible))
	return DevicePage{Count: count, Results: visible[start:end]}, nil
}

func (service *DeviceService) GetDevice(
	ctx context.Context,
	principal identity.Principal,
	query GetDeviceQuery,
) (*dcimdomain.Device, error) {
	if err := service.authorize(ctx, principal, authz.View, 0); err != nil {
		return nil, err
	}
	if err := validatePersistedID(query.ID); err != nil {
		return nil, err
	}
	device, err := service.repository.Get(ctx, query.ID)
	if err != nil {
		return nil, normalizeOperationError("Could not get Device.", err)
	}
	if err := service.authorize(ctx, principal, authz.View, device.ID()); err != nil {
		return nil, err
	}
	return device, nil
}

func (service *DeviceService) CreateDevice(
	ctx context.Context,
	principal identity.Principal,
	command CreateDeviceCommand,
) (*dcimdomain.Device, error) {
	if err := service.authorize(ctx, principal, authz.Add, 0); err != nil {
		return nil, err
	}
	var device *dcimdomain.Device
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		commandValues, err := command.values()
		if err != nil {
			return err
		}
		values, err := service.resolveValues(transactionContext, commandValues)
		if err != nil {
			return err
		}
		now := service.clock.Now()
		candidate, err := dcimdomain.NewDevice(values, now)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Add, candidate.ID()); err != nil {
			return err
		}
		if err := service.validateRackPlacement(transactionContext, candidate); err != nil {
			return err
		}
		if err := service.repository.Create(transactionContext, candidate); err != nil {
			return err
		}
		if err := service.recordDeviceChange(
			transactionContext, principal, candidate, changelog.ActionCreate,
			nil, candidate.Snapshot(), now,
		); err != nil {
			return err
		}
		if err := service.instantiateInterfaces(
			transactionContext, principal, candidate, now,
		); err != nil {
			return err
		}
		device = candidate
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not create Device.", err)
	}
	return device, nil
}

func (service *DeviceService) ReplaceDevice(
	ctx context.Context,
	principal identity.Principal,
	command ReplaceDeviceCommand,
) (*dcimdomain.Device, error) {
	if err := service.authorize(ctx, principal, authz.Change, 0); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var device *dcimdomain.Device
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Change, loaded.ID()); err != nil {
			return err
		}
		commandValues, err := command.values()
		if err != nil {
			return err
		}
		values, err := service.resolveValues(transactionContext, commandValues)
		if err != nil {
			return err
		}
		before, now := loaded.Snapshot(), service.clock.Now()
		if err := loaded.Replace(values, now); err != nil {
			return err
		}
		if err := service.validateRackPlacement(transactionContext, loaded); err != nil {
			return err
		}
		if err := service.repository.Update(transactionContext, loaded); err != nil {
			return err
		}
		if err := service.recordDeviceChange(
			transactionContext, principal, loaded, changelog.ActionUpdate,
			before, loaded.Snapshot(), now,
		); err != nil {
			return err
		}
		device = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not replace Device.", err)
	}
	return device, nil
}

func (service *DeviceService) UpdateDevice(
	ctx context.Context,
	principal identity.Principal,
	command UpdateDeviceCommand,
) (*dcimdomain.Device, error) {
	if err := service.authorize(ctx, principal, authz.Change, 0); err != nil {
		return nil, err
	}
	if err := validatePersistedID(command.ID); err != nil {
		return nil, err
	}
	var device *dcimdomain.Device
	err := service.unitOfWork.WithinTransaction(ctx, func(transactionContext context.Context) error {
		loaded, err := service.repository.GetForUpdate(transactionContext, command.ID)
		if err != nil {
			return err
		}
		if err := service.authorize(transactionContext, principal, authz.Change, loaded.ID()); err != nil {
			return err
		}
		commandPatch, err := command.patch()
		if err != nil {
			return err
		}
		patch, err := service.resolvePatch(transactionContext, commandPatch)
		if err != nil {
			return err
		}
		before, now := loaded.Snapshot(), service.clock.Now()
		if err := loaded.ApplyPatch(patch, now); err != nil {
			return err
		}
		if err := service.validateRackPlacement(transactionContext, loaded); err != nil {
			return err
		}
		if err := service.repository.Update(transactionContext, loaded); err != nil {
			return err
		}
		if err := service.recordDeviceChange(
			transactionContext, principal, loaded, changelog.ActionUpdate,
			before, loaded.Snapshot(), now,
		); err != nil {
			return err
		}
		device = loaded
		return nil
	})
	if err != nil {
		return nil, normalizeOperationError("Could not update Device.", err)
	}
	return device, nil
}

func (service *DeviceService) DeleteDevice(
	ctx context.Context,
	principal identity.Principal,
	command DeleteDeviceCommand,
) error {
	if err := service.authorize(ctx, principal, authz.Delete, 0); err != nil {
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
		if err := service.authorize(transactionContext, principal, authz.Delete, loaded.ID()); err != nil {
			return err
		}
		now := service.clock.Now()
		children, err := service.interfaceCascade.DeleteForDevice(
			transactionContext, loaded.ID(), now,
		)
		if err != nil {
			return err
		}
		before := loaded.Snapshot()
		if err := service.repository.Delete(transactionContext, loaded); err != nil {
			return err
		}
		for _, child := range children {
			change, err := changelog.NewChange(
				principal.ID, child.ObjectType, child.ID, child.Representation,
				changelog.ActionDelete, child.Before, nil, now,
			)
			if err != nil {
				return err
			}
			if err := service.recorder.Record(transactionContext, change); err != nil {
				return err
			}
		}
		return service.recordDeviceChange(
			transactionContext, principal, loaded, changelog.ActionDelete,
			before, nil, now,
		)
	})
	return normalizeOperationError("Could not delete Device.", err)
}

func (service *DeviceService) resolveValues(
	ctx context.Context,
	values deviceCommandValues,
) (dcimdomain.DeviceValues, error) {
	deviceType, err := service.resolveDeviceType(ctx, values.deviceTypeID)
	if err != nil {
		return dcimdomain.DeviceValues{}, err
	}
	role, err := service.resolveRole(ctx, values.roleID)
	if err != nil {
		return dcimdomain.DeviceValues{}, err
	}
	site, err := service.resolveDeviceSite(ctx, values.siteID)
	if err != nil {
		return dcimdomain.DeviceValues{}, err
	}
	rack, err := service.resolveRack(ctx, values.rackID)
	if err != nil {
		return dcimdomain.DeviceValues{}, err
	}
	position, err := resolveDevicePosition(values.position)
	if err != nil {
		return dcimdomain.DeviceValues{}, err
	}
	return dcimdomain.DeviceValues{
		DeviceType: deviceType, Role: role, Name: values.name, Site: site,
		Rack: rack, Position: position, Face: values.face, Status: values.status,
		Serial: values.serial, AssetTag: values.assetTag, Airflow: values.airflow,
		Description: values.description, Comments: values.comments,
	}, nil
}

func (service *DeviceService) resolvePatch(
	ctx context.Context,
	patch deviceCommandPatch,
) (dcimdomain.DevicePatch, error) {
	resolved := dcimdomain.DevicePatch{
		Name: patch.name, Face: patch.face, Status: patch.status, Serial: patch.serial,
		AssetTag: patch.assetTag, Airflow: patch.airflow,
		Description: patch.description, Comments: patch.comments,
	}
	if patch.deviceTypeID != nil {
		value, err := service.resolveDeviceType(ctx, *patch.deviceTypeID)
		if err != nil {
			return dcimdomain.DevicePatch{}, err
		}
		resolved.DeviceType = &value
	}
	if patch.roleID != nil {
		value, err := service.resolveRole(ctx, *patch.roleID)
		if err != nil {
			return dcimdomain.DevicePatch{}, err
		}
		resolved.Role = &value
	}
	if patch.siteID != nil {
		value, err := service.resolveDeviceSite(ctx, *patch.siteID)
		if err != nil {
			return dcimdomain.DevicePatch{}, err
		}
		resolved.Site = &value
	}
	if patch.rackID != nil {
		value, err := service.resolveRack(ctx, *patch.rackID)
		if err != nil {
			return dcimdomain.DevicePatch{}, err
		}
		resolved.Rack = &value
	}
	if patch.position != nil {
		value, err := resolveDevicePosition(*patch.position)
		if err != nil {
			return dcimdomain.DevicePatch{}, err
		}
		resolved.Position = &value
	}
	return resolved, nil
}

func (service *DeviceService) resolveDeviceType(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.DeviceTypeInstanceReference, error) {
	value, err := service.deviceTypes.Get(ctx, id)
	if err != nil {
		return dcimdomain.DeviceTypeInstanceReference{}, relatedDeviceError("device_type", id, err)
	}
	return dcimdomain.NewDeviceTypeInstanceReference(
		value.ID(), value.Model(), value.Slug().String(), value.Manufacturer().Name(),
		value.UHeight(), value.IsFullDepth(), value.Airflow(),
	)
}

func (service *DeviceService) resolveRole(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.DeviceRoleReference, error) {
	value, err := service.roles.Get(ctx, id)
	if err != nil {
		return dcimdomain.DeviceRoleReference{}, relatedDeviceError("role", id, err)
	}
	return dcimdomain.DeviceRoleReference{ID: value.ID(), Display: value.Display()}, nil
}

func (service *DeviceService) resolveDeviceSite(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.SiteReference, error) {
	value, err := service.sites.Get(ctx, id)
	if err != nil {
		return dcimdomain.SiteReference{}, relatedDeviceError("site", id, err)
	}
	return dcimdomain.NewSiteReference(value.ID(), value.Name(), value.Slug().String())
}

func (service *DeviceService) resolveRack(
	ctx context.Context,
	value dcimdomain.DeviceNullable[shared.ID],
) (dcimdomain.DeviceNullable[dcimdomain.RackReference], error) {
	id, present := value.Get()
	if !present {
		return dcimdomain.NullDeviceValue[dcimdomain.RackReference](), nil
	}
	rack, err := service.racks.GetForUpdate(ctx, id)
	if err != nil {
		return dcimdomain.DeviceNullable[dcimdomain.RackReference]{}, relatedDeviceError("rack", id, err)
	}
	reference, err := dcimdomain.NewRackReference(
		rack.ID(), rack.Display(), rack.Site().ID(), rack.StartingUnit(), rack.UHeight(),
	)
	if err != nil {
		return dcimdomain.DeviceNullable[dcimdomain.RackReference]{}, err
	}
	return dcimdomain.NonNullDeviceValue(reference), nil
}

func resolveDevicePosition(
	value dcimdomain.DeviceNullable[string],
) (dcimdomain.DeviceNullable[dcimdomain.RackPosition], error) {
	text, present := value.Get()
	if !present {
		return dcimdomain.NullDeviceValue[dcimdomain.RackPosition](), nil
	}
	position, err := dcimdomain.ParseRackPosition(text)
	if err != nil {
		return dcimdomain.DeviceNullable[dcimdomain.RackPosition]{}, err
	}
	return dcimdomain.NonNullDeviceValue(position), nil
}

func relatedDeviceError(field string, id shared.ID, err error) error {
	if shared.HasReason(err, shared.ErrorReasonNotFound) {
		return shared.NewValidationError(shared.FieldViolation{
			Field: field, Reason: "invalid_choice",
			Description: "Invalid pk \"" + strconv.FormatInt(id.Int64(), 10) + "\" - object does not exist.",
		})
	}
	return err
}

func (service *DeviceService) validateRackPlacement(
	ctx context.Context,
	device *dcimdomain.Device,
) error {
	rackRef, hasRack := device.Rack().Get()
	position, positioned := device.Position().Get()
	if !hasRack || !positioned {
		return nil
	}
	rack, err := service.racks.GetForUpdate(ctx, rackRef.ID())
	if err != nil {
		return relatedDeviceError("rack", rackRef.ID(), err)
	}
	if rack.Site().ID() != device.Site().ID() {
		return shared.NewValidationError(shared.FieldViolation{
			Field: "rack", Reason: "invalid_relationship",
			Description: "The selected rack does not belong to this site.",
		})
	}
	start := uint32(position.HalfUnits())
	height := uint32(device.DeviceType().Height().HalfUnits())
	rackStart := rack.StartingUnit() * 2
	rackEnd := rackStart + rack.UHeight()*2
	if start < rackStart || start+height > rackEnd {
		return devicePlacementError(device)
	}
	occupants, err := service.repository.ListRackOccupantsForUpdate(
		ctx, rackRef.ID(), device.ID(),
	)
	if err != nil {
		return err
	}
	end := start + height
	for _, occupant := range occupants {
		otherStart := uint32(occupant.PositionHalfUnits)
		otherEnd := otherStart + uint32(occupant.HeightHalfUnits)
		if start >= otherEnd || otherStart >= end {
			continue
		}
		if device.DeviceType().IsFullDepth() || occupant.FullDepth ||
			device.Face() == occupant.Face {
			return devicePlacementError(device)
		}
	}
	return nil
}

func devicePlacementError(device *dcimdomain.Device) error {
	position, _ := device.Position().Get()
	deviceType := device.DeviceType()
	return shared.NewValidationError(shared.FieldViolation{
		Field: "position", Reason: "occupied_or_out_of_bounds",
		Description: fmt.Sprintf(
			"U%s is already occupied or does not have sufficient space to accommodate this device type: %s (%sU)",
			position.String(), deviceType.Display(), deviceType.Height().String(),
		),
	})
}

func (service *DeviceService) instantiateInterfaces(
	ctx context.Context,
	principal identity.Principal,
	device *dcimdomain.Device,
	now shared.Timestamp,
) error {
	page, err := service.templates.List(ctx, InterfaceTemplateListCriteria{
		DeviceTypeIDs: []int64{device.DeviceType().ID().Int64()},
		Ordering:      []InterfaceTemplateSort{{Field: InterfaceTemplateSortID}},
	})
	if err != nil {
		return err
	}
	deviceReference, err := dcimdomain.NewDeviceReference(
		device.ID(), device.Name(), device.Display(),
	)
	if err != nil {
		return err
	}
	for _, template := range page.Results {
		if template == nil {
			return shared.NewError(
				shared.ErrorReasonInternal,
				"InterfaceTemplate repository returned an invalid list item.",
			)
		}
		networkInterface, err := dcimdomain.NewInterface(dcimdomain.InterfaceValues{
			Device: deviceReference, Name: template.Name(), Label: template.Label(),
			Type: template.Type().String(), Enabled: template.Enabled(),
			MgmtOnly:    template.MgmtOnly(),
			MTU:         dcimdomain.NullDeviceValue[uint32](),
			Speed:       dcimdomain.NullDeviceValue[uint64](),
			Duplex:      dcimdomain.NullDeviceValue[string](),
			Description: "",
		}, now)
		if err != nil {
			return err
		}
		if err := service.interfaces.Create(ctx, networkInterface); err != nil {
			return err
		}
		change, err := changelog.NewChange(
			principal.ID, dcimdomain.InterfaceObjectType, networkInterface.ID(),
			networkInterface.Display(), changelog.ActionCreate,
			nil, networkInterface.Snapshot(), now,
		)
		if err != nil {
			return err
		}
		if err := service.recorder.Record(ctx, change); err != nil {
			return err
		}
	}
	return nil
}

func (service *DeviceService) recordDeviceChange(
	ctx context.Context,
	principal identity.Principal,
	device *dcimdomain.Device,
	action changelog.Action,
	before shared.ObjectSnapshot,
	after shared.ObjectSnapshot,
	now shared.Timestamp,
) error {
	change, err := changelog.NewChange(
		principal.ID, dcimdomain.DeviceObjectType, device.ID(), device.Display(),
		action, before, after, now,
	)
	if err != nil {
		return err
	}
	return service.recorder.Record(ctx, change)
}

func (service *DeviceService) authorize(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	id shared.ID,
) error {
	if err := service.authorizer.AuthorizeResource(
		ctx, principal, action, authz.ResourceDevice, authz.NewObject(id.Int64()),
	); err != nil {
		return normalizeAuthorizationError(err)
	}
	return nil
}

var _ InterfaceDeviceReader = (*DeviceService)(nil)

func (service *DeviceService) GetDeviceReference(
	ctx context.Context,
	id shared.ID,
) (dcimdomain.DeviceReference, error) {
	device, err := service.repository.Get(ctx, id)
	if err != nil {
		return dcimdomain.DeviceReference{}, err
	}
	return dcimdomain.NewDeviceReference(device.ID(), device.Name(), device.Display())
}
