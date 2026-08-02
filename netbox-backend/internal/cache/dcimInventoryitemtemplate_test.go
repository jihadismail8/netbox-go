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

func newDcimInventoryitemtemplateCache() *gotest.Cache {
	record1 := &model.DcimInventoryitemtemplate{}
	record1.ID = 1
	record2 := &model.DcimInventoryitemtemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimInventoryitemtemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimInventoryitemtemplateCache_Set(t *testing.T) {
	c := newDcimInventoryitemtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemtemplate)
	err := c.ICache.(DcimInventoryitemtemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimInventoryitemtemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimInventoryitemtemplateCache_Get(t *testing.T) {
	c := newDcimInventoryitemtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemtemplate)
	err := c.ICache.(DcimInventoryitemtemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInventoryitemtemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimInventoryitemtemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimInventoryitemtemplateCache_MultiGet(t *testing.T) {
	c := newDcimInventoryitemtemplateCache()
	defer c.Close()

	var testData []*model.DcimInventoryitemtemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInventoryitemtemplate))
	}

	err := c.ICache.(DcimInventoryitemtemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInventoryitemtemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimInventoryitemtemplate))
	}
}

func Test_dcimInventoryitemtemplateCache_MultiSet(t *testing.T) {
	c := newDcimInventoryitemtemplateCache()
	defer c.Close()

	var testData []*model.DcimInventoryitemtemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInventoryitemtemplate))
	}

	err := c.ICache.(DcimInventoryitemtemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInventoryitemtemplateCache_Del(t *testing.T) {
	c := newDcimInventoryitemtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemtemplate)
	err := c.ICache.(DcimInventoryitemtemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInventoryitemtemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimInventoryitemtemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemtemplate)
	err := c.ICache.(DcimInventoryitemtemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimInventoryitemtemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimInventoryitemtemplateCache(t *testing.T) {
	c := NewDcimInventoryitemtemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimInventoryitemtemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimInventoryitemtemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
