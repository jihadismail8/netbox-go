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

func newVpnIpsecpolicyProposalsCache() *gotest.Cache {
	record1 := &model.VpnIpsecpolicyProposals{}
	record1.ID = 1
	record2 := &model.VpnIpsecpolicyProposals{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnIpsecpolicyProposalsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnIpsecpolicyProposalsCache_Set(t *testing.T) {
	c := newVpnIpsecpolicyProposalsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicyProposals)
	err := c.ICache.(VpnIpsecpolicyProposalsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnIpsecpolicyProposalsCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnIpsecpolicyProposalsCache_Get(t *testing.T) {
	c := newVpnIpsecpolicyProposalsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicyProposals)
	err := c.ICache.(VpnIpsecpolicyProposalsCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIpsecpolicyProposalsCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnIpsecpolicyProposalsCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyProposalsCache_MultiGet(t *testing.T) {
	c := newVpnIpsecpolicyProposalsCache()
	defer c.Close()

	var testData []*model.VpnIpsecpolicyProposals
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIpsecpolicyProposals))
	}

	err := c.ICache.(VpnIpsecpolicyProposalsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIpsecpolicyProposalsCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnIpsecpolicyProposals))
	}
}

func Test_vpnIpsecpolicyProposalsCache_MultiSet(t *testing.T) {
	c := newVpnIpsecpolicyProposalsCache()
	defer c.Close()

	var testData []*model.VpnIpsecpolicyProposals
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIpsecpolicyProposals))
	}

	err := c.ICache.(VpnIpsecpolicyProposalsCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyProposalsCache_Del(t *testing.T) {
	c := newVpnIpsecpolicyProposalsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicyProposals)
	err := c.ICache.(VpnIpsecpolicyProposalsCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyProposalsCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnIpsecpolicyProposalsCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecpolicyProposals)
	err := c.ICache.(VpnIpsecpolicyProposalsCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnIpsecpolicyProposalsCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnIpsecpolicyProposalsCache(t *testing.T) {
	c := NewVpnIpsecpolicyProposalsCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnIpsecpolicyProposalsCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnIpsecpolicyProposalsCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
