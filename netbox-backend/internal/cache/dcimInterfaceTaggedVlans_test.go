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

func newDcimInterfaceTaggedVlansCache() *gotest.Cache {
	record1 := &model.DcimInterfaceTaggedVlans{}
	record1.ID = 1
	record2 := &model.DcimInterfaceTaggedVlans{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimInterfaceTaggedVlansCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimInterfaceTaggedVlansCache_Set(t *testing.T) {
	c := newDcimInterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceTaggedVlans)
	err := c.ICache.(DcimInterfaceTaggedVlansCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimInterfaceTaggedVlansCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimInterfaceTaggedVlansCache_Get(t *testing.T) {
	c := newDcimInterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceTaggedVlans)
	err := c.ICache.(DcimInterfaceTaggedVlansCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInterfaceTaggedVlansCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimInterfaceTaggedVlansCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimInterfaceTaggedVlansCache_MultiGet(t *testing.T) {
	c := newDcimInterfaceTaggedVlansCache()
	defer c.Close()

	var testData []*model.DcimInterfaceTaggedVlans
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInterfaceTaggedVlans))
	}

	err := c.ICache.(DcimInterfaceTaggedVlansCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimInterfaceTaggedVlansCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimInterfaceTaggedVlans))
	}
}

func Test_dcimInterfaceTaggedVlansCache_MultiSet(t *testing.T) {
	c := newDcimInterfaceTaggedVlansCache()
	defer c.Close()

	var testData []*model.DcimInterfaceTaggedVlans
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimInterfaceTaggedVlans))
	}

	err := c.ICache.(DcimInterfaceTaggedVlansCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInterfaceTaggedVlansCache_Del(t *testing.T) {
	c := newDcimInterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceTaggedVlans)
	err := c.ICache.(DcimInterfaceTaggedVlansCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimInterfaceTaggedVlansCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimInterfaceTaggedVlansCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimInterfaceTaggedVlans)
	err := c.ICache.(DcimInterfaceTaggedVlansCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimInterfaceTaggedVlansCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimInterfaceTaggedVlansCache(t *testing.T) {
	c := NewDcimInterfaceTaggedVlansCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimInterfaceTaggedVlansCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimInterfaceTaggedVlansCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
