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

func newDcimRegionCache() *gotest.Cache {
	record1 := &model.DcimRegion{}
	record1.ID = 1
	record2 := &model.DcimRegion{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimRegionCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimRegionCache_Set(t *testing.T) {
	c := newDcimRegionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRegion)
	err := c.ICache.(DcimRegionCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimRegionCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimRegionCache_Get(t *testing.T) {
	c := newDcimRegionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRegion)
	err := c.ICache.(DcimRegionCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimRegionCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimRegionCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimRegionCache_MultiGet(t *testing.T) {
	c := newDcimRegionCache()
	defer c.Close()

	var testData []*model.DcimRegion
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimRegion))
	}

	err := c.ICache.(DcimRegionCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimRegionCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimRegion))
	}
}

func Test_dcimRegionCache_MultiSet(t *testing.T) {
	c := newDcimRegionCache()
	defer c.Close()

	var testData []*model.DcimRegion
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimRegion))
	}

	err := c.ICache.(DcimRegionCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimRegionCache_Del(t *testing.T) {
	c := newDcimRegionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRegion)
	err := c.ICache.(DcimRegionCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimRegionCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimRegionCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRegion)
	err := c.ICache.(DcimRegionCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimRegionCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimRegionCache(t *testing.T) {
	c := NewDcimRegionCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimRegionCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimRegionCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
