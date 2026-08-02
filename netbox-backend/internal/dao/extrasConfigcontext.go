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

var _ ExtrasConfigcontextDao = (*extrasConfigcontextDao)(nil)

// ExtrasConfigcontextDao defining the dao interface
type ExtrasConfigcontextDao interface {
	Create(ctx context.Context, table *model.ExtrasConfigcontext) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.ExtrasConfigcontext) error
	GetByID(ctx context.Context, id uint64) (*model.ExtrasConfigcontext, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.ExtrasConfigcontext, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.ExtrasConfigcontext, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontext, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.ExtrasConfigcontext, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasConfigcontext) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasConfigcontext) error
}

type extrasConfigcontextDao struct {
	db    *gorm.DB
	cache cache.ExtrasConfigcontextCache // if nil, the cache is not used.
	sfg   *singleflight.Group            // if cache is nil, the sfg is not used.
}

// NewExtrasConfigcontextDao creating the dao interface
func NewExtrasConfigcontextDao(db *gorm.DB, xCache cache.ExtrasConfigcontextCache) ExtrasConfigcontextDao {
	if xCache == nil {
		return &extrasConfigcontextDao{db: db}
	}
	return &extrasConfigcontextDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *extrasConfigcontextDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new extrasConfigcontext, insert the record and the id value is written back to the table
func (d *extrasConfigcontextDao) Create(ctx context.Context, table *model.ExtrasConfigcontext) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a extrasConfigcontext by id
func (d *extrasConfigcontextDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ExtrasConfigcontext{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a extrasConfigcontext by ids
func (d *extrasConfigcontextDao) UpdateByID(ctx context.Context, table *model.ExtrasConfigcontext) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *extrasConfigcontextDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.ExtrasConfigcontext) error {
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
	if table.Name != "" {
		update["name"] = table.Name
	}
	if table.Weight != 0 {
		update["weight"] = table.Weight
	}
	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.IsActive != nil {
		update["is_active"] = table.IsActive
	}
	if table.Data != nil && table.Data.String() != "" {
		update["data"] = table.Data
	}
	if table.DataFileID != 0 {
		update["data_file_id"] = table.DataFileID
	}
	if table.DataPath != "" {
		update["data_path"] = table.DataPath
	}
	if table.DataSourceID != 0 {
		update["data_source_id"] = table.DataSourceID
	}
	if table.AutoSyncEnabled != nil {
		update["auto_sync_enabled"] = table.AutoSyncEnabled
	}
	if table.DataSynced != nil && table.DataSynced.IsZero() == false {
		update["data_synced"] = table.DataSynced
	}
	if table.ProfileID != 0 {
		update["profile_id"] = table.ProfileID
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a extrasConfigcontext by id
func (d *extrasConfigcontextDao) GetByID(ctx context.Context, id uint64) (*model.ExtrasConfigcontext, error) {
	// no cache
	if d.cache == nil {
		record := &model.ExtrasConfigcontext{}
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
			table := &model.ExtrasConfigcontext{}
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
			if err = d.cache.Set(ctx, id, table, cache.ExtrasConfigcontextExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.ExtrasConfigcontext)
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

// GetByColumns get a paginated list of extrasConfigcontexts by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *extrasConfigcontextDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.ExtrasConfigcontext, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.ExtrasConfigcontextColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.ExtrasConfigcontext{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.ExtrasConfigcontext{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete extrasConfigcontext by ids
func (d *extrasConfigcontextDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.ExtrasConfigcontext{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a extrasConfigcontext by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *extrasConfigcontextDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.ExtrasConfigcontext, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.ExtrasConfigcontextColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.ExtrasConfigcontext{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs Batch get extrasConfigcontext by ids
func (d *extrasConfigcontextDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.ExtrasConfigcontext, error) {
	// no cache
	if d.cache == nil {
		var records []*model.ExtrasConfigcontext
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.ExtrasConfigcontext)
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
			var records []*model.ExtrasConfigcontext
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
				if err = d.cache.MultiSet(ctx, records, cache.ExtrasConfigcontextExpireTime); err != nil {
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

// GetByLastID Get a paginated list of extrasConfigcontexts by last id
func (d *extrasConfigcontextDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.ExtrasConfigcontext, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.ExtrasConfigcontext{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *extrasConfigcontextDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasConfigcontext) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *extrasConfigcontextDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.ExtrasConfigcontext{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *extrasConfigcontextDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasConfigcontext) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
