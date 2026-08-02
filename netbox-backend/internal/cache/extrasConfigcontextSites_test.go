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

func newExtrasConfigcontextSitesCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextSites{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextSites{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextSitesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextSitesCache_Set(t *testing.T) {
	c := newExtrasConfigcontextSitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSites)
	err := c.ICache.(ExtrasConfigcontextSitesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextSitesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextSitesCache_Get(t *testing.T) {
	c := newExtrasConfigcontextSitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSites)
	err := c.ICache.(ExtrasConfigcontextSitesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextSitesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextSitesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextSitesCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextSitesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextSites
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextSites))
	}

	err := c.ICache.(ExtrasConfigcontextSitesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextSitesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextSites))
	}
}

func Test_extrasConfigcontextSitesCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextSitesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextSites
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextSites))
	}

	err := c.ICache.(ExtrasConfigcontextSitesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextSitesCache_Del(t *testing.T) {
	c := newExtrasConfigcontextSitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSites)
	err := c.ICache.(ExtrasConfigcontextSitesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextSitesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextSitesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextSites)
	err := c.ICache.(ExtrasConfigcontextSitesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextSitesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextSitesCache(t *testing.T) {
	c := NewExtrasConfigcontextSitesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextSitesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextSitesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
