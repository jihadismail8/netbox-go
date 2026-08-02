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

func newExtrasConfigcontextRolesCache() *gotest.Cache {
	record1 := &model.ExtrasConfigcontextRoles{}
	record1.ID = 1
	record2 := &model.ExtrasConfigcontextRoles{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasConfigcontextRolesCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasConfigcontextRolesCache_Set(t *testing.T) {
	c := newExtrasConfigcontextRolesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRoles)
	err := c.ICache.(ExtrasConfigcontextRolesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasConfigcontextRolesCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasConfigcontextRolesCache_Get(t *testing.T) {
	c := newExtrasConfigcontextRolesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRoles)
	err := c.ICache.(ExtrasConfigcontextRolesCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextRolesCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasConfigcontextRolesCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextRolesCache_MultiGet(t *testing.T) {
	c := newExtrasConfigcontextRolesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextRoles
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextRoles))
	}

	err := c.ICache.(ExtrasConfigcontextRolesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasConfigcontextRolesCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasConfigcontextRoles))
	}
}

func Test_extrasConfigcontextRolesCache_MultiSet(t *testing.T) {
	c := newExtrasConfigcontextRolesCache()
	defer c.Close()

	var testData []*model.ExtrasConfigcontextRoles
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasConfigcontextRoles))
	}

	err := c.ICache.(ExtrasConfigcontextRolesCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextRolesCache_Del(t *testing.T) {
	c := newExtrasConfigcontextRolesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRoles)
	err := c.ICache.(ExtrasConfigcontextRolesCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextRolesCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasConfigcontextRolesCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasConfigcontextRoles)
	err := c.ICache.(ExtrasConfigcontextRolesCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasConfigcontextRolesCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasConfigcontextRolesCache(t *testing.T) {
	c := NewExtrasConfigcontextRolesCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasConfigcontextRolesCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasConfigcontextRolesCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
