package dao

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"netbox-go/internal/cache"
	"netbox-go/internal/database"
	"netbox-go/internal/model"
)

func newVpnIpsecpolicyProposalsDao() *gotest.Dao {
	testData := &model.VpnIpsecpolicyProposals{}
	testData.ID = 1
	testData.IpsecpolicyID = 1
	// you can set the other fields of testData here, such as:
	//testData.CreatedAt = time.Now()
	//testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(testData.ID): testData})
	c.ICache = cache.NewVpnIpsecpolicyProposalsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewVpnIpsecpolicyProposalsDao(d.DB, c.ICache.(cache.VpnIpsecpolicyProposalsCache))

	return d
}

func Test_vpnIpsecpolicyProposalsDao_Create(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(VpnIpsecpolicyProposalsDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyProposalsDao_DeleteByID(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)
	expectedSQLForDeletion := "DELETE .*"

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(VpnIpsecpolicyProposalsDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(VpnIpsecpolicyProposalsDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyProposalsDao_UpdateByID(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(VpnIpsecpolicyProposalsDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(VpnIpsecpolicyProposalsDao).UpdateByID(d.Ctx, &model.VpnIpsecpolicyProposals{})
	assert.Error(t, err)

}

func Test_vpnIpsecpolicyProposalsDao_GetByID(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(VpnIpsecpolicyProposalsDao).GetByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// notfound error
	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(2).
		WillReturnRows(rows)
	_, err = d.IDao.(VpnIpsecpolicyProposalsDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(VpnIpsecpolicyProposalsDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyProposalsDao_GetByColumns(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(VpnIpsecpolicyProposalsDao).GetByColumns(d.Ctx, &query.Params{
		Page:  0,
		Limit: 10,
		Sort:  "ignore count", // ignore test count(*)
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, _, err = d.IDao.(VpnIpsecpolicyProposalsDao).GetByColumns(d.Ctx, &query.Params{
		Page:  0,
		Limit: 10,
		Columns: []query.Column{
			{
				Name:  "id",
				Exp:   "<",
				Value: 0,
			},
		},
	})
	assert.Error(t, err)

	// error test
	dao := &vpnIpsecpolicyProposalsDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_vpnIpsecpolicyProposalsDao_DeleteByIDs(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("DELETE .*").
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(VpnIpsecpolicyProposalsDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(VpnIpsecpolicyProposalsDao).DeleteByIDs(d.Ctx, []uint64{0})
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyProposalsDao_GetByCondition(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(VpnIpsecpolicyProposalsDao).GetByCondition(d.Ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "id",
				Value: testData.ID,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// notfound error
	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(2, 1).
		WillReturnRows(rows)
	_, err = d.IDao.(VpnIpsecpolicyProposalsDao).GetByCondition(d.Ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "id",
				Value: 2,
			},
		},
	})
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyProposalsDao_GetByIDs(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID).
		WillReturnRows(rows)

	_, err := d.IDao.(VpnIpsecpolicyProposalsDao).GetByIDs(d.Ctx, []uint64{testData.ID})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.IDao.(VpnIpsecpolicyProposalsDao).GetByIDs(d.Ctx, []uint64{111})
	assert.Error(t, err)

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyProposalsDao_GetByLastID(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, err := d.IDao.(VpnIpsecpolicyProposalsDao).GetByLastID(d.Ctx, 0, 10, "")
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, err = d.IDao.(VpnIpsecpolicyProposalsDao).GetByLastID(d.Ctx, 0, 10, "unknown-column")
	assert.Error(t, err)
}

func Test_vpnIpsecpolicyProposalsDao_CreateByTx(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(VpnIpsecpolicyProposalsDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyProposalsDao_DeleteByTx(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)
	expectedSQLForDeletion := "DELETE .*"

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(VpnIpsecpolicyProposalsDao).DeleteByTx(d.Ctx, d.DB, testData.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_vpnIpsecpolicyProposalsDao_UpdateByTx(t *testing.T) {
	d := newVpnIpsecpolicyProposalsDao()
	defer d.Close()
	testData := d.TestData.(*model.VpnIpsecpolicyProposals)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(VpnIpsecpolicyProposalsDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
