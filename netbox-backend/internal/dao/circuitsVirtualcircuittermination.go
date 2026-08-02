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

var _ CircuitsVirtualcircuitterminationDao = (*circuitsVirtualcircuitterminationDao)(nil)

// CircuitsVirtualcircuitterminationDao defining the dao interface
type CircuitsVirtualcircuitterminationDao interface {
	Create(ctx context.Context, table *model.CircuitsVirtualcircuittermination) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.CircuitsVirtualcircuittermination) error
	GetByID(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuittermination, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.CircuitsVirtualcircuittermination, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.CircuitsVirtualcircuittermination, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuittermination, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.CircuitsVirtualcircuittermination, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsVirtualcircuittermination) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsVirtualcircuittermination) error
}

type circuitsVirtualcircuitterminationDao struct {
	db    *gorm.DB
	cache cache.CircuitsVirtualcircuitterminationCache // if nil, the cache is not used.
	sfg   *singleflight.Group                          // if cache is nil, the sfg is not used.
}

// NewCircuitsVirtualcircuitterminationDao creating the dao interface
func NewCircuitsVirtualcircuitterminationDao(db *gorm.DB, xCache cache.CircuitsVirtualcircuitterminationCache) CircuitsVirtualcircuitterminationDao {
	if xCache == nil {
		return &circuitsVirtualcircuitterminationDao{db: db}
	}
	return &circuitsVirtualcircuitterminationDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *circuitsVirtualcircuitterminationDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new circuitsVirtualcircuittermination, insert the record and the id value is written back to the table
func (d *circuitsVirtualcircuitterminationDao) Create(ctx context.Context, table *model.CircuitsVirtualcircuittermination) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a circuitsVirtualcircuittermination by id
func (d *circuitsVirtualcircuitterminationDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.CircuitsVirtualcircuittermination{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a circuitsVirtualcircuittermination by ids
func (d *circuitsVirtualcircuitterminationDao) UpdateByID(ctx context.Context, table *model.CircuitsVirtualcircuittermination) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *circuitsVirtualcircuitterminationDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.CircuitsVirtualcircuittermination) error {
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
	if table.Role != "" {
		update["role"] = table.Role
	}
	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.InterfaceID != 0 {
		update["interface_id"] = table.InterfaceID
	}
	if table.VirtualCircuitID != 0 {
		update["virtual_circuit_id"] = table.VirtualCircuitID
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a circuitsVirtualcircuittermination by id
func (d *circuitsVirtualcircuitterminationDao) GetByID(ctx context.Context, id uint64) (*model.CircuitsVirtualcircuittermination, error) {
	// no cache
	if d.cache == nil {
		record := &model.CircuitsVirtualcircuittermination{}
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
			table := &model.CircuitsVirtualcircuittermination{}
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
			if err = d.cache.Set(ctx, id, table, cache.CircuitsVirtualcircuitterminationExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.CircuitsVirtualcircuittermination)
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

// GetByColumns get a paginated list of circuitsVirtualcircuitterminations by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *circuitsVirtualcircuitterminationDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.CircuitsVirtualcircuittermination, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.CircuitsVirtualcircuitterminationColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.CircuitsVirtualcircuittermination{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.CircuitsVirtualcircuittermination{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete circuitsVirtualcircuittermination by ids
func (d *circuitsVirtualcircuitterminationDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.CircuitsVirtualcircuittermination{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a circuitsVirtualcircuittermination by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *circuitsVirtualcircuitterminationDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.CircuitsVirtualcircuittermination, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.CircuitsVirtualcircuitterminationColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.CircuitsVirtualcircuittermination{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs Batch get circuitsVirtualcircuittermination by ids
func (d *circuitsVirtualcircuitterminationDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.CircuitsVirtualcircuittermination, error) {
	// no cache
	if d.cache == nil {
		var records []*model.CircuitsVirtualcircuittermination
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.CircuitsVirtualcircuittermination)
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
			var records []*model.CircuitsVirtualcircuittermination
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
				if err = d.cache.MultiSet(ctx, records, cache.CircuitsVirtualcircuitterminationExpireTime); err != nil {
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

// GetByLastID Get a paginated list of circuitsVirtualcircuitterminations by last id
func (d *circuitsVirtualcircuitterminationDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.CircuitsVirtualcircuittermination, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.CircuitsVirtualcircuittermination{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *circuitsVirtualcircuitterminationDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsVirtualcircuittermination) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *circuitsVirtualcircuitterminationDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.CircuitsVirtualcircuittermination{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *circuitsVirtualcircuitterminationDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CircuitsVirtualcircuittermination) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
