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

func newVpnIkeproposalCache() *gotest.Cache {
	record1 := &model.VpnIkeproposal{}
	record1.ID = 1
	record2 := &model.VpnIkeproposal{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnIkeproposalCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnIkeproposalCache_Set(t *testing.T) {
	c := newVpnIkeproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkeproposal)
	err := c.ICache.(VpnIkeproposalCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnIkeproposalCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnIkeproposalCache_Get(t *testing.T) {
	c := newVpnIkeproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkeproposal)
	err := c.ICache.(VpnIkeproposalCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIkeproposalCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnIkeproposalCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnIkeproposalCache_MultiGet(t *testing.T) {
	c := newVpnIkeproposalCache()
	defer c.Close()

	var testData []*model.VpnIkeproposal
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIkeproposal))
	}

	err := c.ICache.(VpnIkeproposalCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIkeproposalCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnIkeproposal))
	}
}

func Test_vpnIkeproposalCache_MultiSet(t *testing.T) {
	c := newVpnIkeproposalCache()
	defer c.Close()

	var testData []*model.VpnIkeproposal
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIkeproposal))
	}

	err := c.ICache.(VpnIkeproposalCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIkeproposalCache_Del(t *testing.T) {
	c := newVpnIkeproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkeproposal)
	err := c.ICache.(VpnIkeproposalCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIkeproposalCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnIkeproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIkeproposal)
	err := c.ICache.(VpnIkeproposalCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnIkeproposalCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnIkeproposalCache(t *testing.T) {
	c := NewVpnIkeproposalCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnIkeproposalCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnIkeproposalCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
