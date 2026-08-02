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

func newDcimModuletypeCache() *gotest.Cache {
	record1 := &model.DcimModuletype{}
	record1.ID = 1
	record2 := &model.DcimModuletype{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimModuletypeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimModuletypeCache_Set(t *testing.T) {
	c := newDcimModuletypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletype)
	err := c.ICache.(DcimModuletypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimModuletypeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimModuletypeCache_Get(t *testing.T) {
	c := newDcimModuletypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletype)
	err := c.ICache.(DcimModuletypeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModuletypeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimModuletypeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimModuletypeCache_MultiGet(t *testing.T) {
	c := newDcimModuletypeCache()
	defer c.Close()

	var testData []*model.DcimModuletype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModuletype))
	}

	err := c.ICache.(DcimModuletypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModuletypeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimModuletype))
	}
}

func Test_dcimModuletypeCache_MultiSet(t *testing.T) {
	c := newDcimModuletypeCache()
	defer c.Close()

	var testData []*model.DcimModuletype
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModuletype))
	}

	err := c.ICache.(DcimModuletypeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModuletypeCache_Del(t *testing.T) {
	c := newDcimModuletypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletype)
	err := c.ICache.(DcimModuletypeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModuletypeCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimModuletypeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletype)
	err := c.ICache.(DcimModuletypeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimModuletypeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimModuletypeCache(t *testing.T) {
	c := NewDcimModuletypeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimModuletypeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimModuletypeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
