// Package dcim persists typed DCIM aggregates in the existing Go-owned
// first-profile tables.
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

const siteTableAlias = "sites"

type SiteRepository struct {
	db *gorm.DB
}

var _ applicationdcim.SiteRepository = (*SiteRepository)(nil)

func NewSiteRepository(db *gorm.DB) *SiteRepository {
	if db == nil {
		panic("dcim site repository requires a database")
	}
	return &SiteRepository{db: db}
}

func (repository *SiteRepository) List(
	ctx context.Context,
	criteria applicationdcim.SiteListCriteria,
) (applicationdcim.SitePage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.SitePage{}, translateSiteReadError(
			shared.ID(0),
			"count Sites",
			err,
		)
	}
	if count < 0 {
		return applicationdcim.SitePage{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Site count returned an invalid value.",
		)
	}

	query := selectSiteProjection(base)
	query = applySiteOrdering(query, criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}

	var rows []siteProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationdcim.SitePage{}, translateSiteReadError(
			shared.ID(0),
			"list Sites",
			err,
		)
	}

	sites := make([]*domaindcim.Site, 0, len(rows))
	for _, row := range rows {
		site, err := siteFromProjection(row)
		if err != nil {
			return applicationdcim.SitePage{}, err
		}
		sites = append(sites, site)
	}

	return applicationdcim.SitePage{
		Count:   uint64(count),
		Results: sites,
	}, nil
}

func (repository *SiteRepository) Get(ctx context.Context, id shared.ID) (*domaindcim.Site, error) {
	return repository.get(ctx, id, false)
}

func (repository *SiteRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Site, error) {
	return repository.get(ctx, id, true)
}

func (repository *SiteRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.Site, error) {
	query := repository.database(ctx).
		Table(siteTableExpression()).
		Where(siteTableAlias+".id = ?", id.Int64())
	query = selectSiteProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: siteTableAlias},
		})
	}

	var row siteProjectionRow
	err := query.Take(&row).Error
	if err != nil {
		return nil, translateSiteReadError(id, "get Site", err)
	}
	return siteFromProjection(row)
}

func (repository *SiteRepository) Create(ctx context.Context, site *domaindcim.Site) error {
	if site == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil Site.")
	}
	if site.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted Site.")
	}

	row := siteToRow(*site)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateSiteMutationError("create Site", err)
	}
	if err := site.AssignID(shared.ID(row.ID)); err != nil {
		return err
	}
	return nil
}

func (repository *SiteRepository) Update(ctx context.Context, site *domaindcim.Site) error {
	if site == nil || !site.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted Site.")
	}

	row := siteToRow(*site)
	result := repository.database(ctx).
		Model(&dcimrow.SiteRow{}).
		Where("id = ?", site.ID().Int64()).
		Select(siteUpdateColumns()).
		Updates(&row)
	if result.Error != nil {
		return translateSiteMutationError("update Site", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Site", site.ID())
	}
	return nil
}

func (repository *SiteRepository) Delete(ctx context.Context, site *domaindcim.Site) error {
	if site == nil || !site.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted Site.")
	}

	result := repository.database(ctx).
		Where("id = ?", site.ID().Int64()).
		Delete(&dcimrow.SiteRow{})
	if result.Error != nil {
		return translateSiteMutationError("delete Site", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Site", site.ID())
	}
	return nil
}

func (repository *SiteRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.SiteListCriteria,
) *gorm.DB {
	query := repository.database(ctx).Table(siteTableExpression())
	if len(criteria.IDs) > 0 {
		query = query.Where(siteTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			ids := make([]int64, len(criteria.VisibleObjectIDs))
			for index, id := range criteria.VisibleObjectIDs {
				ids[index] = id.Int64()
			}
			query = query.Where(siteTableAlias+".id IN ?", ids)
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+siteTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+siteTableAlias+".facility) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+siteTableAlias+".description) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+siteTableAlias+".comments) LIKE ? ESCAPE '\\')",
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(siteTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Slugs) > 0 {
		query = query.Where(siteTableAlias+".slug IN ?", criteria.Slugs)
	}
	if len(criteria.Statuses) > 0 {
		statuses := make([]string, len(criteria.Statuses))
		for index, status := range criteria.Statuses {
			statuses[index] = status.String()
		}
		query = query.Where(siteTableAlias+".status IN ?", statuses)
	}
	return query
}

func (repository *SiteRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type siteProjectionRow struct {
	dcimrow.SiteRow
	DeviceCount int64 `gorm:"column:device_count"`
	PrefixCount int64 `gorm:"column:prefix_count"`
	RackCount   int64 `gorm:"column:rack_count"`
}

func selectSiteProjection(query *gorm.DB) *gorm.DB {
	deviceTable := (dcimrow.DeviceRow{}).TableName()
	rackTable := (dcimrow.RackRow{}).TableName()
	return query.Select(
		siteTableAlias + ".*, " +
			"(SELECT COUNT(*) FROM " + deviceTable + " WHERE " + deviceTable + ".site_id = " + siteTableAlias + ".id) AS device_count, " +
			"0 AS prefix_count, " +
			"(SELECT COUNT(*) FROM " + rackTable + " WHERE " + rackTable + ".site_id = " + siteTableAlias + ".id) AS rack_count",
	)
}

func siteTableExpression() string {
	return (dcimrow.SiteRow{}).TableName() + " AS " + siteTableAlias
}

func applySiteOrdering(query *gorm.DB, ordering []applicationdcim.SiteSort) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		column := siteSortColumn(requested.Field)
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: siteTableAlias, Name: column},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationdcim.SiteSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: siteTableAlias, Name: "id"},
		})
	}
	return query
}

func siteSortColumn(field applicationdcim.SiteSortField) string {
	switch field {
	case applicationdcim.SiteSortID:
		return "id"
	case applicationdcim.SiteSortName:
		return "name"
	case applicationdcim.SiteSortSlug:
		return "slug"
	case applicationdcim.SiteSortStatus:
		return "status"
	case applicationdcim.SiteSortCreated:
		return "created"
	case applicationdcim.SiteSortLastUpdated:
		return "last_updated"
	default:
		// The application validates ordering before it reaches this adapter.
		return "id"
	}
}

func containsPattern(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return "%" + value + "%"
}

func siteToRow(site domaindcim.Site) dcimrow.SiteRow {
	state := site.State()
	return dcimrow.SiteRow{
		RowMetadata: dcimrow.RowMetadata{
			ID:          state.ID.Int64(),
			Created:     state.Created.Time,
			LastUpdated: state.LastUpdated.Time,
		},
		Name:        state.Name,
		Slug:        state.Slug,
		Status:      state.Status,
		Facility:    state.Facility,
		Description: state.Description,
		Comments:    state.Comments,
	}
}

func siteFromProjection(row siteProjectionRow) (*domaindcim.Site, error) {
	if row.DeviceCount < 0 || row.PrefixCount < 0 || row.RackCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Persisted Site contains an invalid annotation count.",
		)
	}

	site, err := domaindcim.RestoreSite(domaindcim.SiteState{
		ID:          shared.ID(row.ID),
		Name:        row.Name,
		Slug:        row.Slug,
		Status:      row.Status,
		Facility:    row.Facility,
		Description: row.Description,
		Comments:    row.Comments,
		Created:     shared.NewTimestamp(row.Created),
		LastUpdated: shared.NewTimestamp(row.LastUpdated),
		DeviceCount: uint64(row.DeviceCount),
		PrefixCount: uint64(row.PrefixCount),
		RackCount:   uint64(row.RackCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted Site state.",
			err,
		)
	}
	return site, nil
}

func siteUpdateColumns() []string {
	return []string{
		"name",
		"slug",
		"status",
		"facility",
		"description",
		"comments",
		"last_updated",
	}
}

func translateSiteReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("Site", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal,
		fmt.Sprintf("Could not %s.", operation),
		err,
	)
}

func translateSiteMutationError(operation string, err error) error {
	if field, matched := siteUniqueField(err); matched {
		message := "site with this " + field + " already exists."
		return shared.ConflictWithViolations(
			message,
			err,
			shared.FieldViolation{
				Field:       field,
				Reason:      "unique",
				Description: message,
			},
		)
	}
	if duplicateConstraint(err) {
		// A translated driver error can discard the constraint name. Keep the
		// failure safe and conflict-typed, but do not guess which field failed.
		return shared.Conflict("A Site with this name or slug already exists.", err)
	}
	if foreignKeyConstraint(err) {
		return shared.WrapError(
			shared.ErrorReasonProtected,
			"The Site is referenced by another object.",
			err,
		)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal,
		fmt.Sprintf("Could not %s.", operation),
		err,
	)
}

func siteUniqueField(err error) (string, bool) {
	// pgx includes PostgreSQL's constraint name in PgError.Error(), while
	// SQLite identifies the constrained table column. Matching both keeps the
	// adapter independent of a concrete driver error type and preserves wrapped
	// errors.
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_site_name"),
		strings.Contains(message, "go_dcim_sites.name"):
		return "name", true
	case strings.Contains(message, "uq_go_site_slug"),
		strings.Contains(message, "go_dcim_sites.slug"):
		return "slug", true
	default:
		return "", false
	}
}

func duplicateConstraint(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) || sqlState(err) == "23505" {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "unique failed") ||
		strings.Contains(message, "duplicate key")
}

func foreignKeyConstraint(err error) bool {
	if errors.Is(err, gorm.ErrForeignKeyViolated) || sqlState(err) == "23503" {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "foreign key constraint")
}

func sqlState(err error) string {
	type sqlStateError interface {
		SQLState() string
	}
	var stateError sqlStateError
	if errors.As(err, &stateError) {
		return stateError.SQLState()
	}
	return ""
}
