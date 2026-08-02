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

func newDcimLocationCache() *gotest.Cache {
	record1 := &model.DcimLocation{}
	record1.ID = 1
	record2 := &model.DcimLocation{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimLocationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimLocationCache_Set(t *testing.T) {
	c := newDcimLocationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimLocation)
	err := c.ICache.(DcimLocationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimLocationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimLocationCache_Get(t *testing.T) {
	c := newDcimLocationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimLocation)
	err := c.ICache.(DcimLocationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimLocationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimLocationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimLocationCache_MultiGet(t *testing.T) {
	c := newDcimLocationCache()
	defer c.Close()

	var testData []*model.DcimLocation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimLocation))
	}

	err := c.ICache.(DcimLocationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimLocationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimLocation))
	}
}

func Test_dcimLocationCache_MultiSet(t *testing.T) {
	c := newDcimLocationCache()
	defer c.Close()

	var testData []*model.DcimLocation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimLocation))
	}

	err := c.ICache.(DcimLocationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimLocationCache_Del(t *testing.T) {
	c := newDcimLocationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimLocation)
	err := c.ICache.(DcimLocationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimLocationCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimLocationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimLocation)
	err := c.ICache.(DcimLocationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimLocationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimLocationCache(t *testing.T) {
	c := NewDcimLocationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimLocationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimLocationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
