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

func newCoreAutosyncrecordCache() *gotest.Cache {
	record1 := &model.CoreAutosyncrecord{}
	record1.ID = 1
	record2 := &model.CoreAutosyncrecord{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewCoreAutosyncrecordCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_coreAutosyncrecordCache_Set(t *testing.T) {
	c := newCoreAutosyncrecordCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreAutosyncrecord)
	err := c.ICache.(CoreAutosyncrecordCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(CoreAutosyncrecordCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_coreAutosyncrecordCache_Get(t *testing.T) {
	c := newCoreAutosyncrecordCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreAutosyncrecord)
	err := c.ICache.(CoreAutosyncrecordCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreAutosyncrecordCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(CoreAutosyncrecordCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_coreAutosyncrecordCache_MultiGet(t *testing.T) {
	c := newCoreAutosyncrecordCache()
	defer c.Close()

	var testData []*model.CoreAutosyncrecord
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreAutosyncrecord))
	}

	err := c.ICache.(CoreAutosyncrecordCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(CoreAutosyncrecordCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.CoreAutosyncrecord))
	}
}

func Test_coreAutosyncrecordCache_MultiSet(t *testing.T) {
	c := newCoreAutosyncrecordCache()
	defer c.Close()

	var testData []*model.CoreAutosyncrecord
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.CoreAutosyncrecord))
	}

	err := c.ICache.(CoreAutosyncrecordCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreAutosyncrecordCache_Del(t *testing.T) {
	c := newCoreAutosyncrecordCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreAutosyncrecord)
	err := c.ICache.(CoreAutosyncrecordCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_coreAutosyncrecordCache_SetCacheWithNotFound(t *testing.T) {
	c := newCoreAutosyncrecordCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.CoreAutosyncrecord)
	err := c.ICache.(CoreAutosyncrecordCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(CoreAutosyncrecordCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewCoreAutosyncrecordCache(t *testing.T) {
	c := NewCoreAutosyncrecordCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewCoreAutosyncrecordCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewCoreAutosyncrecordCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
