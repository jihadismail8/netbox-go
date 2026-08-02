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

func newExtrasDashboardCache() *gotest.Cache {
	record1 := &model.ExtrasDashboard{}
	record1.ID = 1
	record2 := &model.ExtrasDashboard{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasDashboardCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasDashboardCache_Set(t *testing.T) {
	c := newExtrasDashboardCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasDashboard)
	err := c.ICache.(ExtrasDashboardCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasDashboardCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasDashboardCache_Get(t *testing.T) {
	c := newExtrasDashboardCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasDashboard)
	err := c.ICache.(ExtrasDashboardCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasDashboardCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasDashboardCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasDashboardCache_MultiGet(t *testing.T) {
	c := newExtrasDashboardCache()
	defer c.Close()

	var testData []*model.ExtrasDashboard
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasDashboard))
	}

	err := c.ICache.(ExtrasDashboardCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasDashboardCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasDashboard))
	}
}

func Test_extrasDashboardCache_MultiSet(t *testing.T) {
	c := newExtrasDashboardCache()
	defer c.Close()

	var testData []*model.ExtrasDashboard
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasDashboard))
	}

	err := c.ICache.(ExtrasDashboardCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasDashboardCache_Del(t *testing.T) {
	c := newExtrasDashboardCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasDashboard)
	err := c.ICache.(ExtrasDashboardCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasDashboardCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasDashboardCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasDashboard)
	err := c.ICache.(ExtrasDashboardCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasDashboardCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasDashboardCache(t *testing.T) {
	c := NewExtrasDashboardCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasDashboardCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasDashboardCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
