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

func newDcimPoweroutletCache() *gotest.Cache {
	record1 := &model.DcimPoweroutlet{}
	record1.ID = 1
	record2 := &model.DcimPoweroutlet{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimPoweroutletCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimPoweroutletCache_Set(t *testing.T) {
	c := newDcimPoweroutletCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPoweroutlet)
	err := c.ICache.(DcimPoweroutletCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimPoweroutletCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimPoweroutletCache_Get(t *testing.T) {
	c := newDcimPoweroutletCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPoweroutlet)
	err := c.ICache.(DcimPoweroutletCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPoweroutletCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimPoweroutletCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimPoweroutletCache_MultiGet(t *testing.T) {
	c := newDcimPoweroutletCache()
	defer c.Close()

	var testData []*model.DcimPoweroutlet
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPoweroutlet))
	}

	err := c.ICache.(DcimPoweroutletCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimPoweroutletCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimPoweroutlet))
	}
}

func Test_dcimPoweroutletCache_MultiSet(t *testing.T) {
	c := newDcimPoweroutletCache()
	defer c.Close()

	var testData []*model.DcimPoweroutlet
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimPoweroutlet))
	}

	err := c.ICache.(DcimPoweroutletCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPoweroutletCache_Del(t *testing.T) {
	c := newDcimPoweroutletCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPoweroutlet)
	err := c.ICache.(DcimPoweroutletCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimPoweroutletCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimPoweroutletCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimPoweroutlet)
	err := c.ICache.(DcimPoweroutletCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimPoweroutletCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimPoweroutletCache(t *testing.T) {
	c := NewDcimPoweroutletCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimPoweroutletCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimPoweroutletCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
