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

func newCoreManagedfileCache() *gotest.Cache {
	record1 := &model.CoreManagedfile{}
	record1.ID = 1
	record2 := &model.CoreManagedfile{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCoreManagedfileCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_coreManagedfileCache_Set(t *testing.T) {
	c := newCoreManagedfileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreManagedfile)
	err := c.ICache.(CoreManagedfileCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CoreManagedfileCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_coreManagedfileCache_Get(t *testing.T) {
	c := newCoreManagedfileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreManagedfile)
	err := c.ICache.(CoreManagedfileCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreManagedfileCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CoreManagedfileCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_coreManagedfileCache_MultiGet(t *testing.T) {
	c := newCoreManagedfileCache()
	defer c.Close()

	var testData []*model.CoreManagedfile
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreManagedfile))
	}

	err := c.ICache.(CoreManagedfileCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreManagedfileCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CoreManagedfile))
	}
}

func Test_coreManagedfileCache_MultiSet(t *testing.T) {
	c := newCoreManagedfileCache()
	defer c.Close()

	var testData []*model.CoreManagedfile
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreManagedfile))
	}

	err := c.ICache.(CoreManagedfileCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreManagedfileCache_Del(t *testing.T) {
	c := newCoreManagedfileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreManagedfile)
	err := c.ICache.(CoreManagedfileCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreManagedfileCache_SetCacheWithNotFound(t *testing.T) {
	c := newCoreManagedfileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreManagedfile)
	err := c.ICache.(CoreManagedfileCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CoreManagedfileCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCoreManagedfileCache(t *testing.T) {
	c := NewCoreManagedfileCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCoreManagedfileCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCoreManagedfileCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
