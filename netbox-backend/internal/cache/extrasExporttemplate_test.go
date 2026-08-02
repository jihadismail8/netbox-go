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

func newExtrasExporttemplateCache() *gotest.Cache {
	record1 := &model.ExtrasExporttemplate{}
	record1.ID = 1
	record2 := &model.ExtrasExporttemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasExporttemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasExporttemplateCache_Set(t *testing.T) {
	c := newExtrasExporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplate)
	err := c.ICache.(ExtrasExporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasExporttemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasExporttemplateCache_Get(t *testing.T) {
	c := newExtrasExporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplate)
	err := c.ICache.(ExtrasExporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasExporttemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasExporttemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasExporttemplateCache_MultiGet(t *testing.T) {
	c := newExtrasExporttemplateCache()
	defer c.Close()

	var testData []*model.ExtrasExporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasExporttemplate))
	}

	err := c.ICache.(ExtrasExporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasExporttemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasExporttemplate))
	}
}

func Test_extrasExporttemplateCache_MultiSet(t *testing.T) {
	c := newExtrasExporttemplateCache()
	defer c.Close()

	var testData []*model.ExtrasExporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasExporttemplate))
	}

	err := c.ICache.(ExtrasExporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasExporttemplateCache_Del(t *testing.T) {
	c := newExtrasExporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplate)
	err := c.ICache.(ExtrasExporttemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasExporttemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasExporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasExporttemplate)
	err := c.ICache.(ExtrasExporttemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasExporttemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasExporttemplateCache(t *testing.T) {
	c := NewExtrasExporttemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasExporttemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasExporttemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
