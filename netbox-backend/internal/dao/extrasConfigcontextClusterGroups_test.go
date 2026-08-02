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

func newExtrasConfigcontextClusterGroupsDao() *gotest.Dao {
	testData := &model.ExtrasConfigcontextClusterGroups{}
	testData.ID = 1
	testData.ConfigcontextID = 1
	// you can set the other fields of testData here, such as:
	//testData.CreatedAt = time.Now()
	//testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(testData.ID): testData})
	c.ICache = cache.NewExtrasConfigcontextClusterGroupsCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewExtrasConfigcontextClusterGroupsDao(d.DB, c.ICache.(cache.ExtrasConfigcontextClusterGroupsCache))

	return d
}

func Test_extrasConfigcontextClusterGroupsDao_Create(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterGroupsDao_DeleteByID(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)
	expectedSQLForDeletion := "DELETE .*"

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterGroupsDao_UpdateByID(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).UpdateByID(d.Ctx, &model.ExtrasConfigcontextClusterGroups{})
	assert.Error(t, err)

}

func Test_extrasConfigcontextClusterGroupsDao_GetByID(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByID(d.Ctx, testData.ID)
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
	_, err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterGroupsDao_GetByColumns(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByColumns(d.Ctx, &query.Params{
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
	_, _, err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByColumns(d.Ctx, &query.Params{
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
	dao := &extrasConfigcontextClusterGroupsDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_extrasConfigcontextClusterGroupsDao_DeleteByIDs(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("DELETE .*").
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).DeleteByIDs(d.Ctx, []uint64{0})
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterGroupsDao_GetByCondition(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByCondition(d.Ctx, &query.Conditions{
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
	_, err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByCondition(d.Ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "id",
				Value: 2,
			},
		},
	})
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterGroupsDao_GetByIDs(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID).
		WillReturnRows(rows)

	_, err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByIDs(d.Ctx, []uint64{testData.ID})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByIDs(d.Ctx, []uint64{111})
	assert.Error(t, err)

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterGroupsDao_GetByLastID(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByLastID(d.Ctx, 0, 10, "")
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, err = d.IDao.(ExtrasConfigcontextClusterGroupsDao).GetByLastID(d.Ctx, 0, 10, "unknown-column")
	assert.Error(t, err)
}

func Test_extrasConfigcontextClusterGroupsDao_CreateByTx(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterGroupsDao_DeleteByTx(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)
	expectedSQLForDeletion := "DELETE .*"

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).DeleteByTx(d.Ctx, d.DB, testData.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_extrasConfigcontextClusterGroupsDao_UpdateByTx(t *testing.T) {
	d := newExtrasConfigcontextClusterGroupsDao()
	defer d.Close()
	testData := d.TestData.(*model.ExtrasConfigcontextClusterGroups)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(ExtrasConfigcontextClusterGroupsDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
