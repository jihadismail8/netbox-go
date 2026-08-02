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

func newDcimInventoryitemroleCache() *gotest.Cache {
	record1 := &model.DcimInventoryitemrole{}
	record1.ID = 1
	record2 := &model.DcimInventoryitemrole{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimInventoryitemroleCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimInventoryitemroleCache_Set(t *testing.T) {
	c := newDcimInventoryitemroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemrole)
	err := c.ICache.(DcimInventoryitemroleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimInventoryitemroleCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimInventoryitemroleCache_Get(t *testing.T) {
	c := newDcimInventoryitemroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemrole)
	err := c.ICache.(DcimInventoryitemroleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInventoryitemroleCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimInventoryitemroleCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimInventoryitemroleCache_MultiGet(t *testing.T) {
	c := newDcimInventoryitemroleCache()
	defer c.Close()

	var testData []*model.DcimInventoryitemrole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInventoryitemrole))
	}

	err := c.ICache.(DcimInventoryitemroleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInventoryitemroleCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimInventoryitemrole))
	}
}

func Test_dcimInventoryitemroleCache_MultiSet(t *testing.T) {
	c := newDcimInventoryitemroleCache()
	defer c.Close()

	var testData []*model.DcimInventoryitemrole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInventoryitemrole))
	}

	err := c.ICache.(DcimInventoryitemroleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInventoryitemroleCache_Del(t *testing.T) {
	c := newDcimInventoryitemroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemrole)
	err := c.ICache.(DcimInventoryitemroleCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInventoryitemroleCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimInventoryitemroleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInventoryitemrole)
	err := c.ICache.(DcimInventoryitemroleCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimInventoryitemroleCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimInventoryitemroleCache(t *testing.T) {
	c := NewDcimInventoryitemroleCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimInventoryitemroleCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimInventoryitemroleCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
