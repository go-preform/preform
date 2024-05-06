package model_test

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/test/pg/config"
	"github.com/go-preform/preform/test/pg/mainModel"
	"github.com/go-preform/preform/test/pg/mainModel/src/types"
	preformTestUtil "github.com/go-preform/preform/testUtil"
	preformTracer "github.com/go-preform/preform/tracer"
	preformTypes "github.com/go-preform/preform/types"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	pgxConn *sql.DB
)

func init() {
	pgxConn, _ = sql.Open("pgx", config.PgConnStr)
}

func TestInit(t *testing.T) {
	mainModel.Init(pgxConn)
	_, err := mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user CASCADE;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user_log CASCADE;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.foo CASCADE;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_b.bar CASCADE;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("ALTER SEQUENCE preform_test_a.user_id_seq RESTART WITH 1;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("ALTER SEQUENCE preform_test_a.foo_id_seq RESTART WITH 1;")
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

var (
	parentUuid = uuid.NewV4()
)

func TestUserInsert(t *testing.T) {
	_, err := mainModel.PreformTestA.Exec("SET session_replication_role = 'replica';")
	assert.Nil(t, err)

	user := mainModel.UserBody{
		Name: "test1",
		Detail: preformTypes.NewNull(preformTypes.NewJsonRaw(&types.UserDetail{
			Age: 11,
		})),
		UserLogs: []*mainModel.UserLogBody{
			{
				Id:   parentUuid,
				Type: mainModel.PreformTestALogTypes.Register,
				Detail: mainModel.PreformTestALogDetail{
					UserAgent: "test1",
					SessionId: uuid.NewV4(),
				},
			},
			{
				Id:           uuid.NewV4(),
				Type:         mainModel.PreformTestALogTypes.Login,
				RelatedLogId: preformTypes.NewNull(parentUuid),
				Detail: mainModel.PreformTestALogDetail{
					UserAgent: "test2",
				},
			},
		},
	}
	err = user.Insert(preform.EditConfig{Cascading: true})
	assert.Nil(t, err)
	assert.Equal(t, int32(1), user.Id)

	_, err = mainModel.PreformTestA.Exec("SET session_replication_role = 'origin';")
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
			Name:       "test2",
			ManagerIds: preformTypes.Array[int32]{1},
			CreatedBy:  1,
		},
		{
			Name:       "test3",
			ManagerIds: preformTypes.Array[int32]{1, 2},
			CreatedBy:  2,
		},
	}
	err = mainModel.PreformTestA.User.Insert(users)
	assert.Nil(t, err)
	assert.Equal(t, int32(2), users[0].Id)
	assert.Equal(t, int32(3), users[1].Id)

	users = []*mainModel.UserBody{
		{
			Name:       "test4",
			ManagerIds: preformTypes.Array[int32]{1},
			CreatedBy:  1,
		},
		{
			Name:       "test5",
			ManagerIds: preformTypes.Array[int32]{1, 2},
			CreatedBy:  2,
			UserLogs: []*mainModel.UserLogBody{
				{
					Id:   parentUuid,
					Type: mainModel.PreformTestALogTypes.Register,
					Detail: mainModel.PreformTestALogDetail{
						UserAgent: "test1",
						SessionId: uuid.NewV4(),
					},
				},
				{
					Id:           uuid.NewV4(),
					Type:         mainModel.PreformTestALogTypes.Login,
					RelatedLogId: preformTypes.NewNull(parentUuid),
					Detail: mainModel.PreformTestALogDetail{
						UserAgent: "test2",
					},
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
}

func TestUserSelect(t *testing.T) {
	user, err := mainModel.PreformTestA.User.GetOne(1)
	assert.Nil(t, err)
	assert.Equal(t, "test1", user.Name)
	assert.Equal(t, int32(1), user.Id)
	assert.True(t, user.Detail.Valid)
	assert.Equal(t, uint32(11), user.Detail.V.Get().Age)

	mainModel.PreformTestA.User.Use(func(f *mainModel.FactoryUser) {
		user, err := f.Select(f.Id).GetOne(1)
		assert.Nil(t, err)
		assert.Equal(t, int32(1), user.Id)
		assert.Equal(t, "", user.Name)
		assert.Len(t, user.ManagerIds, 0)
		assert.Equal(t, false, user.Detail.Valid)
	})

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
	assert.Equal(t, false, users[0].Detail.Valid)

	users, err = mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.ManagerIds.Eq(preformTypes.Array[int32]{1, 2})).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 2)

}

func TestUserSelectRelation(t *testing.T) {
	user, err := mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs).GetOne(1)
	assert.Nil(t, err)
	assert.Len(t, user.UserLogs, 2)
	assert.Equal(t, "test1", user.UserLogs[0].Detail.UserAgent)
	assert.Equal(t, "test2", user.UserLogs[1].Detail.UserAgent)
	assert.Equal(t, uuid.UUID{}, user.UserLogs[1].Detail.SessionId)
	logs, err := user.LoadUserLogsByUserLogUserFkRegister()
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	members, err := user.LoadUsersByUserUserFk()
	assert.Nil(t, err)
	assert.Len(t, members, 2)
	user, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogsByUserLogUserFkRegister).GetOne(1)
	assert.Nil(t, err)
	assert.Len(t, user.UserLogsByUserLogUserFkRegister, 1)

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).Eager(mainModel.PreformTestA.User.UserByUserUserFk).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
	assert.NotNil(t, users[0].UserByUserUserFk)
	assert.NotNil(t, users[1].UserByUserUserFk)
	assert.NotNil(t, users[2].UserByUserUserFk)
	assert.NotNil(t, users[3].UserByUserUserFk)

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
	assert.Equal(t, preformTypes.Array[int32]{1, 2}, user.ManagerIds)
	assert.Equal(t, int32(2), user.CreatedBy)
	assert.Equal(t, int32(5), user.Id)

	user.Name = "test5-1"
	user.ManagerIds = append(user.ManagerIds, 3)
	user.CreatedBy = 3
	user.Detail.Valid = true
	user.Detail.V.Set(&types.UserDetail{
		Age: 18,
	})
	user.Config = preformTypes.NewJsonRaw[*types.UserConfig](nil)
	user.ExtraConfig = preformTypes.NewJsonRaw[types.UserConfig](types.UserConfig{})
	affected, err := user.Update(preform.UpdateConfig{Cols: []preform.ICol{
		mainModel.PreformTestA.User.Name,
		mainModel.PreformTestA.User.ManagerIds,
		//mainModel.PreformTestA.User.CreatedBy,
		mainModel.PreformTestA.User.Detail,
		mainModel.PreformTestA.User.Config,
	}})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), affected)

	user, err = mainModel.PreformTestA.User.GetOne(5)
	assert.Nil(t, err)
	assert.Equal(t, "test5-1", user.Name)
	assert.Equal(t, preformTypes.Array[int32]{1, 2, 3}, user.ManagerIds)
	assert.Equal(t, int32(2), user.CreatedBy)
	assert.Equal(t, int32(5), user.Id)
	assert.Equal(t, uint32(18), user.Detail.V.Get().Age)
	assert.Equal(t, "{}", string(user.Config.Src()))
	assert.Equal(t, "{}", string(user.ExtraConfig.Src()))

	userLog, err := mainModel.PreformTestA.UserLog.GetOne(parentUuid)
	assert.Nil(t, err)
	assert.False(t, userLog.RelatedLogId.Valid)
	assert.Equal(t, "test1", userLog.Detail.UserAgent)
	userLog.Detail.UserAgent = "test1-1"
	affected, err = userLog.Update()
	assert.Nil(t, err)
	userLog, err = mainModel.PreformTestA.UserLog.GetOne(parentUuid)
	assert.Nil(t, err)
	assert.Equal(t, "test1-1", userLog.Detail.UserAgent)

	affected, err = mainModel.PreformTestA.User.Update().Set(mainModel.PreformTestA.User.CreatedAt, time.Now()).Where(mainModel.PreformTestA.User.Id.Gt(2)).Exec()
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
	users, err = mainModel.UserAndLog.Select(mainModel.UserAndLog.UserName).Where(mainModel.UserAndLog.UserLogId.NotEq(parentUuid)).GetAll()
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
		Id:         1,
		Name:       "dummy",
		ManagerIds: preformTypes.Array[int32]{1},
		CreatedBy:  1,
		CreatedAt:  time.Now(),
	}
	queryRunner.AddToQueryRows([][]driver.Value{{[]string{"id", "name", "manager_ids", "created_by", "created_at", "logined_at", "detail", "config", "extra_config"}}, {1, dummyUser.Name, "{1}", 1, dummyUser.CreatedAt, nil, nil, "{}", "{}"}})
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
	fakeUUid := uuid.NewV4()
	userLogScanner.BodiesQueue = append(userLogScanner.BodiesQueue, []mainModel.UserLogBody{{Id: fakeUUid, UserId: 1}})
	users, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.Len(t, users[0].UserLogs, 1)
	assert.Equal(t, fakeUUid, users[0].UserLogs[0].Id)

}
