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

func newDcimModulebayCache() *gotest.Cache {
	record1 := &model.DcimModulebay{}
	record1.ID = 1
	record2 := &model.DcimModulebay{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimModulebayCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimModulebayCache_Set(t *testing.T) {
	c := newDcimModulebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebay)
	err := c.ICache.(DcimModulebayCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimModulebayCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimModulebayCache_Get(t *testing.T) {
	c := newDcimModulebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebay)
	err := c.ICache.(DcimModulebayCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModulebayCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimModulebayCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimModulebayCache_MultiGet(t *testing.T) {
	c := newDcimModulebayCache()
	defer c.Close()

	var testData []*model.DcimModulebay
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModulebay))
	}

	err := c.ICache.(DcimModulebayCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModulebayCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimModulebay))
	}
}

func Test_dcimModulebayCache_MultiSet(t *testing.T) {
	c := newDcimModulebayCache()
	defer c.Close()

	var testData []*model.DcimModulebay
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModulebay))
	}

	err := c.ICache.(DcimModulebayCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebayCache_Del(t *testing.T) {
	c := newDcimModulebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebay)
	err := c.ICache.(DcimModulebayCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebayCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimModulebayCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModulebay)
	err := c.ICache.(DcimModulebayCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimModulebayCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimModulebayCache(t *testing.T) {
	c := NewDcimModulebayCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimModulebayCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimModulebayCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
