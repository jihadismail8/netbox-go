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

func newDcimConsoleporttemplateCache() *gotest.Cache {
	record1 := &model.DcimConsoleporttemplate{}
	record1.ID = 1
	record2 := &model.DcimConsoleporttemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimConsoleporttemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimConsoleporttemplateCache_Set(t *testing.T) {
	c := newDcimConsoleporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleporttemplate)
	err := c.ICache.(DcimConsoleporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimConsoleporttemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimConsoleporttemplateCache_Get(t *testing.T) {
	c := newDcimConsoleporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleporttemplate)
	err := c.ICache.(DcimConsoleporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleporttemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimConsoleporttemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimConsoleporttemplateCache_MultiGet(t *testing.T) {
	c := newDcimConsoleporttemplateCache()
	defer c.Close()

	var testData []*model.DcimConsoleporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleporttemplate))
	}

	err := c.ICache.(DcimConsoleporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimConsoleporttemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimConsoleporttemplate))
	}
}

func Test_dcimConsoleporttemplateCache_MultiSet(t *testing.T) {
	c := newDcimConsoleporttemplateCache()
	defer c.Close()

	var testData []*model.DcimConsoleporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimConsoleporttemplate))
	}

	err := c.ICache.(DcimConsoleporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleporttemplateCache_Del(t *testing.T) {
	c := newDcimConsoleporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleporttemplate)
	err := c.ICache.(DcimConsoleporttemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimConsoleporttemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimConsoleporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimConsoleporttemplate)
	err := c.ICache.(DcimConsoleporttemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimConsoleporttemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimConsoleporttemplateCache(t *testing.T) {
	c := NewDcimConsoleporttemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimConsoleporttemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimConsoleporttemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
