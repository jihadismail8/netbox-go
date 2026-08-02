package dao

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"

	"netbox-go/internal/cache"
	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

var _ ThumbnailKvstoreDao = (*thumbnailKvstoreDao)(nil)

// ThumbnailKvstoreDao defining the dao interface
type ThumbnailKvstoreDao interface {
	Create(ctx context.Context, table *model.ThumbnailKvstore) error
	DeleteByKey(ctx context.Context, key string) error
	UpdateByKey(ctx context.Context, table *model.ThumbnailKvstore) error
	GetByKey(ctx context.Context, key string) (*model.ThumbnailKvstore, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.ThumbnailKvstore, int64, error)

	DeleteByKeys(ctx context.Context, keys []string) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.ThumbnailKvstore, error)
	GetByKeys(ctx context.Context, keys []string) (map[string]*model.ThumbnailKvstore, error)
	GetByLastKey(ctx context.Context, lastKey string, limit int, sort string) ([]*model.ThumbnailKvstore, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ThumbnailKvstore) (string, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, key string) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ThumbnailKvstore) error
}

type thumbnailKvstoreDao struct {
	db    *gorm.DB
	cache cache.ThumbnailKvstoreCache // if nil, the cache is not used.
	sfg   *singleflight.Group         // if cache is nil, the sfg is not used.
}

// NewThumbnailKvstoreDao creating the dao interface
func NewThumbnailKvstoreDao(db *gorm.DB, xCache cache.ThumbnailKvstoreCache) ThumbnailKvstoreDao {
	if xCache == nil {
		return &thumbnailKvstoreDao{db: db}
	}
	return &thumbnailKvstoreDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *thumbnailKvstoreDao) deleteCache(ctx context.Context, key string) error {
	if d.cache != nil {
		return d.cache.Del(ctx, key)
	}
	return nil
}

// Create a new thumbnailKvstore, insert the record and the key value is written back to the table
func (d *thumbnailKvstoreDao) Create(ctx context.Context, table *model.ThumbnailKvstore) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByKey delete a thumbnailKvstore by key
func (d *thumbnailKvstoreDao) DeleteByKey(ctx context.Context, key string) error {
	err := d.db.WithContext(ctx).Where("key = ?", key).Delete(&model.ThumbnailKvstore{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, key)

	return nil
}

// UpdateByKey update a thumbnailKvstore by key
func (d *thumbnailKvstoreDao) UpdateByKey(ctx context.Context, table *model.ThumbnailKvstore) error {
	err := d.updateDataByKey(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.Key)

	return err
}

func (d *thumbnailKvstoreDao) updateDataByKey(ctx context.Context, db *gorm.DB, table *model.ThumbnailKvstore) error {
	if table.Key == "" {
		return errors.New("key cannot be empty")
	}

	update := map[string]interface{}{}

	if table.Key != "" {
		update["key"] = table.Key
	}
	if table.Value != "" {
		update["value"] = table.Value
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByKey get a thumbnailKvstore by key
func (d *thumbnailKvstoreDao) GetByKey(ctx context.Context, key string) (*model.ThumbnailKvstore, error) {
	// no cache
	if d.cache == nil {
		record := &model.ThumbnailKvstore{}
		err := d.db.WithContext(ctx).Where("key = ?", key).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, key)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same key, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(key, func() (interface{}, error) {

			table := &model.ThumbnailKvstore{}
			err = d.db.WithContext(ctx).Where("key = ?", key).First(table).Error
			if err != nil {
				// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, database.ErrRecordNotFound) {
					if err = d.cache.SetPlaceholder(ctx, key); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("key", key))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, key, table, cache.ThumbnailKvstoreExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("key", key))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.ThumbnailKvstore)
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

// GetByColumns get a paginated list of thumbnailKvstores by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *thumbnailKvstoreDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.ThumbnailKvstore, int64, error) {
	if params.Sort == "" {
		params.Sort = "-key"
	}
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.ThumbnailKvstoreColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.ThumbnailKvstore{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.ThumbnailKvstore{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByKeys batch delete thumbnailKvstores by keys
func (d *thumbnailKvstoreDao) DeleteByKeys(ctx context.Context, keys []string) error {
	err := d.db.WithContext(ctx).Where("key IN (?)", keys).Delete(&model.ThumbnailKvstore{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, key := range keys {
		_ = d.deleteCache(ctx, key)
	}

	return nil
}

// GetByCondition get a thumbnailKvstore by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *thumbnailKvstoreDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.ThumbnailKvstore, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.ThumbnailKvstoreColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.ThumbnailKvstore{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByKeys batch get thumbnailKvstores by keys
func (d *thumbnailKvstoreDao) GetByKeys(ctx context.Context, keys []string) (map[string]*model.ThumbnailKvstore, error) {
	// no cache
	if d.cache == nil {
		var records []*model.ThumbnailKvstore
		err := d.db.WithContext(ctx).Where("key IN (?)", keys).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[string]*model.ThumbnailKvstore)
		for _, record := range records {
			itemMap[record.Key] = record
		}
		return itemMap, nil
	}

	// get form cache
	itemMap, err := d.cache.MultiGet(ctx, keys)
	if err != nil {
		return nil, err
	}

	var missedKeys []string
	for _, key := range keys {
		if _, ok := itemMap[key]; !ok {
			missedKeys = append(missedKeys, key)
		}
	}

	// get missed data
	if len(missedKeys) > 0 {
		// find the key of an active placeholder, i.e. an key that does not exist in database
		var realMissedKeys []string
		for _, key := range missedKeys {
			_, err = d.cache.Get(ctx, key)
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissedKeys = append(realMissedKeys, key)
		}

		if len(realMissedKeys) > 0 {
			var records []*model.ThumbnailKvstore
			var recordKeyMap = make(map[string]struct{})
			err = d.db.WithContext(ctx).Where("key IN (?)", realMissedKeys).Find(&records).Error
			if err != nil {
				return nil, err
			}

			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.Key] = record
					recordKeyMap[record.Key] = struct{}{}
				}
				err = d.cache.MultiSet(ctx, records, cache.ThumbnailKvstoreExpireTime)
				if err != nil {
					logger.Warn("cache.MultiSet error", logger.Err(err), logger.Any("keys", records))
				}
				if len(records) == len(realMissedKeys) {
					return itemMap, nil
				}
			}
			for _, key := range realMissedKeys {
				if _, ok := recordKeyMap[key]; !ok {
					if err = d.cache.SetPlaceholder(ctx, key); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("key", key))
					}
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLastKey get a paginated list of thumbnailKvstores by last key
func (d *thumbnailKvstoreDao) GetByLastKey(ctx context.Context, lastKey string, limit int, sort string) ([]*model.ThumbnailKvstore, error) {
	if sort == "" {
		sort = "-key"
	}
	page := query.NewPage(0, limit, sort)

	records := []*model.ThumbnailKvstore{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("key < ?", lastKey).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *thumbnailKvstoreDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ThumbnailKvstore) (string, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.Key, err
}

// DeleteByTx delete a record by key in the database using the provided transaction
func (d *thumbnailKvstoreDao) DeleteByTx(ctx context.Context, tx *gorm.DB, key string) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.ThumbnailKvstore{}).Where("key = ?", key).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, key)

	return nil
}

// UpdateByTx update a record by key in the database using the provided transaction
func (d *thumbnailKvstoreDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ThumbnailKvstore) error {
	err := d.updateDataByKey(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.Key)

	return err
}
