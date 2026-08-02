package dcim

import (
	"context"
	"errors"
	"fmt"
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
	rackTypeTableAlias             = "rack_types"
	rackTypeManufacturerTableAlias = "rack_type_manufacturers"
)

type RackTypeRepository struct{ db *gorm.DB }

var _ applicationdcim.RackTypeRepository = (*RackTypeRepository)(nil)

func NewRackTypeRepository(db *gorm.DB) *RackTypeRepository {
	if db == nil {
		panic("dcim rack type repository requires a database")
	}
	return &RackTypeRepository{db: db}
}

func (repository *RackTypeRepository) List(
	ctx context.Context,
	criteria applicationdcim.RackTypeListCriteria,
) (applicationdcim.RackTypePage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.RackTypePage{}, translateRackTypeReadError(0, "count RackTypes", err)
	}
	if count < 0 {
		return applicationdcim.RackTypePage{}, shared.NewError(
			shared.ErrorReasonInternal, "RackType count returned an invalid value.",
		)
	}
	query := applyRackTypeOrdering(selectRackTypeProjection(base), criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []rackTypeProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.RackTypePage{}, translateRackTypeReadError(0, "list RackTypes", err)
	}
	results := make([]*domaindcim.RackType, 0, len(rows))
	for _, row := range rows {
		rackType, err := rackTypeFromProjection(row)
		if err != nil {
			return applicationdcim.RackTypePage{}, err
		}
		results = append(results, rackType)
	}
	return applicationdcim.RackTypePage{Count: uint64(count), Results: results}, nil
}

func (repository *RackTypeRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.RackType, error) {
	return repository.get(ctx, id, false)
}

func (repository *RackTypeRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.RackType, error) {
	return repository.get(ctx, id, true)
}

func (repository *RackTypeRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.RackType, error) {
	query := repository.baseQuery(ctx).Where(rackTypeTableAlias+".id = ?", id.Int64())
	query = selectRackTypeProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: rackTypeTableAlias},
		})
	}
	var row rackTypeProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateRackTypeReadError(id, "get RackType", err)
	}
	return rackTypeFromProjection(row)
}

func (repository *RackTypeRepository) Create(
	ctx context.Context,
	rackType *domaindcim.RackType,
) error {
	if rackType == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil RackType.")
	}
	if rackType.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted RackType.")
	}
	row := rackTypeToRow(*rackType)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateRackTypeMutationError("create RackType", err)
	}
	return rackType.AssignID(shared.ID(row.ID))
}

func (repository *RackTypeRepository) Update(
	ctx context.Context,
	rackType *domaindcim.RackType,
) error {
	if rackType == nil || !rackType.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted RackType.")
	}
	row := rackTypeToRow(*rackType)
	result := repository.database(ctx).
		Model(&dcimrow.RackTypeRow{}).
		Where("id = ?", rackType.ID().Int64()).
		Select(
			"manufacturer_id", "model", "slug", "form_factor", "width", "u_height",
			"starting_unit", "desc_units", "description", "comments", "last_updated",
		).
		Updates(&row)
	if result.Error != nil {
		return translateRackTypeMutationError("update RackType", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("RackType", rackType.ID())
	}
	return nil
}

func (repository *RackTypeRepository) Delete(
	ctx context.Context,
	rackType *domaindcim.RackType,
) error {
	if rackType == nil || !rackType.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted RackType.")
	}
	result := repository.database(ctx).
		Where("id = ?", rackType.ID().Int64()).
		Delete(&dcimrow.RackTypeRow{})
	if result.Error != nil {
		return translateRackTypeMutationError("delete RackType", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("RackType", rackType.ID())
	}
	return nil
}

func (repository *RackTypeRepository) PropagateToRacks(
	ctx context.Context,
	rackTypeID shared.ID,
	attributes domaindcim.RackPhysicalAttributes,
	now shared.Timestamp,
) ([]applicationdcim.RackPropagationChange, error) {
	if !rackTypeID.IsValid() || now.IsZero() {
		return nil, shared.NewError(shared.ErrorReasonInternal, "Cannot propagate invalid RackType state.")
	}
	var rows []dcimrow.RackRow
	err := repository.database(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("rack_type_id = ?", rackTypeID.Int64()).
		Order("id").
		Find(&rows).Error
	if err != nil {
		return nil, shared.WrapError(shared.ErrorReasonInternal, "Could not lock RackType Racks.", err)
	}
	changes := make([]applicationdcim.RackPropagationChange, 0, len(rows))
	for index := range rows {
		row := &rows[index]
		before, snapshotErr := rackSnapshotFromRow(*row)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		factor := attributes.FormFactor.String()
		row.FormFactor = &factor
		row.Width = int64(attributes.Width.Uint32())
		row.UHeight = int64(attributes.UHeight)
		row.StartingUnit = int64(attributes.StartingUnit)
		row.DescUnits = attributes.DescUnits
		row.LastUpdated = now.Time

		result := repository.database(ctx).
			Model(&dcimrow.RackRow{}).
			Where("id = ?", row.ID).
			Select("form_factor", "width", "u_height", "starting_unit", "desc_units", "last_updated").
			Updates(row)
		if result.Error != nil {
			return nil, shared.WrapError(
				shared.ErrorReasonInternal, "Could not propagate RackType attributes to Rack.", result.Error,
			)
		}
		if result.RowsAffected == 0 {
			return nil, shared.NotFound("Rack", shared.ID(row.ID))
		}
		after, snapshotErr := rackSnapshotFromRow(*row)
		if snapshotErr != nil {
			return nil, snapshotErr
		}
		changes = append(changes, applicationdcim.RackPropagationChange{
			ID: shared.ID(row.ID), Representation: rackRepresentation(*row), Before: before, After: after,
		})
	}
	return changes, nil
}

func (repository *RackTypeRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.RackTypeListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(rackTypeTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(rackTypeTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs))
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+rackTypeTableAlias+".model) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTypeTableAlias+".description) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackTypeTableAlias+".comments) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern,
		)
	}
	if len(criteria.ManufacturerIDs) > 0 {
		query = query.Where(rackTypeTableAlias+".manufacturer_id IN ?", criteria.ManufacturerIDs)
	}
	if len(criteria.ManufacturerSlugs) > 0 {
		query = query.Where(rackTypeManufacturerTableAlias+".slug IN ?", criteria.ManufacturerSlugs)
	}
	if len(criteria.Models) > 0 {
		query = query.Where(rackTypeTableAlias+".model IN ?", criteria.Models)
	}
	if len(criteria.Slugs) > 0 {
		query = query.Where(rackTypeTableAlias+".slug IN ?", criteria.Slugs)
	}
	return query
}

func (repository *RackTypeRepository) baseQuery(ctx context.Context) *gorm.DB {
	manufacturerTable := (dcimrow.ManufacturerRow{}).TableName()
	return repository.database(ctx).
		Table(rackTypeTableExpression()).
		Joins(
			"JOIN " + manufacturerTable + " AS " + rackTypeManufacturerTableAlias +
				" ON " + rackTypeManufacturerTableAlias + ".id = " + rackTypeTableAlias + ".manufacturer_id",
		)
}

func (repository *RackTypeRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type rackTypeProjectionRow struct {
	dcimrow.RackTypeRow
	ManufacturerName string `gorm:"column:manufacturer_name"`
	ManufacturerSlug string `gorm:"column:manufacturer_slug"`
}

func selectRackTypeProjection(query *gorm.DB) *gorm.DB {
	return query.Select(
		rackTypeTableAlias + ".*, " +
			rackTypeManufacturerTableAlias + ".name AS manufacturer_name, " +
			rackTypeManufacturerTableAlias + ".slug AS manufacturer_slug",
	)
}

func rackTypeTableExpression() string {
	return (dcimrow.RackTypeRow{}).TableName() + " AS " + rackTypeTableAlias
}

func applyRackTypeOrdering(query *gorm.DB, ordering []applicationdcim.RackTypeSort) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: rackTypeTableAlias, Name: rackTypeSortColumn(requested.Field)},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationdcim.RackTypeSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: rackTypeTableAlias, Name: "id"},
		})
	}
	return query
}

func rackTypeSortColumn(field applicationdcim.RackTypeSortField) string {
	switch field {
	case applicationdcim.RackTypeSortID:
		return "id"
	case applicationdcim.RackTypeSortManufacturer:
		return "manufacturer_id"
	case applicationdcim.RackTypeSortModel:
		return "model"
	case applicationdcim.RackTypeSortSlug:
		return "slug"
	case applicationdcim.RackTypeSortUHeight:
		return "u_height"
	case applicationdcim.RackTypeSortCreated:
		return "created"
	case applicationdcim.RackTypeSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func rackTypeToRow(rackType domaindcim.RackType) dcimrow.RackTypeRow {
	state := rackType.State()
	return dcimrow.RackTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		ManufacturerID: state.Manufacturer.ID().Int64(), Model: state.Model, Slug: state.Slug,
		FormFactor: state.FormFactor, Width: int64(state.Width), UHeight: int64(state.UHeight),
		StartingUnit: int64(state.StartingUnit), DescUnits: state.DescUnits,
		Description: state.Description, Comments: state.Comments,
	}
}

func rackTypeFromProjection(row rackTypeProjectionRow) (*domaindcim.RackType, error) {
	manufacturer, err := domaindcim.NewManufacturerReference(
		shared.ID(row.ManufacturerID), row.ManufacturerName, row.ManufacturerSlug,
	)
	if err != nil {
		return nil, err
	}
	rackType, err := domaindcim.RestoreRackType(domaindcim.RackTypeState{
		ID: shared.ID(row.ID), Manufacturer: manufacturer, Model: row.Model, Slug: row.Slug,
		FormFactor: row.FormFactor, Width: uint32(row.Width), UHeight: uint32(row.UHeight),
		StartingUnit: uint32(row.StartingUnit), DescUnits: row.DescUnits,
		Description: row.Description, Comments: row.Comments,
		Created: shared.NewTimestamp(row.Created), LastUpdated: shared.NewTimestamp(row.LastUpdated),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not restore persisted RackType state.", err,
		)
	}
	return rackType, nil
}

func rackSnapshotFromRow(row dcimrow.RackRow) (domaindcim.RackSnapshot, error) {
	if row.ID <= 0 || row.SiteID <= 0 || row.Name == "" {
		return domaindcim.RackSnapshot{}, shared.NewError(
			shared.ErrorReasonInternal, "Persisted Rack contains invalid propagation state.",
		)
	}
	return domaindcim.RackSnapshot{
		SiteID: shared.ID(row.SiteID), Name: row.Name, FacilityID: copyStringPointer(row.FacilityID),
		RackTypeID: copySharedIDPointer(row.RackTypeID), Status: row.Status,
		RoleID: copySharedIDPointer(row.RoleID), Serial: row.Serial,
		AssetTag: copyStringPointer(row.AssetTag), FormFactor: copyStringPointer(row.FormFactor),
		Width: uint32(row.Width), UHeight: uint32(row.UHeight), StartingUnit: uint32(row.StartingUnit),
		DescUnits: row.DescUnits, Airflow: copyStringPointer(row.Airflow),
		Description: row.Description, Comments: row.Comments,
	}, nil
}

func rackRepresentation(row dcimrow.RackRow) string {
	if row.FacilityID != nil && strings.TrimSpace(*row.FacilityID) != "" {
		return row.Name + " (" + *row.FacilityID + ")"
	}
	return row.Name
}

func copyStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func copySharedIDPointer(value *int64) *shared.ID {
	if value == nil {
		return nil
	}
	cloned := shared.ID(*value)
	return &cloned
}

func translateRackTypeReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("RackType", id)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func translateRackTypeMutationError(operation string, err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_rack_type_slug"),
		strings.Contains(message, "go_dcim_rack_types.slug"):
		const description = "rack type with this slug already exists."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{Field: "slug", Reason: "unique", Description: description},
		)
	case strings.Contains(message, "uq_go_rack_type_model"),
		strings.Contains(message, "go_dcim_rack_types.manufacturer_id") &&
			strings.Contains(message, "go_dcim_rack_types.model"):
		const description = "Rack type with this Manufacturer and Model already exists."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{Field: "non_field_errors", Reason: "unique_together", Description: description},
		)
	case duplicateConstraint(err):
		return shared.Conflict("A matching RackType already exists.", err)
	case foreignKeyConstraint(err):
		return shared.WrapError(
			shared.ErrorReasonProtected, "The RackType is referenced by another object.", err,
		)
	default:
		return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
	}
}
