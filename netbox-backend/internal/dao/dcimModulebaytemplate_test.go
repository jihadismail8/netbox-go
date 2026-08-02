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

func newDcimModulebaytemplateDao() *gotest.Dao {
	testData := &model.DcimModulebaytemplate{}
	testData.ID = 1
	testData.Name = "a"
	// you can set the other fields of testData here, such as:
	//testData.CreatedAt = time.Now()
	//testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(testData.ID): testData})
	c.ICache = cache.NewDcimModulebaytemplateCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewDcimModulebaytemplateDao(d.DB, c.ICache.(cache.DcimModulebaytemplateCache))

	return d
}

func Test_dcimModulebaytemplateDao_Create(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(DcimModulebaytemplateDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebaytemplateDao_DeleteByID(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)
	expectedSQLForDeletion := "DELETE .*"

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(DcimModulebaytemplateDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(DcimModulebaytemplateDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_dcimModulebaytemplateDao_UpdateByID(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(DcimModulebaytemplateDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(DcimModulebaytemplateDao).UpdateByID(d.Ctx, &model.DcimModulebaytemplate{})
	assert.Error(t, err)

}

func Test_dcimModulebaytemplateDao_GetByID(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(DcimModulebaytemplateDao).GetByID(d.Ctx, testData.ID)
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
	_, err = d.IDao.(DcimModulebaytemplateDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(DcimModulebaytemplateDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_dcimModulebaytemplateDao_GetByColumns(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(DcimModulebaytemplateDao).GetByColumns(d.Ctx, &query.Params{
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
	_, _, err = d.IDao.(DcimModulebaytemplateDao).GetByColumns(d.Ctx, &query.Params{
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
	dao := &dcimModulebaytemplateDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_dcimModulebaytemplateDao_DeleteByIDs(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("DELETE .*").
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(DcimModulebaytemplateDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(DcimModulebaytemplateDao).DeleteByIDs(d.Ctx, []uint64{0})
	assert.Error(t, err)
}

func Test_dcimModulebaytemplateDao_GetByCondition(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(DcimModulebaytemplateDao).GetByCondition(d.Ctx, &query.Conditions{
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
	_, err = d.IDao.(DcimModulebaytemplateDao).GetByCondition(d.Ctx, &query.Conditions{
		Columns: []query.Column{
			{
				Name:  "id",
				Value: 2,
			},
		},
	})
	assert.Error(t, err)
}

func Test_dcimModulebaytemplateDao_GetByIDs(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID).
		WillReturnRows(rows)

	_, err := d.IDao.(DcimModulebaytemplateDao).GetByIDs(d.Ctx, []uint64{testData.ID})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.IDao.(DcimModulebaytemplateDao).GetByIDs(d.Ctx, []uint64{111})
	assert.Error(t, err)

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebaytemplateDao_GetByLastID(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, err := d.IDao.(DcimModulebaytemplateDao).GetByLastID(d.Ctx, 0, 10, "")
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, err = d.IDao.(DcimModulebaytemplateDao).GetByLastID(d.Ctx, 0, 10, "unknown-column")
	assert.Error(t, err)
}

func Test_dcimModulebaytemplateDao_CreateByTx(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(d.GetAnyArgs(testData)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(DcimModulebaytemplateDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebaytemplateDao_DeleteByTx(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)
	expectedSQLForDeletion := "DELETE .*"

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(DcimModulebaytemplateDao).DeleteByTx(d.Ctx, d.DB, testData.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_dcimModulebaytemplateDao_UpdateByTx(t *testing.T) {
	d := newDcimModulebaytemplateDao()
	defer d.Close()
	testData := d.TestData.(*model.DcimModulebaytemplate)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(DcimModulebaytemplateDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
