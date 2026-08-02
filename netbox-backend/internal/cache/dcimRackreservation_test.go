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

func newDcimRackreservationCache() *gotest.Cache {
	record1 := &model.DcimRackreservation{}
	record1.ID = 1
	record2 := &model.DcimRackreservation{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimRackreservationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimRackreservationCache_Set(t *testing.T) {
	c := newDcimRackreservationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRackreservation)
	err := c.ICache.(DcimRackreservationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimRackreservationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimRackreservationCache_Get(t *testing.T) {
	c := newDcimRackreservationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRackreservation)
	err := c.ICache.(DcimRackreservationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimRackreservationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimRackreservationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimRackreservationCache_MultiGet(t *testing.T) {
	c := newDcimRackreservationCache()
	defer c.Close()

	var testData []*model.DcimRackreservation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimRackreservation))
	}

	err := c.ICache.(DcimRackreservationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimRackreservationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimRackreservation))
	}
}

func Test_dcimRackreservationCache_MultiSet(t *testing.T) {
	c := newDcimRackreservationCache()
	defer c.Close()

	var testData []*model.DcimRackreservation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimRackreservation))
	}

	err := c.ICache.(DcimRackreservationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimRackreservationCache_Del(t *testing.T) {
	c := newDcimRackreservationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRackreservation)
	err := c.ICache.(DcimRackreservationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimRackreservationCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimRackreservationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimRackreservation)
	err := c.ICache.(DcimRackreservationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimRackreservationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimRackreservationCache(t *testing.T) {
	c := NewDcimRackreservationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimRackreservationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimRackreservationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
