package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

func newDjangoMigrationsCache() *gotest.Cache {
	record1 := &model.DjangoMigrations{}
	record1.ID = 1
	record2 := &model.DjangoMigrations{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDjangoMigrationsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_djangoMigrationsCache_Set(t *testing.T) {
	c := newDjangoMigrationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DjangoMigrations)
	err := c.ICache.(DjangoMigrationsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DjangoMigrationsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_djangoMigrationsCache_Get(t *testing.T) {
	c := newDjangoMigrationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DjangoMigrations)
	err := c.ICache.(DjangoMigrationsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DjangoMigrationsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DjangoMigrationsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_djangoMigrationsCache_MultiGet(t *testing.T) {
	c := newDjangoMigrationsCache()
	defer c.Close()

	var testData []*model.DjangoMigrations
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DjangoMigrations))
	}

	err := c.ICache.(DjangoMigrationsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DjangoMigrationsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DjangoMigrations))
	}
}

func Test_djangoMigrationsCache_MultiSet(t *testing.T) {
	c := newDjangoMigrationsCache()
	defer c.Close()

	var testData []*model.DjangoMigrations
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DjangoMigrations))
	}

	err := c.ICache.(DjangoMigrationsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_djangoMigrationsCache_Del(t *testing.T) {
	c := newDjangoMigrationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DjangoMigrations)
	err := c.ICache.(DjangoMigrationsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_djangoMigrationsCache_SetCacheWithNotFound(t *testing.T) {
	c := newDjangoMigrationsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DjangoMigrations)
	err := c.ICache.(DjangoMigrationsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DjangoMigrationsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDjangoMigrationsCache(t *testing.T) {
	c := NewDjangoMigrationsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDjangoMigrationsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDjangoMigrationsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
