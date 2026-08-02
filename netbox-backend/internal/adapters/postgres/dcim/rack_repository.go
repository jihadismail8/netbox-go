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
	rackTableAlias          = "racks"
	rackSiteTableAlias      = "rack_sites"
	rackTypeJoinTableAlias  = "rack_type_refs"
	rackReferencedRoleAlias = "rack_roles"
)

type RackRepository struct{ db *gorm.DB }

var _ applicationdcim.RackRepository = (*RackRepository)(nil)

func NewRackRepository(db *gorm.DB) *RackRepository {
	if db == nil {
		panic("dcim rack repository requires a database")
	}
	return &RackRepository{db: db}
}

func (repository *RackRepository) List(
	ctx context.Context,
	criteria applicationdcim.RackListCriteria,
) (applicationdcim.RackPage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.RackPage{}, translateRackReadError(0, "count Racks", err)
	}
	if count < 0 {
		return applicationdcim.RackPage{}, shared.NewError(
			shared.ErrorReasonInternal, "Rack count returned an invalid value.",
		)
	}
	query := applyRackOrdering(selectRackProjection(base), criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []rackProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.RackPage{}, translateRackReadError(0, "list Racks", err)
	}
	results := make([]*domaindcim.Rack, 0, len(rows))
	for _, row := range rows {
		rack, err := rackFromProjection(row)
		if err != nil {
			return applicationdcim.RackPage{}, err
		}
		results = append(results, rack)
	}
	return applicationdcim.RackPage{Count: uint64(count), Results: results}, nil
}

func (repository *RackRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Rack, error) {
	return repository.get(ctx, id, false)
}

func (repository *RackRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Rack, error) {
	return repository.get(ctx, id, true)
}

func (repository *RackRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.Rack, error) {
	query := repository.baseQuery(ctx).Where(rackTableAlias+".id = ?", id.Int64())
	query = selectRackProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: rackTableAlias},
		})
	}
	var row rackProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateRackReadError(id, "get Rack", err)
	}
	return rackFromProjection(row)
}

func (repository *RackRepository) Create(ctx context.Context, rack *domaindcim.Rack) error {
	if rack == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil Rack.")
	}
	if rack.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted Rack.")
	}
	row := rackToRow(*rack)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateRackMutationError("create Rack", err)
	}
	return rack.AssignID(shared.ID(row.ID))
}

func (repository *RackRepository) Update(ctx context.Context, rack *domaindcim.Rack) error {
	if rack == nil || !rack.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted Rack.")
	}
	row := rackToRow(*rack)
	result := repository.database(ctx).
		Model(&dcimrow.RackRow{}).
		Where("id = ?", rack.ID().Int64()).
		Select(
			"site_id", "name", "facility_id", "rack_type_id", "status", "role_id",
			"serial", "asset_tag", "form_factor", "width", "u_height", "starting_unit",
			"desc_units", "airflow", "description", "comments", "last_updated",
		).
		Updates(&row)
	if result.Error != nil {
		return translateRackMutationError("update Rack", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Rack", rack.ID())
	}
	return nil
}

func (repository *RackRepository) Delete(ctx context.Context, rack *domaindcim.Rack) error {
	if rack == nil || !rack.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted Rack.")
	}
	result := repository.database(ctx).
		Where("id = ?", rack.ID().Int64()).
		Delete(&dcimrow.RackRow{})
	if result.Error != nil {
		return translateRackMutationError("delete Rack", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Rack", rack.ID())
	}
	return nil
}

func (repository *RackRepository) MountedDevices(
	ctx context.Context,
	rackID shared.ID,
) ([]applicationdcim.RackDevicePlacement, error) {
	if !rackID.IsValid() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot inspect an invalid Rack.")
	}
	deviceTable := (dcimrow.DeviceRow{}).TableName()
	deviceTypeTable := (dcimrow.DeviceTypeRow{}).TableName()
	var rows []struct {
		ID       int64
		Position float64
		UHeight  float64
	}
	err := repository.database(ctx).
		Table(deviceTable+" AS mounted_devices").
		Select("mounted_devices.id, mounted_devices.position, mounted_types.u_height").
		Joins(
			"JOIN "+deviceTypeTable+" AS mounted_types ON "+
				"mounted_types.id = mounted_devices.device_type_id",
		).
		Where("mounted_devices.rack_id = ? AND mounted_devices.position IS NOT NULL", rackID.Int64()).
		Order("mounted_devices.position, mounted_devices.id").
		Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: "mounted_devices"},
		}).
		Find(&rows).Error
	if err != nil {
		return nil, shared.WrapError(shared.ErrorReasonInternal, "Could not inspect mounted Rack Devices.", err)
	}
	placements := make([]applicationdcim.RackDevicePlacement, 0, len(rows))
	for _, row := range rows {
		position := row.Position * 2
		height := row.UHeight * 2
		if row.ID <= 0 || math.Trunc(position) != position || position < 0 ||
			math.Trunc(height) != height || height < 0 || height > math.MaxUint16 {
			return nil, shared.NewError(
				shared.ErrorReasonInternal, "Persisted Device contains invalid Rack placement state.",
			)
		}
		placements = append(placements, applicationdcim.RackDevicePlacement{
			ID: shared.ID(row.ID), PositionHalfUnits: int32(position), HeightHalfUnits: uint16(height),
		})
	}
	return placements, nil
}

func (repository *RackRepository) PropagateSiteToDevices(
	ctx context.Context,
	rackID shared.ID,
	siteID shared.ID,
	now shared.Timestamp,
) ([]applicationdcim.RackSitePropagationChange, error) {
	if !rackID.IsValid() || !siteID.IsValid() || now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot propagate invalid Rack Site state.")
	}
	var rows []dcimrow.DeviceRow
	err := repository.database(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("rack_id = ?", rackID.Int64()).
		Order("id").
		Find(&rows).Error
	if err != nil {
		return nil, shared.WrapError(shared.ErrorReasonInternal, "Could not lock Rack Devices.", err)
	}
	changes := make([]applicationdcim.RackSitePropagationChange, 0, len(rows))
	for index := range rows {
		row := &rows[index]
		before, snapshotErr := rackDeviceSnapshot(*row)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		row.SiteID = siteID.Int64()
		row.LastUpdated = now.Time
		result := repository.database(ctx).
			Model(&dcimrow.DeviceRow{}).
			Where("id = ?", row.ID).
			Select("site_id", "last_updated").
			Updates(row)
		if result.Error != nil {
			return nil, translateRackDeviceSiteError(result.Error)
		}
		if result.RowsAffected == 0 {
			return nil, shared.NotFound("Device", shared.ID(row.ID))
		}
		after, snapshotErr := rackDeviceSnapshot(*row)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		changes = append(changes, applicationdcim.RackSitePropagationChange{
			ID: shared.ID(row.ID), Representation: rackDeviceRepresentation(*row),
			Before: before, After: after,
		})
	}
	return changes, nil
}

func (repository *RackRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.RackListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(rackTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(rackTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs))
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+rackTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTableAlias+".facility_id) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTableAlias+".serial) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTableAlias+".asset_tag) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTableAlias+".description) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTableAlias+".comments) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern, pattern, pattern, pattern,
		)
	}
	if len(criteria.SiteIDs) > 0 {
		query = query.Where(rackTableAlias+".site_id IN ?", criteria.SiteIDs)
	}
	if len(criteria.SiteSlugs) > 0 {
		query = query.Where(rackSiteTableAlias+".slug IN ?", criteria.SiteSlugs)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(rackTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Statuses) > 0 {
		statuses := make([]string, len(criteria.Statuses))
		for index, status := range criteria.Statuses {
			statuses[index] = status.String()
		}
		query = query.Where(rackTableAlias+".status IN ?", statuses)
	}
	if len(criteria.RoleIDs) > 0 {
		query = query.Where(rackTableAlias+".role_id IN ?", criteria.RoleIDs)
	}
	if len(criteria.RoleSlugs) > 0 {
		query = query.Where(rackReferencedRoleAlias+".slug IN ?", criteria.RoleSlugs)
	}
	if len(criteria.RackTypeIDs) > 0 {
		query = query.Where(rackTableAlias+".rack_type_id IN ?", criteria.RackTypeIDs)
	}
	if len(criteria.RackTypeSlugs) > 0 {
		query = query.Where(rackTypeJoinTableAlias+".slug IN ?", criteria.RackTypeSlugs)
	}
	return query
}

func (repository *RackRepository) baseQuery(ctx context.Context) *gorm.DB {
	siteTable := (dcimrow.SiteRow{}).TableName()
	rackTypeTable := (dcimrow.RackTypeRow{}).TableName()
	rackRoleTable := (dcimrow.RackRoleRow{}).TableName()
	return repository.database(ctx).
		Table(rackTableExpression()).
		Joins(
			"JOIN " + siteTable + " AS " + rackSiteTableAlias +
				" ON " + rackSiteTableAlias + ".id = " + rackTableAlias + ".site_id",
		).
		Joins(
			"LEFT JOIN " + rackTypeTable + " AS " + rackTypeJoinTableAlias +
				" ON " + rackTypeJoinTableAlias + ".id = " + rackTableAlias + ".rack_type_id",
		).
		Joins(
			"LEFT JOIN " + rackRoleTable + " AS " + rackReferencedRoleAlias +
				" ON " + rackReferencedRoleAlias + ".id = " + rackTableAlias + ".role_id",
		)
}

func (repository *RackRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type rackProjectionRow struct {
	dcimrow.RackRow
	SiteName             string `gorm:"column:site_name"`
	SiteSlug             string `gorm:"column:site_slug"`
	RackTypeModel        string `gorm:"column:rack_type_model"`
	RackTypeSlug         string `gorm:"column:rack_type_slug"`
	RackTypeFormFactor   string `gorm:"column:rack_type_form_factor"`
	RackTypeWidth        int64  `gorm:"column:rack_type_width"`
	RackTypeUHeight      int64  `gorm:"column:rack_type_u_height"`
	RackTypeStartingUnit int64  `gorm:"column:rack_type_starting_unit"`
	RackTypeDescUnits    bool   `gorm:"column:rack_type_desc_units"`
	RoleName             string `gorm:"column:role_name"`
	RoleSlug             string `gorm:"column:role_slug"`
	DeviceCount          int64  `gorm:"column:device_count"`
}

func selectRackProjection(query *gorm.DB) *gorm.DB {
	deviceTable := (dcimrow.DeviceRow{}).TableName()
	return query.Select(
		rackTableAlias + ".*, " +
			rackSiteTableAlias + ".name AS site_name, " +
			rackSiteTableAlias + ".slug AS site_slug, " +
			rackTypeJoinTableAlias + ".model AS rack_type_model, " +
			rackTypeJoinTableAlias + ".slug AS rack_type_slug, " +
			rackTypeJoinTableAlias + ".form_factor AS rack_type_form_factor, " +
			rackTypeJoinTableAlias + ".width AS rack_type_width, " +
			rackTypeJoinTableAlias + ".u_height AS rack_type_u_height, " +
			rackTypeJoinTableAlias + ".starting_unit AS rack_type_starting_unit, " +
			rackTypeJoinTableAlias + ".desc_units AS rack_type_desc_units, " +
			rackReferencedRoleAlias + ".name AS role_name, " +
			rackReferencedRoleAlias + ".slug AS role_slug, " +
			"(SELECT COUNT(*) FROM " + deviceTable + " WHERE " +
			deviceTable + ".rack_id = " + rackTableAlias + ".id) AS device_count",
	)
}

func rackTableExpression() string {
	return (dcimrow.RackRow{}).TableName() + " AS " + rackTableAlias
}

func applyRackOrdering(query *gorm.DB, ordering []applicationdcim.RackSort) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: rackTableAlias, Name: rackSortColumn(requested.Field)},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationdcim.RackSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: rackTableAlias, Name: "id"},
		})
	}
	return query
}

func rackSortColumn(field applicationdcim.RackSortField) string {
	switch field {
	case applicationdcim.RackSortID:
		return "id"
	case applicationdcim.RackSortSite:
		return "site_id"
	case applicationdcim.RackSortName:
		return "name"
	case applicationdcim.RackSortFacilityID:
		return "facility_id"
	case applicationdcim.RackSortStatus:
		return "status"
	case applicationdcim.RackSortUHeight:
		return "u_height"
	case applicationdcim.RackSortCreated:
		return "created"
	case applicationdcim.RackSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func rackToRow(rack domaindcim.Rack) dcimrow.RackRow {
	state := rack.State()
	return dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		SiteID: state.Site.ID().Int64(), Name: state.Name,
		FacilityID: nullableRackString(state.FacilityID),
		RackTypeID: nullableRackTypeReferenceID(state.RackType),
		Status:     state.Status, RoleID: nullableRackRoleReferenceID(state.Role),
		Serial: state.Serial, AssetTag: nullableRackString(state.AssetTag),
		FormFactor: nullableRackString(state.FormFactor), Width: int64(state.Width),
		UHeight: int64(state.UHeight), StartingUnit: int64(state.StartingUnit),
		DescUnits: state.DescUnits, Airflow: nullableRackString(state.Airflow),
		Description: state.Description, Comments: state.Comments,
	}
}

func rackFromProjection(row rackProjectionRow) (*domaindcim.Rack, error) {
	if row.DeviceCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Persisted Rack contains an invalid Device count.",
		)
	}
	site, err := domaindcim.NewSiteReference(shared.ID(row.SiteID), row.SiteName, row.SiteSlug)
	if err != nil {
		return nil, err
	}
	rackType := domaindcim.NullRackValue[domaindcim.RackTypeReference]()
	if row.RackTypeID != nil {
		formFactor, validFactor := domaindcim.ParseRackFormFactor(row.RackTypeFormFactor)
		width, validWidth := domaindcim.ParseRackWidth(uint32(row.RackTypeWidth))
		if !validFactor || !validWidth || row.RackTypeUHeight <= 0 || row.RackTypeStartingUnit <= 0 {
			return nil, shared.NewError(
				shared.ErrorReasonInternal, "Persisted RackType reference violates domain invariants.",
			)
		}
		reference, referenceErr := domaindcim.NewRackTypeReference(
			shared.ID(*row.RackTypeID), row.RackTypeModel, row.RackTypeSlug,
			domaindcim.RackPhysicalAttributes{
				FormFactor: formFactor, Width: width, UHeight: uint32(row.RackTypeUHeight),
				StartingUnit: uint32(row.RackTypeStartingUnit), DescUnits: row.RackTypeDescUnits,
			},
		)
		if referenceErr != nil {
			return nil, referenceErr
		}
		rackType = domaindcim.NonNullRackValue(reference)
	}
	role := domaindcim.NullRackValue[domaindcim.RackRoleReference]()
	if row.RoleID != nil {
		reference, referenceErr := domaindcim.NewRackRoleReference(
			shared.ID(*row.RoleID), row.RoleName, row.RoleSlug,
		)
		if referenceErr != nil {
			return nil, referenceErr
		}
		role = domaindcim.NonNullRackValue(reference)
	}
	rack, err := domaindcim.RestoreRack(domaindcim.RackState{
		ID: shared.ID(row.ID), Site: site, Name: row.Name,
		FacilityID: nullableRackStringFromPointer(row.FacilityID), RackType: rackType,
		Status: row.Status, Role: role, Serial: row.Serial,
		AssetTag:   nullableRackStringFromPointer(row.AssetTag),
		FormFactor: nullableRackStringFromPointer(row.FormFactor),
		Width:      uint32(row.Width), UHeight: uint32(row.UHeight),
		StartingUnit: uint32(row.StartingUnit), DescUnits: row.DescUnits,
		Airflow:     nullableRackStringFromPointer(row.Airflow),
		Description: row.Description, Comments: row.Comments,
		Created: shared.NewTimestamp(row.Created), LastUpdated: shared.NewTimestamp(row.LastUpdated),
		DeviceCount: uint64(row.DeviceCount),
	})
	if err != nil {
		return nil, shared.WrapError(shared.ErrorReasonInternal, "Could not restore persisted Rack state.", err)
	}
	return rack, nil
}

func nullableRackString(value domaindcim.RackNullable[string]) *string {
	text, present := value.Get()
	if !present {
		return nil
	}
	return &text
}

func nullableRackStringFromPointer(value *string) domaindcim.RackNullable[string] {
	if value == nil {
		return domaindcim.NullRackValue[string]()
	}
	return domaindcim.NonNullRackValue(*value)
}

func nullableRackTypeReferenceID(
	value domaindcim.RackNullable[domaindcim.RackTypeReference],
) *int64 {
	reference, present := value.Get()
	if !present {
		return nil
	}
	id := reference.ID().Int64()
	return &id
}

func nullableRackRoleReferenceID(
	value domaindcim.RackNullable[domaindcim.RackRoleReference],
) *int64 {
	reference, present := value.Get()
	if !present {
		return nil
	}
	id := reference.ID().Int64()
	return &id
}

func rackDeviceSnapshot(row dcimrow.DeviceRow) (domaindcim.DeviceSnapshot, error) {
	if row.ID <= 0 || row.DeviceTypeID <= 0 || row.RoleID <= 0 || row.SiteID <= 0 {
		return domaindcim.DeviceSnapshot{}, shared.NewError(
			shared.ErrorReasonInternal, "Persisted Device contains invalid Rack Site propagation state.",
		)
	}
	var rackID *shared.ID
	if row.RackID != nil {
		id := shared.ID(*row.RackID)
		rackID = &id
	}
	var position *string
	if row.Position != nil {
		text := strconv.FormatFloat(*row.Position, 'f', -1, 64)
		position = &text
	}
	return domaindcim.DeviceSnapshot{
		DeviceTypeID: shared.ID(row.DeviceTypeID), RoleID: shared.ID(row.RoleID),
		Name: copyStringPointer(row.Name), SiteID: shared.ID(row.SiteID), RackID: rackID,
		Position: position, Face: row.Face, Status: row.Status, Serial: row.Serial,
		AssetTag: copyStringPointer(row.AssetTag), Airflow: copyStringPointer(row.Airflow),
		Description: row.Description, Comments: row.Comments,
	}, nil
}

func rackDeviceRepresentation(row dcimrow.DeviceRow) string {
	if row.Name != nil && *row.Name != "" {
		if row.AssetTag != nil && *row.AssetTag != "" {
			return *row.Name + " (" + *row.AssetTag + ")"
		}
		return *row.Name
	}
	return "device " + strconv.FormatInt(row.ID, 10)
}

func translateRackReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("Rack", id)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func translateRackMutationError(operation string, err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_rack_asset_tag"),
		strings.Contains(message, "go_dcim_racks.asset_tag"):
		const description = "rack with this asset tag already exists."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{Field: "asset_tag", Reason: "unique", Description: description},
		)
	case duplicateConstraint(err):
		return shared.Conflict("A matching Rack already exists.", err)
	case foreignKeyConstraint(err):
		return shared.WrapError(shared.ErrorReasonProtected, "The Rack is referenced by another object.", err)
	default:
		return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
	}
}

func translateRackDeviceSiteError(err error) error {
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "uq_go_device_site_name_ci") ||
		(strings.Contains(message, "go_dcim_devices.site_id") &&
			strings.Contains(message, "go_dcim_devices.name")) ||
		duplicateConstraint(err) {
		return shared.Conflict("A device with this name already exists at this site.", err)
	}
	return shared.WrapError(shared.ErrorReasonInternal, "Could not propagate Rack Site to Device.", err)
}
