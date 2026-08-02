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

func newDcimPowerporttemplateCache() *gotest.Cache {
	record1 := &model.DcimPowerporttemplate{}
	record1.ID = 1
	record2 := &model.DcimPowerporttemplate{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimPowerporttemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimPowerporttemplateCache_Set(t *testing.T) {
	c := newDcimPowerporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerporttemplate)
	err := c.ICache.(DcimPowerporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimPowerporttemplateCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimPowerporttemplateCache_Get(t *testing.T) {
	c := newDcimPowerporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerporttemplate)
	err := c.ICache.(DcimPowerporttemplateCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPowerporttemplateCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimPowerporttemplateCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimPowerporttemplateCache_MultiGet(t *testing.T) {
	c := newDcimPowerporttemplateCache()
	defer c.Close()

	var testData []*model.DcimPowerporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPowerporttemplate))
	}

	err := c.ICache.(DcimPowerporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPowerporttemplateCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimPowerporttemplate))
	}
}

func Test_dcimPowerporttemplateCache_MultiSet(t *testing.T) {
	c := newDcimPowerporttemplateCache()
	defer c.Close()

	var testData []*model.DcimPowerporttemplate
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPowerporttemplate))
	}

	err := c.ICache.(DcimPowerporttemplateCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPowerporttemplateCache_Del(t *testing.T) {
	c := newDcimPowerporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerporttemplate)
	err := c.ICache.(DcimPowerporttemplateCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPowerporttemplateCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimPowerporttemplateCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPowerporttemplate)
	err := c.ICache.(DcimPowerporttemplateCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimPowerporttemplateCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimPowerporttemplateCache(t *testing.T) {
	c := NewDcimPowerporttemplateCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimPowerporttemplateCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimPowerporttemplateCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
