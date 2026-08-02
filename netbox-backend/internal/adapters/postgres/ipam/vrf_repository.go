// Package ipam persists typed IPAM aggregates in the existing Go-owned
// first-profile tables.
package ipam

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

const vrfTableAlias = "vrfs"

type VRFRepository struct {
	db *gorm.DB
}

var _ applicationipam.VRFRepository = (*VRFRepository)(nil)

func NewVRFRepository(db *gorm.DB) *VRFRepository {
	if db == nil {
		panic("ipam VRF repository requires a database")
	}
	return &VRFRepository{db: db}
}

func (repository *VRFRepository) List(
	ctx context.Context,
	criteria applicationipam.VRFListCriteria,
) (applicationipam.VRFPage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationipam.VRFPage{}, translateVRFReadError(0, "count VRFs", err)
	}
	if count < 0 {
		return applicationipam.VRFPage{}, shared.NewError(
			shared.ErrorReasonInternal,
			"VRF count returned an invalid value.",
		)
	}

	query := selectVRFProjection(base)
	query = applyVRFOrdering(query, criteria.Ordering)
	if !criteria.DeferPagination && criteria.Offset > 0 {
		query = query.Offset(int(criteria.Offset))
	}
	if !criteria.DeferPagination && criteria.Limit > 0 {
		query = query.Limit(int(criteria.Limit))
	}

	var rows []vrfProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationipam.VRFPage{}, translateVRFReadError(0, "list VRFs", err)
	}
	vrfs := make([]*domainipam.VRF, 0, len(rows))
	for _, row := range rows {
		vrf, err := vrfFromProjection(row)
		if err != nil {
			return applicationipam.VRFPage{}, err
		}
		vrfs = append(vrfs, vrf)
	}
	return applicationipam.VRFPage{Count: uint64(count), Results: vrfs}, nil
}

func (repository *VRFRepository) Get(ctx context.Context, id shared.ID) (*domainipam.VRF, error) {
	return repository.get(ctx, id, false)
}

func (repository *VRFRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domainipam.VRF, error) {
	return repository.get(ctx, id, true)
}

func (repository *VRFRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domainipam.VRF, error) {
	query := repository.database(ctx).
		Table(vrfTableExpression()).
		Where(vrfTableAlias+".id = ?", id.Int64())
	query = selectVRFProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: vrfTableAlias},
		})
	}

	var row vrfProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateVRFReadError(id, "get VRF", err)
	}
	return vrfFromProjection(row)
}

func (repository *VRFRepository) Create(ctx context.Context, vrf *domainipam.VRF) error {
	if vrf == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil VRF.")
	}
	if vrf.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted VRF.")
	}

	row := vrfToRow(*vrf)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateVRFMutationError("create VRF", err)
	}
	return vrf.AssignID(shared.ID(row.ID))
}

func (repository *VRFRepository) Update(ctx context.Context, vrf *domainipam.VRF) error {
	if vrf == nil || !vrf.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted VRF.")
	}

	row := vrfToRow(*vrf)
	result := repository.database(ctx).
		Model(&ipamrow.VRFRow{}).
		Where("id = ?", vrf.ID().Int64()).
		Select(vrfUpdateColumns()).
		Updates(&row)
	if result.Error != nil {
		return translateVRFMutationError("update VRF", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("VRF", vrf.ID())
	}
	return nil
}

func (repository *VRFRepository) Delete(ctx context.Context, vrf *domainipam.VRF) error {
	if vrf == nil || !vrf.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted VRF.")
	}

	result := repository.database(ctx).
		Where("id = ?", vrf.ID().Int64()).
		Delete(&ipamrow.VRFRow{})
	if result.Error != nil {
		return translateVRFMutationError("delete VRF", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("VRF", vrf.ID())
	}
	return nil
}

func (repository *VRFRepository) filteredQuery(
	ctx context.Context,
	criteria applicationipam.VRFListCriteria,
) *gorm.DB {
	query := repository.database(ctx).Table(vrfTableExpression())
	if len(criteria.IDs) > 0 {
		query = query.Where(vrfTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			ids := make([]int64, len(criteria.VisibleObjectIDs))
			for index, id := range criteria.VisibleObjectIDs {
				ids[index] = id.Int64()
			}
			query = query.Where(vrfTableAlias+".id IN ?", ids)
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+vrfTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+vrfTableAlias+".rd) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+vrfTableAlias+".description) LIKE ? ESCAPE '\\')",
			pattern,
			pattern,
			pattern,
		)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(vrfTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.RDs) > 0 {
		rds := make([]string, len(criteria.RDs))
		for index, rd := range criteria.RDs {
			rds[index] = rd.String()
		}
		query = query.Where(vrfTableAlias+".rd IN ?", rds)
	}
	if criteria.EnforceUnique != nil {
		query = query.Where(vrfTableAlias+".enforce_unique = ?", *criteria.EnforceUnique)
	}
	return query
}

func (repository *VRFRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type vrfProjectionRow struct {
	ipamrow.VRFRow
	IPAddressCount int64 `gorm:"column:ipaddress_count"`
	PrefixCount    int64 `gorm:"column:prefix_count"`
}

func selectVRFProjection(query *gorm.DB) *gorm.DB {
	ipAddressTable := (ipamrow.IPAddressRow{}).TableName()
	prefixTable := (ipamrow.PrefixRow{}).TableName()
	return query.Select(
		vrfTableAlias + ".*, " +
			"(SELECT COUNT(*) FROM " + ipAddressTable + " WHERE " + ipAddressTable + ".vrf_id = " + vrfTableAlias + ".id) AS ipaddress_count, " +
			"(SELECT COUNT(*) FROM " + prefixTable + " WHERE " + prefixTable + ".vrf_id = " + vrfTableAlias + ".id) AS prefix_count",
	)
}

func vrfTableExpression() string {
	return (ipamrow.VRFRow{}).TableName() + " AS " + vrfTableAlias
}

func applyVRFOrdering(query *gorm.DB, ordering []applicationipam.VRFSort) *gorm.DB {
	if len(ordering) == 0 {
		ordering = []applicationipam.VRFSort{
			{Field: applicationipam.VRFSortName},
			{Field: applicationipam.VRFSortRD},
		}
	}
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: vrfTableAlias, Name: vrfSortColumn(requested.Field)},
			Desc:   requested.Descending,
		})
		if requested.Field == applicationipam.VRFSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: vrfTableAlias, Name: "id"},
		})
	}
	return query
}

func vrfSortColumn(field applicationipam.VRFSortField) string {
	switch field {
	case applicationipam.VRFSortID:
		return "id"
	case applicationipam.VRFSortName:
		return "name"
	case applicationipam.VRFSortRD:
		return "rd"
	case applicationipam.VRFSortCreated:
		return "created"
	case applicationipam.VRFSortLastUpdated:
		return "last_updated"
	default:
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

func vrfToRow(vrf domainipam.VRF) ipamrow.VRFRow {
	state := vrf.State()
	return ipamrow.VRFRow{
		RowMetadata: ipamrow.RowMetadata{
			ID:          state.ID.Int64(),
			Created:     state.Created.Time,
			LastUpdated: state.LastUpdated.Time,
		},
		Name:          state.Name,
		RD:            routeDistinguisherPointer(state.RD),
		EnforceUnique: state.EnforceUnique,
		Description:   state.Description,
		Comments:      state.Comments,
	}
}

func routeDistinguisherPointer(nullable domainipam.NullableRouteDistinguisher) *string {
	rd, present := nullable.Get()
	if !present {
		return nil
	}
	value := rd.String()
	return &value
}

func nullableRouteDistinguisher(value *string) (domainipam.NullableRouteDistinguisher, error) {
	if value == nil {
		return domainipam.NullRouteDistinguisher(), nil
	}
	rd, err := domainipam.ParseRouteDistinguisher(*value)
	if err != nil {
		return domainipam.NullableRouteDistinguisher{}, err
	}
	return domainipam.NonNullRouteDistinguisher(rd), nil
}

func vrfFromProjection(row vrfProjectionRow) (*domainipam.VRF, error) {
	if row.IPAddressCount < 0 || row.PrefixCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Persisted VRF contains an invalid annotation count.",
		)
	}
	rd, err := nullableRouteDistinguisher(row.RD)
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted VRF contains an invalid route distinguisher.",
			err,
		)
	}
	vrf, err := domainipam.RestoreVRF(domainipam.VRFState{
		ID:             shared.ID(row.ID),
		Name:           row.Name,
		RD:             rd,
		EnforceUnique:  row.EnforceUnique,
		Description:    row.Description,
		Comments:       row.Comments,
		Created:        shared.NewTimestamp(row.Created),
		LastUpdated:    shared.NewTimestamp(row.LastUpdated),
		IPAddressCount: uint64(row.IPAddressCount),
		PrefixCount:    uint64(row.PrefixCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted VRF state.",
			err,
		)
	}
	return vrf, nil
}

func vrfUpdateColumns() []string {
	return []string{"name", "rd", "enforce_unique", "description", "comments", "last_updated"}
}

func translateVRFReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("VRF", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal,
		fmt.Sprintf("Could not %s.", operation),
		err,
	)
}

func translateVRFMutationError(operation string, err error) error {
	if vrfRDConstraint(err) {
		const message = "VRF with this route distinguisher already exists."
		return shared.ConflictWithViolations(
			message,
			err,
			shared.FieldViolation{Field: "rd", Reason: "unique", Description: message},
		)
	}
	if duplicateConstraint(err) {
		return shared.Conflict("A VRF with this route distinguisher already exists.", err)
	}
	if foreignKeyConstraint(err) {
		return shared.WrapError(
			shared.ErrorReasonProtected,
			"The VRF is referenced by a Prefix or IP address.",
			err,
		)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal,
		fmt.Sprintf("Could not %s.", operation),
		err,
	)
}

func vrfRDConstraint(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "uq_go_vrf_rd") ||
		strings.Contains(message, "go_ipam_vrfs.rd")
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
	return strings.Contains(strings.ToLower(err.Error()), "foreign key constraint")
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
