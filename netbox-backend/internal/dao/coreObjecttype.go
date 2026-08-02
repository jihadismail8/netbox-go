package dao

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"netbox-go/internal/cache"
	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

var _ CoreObjecttypeDao = (*coreObjecttypeDao)(nil)

// CoreObjecttypeDao defining the dao interface
type CoreObjecttypeDao interface {
	Create(ctx context.Context, table *model.CoreObjecttype) error
	DeleteByContenttypePtrID(ctx context.Context, contenttypePtrID int) error
	UpdateByContenttypePtrID(ctx context.Context, table *model.CoreObjecttype) error
	GetByContenttypePtrID(ctx context.Context, contenttypePtrID int) (*model.CoreObjecttype, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.CoreObjecttype, int64, error)

	DeleteByContenttypePtrIDs(ctx context.Context, contenttypePtrIDs []int) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.CoreObjecttype, error)
	GetByContenttypePtrIDs(ctx context.Context, contenttypePtrIDs []int) (map[int]*model.CoreObjecttype, error)
	GetByLastContenttypePtrID(ctx context.Context, lastContenttypePtrID int, limit int, sort string) ([]*model.CoreObjecttype, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreObjecttype) (int, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, contenttypePtrID int) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreObjecttype) error
}

type coreObjecttypeDao struct {
	db    *gorm.DB
	cache cache.CoreObjecttypeCache // if nil, the cache is not used.
	sfg   *singleflight.Group       // if cache is nil, the sfg is not used.
}

// NewCoreObjecttypeDao creating the dao interface
func NewCoreObjecttypeDao(db *gorm.DB, xCache cache.CoreObjecttypeCache) CoreObjecttypeDao {
	if xCache == nil {
		return &coreObjecttypeDao{db: db}
	}
	return &coreObjecttypeDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *coreObjecttypeDao) deleteCache(ctx context.Context, contenttypePtrID int) error {
	if d.cache != nil {
		return d.cache.Del(ctx, contenttypePtrID)
	}
	return nil
}

// Create a new coreObjecttype, insert the record and the contenttypePtrID value is written back to the table
func (d *coreObjecttypeDao) Create(ctx context.Context, table *model.CoreObjecttype) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteByContenttypePtrID delete a coreObjecttype by contenttypePtrID
func (d *coreObjecttypeDao) DeleteByContenttypePtrID(ctx context.Context, contenttypePtrID int) error {
	err := d.db.WithContext(ctx).Where("contenttype_ptr_id = ?", contenttypePtrID).Delete(&model.CoreObjecttype{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, contenttypePtrID)

	return nil
}

// UpdateByContenttypePtrID update a coreObjecttype by contenttypePtrID
func (d *coreObjecttypeDao) UpdateByContenttypePtrID(ctx context.Context, table *model.CoreObjecttype) error {
	err := d.updateDataByContenttypePtrID(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ContenttypePtrID)

	return err
}

func (d *coreObjecttypeDao) updateDataByContenttypePtrID(ctx context.Context, db *gorm.DB, table *model.CoreObjecttype) error {
	if table.ContenttypePtrID < 1 {
		return errors.New("contenttypePtrID cannot be 0")
	}

	update := map[string]interface{}{}

	if table.ContenttypePtrID != 0 {
		update["contenttype_ptr_id"] = table.ContenttypePtrID
	}
	if table.Public != nil {
		update["public"] = table.Public
	}
	if table.Features != "" {
		update["features"] = table.Features
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetByContenttypePtrID get a coreObjecttype by contenttypePtrID
func (d *coreObjecttypeDao) GetByContenttypePtrID(ctx context.Context, contenttypePtrID int) (*model.CoreObjecttype, error) {
	// no cache
	if d.cache == nil {
		record := &model.CoreObjecttype{}
		err := d.db.WithContext(ctx).Where("contenttype_ptr_id = ?", contenttypePtrID).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, contenttypePtrID)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same contenttypePtrID, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(utils.IntToStr(contenttypePtrID), func() (interface{}, error) {

			table := &model.CoreObjecttype{}
			err = d.db.WithContext(ctx).Where("contenttype_ptr_id = ?", contenttypePtrID).First(table).Error
			if err != nil {
				// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, database.ErrRecordNotFound) {
					if err = d.cache.SetPlaceholder(ctx, contenttypePtrID); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("contenttypePtrID", contenttypePtrID))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, contenttypePtrID, table, cache.CoreObjecttypeExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("contenttypePtrID", contenttypePtrID))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.CoreObjecttype)
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

// GetByColumns get a paginated list of coreObjecttypes by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *coreObjecttypeDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.CoreObjecttype, int64, error) {
	if params.Sort == "" {
		params.Sort = "-contenttype_ptr_id"
	}
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.CoreObjecttypeColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.CoreObjecttype{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.CoreObjecttype{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteByContenttypePtrIDs batch delete coreObjecttypes by contenttypePtrIDs
func (d *coreObjecttypeDao) DeleteByContenttypePtrIDs(ctx context.Context, contenttypePtrIDs []int) error {
	err := d.db.WithContext(ctx).Where("contenttype_ptr_id IN (?)", contenttypePtrIDs).Delete(&model.CoreObjecttype{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, contenttypePtrID := range contenttypePtrIDs {
		_ = d.deleteCache(ctx, contenttypePtrID)
	}

	return nil
}

// GetByCondition get a coreObjecttype by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *coreObjecttypeDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.CoreObjecttype, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.CoreObjecttypeColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.CoreObjecttype{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetByContenttypePtrIDs batch get coreObjecttypes by contenttypePtrIDs
func (d *coreObjecttypeDao) GetByContenttypePtrIDs(ctx context.Context, contenttypePtrIDs []int) (map[int]*model.CoreObjecttype, error) {
	// no cache
	if d.cache == nil {
		var records []*model.CoreObjecttype
		err := d.db.WithContext(ctx).Where("contenttype_ptr_id IN (?)", contenttypePtrIDs).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[int]*model.CoreObjecttype)
		for _, record := range records {
			itemMap[record.ContenttypePtrID] = record
		}
		return itemMap, nil
	}

	// get form cache
	itemMap, err := d.cache.MultiGet(ctx, contenttypePtrIDs)
	if err != nil {
		return nil, err
	}

	var missedContenttypePtrIDs []int
	for _, contenttypePtrID := range contenttypePtrIDs {
		if _, ok := itemMap[contenttypePtrID]; !ok {
			missedContenttypePtrIDs = append(missedContenttypePtrIDs, contenttypePtrID)
		}
	}

	// get missed data
	if len(missedContenttypePtrIDs) > 0 {
		// find the contenttypePtrID of an active placeholder, i.e. an contenttypePtrID that does not exist in database
		var realMissedContenttypePtrIDs []int
		for _, contenttypePtrID := range missedContenttypePtrIDs {
			_, err = d.cache.Get(ctx, contenttypePtrID)
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissedContenttypePtrIDs = append(realMissedContenttypePtrIDs, contenttypePtrID)
		}

		if len(realMissedContenttypePtrIDs) > 0 {
			var records []*model.CoreObjecttype
			var recordContenttypePtrIDMap = make(map[int]struct{})
			err = d.db.WithContext(ctx).Where("contenttype_ptr_id IN (?)", realMissedContenttypePtrIDs).Find(&records).Error
			if err != nil {
				return nil, err
			}

			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.ContenttypePtrID] = record
					recordContenttypePtrIDMap[record.ContenttypePtrID] = struct{}{}
				}
				err = d.cache.MultiSet(ctx, records, cache.CoreObjecttypeExpireTime)
				if err != nil {
					logger.Warn("cache.MultiSet error", logger.Err(err), logger.Any("contenttypePtrIDs", records))
				}
				if len(records) == len(realMissedContenttypePtrIDs) {
					return itemMap, nil
				}
			}
			for _, contenttypePtrID := range realMissedContenttypePtrIDs {
				if _, ok := recordContenttypePtrIDMap[contenttypePtrID]; !ok {
					if err = d.cache.SetPlaceholder(ctx, contenttypePtrID); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("contenttypePtrID", contenttypePtrID))
					}
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLastContenttypePtrID get a paginated list of coreObjecttypes by last contenttypePtrID
func (d *coreObjecttypeDao) GetByLastContenttypePtrID(ctx context.Context, lastContenttypePtrID int, limit int, sort string) ([]*model.CoreObjecttype, error) {
	if sort == "" {
		sort = "-contenttype_ptr_id"
	}
	page := query.NewPage(0, limit, sort)

	records := []*model.CoreObjecttype{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("contenttype_ptr_id < ?", lastContenttypePtrID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *coreObjecttypeDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreObjecttype) (int, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.ContenttypePtrID, err
}

// DeleteByTx delete a record by contenttypePtrID in the database using the provided transaction
func (d *coreObjecttypeDao) DeleteByTx(ctx context.Context, tx *gorm.DB, contenttypePtrID int) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.CoreObjecttype{}).Where("contenttype_ptr_id = ?", contenttypePtrID).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, contenttypePtrID)

	return nil
}

// UpdateByTx update a record by contenttypePtrID in the database using the provided transaction
func (d *coreObjecttypeDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.CoreObjecttype) error {
	err := d.updateDataByContenttypePtrID(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.ContenttypePtrID)

	return err
}
