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

func newDcimPowerpanelCache() *gotest.Cache {
	record1 := &model.DcimPowerpanel{}
	record1.ID = 1
	record2 := &model.DcimPowerpanel{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimPowerpanelCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimPowerpanelCache_Set(t *testing.T) {
	c := newDcimPowerpanelCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerpanel)
	err := c.ICache.(DcimPowerpanelCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimPowerpanelCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimPowerpanelCache_Get(t *testing.T) {
	c := newDcimPowerpanelCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerpanel)
	err := c.ICache.(DcimPowerpanelCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPowerpanelCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimPowerpanelCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimPowerpanelCache_MultiGet(t *testing.T) {
	c := newDcimPowerpanelCache()
	defer c.Close()

	var testData []*model.DcimPowerpanel
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPowerpanel))
	}

	err := c.ICache.(DcimPowerpanelCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPowerpanelCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimPowerpanel))
	}
}

func Test_dcimPowerpanelCache_MultiSet(t *testing.T) {
	c := newDcimPowerpanelCache()
	defer c.Close()

	var testData []*model.DcimPowerpanel
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPowerpanel))
	}

	err := c.ICache.(DcimPowerpanelCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPowerpanelCache_Del(t *testing.T) {
	c := newDcimPowerpanelCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerpanel)
	err := c.ICache.(DcimPowerpanelCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPowerpanelCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimPowerpanelCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerpanel)
	err := c.ICache.(DcimPowerpanelCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimPowerpanelCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimPowerpanelCache(t *testing.T) {
	c := NewDcimPowerpanelCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimPowerpanelCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimPowerpanelCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
