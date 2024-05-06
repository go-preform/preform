package model_test

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"flag"
	"fmt"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/test/sqlite/config"
	"github.com/go-preform/preform/test/sqlite/mainModel"
	"github.com/go-preform/preform/test/sqlite/mainModel/src/types"
	preformTestUtil "github.com/go-preform/preform/testUtil"
	preformTracer "github.com/go-preform/preform/tracer"
	preformTypes "github.com/go-preform/preform/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var (
	conn *sql.DB
)

func TestMain(m *testing.M) {
	wdd := flag.String("root", "", "")
	flag.Parse()
	wd, _ := os.Getwd()
	if wdd != nil && *wdd != "" {
		wd = *wdd
	}
	conn = config.InitDb(wd)
	_, err := conn.Exec(fmt.Sprintf(`
attach database '%s/preform_test_a.db' as 'preform_test_a';
attach database '%s/preform_test_b.db' as preform_test_b;
`, wd, wd))
	if err != nil {
		panic(err)
	}
}

func TestInit(t *testing.T) {
	mainModel.Init(conn)
	_, err := mainModel.PreformTestA.Exec("DELETE FROM preform_test_a.user;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("DELETE FROM preform_test_a.user_log;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("DELETE FROM preform_test_a.foo;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("DELETE FROM preform_test_b.bar;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("delete from preform_test_a.sqlite_sequence where name='user';")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("delete from preform_test_a.sqlite_sequence where name='user_log';")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("delete from preform_test_a.sqlite_sequence where name='foo';")
	assert.Nil(t, err)
}

func TestTrace(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	mainModel.PreformTestA.SetTracerToDb(preformTracer.NewZeroLogTracer(zerolog.New(writer), 7, 0))
	//_, err := mainModel.PreformTestA.Query("SELECT * FROM preform_test_a.user")
	//assert.Nil(t, err)
	//log := writer.String()
	//assert.Contains(t, log, "SELECT * FROM preform_test_a.user")
	mainModel.PreformTestA.SetTracerToDb(preformTracer.NewChainTracer(preformTracer.NewZeroLogTracer(zerolog.New(zerolog.NewConsoleWriter()), 7, 0)))
}

var (
	parentUuid int64 = 1
)

func TestUserInsert(t *testing.T) {

	user := mainModel.UserBody{
		Name: "test1",
		Detail: preformTypes.NewNull(preformTypes.NewJsonRaw(&types.UserDetail{
			Age: 11,
		})),
		UserLogs: []*mainModel.UserLogBody{
			{
				Id:   parentUuid,
				Type: 1,
			},
			{
				Id:           2,
				Type:         2,
				RelatedLogId: preformTypes.NewNull(parentUuid),
			},
		},
	}
	err := user.Insert(preform.EditConfig{Cascading: true})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), user.Id)

	foos := []mainModel.FooBody{
		{
			Fk: 1,
		},
		{
			Fk: 2,
		},
	}

	err = mainModel.PreformTestA.Foo.InsertBatch(foos)
	assert.Nil(t, err)

	bars := []mainModel.BarBody{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
	}

	err = mainModel.PreformTestB.Bar.InsertBatch(bars)
	assert.Nil(t, err)

	users := []*mainModel.UserBody{
		{
			Name:      "test2",
			CreatedBy: 1,
		},
		{
			Name:      "test3",
			CreatedBy: 2,
		},
	}
	err = mainModel.PreformTestA.User.Insert(users)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), users[0].Id)
	assert.Equal(t, int64(3), users[1].Id)

	users = []*mainModel.UserBody{
		{
			Name:      "test4",
			CreatedBy: 1,
		},
		{
			Name:      "test5",
			CreatedBy: 2,
			UserLogs: []*mainModel.UserLogBody{
				{
					Id:   parentUuid,
					Type: 1,
				},
				{
					Id:           2,
					Type:         2,
					RelatedLogId: preformTypes.NewNull(parentUuid),
				},
			},
		},
	}

	err = mainModel.PreformTestA.User.InsertBatch(users)
	assert.Nil(t, err)

	allUsers, err := mainModel.PreformTestA.User.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, allUsers, 5)
}

func TestUserSelect(t *testing.T) {
	user, err := mainModel.PreformTestA.User.GetOne(1)
	assert.Nil(t, err)
	assert.Equal(t, "test1", user.Name)
	assert.Equal(t, int64(1), user.Id)
	assert.True(t, user.Detail.Valid)
	assert.Equal(t, uint32(11), user.Detail.V.Get().Age)

	mainModel.PreformTestA.User.Use(func(f *mainModel.FactoryUser) {
		user, err := f.Select(f.Id).GetOne(1)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, "", user.Name)
		assert.Equal(t, false, user.Detail.Valid)
	})

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
	assert.Equal(t, false, users[0].Detail.Valid)

}

func TestUserSelectRelation(t *testing.T) {
	user, err := mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs).GetOne(1)
	assert.Nil(t, err)
	assert.Len(t, user.UserLogs, 2)
	logs, err := user.LoadUserLogsByUserLogUserFkRegister()
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	members, err := user.LoadUsersByFkUserCreatedById()
	assert.Nil(t, err)
	assert.Len(t, members, 2)
	user, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogsByUserLogUserFkRegister).GetOne(1)
	assert.Nil(t, err)
	assert.Len(t, user.UserLogsByUserLogUserFkRegister, 1)

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).Eager(mainModel.PreformTestA.User.UserByFkUserCreatedById).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
	assert.NotNil(t, users[0].UserByFkUserCreatedById)
	assert.NotNil(t, users[1].UserByFkUserCreatedById)
	assert.NotNil(t, users[2].UserByFkUserCreatedById)
	assert.NotNil(t, users[3].UserByFkUserCreatedById)

	rows, err := mainModel.PreformTestA.User.Select("*").JoinRelation(mainModel.PreformTestA.User.UserLogs).Where(mainModel.PreformTestA.User.Id.Eq(1)).QueryRaw()
	assert.Nil(t, err)
	assert.Len(t, rows.Rows, 2)

	foos, err := mainModel.PreformTestA.Foo.Select().Eager(mainModel.PreformTestA.Foo.Bar).GetAll()
	assert.Nil(t, err)
	assert.Len(t, foos, 2)
	assert.NotNil(t, foos[0].Bar)
	assert.NotNil(t, foos[1].Bar)
}

func TestUserUpdate(t *testing.T) {
	user, err := mainModel.PreformTestA.User.GetOne(5)
	assert.Nil(t, err)
	assert.Equal(t, "test5", user.Name)
	assert.Equal(t, int64(2), user.CreatedBy)
	assert.Equal(t, int64(5), user.Id)

	user.Name = "test5-1"
	user.CreatedBy = 3
	user.Detail.Valid = true
	user.Detail.V.Set(&types.UserDetail{
		Age: 18,
	})
	user.Config = preformTypes.NewJsonRaw[*types.UserConfig](nil)
	user.ExtraConfig = preformTypes.NewJsonRaw[types.UserConfig](types.UserConfig{})
	affected, err := user.Update(preform.UpdateConfig{Cols: []preform.ICol{
		mainModel.PreformTestA.User.Name,
		//mainModel.PreformTestA.User.CreatedBy,
		mainModel.PreformTestA.User.Detail,
		mainModel.PreformTestA.User.Config,
	}})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	user, err = mainModel.PreformTestA.User.GetOne(5)
	assert.Nil(t, err)
	assert.Equal(t, "test5-1", user.Name)
	assert.Equal(t, int64(2), user.CreatedBy)
	assert.Equal(t, int64(5), user.Id)
	assert.Equal(t, uint32(18), user.Detail.V.Get().Age)
	assert.Equal(t, "null", string(user.Config.Src()))
	assert.Equal(t, "", string(user.ExtraConfig.Src()))

	userLog, err := mainModel.PreformTestA.UserLog.GetOne(parentUuid)
	assert.Nil(t, err)
	assert.False(t, userLog.RelatedLogId.Valid)
	affected, err = userLog.Update()
	assert.Nil(t, err)
	userLog, err = mainModel.PreformTestA.UserLog.GetOne(parentUuid)
	assert.Nil(t, err)

	affected, err = mainModel.PreformTestA.User.Update().Set(mainModel.PreformTestA.User.CreatedAt, preformTypes.SqliteTime(time.Now())).Where(mainModel.PreformTestA.User.Id.Gt(2)).Exec()
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

	affected, err = user.Delete()
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	allUsers, err := mainModel.PreformTestA.User.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, allUsers, 3)
}

func TestPrebuildQuery(t *testing.T) {
	userLogs, err := mainModel.UserAndLog.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, userLogs, 2)
	assert.Equal(t, int64(1), userLogs[0].UserId)
	assert.Equal(t, int64(1), userLogs[1].UserId)
	userLogs, err = mainModel.UserAndLog.Select().GetAllFast()
	assert.Nil(t, err)
	assert.Len(t, userLogs, 2)
	assert.Equal(t, int64(1), userLogs[0].UserId)
	assert.Equal(t, int64(1), userLogs[1].UserId)
	userLogs, err = mainModel.UserAndLog.Select(mainModel.UserAndLog.UserName).Where(mainModel.UserAndLog.UserLogId.NotEq(parentUuid)).GetAll()
	assert.Nil(t, err)
	assert.Len(t, userLogs, 1)
	assert.Equal(t, int64(0), userLogs[0].UserId)
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
		CreatedAt: preformTypes.SqliteTime(time.Now()),
	}
	queryRunner.AddToQueryRows([][]driver.Value{{[]string{"id", "name", "created_by", "created_at", "logined_at", "detail", "config", "extra_config"}}, {1, dummyUser.Name, 1, dummyUser.CreatedAt, nil, nil, "{}", "{}"}})
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
	fakeUUid := int64(123)
	userLogScanner.BodiesQueue = append(userLogScanner.BodiesQueue, []mainModel.UserLogBody{{Id: fakeUUid, UserId: 1}})
	users, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.Len(t, users[0].UserLogs, 1)
	assert.Equal(t, fakeUUid, users[0].UserLogs[0].Id)
}
