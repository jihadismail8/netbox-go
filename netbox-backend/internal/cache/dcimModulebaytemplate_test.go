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

func newDcimModulebaytemplateCache() *gotest.Cache {
	record1 := &model.DcimModulebaytemplate{}
	record1.ID = 1
	record2 := &model.DcimModulebaytemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimModulebaytemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimModulebaytemplateCache_Set(t *testing.T) {
	c := newDcimModulebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebaytemplate)
	err := c.ICache.(DcimModulebaytemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimModulebaytemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimModulebaytemplateCache_Get(t *testing.T) {
	c := newDcimModulebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebaytemplate)
	err := c.ICache.(DcimModulebaytemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModulebaytemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimModulebaytemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimModulebaytemplateCache_MultiGet(t *testing.T) {
	c := newDcimModulebaytemplateCache()
	defer c.Close()

	var testData []*model.DcimModulebaytemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModulebaytemplate))
	}

	err := c.ICache.(DcimModulebaytemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModulebaytemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimModulebaytemplate))
	}
}

func Test_dcimModulebaytemplateCache_MultiSet(t *testing.T) {
	c := newDcimModulebaytemplateCache()
	defer c.Close()

	var testData []*model.DcimModulebaytemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModulebaytemplate))
	}

	err := c.ICache.(DcimModulebaytemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebaytemplateCache_Del(t *testing.T) {
	c := newDcimModulebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebaytemplate)
	err := c.ICache.(DcimModulebaytemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebaytemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimModulebaytemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebaytemplate)
	err := c.ICache.(DcimModulebaytemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimModulebaytemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimModulebaytemplateCache(t *testing.T) {
	c := NewDcimModulebaytemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimModulebaytemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimModulebaytemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
