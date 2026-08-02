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

var _ DjangoSessionDao = (*djangoSessionDao)(nil)

// DjangoSessionDao defining the dao interface
type DjangoSessionDao interface {
	Create(ctx context.Context, table *model.DjangoSession) error
	DeleteBySessionKey(ctx context.Context, sessionKey string) error
	UpdateBySessionKey(ctx context.Context, table *model.DjangoSession) error
	GetBySessionKey(ctx context.Context, sessionKey string) (*model.DjangoSession, error)
	GetByColumns(ctx context.Context, params *query.Params) ([]*model.DjangoSession, int64, error)

	DeleteBySessionKeys(ctx context.Context, sessionKeys []string) error
	GetByCondition(ctx context.Context, condition *query.Conditions) (*model.DjangoSession, error)
	GetBySessionKeys(ctx context.Context, sessionKeys []string) (map[string]*model.DjangoSession, error)
	GetByLastSessionKey(ctx context.Context, lastSessionKey string, limit int, sort string) ([]*model.DjangoSession, error)

	CreateByTx(ctx context.Context, tx *gorm.DB, table *model.DjangoSession) (string, error)
	DeleteByTx(ctx context.Context, tx *gorm.DB, sessionKey string) error
	UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.DjangoSession) error
}

type djangoSessionDao struct {
	db    *gorm.DB
	cache cache.DjangoSessionCache // if nil, the cache is not used.
	sfg   *singleflight.Group      // if cache is nil, the sfg is not used.
}

// NewDjangoSessionDao creating the dao interface
func NewDjangoSessionDao(db *gorm.DB, xCache cache.DjangoSessionCache) DjangoSessionDao {
	if xCache == nil {
		return &djangoSessionDao{db: db}
	}
	return &djangoSessionDao{
		db:    db,
		cache: xCache,
		sfg:   new(singleflight.Group),
	}
}

func (d *djangoSessionDao) deleteCache(ctx context.Context, sessionKey string) error {
	if d.cache != nil {
		return d.cache.Del(ctx, sessionKey)
	}
	return nil
}

// Create a new djangoSession, insert the record and the sessionKey value is written back to the table
func (d *djangoSessionDao) Create(ctx context.Context, table *model.DjangoSession) error {
	return d.db.WithContext(ctx).Create(table).Error
}

// DeleteBySessionKey delete a djangoSession by sessionKey
func (d *djangoSessionDao) DeleteBySessionKey(ctx context.Context, sessionKey string) error {
	err := d.db.WithContext(ctx).Where("session_key = ?", sessionKey).Delete(&model.DjangoSession{}).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, sessionKey)

	return nil
}

// UpdateBySessionKey update a djangoSession by sessionKey
func (d *djangoSessionDao) UpdateBySessionKey(ctx context.Context, table *model.DjangoSession) error {
	err := d.updateDataBySessionKey(ctx, d.db, table)

	// delete cache
	_ = d.deleteCache(ctx, table.SessionKey)

	return err
}

func (d *djangoSessionDao) updateDataBySessionKey(ctx context.Context, db *gorm.DB, table *model.DjangoSession) error {
	if table.SessionKey == "" {
		return errors.New("sessionKey cannot be empty")
	}

	update := map[string]interface{}{}

	if table.SessionKey != "" {
		update["session_key"] = table.SessionKey
	}
	if table.SessionData != "" {
		update["session_data"] = table.SessionData
	}
	if table.ExpireDate != nil && table.ExpireDate.IsZero() == false {
		update["expire_date"] = table.ExpireDate
	}

	return db.WithContext(ctx).Model(table).Updates(update).Error
}

// GetBySessionKey get a djangoSession by sessionKey
func (d *djangoSessionDao) GetBySessionKey(ctx context.Context, sessionKey string) (*model.DjangoSession, error) {
	// no cache
	if d.cache == nil {
		record := &model.DjangoSession{}
		err := d.db.WithContext(ctx).Where("session_key = ?", sessionKey).First(record).Error
		return record, err
	}

	// get from cache
	record, err := d.cache.Get(ctx, sessionKey)
	if err == nil {
		return record, nil
	}

	// get from database
	if errors.Is(err, database.ErrCacheNotFound) {
		// for the same sessionKey, prevent high concurrent simultaneous access to database
		val, err, _ := d.sfg.Do(sessionKey, func() (interface{}, error) {

			table := &model.DjangoSession{}
			err = d.db.WithContext(ctx).Where("session_key = ?", sessionKey).First(table).Error
			if err != nil {
				// set placeholder cache to prevent cache penetration, default expiration time 10 minutes
				if errors.Is(err, database.ErrRecordNotFound) {
					if err = d.cache.SetPlaceholder(ctx, sessionKey); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("sessionKey", sessionKey))
					}
					return nil, database.ErrRecordNotFound
				}
				return nil, err
			}
			// set cache
			if err = d.cache.Set(ctx, sessionKey, table, cache.DjangoSessionExpireTime); err != nil {
				logger.Warn("cache.Set error", logger.Err(err), logger.Any("sessionKey", sessionKey))
			}
			return table, nil
		})
		if err != nil {
			return nil, err
		}
		table, ok := val.(*model.DjangoSession)
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

// GetByColumns get a paginated list of djangoSessions by custom conditions.
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html
func (d *djangoSessionDao) GetByColumns(ctx context.Context, params *query.Params) ([]*model.DjangoSession, int64, error) {
	if params.Sort == "" {
		params.Sort = "-session_key"
	}
	queryStr, args, err := params.ConvertToGormConditions(query.WithWhitelistNames(model.DjangoSessionColumnNames))
	if err != nil {
		return nil, 0, errors.New("query params error: " + err.Error())
	}

	var total int64
	if params.Sort != "ignore count" { // determine if count is required
		err = d.db.WithContext(ctx).Model(&model.DjangoSession{}).Where(queryStr, args...).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return nil, total, nil
		}
	}

	records := []*model.DjangoSession{}
	order, limit, offset := params.ConvertToPage()
	err = d.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Where(queryStr, args...).Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, err
}

// DeleteBySessionKeys batch delete djangoSessions by sessionKeys
func (d *djangoSessionDao) DeleteBySessionKeys(ctx context.Context, sessionKeys []string) error {
	err := d.db.WithContext(ctx).Where("session_key IN (?)", sessionKeys).Delete(&model.DjangoSession{}).Error
	if err != nil {
		return err
	}

	// delete cache
	for _, sessionKey := range sessionKeys {
		_ = d.deleteCache(ctx, sessionKey)
	}

	return nil
}

// GetByCondition get a djangoSession by custom condition
// For more details, please refer to https://go-sponge.com/component/data/custom-page-query.html#_2-condition-parameters-optional
func (d *djangoSessionDao) GetByCondition(ctx context.Context, c *query.Conditions) (*model.DjangoSession, error) {
	queryStr, args, err := c.ConvertToGorm(query.WithWhitelistNames(model.DjangoSessionColumnNames))
	if err != nil {
		return nil, err
	}

	table := &model.DjangoSession{}
	err = d.db.WithContext(ctx).Where(queryStr, args...).First(table).Error
	if err != nil {
		return nil, err
	}

	return table, nil
}

// GetBySessionKeys batch get djangoSessions by sessionKeys
func (d *djangoSessionDao) GetBySessionKeys(ctx context.Context, sessionKeys []string) (map[string]*model.DjangoSession, error) {
	// no cache
	if d.cache == nil {
		var records []*model.DjangoSession
		err := d.db.WithContext(ctx).Where("session_key IN (?)", sessionKeys).Find(&records).Error
		if err != nil {
			return nil, err
		}
		itemMap := make(map[string]*model.DjangoSession)
		for _, record := range records {
			itemMap[record.SessionKey] = record
		}
		return itemMap, nil
	}

	// get form cache
	itemMap, err := d.cache.MultiGet(ctx, sessionKeys)
	if err != nil {
		return nil, err
	}

	var missedSessionKeys []string
	for _, sessionKey := range sessionKeys {
		if _, ok := itemMap[sessionKey]; !ok {
			missedSessionKeys = append(missedSessionKeys, sessionKey)
		}
	}

	// get missed data
	if len(missedSessionKeys) > 0 {
		// find the sessionKey of an active placeholder, i.e. an sessionKey that does not exist in database
		var realMissedSessionKeys []string
		for _, sessionKey := range missedSessionKeys {
			_, err = d.cache.Get(ctx, sessionKey)
			if d.cache.IsPlaceholderErr(err) {
				continue
			}
			realMissedSessionKeys = append(realMissedSessionKeys, sessionKey)
		}

		if len(realMissedSessionKeys) > 0 {
			var records []*model.DjangoSession
			var recordSessionKeyMap = make(map[string]struct{})
			err = d.db.WithContext(ctx).Where("session_key IN (?)", realMissedSessionKeys).Find(&records).Error
			if err != nil {
				return nil, err
			}

			if len(records) > 0 {
				for _, record := range records {
					itemMap[record.SessionKey] = record
					recordSessionKeyMap[record.SessionKey] = struct{}{}
				}
				err = d.cache.MultiSet(ctx, records, cache.DjangoSessionExpireTime)
				if err != nil {
					logger.Warn("cache.MultiSet error", logger.Err(err), logger.Any("sessionKeys", records))
				}
				if len(records) == len(realMissedSessionKeys) {
					return itemMap, nil
				}
			}
			for _, sessionKey := range realMissedSessionKeys {
				if _, ok := recordSessionKeyMap[sessionKey]; !ok {
					if err = d.cache.SetPlaceholder(ctx, sessionKey); err != nil {
						logger.Warn("cache.SetPlaceholder error", logger.Err(err), logger.Any("sessionKey", sessionKey))
					}
				}
			}
		}
	}

	return itemMap, nil
}

// GetByLastSessionKey get a paginated list of djangoSessions by last sessionKey
func (d *djangoSessionDao) GetByLastSessionKey(ctx context.Context, lastSessionKey string, limit int, sort string) ([]*model.DjangoSession, error) {
	if sort == "" {
		sort = "-session_key"
	}
	page := query.NewPage(0, limit, sort)

	records := []*model.DjangoSession{}
	err := d.db.WithContext(ctx).Order(page.Sort()).Limit(page.Limit()).Where("session_key < ?", lastSessionKey).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// CreateByTx create a record in the database using the provided transaction
func (d *djangoSessionDao) CreateByTx(ctx context.Context, tx *gorm.DB, table *model.DjangoSession) (string, error) {
	err := tx.WithContext(ctx).Create(table).Error
	return table.SessionKey, err
}

// DeleteByTx delete a record by sessionKey in the database using the provided transaction
func (d *djangoSessionDao) DeleteByTx(ctx context.Context, tx *gorm.DB, sessionKey string) error {
	update := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	err := tx.WithContext(ctx).Model(&model.DjangoSession{}).Where("session_key = ?", sessionKey).Updates(update).Error
	if err != nil {
		return err
	}

	// delete cache
	_ = d.deleteCache(ctx, sessionKey)

	return nil
}

// UpdateByTx update a record by sessionKey in the database using the provided transaction
func (d *djangoSessionDao) UpdateByTx(ctx context.Context, tx *gorm.DB, table *model.DjangoSession) error {
	err := d.updateDataBySessionKey(ctx, tx, table)

	// delete cache
	_ = d.deleteCache(ctx, table.SessionKey)

	return err
}
