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

var _ DcimRearporttemplateDao = (*dcimRearporttemplateDao)(nil)

// DcimRearporttemplateDao defining the dao interface
type DcimRearporttemplateDao interface {
	Create(ctx context.Context, table *model.DcimRearporttemplate) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.DcimRearporttemplate) error
	GetByID(ctx context.Context, id uint64) (*model.DcimRearporttemplate, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.DcimRearporttemplate, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.DcimRearporttemplate, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRearporttemplate, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.DcimRearporttemplate, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimRearporttemplate) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimRearporttemplate) error
}

type dcimRearporttemplateDao struct {
	db    *gorm.DB
	cache cache.DcimRearporttemplateCache // if nil, the cache is not used.
	sfg   *singleflight.Group             // if cache is nil, the sfg is not used.
}

// NewDcimRearporttemplateDao creating the dao interface
func NewDcimRearporttemplateDao(db *gorm.DB, xCache cache.DcimRearporttemplateCache) DcimRearporttemplateDao {
	if xCache == nil {
		return &dcimRearporttemplateDao{db: db}
	}
	return &dcimRearporttemplateDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *dcimRearporttemplateDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new dcimRearporttemplate, insert the record and the id value is written back to the table
func (d *dcimRearporttemplateDao) Create(ctx context.Context, table *model.DcimRearporttemplate) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a dcimRearporttemplate by id
func (d *dcimRearporttemplateDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.DcimRearporttemplate{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a dcimRearporttemplate by ids
func (d *dcimRearporttemplateDao) UpdateByID(ctx context.Context, table *model.DcimRearporttemplate) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *dcimRearporttemplateDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.DcimRearporttemplate) error {
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
	if table.Label != "" {
		update["label"] = table.Label
	}
	if table.Description != "" {
		update["description"] = table.Description
	}
	if table.Type != "" {
		update["type"] = table.Type
	}
	if table.Positions != 0 {
		update["positions"] = table.Positions
	}
	if table.DeviceTypeID != 0 {
		update["device_type_id"] = table.DeviceTypeID
	}
	if table.Color != "" {
		update["color"] = table.Color
	}
	if table.ModuleTypeID != 0 {
		update["module_type_id"] = table.ModuleTypeID
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a dcimRearporttemplate by id
func (d *dcimRearporttemplateDao) GetByID(ctx context.Context, id uint64) (*model.DcimRearporttemplate, error) {
	// no cache
	if d.cache == nil {
		record := &model.DcimRearporttemplate{}
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
			table := &model.DcimRearporttemplate{}
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
			if err = d.cache.Set(ctx, id, table, cache.DcimRearporttemplateExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.DcimRearporttemplate)
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

// GetByColumns get a paginated list of dcimRearporttemplates by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *dcimRearporttemplateDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.DcimRearporttemplate, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.DcimRearporttemplateColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.DcimRearporttemplate{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.DcimRearporttemplate{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete dcimRearporttemplate by ids
func (d *dcimRearporttemplateDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.DcimRearporttemplate{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a dcimRearporttemplate by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *dcimRearporttemplateDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.DcimRearporttemplate, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.DcimRearporttemplateColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.DcimRearporttemplate{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs Batch get dcimRearporttemplate by ids
func (d *dcimRearporttemplateDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.DcimRearporttemplate, error) {
	// no cache
	if d.cache == nil {
		var records []*model.DcimRearporttemplate
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.DcimRearporttemplate)
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
			var records []*model.DcimRearporttemplate
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
				if err = d.cache.MultiSet(ctx, records, cache.DcimRearporttemplateExpireTime); err != nil {
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

// GetByLastID Get a paginated list of dcimRearporttemplates by last id
func (d *dcimRearporttemplateDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.DcimRearporttemplate, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.DcimRearporttemplate{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *dcimRearporttemplateDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimRearporttemplate) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *dcimRearporttemplateDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.DcimRearporttemplate{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *dcimRearporttemplateDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimRearporttemplate) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
