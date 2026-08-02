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

func newVpnIpsecproposalCache() *gotest.Cache {
	record1 := &model.VpnIpsecproposal{}
	record1.ID = 1
	record2 := &model.VpnIpsecproposal{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewVpnIpsecproposalCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_vpnIpsecproposalCache_Set(t *testing.T) {
	c := newVpnIpsecproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecproposal)
	err := c.ICache.(VpnIpsecproposalCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(VpnIpsecproposalCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_vpnIpsecproposalCache_Get(t *testing.T) {
	c := newVpnIpsecproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecproposal)
	err := c.ICache.(VpnIpsecproposalCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIpsecproposalCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(VpnIpsecproposalCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnIpsecproposalCache_MultiGet(t *testing.T) {
	c := newVpnIpsecproposalCache()
	defer c.Close()

	var testData []*model.VpnIpsecproposal
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIpsecproposal))
	}

	err := c.ICache.(VpnIpsecproposalCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(VpnIpsecproposalCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.VpnIpsecproposal))
	}
}

func Test_vpnIpsecproposalCache_MultiSet(t *testing.T) {
	c := newVpnIpsecproposalCache()
	defer c.Close()

	var testData []*model.VpnIpsecproposal
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.VpnIpsecproposal))
	}

	err := c.ICache.(VpnIpsecproposalCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecproposalCache_Del(t *testing.T) {
	c := newVpnIpsecproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecproposal)
	err := c.ICache.(VpnIpsecproposalCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecproposalCache_SetCacheWithNotFound(t *testing.T) {
	c := newVpnIpsecproposalCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.VpnIpsecproposal)
	err := c.ICache.(VpnIpsecproposalCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(VpnIpsecproposalCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewVpnIpsecproposalCache(t *testing.T) {
	c := NewVpnIpsecproposalCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewVpnIpsecproposalCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewVpnIpsecproposalCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
