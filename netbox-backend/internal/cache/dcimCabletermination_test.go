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

func newDcimCableterminationCache() *gotest.Cache {
	record1 := &model.DcimCabletermination{}
	record1.ID = 1
	record2 := &model.DcimCabletermination{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimCableterminationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimCableterminationCache_Set(t *testing.T) {
	c := newDcimCableterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimCabletermination)
	err := c.ICache.(DcimCableterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimCableterminationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimCableterminationCache_Get(t *testing.T) {
	c := newDcimCableterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimCabletermination)
	err := c.ICache.(DcimCableterminationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimCableterminationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimCableterminationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimCableterminationCache_MultiGet(t *testing.T) {
	c := newDcimCableterminationCache()
	defer c.Close()

	var testData []*model.DcimCabletermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimCabletermination))
	}

	err := c.ICache.(DcimCableterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimCableterminationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimCabletermination))
	}
}

func Test_dcimCableterminationCache_MultiSet(t *testing.T) {
	c := newDcimCableterminationCache()
	defer c.Close()

	var testData []*model.DcimCabletermination
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimCabletermination))
	}

	err := c.ICache.(DcimCableterminationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimCableterminationCache_Del(t *testing.T) {
	c := newDcimCableterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimCabletermination)
	err := c.ICache.(DcimCableterminationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimCableterminationCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimCableterminationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimCabletermination)
	err := c.ICache.(DcimCableterminationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimCableterminationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimCableterminationCache(t *testing.T) {
	c := NewDcimCableterminationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimCableterminationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimCableterminationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
