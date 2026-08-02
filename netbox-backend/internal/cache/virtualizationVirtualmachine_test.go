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

func newVirtualizationVirtualmachineCache() *gotest.Cache {
	record1 := &model.VirtualizationVirtualmachine{}
	record1.ID = 1
	record2 := &model.VirtualizationVirtualmachine{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVirtualizationVirtualmachineCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_virtualizationVirtualmachineCache_Set(t *testing.T) {
	c := newVirtualizationVirtualmachineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualmachine)
	err := c.ICache.(VirtualizationVirtualmachineCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VirtualizationVirtualmachineCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_virtualizationVirtualmachineCache_Get(t *testing.T) {
	c := newVirtualizationVirtualmachineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualmachine)
	err := c.ICache.(VirtualizationVirtualmachineCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVirtualmachineCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VirtualizationVirtualmachineCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_virtualizationVirtualmachineCache_MultiGet(t *testing.T) {
	c := newVirtualizationVirtualmachineCache()
	defer c.Close()

	var testData []*model.VirtualizationVirtualmachine
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVirtualmachine))
	}

	err := c.ICache.(VirtualizationVirtualmachineCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVirtualmachineCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VirtualizationVirtualmachine))
	}
}

func Test_virtualizationVirtualmachineCache_MultiSet(t *testing.T) {
	c := newVirtualizationVirtualmachineCache()
	defer c.Close()

	var testData []*model.VirtualizationVirtualmachine
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVirtualmachine))
	}

	err := c.ICache.(VirtualizationVirtualmachineCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVirtualmachineCache_Del(t *testing.T) {
	c := newVirtualizationVirtualmachineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualmachine)
	err := c.ICache.(VirtualizationVirtualmachineCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVirtualmachineCache_SetCacheWithNotFound(t *testing.T) {
	c := newVirtualizationVirtualmachineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualmachine)
	err := c.ICache.(VirtualizationVirtualmachineCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VirtualizationVirtualmachineCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVirtualizationVirtualmachineCache(t *testing.T) {
	c := NewVirtualizationVirtualmachineCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVirtualizationVirtualmachineCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVirtualizationVirtualmachineCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
