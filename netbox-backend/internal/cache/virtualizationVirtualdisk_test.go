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

func newVirtualizationVirtualdiskCache() *gotest.Cache {
	record1 := &model.VirtualizationVirtualdisk{}
	record1.ID = 1
	record2 := &model.VirtualizationVirtualdisk{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVirtualizationVirtualdiskCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_virtualizationVirtualdiskCache_Set(t *testing.T) {
	c := newVirtualizationVirtualdiskCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualdisk)
	err := c.ICache.(VirtualizationVirtualdiskCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VirtualizationVirtualdiskCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_virtualizationVirtualdiskCache_Get(t *testing.T) {
	c := newVirtualizationVirtualdiskCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualdisk)
	err := c.ICache.(VirtualizationVirtualdiskCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVirtualdiskCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VirtualizationVirtualdiskCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_virtualizationVirtualdiskCache_MultiGet(t *testing.T) {
	c := newVirtualizationVirtualdiskCache()
	defer c.Close()

	var testData []*model.VirtualizationVirtualdisk
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVirtualdisk))
	}

	err := c.ICache.(VirtualizationVirtualdiskCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationVirtualdiskCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VirtualizationVirtualdisk))
	}
}

func Test_virtualizationVirtualdiskCache_MultiSet(t *testing.T) {
	c := newVirtualizationVirtualdiskCache()
	defer c.Close()

	var testData []*model.VirtualizationVirtualdisk
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationVirtualdisk))
	}

	err := c.ICache.(VirtualizationVirtualdiskCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVirtualdiskCache_Del(t *testing.T) {
	c := newVirtualizationVirtualdiskCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualdisk)
	err := c.ICache.(VirtualizationVirtualdiskCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationVirtualdiskCache_SetCacheWithNotFound(t *testing.T) {
	c := newVirtualizationVirtualdiskCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationVirtualdisk)
	err := c.ICache.(VirtualizationVirtualdiskCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VirtualizationVirtualdiskCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVirtualizationVirtualdiskCache(t *testing.T) {
	c := NewVirtualizationVirtualdiskCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVirtualizationVirtualdiskCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVirtualizationVirtualdiskCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
