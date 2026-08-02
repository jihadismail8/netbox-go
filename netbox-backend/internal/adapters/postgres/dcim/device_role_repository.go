package dcim

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const deviceRoleTableAlias = "device_roles"

type DeviceRoleRepository struct{ db *gorm.DB }

var _ applicationdcim.DeviceRoleRepository = (*DeviceRoleRepository)(nil)

func NewDeviceRoleRepository(db *gorm.DB) *DeviceRoleRepository {
	if db == nil {
		panic("dcim device-role repository requires a database")
	}
	return &DeviceRoleRepository{db: db}
}

func (repository *DeviceRoleRepository) List(
	ctx context.Context,
	criteria applicationdcim.DeviceRoleListCriteria,
) (applicationdcim.DeviceRolePage, error) {
	base := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := base.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return applicationdcim.DeviceRolePage{}, translateDeviceRoleReadError(0, "count DeviceRoles", err)
	}
	if count < 0 {
		return applicationdcim.DeviceRolePage{}, shared.NewError(
			shared.ErrorReasonInternal, "DeviceRole count returned an invalid value.",
		)
	}
	var matchedIDs []int64
	if err := base.Select(deviceRoleTableAlias + ".id").Scan(&matchedIDs).Error; err != nil {
		return applicationdcim.DeviceRolePage{}, translateDeviceRoleReadError(0, "list DeviceRoles", err)
	}
	projected, err := repository.projectedHierarchy(ctx, false)
	if err != nil {
		return applicationdcim.DeviceRolePage{}, err
	}
	results := make([]*domaindcim.DeviceRole, 0, len(matchedIDs))
	for _, id := range matchedIDs {
		role := projected[id]
		if role == nil {
			return applicationdcim.DeviceRolePage{}, shared.NewError(
				shared.ErrorReasonInternal, "DeviceRole list projection omitted a matched object.",
			)
		}
		results = append(results, role)
	}
	sortProjectedDeviceRoles(results, projected, criteria)
	if !criteria.DeferPagination {
		start := min(int(criteria.Offset), len(results))
		end := min(start+int(criteria.Limit), len(results))
		results = results[start:end]
	}
	return applicationdcim.DeviceRolePage{Count: uint64(count), Results: results}, nil
}

func (repository *DeviceRoleRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.DeviceRole, error) {
	projected, err := repository.projectedHierarchy(ctx, false)
	if err != nil {
		return nil, err
	}
	role := projected[id.Int64()]
	if role == nil {
		return nil, shared.NotFound("DeviceRole", id)
	}
	return role, nil
}

func (repository *DeviceRoleRepository) ListHierarchyForUpdate(
	ctx context.Context,
) ([]*domaindcim.DeviceRole, error) {
	projected, err := repository.projectedHierarchy(ctx, true)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(projected))
	for id := range projected {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	roles := make([]*domaindcim.DeviceRole, 0, len(ids))
	for _, id := range ids {
		roles = append(roles, projected[id])
	}
	return roles, nil
}

func (repository *DeviceRoleRepository) Create(
	ctx context.Context,
	role *domaindcim.DeviceRole,
) error {
	if role == nil {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot persist a nil DeviceRole.")
	}
	if role.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot create an already persisted DeviceRole.")
	}
	row := deviceRoleToRow(*role)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateDeviceRoleMutationError("create DeviceRole", role, err)
	}
	return role.AssignID(shared.ID(row.ID))
}

func (repository *DeviceRoleRepository) Update(
	ctx context.Context,
	role *domaindcim.DeviceRole,
) error {
	if role == nil || !role.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot update an unpersisted DeviceRole.")
	}
	row := deviceRoleToRow(*role)
	result := repository.database(ctx).
		Model(&dcimrow.DeviceRoleRow{}).
		Where("id = ?", role.ID().Int64()).
		Select("parent_id", "name", "slug", "color", "vm_role", "description", "comments", "last_updated").
		Updates(&row)
	if result.Error != nil {
		return translateDeviceRoleMutationError("update DeviceRole", role, result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("DeviceRole", role.ID())
	}
	return nil
}

func (repository *DeviceRoleRepository) FindDeviceUsingRoles(
	ctx context.Context,
	ids []shared.ID,
) (*applicationdcim.DeviceRoleDependent, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	type dependentRow struct {
		ID       int64
		Name     *string
		AssetTag *string
	}
	var row dependentRow
	err := repository.database(ctx).
		Model(&dcimrow.DeviceRow{}).
		Select("id", "name", "asset_tag").
		Where("role_id IN ?", primitiveIDs(ids)).
		Order("id").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal, "Could not check DeviceRole dependencies.", err,
		)
	}
	display := "device " + shared.ID(row.ID).String()
	if row.Name != nil && strings.TrimSpace(*row.Name) != "" {
		display = *row.Name
		if row.AssetTag != nil && strings.TrimSpace(*row.AssetTag) != "" {
			display += " (" + *row.AssetTag + ")"
		}
	}
	return &applicationdcim.DeviceRoleDependent{ID: shared.ID(row.ID), Display: display}, nil
}

func (repository *DeviceRoleRepository) Delete(
	ctx context.Context,
	role *domaindcim.DeviceRole,
) error {
	if role == nil || !role.ID().IsValid() {
		return shared.NewError(shared.ErrorReasonInternal, "Cannot delete an unpersisted DeviceRole.")
	}
	result := repository.database(ctx).
		Where("id = ?", role.ID().Int64()).
		Delete(&dcimrow.DeviceRoleRow{})
	if result.Error != nil {
		return translateDeviceRoleMutationError("delete DeviceRole", role, result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("DeviceRole", role.ID())
	}
	return nil
}

func (repository *DeviceRoleRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.DeviceRoleListCriteria,
) *gorm.DB {
	query := repository.database(ctx).Table(deviceRoleTableExpression())
	if len(criteria.IDs) > 0 {
		query = query.Where(deviceRoleTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(deviceRoleTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs))
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+deviceRoleTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceRoleTableAlias+".slug) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceRoleTableAlias+".description) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+deviceRoleTableAlias+".comments) LIKE ? ESCAPE '\\')",
			pattern, pattern, pattern, pattern,
		)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(deviceRoleTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Slugs) > 0 {
		query = query.Where(deviceRoleTableAlias+".slug IN ?", criteria.Slugs)
	}
	return query
}

func (repository *DeviceRoleRepository) projectedHierarchy(
	ctx context.Context,
	forUpdate bool,
) (map[int64]*domaindcim.DeviceRole, error) {
	query := repository.database(ctx).Model(&dcimrow.DeviceRoleRow{}).Order("id")
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	var rows []dcimrow.DeviceRoleRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, translateDeviceRoleReadError(0, "load DeviceRole hierarchy", err)
	}
	directCounts, err := repository.directDeviceCounts(ctx)
	if err != nil {
		return nil, err
	}
	return restoreDeviceRoleHierarchy(rows, directCounts)
}

func (repository *DeviceRoleRepository) directDeviceCounts(ctx context.Context) (map[int64]uint64, error) {
	type countRow struct {
		RoleID int64
		Count  int64
	}
	var rows []countRow
	if err := repository.database(ctx).
		Model(&dcimrow.DeviceRow{}).
		Select("role_id, COUNT(*) AS count").
		Group("role_id").
		Scan(&rows).Error; err != nil {
		return nil, shared.WrapError(shared.ErrorReasonInternal, "Could not count Devices by role.", err)
	}
	counts := make(map[int64]uint64, len(rows))
	for _, row := range rows {
		if row.RoleID <= 0 || row.Count < 0 {
			return nil, shared.NewError(
				shared.ErrorReasonInternal, "DeviceRole projection contains an invalid Device count.",
			)
		}
		counts[row.RoleID] = uint64(row.Count)
	}
	return counts, nil
}

func restoreDeviceRoleHierarchy(
	rows []dcimrow.DeviceRoleRow,
	directCounts map[int64]uint64,
) (map[int64]*domaindcim.DeviceRole, error) {
	byID := make(map[int64]dcimrow.DeviceRoleRow, len(rows))
	children := make(map[int64][]int64)
	for _, row := range rows {
		if row.ID <= 0 {
			return nil, shared.NewError(shared.ErrorReasonInternal, "Persisted DeviceRole has an invalid ID.")
		}
		byID[row.ID] = row
		if row.ParentID != nil {
			children[*row.ParentID] = append(children[*row.ParentID], row.ID)
		}
	}
	for _, row := range rows {
		if row.ParentID != nil {
			if _, present := byID[*row.ParentID]; !present {
				return nil, shared.NewError(
					shared.ErrorReasonInternal, "Persisted DeviceRole hierarchy contains a missing parent.",
				)
			}
		}
	}
	type projection struct {
		depth uint32
		count uint64
	}
	projections := make(map[int64]projection, len(rows))
	depths := make(map[int64]uint32, len(rows))
	depthActive := make(map[int64]bool)
	depthComplete := make(map[int64]bool)
	var projectDepth func(int64) (uint32, error)
	projectDepth = func(id int64) (uint32, error) {
		if depthComplete[id] {
			return depths[id], nil
		}
		if depthActive[id] {
			return 0, shared.NewError(
				shared.ErrorReasonInternal, "Persisted DeviceRole hierarchy contains a cycle.",
			)
		}
		depthActive[id] = true
		row := byID[id]
		var depth uint32
		if row.ParentID != nil {
			parentDepth, err := projectDepth(*row.ParentID)
			if err != nil {
				return 0, err
			}
			if parentDepth == ^uint32(0) {
				return 0, shared.NewError(shared.ErrorReasonInternal, "DeviceRole depth exceeds uint32.")
			}
			depth = parentDepth + 1
		}
		depthActive[id] = false
		depthComplete[id] = true
		depths[id] = depth
		return depth, nil
	}
	counts := make(map[int64]uint64, len(rows))
	countActive := make(map[int64]bool)
	countComplete := make(map[int64]bool)
	var projectCount func(int64) (uint64, error)
	projectCount = func(id int64) (uint64, error) {
		if countComplete[id] {
			return counts[id], nil
		}
		if countActive[id] {
			return 0, shared.NewError(
				shared.ErrorReasonInternal, "Persisted DeviceRole hierarchy contains a cycle.",
			)
		}
		countActive[id] = true
		count := directCounts[id]
		for _, childID := range children[id] {
			childCount, err := projectCount(childID)
			if err != nil {
				return 0, err
			}
			if ^uint64(0)-count < childCount {
				return 0, shared.NewError(shared.ErrorReasonInternal, "DeviceRole Device count overflowed.")
			}
			count += childCount
		}
		countActive[id] = false
		countComplete[id] = true
		counts[id] = count
		return count, nil
	}
	for id := range byID {
		depth, err := projectDepth(id)
		if err != nil {
			return nil, err
		}
		count, err := projectCount(id)
		if err != nil {
			return nil, err
		}
		projections[id] = projection{depth: depth, count: count}
	}
	result := make(map[int64]*domaindcim.DeviceRole, len(rows))
	for id, row := range byID {
		parent := domaindcim.RootDeviceRoleParent()
		parentDisplay := ""
		if row.ParentID != nil {
			parent = domaindcim.NonRootDeviceRoleParent(shared.ID(*row.ParentID))
			parentDisplay = byID[*row.ParentID].Name
		}
		role, err := domaindcim.RestoreDeviceRole(domaindcim.DeviceRoleState{
			ID: shared.ID(row.ID), Parent: parent, ParentDisplay: parentDisplay,
			Name: row.Name, Slug: row.Slug, Color: row.Color, VMRole: row.VMRole,
			Description: row.Description, Comments: row.Comments,
			Created: shared.NewTimestamp(row.Created), LastUpdated: shared.NewTimestamp(row.LastUpdated),
			DeviceCount: projections[id].count, Depth: projections[id].depth,
		})
		if err != nil {
			return nil, shared.WrapError(
				shared.ErrorReasonInternal, "Could not restore persisted DeviceRole state.", err,
			)
		}
		result[id] = role
	}
	return result, nil
}

type deviceRoleOrderPart struct {
	name string
	id   shared.ID
}

func sortProjectedDeviceRoles(
	roles []*domaindcim.DeviceRole,
	hierarchy map[int64]*domaindcim.DeviceRole,
	criteria applicationdcim.DeviceRoleListCriteria,
) {
	if criteria.DefaultTreeOrder {
		paths := make(map[shared.ID][]deviceRoleOrderPart, len(roles))
		for _, role := range roles {
			paths[role.ID()] = deviceRoleTreePath(role, hierarchy)
		}
		sort.SliceStable(roles, func(left, right int) bool {
			return compareDeviceRolePaths(paths[roles[left].ID()], paths[roles[right].ID()]) < 0
		})
		return
	}
	sort.SliceStable(roles, func(left, right int) bool {
		for _, requested := range criteria.Ordering {
			comparison := compareDeviceRoleField(roles[left], roles[right], requested.Field)
			if comparison == 0 {
				continue
			}
			if requested.Descending {
				return comparison > 0
			}
			return comparison < 0
		}
		return roles[left].ID() < roles[right].ID()
	})
}

func deviceRoleTreePath(
	role *domaindcim.DeviceRole,
	hierarchy map[int64]*domaindcim.DeviceRole,
) []deviceRoleOrderPart {
	var reversed []deviceRoleOrderPart
	visited := make(map[shared.ID]struct{})
	for role != nil {
		if _, duplicate := visited[role.ID()]; duplicate {
			break
		}
		visited[role.ID()] = struct{}{}
		reversed = append(reversed, deviceRoleOrderPart{name: role.Name(), id: role.ID()})
		parentID, present := role.Parent().Get()
		if !present {
			break
		}
		role = hierarchy[parentID.Int64()]
	}
	path := make([]deviceRoleOrderPart, len(reversed))
	for index := range reversed {
		path[len(reversed)-1-index] = reversed[index]
	}
	return path
}

func compareDeviceRolePaths(left, right []deviceRoleOrderPart) int {
	for index := 0; index < min(len(left), len(right)); index++ {
		if left[index].name < right[index].name {
			return -1
		}
		if left[index].name > right[index].name {
			return 1
		}
		if left[index].id < right[index].id {
			return -1
		}
		if left[index].id > right[index].id {
			return 1
		}
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	return 0
}

func compareDeviceRoleField(
	left *domaindcim.DeviceRole,
	right *domaindcim.DeviceRole,
	field applicationdcim.DeviceRoleSortField,
) int {
	switch field {
	case applicationdcim.DeviceRoleSortID:
		return compareInt64(left.ID().Int64(), right.ID().Int64())
	case applicationdcim.DeviceRoleSortName:
		return strings.Compare(left.Name(), right.Name())
	case applicationdcim.DeviceRoleSortSlug:
		return strings.Compare(left.Slug().String(), right.Slug().String())
	case applicationdcim.DeviceRoleSortCreated:
		return compareInt64(left.Created().UnixNano(), right.Created().UnixNano())
	case applicationdcim.DeviceRoleSortLastUpdated:
		return compareInt64(left.LastUpdated().UnixNano(), right.LastUpdated().UnixNano())
	default:
		return 0
	}
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

func deviceRoleToRow(role domaindcim.DeviceRole) dcimrow.DeviceRoleRow {
	values := role.Values()
	var parentID *int64
	if id, present := values.Parent.Get(); present {
		value := id.Int64()
		parentID = &value
	}
	return dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: role.ID().Int64(), Created: role.Created().Time, LastUpdated: role.LastUpdated().Time,
		},
		ParentID: parentID, Name: values.Name, Slug: values.Slug, Color: values.Color,
		VMRole: values.VMRole, Description: values.Description, Comments: values.Comments,
	}
}

func deviceRoleTableExpression() string {
	return (dcimrow.DeviceRoleRow{}).TableName() + " AS " + deviceRoleTableAlias
}

func (repository *DeviceRoleRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

func translateDeviceRoleReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("DeviceRole", id)
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func translateDeviceRoleMutationError(
	operation string,
	role *domaindcim.DeviceRole,
	err error,
) error {
	if field, matched := deviceRoleUniqueField(err); matched {
		parent := domaindcim.RootDeviceRoleParent()
		if role != nil {
			parent = role.Parent()
		}
		message := deviceRoleUniqueMessage(parent, field)
		return shared.ConflictWithViolations(
			message,
			err,
			shared.FieldViolation{Field: "non_field_errors", Reason: "unique", Description: message},
		)
	}
	if duplicateConstraint(err) {
		return shared.Conflict("A matching device role already exists under this parent.", err)
	}
	if foreignKeyConstraint(err) {
		if strings.HasPrefix(operation, "delete") {
			return shared.WrapError(
				shared.ErrorReasonProtected, "The DeviceRole is referenced by another object.", err,
			)
		}
		return shared.NewValidationError(shared.FieldViolation{
			Field: "parent", Reason: "does_not_exist", Description: "The related object does not exist.",
		})
	}
	return shared.WrapError(shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err)
}

func deviceRoleUniqueField(err error) (string, bool) {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_device_role_parent_name"),
		strings.Contains(message, "go_dcim_device_roles.name"):
		return "name", true
	case strings.Contains(message, "uq_go_device_role_parent_slug"),
		strings.Contains(message, "go_dcim_device_roles.slug"):
		return "slug", true
	default:
		return "", false
	}
}

func deviceRoleUniqueMessage(parent domaindcim.DeviceRoleParent, field string) string {
	if parent.IsRoot() {
		return "A top-level device role with this " + field + " already exists."
	}
	label := map[string]string{"name": "Name", "slug": "Slug"}[field]
	return "Device role with this Parent and " + label + " already exists."
}
