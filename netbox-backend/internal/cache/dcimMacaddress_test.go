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

func newDcimMacaddressCache() *gotest.Cache {
	record1 := &model.DcimMacaddress{}
	record1.ID = 1
	record2 := &model.DcimMacaddress{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimMacaddressCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimMacaddressCache_Set(t *testing.T) {
	c := newDcimMacaddressCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimMacaddress)
	err := c.ICache.(DcimMacaddressCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimMacaddressCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimMacaddressCache_Get(t *testing.T) {
	c := newDcimMacaddressCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimMacaddress)
	err := c.ICache.(DcimMacaddressCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimMacaddressCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimMacaddressCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimMacaddressCache_MultiGet(t *testing.T) {
	c := newDcimMacaddressCache()
	defer c.Close()

	var testData []*model.DcimMacaddress
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimMacaddress))
	}

	err := c.ICache.(DcimMacaddressCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimMacaddressCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimMacaddress))
	}
}

func Test_dcimMacaddressCache_MultiSet(t *testing.T) {
	c := newDcimMacaddressCache()
	defer c.Close()

	var testData []*model.DcimMacaddress
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimMacaddress))
	}

	err := c.ICache.(DcimMacaddressCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimMacaddressCache_Del(t *testing.T) {
	c := newDcimMacaddressCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimMacaddress)
	err := c.ICache.(DcimMacaddressCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimMacaddressCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimMacaddressCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimMacaddress)
	err := c.ICache.(DcimMacaddressCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimMacaddressCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimMacaddressCache(t *testing.T) {
	c := NewDcimMacaddressCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimMacaddressCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimMacaddressCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
