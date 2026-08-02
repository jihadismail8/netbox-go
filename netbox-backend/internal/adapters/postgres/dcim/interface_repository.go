package dcim

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

const (
	interfaceTableAlias        = "typed_interfaces"
	interfaceDeviceAlias       = "interface_devices"
	interfaceDeviceTypeAlias   = "interface_device_types"
	interfaceManufacturerAlias = "interface_manufacturers"
)

type InterfaceRepository struct {
	db *gorm.DB
}

var _ applicationdcim.InterfaceRepository = (*InterfaceRepository)(nil)
var _ applicationdcim.InterfaceDeviceReader = (*InterfaceRepository)(nil)

func NewInterfaceRepository(db *gorm.DB) *InterfaceRepository {
	if db == nil {
		panic("postgres Interface repository requires a database")
	}
	return &InterfaceRepository{db: db}
}

func (repository *InterfaceRepository) List(
	ctx context.Context,
	criteria applicationdcim.InterfaceListCriteria,
) (applicationdcim.InterfacePage, error) {
	filtered := repository.filteredQuery(ctx, criteria)
	var count int64
	if err := filtered.Count(&count).Error; err != nil {
		return applicationdcim.InterfacePage{}, shared.WrapError(
			shared.ErrorReasonInternal, "Could not count Interfaces.", err,
		)
	}
	query := applyInterfaceOrdering(filtered, criteria.Ordering)
	if !criteria.DeferPagination {
		query = query.Limit(int(criteria.Limit)).Offset(int(criteria.Offset))
	}
	var rows []interfaceProjectionRow
	if err := selectInterfaceProjection(query).Find(&rows).Error; err != nil {
		return applicationdcim.InterfacePage{}, shared.WrapError(
			shared.ErrorReasonInternal, "Could not list Interfaces.", err,
		)
	}
	results := make([]*domaindcim.Interface, 0, len(rows))
	for _, row := range rows {
		networkInterface, err := interfaceFromProjection(row)
		if err != nil {
			return applicationdcim.InterfacePage{}, err
		}
		results = append(results, networkInterface)
	}
	return applicationdcim.InterfacePage{Count: uint64(count), Results: results}, nil
}

func (repository *InterfaceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Interface, error) {
	return repository.get(ctx, id, false)
}

func (repository *InterfaceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domaindcim.Interface, error) {
	return repository.get(ctx, id, true)
}

func (repository *InterfaceRepository) get(
	ctx context.Context,
	id shared.ID,
	forUpdate bool,
) (*domaindcim.Interface, error) {
	query := repository.baseQuery(ctx).Where(interfaceTableAlias+".id = ?", id.Int64())
	query = selectInterfaceProjection(query)
	if forUpdate {
		query = query.Clauses(clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: interfaceTableAlias},
		})
	}
	var row interfaceProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateInterfaceReadError(id, "get Interface", err)
	}
	return interfaceFromProjection(row)
}

func (repository *InterfaceRepository) ListForDeviceForUpdate(
	ctx context.Context,
	deviceID shared.ID,
) ([]*domaindcim.Interface, error) {
	query := selectInterfaceProjection(
		repository.baseQuery(ctx).
			Where(interfaceTableAlias+".device_id = ?", deviceID.Int64()).
			Order(clause.OrderByColumn{
				Column: clause.Column{Table: interfaceTableAlias, Name: "id"},
			}),
	).Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
		Table:    clause.Table{Name: interfaceTableAlias},
	})
	var rows []interfaceProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not lock Device Interfaces for deletion.",
			err,
		)
	}
	results := make([]*domaindcim.Interface, 0, len(rows))
	for _, row := range rows {
		networkInterface, err := interfaceFromProjection(row)
		if err != nil {
			return nil, err
		}
		results = append(results, networkInterface)
	}
	return results, nil
}

func (repository *InterfaceRepository) Create(
	ctx context.Context,
	networkInterface *domaindcim.Interface,
) error {
	if networkInterface == nil {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot persist a nil Interface.",
		)
	}
	if networkInterface.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot create an already persisted Interface.",
		)
	}
	row := interfaceToRow(*networkInterface)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateInterfaceMutationError("create Interface", err)
	}
	return networkInterface.AssignID(shared.ID(row.ID))
}

func (repository *InterfaceRepository) Update(
	ctx context.Context,
	networkInterface *domaindcim.Interface,
) error {
	if networkInterface == nil || !networkInterface.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot update an unpersisted Interface.",
		)
	}
	row := interfaceToRow(*networkInterface)
	result := repository.database(ctx).
		Model(&dcimrow.InterfaceRow{}).
		Where("id = ?", networkInterface.ID().Int64()).
		Select(
			"device_id", "name", "label", "type", "enabled", "mgmt_only",
			"mtu", "speed", "duplex", "description", "last_updated",
		).
		Updates(&row)
	if result.Error != nil {
		return translateInterfaceMutationError("update Interface", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Interface", networkInterface.ID())
	}
	return nil
}

func (repository *InterfaceRepository) Delete(
	ctx context.Context,
	networkInterface *domaindcim.Interface,
) error {
	if networkInterface == nil || !networkInterface.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot delete an unpersisted Interface.",
		)
	}
	result := repository.database(ctx).
		Where("id = ?", networkInterface.ID().Int64()).
		Delete(&dcimrow.InterfaceRow{})
	if result.Error != nil {
		return translateInterfaceMutationError("delete Interface", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("Interface", networkInterface.ID())
	}
	return nil
}

func (repository *InterfaceRepository) GetDeviceReference(
	ctx context.Context,
	id shared.ID,
) (domaindcim.DeviceReference, error) {
	var row deviceReferenceProjectionRow
	query := repository.database(ctx).
		Table((dcimrow.DeviceRow{}).TableName()+" AS "+interfaceDeviceAlias).
		Joins(
			"JOIN "+(dcimrow.DeviceTypeRow{}).TableName()+" AS "+
				interfaceDeviceTypeAlias+" ON "+interfaceDeviceTypeAlias+".id = "+
				interfaceDeviceAlias+".device_type_id",
		).
		Joins(
			"JOIN "+(dcimrow.ManufacturerRow{}).TableName()+" AS "+
				interfaceManufacturerAlias+" ON "+interfaceManufacturerAlias+".id = "+
				interfaceDeviceTypeAlias+".manufacturer_id",
		).
		Select(
			interfaceDeviceAlias+".id, "+interfaceDeviceAlias+".name, "+
				interfaceDeviceAlias+".asset_tag, "+interfaceDeviceTypeAlias+
				".model AS device_type_model, "+interfaceManufacturerAlias+
				".name AS manufacturer_name",
		).
		Where(interfaceDeviceAlias+".id = ?", id.Int64())
	if err := query.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domaindcim.DeviceReference{}, shared.NewValidationError(
				shared.FieldViolation{
					Field: "device", Reason: "invalid_choice",
					Description: "Select a valid choice.",
				},
			)
		}
		return domaindcim.DeviceReference{}, shared.WrapError(
			shared.ErrorReasonInternal, "Could not resolve selected Device.", err,
		)
	}
	return deviceReferenceFromProjection(row)
}

func (repository *InterfaceRepository) filteredQuery(
	ctx context.Context,
	criteria applicationdcim.InterfaceListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(interfaceTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where(
				interfaceTableAlias+".id IN ?", primitiveIDs(criteria.VisibleObjectIDs),
			)
		}
	}
	if criteria.Query != "" {
		pattern := containsPattern(criteria.Query)
		query = query.Where(
			"(LOWER("+interfaceTableAlias+".name) LIKE ? ESCAPE '\\' OR "+
				"LOWER("+interfaceTableAlias+".description) LIKE ? ESCAPE '\\')",
			pattern, pattern,
		)
	}
	if len(criteria.DeviceIDs) > 0 {
		query = query.Where(interfaceTableAlias+".device_id IN ?", criteria.DeviceIDs)
	}
	if len(criteria.DeviceNames) > 0 {
		query = query.Where(interfaceDeviceAlias+".name IN ?", criteria.DeviceNames)
	}
	if len(criteria.Names) > 0 {
		query = query.Where(interfaceTableAlias+".name IN ?", criteria.Names)
	}
	if len(criteria.Types) > 0 {
		values := make([]string, len(criteria.Types))
		for index, value := range criteria.Types {
			values[index] = value.String()
		}
		query = query.Where(interfaceTableAlias+".type IN ?", values)
	}
	if criteria.Enabled != nil {
		query = query.Where(interfaceTableAlias+".enabled = ?", *criteria.Enabled)
	}
	if criteria.MgmtOnly != nil {
		query = query.Where(interfaceTableAlias+".mgmt_only = ?", *criteria.MgmtOnly)
	}
	return query
}

func (repository *InterfaceRepository) baseQuery(ctx context.Context) *gorm.DB {
	return repository.database(ctx).
		Table(interfaceTableExpression()).
		Joins(
			"JOIN " + (dcimrow.DeviceRow{}).TableName() + " AS " +
				interfaceDeviceAlias + " ON " + interfaceDeviceAlias + ".id = " +
				interfaceTableAlias + ".device_id",
		).
		Joins(
			"JOIN " + (dcimrow.DeviceTypeRow{}).TableName() + " AS " +
				interfaceDeviceTypeAlias + " ON " + interfaceDeviceTypeAlias + ".id = " +
				interfaceDeviceAlias + ".device_type_id",
		).
		Joins(
			"JOIN " + (dcimrow.ManufacturerRow{}).TableName() + " AS " +
				interfaceManufacturerAlias + " ON " + interfaceManufacturerAlias + ".id = " +
				interfaceDeviceTypeAlias + ".manufacturer_id",
		)
}

func (repository *InterfaceRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type interfaceProjectionRow struct {
	dcimrow.InterfaceRow
	DeviceName       *string `gorm:"column:device_name"`
	DeviceAssetTag   *string `gorm:"column:device_asset_tag"`
	DeviceTypeModel  string  `gorm:"column:device_type_model"`
	ManufacturerName string  `gorm:"column:manufacturer_name"`
	IPAddressCount   int64   `gorm:"column:count_ipaddresses"`
}

type deviceReferenceProjectionRow struct {
	ID               int64
	Name             *string
	AssetTag         *string
	DeviceTypeModel  string `gorm:"column:device_type_model"`
	ManufacturerName string `gorm:"column:manufacturer_name"`
}

func selectInterfaceProjection(query *gorm.DB) *gorm.DB {
	ipAddressTable := (ipamrow.IPAddressRow{}).TableName()
	return query.Select(
		interfaceTableAlias+".*, "+
			interfaceDeviceAlias+".name AS device_name, "+
			interfaceDeviceAlias+".asset_tag AS device_asset_tag, "+
			interfaceDeviceTypeAlias+".model AS device_type_model, "+
			interfaceManufacturerAlias+".name AS manufacturer_name, "+
			"(SELECT COUNT(*) FROM "+ipAddressTable+" AS interface_ip_addresses "+
			"WHERE interface_ip_addresses.assigned_object_type = ? AND "+
			"interface_ip_addresses.assigned_object_id = "+interfaceTableAlias+
			".id) AS count_ipaddresses",
		domaindcim.InterfaceObjectType,
	)
}

func interfaceTableExpression() string {
	return (dcimrow.InterfaceRow{}).TableName() + " AS " + interfaceTableAlias
}

func applyInterfaceOrdering(
	query *gorm.DB,
	ordering []applicationdcim.InterfaceSort,
) *gorm.DB {
	hasUniqueOrdering := false
	for _, requested := range ordering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{
				Table: interfaceTableAlias,
				Name:  interfaceSortColumn(requested.Field),
			},
			Desc: requested.Descending,
		})
		if requested.Field == applicationdcim.InterfaceSortID {
			hasUniqueOrdering = true
		}
	}
	if !hasUniqueOrdering {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Table: interfaceTableAlias, Name: "id"},
		})
	}
	return query
}

func interfaceSortColumn(field applicationdcim.InterfaceSortField) string {
	switch field {
	case applicationdcim.InterfaceSortID:
		return "id"
	case applicationdcim.InterfaceSortDevice:
		return "device_id"
	case applicationdcim.InterfaceSortName:
		return "name"
	case applicationdcim.InterfaceSortType:
		return "type"
	case applicationdcim.InterfaceSortCreated:
		return "created"
	case applicationdcim.InterfaceSortLastUpdated:
		return "last_updated"
	default:
		return "id"
	}
}

func interfaceToRow(networkInterface domaindcim.Interface) dcimrow.InterfaceRow {
	state := networkInterface.State()
	return dcimrow.InterfaceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time,
			LastUpdated: state.LastUpdated.Time,
		},
		DeviceID: state.Device.ID().Int64(), Name: state.Name, Label: state.Label,
		Type: state.Type, Enabled: state.Enabled, MgmtOnly: state.MgmtOnly,
		MTU: nullableUint32ToInt64(state.MTU), Speed: nullableUint64ToInt64(state.Speed),
		Duplex: nullableDuplexToString(state.Duplex), Description: state.Description,
	}
}

func interfaceFromProjection(
	row interfaceProjectionRow,
) (*domaindcim.Interface, error) {
	reference, err := deviceReferenceFromProjection(deviceReferenceProjectionRow{
		ID: row.DeviceID, Name: row.DeviceName, AssetTag: row.DeviceAssetTag,
		DeviceTypeModel: row.DeviceTypeModel, ManufacturerName: row.ManufacturerName,
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted Interface relationship.",
			err,
		)
	}
	if row.IPAddressCount < 0 {
		return nil, shared.NewError(
			shared.ErrorReasonInternal,
			"Persisted Interface has an invalid IP address count.",
		)
	}
	networkInterface, err := domaindcim.RestoreInterface(domaindcim.InterfaceState{
		ID: shared.ID(row.ID), Device: reference, Name: row.Name, Label: row.Label,
		Type: row.Type, Enabled: row.Enabled, MgmtOnly: row.MgmtOnly,
		MTU: nullableInt64ToUint32(row.MTU), Speed: nullableInt64ToUint64(row.Speed),
		Duplex: nullableString(row.Duplex), Description: row.Description,
		Created:        shared.NewTimestamp(row.Created),
		LastUpdated:    shared.NewTimestamp(row.LastUpdated),
		IPAddressCount: uint64(row.IPAddressCount),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted Interface state.",
			err,
		)
	}
	return networkInterface, nil
}

func deviceReferenceFromProjection(
	row deviceReferenceProjectionRow,
) (domaindcim.DeviceReference, error) {
	name := domaindcim.NullDeviceValue[string]()
	if row.Name != nil {
		name = domaindcim.NonNullDeviceValue(*row.Name)
	}
	display := strings.TrimSpace(row.ManufacturerName) + " " +
		strings.TrimSpace(row.DeviceTypeModel)
	if row.Name != nil && strings.TrimSpace(*row.Name) != "" {
		display = strings.TrimSpace(*row.Name)
	}
	if row.AssetTag != nil && strings.TrimSpace(*row.AssetTag) != "" {
		display += " (" + strings.TrimSpace(*row.AssetTag) + ")"
	} else if (row.Name == nil || strings.TrimSpace(*row.Name) == "") && row.ID > 0 {
		display += fmt.Sprintf(" (%d)", row.ID)
	}
	return domaindcim.NewDeviceReference(shared.ID(row.ID), name, display)
}

func nullableUint32ToInt64(value domaindcim.DeviceNullable[uint32]) *int64 {
	raw, present := value.Get()
	if !present {
		return nil
	}
	converted := int64(raw)
	return &converted
}

func nullableUint64ToInt64(value domaindcim.DeviceNullable[uint64]) *int64 {
	raw, present := value.Get()
	if !present {
		return nil
	}
	converted := int64(raw)
	return &converted
}

func nullableDuplexToString(value domaindcim.DeviceNullable[string]) *string {
	raw, present := value.Get()
	if !present {
		return nil
	}
	return &raw
}

func nullableInt64ToUint32(value *int64) domaindcim.DeviceNullable[uint32] {
	if value == nil {
		return domaindcim.NullDeviceValue[uint32]()
	}
	return domaindcim.NonNullDeviceValue(uint32(*value))
}

func nullableInt64ToUint64(value *int64) domaindcim.DeviceNullable[uint64] {
	if value == nil {
		return domaindcim.NullDeviceValue[uint64]()
	}
	return domaindcim.NonNullDeviceValue(uint64(*value))
}

func nullableString(value *string) domaindcim.DeviceNullable[string] {
	if value == nil {
		return domaindcim.NullDeviceValue[string]()
	}
	return domaindcim.NonNullDeviceValue(*value)
}

func translateInterfaceReadError(id shared.ID, operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("Interface", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}

func translateInterfaceMutationError(operation string, err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "uq_go_interface_name"),
		strings.Contains(message, "go_dcim_interfaces.device_id") &&
			strings.Contains(message, "go_dcim_interfaces.name"):
		const description = "The fields device, name must make a unique set."
		return shared.ConflictWithViolations(
			description, err,
			shared.FieldViolation{
				Field: "non_field_errors", Reason: "unique_together",
				Description: description,
			},
		)
	case duplicateConstraint(err):
		return shared.Conflict("A matching Interface already exists.", err)
	case foreignKeyConstraint(err):
		return shared.Conflict("The selected Device no longer exists.", err)
	default:
		return shared.WrapError(
			shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
		)
	}
}
