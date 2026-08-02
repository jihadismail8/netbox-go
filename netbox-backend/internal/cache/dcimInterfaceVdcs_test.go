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

func newDcimInterfaceVdcsCache() *gotest.Cache {
	record1 := &model.DcimInterfaceVdcs{}
	record1.ID = 1
	record2 := &model.DcimInterfaceVdcs{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimInterfaceVdcsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimInterfaceVdcsCache_Set(t *testing.T) {
	c := newDcimInterfaceVdcsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceVdcs)
	err := c.ICache.(DcimInterfaceVdcsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimInterfaceVdcsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimInterfaceVdcsCache_Get(t *testing.T) {
	c := newDcimInterfaceVdcsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceVdcs)
	err := c.ICache.(DcimInterfaceVdcsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInterfaceVdcsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimInterfaceVdcsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimInterfaceVdcsCache_MultiGet(t *testing.T) {
	c := newDcimInterfaceVdcsCache()
	defer c.Close()

	var testData []*model.DcimInterfaceVdcs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInterfaceVdcs))
	}

	err := c.ICache.(DcimInterfaceVdcsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInterfaceVdcsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimInterfaceVdcs))
	}
}

func Test_dcimInterfaceVdcsCache_MultiSet(t *testing.T) {
	c := newDcimInterfaceVdcsCache()
	defer c.Close()

	var testData []*model.DcimInterfaceVdcs
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInterfaceVdcs))
	}

	err := c.ICache.(DcimInterfaceVdcsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInterfaceVdcsCache_Del(t *testing.T) {
	c := newDcimInterfaceVdcsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceVdcs)
	err := c.ICache.(DcimInterfaceVdcsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInterfaceVdcsCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimInterfaceVdcsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceVdcs)
	err := c.ICache.(DcimInterfaceVdcsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimInterfaceVdcsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimInterfaceVdcsCache(t *testing.T) {
	c := NewDcimInterfaceVdcsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimInterfaceVdcsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimInterfaceVdcsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
