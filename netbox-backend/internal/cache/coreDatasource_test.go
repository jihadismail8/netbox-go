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

func newCoreDatasourceCache() *gotest.Cache {
	record1 := &model.CoreDatasource{}
	record1.ID = 1
	record2 := &model.CoreDatasource{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCoreDatasourceCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_coreDatasourceCache_Set(t *testing.T) {
	c := newCoreDatasourceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreDatasource)
	err := c.ICache.(CoreDatasourceCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CoreDatasourceCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_coreDatasourceCache_Get(t *testing.T) {
	c := newCoreDatasourceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreDatasource)
	err := c.ICache.(CoreDatasourceCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreDatasourceCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CoreDatasourceCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_coreDatasourceCache_MultiGet(t *testing.T) {
	c := newCoreDatasourceCache()
	defer c.Close()

	var testData []*model.CoreDatasource
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreDatasource))
	}

	err := c.ICache.(CoreDatasourceCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreDatasourceCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CoreDatasource))
	}
}

func Test_coreDatasourceCache_MultiSet(t *testing.T) {
	c := newCoreDatasourceCache()
	defer c.Close()

	var testData []*model.CoreDatasource
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreDatasource))
	}

	err := c.ICache.(CoreDatasourceCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreDatasourceCache_Del(t *testing.T) {
	c := newCoreDatasourceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreDatasource)
	err := c.ICache.(CoreDatasourceCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreDatasourceCache_SetCacheWithNotFound(t *testing.T) {
	c := newCoreDatasourceCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreDatasource)
	err := c.ICache.(CoreDatasourceCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CoreDatasourceCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCoreDatasourceCache(t *testing.T) {
	c := NewCoreDatasourceCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCoreDatasourceCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCoreDatasourceCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
