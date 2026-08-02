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

var _ CoreAutosyncrecordDao = (*coreAutosyncrecordDao)(nil)

// CoreAutosyncrecordDao defining the dao interface
type CoreAutosyncrecordDao interface {
	Create(ctx context.Context, table *model.CoreAutosyncrecord) error
	DeleteByID(ctx context.Context, id uint64) error
	UpdateByID(ctx context.Context, table *model.CoreAutosyncrecord) error
	GetByID(ctx context.Context, id uint64) (*model.CoreAutosyncrecord, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.CoreAutosyncrecord, int64, error)

	DeleteByIDs(ctx context.Context, ids []uint64) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.CoreAutosyncrecord, error)
	GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.CoreAutosyncrecord, error)
	GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.CoreAutosyncrecord, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreAutosyncrecord) (uint64, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreAutosyncrecord) error
}

type coreAutosyncrecordDao struct {
	db    *gorm.DB
	cache cache.CoreAutosyncrecordCache // if nil, the cache is not used.
	sfg   *singleflight.Group           // if cache is nil, the sfg is not used.
}

// NewCoreAutosyncrecordDao creating the dao interface
func NewCoreAutosyncrecordDao(db *gorm.DB, xCache cache.CoreAutosyncrecordCache) CoreAutosyncrecordDao {
	if xCache == nil {
		return &coreAutosyncrecordDao{db: db}
	}
	return &coreAutosyncrecordDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *coreAutosyncrecordDao) deleteCache(ctx context.Context, id uint64) error {
	if d.cache != nil {
		return d.cache.Del(ctx, id)
	}
	return nil
}

// Create a new coreAutosyncrecord, insert the record and the id value is written back to the table
func (d *coreAutosyncrecordDao) Create(ctx context.Context, table *model.CoreAutosyncrecord) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByID delete a coreAutosyncrecord by id
func (d *coreAutosyncrecordDao) DeleteByID(ctx context.Context, id uint64) error {
	err := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.CoreAutosyncrecord{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByID update a coreAutosyncrecord by ids
func (d *coreAutosyncrecordDao) UpdateByID(ctx context.Context, table *model.CoreAutosyncrecord) error {
	err := d.updateDataByID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}

func (d *coreAutosyncrecordDao) updateDataByID(ctx context.Context, db *gorm.DB, table *model.CoreAutosyncrecord) error {
	if table.ID < 1 {
		return errors.New("id cannot be 0")
	}

	update := map[string]interface{}{}

	if table.ObjectID != 0 {
		update["object_id"] = table.ObjectID
	}
	if table.DatafileID != 0 {
		update["datafile_id"] = table.DatafileID
	}
	if table.ObjectTypeID != 0 {
		update["object_type_id"] = table.ObjectTypeID
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByID get a coreAutosyncrecord by id
func (d *coreAutosyncrecordDao) GetByID(ctx context.Context, id uint64) (*model.CoreAutosyncrecord, error) {
	// no cache
	if d.cache == nil {
		record := &model.CoreAutosyncrecord{}
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
			table := &model.CoreAutosyncrecord{}
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
			if err = d.cache.Set(ctx, id, table, cache.CoreAutosyncrecordExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("id", id))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.CoreAutosyncrecord)
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

// GetByColumns get a paginated list of coreAutosyncrecords by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *coreAutosyncrecordDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.CoreAutosyncrecord, int64, error) {
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.CoreAutosyncrecordColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.CoreAutosyncrecord{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.CoreAutosyncrecord{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByIDs batch delete coreAutosyncrecord by ids
func (d *coreAutosyncrecordDao) DeleteByIDs(ctx context.Context, ids []uint64) error {
	err := d.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&model.CoreAutosyncrecord{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, id := range ids {
		_ = d.deleteCache(ctx, id)
	}

	return nil
}

// GetByCondition get a coreAutosyncrecord by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *coreAutosyncrecordDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.CoreAutosyncrecord, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.CoreAutosyncrecordColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.CoreAutosyncrecord{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByIDs Batch get coreAutosyncrecord by ids
func (d *coreAutosyncrecordDao) GetByIDs(ctx context.Context, ids []uint64) (map[uint64]*model.CoreAutosyncrecord, error) {
	// no cache
	if d.cache == nil {
		var records []*model.CoreAutosyncrecord
		err := d.db.WithContext(ctx).Where("id IN (?)", ids).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[uint64]*model.CoreAutosyncrecord)
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
			var records []*model.CoreAutosyncrecord
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
				if err = d.cache.MultiSet(ctx, records, cache.CoreAutosyncrecordExpireTime); err != nil {
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

// GetByLastID Get a paginated list of coreAutosyncrecords by last id
func (d *coreAutosyncrecordDao) GetByLastID(ctx context.Context, lastID uint64, limit int, sort string) ([]*model.CoreAutosyncrecord, error) {
	page := query.NewPage(0, limit, sort)

	records := []*model.CoreAutosyncrecord{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("id < ?", lastID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *coreAutosyncrecordDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreAutosyncrecord) (uint64, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ID, err
}

// DeleteByTx delete a record by id in the database using the provided transaction
func (d *coreAutosyncrecordDao) DeleteByTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	err := tx.WithContext(ctx).Where("id = ?", id).Delete(&model.CoreAutosyncrecord{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, id)

	return nil
}

// UpdateByTx update a record by id in the database using the provided transaction
func (d *coreAutosyncrecordDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreAutosyncrecord) error {
	err := d.updateDataByID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ID)

	return err
}
