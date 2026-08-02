package dao

import (
	"context"
	"errors"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"

	"netbox-go/internal/cache"
	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

var _ ExtrasCachedvalueDao = (*extrasCachedvalueDao)(nil)

// ExtrasCachedvalueDao defining the dao interface
type ExtrasCachedvalueDao interface {
	Create(ctx context.Context, table *model.ExtrasCachedvalue) error
	DeleteByID(ctx context.Context, id string) error
	UpdateByID(ctx context.Context, table *model.ExtrasCachedvalue) error
	GetByID(ctx context.Context, id string) (*model.ExtrasCachedvalue, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.ExtrasCachedvalue, int64, error)

	DeleteByIDs(ctx context.Context, ids []string) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.ExtrasCachedvalue, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]*model.ExtrasCachedvalue, error)
	GetByLastID(ctx context.Context, lastID string, limit int, sort string) ([]*model.ExtrasCachedvalue, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasCachedvalue) (string, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id string) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasCachedvalue) error
}

type extrasCachedvalueDao struct {
	db    *gorm.DB
	cache cache.ExtrasCachedvalueCache // if nil, the cache is not used.
	sfg   *singleflight.Group          // if cache is nil, the sfg is not used.
}

// NewExtrasCachedvalueDao creating the dao interface
func NewExtrasCachedvalueDao(db *gorm.DB, xCache cache.ExtrasCachedvalueCache) ExtrasCachedvalueDao {
	if xCache == nil {
		return &extrasCachedvalueDao{db: db}
	}
	return &extrasCachedvalueDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *extrasCachedvalueDao) deleteCache(ctx context.Context, id string) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new extrasCachedvalue, insert the record and the id value is written back to the table
func (d *extrasCachedvalueDao) Create(ctx context.Context, table *model.ExtrasCachedvalue) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a extrasCachedvalue by id
func (d *extrasCachedvalueDao) DeleteByID(ctx context.Context, id string) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ExtrasCachedvalue{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a extrasCachedvalue by id
func (d *extrasCachedvalueDao) UpdateByID(ctx context.Context, table *model.ExtrasCachedvalue) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *extrasCachedvalueDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.ExtrasCachedvalue) error {
	if table.ID == "" {
		return errors.New("id cannot be empty")
	}

	update := map[string]interface{}{}

	if table.Timestamp != nil && table.Timestamp.IsZero() == false {
		update["timestamp"] = table.Timestamp
	}
	if table.ObjectID != 0 {
		update["object_id"] = table.ObjectID
	}
	if table.Field != "" {
		update["field"] = table.Field
	}
	if table.Type != "" {
		update["type"] = table.Type
	}
	if table.Value != "" {
		update["value"] = table.Value
	}
	if table.Weight != 0 {
		update["weight"] = table.Weight
	}
	if table.ObjectTypeID != 0 {
		update["object_type_id"] = table.ObjectTypeID
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a extrasCachedvalue by id
func (d *extrasCachedvalueDao) GetByID(ctx context.Context, id string) (*model.ExtrasCachedvalue, error) {
	// no cache
	if d.cache == nil {
		record := &model.ExtrasCachedvalue{}
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
		val, err, _ := d.sfg.Do(id, func() (interface{}, error) {

			table := &model.ExtrasCachedvalue{}
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
			if err = d.cache.Set(ctx, id, table, cache.ExtrasCachedvalueExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.ExtrasCachedvalue)
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

// GetByColumns get a paginated list of extrasCachedvalues by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *extrasCachedvalueDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.ExtrasCachedvalue, int64, error) {
	if params.Sort == "" {
		params.Sort = "-id"
	}
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.ExtrasCachedvalueColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.ExtrasCachedvalue{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.ExtrasCachedvalue{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete extrasCachedvalues by ids
func (d *extrasCachedvalueDao) DeleteByIDs(ctx context.Context, ids []string) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.ExtrasCachedvalue{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a extrasCachedvalue by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *extrasCachedvalueDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.ExtrasCachedvalue, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.ExtrasCachedvalueColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.ExtrasCachedvalue{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs batch get extrasCachedvalues by ids
func (d *extrasCachedvalueDao) GetByIDs(ctx context.Context, ids []string) (map[string]*model.ExtrasCachedvalue, error) {
	// no cache
	if d.cache == nil {
		var records []*model.ExtrasCachedvalue
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[string]*model.ExtrasCachedvalue)
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

	var missedIDs []string
	for _, id := range ids {
		if _, ok := itemMap[id]; !ok {
			missedIDs = append(missedIDs, id)
		}
	}

	// get missed data
	if len(missedIDs) > 0 {
		// find the id of an active placeholder, i.e. an id that does not exist in database
		var realMissedIDs []string
		for _, id := range missedIDs {
			_, err = d.cache.Get(ctx, id)
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissedIDs = append(realMissedIDs, id)
		}

		if len(realMissedIDs) > 0 {
			var records []*model.ExtrasCachedvalue
			var recordIDMap = make(map[string]struct{})
			err = d.db.WithContext(ctx).Where("id IN (?)", realMissedIDs).Find(&records).Error
			if err != nil {
				return nil, err
			}

			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.ID] = record
					recordIDMap[record.ID] = struct{}{}
				}
				err = d.cache.MultiSet(ctx, records, cache.ExtrasCachedvalueExpireTime)
				if err != nil {
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

// GetByLastID get a paginated list of extrasCachedvalues by last id
func (d *extrasCachedvalueDao) GetByLastID(ctx context.Context, lastID string, limit int, sort string) ([]*model.ExtrasCachedvalue, error) {
	if sort == "" {
		sort = "-id"
	}
	page := query.NewPage(0, limit, sort)

	records := []*model.ExtrasCachedvalue{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *extrasCachedvalueDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasCachedvalue) (string, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *extrasCachedvalueDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id string) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.ExtrasCachedvalue{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *extrasCachedvalueDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.ExtrasCachedvalue) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
