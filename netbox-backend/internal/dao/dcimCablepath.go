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

var _ DcimCablepathDao = (*dcimCablepathDao)(nil)

// DcimCablepathDao defining the dao interface
type DcimCablepathDao interface {
	Create(ctx context.Context, table *model.DcimCablepath) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.DcimCablepath) error
	GetByID(ctx context.Context, id uint64) (*model.DcimCablepath, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.DcimCablepath, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.DcimCablepath, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.DcimCablepath, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.DcimCablepath, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimCablepath) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimCablepath) error
}

type dcimCablepathDao struct {
	db    *gorm.DB
	cache cache.DcimCablepathCache // if nil, the cache is not used.
	sfg   *singleflight.Group      // if cache is nil, the sfg is not used.
}

// NewDcimCablepathDao creating the dao interface
func NewDcimCablepathDao(db *gorm.DB, xCache cache.DcimCablepathCache) DcimCablepathDao {
	if xCache == nil {
		return &dcimCablepathDao{db: db}
	}
	return &dcimCablepathDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *dcimCablepathDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new dcimCablepath, insert the record and the id value is written back to the table
func (d *dcimCablepathDao) Create(ctx context.Context, table *model.DcimCablepath) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a dcimCablepath by id
func (d *dcimCablepathDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.DcimCablepath{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a dcimCablepath by ids
func (d *dcimCablepathDao) UpdateByID(ctx context.Context, table *model.DcimCablepath) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *dcimCablepathDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.DcimCablepath) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.XNodes != "" {
		update["_nodes"] = table.XNodes
	}
	if table.IsActive != nil {
		update["is_active"] = table.IsActive
	}
	if table.IsSplit != nil {
		update["is_split"] = table.IsSplit
	}
	if table.Path != nil && table.Path.String() != "" {
		update["path"] = table.Path
	}
	if table.IsComplete != nil {
		update["is_complete"] = table.IsComplete
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a dcimCablepath by id
func (d *dcimCablepathDao) GetByID(ctx context.Context, id uint64) (*model.DcimCablepath, error) {
	// no cache
	if d.cache == nil {
		record := &model.DcimCablepath{}
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
			table := &model.DcimCablepath{}
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
			if err = d.cache.Set(ctx, id, table, cache.DcimCablepathExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.DcimCablepath)
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

// GetByColumns get a paginated list of dcimCablepaths by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *dcimCablepathDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.DcimCablepath, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.DcimCablepathColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.DcimCablepath{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.DcimCablepath{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete dcimCablepath by ids
func (d *dcimCablepathDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.DcimCablepath{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a dcimCablepath by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *dcimCablepathDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.DcimCablepath, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.DcimCablepathColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.DcimCablepath{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs Batch get dcimCablepath by ids
func (d *dcimCablepathDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.DcimCablepath, error) {
	// no cache
	if d.cache == nil {
		var records []*model.DcimCablepath
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.DcimCablepath)
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
			var records []*model.DcimCablepath
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
				if err = d.cache.MultiSet(ctx, records, cache.DcimCablepathExpireTime); err != nil {
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

// GetByLastID Get a paginated list of dcimCablepaths by last id
func (d *dcimCablepathDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.DcimCablepath, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.DcimCablepath{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *dcimCablepathDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimCablepath) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *dcimCablepathDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.DcimCablepath{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *dcimCablepathDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.DcimCablepath) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
