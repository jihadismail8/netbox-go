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

const rackRoleTableAlias = "rack_roles"

type RackRoleRepository struct{ db *gorm.DB }

var _ applicationdcim.RackRoleRepository = (*RackRoleRepository)(nil)

func NewRackRoleRepository(db *gorm.DB) *RackRoleRepository {
	if db == nil {
		panic("dcim rack-role repository requires a database")
	}
	return &RackRoleRepository{db: db}
}

func (repository *RackRoleRepository) List(
	ctx context.Context,
	criteria applicationdcim.RackRoleListCriteria,
) (applicationdcim.RackRolePage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.RackRolePage{}, translateRackRoleReadError(0, "count RackRoles", err)
	}
	if count < 0 {
		return applicationdcim.RackRolePage{}, shared.NewError(
			shared.ErrorReasonInternal, "RackRole count returned an invalid value.",
		)
	}
	query := applyRackRoleOrdering(selectRackRoleProjection(base), criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}
	var rows []rackRoleProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.RackRolePage{}, translateRackRoleReadError(0, "list RackRoles", err)
	}
	results := make([]*domaindcim.RackRole, 0, len(rows))
	for _, row := range rows {
		role, err := rackRoleFromProjection(row)
		if err != nil {
			return applicationdcim.RackRolePage{}, err
		}
		results = append(results, role)
	}
	return applicationdcim.RackRolePage{Count: uint64(count), Results: results}, nil
}

func (repository *RackRoleRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.RackRole, error) {
	return repository.get(ctx, id, false)
}

func (repository *RackRoleRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.RackRole, error) {
	return repository.get(ctx, id, true)
}

func (repository *RackRoleRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.RackRole, error) {
	query := repository.database(ctx).
		Table(rackRoleTableExpression()).
		Where(rackRoleTableAlias+".id = ?", id.Int64())
	query = selectRackRoleProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: rackRoleTableAlias},
		})
	}
	var row rackRoleProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateRackRoleReadError(id, "get RackRole", err)
	}
	return rackRoleFromProjection(row)
}

func (repository *RackRoleRepository) Create(ctx context.Context, role *domaindcim.RackRole) error {
	if role == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil RackRole.")
	}
	if role.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted RackRole.")
	}
	row := rackRoleToRow(*role)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateRackRoleMutationError("create RackRole", err)
	}
	return role.AssignID(shared.ID(row.ID))
}

func (repository *RackRoleRepository) Update(ctx context.Context, role *domaindcim.RackRole) error {
	if role == nil || !role.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted RackRole.")
	}
	row := rackRoleToRow(*role)
	result := repository.database(ctx).
		Model(&dcimrow.RackRoleRow{}).
		Where("id = ?", role.ID().Int64()).
		Select("name", "slug", "color", "description", "last_updated").
		Updates(&row)
	if result.Error != nil {
		return translateRackRoleMutationError("update RackRole", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("RackRole", role.ID())
	}
	return nil
}

func (repository *RackRoleRepository) Delete(ctx context.Context, role *domaindcim.RackRole) error {
	if role == nil || !role.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted RackRole.")
	}
	result := repository.database(ctx).
		Where("id = ?", role.ID().Int64()).
		Delete(&dcimrow.RackRoleRow{})
	if result.Error != nil {
		return translateRackRoleMutationError("delete RackRole", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("RackRole", role.ID())
	}
	return nil
}

func (repository *RackRoleRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.RackRoleListCriteria,
) *gorm.DB {
	query := repository.database(ctx).Table(rackRoleTableExpression())
	if len(criteria.IDs) > 0 {
		query = query.Where(rackRoleTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(rackRoleTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs))
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+rackRoleTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackRoleTableAlias+".slug) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+rackRoleTableAlias+".description) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern,
		)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(rackRoleTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Slugs) > 0 {
		query = query.Where(rackRoleTableAlias+".slug IN ?", criteria.Slugs)
	}
	return query
}

func (repository *RackRoleRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type rackRoleProjectionRow struct {
	dcimrow.RackRoleRow
	RackCount int64 `gorm:"column:rack_count"`
}

func selectRackRoleProjection(query *gorm.DB) *gorm.DB {
	rackTable := (dcimrow.RackRow{}).TableName()
	return query.Select(
		rackRoleTableAlias + ".*, " +
			"(SELECT COUNT(*) FROM " + rackTable + " WHERE " + rackTable +
			".role_id = " + rackRoleTableAlias + ".id) AS rack_count",
	)
}

func rackRoleTableExpression() string {
	return (dcimrow.RackRoleRow{}).TableName() + " AS " + rackRoleTableAlias
}

func applyRackRoleOrdering(query *gorm.DB, ordering []applicationdcim.RackRoleSort) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: rackRoleTableAlias, Name: rackRoleSortColumn(requested.Field)},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationdcim.RackRoleSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: rackRoleTableAlias, Name: "id"},
		})
	}
	return query
}

func rackRoleSortColumn(field applicationdcim.RackRoleSortField) string {
	switch field {
	case applicationdcim.RackRoleSortID:
		return "id"
	case applicationdcim.RackRoleSortName:
		return "name"
	case applicationdcim.RackRoleSortSlug:
		return "slug"
	case applicationdcim.RackRoleSortCreated:
		return "created"
	case applicationdcim.RackRoleSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func rackRoleToRow(role domaindcim.RackRole) dcimrow.RackRoleRow {
	state := role.State()
	return dcimrow.RackRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		Name: state.Name, Slug: state.Slug, Color: state.Color, Description: state.Description,
	}
}

func rackRoleFromProjection(row rackRoleProjectionRow) (*domaindcim.RackRole, error) {
	if row.RackCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal, "Persisted RackRole contains an invalid annotation count.",
		)
	}
	role, err := domaindcim.RestoreRackRole(domaindcim.RackRoleState{
		ID: shared.ID(row.ID), Name: row.Name, Slug: row.Slug, Color: row.Color,
		Description: row.Description, Created: shared.NewTimestamp(row.Created),
		LastUpdated: shared.NewTimestamp(row.LastUpdated), RackCount: uint64(row.RackCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not restore persisted RackRole state.", err,
		)
	}
	return role, nil
}

func translateRackRoleReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("RackRole", id)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func translateRackRoleMutationError(operation string, err error) error {
	if field, matched := rackRoleUniqueField(err); matched {
		message := "rack role with this " + field + " already exists."
		return shared.ConflictWithViolations(
			message,
			err,
			shared.FieldViolation{Field: field, Reason: "unique", Description: message},
		)
	}
	if duplicateConstraint(err) {
		return shared.Conflict("A RackRole with this name or slug already exists.", err)
	}
	if foreignKeyConstraint(err) {
		return shared.WrapError(
			shared.ErrorReasonProtected, "The RackRole is referenced by another object.", err,
		)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func rackRoleUniqueField(err error) (string, bool) {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_rack_role_name"),
		strings.Contains(message, "go_dcim_rack_roles.name"):
		return "name", true
	case strings.Contains(message, "uq_go_rack_role_slug"),
		strings.Contains(message, "go_dcim_rack_roles.slug"):
		return "slug", true
	default:
		return "", false
	}
}
