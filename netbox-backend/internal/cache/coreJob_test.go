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

func newCoreJobCache() *gotest.Cache {
	record1 := &model.CoreJob{}
	record1.ID = 1
	record2 := &model.CoreJob{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCoreJobCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_coreJobCache_Set(t *testing.T) {
	c := newCoreJobCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreJob)
	err := c.ICache.(CoreJobCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CoreJobCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_coreJobCache_Get(t *testing.T) {
	c := newCoreJobCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreJob)
	err := c.ICache.(CoreJobCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreJobCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CoreJobCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_coreJobCache_MultiGet(t *testing.T) {
	c := newCoreJobCache()
	defer c.Close()

	var testData []*model.CoreJob
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreJob))
	}

	err := c.ICache.(CoreJobCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreJobCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CoreJob))
	}
}

func Test_coreJobCache_MultiSet(t *testing.T) {
	c := newCoreJobCache()
	defer c.Close()

	var testData []*model.CoreJob
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreJob))
	}

	err := c.ICache.(CoreJobCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreJobCache_Del(t *testing.T) {
	c := newCoreJobCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreJob)
	err := c.ICache.(CoreJobCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreJobCache_SetCacheWithNotFound(t *testing.T) {
	c := newCoreJobCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreJob)
	err := c.ICache.(CoreJobCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CoreJobCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCoreJobCache(t *testing.T) {
	c := NewCoreJobCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCoreJobCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCoreJobCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
