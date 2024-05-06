package model_test

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/test/mysql/config"
	"github.com/go-preform/preform/test/mysql/mainModel"
	preformTestUtil "github.com/go-preform/preform/testUtil"
	preformTracer "github.com/go-preform/preform/tracer"
	preformTypes "github.com/go-preform/preform/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	myConn *sql.DB
)

func init() {
	myConn, _ = sql.Open("mysql", config.MysqlConnStr)
}

func TestInit(t *testing.T) {
	mainModel.Init(myConn)
	_, err := myConn.Exec("SET FOREIGN_KEY_CHECKS=0;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user_log;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user_manager;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.foo;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_b.bar;")
	assert.Nil(t, err)
	_, err = myConn.Exec("SET FOREIGN_KEY_CHECKS=1;")
	assert.Nil(t, err)
}

func TestTrace(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	mainModel.PreformTestA.SetTracerToDb(preformTracer.NewZeroLogTracer(zerolog.New(writer), 7, 0))
	_, err := mainModel.PreformTestA.Query("SELECT 1")
	assert.Nil(t, err)
	log := writer.String()
	assert.Contains(t, log, "SELECT 1")
	mainModel.PreformTestA.SetTracerToDb(preformTracer.NewChainTracer(preformTracer.NewZeroLogTracer(zerolog.New(zerolog.NewConsoleWriter()), 7, 0)))
}

func TestUserInsert(t *testing.T) {
	_, err := myConn.Exec("SET FOREIGN_KEY_CHECKS=0;")
	assert.Nil(t, err)

	user := mainModel.UserBody{
		Name:      "test1",
		CreatedAt: time.Now(),
		UserLogs: []*mainModel.UserLogBody{
			{
				Id:   1,
				Type: mainModel.PreformTestAUserLogTypes.Register,
			},
			{
				Id:           2,
				Type:         mainModel.PreformTestAUserLogTypes.Login,
				RelatedLogId: preformTypes.NewNull(int64(1)),
			},
		},
	}
	err = user.Insert(preform.EditConfig{Cascading: true})
	assert.Nil(t, err)
	assert.Equal(t, int32(1), user.Id)

	_, err = myConn.Exec("SET FOREIGN_KEY_CHECKS=1;")
	assert.Nil(t, err)

	foos := []mainModel.FooBody{
		{
			Fk1: 1,
			Fk2: 2,
		},
		{
			Fk1: 3,
			Fk2: 4,
		},
	}

	err = mainModel.PreformTestA.Foo.InsertBatch(foos)
	assert.Nil(t, err)

	bars := []mainModel.BarBody{
		{
			Id1: 1,
			Id2: 2,
		},
		{
			Id1: 3,
			Id2: 4,
		},
	}

	err = mainModel.PreformTestB.Bar.InsertBatch(bars)
	assert.Nil(t, err)

	users := []*mainModel.UserBody{
		{
			Name:      "test2",
			CreatedBy: 1,
			CreatedAt: time.Now(),
		},
		{
			Name:      "test3",
			CreatedBy: 2,
			CreatedAt: time.Now(),
		},
	}
	err = mainModel.PreformTestA.User.Insert(users)
	assert.Nil(t, err)
	assert.Equal(t, int32(2), users[0].Id)
	assert.Equal(t, int32(3), users[1].Id)

	users = []*mainModel.UserBody{
		{
			Name:      "test4",
			CreatedBy: 1,
			CreatedAt: time.Now(),
		},
		{
			Name:      "test5",
			CreatedBy: 2,
			CreatedAt: time.Now(),
			UserLogs: []*mainModel.UserLogBody{
				{
					Id:   3,
					Type: mainModel.PreformTestAUserLogTypes.Register,
				},
				{
					Id:           4,
					Type:         mainModel.PreformTestAUserLogTypes.Login,
					RelatedLogId: preformTypes.NewNull(int64(3)),
				},
			},
		},
	}
	err = mainModel.PreformTestA.User.InsertBatch(users, preform.EditConfig{Cascading: true})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cascading is not supported for batch insert")

	err = mainModel.PreformTestA.User.InsertBatch(users)
	assert.Nil(t, err)

	allUsers, err := mainModel.PreformTestA.User.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, allUsers, 5)

	err = mainModel.PreformTestA.User.UserByUserManagerUserId.LinkModels(mainModel.UserManagerBody{UserId: 2, ManagerId: 1}, mainModel.UserManagerBody{UserId: 3, ManagerId: 1}, mainModel.UserManagerBody{UserId: 3, ManagerId: 2})
	assert.Nil(t, err)
}

func TestUserSelect(t *testing.T) {
	user, err := mainModel.PreformTestA.User.GetOne(1)
	assert.Nil(t, err)
	assert.Equal(t, "test1", user.Name)
	assert.Equal(t, int32(1), user.Id)

	mainModel.PreformTestA.User.Use(func(f *mainModel.FactoryUser) {
		user, err := f.Select(f.Id).GetOne(1)
		assert.Nil(t, err)
		assert.Equal(t, int32(1), user.Id)
		assert.Equal(t, "", user.Name)
	})

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)

}

func TestUserSelectRelation(t *testing.T) {
	user, err := mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs).GetOne(1)
	assert.Nil(t, err)
	assert.Len(t, user.UserLogs, 2)
	logs, err := user.LoadUserLogsByUserLogUserFkRegister()
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	members, err := user.LoadUsersByUserFk()
	assert.Nil(t, err)
	assert.Len(t, members, 2)
	staffs, err := user.LoadUserByUserManagerUserId()
	assert.Nil(t, err)
	assert.Len(t, staffs, 2)
	user, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserByUserManagerUserId).GetOne(2)
	assert.Nil(t, err)
	assert.Len(t, user.UserByUserManagerUserId, 1)
	user, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserByUserManagerManagerId).GetOne(3)
	assert.Nil(t, err)
	assert.Len(t, user.UserByUserManagerManagerId, 2)
	user, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogsByUserLogUserFkRegister).GetOne(1)
	assert.Nil(t, err)
	assert.Len(t, user.UserLogsByUserLogUserFkRegister, 1)

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).Eager(mainModel.PreformTestA.User.UserByUserFk).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
	assert.NotNil(t, users[0].UserByUserFk)
	assert.NotNil(t, users[1].UserByUserFk)
	assert.NotNil(t, users[2].UserByUserFk)
	assert.NotNil(t, users[3].UserByUserFk)

	rows, err := mainModel.PreformTestA.User.Select("*").JoinRelation(mainModel.PreformTestA.User.UserLogs).Where(mainModel.PreformTestA.User.Id.Eq(1)).QueryRaw()
	assert.Nil(t, err)
	assert.Len(t, rows.Rows, 2)

	foos, err := mainModel.PreformTestA.Foo.Select().Eager(mainModel.PreformTestA.Foo.Bars).GetAll()
	assert.Nil(t, err)
	assert.Len(t, foos, 2)
	assert.Len(t, foos[0].Bars, 1)
	assert.Len(t, foos[1].Bars, 1)
}

func TestUserUpdate(t *testing.T) {
	user, err := mainModel.PreformTestA.User.GetOne(5)
	assert.Nil(t, err)
	assert.Equal(t, "test5", user.Name)
	assert.Equal(t, int32(2), user.CreatedBy)
	assert.Equal(t, int32(5), user.Id)

	user.Name = "test5-1"
	user.CreatedBy = 3
	affected, err := user.Update(preform.UpdateConfig{Cols: []preform.ICol{
		mainModel.PreformTestA.User.Name,
		//mainModel.PreformTestA.User.CreatedBy,
	}})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	user, err = mainModel.PreformTestA.User.GetOne(5)
	assert.Nil(t, err)
	assert.Equal(t, "test5-1", user.Name)
	assert.Equal(t, int32(2), user.CreatedBy)
	assert.Equal(t, int32(5), user.Id)

	userLog, err := mainModel.PreformTestA.UserLog.GetOne(1)
	assert.Nil(t, err)
	assert.False(t, userLog.RelatedLogId.Valid)
	userLog.RelatedLogId = preformTypes.NewNull(int64(2))
	affected, err = userLog.Update()
	assert.Nil(t, err)
	userLog, err = mainModel.PreformTestA.UserLog.GetOne(1)
	assert.Nil(t, err)
	assert.True(t, userLog.RelatedLogId.Valid)
	assert.Equal(t, int64(2), userLog.RelatedLogId.V)

	affected, err = mainModel.PreformTestA.User.Update().Set(mainModel.PreformTestA.User.CreatedAt, time.Now().Add(time.Hour*24)).Where(mainModel.PreformTestA.User.Id.Gt(2)).Exec()
	assert.Nil(t, err)
	assert.Equal(t, int64(3), affected)
}

func TestUserDelete(t *testing.T) {
	affected, err := mainModel.PreformTestA.User.Delete().Where(mainModel.PreformTestA.User.Id.Gt(4)).Exec()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	user, err := mainModel.PreformTestA.User.GetOne(3)
	assert.Nil(t, err)
	assert.Equal(t, "test3", user.Name)

	affected, err = user.Delete(preform.EditConfig{Cascading: true})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	allUsers, err := mainModel.PreformTestA.User.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, allUsers, 3)
}

func TestPrebuildQuery(t *testing.T) {
	users, err := mainModel.UserAndLog.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, int32(1), users[0].UserId)
	assert.Equal(t, int32(1), users[1].UserId)
	users, err = mainModel.UserAndLog.Select().GetAllFast()
	assert.Nil(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, int32(1), users[0].UserId)
	assert.Equal(t, int32(1), users[1].UserId)
	users, err = mainModel.UserAndLog.Select(mainModel.UserAndLog.UserName).Where(mainModel.UserAndLog.UserLogId.NotEq(1)).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, int32(0), users[0].UserId)
}

func TestTesters(t *testing.T) {
	dummyDb := preformTestUtil.NewTestDB("pgx")
	mainModel.Init(dummyDb)
	queryRunner := preformTestUtil.NewTestQueryRunner()
	mainModel.PreformTestA.SetConn(dummyDb, queryRunner)
	dummyUser := mainModel.UserBody{
		Id:        1,
		Name:      "dummy",
		CreatedBy: 1,
		CreatedAt: time.Now(),
	}
	queryRunner.AddToQueryRows([][]driver.Value{{[]string{"id", "name", "created_by", "created_at", "logined_at"}}, {1, dummyUser.Name, 1, dummyUser.CreatedAt, nil}})
	users, err := mainModel.PreformTestA.User.GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	queryRunner.ErrorQueue = append(queryRunner.ErrorQueue, sql.ErrNoRows)
	_, err = mainModel.PreformTestA.User.GetOne(1)
	assert.Equal(t, sql.ErrNoRows, err)
	userScanner := preformTestUtil.NewTestModelScanner[mainModel.UserBody]()
	mainModel.PreformTestA.User.SetModelScanner(userScanner)
	userLogScanner := preformTestUtil.NewTestModelScanner[mainModel.UserLogBody]()
	mainModel.PreformTestA.UserLog.SetModelScanner(userLogScanner)
	userScanner.BodiesQueue = append(userScanner.BodiesQueue, []mainModel.UserBody{dummyUser})
	userLogScanner.BodiesQueue = append(userLogScanner.BodiesQueue, []mainModel.UserLogBody{{Id: 9527, UserId: 1}})
	users, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.Len(t, users[0].UserLogs, 1)
	assert.Equal(t, int64(9527), users[0].UserLogs[0].Id)

}
