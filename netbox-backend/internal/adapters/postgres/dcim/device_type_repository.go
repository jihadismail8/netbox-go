package dcim

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
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
	deviceTypeTableAlias             = "device_types"
	deviceTypeManufacturerTableAlias = "device_type_manufacturers"
)

type DeviceTypeRepository struct{ db *gorm.DB }

var _ applicationdcim.DeviceTypeRepository = (*DeviceTypeRepository)(nil)

func NewDeviceTypeRepository(db *gorm.DB) *DeviceTypeRepository {
	if db == nil {
		panic("dcim device-type repository requires a database")
	}
	return &DeviceTypeRepository{db: db}
}

func (repository *DeviceTypeRepository) List(
	ctx context.Context,
	criteria applicationdcim.DeviceTypeListCriteria,
) (applicationdcim.DeviceTypePage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.DeviceTypePage{},
			translateDeviceTypeReadError(0, "count DeviceTypes", err)
	}
	if count < 0 {
		return applicationdcim.DeviceTypePage{}, shared.NewError(
			shared.ErrorReasonInternal, "DeviceType count returned an invalid value.",
		)
	}
	query := applyDeviceTypeOrdering(selectDeviceTypeProjection(base), criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []deviceTypeProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.DeviceTypePage{},
			translateDeviceTypeReadError(0, "list DeviceTypes", err)
	}
	results := make([]*domaindcim.DeviceType, 0, len(rows))
	for _, row := range rows {
		deviceType, err := deviceTypeFromProjection(row)
		if err != nil {
			return applicationdcim.DeviceTypePage{}, err
		}
		results = append(results, deviceType)
	}
	return applicationdcim.DeviceTypePage{Count: uint64(count), Results: results}, nil
}

func (repository *DeviceTypeRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.DeviceType, error) {
	return repository.get(ctx, id, false)
}

func (repository *DeviceTypeRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.DeviceType, error) {
	return repository.get(ctx, id, true)
}

func (repository *DeviceTypeRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.DeviceType, error) {
	query := selectDeviceTypeProjection(
		repository.baseQuery(ctx).Where(deviceTypeTableAlias+".id = ?", id.Int64()),
	)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: deviceTypeTableAlias},
		})
	}
	var row deviceTypeProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateDeviceTypeReadError(id, "get DeviceType", err)
	}
	return deviceTypeFromProjection(row)
}

func (repository *DeviceTypeRepository) Create(
	ctx context.Context,
	deviceType *domaindcim.DeviceType,
) error {
	if deviceType == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil DeviceType.")
	}
	if deviceType.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot create an already persisted DeviceType.",
		)
	}
	row := deviceTypeToRow(*deviceType)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateDeviceTypeMutationError("create DeviceType", err)
	}
	return deviceType.AssignID(shared.ID(row.ID))
}

func (repository *DeviceTypeRepository) Update(
	ctx context.Context,
	deviceType *domaindcim.DeviceType,
) error {
	if deviceType == nil || !deviceType.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot update an unpersisted DeviceType.",
		)
	}
	row := deviceTypeToRow(*deviceType)
	result := repository.database(ctx).
		Model(&dcimrow.DeviceTypeRow{}).
		Where("id = ?", deviceType.ID().Int64()).
		Select(
			"manufacturer_id", "model", "slug", "part_number", "u_height",
			"exclude_from_utilization", "is_full_depth", "airflow",
			"description", "comments", "last_updated",
		).
		Updates(&row)
	if result.Error != nil {
		return translateDeviceTypeMutationError("update DeviceType", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("DeviceType", deviceType.ID())
	}
	return nil
}

func (repository *DeviceTypeRepository) ListPositionedDevicesForUpdate(
	ctx context.Context,
) ([]applicationdcim.PositionedDevice, error) {
	deviceTable := (dcimrow.DeviceRow{}).TableName()
	rackTable := (dcimrow.RackRow{}).TableName()
	typeTable := (dcimrow.DeviceTypeRow{}).TableName()
	type positionedRow struct {
		ID               int64
		DeviceTypeID     int64
		RackID           int64
		Position         float64
		Face             string
		StoredUHeight    float64
		StoredFullDepth  bool
		RackStartingUnit int64
		RackUHeight      int64
	}
	var rows []positionedRow
	err := repository.database(ctx).
		Table(deviceTable + " AS positioned_devices").
		Select(
			"positioned_devices.id, positioned_devices.device_type_id, " +
				"positioned_devices.rack_id, positioned_devices.position, positioned_devices.face, " +
				"positioned_types.u_height AS stored_u_height, " +
				"positioned_types.is_full_depth AS stored_full_depth, " +
				"positioned_racks.starting_unit AS rack_starting_unit, " +
				"positioned_racks.u_height AS rack_u_height",
		).
		Joins(
			"JOIN " + typeTable + " AS positioned_types ON " +
				"positioned_types.id = positioned_devices.device_type_id",
		).
		Joins(
			"JOIN " + rackTable + " AS positioned_racks ON " +
				"positioned_racks.id = positioned_devices.rack_id",
		).
		Where("positioned_devices.rack_id IS NOT NULL").
		Where("positioned_devices.position IS NOT NULL").
		Order("positioned_devices.id").
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Find(&rows).Error
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not lock positioned Devices.", err,
		)
	}
	placements := make([]applicationdcim.PositionedDevice, 0, len(rows))
	for _, row := range rows {
		if row.ID <= 0 || row.DeviceTypeID <= 0 || row.RackID <= 0 ||
			row.Position < 0 || math.Trunc(row.Position*2) != row.Position*2 ||
			row.RackStartingUnit < 0 || row.RackUHeight < 0 ||
			row.RackStartingUnit > math.MaxUint32 || row.RackUHeight > math.MaxUint32 {
			return nil, shared.NewError(
				shared.ErrorReasonInternal,
				"Persisted positioned Device contains invalid rack state.",
			)
		}
		storedHeight, err := domaindcim.ParseDeviceHeight(
			strconv.FormatFloat(row.StoredUHeight, 'f', -1, 64),
		)
		if err != nil {
			return nil, shared.WrapError(
				shared.ErrorReasonInternal,
				"Persisted positioned Device has an invalid DeviceType height.",
				err,
			)
		}
		placements = append(placements, applicationdcim.PositionedDevice{
			ID: shared.ID(row.ID), DeviceTypeID: shared.ID(row.DeviceTypeID),
			RackID:                shared.ID(row.RackID),
			PositionHalfUnits:     uint32(math.Round(row.Position * 2)),
			Face:                  row.Face,
			StoredHeightHalfUnits: uint32(storedHeight.HalfUnits()),
			StoredFullDepth:       row.StoredFullDepth,
			RackStartingUnit:      uint32(row.RackStartingUnit),
			RackUHeight:           uint32(row.RackUHeight),
		})
	}
	return placements, nil
}

func (repository *DeviceTypeRepository) FindDeviceUsingDeviceType(
	ctx context.Context,
	id shared.ID,
) (*applicationdcim.DeviceTypeDependent, error) {
	type dependentRow struct {
		ID       int64
		Name     *string
		AssetTag *string
	}
	var row dependentRow
	err := repository.database(ctx).
		Model(&dcimrow.DeviceRow{}).
		Select("id", "name", "asset_tag").
		Where("device_type_id = ?", id.Int64()).
		Order("id").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not check DeviceType dependencies.", err,
		)
	}
	display := "device " + shared.ID(row.ID).String()
	if row.Name != nil && strings.TrimSpace(*row.Name) != "" {
		display = *row.Name
		if row.AssetTag != nil && strings.TrimSpace(*row.AssetTag) != "" {
			display += " (" + *row.AssetTag + ")"
		}
	}
	return &applicationdcim.DeviceTypeDependent{
		ID: shared.ID(row.ID), Display: display,
	}, nil
}

func (repository *DeviceTypeRepository) ListInterfaceTemplatesForUpdate(
	ctx context.Context,
	id shared.ID,
) ([]applicationdcim.InterfaceTemplateDeletion, error) {
	var rows []dcimrow.InterfaceTemplateRow
	err := repository.database(ctx).
		Where("device_type_id = ?", id.Int64()).
		Order("id").
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Find(&rows).Error
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not lock DeviceType InterfaceTemplates.",
			err,
		)
	}
	templates := make([]applicationdcim.InterfaceTemplateDeletion, 0, len(rows))
	for _, row := range rows {
		if row.ID <= 0 || row.DeviceTypeID <= 0 || strings.TrimSpace(row.Name) == "" {
			return nil, shared.NewError(
				shared.ErrorReasonInternal,
				"Persisted InterfaceTemplate contains invalid deletion state.",
			)
		}
		display := row.Name
		if strings.TrimSpace(row.Label) != "" {
			display += " (" + row.Label + ")"
		}
		templates = append(templates, applicationdcim.InterfaceTemplateDeletion{
			ID: shared.ID(row.ID), Representation: display,
			Snapshot: domaindcim.InterfaceTemplateSnapshot{
				DeviceTypeID: shared.ID(row.DeviceTypeID), Name: row.Name,
				Label: row.Label, Type: row.Type, Enabled: row.Enabled,
				MgmtOnly: row.MgmtOnly, Description: row.Description,
			},
		})
	}
	return templates, nil
}

func (repository *DeviceTypeRepository) DeleteInterfaceTemplate(
	ctx context.Context,
	id shared.ID,
) error {
	result := repository.database(ctx).
		Where("id = ?", id.Int64()).
		Delete(&dcimrow.InterfaceTemplateRow{})
	if result.Error != nil {
		return shared.WrapError(
			shared.ErrorReasonInternal, "Could not delete InterfaceTemplate.", result.Error,
		)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("InterfaceTemplate", id)
	}
	return nil
}

func (repository *DeviceTypeRepository) Delete(
	ctx context.Context,
	deviceType *domaindcim.DeviceType,
) error {
	if deviceType == nil || !deviceType.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot delete an unpersisted DeviceType.",
		)
	}
	result := repository.database(ctx).
		Where("id = ?", deviceType.ID().Int64()).
		Delete(&dcimrow.DeviceTypeRow{})
	if result.Error != nil {
		return translateDeviceTypeMutationError("delete DeviceType", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("DeviceType", deviceType.ID())
	}
	return nil
}

func (repository *DeviceTypeRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.DeviceTypeListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(deviceTypeTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(
				deviceTypeTableAlias+".id IN ?",
				primitiveIDs(criteria.VisibleObjectIDs),
			)
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+deviceTypeTableAlias+".model) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTypeTableAlias+".slug) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTypeTableAlias+".part_number) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTypeTableAlias+".description) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceTypeTableAlias+".comments) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern, pattern, pattern,
		)
	}
	if len(criteria.ManufacturerIDs) > 0 {
		query = query.Where(
			deviceTypeTableAlias+".manufacturer_id IN ?", criteria.ManufacturerIDs,
		)
	}
	if len(criteria.ManufacturerSlugs) > 0 {
		query = query.Where(
			deviceTypeManufacturerTableAlias+".slug IN ?", criteria.ManufacturerSlugs,
		)
	}
	if len(criteria.Models) > 0 {
		query = query.Where(deviceTypeTableAlias+".model IN ?", criteria.Models)
	}
	if len(criteria.Slugs) > 0 {
		query = query.Where(deviceTypeTableAlias+".slug IN ?", criteria.Slugs)
	}
	return query
}

func (repository *DeviceTypeRepository) baseQuery(ctx context.Context) *gorm.DB {
	manufacturerTable := (dcimrow.ManufacturerRow{}).TableName()
	return repository.database(ctx).
		Table(deviceTypeTableExpression()).
		Joins(
			"JOIN " + manufacturerTable + " AS " + deviceTypeManufacturerTableAlias +
				" ON " + deviceTypeManufacturerTableAlias + ".id = " +
				deviceTypeTableAlias + ".manufacturer_id",
		)
}

func (repository *DeviceTypeRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type deviceTypeProjectionRow struct {
	dcimrow.DeviceTypeRow
	ManufacturerName       string `gorm:"column:manufacturer_name"`
	ManufacturerSlug       string `gorm:"column:manufacturer_slug"`
	DeviceCount            int64  `gorm:"column:device_count"`
	InterfaceTemplateCount int64  `gorm:"column:interface_template_count"`
}

func selectDeviceTypeProjection(query *gorm.DB) *gorm.DB {
	deviceTable := (dcimrow.DeviceRow{}).TableName()
	templateTable := (dcimrow.InterfaceTemplateRow{}).TableName()
	return query.Select(
		deviceTypeTableAlias + ".*, " +
			deviceTypeManufacturerTableAlias + ".name AS manufacturer_name, " +
			deviceTypeManufacturerTableAlias + ".slug AS manufacturer_slug, " +
			"(SELECT COUNT(*) FROM " + deviceTable + " WHERE " + deviceTable +
			".device_type_id = " + deviceTypeTableAlias + ".id) AS device_count, " +
			"(SELECT COUNT(*) FROM " + templateTable + " WHERE " + templateTable +
			".device_type_id = " + deviceTypeTableAlias +
			".id) AS interface_template_count",
	)
}

func deviceTypeTableExpression() string {
	return (dcimrow.DeviceTypeRow{}).TableName() + " AS " + deviceTypeTableAlias
}

func applyDeviceTypeOrdering(
	query *gorm.DB,
	ordering []applicationdcim.DeviceTypeSort,
) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{
				Table: deviceTypeTableAlias,
				Name:  deviceTypeSortColumn(requested.Field),
			},
			Desc: requested.Descending,
		})
		if requested.Field == applicationdcim.DeviceTypeSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: deviceTypeTableAlias, Name: "id"},
		})
	}
	return query
}

func deviceTypeSortColumn(field applicationdcim.DeviceTypeSortField) string {
	switch field {
	case applicationdcim.DeviceTypeSortID:
		return "id"
	case applicationdcim.DeviceTypeSortManufacturer:
		return "manufacturer_id"
	case applicationdcim.DeviceTypeSortModel:
		return "model"
	case applicationdcim.DeviceTypeSortSlug:
		return "slug"
	case applicationdcim.DeviceTypeSortUHeight:
		return "u_height"
	case applicationdcim.DeviceTypeSortCreated:
		return "created"
	case applicationdcim.DeviceTypeSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func deviceTypeToRow(deviceType domaindcim.DeviceType) dcimrow.DeviceTypeRow {
	state := deviceType.State()
	return dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time,
			LastUpdated: state.LastUpdated.Time,
		},
		ManufacturerID: state.Manufacturer.ID().Int64(),
		Model:          state.Model, Slug: state.Slug, PartNumber: state.PartNumber,
		UHeight:                stateUHeight(state.UHeight),
		ExcludeFromUtilization: state.ExcludeFromUtilization,
		IsFullDepth:            state.IsFullDepth, Airflow: nullableAirflowPointer(state.Airflow),
		Description: state.Description, Comments: state.Comments,
	}
}

func stateUHeight(value string) float64 {
	height, _ := domaindcim.ParseDeviceHeight(value)
	return height.Float64()
}

func nullableAirflowPointer(
	nullable domaindcim.NullableDeviceAirflow,
) *string {
	value, present := nullable.Get()
	if !present {
		return nil
	}
	text := value.String()
	return &text
}

func deviceTypeFromProjection(
	row deviceTypeProjectionRow,
) (*domaindcim.DeviceType, error) {
	if row.DeviceCount < 0 || row.InterfaceTemplateCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Persisted DeviceType contains an invalid annotation count.",
		)
	}
	manufacturer, err := domaindcim.NewManufacturerReference(
		shared.ID(row.ManufacturerID), row.ManufacturerName, row.ManufacturerSlug,
	)
	if err != nil {
		return nil, err
	}
	airflow := domaindcim.NullDeviceAirflow()
	if row.Airflow != nil {
		airflow = domaindcim.NonNullDeviceAirflow(
			domaindcim.DeviceAirflow(*row.Airflow),
		)
	}
	deviceType, err := domaindcim.RestoreDeviceType(domaindcim.DeviceTypeState{
		ID: shared.ID(row.ID), Manufacturer: manufacturer,
		Model: row.Model, Slug: row.Slug, PartNumber: row.PartNumber,
		UHeight:                strconv.FormatFloat(row.UHeight, 'f', -1, 64),
		ExcludeFromUtilization: row.ExcludeFromUtilization,
		IsFullDepth:            row.IsFullDepth, Airflow: airflow,
		Description: row.Description, Comments: row.Comments,
		Created:                shared.NewTimestamp(row.Created),
		LastUpdated:            shared.NewTimestamp(row.LastUpdated),
		DeviceCount:            uint64(row.DeviceCount),
		InterfaceTemplateCount: uint64(row.InterfaceTemplateCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted DeviceType state.",
			err,
		)
	}
	return deviceType, nil
}

func translateDeviceTypeReadError(
	id shared.ID,
	operation string,
	err error,
) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("DeviceType", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}

func translateDeviceTypeMutationError(operation string, err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_device_type_model"),
		strings.Contains(message, "go_dcim_device_types.manufacturer_id") &&
			strings.Contains(message, "go_dcim_device_types.model"):
		const description = "Device type with this Manufacturer and Model already exists."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{
				Field: "non_field_errors", Reason: "unique_together",
				Description: description,
			},
		)
	case strings.Contains(message, "uq_go_device_type_slug"),
		strings.Contains(message, "go_dcim_device_types.manufacturer_id") &&
			strings.Contains(message, "go_dcim_device_types.slug"):
		const description = "Device type with this Manufacturer and Slug already exists."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{
				Field: "non_field_errors", Reason: "unique_together",
				Description: description,
			},
		)
	case duplicateConstraint(err):
		return shared.Conflict(
			"A matching DeviceType already exists for this Manufacturer.", err,
		)
	case foreignKeyConstraint(err):
		return shared.WrapError(
			shared.ErrorReasonProtected,
			"The DeviceType is referenced by another object.",
			err,
		)
	default:
		return shared.WrapError(
			shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
		)
	}
}
