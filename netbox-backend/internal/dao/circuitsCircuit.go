package dao

import (
	"context"
	"errors"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"netbox-go/internal/cache"
	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

var _ CircuitsCircuitDao = (*circuitsCircuitDao)(nil)

// CircuitsCircuitDao defining the dao interface
type CircuitsCircuitDao interface {
	Create(ctx context.Context, table *model.CircuitsCircuit) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.CircuitsCircuit) error
	GetByID(ctx context.Context, id uint64) (*model.CircuitsCircuit, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.CircuitsCircuit, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.CircuitsCircuit, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuit, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.CircuitsCircuit, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsCircuit) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsCircuit) error
}

type circuitsCircuitDao struct {
	db    *gorm.DB
	cache cache.CircuitsCircuitCache // if nil, the cache is not used.
	sfg   *singleflight.Group        // if cache is nil, the sfg is not used.
}

// NewCircuitsCircuitDao creating the dao interface
func NewCircuitsCircuitDao(db *gorm.DB, xCache cache.CircuitsCircuitCache) CircuitsCircuitDao {
	if xCache == nil {
		return &circuitsCircuitDao{db: db}
	}
	return &circuitsCircuitDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *circuitsCircuitDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new circuitsCircuit, insert the record and the id value is written back to the table
func (d *circuitsCircuitDao) Create(ctx context.Context, table *model.CircuitsCircuit) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a circuitsCircuit by id
func (d *circuitsCircuitDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.CircuitsCircuit{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a circuitsCircuit by ids
func (d *circuitsCircuitDao) UpdateByID(ctx context.Context, table *model.CircuitsCircuit) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *circuitsCircuitDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.CircuitsCircuit) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.Created != nil && table.Created.IsZero() == false {
		update["created"] = table.Created
	}
	if table.LastUpdated != nil && table.LastUpdated.IsZero() == false {
		update["last_updated"] = table.LastUpdated
	}
	if table.CustomFieldData != nil && table.CustomFieldData.String() != "" {
		update["custom_field_data"] = table.CustomFieldData
	}
	if table.Cid != "" {
		update["cid"] = table.Cid
	}
	if table.Status != "" {
		update["status"] = table.Status
	}
	if table.InstallDate != nil && table.InstallDate.IsZero() == false {
		update["install_date"] = table.InstallDate
	}
	if table.CommitRate != 0 {
		update["commit_rate"] = table.CommitRate
	}
	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.Comments != "" {
		update["comments"] = table.Comments
	}
	if table.ProviderID != 0 {
		update["provider_id"] = table.ProviderID
	}
	if table.TenantID != 0 {
		update["tenant_id"] = table.TenantID
	}
	if table.TerminationAID != 0 {
		update["termination_a_id"] = table.TerminationAID
	}
	if table.TerminationZID != 0 {
		update["termination_z_id"] = table.TerminationZID
	}
	if table.TypeID != 0 {
		update["type_id"] = table.TypeID
	}
	if table.TerminationDate != nil && table.TerminationDate.IsZero() == false {
		update["termination_date"] = table.TerminationDate
	}
	if table.ProviderAccountID != 0 {
		update["provider_account_id"] = table.ProviderAccountID
	}
	if table.XAbsDistance != nil && table.XAbsDistance.IsZero() == false {
		update["_abs_distance"] = table.XAbsDistance
	}
	if table.Distance != nil && table.Distance.IsZero() == false {
		update["distance"] = table.Distance
	}
	if table.DistanceUnit != "" {
		update["distance_unit"] = table.DistanceUnit
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a circuitsCircuit by id
func (d *circuitsCircuitDao) GetByID(ctx context.Context, id uint64) (*model.CircuitsCircuit, error) {
	// no cache
	if d.cache == nil {
		record := &model.CircuitsCircuit{}
		err := d.db.WithContext(ctx).Where("id = ?", id).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, id)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same id, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(utils.Uint64ToStr(id), func() (interface{}, error) {
			table := &model.CircuitsCircuit{}
			err = d.db.WithContext(ctx).Where("id = ?", id).First(table).Error
			if err != nil {
				// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, database.ErrRecordNotFound) {
					if err = d.cache.SetPlaceholder(ctx, id); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("id", id))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, id, table, cache.CircuitsCircuitExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.CircuitsCircuit)
		if !ok {
			return nil, database.ErrRecordNotFound
		}
		return table, nil
	}

	if d.cache.IsPlaceholderErr(err) {
		return nil, database.ErrRecordNotFound
	}

	return nil, err
}

// GetByColumns get a paginated list of circuitsCircuits by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *circuitsCircuitDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.CircuitsCircuit, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.CircuitsCircuitColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.CircuitsCircuit{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.CircuitsCircuit{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete circuitsCircuit by ids
func (d *circuitsCircuitDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.CircuitsCircuit{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a circuitsCircuit by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *circuitsCircuitDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.CircuitsCircuit, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.CircuitsCircuitColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.CircuitsCircuit{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs Batch get circuitsCircuit by ids
func (d *circuitsCircuitDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsCircuit, error) {
	// no cache
	if d.cache == nil {
		var records []*model.CircuitsCircuit
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.CircuitsCircuit)
		for _, record := range records {
			itemMap[record.ID] = record
		}
		return itemMap, nil
	}

	// get form cache
	itemMap, err := d.cache.MultiGet(ctx, ids)
	if err != nil {
		return nil, err
	}

	var missedIDs []uint64
	for _, id := range ids {
		if _, ok := itemMap[id]; !ok {
			missedIDs = append(missedIDs, id)
		}
	}

	// get missed data
	if len(missedIDs) > 0 {
		// find the id of an active placeholder, i.e. an id that does not exist in database
		var realMissedIDs []uint64
		for _, id := range missedIDs {
			_, err = d.cache.Get(ctx, id)
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissedIDs = append(realMissedIDs, id)
		}

		// get missed id from database
		if len(realMissedIDs) > 0 {
			var records []*model.CircuitsCircuit
			var recordIDMap = make(map[uint64]struct{})
			err = d.db.WithContext(ctx).Where("id IN (?)", realMissedIDs).Find(&records).Error
			if err != nil {
				return nil, err
			}
			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.ID] = record
					recordIDMap[record.ID] = struct{}{}
				}
				if err = d.cache.MultiSet(ctx, records, cache.CircuitsCircuitExpireTime); err != nil {
					logger.Warn("cache.MultiSet error", logger.Err(err), logger.Any("ids", records))
				}
				if len(records) == len(realMissedIDs) {
					return itemMap, nil
				}
			}
			for _, id := range realMissedIDs {
				if _, ok := recordIDMap[id]; !ok {
					if err = d.cache.SetPlaceholder(ctx, id); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("id", id))
					}
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLastID Get a paginated list of circuitsCircuits by last id
func (d *circuitsCircuitDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.CircuitsCircuit, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.CircuitsCircuit{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *circuitsCircuitDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsCircuit) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *circuitsCircuitDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.CircuitsCircuit{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *circuitsCircuitDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsCircuit) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
