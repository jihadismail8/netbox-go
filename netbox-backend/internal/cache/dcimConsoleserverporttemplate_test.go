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

func newDcimConsoleserverporttemplateCache() *gotest.Cache {
	record1 := &model.DcimConsoleserverporttemplate{}
	record1.ID = 1
	record2 := &model.DcimConsoleserverporttemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimConsoleserverporttemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimConsoleserverporttemplateCache_Set(t *testing.T) {
	c := newDcimConsoleserverporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverporttemplate)
	err := c.ICache.(DcimConsoleserverporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimConsoleserverporttemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimConsoleserverporttemplateCache_Get(t *testing.T) {
	c := newDcimConsoleserverporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverporttemplate)
	err := c.ICache.(DcimConsoleserverporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleserverporttemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimConsoleserverporttemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimConsoleserverporttemplateCache_MultiGet(t *testing.T) {
	c := newDcimConsoleserverporttemplateCache()
	defer c.Close()

	var testData []*model.DcimConsoleserverporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleserverporttemplate))
	}

	err := c.ICache.(DcimConsoleserverporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleserverporttemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimConsoleserverporttemplate))
	}
}

func Test_dcimConsoleserverporttemplateCache_MultiSet(t *testing.T) {
	c := newDcimConsoleserverporttemplateCache()
	defer c.Close()

	var testData []*model.DcimConsoleserverporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleserverporttemplate))
	}

	err := c.ICache.(DcimConsoleserverporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleserverporttemplateCache_Del(t *testing.T) {
	c := newDcimConsoleserverporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverporttemplate)
	err := c.ICache.(DcimConsoleserverporttemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleserverporttemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimConsoleserverporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleserverporttemplate)
	err := c.ICache.(DcimConsoleserverporttemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimConsoleserverporttemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimConsoleserverporttemplateCache(t *testing.T) {
	c := NewDcimConsoleserverporttemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimConsoleserverporttemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimConsoleserverporttemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
