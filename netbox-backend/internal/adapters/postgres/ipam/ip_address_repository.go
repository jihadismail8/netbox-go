package ipam

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

const (
	ipAddressTableAlias        = "typed_ip_addresses"
	ipAddressVRFAlias          = "ip_address_vrfs"
	ipAddressInterfaceAlias    = "ip_address_interfaces"
	ipAddressDeviceAlias       = "ip_address_devices"
	ipAddressDeviceTypeAlias   = "ip_address_device_types"
	ipAddressManufacturerAlias = "ip_address_manufacturers"
)

type IPAddressRepository struct{ db *gorm.DB }

var _ applicationipam.IPAddressRepository = (*IPAddressRepository)(nil)

func NewIPAddressRepository(db *gorm.DB) *IPAddressRepository {
	if db == nil {
		panic("postgres IPAddress repository requires a database")
	}
	return &IPAddressRepository{db: db}
}

func (repository *IPAddressRepository) List(
	ctx context.Context,
	criteria applicationipam.IPAddressListCriteria,
) (applicationipam.IPAddressPage, error) {
	var rows []ipAddressProjectionRow
	if err := selectIPAddressProjection(
		repository.filteredQuery(ctx, criteria),
	).Find(&rows).Error; err != nil {
		return applicationipam.IPAddressPage{},
			translateIPAddressReadError(0, "list IP addresses", err)
	}
	addresses := make([]*domainipam.IPAddress, 0, len(rows))
	for _, row := range rows {
		address, err := ipAddressFromProjection(row)
		if err != nil {
			return applicationipam.IPAddressPage{}, err
		}
		if matchesIPAddressCriteria(address, criteria) {
			addresses = append(addresses, address)
		}
	}
	sortIPAddresses(addresses, criteria.Ordering)
	count := uint64(len(addresses))
	if !criteria.DeferPagination {
		start := min(int(criteria.Offset), len(addresses))
		end := min(start+int(criteria.Limit), len(addresses))
		addresses = addresses[start:end]
	}
	return applicationipam.IPAddressPage{Count: count, Results: addresses}, nil
}

func (repository *IPAddressRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*domainipam.IPAddress, error) {
	return repository.get(ctx, id)
}

func (repository *IPAddressRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*domainipam.IPAddress, error) {
	// Lock the base row before reading the joined projection. Combining the
	// lock with LEFT JOINs can yield a mixed PostgreSQL READ COMMITTED result
	// after a concurrent updater commits: the target row is rechecked while
	// nullable join columns still reflect the pre-wait snapshot.
	var locked ipamrow.IPAddressRow
	if err := repository.database(ctx).
		Select("id").
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("id = ?", id.Int64()).
		Take(&locked).Error; err != nil {
		return nil, translateIPAddressReadError(
			id, "lock IP address for update", err,
		)
	}
	return repository.get(ctx, id)
}

func (repository *IPAddressRepository) get(
	ctx context.Context,
	id shared.ID,
) (*domainipam.IPAddress, error) {
	query := selectIPAddressProjection(
		repository.baseQuery(ctx).
			Where(ipAddressTableAlias+".id = ?", id.Int64()),
	)
	var row ipAddressProjectionRow
	if err := query.Take(&row).Error; err != nil {
		return nil, translateIPAddressReadError(id, "get IP address", err)
	}
	return ipAddressFromProjection(row)
}

func (repository *IPAddressRepository) Create(
	ctx context.Context,
	address *domainipam.IPAddress,
) error {
	if address == nil {
		return shared.NewError(
			shared.ErrorReasonInternal, "Cannot persist a nil IPAddress.",
		)
	}
	if address.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot create an already persisted IPAddress.",
		)
	}
	row := ipAddressToRow(*address)
	if err := repository.database(ctx).Create(&row).Error; err != nil {
		return translateIPAddressMutationError("create IP address", err)
	}
	return address.AssignID(shared.ID(row.ID))
}

func (repository *IPAddressRepository) Update(
	ctx context.Context,
	address *domainipam.IPAddress,
) error {
	if address == nil || !address.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot update an unpersisted IPAddress.",
		)
	}
	row := ipAddressToRow(*address)
	result := repository.database(ctx).
		Model(&ipamrow.IPAddressRow{}).
		Where("id = ?", address.ID().Int64()).
		Select(ipAddressUpdateColumns()).
		Updates(&row)
	if result.Error != nil {
		return translateIPAddressMutationError("update IP address", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("IPAddress", address.ID())
	}
	return nil
}

func (repository *IPAddressRepository) Delete(
	ctx context.Context,
	address *domainipam.IPAddress,
) error {
	if address == nil || !address.ID().IsValid() {
		return shared.NewError(
			shared.ErrorReasonInternal,
			"Cannot delete an unpersisted IPAddress.",
		)
	}
	result := repository.database(ctx).
		Where("id = ?", address.ID().Int64()).
		Delete(&ipamrow.IPAddressRow{})
	if result.Error != nil {
		return translateIPAddressMutationError("delete IP address", result.Error)
	}
	if result.RowsAffected == 0 {
		return shared.NotFound("IPAddress", address.ID())
	}
	return nil
}

func (repository *IPAddressRepository) LockUniqueness(
	ctx context.Context,
	vrf domainipam.NullableVRFReference,
	address domainipam.HostAddress,
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
	_, _ = hasher.Write([]byte(
		"ip-address:" + scope + ":" + address.Host().String(),
	))
	if err := db.Exec(
		"SELECT pg_advisory_xact_lock(?)", int64(hasher.Sum64()),
	).Error; err != nil {
		return shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not lock the IPAddress uniqueness scope.",
			err,
		)
	}
	return nil
}

func (repository *IPAddressRepository) FindDuplicates(
	ctx context.Context,
	vrf domainipam.NullableVRFReference,
	address domainipam.HostAddress,
	exclude shared.ID,
) ([]*domainipam.IPAddress, error) {
	query := repository.baseQuery(ctx).
		Where(ipAddressTableAlias+".address LIKE ?", address.Host().String()+"/%")
	if reference, present := vrf.Get(); present {
		query = query.Where(
			ipAddressTableAlias+".vrf_id = ?", reference.ID().Int64(),
		)
	} else {
		query = query.Where(ipAddressTableAlias + ".vrf_id IS NULL")
	}
	if exclude.IsValid() {
		query = query.Where(ipAddressTableAlias+".id <> ?", exclude.Int64())
	}
	var rows []ipAddressProjectionRow
	if err := selectIPAddressProjection(query).
		Order(ipAddressTableAlias + ".id").
		Find(&rows).Error; err != nil {
		return nil, translateIPAddressReadError(
			0, "find duplicate IP addresses", err,
		)
	}
	results := make([]*domainipam.IPAddress, 0, len(rows))
	for _, row := range rows {
		candidate, err := ipAddressFromProjection(row)
		if err != nil {
			return nil, err
		}
		if candidate.Address().Host() == address.Host() {
			results = append(results, candidate)
		}
	}
	return results, nil
}

func (repository *IPAddressRepository) ListAssignedToInterfaceForUpdate(
	ctx context.Context,
	interfaceID shared.ID,
) ([]*domainipam.IPAddress, error) {
	query := selectIPAddressProjection(
		repository.baseQuery(ctx).
			Where(
				ipAddressTableAlias+".assigned_object_type = ?",
				domainipam.IPAddressAssignmentType,
			).
			Where(
				ipAddressTableAlias+".assigned_object_id = ?",
				interfaceID.Int64(),
			).
			Order(ipAddressTableAlias + ".id"),
	).Clauses(clause.Locking{
		Strength: clause.LockingStrengthUpdate,
		Table:    clause.Table{Name: ipAddressTableAlias},
	})
	var rows []ipAddressProjectionRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, translateIPAddressReadError(
			0, "lock Interface IP addresses for deletion", err,
		)
	}
	results := make([]*domainipam.IPAddress, 0, len(rows))
	for _, row := range rows {
		address, err := ipAddressFromProjection(row)
		if err != nil {
			return nil, err
		}
		results = append(results, address)
	}
	return results, nil
}

func (repository *IPAddressRepository) filteredQuery(
	ctx context.Context,
	criteria applicationipam.IPAddressListCriteria,
) *gorm.DB {
	query := repository.baseQuery(ctx)
	if len(criteria.IDs) > 0 {
		query = query.Where(ipAddressTableAlias+".id IN ?", criteria.IDs)
	}
	if criteria.VisibilityConstrained {
		if len(criteria.VisibleObjectIDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			ids := make([]int64, len(criteria.VisibleObjectIDs))
			for index, id := range criteria.VisibleObjectIDs {
				ids[index] = id.Int64()
			}
			query = query.Where(ipAddressTableAlias+".id IN ?", ids)
		}
	}
	if len(criteria.VRFIDs) > 0 {
		query = query.Where(ipAddressTableAlias+".vrf_id IN ?", criteria.VRFIDs)
	}
	if len(criteria.VRFRDs) > 0 {
		query = query.Where(ipAddressVRFAlias+".rd IN ?", criteria.VRFRDs)
	}
	if criteria.Family != nil {
		switch *criteria.Family {
		case 4:
			query = query.Where(ipAddressTableAlias+".address NOT LIKE ?", "%:%")
		case 6:
			query = query.Where(ipAddressTableAlias+".address LIKE ?", "%:%")
		default:
			query = query.Where("1 = 0")
		}
	}
	if len(criteria.Statuses) > 0 {
		values := make([]string, len(criteria.Statuses))
		for index, status := range criteria.Statuses {
			values[index] = status.String()
		}
		query = query.Where(ipAddressTableAlias+".status IN ?", values)
	}
	if criteria.Assigned != nil {
		if *criteria.Assigned {
			query = query.Where(
				ipAddressTableAlias+".assigned_object_type = ? AND "+
					ipAddressTableAlias+".assigned_object_id IS NOT NULL",
				domainipam.IPAddressAssignmentType,
			)
		} else {
			query = query.Where(
				ipAddressTableAlias + ".assigned_object_type IS NULL AND " +
					ipAddressTableAlias + ".assigned_object_id IS NULL",
			)
		}
	}
	if len(criteria.InterfaceIDs) > 0 {
		query = query.Where(
			ipAddressTableAlias+".assigned_object_type = ? AND "+
				ipAddressTableAlias+".assigned_object_id IN ?",
			domainipam.IPAddressAssignmentType,
			criteria.InterfaceIDs,
		)
	}
	if len(criteria.DeviceIDs) > 0 {
		query = query.Where(
			ipAddressInterfaceAlias+".device_id IN ?", criteria.DeviceIDs,
		)
	}
	return query
}

func (repository *IPAddressRepository) baseQuery(ctx context.Context) *gorm.DB {
	return repository.database(ctx).
		Table(ipAddressTableExpression()).
		Joins(
			"LEFT JOIN "+(ipamrow.VRFRow{}).TableName()+" AS "+
				ipAddressVRFAlias+" ON "+ipAddressVRFAlias+".id = "+
				ipAddressTableAlias+".vrf_id",
		).
		Joins(
			"LEFT JOIN "+(dcimrow.InterfaceRow{}).TableName()+" AS "+
				ipAddressInterfaceAlias+" ON "+ipAddressTableAlias+
				".assigned_object_type = ? AND "+ipAddressInterfaceAlias+
				".id = "+ipAddressTableAlias+".assigned_object_id",
			domainipam.IPAddressAssignmentType,
		).
		Joins(
			"LEFT JOIN " + (dcimrow.DeviceRow{}).TableName() + " AS " +
				ipAddressDeviceAlias + " ON " + ipAddressDeviceAlias + ".id = " +
				ipAddressInterfaceAlias + ".device_id",
		).
		Joins(
			"LEFT JOIN " + (dcimrow.DeviceTypeRow{}).TableName() + " AS " +
				ipAddressDeviceTypeAlias + " ON " + ipAddressDeviceTypeAlias + ".id = " +
				ipAddressDeviceAlias + ".device_type_id",
		).
		Joins(
			"LEFT JOIN " + (dcimrow.ManufacturerRow{}).TableName() + " AS " +
				ipAddressManufacturerAlias + " ON " + ipAddressManufacturerAlias +
				".id = " + ipAddressDeviceTypeAlias + ".manufacturer_id",
		)
}

func (repository *IPAddressRepository) database(ctx context.Context) *gorm.DB {
	db := repository.db
	if tx, ok := postgresTransaction.FromContext(ctx); ok {
		db = tx
	}
	return db.WithContext(ctx)
}

type ipAddressProjectionRow struct {
	ipamrow.IPAddressRow
	VRFName                string     `gorm:"column:vrf_name"`
	VRFRD                  *string    `gorm:"column:vrf_rd"`
	VRFEnforceUnique       bool       `gorm:"column:vrf_enforce_unique"`
	InterfaceID            *int64     `gorm:"column:interface_id"`
	InterfaceDeviceID      *int64     `gorm:"column:interface_device_id"`
	InterfaceName          *string    `gorm:"column:interface_name"`
	InterfaceLabel         *string    `gorm:"column:interface_label"`
	InterfaceType          *string    `gorm:"column:interface_type"`
	InterfaceEnabled       *bool      `gorm:"column:interface_enabled"`
	InterfaceMgmtOnly      *bool      `gorm:"column:interface_mgmt_only"`
	InterfaceMTU           *int64     `gorm:"column:interface_mtu"`
	InterfaceSpeed         *int64     `gorm:"column:interface_speed"`
	InterfaceDuplex        *string    `gorm:"column:interface_duplex"`
	InterfaceDescription   *string    `gorm:"column:interface_description"`
	InterfaceCreated       *time.Time `gorm:"column:interface_created"`
	InterfaceLastUpdated   *time.Time `gorm:"column:interface_last_updated"`
	DeviceName             *string    `gorm:"column:device_name"`
	DeviceAssetTag         *string    `gorm:"column:device_asset_tag"`
	DeviceTypeModel        *string    `gorm:"column:device_type_model"`
	DeviceManufacturerName *string    `gorm:"column:device_manufacturer_name"`
}

func selectIPAddressProjection(query *gorm.DB) *gorm.DB {
	return query.Select(
		ipAddressTableAlias + ".*, " +
			ipAddressVRFAlias + ".name AS vrf_name, " +
			ipAddressVRFAlias + ".rd AS vrf_rd, " +
			ipAddressVRFAlias + ".enforce_unique AS vrf_enforce_unique, " +
			ipAddressInterfaceAlias + ".id AS interface_id, " +
			ipAddressInterfaceAlias + ".device_id AS interface_device_id, " +
			ipAddressInterfaceAlias + ".name AS interface_name, " +
			ipAddressInterfaceAlias + ".label AS interface_label, " +
			ipAddressInterfaceAlias + ".type AS interface_type, " +
			ipAddressInterfaceAlias + ".enabled AS interface_enabled, " +
			ipAddressInterfaceAlias + ".mgmt_only AS interface_mgmt_only, " +
			ipAddressInterfaceAlias + ".mtu AS interface_mtu, " +
			ipAddressInterfaceAlias + ".speed AS interface_speed, " +
			ipAddressInterfaceAlias + ".duplex AS interface_duplex, " +
			ipAddressInterfaceAlias + ".description AS interface_description, " +
			ipAddressInterfaceAlias + ".created AS interface_created, " +
			ipAddressInterfaceAlias + ".last_updated AS interface_last_updated, " +
			ipAddressDeviceAlias + ".name AS device_name, " +
			ipAddressDeviceAlias + ".asset_tag AS device_asset_tag, " +
			ipAddressDeviceTypeAlias + ".model AS device_type_model, " +
			ipAddressManufacturerAlias + ".name AS device_manufacturer_name",
	)
}

func ipAddressTableExpression() string {
	return (ipamrow.IPAddressRow{}).TableName() + " AS " +
		ipAddressTableAlias
}

func ipAddressFromProjection(
	row ipAddressProjectionRow,
) (*domainipam.IPAddress, error) {
	vrf, err := nullableIPAddressVRF(row)
	if err != nil {
		return nil, err
	}
	assignment, err := nullableIPAddressAssignment(row)
	if err != nil {
		return nil, err
	}
	role := domainipam.NullIPAddressRole()
	if row.Role != nil {
		role = domainipam.NonNullIPAddressRole(
			domainipam.IPAddressRole(*row.Role),
		)
	}
	address, err := domainipam.RestoreIPAddress(domainipam.IPAddressState{
		ID: shared.ID(row.ID), Address: row.Address, VRF: vrf,
		Status: row.Status, Role: role, DNSName: row.DNSName,
		Description: row.Description, Comments: row.Comments,
		Assignment: assignment, Created: shared.NewTimestamp(row.Created),
		LastUpdated: shared.NewTimestamp(row.LastUpdated),
	})
	if err != nil {
		return nil, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted IPAddress state.",
			err,
		)
	}
	return address, nil
}

func nullableIPAddressVRF(
	row ipAddressProjectionRow,
) (domainipam.NullableVRFReference, error) {
	if row.VRFID == nil {
		return domainipam.NullVRFReference(), nil
	}
	rd, err := nullableRouteDistinguisher(row.VRFRD)
	if err != nil {
		return domainipam.NullableVRFReference{}, err
	}
	reference, err := domainipam.NewVRFReference(
		shared.ID(*row.VRFID),
		row.VRFName,
		rd,
		row.VRFEnforceUnique,
	)
	if err != nil {
		return domainipam.NullableVRFReference{}, shared.WrapError(
			shared.ErrorReasonInternal,
			"Persisted IPAddress references an invalid VRF.",
			err,
		)
	}
	return domainipam.NonNullVRFReference(reference), nil
}

func nullableIPAddressAssignment(
	row ipAddressProjectionRow,
) (domainipam.NullableInterfaceAssignment, error) {
	if row.AssignedObjectType == nil && row.AssignedObjectID == nil {
		return domainipam.NullInterfaceAssignment(), nil
	}
	if row.AssignedObjectType == nil || row.AssignedObjectID == nil ||
		*row.AssignedObjectType != domainipam.IPAddressAssignmentType ||
		row.InterfaceID == nil || row.InterfaceDeviceID == nil ||
		row.InterfaceName == nil || row.InterfaceLabel == nil ||
		row.InterfaceType == nil || row.InterfaceEnabled == nil ||
		row.InterfaceMgmtOnly == nil || row.InterfaceDescription == nil ||
		row.InterfaceCreated == nil || row.InterfaceLastUpdated == nil ||
		row.DeviceTypeModel == nil || row.DeviceManufacturerName == nil {
		return domainipam.NullableInterfaceAssignment{}, shared.NewError(
			shared.ErrorReasonInternal,
			"Persisted IPAddress contains an invalid Interface assignment.",
		)
	}
	deviceReference, err := ipAddressDeviceReference(row)
	if err != nil {
		return domainipam.NullableInterfaceAssignment{}, err
	}
	networkInterface, err := domaindcim.RestoreInterface(
		domaindcim.InterfaceState{
			ID: shared.ID(*row.InterfaceID), Device: deviceReference,
			Name: *row.InterfaceName, Label: *row.InterfaceLabel,
			Type: *row.InterfaceType, Enabled: *row.InterfaceEnabled,
			MgmtOnly:    *row.InterfaceMgmtOnly,
			MTU:         nullableIPAddressInterfaceUint32(row.InterfaceMTU),
			Speed:       nullableIPAddressInterfaceUint64(row.InterfaceSpeed),
			Duplex:      nullableIPAddressInterfaceString(row.InterfaceDuplex),
			Description: *row.InterfaceDescription,
			Created:     shared.NewTimestamp(*row.InterfaceCreated),
			LastUpdated: shared.NewTimestamp(*row.InterfaceLastUpdated),
		},
	)
	if err != nil {
		return domainipam.NullableInterfaceAssignment{}, shared.WrapError(
			shared.ErrorReasonInternal,
			"Could not restore persisted IPAddress Interface assignment.",
			err,
		)
	}
	assignment, err := domainipam.NewInterfaceAssignment(networkInterface)
	if err != nil {
		return domainipam.NullableInterfaceAssignment{}, err
	}
	return domainipam.NonNullInterfaceAssignment(assignment), nil
}

func ipAddressDeviceReference(
	row ipAddressProjectionRow,
) (domaindcim.DeviceReference, error) {
	name := domaindcim.NullDeviceValue[string]()
	if row.DeviceName != nil {
		name = domaindcim.NonNullDeviceValue(*row.DeviceName)
	}
	display := strings.TrimSpace(*row.DeviceManufacturerName) + " " +
		strings.TrimSpace(*row.DeviceTypeModel)
	if row.DeviceName != nil && strings.TrimSpace(*row.DeviceName) != "" {
		display = strings.TrimSpace(*row.DeviceName)
	}
	if row.DeviceAssetTag != nil && strings.TrimSpace(*row.DeviceAssetTag) != "" {
		display += " (" + strings.TrimSpace(*row.DeviceAssetTag) + ")"
	} else if (row.DeviceName == nil ||
		strings.TrimSpace(*row.DeviceName) == "") &&
		row.InterfaceDeviceID != nil {
		display += fmt.Sprintf(" (%d)", *row.InterfaceDeviceID)
	}
	return domaindcim.NewDeviceReference(
		shared.ID(*row.InterfaceDeviceID), name, display,
	)
}

func nullableIPAddressInterfaceUint32(
	value *int64,
) domaindcim.DeviceNullable[uint32] {
	if value == nil {
		return domaindcim.NullDeviceValue[uint32]()
	}
	return domaindcim.NonNullDeviceValue(uint32(*value))
}

func nullableIPAddressInterfaceUint64(
	value *int64,
) domaindcim.DeviceNullable[uint64] {
	if value == nil {
		return domaindcim.NullDeviceValue[uint64]()
	}
	return domaindcim.NonNullDeviceValue(uint64(*value))
}

func nullableIPAddressInterfaceString(
	value *string,
) domaindcim.DeviceNullable[string] {
	if value == nil {
		return domaindcim.NullDeviceValue[string]()
	}
	return domaindcim.NonNullDeviceValue(*value)
}

func ipAddressToRow(address domainipam.IPAddress) ipamrow.IPAddressRow {
	state := address.State()
	var vrfID *int64
	if reference, present := state.VRF.Get(); present {
		value := reference.ID().Int64()
		vrfID = &value
	}
	var role *string
	if value, present := state.Role.Get(); present {
		text := value.String()
		role = &text
	}
	var assignedObjectType *string
	var assignedObjectID *int64
	if assignment, present := state.Assignment.Get(); present {
		objectType := domainipam.IPAddressAssignmentType
		id := assignment.ID().Int64()
		assignedObjectType, assignedObjectID = &objectType, &id
	}
	return ipamrow.IPAddressRow{
		RowMetadata: ipamrow.RowMetadata{
			ID: state.ID.Int64(), Created: state.Created.Time,
			LastUpdated: state.LastUpdated.Time,
		},
		Address: state.Address, VRFID: vrfID, Status: state.Status,
		Role: role, DNSName: state.DNSName, Description: state.Description,
		Comments: state.Comments, AssignedObjectType: assignedObjectType,
		AssignedObjectID: assignedObjectID,
	}
}

func matchesIPAddressCriteria(
	address *domainipam.IPAddress,
	criteria applicationipam.IPAddressListCriteria,
) bool {
	if address == nil {
		return false
	}
	if criteria.Query != "" && !matchesIPAddressSearch(address, criteria.Query) {
		return false
	}
	if criteria.AddressesPresent {
		matched := false
		for _, filter := range criteria.Addresses {
			if !filter.Valid {
				continue
			}
			if filter.ExplicitMask {
				matched = address.Address().Compare(filter.Address) == 0
			} else {
				matched = address.Address().Host() == filter.Address.Host()
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}
	if criteria.Parent != nil {
		if !criteria.Parent.Valid ||
			!criteria.Parent.Network.Contains(address.Address().Host()) {
			return false
		}
	}
	return true
}

func matchesIPAddressSearch(
	address *domainipam.IPAddress,
	query string,
) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower == "" {
		return true
	}
	return strings.Contains(strings.ToLower(address.Address().String()), lower) ||
		strings.Contains(strings.ToLower(address.DNSName()), lower) ||
		strings.Contains(strings.ToLower(address.Description()), lower) ||
		strings.Contains(strings.ToLower(address.Comments()), lower)
}

func sortIPAddresses(
	addresses []*domainipam.IPAddress,
	ordering []applicationipam.IPAddressSort,
) {
	if len(ordering) == 0 {
		ordering = []applicationipam.IPAddressSort{
			{Field: applicationipam.IPAddressSortVRF},
			{Field: applicationipam.IPAddressSortAddress},
		}
	}
	hasID := false
	for _, requested := range ordering {
		if requested.Field == applicationipam.IPAddressSortID {
			hasID = true
		}
	}
	if !hasID {
		ordering = append(ordering, applicationipam.IPAddressSort{
			Field: applicationipam.IPAddressSortID,
		})
	}
	sort.SliceStable(addresses, func(leftIndex, rightIndex int) bool {
		left, right := addresses[leftIndex], addresses[rightIndex]
		for _, requested := range ordering {
			compared := compareIPAddressField(left, right, requested.Field)
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

func compareIPAddressField(
	left, right *domainipam.IPAddress,
	field applicationipam.IPAddressSortField,
) int {
	switch field {
	case applicationipam.IPAddressSortID:
		return compareInt64(left.ID().Int64(), right.ID().Int64())
	case applicationipam.IPAddressSortVRF:
		return comparePrefixVRF(left.VRF(), right.VRF())
	case applicationipam.IPAddressSortAddress:
		return left.Address().Compare(right.Address())
	case applicationipam.IPAddressSortStatus:
		return strings.Compare(left.Status().String(), right.Status().String())
	case applicationipam.IPAddressSortDNSName:
		return strings.Compare(left.DNSName(), right.DNSName())
	case applicationipam.IPAddressSortCreated:
		return left.Created().Compare(right.Created().Time)
	case applicationipam.IPAddressSortLastUpdated:
		return left.LastUpdated().Compare(right.LastUpdated().Time)
	default:
		return 0
	}
}

func ipAddressUpdateColumns() []string {
	return []string{
		"address", "vrf_id", "status", "role", "dns_name", "description",
		"comments", "assigned_object_type", "assigned_object_id",
		"last_updated",
	}
}

func translateIPAddressReadError(
	id shared.ID,
	operation string,
	err error,
) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shared.NotFound("IPAddress", id)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}

func translateIPAddressMutationError(operation string, err error) error {
	if foreignKeyConstraint(err) {
		return shared.Conflict(
			"The selected VRF or Interface no longer exists.", err,
		)
	}
	return shared.WrapError(
		shared.ErrorReasonInternal, fmt.Sprintf("Could not %s.", operation), err,
	)
}
