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

func newDcimModuleCache() *gotest.Cache {
	record1 := &model.DcimModule{}
	record1.ID = 1
	record2 := &model.DcimModule{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimModuleCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimModuleCache_Set(t *testing.T) {
	c := newDcimModuleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModule)
	err := c.ICache.(DcimModuleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimModuleCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimModuleCache_Get(t *testing.T) {
	c := newDcimModuleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModule)
	err := c.ICache.(DcimModuleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModuleCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimModuleCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimModuleCache_MultiGet(t *testing.T) {
	c := newDcimModuleCache()
	defer c.Close()

	var testData []*model.DcimModule
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModule))
	}

	err := c.ICache.(DcimModuleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModuleCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimModule))
	}
}

func Test_dcimModuleCache_MultiSet(t *testing.T) {
	c := newDcimModuleCache()
	defer c.Close()

	var testData []*model.DcimModule
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModule))
	}

	err := c.ICache.(DcimModuleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModuleCache_Del(t *testing.T) {
	c := newDcimModuleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModule)
	err := c.ICache.(DcimModuleCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModuleCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimModuleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModule)
	err := c.ICache.(DcimModuleCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimModuleCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimModuleCache(t *testing.T) {
	c := NewDcimModuleCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimModuleCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimModuleCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
