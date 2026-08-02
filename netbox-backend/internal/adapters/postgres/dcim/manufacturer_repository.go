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

const manufacturerTableAlias = "manufacturers"

type ManufacturerRepository struct{ db *gorm.DB }

var _ applicationdcim.ManufacturerRepository = (*ManufacturerRepository)(nil)

func NewManufacturerRepository(db *gorm.DB) *ManufacturerRepository {
	if db == nil {
		panic("dcim manufacturer repository requires a database")
	}
	return &ManufacturerRepository{db: db}
}

func (repository *ManufacturerRepository) List(
	ctx context.Context,
	criteria applicationdcim.ManufacturerListCriteria,
) (applicationdcim.ManufacturerPage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.ManufacturerPage{}, translateManufacturerReadError(0, "count Manufacturers", err)
	}
	if count < 0 {
		return applicationdcim.ManufacturerPage{}, shared.NewError(
			shared.ErrorReasonInternal, "Manufacturer count returned an invalid value.",
		)
	}
	query := applyManufacturerOrdering(selectManufacturerProjection(base), criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []manufacturerProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.ManufacturerPage{}, translateManufacturerReadError(0, "list Manufacturers", err)
	}
	results := make([]*domaindcim.Manufacturer, 0, len(rows))
	for _, row := range rows {
		manufacturer, err := manufacturerFromProjection(row)
		if err != nil {
			return applicationdcim.ManufacturerPage{}, err
		}
		results = append(results, manufacturer)
	}
	return applicationdcim.ManufacturerPage{Count: uint64(count), Results: results}, nil
}

func (repository *ManufacturerRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Manufacturer, error) {
	return repository.get(ctx, id, false)
}

func (repository *ManufacturerRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Manufacturer, error) {
	return repository.get(ctx, id, true)
}

func (repository *ManufacturerRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.Manufacturer, error) {
	query := repository.database(ctx).
		Table(manufacturerTableExpression()).
		Where(manufacturerTableAlias+".id = ?", id.Int64())
	query = selectManufacturerProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: manufacturerTableAlias},
		})
	}
	var row manufacturerProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateManufacturerReadError(id, "get Manufacturer", err)
	}
	return manufacturerFromProjection(row)
}

func (repository *ManufacturerRepository) Create(
	ctx context.Context,
	manufacturer *domaindcim.Manufacturer,
) error {
	if manufacturer == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil Manufacturer.")
	}
	if manufacturer.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted Manufacturer.")
	}
	row := manufacturerToRow(*manufacturer)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateManufacturerMutationError("create Manufacturer", err)
	}
	return manufacturer.AssignID(shared.ID(row.ID))
}

func (repository *ManufacturerRepository) Update(
	ctx context.Context,
	manufacturer *domaindcim.Manufacturer,
) error {
	if manufacturer == nil || !manufacturer.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted Manufacturer.")
	}
	row := manufacturerToRow(*manufacturer)
	result := repository.database(ctx).
		Model(&dcimrow.ManufacturerRow{}).
		Where("id = ?", manufacturer.ID().Int64()).
		Select("name", "slug", "description", "last_updated").
		Updates(&row)
	if result.Error != nil {
		return translateManufacturerMutationError("update Manufacturer", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Manufacturer", manufacturer.ID())
	}
	return nil
}

func (repository *ManufacturerRepository) Delete(
	ctx context.Context,
	manufacturer *domaindcim.Manufacturer,
) error {
	if manufacturer == nil || !manufacturer.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted Manufacturer.")
	}
	result := repository.database(ctx).
		Where("id = ?", manufacturer.ID().Int64()).
		Delete(&dcimrow.ManufacturerRow{})
	if result.Error != nil {
		return translateManufacturerMutationError("delete Manufacturer", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Manufacturer", manufacturer.ID())
	}
	return nil
}

func (repository *ManufacturerRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.ManufacturerListCriteria,
) *gorm.DB {
	query := repository.database(ctx).Table(manufacturerTableExpression())
	if len(criteria.IDs) > 0 {
		query = query.Where(manufacturerTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(manufacturerTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs))
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+manufacturerTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+manufacturerTableAlias+".slug) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+manufacturerTableAlias+".description) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern,
		)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(manufacturerTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Slugs) > 0 {
		query = query.Where(manufacturerTableAlias+".slug IN ?", criteria.Slugs)
	}
	return query
}

func (repository *ManufacturerRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type manufacturerProjectionRow struct {
	dcimrow.ManufacturerRow
	DeviceTypeCount int64 `gorm:"column:devicetype_count"`
}

func selectManufacturerProjection(query *gorm.DB) *gorm.DB {
	deviceTypeTable := (dcimrow.DeviceTypeRow{}).TableName()
	return query.Select(
		manufacturerTableAlias + ".*, " +
			"(SELECT COUNT(*) FROM " + deviceTypeTable + " WHERE " + deviceTypeTable +
			".manufacturer_id = " + manufacturerTableAlias + ".id) AS devicetype_count",
	)
}

func manufacturerTableExpression() string {
	return (dcimrow.ManufacturerRow{}).TableName() + " AS " + manufacturerTableAlias
}

func applyManufacturerOrdering(
	query *gorm.DB,
	ordering []applicationdcim.ManufacturerSort,
) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: manufacturerTableAlias, Name: manufacturerSortColumn(requested.Field)},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationdcim.ManufacturerSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: manufacturerTableAlias, Name: "id"},
		})
	}
	return query
}

func manufacturerSortColumn(field applicationdcim.ManufacturerSortField) string {
	switch field {
	case applicationdcim.ManufacturerSortID:
		return "id"
	case applicationdcim.ManufacturerSortName:
		return "name"
	case applicationdcim.ManufacturerSortSlug:
		return "slug"
	case applicationdcim.ManufacturerSortCreated:
		return "created"
	case applicationdcim.ManufacturerSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func manufacturerToRow(manufacturer domaindcim.Manufacturer) dcimrow.ManufacturerRow {
	state := manufacturer.State()
	return dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		Name: state.Name, Slug: state.Slug, Description: state.Description,
	}
}

func manufacturerFromProjection(row manufacturerProjectionRow) (*domaindcim.Manufacturer, error) {
	if row.DeviceTypeCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Persisted Manufacturer contains an invalid annotation count.",
		)
	}
	manufacturer, err := domaindcim.RestoreManufacturer(domaindcim.ManufacturerState{
		ID: shared.ID(row.ID), Name: row.Name, Slug: row.Slug, Description: row.Description,
		Created: shared.NewTimestamp(row.Created), LastUpdated: shared.NewTimestamp(row.LastUpdated),
		DeviceTypeCount: uint64(row.DeviceTypeCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not restore persisted Manufacturer state.", err,
		)
	}
	return manufacturer, nil
}

func translateManufacturerReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("Manufacturer", id)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func translateManufacturerMutationError(operation string, err error) error {
	if field, matched := manufacturerUniqueField(err); matched {
		message := "manufacturer with this " + field + " already exists."
		return shared.ConflictWithViolations(
			message,
			err,
			shared.FieldViolation{Field: field, Reason: "unique", Description: message},
		)
	}
	if duplicateConstraint(err) {
		return shared.Conflict("A Manufacturer with this name or slug already exists.", err)
	}
	if foreignKeyConstraint(err) {
		return shared.WrapError(
			shared.ErrorReasonProtected, "The Manufacturer is referenced by another object.", err,
		)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func manufacturerUniqueField(err error) (string, bool) {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_manufacturer_name"),
		strings.Contains(message, "go_dcim_manufacturers.name"):
		return "name", true
	case strings.Contains(message, "uq_go_manufacturer_slug"),
		strings.Contains(message, "go_dcim_manufacturers.slug"):
		return "slug", true
	default:
		return "", false
	}
}

func primitiveIDs(ids []shared.ID) []int64 {
	values := make([]int64, len(ids))
	for index, id := range ids {
		values[index] = id.Int64()
	}
	return values
}
