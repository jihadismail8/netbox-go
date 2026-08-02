package dcim

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	deviceTableAlias        = "typed_devices"
	deviceTypeAlias         = "device_type_refs"
	deviceManufacturerAlias = "device_manufacturers"
	deviceRoleAlias         = "device_role_refs"
	deviceSiteAlias         = "device_site_refs"
	deviceRackAlias         = "device_rack_refs"
)

type DeviceRepository struct{ db *gorm.DB }

var _ applicationdcim.DeviceRepository = (*DeviceRepository)(nil)
var _ applicationdcim.InterfaceDeviceReader = (*DeviceRepository)(nil)

func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	if db == nil {
		panic("postgres Device repository requires a database")
	}
	return &DeviceRepository{db: db}
}

func (repository *DeviceRepository) List(
	ctx context.Context,
	criteria applicationdcim.DeviceListCriteria,
) (applicationdcim.DevicePage, error) {
	filtered := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := filtered.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.DevicePage{}, translateDeviceReadError(0, "count Devices", err)
	}
	if count < 0 {
		return applicationdcim.DevicePage{}, shared.NewError(
			shared.ErrorReasonInternal, "Device count returned an invalid value.",
		)
	}
	query := applyDeviceOrdering(selectDeviceProjection(filtered), criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []deviceProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.DevicePage{}, translateDeviceReadError(0, "list Devices", err)
	}
	results := make([]*domaindcim.Device, 0, len(rows))
	for _, row := range rows {
		device, err := deviceFromProjection(row)
		if err != nil {
			return applicationdcim.DevicePage{}, err
		}
		results = append(results, device)
	}
	return applicationdcim.DevicePage{Count: uint64(count), Results: results}, nil
}

func (repository *DeviceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Device, error) {
	return repository.get(ctx, id, false)
}

func (repository *DeviceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Device, error) {
	return repository.get(ctx, id, true)
}

func (repository *DeviceRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.Device, error) {
	query := selectDeviceProjection(
		repository.baseQuery(ctx).Where(deviceTableAlias+".id = ?", id.Int64()),
	)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: deviceTableAlias},
		})
	}
	var row deviceProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateDeviceReadError(id, "get Device", err)
	}
	return deviceFromProjection(row)
}

func (repository *DeviceRepository) GetDeviceReference(
	ctx context.Context,
	id shared.ID,
) (domaindcim.DeviceReference, error) {
	device, err := repository.Get(ctx, id)
	if err != nil {
		return domaindcim.DeviceReference{}, err
	}
	return domaindcim.NewDeviceReference(device.ID(), device.Name(), device.Display())
}

func (repository *DeviceRepository) Create(
	ctx context.Context,
	device *domaindcim.Device,
) error {
	if device == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil Device.")
	}
	if device.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted Device.")
	}
	row := deviceToRow(*device)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateDeviceMutationError("create Device", err)
	}
	return device.AssignID(shared.ID(row.ID))
}

func (repository *DeviceRepository) Update(
	ctx context.Context,
	device *domaindcim.Device,
) error {
	if device == nil || !device.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted Device.")
	}
	row := deviceToRow(*device)
	result := repository.database(ctx).
		Model(&dcimrow.DeviceRow{}).
		Where("id = ?", device.ID().Int64()).
		Select(
			"device_type_id", "role_id", "name", "site_id", "rack_id", "position",
			"face", "status", "serial", "asset_tag", "airflow", "description",
			"comments", "last_updated",
		).
		Updates(&row)
	if result.Error != nil {
		return translateDeviceMutationError("update Device", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Device", device.ID())
	}
	return nil
}

func (repository *DeviceRepository) Delete(
	ctx context.Context,
	device *domaindcim.Device,
) error {
	if device == nil || !device.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted Device.")
	}
	result := repository.database(ctx).
		Where("id = ?", device.ID().Int64()).
		Delete(&dcimrow.DeviceRow{})
	if result.Error != nil {
		return translateDeviceMutationError("delete Device", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Device", device.ID())
	}
	return nil
}

func (repository *DeviceRepository) ListRackOccupantsForUpdate(
	ctx context.Context,
	rackID shared.ID,
	excludeID shared.ID,
) ([]applicationdcim.DeviceRackOccupant, error) {
	if !rackID.IsValid() {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Cannot inspect an invalid Device Rack.",
		)
	}
	var rows []struct {
		ID          int64
		Position    float64
		UHeight     float64
		Face        string
		IsFullDepth bool
	}
	query := repository.database(ctx).
		Table((dcimrow.DeviceRow{}).TableName()+" AS rack_devices").
		Select(
			"rack_devices.id, rack_devices.position, rack_device_types.u_height, "+
				"rack_devices.face, rack_device_types.is_full_depth",
		).
		Joins(
			"JOIN "+(dcimrow.DeviceTypeRow{}).TableName()+" AS rack_device_types "+
				"ON rack_device_types.id = rack_devices.device_type_id",
		).
		Where("rack_devices.rack_id = ? AND rack_devices.position IS NOT NULL", rackID.Int64()).
		Order("rack_devices.id").
		Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: "rack_devices"},
		})
	if excludeID.IsValid() {
		query = query.Where("rack_devices.id <> ?", excludeID.Int64())
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not inspect Device Rack occupancy.", err,
		)
	}
	result := make([]applicationdcim.DeviceRackOccupant, 0, len(rows))
	for _, row := range rows {
		position := row.Position * 2
		height := row.UHeight * 2
		face, validFace := domaindcim.ParseDeviceFace(row.Face)
		if row.ID <= 0 || math.Trunc(position) != position || position < 2 ||
			position > math.MaxUint16 || math.Trunc(height) != height || height < 0 ||
			height > math.MaxUint16 || !validFace {
			return nil, shared.NewError(
				shared.ErrorReasonInternal,
				"Persisted Device contains invalid Rack occupancy state.",
			)
		}
		result = append(result, applicationdcim.DeviceRackOccupant{
			ID: shared.ID(row.ID), PositionHalfUnits: uint16(position),
			HeightHalfUnits: uint16(height), Face: face, FullDepth: row.IsFullDepth,
		})
	}
	return result, nil
}

func (repository *DeviceRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.DeviceListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(deviceTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(deviceTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs))
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+deviceTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTableAlias+".serial) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTableAlias+".asset_tag) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTableAlias+".description) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTableAlias+".comments) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern, pattern, pattern,
		)
	}
	if len(criteria.SiteIDs) > 0 {
		query = query.Where(deviceTableAlias+".site_id IN ?", criteria.SiteIDs)
	}
	if len(criteria.SiteSlugs) > 0 {
		query = query.Where(deviceSiteAlias+".slug IN ?", criteria.SiteSlugs)
	}
	if len(criteria.RackIDs) > 0 {
		query = query.Where(deviceTableAlias+".rack_id IN ?", criteria.RackIDs)
	}
	if len(criteria.DeviceTypeIDs) > 0 {
		query = query.Where(deviceTableAlias+".device_type_id IN ?", criteria.DeviceTypeIDs)
	}
	if len(criteria.DeviceTypeSlugs) > 0 {
		query = query.Where(deviceTypeAlias+".slug IN ?", criteria.DeviceTypeSlugs)
	}
	if len(criteria.RoleIDs) > 0 {
		query = query.Where(deviceTableAlias+".role_id IN ?", criteria.RoleIDs)
	}
	if len(criteria.RoleSlugs) > 0 {
		query = query.Where(deviceRoleAlias+".slug IN ?", criteria.RoleSlugs)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(deviceTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Statuses) > 0 {
		values := make([]string, len(criteria.Statuses))
		for index, status := range criteria.Statuses {
			values[index] = status.String()
		}
		query = query.Where(deviceTableAlias+".status IN ?", values)
	}
	return query
}

func (repository *DeviceRepository) baseQuery(ctx context.Context) *gorm.DB {
	return repository.database(ctx).
		Table((dcimrow.DeviceRow{}).TableName() + " AS " + deviceTableAlias).
		Joins(
			"JOIN " + (dcimrow.DeviceTypeRow{}).TableName() + " AS " + deviceTypeAlias +
				" ON " + deviceTypeAlias + ".id = " + deviceTableAlias + ".device_type_id",
		).
		Joins(
			"JOIN " + (dcimrow.ManufacturerRow{}).TableName() + " AS " + deviceManufacturerAlias +
				" ON " + deviceManufacturerAlias + ".id = " + deviceTypeAlias + ".manufacturer_id",
		).
		Joins(
			"JOIN " + (dcimrow.DeviceRoleRow{}).TableName() + " AS " + deviceRoleAlias +
				" ON " + deviceRoleAlias + ".id = " + deviceTableAlias + ".role_id",
		).
		Joins(
			"JOIN " + (dcimrow.SiteRow{}).TableName() + " AS " + deviceSiteAlias +
				" ON " + deviceSiteAlias + ".id = " + deviceTableAlias + ".site_id",
		).
		Joins(
			"LEFT JOIN " + (dcimrow.RackRow{}).TableName() + " AS " + deviceRackAlias +
				" ON " + deviceRackAlias + ".id = " + deviceTableAlias + ".rack_id",
		)
}

func (repository *DeviceRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type deviceProjectionRow struct {
	dcimrow.DeviceRow
	DeviceTypeModel       string  `gorm:"column:device_type_model"`
	DeviceTypeSlug        string  `gorm:"column:device_type_slug"`
	DeviceTypeUHeight     float64 `gorm:"column:device_type_u_height"`
	DeviceTypeIsFullDepth bool    `gorm:"column:device_type_is_full_depth"`
	DeviceTypeAirflow     *string `gorm:"column:device_type_airflow"`
	ManufacturerName      string  `gorm:"column:manufacturer_name"`
	RoleName              string  `gorm:"column:role_name"`
	SiteName              string  `gorm:"column:site_name"`
	SiteSlug              string  `gorm:"column:site_slug"`
	RackName              string  `gorm:"column:rack_name"`
	RackFacilityID        *string `gorm:"column:rack_facility_id"`
	RackSiteID            *int64  `gorm:"column:rack_site_id"`
	RackStartingUnit      *int64  `gorm:"column:rack_starting_unit"`
	RackUHeight           *int64  `gorm:"column:rack_u_height"`
	InterfaceCount        int64   `gorm:"column:interface_count"`
}

func selectDeviceProjection(query *gorm.DB) *gorm.DB {
	interfaceTable := (dcimrow.InterfaceRow{}).TableName()
	return query.Select(
		deviceTableAlias + ".*, " +
			deviceTypeAlias + ".model AS device_type_model, " +
			deviceTypeAlias + ".slug AS device_type_slug, " +
			deviceTypeAlias + ".u_height AS device_type_u_height, " +
			deviceTypeAlias + ".is_full_depth AS device_type_is_full_depth, " +
			deviceTypeAlias + ".airflow AS device_type_airflow, " +
			deviceManufacturerAlias + ".name AS manufacturer_name, " +
			deviceRoleAlias + ".name AS role_name, " +
			deviceSiteAlias + ".name AS site_name, " +
			deviceSiteAlias + ".slug AS site_slug, " +
			deviceRackAlias + ".name AS rack_name, " +
			deviceRackAlias + ".facility_id AS rack_facility_id, " +
			deviceRackAlias + ".site_id AS rack_site_id, " +
			deviceRackAlias + ".starting_unit AS rack_starting_unit, " +
			deviceRackAlias + ".u_height AS rack_u_height, " +
			"(SELECT COUNT(*) FROM " + interfaceTable + " WHERE " +
			interfaceTable + ".device_id = " + deviceTableAlias + ".id) AS interface_count",
	)
}

func applyDeviceOrdering(
	query *gorm.DB,
	ordering []applicationdcim.DeviceSort,
) *gorm.DB {
	hasID := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: deviceTableAlias, Name: deviceSortColumn(requested.Field)},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationdcim.DeviceSortID {
			hasID = true
		}
	}
	if !hasID {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: deviceTableAlias, Name: "id"},
		})
	}
	return query
}

func deviceSortColumn(field applicationdcim.DeviceSortField) string {
	switch field {
	case applicationdcim.DeviceSortID:
		return "id"
	case applicationdcim.DeviceSortSite:
		return "site_id"
	case applicationdcim.DeviceSortRack:
		return "rack_id"
	case applicationdcim.DeviceSortPosition:
		return "position"
	case applicationdcim.DeviceSortName:
		return "name"
	case applicationdcim.DeviceSortStatus:
		return "status"
	case applicationdcim.DeviceSortCreated:
		return "created"
	case applicationdcim.DeviceSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func deviceToRow(device domaindcim.Device) dcimrow.DeviceRow {
	state := device.State()
	var rackID *int64
	if rack, present := state.Rack.Get(); present {
		value := rack.ID().Int64()
		rackID = &value
	}
	var position *float64
	if value, present := state.Position.Get(); present {
		number := float64(value.HalfUnits()) / 2
		position = &number
	}
	return dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		DeviceTypeID: state.DeviceType.ID().Int64(), RoleID: state.Role.ID.Int64(),
		Name: deviceNullableString(state.Name), SiteID: state.Site.ID().Int64(),
		RackID: rackID, Position: position, Face: state.Face, Status: state.Status,
		Serial: state.Serial, AssetTag: deviceNullableString(state.AssetTag),
		Airflow:     nullableDeviceAirflow(state.Airflow),
		Description: state.Description, Comments: state.Comments,
	}
}

func deviceFromProjection(row deviceProjectionRow) (*domaindcim.Device, error) {
	if row.InterfaceCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Persisted Device contains an invalid Interface count.",
		)
	}
	typeHeight := row.DeviceTypeUHeight * 2
	if math.Trunc(typeHeight) != typeHeight || typeHeight < 0 || typeHeight > math.MaxUint16 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Persisted DeviceType reference contains an invalid height.",
		)
	}
	height, err := domaindcim.DeviceHeightFromHalfUnits(uint16(typeHeight))
	if err != nil {
		return nil, err
	}
	deviceTypeAirflow, err := nullableDeviceAirflowFromPointer(row.DeviceTypeAirflow)
	if err != nil {
		return nil, err
	}
	deviceType, err := domaindcim.NewDeviceTypeInstanceReference(
		shared.ID(row.DeviceTypeID), row.DeviceTypeModel, row.DeviceTypeSlug,
		row.ManufacturerName, height, row.DeviceTypeIsFullDepth, deviceTypeAirflow,
	)
	if err != nil {
		return nil, err
	}
	site, err := domaindcim.NewSiteReference(
		shared.ID(row.SiteID), row.SiteName, row.SiteSlug,
	)
	if err != nil {
		return nil, err
	}
	rack := domaindcim.NullDeviceValue[domaindcim.RackReference]()
	if row.RackID != nil {
		if row.RackSiteID == nil || row.RackStartingUnit == nil || row.RackUHeight == nil {
			return nil, shared.NewError(
				shared.ErrorReasonInternal, "Persisted Device contains an invalid Rack reference.",
			)
		}
		display := row.RackName
		if row.RackFacilityID != nil && *row.RackFacilityID != "" {
			display += " (" + *row.RackFacilityID + ")"
		}
		reference, referenceErr := domaindcim.NewRackReference(
			shared.ID(*row.RackID), display, shared.ID(*row.RackSiteID),
			uint32(*row.RackStartingUnit), uint32(*row.RackUHeight),
		)
		if referenceErr != nil {
			return nil, referenceErr
		}
		rack = domaindcim.NonNullDeviceValue(reference)
	}
	position := domaindcim.NullDeviceValue[domaindcim.RackPosition]()
	if row.Position != nil {
		halfUnits := *row.Position * 2
		if math.Trunc(halfUnits) != halfUnits || halfUnits < 0 || halfUnits > math.MaxUint16 {
			return nil, shared.NewError(
				shared.ErrorReasonInternal, "Persisted Device contains an invalid Rack position.",
			)
		}
		value, positionErr := domaindcim.RackPositionFromHalfUnits(uint16(halfUnits))
		if positionErr != nil {
			return nil, positionErr
		}
		position = domaindcim.NonNullDeviceValue(value)
	}
	airflow, err := nullableDeviceAirflowFromPointer(row.Airflow)
	if err != nil {
		return nil, err
	}
	device, err := domaindcim.RestoreDevice(domaindcim.DeviceState{
		ID: shared.ID(row.ID), DeviceType: deviceType,
		Role: domaindcim.DeviceRoleReference{ID: shared.ID(row.RoleID), Display: row.RoleName},
		Name: deviceNullableStringFromPointer(row.Name), Site: site, Rack: rack,
		Position: position, Face: row.Face, Status: row.Status, Serial: row.Serial,
		AssetTag: deviceNullableStringFromPointer(row.AssetTag), Airflow: airflow,
		Description: row.Description, Comments: row.Comments,
		Created: shared.NewTimestamp(row.Created), LastUpdated: shared.NewTimestamp(row.LastUpdated),
		InterfaceCount: uint64(row.InterfaceCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not restore persisted Device state.", err,
		)
	}
	return device, nil
}

func deviceNullableString(value domaindcim.DeviceNullable[string]) *string {
	text, present := value.Get()
	if !present {
		return nil
	}
	return &text
}

func deviceNullableStringFromPointer(value *string) domaindcim.DeviceNullable[string] {
	if value == nil {
		return domaindcim.NullDeviceValue[string]()
	}
	return domaindcim.NonNullDeviceValue(*value)
}

func nullableDeviceAirflow(value domaindcim.NullableDeviceAirflow) *string {
	airflow, present := value.Get()
	if !present {
		return nil
	}
	text := airflow.String()
	return &text
}

func nullableDeviceAirflowFromPointer(
	value *string,
) (domaindcim.NullableDeviceAirflow, error) {
	if value == nil {
		return domaindcim.NullDeviceAirflow(), nil
	}
	airflow, valid := domaindcim.ParseDeviceAirflow(*value)
	if !valid {
		return domaindcim.NullableDeviceAirflow{}, shared.NewError(
			shared.ErrorReasonInternal, "Persisted Device contains an invalid airflow.",
		)
	}
	return domaindcim.NonNullDeviceAirflow(airflow), nil
}

func translateDeviceReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("Device", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}

func translateDeviceMutationError(operation string, err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_device_site_name_ci"),
		strings.Contains(message, "go_dcim_devices.site_id") &&
			strings.Contains(message, "go_dcim_devices.name"):
		const description = "A device with this name already exists at this site."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{Field: "name", Reason: "unique", Description: description},
		)
	case strings.Contains(message, "uq_go_device_asset_tag"),
		strings.Contains(message, "go_dcim_devices.asset_tag"):
		const description = "device with this asset tag already exists."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{Field: "asset_tag", Reason: "unique", Description: description},
		)
	case strings.Contains(message, "uq_go_device_rack_position_face"):
		const description = "The selected rack position is already occupied."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{
				Field: "position", Reason: "occupied_or_out_of_bounds",
				Description: description,
			},
		)
	case duplicateConstraint(err):
		return shared.Conflict("A matching Device already exists.", err)
	case foreignKeyConstraint(err):
		return shared.WrapError(
			shared.ErrorReasonProtected, "The Device is referenced by another object.", err,
		)
	default:
		return shared.WrapError(
			shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
		)
	}
}
