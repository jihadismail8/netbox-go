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

func newVirtualizationClustertypeCache() *gotest.Cache {
	record1 := &model.VirtualizationClustertype{}
	record1.ID = 1
	record2 := &model.VirtualizationClustertype{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVirtualizationClustertypeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_virtualizationClustertypeCache_Set(t *testing.T) {
	c := newVirtualizationClustertypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustertype)
	err := c.ICache.(VirtualizationClustertypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VirtualizationClustertypeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_virtualizationClustertypeCache_Get(t *testing.T) {
	c := newVirtualizationClustertypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustertype)
	err := c.ICache.(VirtualizationClustertypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationClustertypeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VirtualizationClustertypeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_virtualizationClustertypeCache_MultiGet(t *testing.T) {
	c := newVirtualizationClustertypeCache()
	defer c.Close()

	var testData []*model.VirtualizationClustertype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationClustertype))
	}

	err := c.ICache.(VirtualizationClustertypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VirtualizationClustertypeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VirtualizationClustertype))
	}
}

func Test_virtualizationClustertypeCache_MultiSet(t *testing.T) {
	c := newVirtualizationClustertypeCache()
	defer c.Close()

	var testData []*model.VirtualizationClustertype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VirtualizationClustertype))
	}

	err := c.ICache.(VirtualizationClustertypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationClustertypeCache_Del(t *testing.T) {
	c := newVirtualizationClustertypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustertype)
	err := c.ICache.(VirtualizationClustertypeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_virtualizationClustertypeCache_SetCacheWithNotFound(t *testing.T) {
	c := newVirtualizationClustertypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VirtualizationClustertype)
	err := c.ICache.(VirtualizationClustertypeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VirtualizationClustertypeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVirtualizationClustertypeCache(t *testing.T) {
	c := NewVirtualizationClustertypeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVirtualizationClustertypeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVirtualizationClustertypeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
