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

func newExtrasConfigcontextClustersCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextClusters{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextClusters{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextClustersCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextClustersCache_Set(t *testing.T) {
	c := newExtrasConfigcontextClustersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusters)
	err := c.ICache.(ExtrasConfigcontextClustersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextClustersCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextClustersCache_Get(t *testing.T) {
	c := newExtrasConfigcontextClustersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusters)
	err := c.ICache.(ExtrasConfigcontextClustersCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextClustersCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextClustersCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextClustersCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextClustersCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextClusters
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextClusters))
	}

	err := c.ICache.(ExtrasConfigcontextClustersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextClustersCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextClusters))
	}
}

func Test_extrasConfigcontextClustersCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextClustersCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextClusters
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextClusters))
	}

	err := c.ICache.(ExtrasConfigcontextClustersCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClustersCache_Del(t *testing.T) {
	c := newExtrasConfigcontextClustersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusters)
	err := c.ICache.(ExtrasConfigcontextClustersCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClustersCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextClustersCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextClusters)
	err := c.ICache.(ExtrasConfigcontextClustersCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextClustersCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextClustersCache(t *testing.T) {
	c := NewExtrasConfigcontextClustersCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextClustersCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextClustersCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
