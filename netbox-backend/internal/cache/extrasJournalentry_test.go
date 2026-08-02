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

func newExtrasJournalentryCache() *gotest.Cache {
	record1 := &model.ExtrasJournalentry{}
	record1.ID = 1
	record2 := &model.ExtrasJournalentry{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewExtrasJournalentryCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_extrasJournalentryCache_Set(t *testing.T) {
	c := newExtrasJournalentryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasJournalentry)
	err := c.ICache.(ExtrasJournalentryCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(ExtrasJournalentryCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_extrasJournalentryCache_Get(t *testing.T) {
	c := newExtrasJournalentryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasJournalentry)
	err := c.ICache.(ExtrasJournalentryCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasJournalentryCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(ExtrasJournalentryCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasJournalentryCache_MultiGet(t *testing.T) {
	c := newExtrasJournalentryCache()
	defer c.Close()

	var testData []*model.ExtrasJournalentry
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasJournalentry))
	}

	err := c.ICache.(ExtrasJournalentryCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(ExtrasJournalentryCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.ExtrasJournalentry))
	}
}

func Test_extrasJournalentryCache_MultiSet(t *testing.T) {
	c := newExtrasJournalentryCache()
	defer c.Close()

	var testData []*model.ExtrasJournalentry
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.ExtrasJournalentry))
	}

	err := c.ICache.(ExtrasJournalentryCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasJournalentryCache_Del(t *testing.T) {
	c := newExtrasJournalentryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasJournalentry)
	err := c.ICache.(ExtrasJournalentryCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasJournalentryCache_SetCacheWithNotFound(t *testing.T) {
	c := newExtrasJournalentryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.ExtrasJournalentry)
	err := c.ICache.(ExtrasJournalentryCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(ExtrasJournalentryCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewExtrasJournalentryCache(t *testing.T) {
	c := NewExtrasJournalentryCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewExtrasJournalentryCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewExtrasJournalentryCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
