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

func newDcimModuletypeprofileCache() *gotest.Cache {
	record1 := &model.DcimModuletypeprofile{}
	record1.ID = 1
	record2 := &model.DcimModuletypeprofile{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewDcimModuletypeprofileCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_dcimModuletypeprofileCache_Set(t *testing.T) {
	c := newDcimModuletypeprofileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletypeprofile)
	err := c.ICache.(DcimModuletypeprofileCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(DcimModuletypeprofileCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_dcimModuletypeprofileCache_Get(t *testing.T) {
	c := newDcimModuletypeprofileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletypeprofile)
	err := c.ICache.(DcimModuletypeprofileCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModuletypeprofileCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(DcimModuletypeprofileCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimModuletypeprofileCache_MultiGet(t *testing.T) {
	c := newDcimModuletypeprofileCache()
	defer c.Close()

	var testData []*model.DcimModuletypeprofile
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModuletypeprofile))
	}

	err := c.ICache.(DcimModuletypeprofileCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(DcimModuletypeprofileCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.DcimModuletypeprofile))
	}
}

func Test_dcimModuletypeprofileCache_MultiSet(t *testing.T) {
	c := newDcimModuletypeprofileCache()
	defer c.Close()

	var testData []*model.DcimModuletypeprofile
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.DcimModuletypeprofile))
	}

	err := c.ICache.(DcimModuletypeprofileCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModuletypeprofileCache_Del(t *testing.T) {
	c := newDcimModuletypeprofileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletypeprofile)
	err := c.ICache.(DcimModuletypeprofileCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModuletypeprofileCache_SetCacheWithNotFound(t *testing.T) {
	c := newDcimModuletypeprofileCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.DcimModuletypeprofile)
	err := c.ICache.(DcimModuletypeprofileCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(DcimModuletypeprofileCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewDcimModuletypeprofileCache(t *testing.T) {
	c := NewDcimModuletypeprofileCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewDcimModuletypeprofileCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewDcimModuletypeprofileCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
