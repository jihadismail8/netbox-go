package ipam

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

const (
	prefixTableAlias = "prefixes"
	prefixVRFAlias   = "prefix_vrfs"
)

type PrefixRepository struct{ db *gorm.DB }

var _ applicationipam.PrefixRepository = (*PrefixRepository)(nil)

func NewPrefixRepository(db *gorm.DB) *PrefixRepository {
	if db == nil {
		panic("ipam Prefix repository requires a database")
	}
	return &PrefixRepository{db: db}
}

func (repository *PrefixRepository) List(
	ctx context.Context,
	criteria applicationipam.PrefixListCriteria,
) (applicationipam.PrefixPage, error) {
	query := repository.filteredQuery(ctx, criteria)
	query = selectPrefixProjection(query)
	var rows []prefixProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return applicationipam.PrefixPage{}, translatePrefixReadError(0, "list Prefixes", err)
	}
	hierarchy, err := repository.loadHierarchy(ctx)
	if err != nil {
		return applicationipam.PrefixPage{}, err
	}
	prefixes := make([]*domainipam.Prefix, 0, len(rows))
	for _, row := range rows {
		prefix, err := prefixFromProjection(row, hierarchy)
		if err != nil {
			return applicationipam.PrefixPage{}, err
		}
		if matchesPrefixCriteria(prefix, criteria) {
			prefixes = append(prefixes, prefix)
		}
	}
	sortPrefixes(prefixes, criteria.Ordering)
	count := uint64(len(prefixes))
	if !criteria.DeferPagination {
		start := min(int(criteria.Offset), len(prefixes))
		end := min(start+int(criteria.Limit), len(prefixes))
		prefixes = prefixes[start:end]
	}
	return applicationipam.PrefixPage{Count: count, Results: prefixes}, nil
}

func (repository *PrefixRepository) Get(ctx context.Context, id shared.ID) (*domainipam.Prefix, error) {
	return repository.get(ctx, id, false)
}

func (repository *PrefixRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domainipam.Prefix, error) {
	return repository.get(ctx, id, true)
}

func (repository *PrefixRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domainipam.Prefix, error) {
	query := repository.baseQuery(ctx).Where(prefixTableAlias+".id = ?", id.Int64())
	query = selectPrefixProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: prefixTableAlias},
		})
	}
	var row prefixProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translatePrefixReadError(id, "get Prefix", err)
	}
	hierarchy, err := repository.loadHierarchy(ctx)
	if err != nil {
		return nil, err
	}
	return prefixFromProjection(row, hierarchy)
}

func (repository *PrefixRepository) Create(ctx context.Context, prefix *domainipam.Prefix) error {
	if prefix == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil Prefix.")
	}
	if prefix.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted Prefix.")
	}
	row := prefixToRow(*prefix)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translatePrefixMutationError("create Prefix", err)
	}
	return prefix.AssignID(shared.ID(row.ID))
}

func (repository *PrefixRepository) Update(ctx context.Context, prefix *domainipam.Prefix) error {
	if prefix == nil || !prefix.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted Prefix.")
	}
	row := prefixToRow(*prefix)
	result := repository.database(ctx).
		Model(&ipamrow.PrefixRow{}).
		Where("id = ?", prefix.ID().Int64()).
		Select(prefixUpdateColumns()).
		Updates(&row)
	if result.Error != nil {
		return translatePrefixMutationError("update Prefix", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Prefix", prefix.ID())
	}
	return nil
}

func (repository *PrefixRepository) Delete(ctx context.Context, prefix *domainipam.Prefix) error {
	if prefix == nil || !prefix.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted Prefix.")
	}
	result := repository.database(ctx).
		Where("id = ?", prefix.ID().Int64()).
		Delete(&ipamrow.PrefixRow{})
	if result.Error != nil {
		return translatePrefixMutationError("delete Prefix", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Prefix", prefix.ID())
	}
	return nil
}

func (repository *PrefixRepository) LockUniqueness(
	ctx context.Context,
	vrf domainipam.NullableVRFReference,
	network domainipam.PrefixNetwork,
) error {
	db := repository.database(ctx)
	if db.Name() != "postgres" {
		return nil
	}
	scope := "global"
	if reference, present := vrf.Get(); present {
		scope = reference.ID().String()
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte("prefix:" + scope + ":" + network.String()))
	key := int64(hasher.Sum64())
	if err := db.Exec("SELECT pg_advisory_xact_lock(?)", key).Error; err != nil {
		return shared.WrapError(
			shared.ErrorReasonInternal, "Could not lock the Prefix uniqueness scope.", err,
		)
	}
	return nil
}

func (repository *PrefixRepository) FindDuplicate(
	ctx context.Context,
	vrf domainipam.NullableVRFReference,
	network domainipam.PrefixNetwork,
	exclude shared.ID,
) (*domainipam.Prefix, error) {
	query := repository.baseQuery(ctx).Where(prefixTableAlias+".prefix = ?", network.String())
	if reference, present := vrf.Get(); present {
		query = query.Where(prefixTableAlias+".vrf_id = ?", reference.ID().Int64())
	} else {
		query = query.Where(prefixTableAlias + ".vrf_id IS NULL")
	}
	if exclude.IsValid() {
		query = query.Where(prefixTableAlias+".id <> ?", exclude.Int64())
	}
	var row prefixProjectionRow
	if err := selectPrefixProjection(query).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, translatePrefixReadError(0, "find duplicate Prefix", err)
	}
	return prefixFromProjection(row, nil)
}

func (repository *PrefixRepository) filteredQuery(
	ctx context.Context,
	criteria applicationipam.PrefixListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(prefixTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			ids := make([]int64, len(criteria.VisibleObjectIDs))
			for index, id := range criteria.VisibleObjectIDs {
				ids[index] = id.Int64()
			}
			query = query.Where(prefixTableAlias+".id IN ?", ids)
		}
	}
	if len(criteria.VRFIDs) > 0 {
		query = query.Where(prefixTableAlias+".vrf_id IN ?", criteria.VRFIDs)
	}
	if len(criteria.VRFRDs) > 0 {
		query = query.Where(prefixVRFAlias+".rd IN ?", criteria.VRFRDs)
	}
	if criteria.PrefixesPresent {
		if len(criteria.Prefixes) == 0 {
			query = query.Where("1 = 0")
		} else {
			values := make([]string, len(criteria.Prefixes))
			for index, network := range criteria.Prefixes {
				values[index] = network.String()
			}
			query = query.Where(prefixTableAlias+".prefix IN ?", values)
		}
	}
	if criteria.Family != nil {
		switch *criteria.Family {
		case 4:
			query = query.Where(prefixTableAlias+".prefix NOT LIKE ?", "%:%")
		case 6:
			query = query.Where(prefixTableAlias+".prefix LIKE ?", "%:%")
		default:
			query = query.Where("1 = 0")
		}
	}
	if len(criteria.Statuses) > 0 {
		values := make([]string, len(criteria.Statuses))
		for index, status := range criteria.Statuses {
			values[index] = status.String()
		}
		query = query.Where(prefixTableAlias+".status IN ?", values)
	}
	return query
}

func (repository *PrefixRepository) baseQuery(ctx context.Context) *gorm.DB {
	vrfTable := (ipamrow.VRFRow{}).TableName()
	return repository.database(ctx).
		Table(prefixTableExpression()).
		Joins("LEFT JOIN " + vrfTable + " AS " + prefixVRFAlias +
			" ON " + prefixVRFAlias + ".id = " + prefixTableAlias + ".vrf_id")
}

func (repository *PrefixRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type prefixProjectionRow struct {
	ipamrow.PrefixRow
	VRFName          string  `gorm:"column:vrf_name"`
	VRFRD            *string `gorm:"column:vrf_rd"`
	VRFEnforceUnique bool    `gorm:"column:vrf_enforce_unique"`
}

type prefixHierarchyRow struct {
	ID     int64
	Prefix string
	VRFID  *int64
}

func selectPrefixProjection(query *gorm.DB) *gorm.DB {
	return query.Select(
		prefixTableAlias + ".*, " +
			prefixVRFAlias + ".name AS vrf_name, " +
			prefixVRFAlias + ".rd AS vrf_rd, " +
			prefixVRFAlias + ".enforce_unique AS vrf_enforce_unique",
	)
}

func prefixTableExpression() string {
	return (ipamrow.PrefixRow{}).TableName() + " AS " + prefixTableAlias
}

func (repository *PrefixRepository) loadHierarchy(ctx context.Context) ([]prefixHierarchyRow, error) {
	var rows []prefixHierarchyRow
	if err := repository.database(ctx).
		Model(&ipamrow.PrefixRow{}).
		Select("id", "prefix", "vrf_id").
		Find(&rows).Error; err != nil {
		return nil, translatePrefixReadError(0, "load Prefix hierarchy", err)
	}
	return rows, nil
}

func prefixFromProjection(
	row prefixProjectionRow,
	hierarchy []prefixHierarchyRow,
) (*domainipam.Prefix, error) {
	vrf, err := nullablePrefixVRF(row)
	if err != nil {
		return nil, err
	}
	network, err := domainipam.ParsePrefixNetwork(row.Prefix)
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Persisted Prefix contains an invalid network.", err,
		)
	}
	var children uint64
	var depth uint32
	for _, other := range hierarchy {
		if other.ID == row.ID || !samePrefixVRFID(row.VRFID, other.VRFID) {
			continue
		}
		candidate, parseErr := domainipam.ParsePrefixNetwork(other.Prefix)
		if parseErr != nil {
			return nil, shared.WrapError(
				shared.ErrorReasonInternal, "Persisted Prefix hierarchy contains an invalid network.", parseErr,
			)
		}
		if network.Contains(candidate, false) {
			children++
		}
		if candidate.Contains(network, false) {
			depth++
		}
	}
	prefix, err := domainipam.RestorePrefix(domainipam.PrefixState{
		ID: shared.ID(row.ID), Prefix: row.Prefix, VRF: vrf, Status: row.Status,
		IsPool: row.IsPool, MarkUtilized: row.MarkUtilized,
		Description: row.Description, Comments: row.Comments,
		Created: shared.NewTimestamp(row.Created), LastUpdated: shared.NewTimestamp(row.LastUpdated),
		Children: children, Depth: depth,
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not restore persisted Prefix state.", err,
		)
	}
	return prefix, nil
}

func nullablePrefixVRF(row prefixProjectionRow) (domainipam.NullableVRFReference, error) {
	if row.VRFID == nil {
		return domainipam.NullVRFReference(), nil
	}
	rd, err := nullableRouteDistinguisher(row.VRFRD)
	if err != nil {
		return domainipam.NullableVRFReference{}, err
	}
	reference, err := domainipam.NewVRFReference(
		shared.ID(*row.VRFID), row.VRFName, rd, row.VRFEnforceUnique,
	)
	if err != nil {
		return domainipam.NullableVRFReference{}, shared.WrapError(
			shared.ErrorReasonInternal, "Persisted Prefix references an invalid VRF.", err,
		)
	}
	return domainipam.NonNullVRFReference(reference), nil
}

func samePrefixVRFID(left, right *int64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func matchesPrefixCriteria(prefix *domainipam.Prefix, criteria applicationipam.PrefixListCriteria) bool {
	if prefix == nil {
		return false
	}
	if criteria.Query != "" && !matchesPrefixSearch(prefix, criteria.Query) {
		return false
	}
	if criteria.Within != nil {
		if !criteria.Within.Valid || !criteria.Within.Network.Contains(prefix.Network(), false) {
			return false
		}
	}
	if criteria.WithinInclude != nil {
		if !criteria.WithinInclude.Valid || !criteria.WithinInclude.Network.Contains(prefix.Network(), true) {
			return false
		}
	}
	if criteria.Contains != nil {
		if !criteria.Contains.Valid || !prefix.Network().Contains(
			criteria.Contains.Network, criteria.Contains.ExplicitMask,
		) {
			return false
		}
	}
	return true
}

func matchesPrefixSearch(prefix *domainipam.Prefix, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}
	lower := strings.ToLower(query)
	if strings.Contains(strings.ToLower(prefix.Description()), lower) ||
		strings.Contains(strings.ToLower(prefix.Network().String()), lower) {
		return true
	}
	queryNetwork, _, err := domainipam.ParsePrefixFilter(query)
	return err == nil && prefix.Network().Contains(queryNetwork, true)
}

func sortPrefixes(prefixes []*domainipam.Prefix, ordering []applicationipam.PrefixSort) {
	if len(ordering) == 0 {
		ordering = []applicationipam.PrefixSort{
			{Field: applicationipam.PrefixSortVRF},
			{Field: applicationipam.PrefixSortPrefix},
		}
	}
	hasID := false
	for _, requested := range ordering {
		if requested.Field == applicationipam.PrefixSortID {
			hasID = true
		}
	}
	if !hasID {
		ordering = append(ordering, applicationipam.PrefixSort{Field: applicationipam.PrefixSortID})
	}
	sort.SliceStable(prefixes, func(leftIndex, rightIndex int) bool {
		left, right := prefixes[leftIndex], prefixes[rightIndex]
		for _, requested := range ordering {
			compared := comparePrefixField(left, right, requested.Field)
			if compared == 0 {
				continue
			}
			if requested.Descending {
				return compared > 0
			}
			return compared < 0
		}
		return false
	})
}

func comparePrefixField(left, right *domainipam.Prefix, field applicationipam.PrefixSortField) int {
	switch field {
	case applicationipam.PrefixSortID:
		return compareInt64(left.ID().Int64(), right.ID().Int64())
	case applicationipam.PrefixSortVRF:
		return comparePrefixVRF(left.VRF(), right.VRF())
	case applicationipam.PrefixSortPrefix:
		return left.Network().Compare(right.Network())
	case applicationipam.PrefixSortStatus:
		return strings.Compare(left.Status().String(), right.Status().String())
	case applicationipam.PrefixSortCreated:
		return left.Created().Compare(right.Created().Time)
	case applicationipam.PrefixSortLastUpdated:
		return left.LastUpdated().Compare(right.LastUpdated().Time)
	default:
		return 0
	}
}

func comparePrefixVRF(left, right domainipam.NullableVRFReference) int {
	leftReference, leftPresent := left.Get()
	rightReference, rightPresent := right.Get()
	if !leftPresent && !rightPresent {
		return 0
	}
	if !leftPresent {
		return -1
	}
	if !rightPresent {
		return 1
	}
	return compareInt64(leftReference.ID().Int64(), rightReference.ID().Int64())
}

func compareInt64(left, right int64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func prefixToRow(prefix domainipam.Prefix) ipamrow.PrefixRow {
	state := prefix.State()
	var vrfID *int64
	if reference, present := state.VRF.Get(); present {
		value := reference.ID().Int64()
		vrfID = &value
	}
	return ipamrow.PrefixRow{
		RowMetadata: ipamrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time, LastUpdated: state.LastUpdated.Time,
		},
		Prefix: state.Prefix, VRFID: vrfID, Status: state.Status,
		IsPool: state.IsPool, MarkUtilized: state.MarkUtilized,
		Description: state.Description, Comments: state.Comments,
	}
}

func prefixUpdateColumns() []string {
	return []string{
		"prefix", "vrf_id", "status", "is_pool", "mark_utilized",
		"description", "comments", "last_updated",
	}
}

func translatePrefixReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("Prefix", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}

func translatePrefixMutationError(operation string, err error) error {
	if foreignKeyConstraint(err) {
		return shared.WrapError(
			shared.ErrorReasonProtected, "The Prefix is referenced by another object.", err,
		)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}
