package model_test

import (
	"bytes"
	"context"
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/test/clickhouse/config"
	"github.com/go-preform/preform/test/clickhouse/mainModel"
	preformTracer "github.com/go-preform/preform/tracer"
	preformTypes "github.com/go-preform/preform/types"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	chConn *sql.DB
)

func init() {
	chConn = config.ChConn()
}

func TestInit(t *testing.T) {
	mainModel.Init(chConn)
	_, err := mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user;")
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.Exec("TRUNCATE TABLE preform_test_a.user_log;")
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

	tx, err := mainModel.PreformTestA.BeginTx(context.Background())
	assert.Nil(t, err)
	user := mainModel.UserBody{
		Id:        1,
		Name:      "test1",
		CreatedAt: time.Now(),
		UserLogs: []*mainModel.UserLogBody{
			{
				Id:   parentUuid,
				Type: mainModel.PreformTestAUserLogTypes.Register,
				Detail: map[string]string{
					"UserAgent": "test1",
					"SessionId": uuid.NewV4().String(),
				},
			},
			{
				Id:           uuid.NewV4(),
				Type:         mainModel.PreformTestAUserLogTypes.Login,
				RelatedLogId: uuid.NullUUID{UUID: parentUuid, Valid: true},
				Detail: map[string]string{
					"UserAgent": "test2",
				},
			},
		},
		DetailAge:         []int32{11},
		DetailDateOfBirth: []time.Time{time.Now().Add(-time.Hour * 24 * 365 * 11)},
	}
	err = user.Insert(preform.EditConfig{Cascading: true, Tx: tx})
	assert.Nil(t, err)
	assert.Equal(t, int32(1), user.Id)

	users := []*mainModel.UserBody{
		{
			Id:         2,
			Name:       "test2",
			ManagerIds: preformTypes.Array[int32]{1},
			CreatedBy:  1,
			CreatedAt:  time.Now(),
		},
		{
			Id:         3,
			Name:       "test3",
			ManagerIds: preformTypes.Array[int32]{1, 2},
			CreatedBy:  2,
			CreatedAt:  time.Now(),
		},
	}
	err = mainModel.PreformTestA.User.Insert(users, preform.EditConfig{Tx: tx})
	assert.Nil(t, err)
	assert.Equal(t, int32(2), users[0].Id)
	assert.Equal(t, int32(3), users[1].Id)

	users = []*mainModel.UserBody{
		{
			Id:         4,
			Name:       "test4",
			ManagerIds: preformTypes.Array[int32]{1},
			CreatedBy:  1,
			CreatedAt:  time.Now(),
		},
		{
			Id:         5,
			Name:       "test5",
			ManagerIds: preformTypes.Array[int32]{1, 2},
			CreatedBy:  2,
			CreatedAt:  time.Now(),
			UserLogs: []*mainModel.UserLogBody{
				{
					Id:   parentUuid,
					Type: mainModel.PreformTestAUserLogTypes.Register,
					Detail: map[string]string{
						"UserAgent": "test1",
						"SessionId": uuid.NewV4().String(),
					},
				},
				{
					Id:           uuid.NewV4(),
					Type:         mainModel.PreformTestAUserLogTypes.Login,
					RelatedLogId: uuid.NullUUID{UUID: parentUuid, Valid: true},
					Detail:       map[string]string{"UserAgent": "test2"},
				},
			},
		},
	}
	err = mainModel.PreformTestA.User.InsertBatch(users, preform.EditConfig{Cascading: true, Tx: tx})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cascading is not supported for batch insert")

	err = mainModel.PreformTestA.User.InsertBatch(users, preform.EditConfig{Tx: tx})
	assert.Nil(t, err)

	allUsers, err := mainModel.PreformTestA.User.Select().GetAll()
	assert.Nil(t, err)
	assert.Len(t, allUsers, 5)

	assert.Nil(t, tx.Commit())
}

func TestUserSelect(t *testing.T) {
	user, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Eq(1)).GetOne()
	assert.Nil(t, err)
	assert.Equal(t, "test1", user.Name)
	assert.Equal(t, int32(1), user.Id)
	assert.Nil(t, user.LoginedAt)
	assert.Equal(t, user.DetailAge[0], int32(11))

	mainModel.PreformTestA.User.Use(func(f *mainModel.FactoryUser) {
		user, err := f.Select(f.Id).Where(mainModel.PreformTestA.User.Id.Eq(1)).GetOne()
		assert.Nil(t, err)
		assert.Equal(t, int32(1), user.Id)
		assert.Equal(t, "", user.Name)
		assert.Len(t, user.ManagerIds, 0)
	})

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)

	users, err = mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.ManagerIds.Eq([]int32{1, 2})).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 2)

}

func TestUserSelectRelation(t *testing.T) {
	user, err := mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogs.OrderBy((any(mainModel.PreformTestA.User.UserLogs.TargetFactory())).(*mainModel.FactoryUserLog).RelatedLogId.Desc())).Where(mainModel.PreformTestA.User.Id.Eq(1)).GetOne()
	assert.Nil(t, err)
	assert.Len(t, user.UserLogs, 2)
	assert.Equal(t, "test1", user.UserLogs[1].Detail["UserAgent"])
	assert.Equal(t, "test2", user.UserLogs[0].Detail["UserAgent"])
	assert.Equal(t, "", user.UserLogs[0].Detail["SessionId"])
	logs, err := user.LoadUserLogsByUserLogUserFkRegister()
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	members, err := user.LoadUsersByCommentCreatedBy0()
	assert.Nil(t, err)
	assert.Len(t, members, 2)
	user, err = mainModel.PreformTestA.User.Select().Eager(mainModel.PreformTestA.User.UserLogsByUserLogUserFkRegister).Where(mainModel.PreformTestA.User.Id.Eq(1)).GetOne()
	assert.Nil(t, err)
	assert.Len(t, user.UserLogsByUserLogUserFkRegister, 1)

	users, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Gt(1)).Eager(mainModel.PreformTestA.User.UserByCommentCreatedBy0).GetAll()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
	assert.NotNil(t, users[0].UserByCommentCreatedBy0)
	assert.NotNil(t, users[1].UserByCommentCreatedBy0)
	assert.NotNil(t, users[2].UserByCommentCreatedBy0)
	assert.NotNil(t, users[3].UserByCommentCreatedBy0)

	rows, err := mainModel.PreformTestA.User.Select("*").JoinRelation(mainModel.PreformTestA.User.UserLogs).Where(mainModel.PreformTestA.User.Id.Eq(1)).QueryRaw()
	assert.Nil(t, err)
	assert.Len(t, rows.Rows, 2)

}

func TestUserUpdate(t *testing.T) {
	user, err := mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Eq(5)).GetOne()
	assert.Nil(t, err)
	assert.Equal(t, "test5", user.Name)
	assert.Equal(t, []int32{1, 2}, user.ManagerIds)
	assert.Equal(t, int32(2), user.CreatedBy)
	assert.Equal(t, int32(5), user.Id)

	user.Name = "test5-1"
	user.ManagerIds = append(user.ManagerIds, 3)
	user.CreatedBy = 3
	user.DetailAge = []int32{12}
	user.DetailDateOfBirth = []time.Time{time.Now().Add(-time.Hour * 24 * 365 * 12)}

	tx, err := mainModel.PreformTestA.BeginTx(context.Background())
	assert.Nil(t, err)
	_, err = mainModel.PreformTestA.User.Update().SetBodies(user).Columns(mainModel.PreformTestA.User.Name,
		mainModel.PreformTestA.User.ManagerIds,
		//mainModel.PreformTestA.User.CreatedBy,
		mainModel.PreformTestA.User.DetailAge,
		mainModel.PreformTestA.User.DetailDateOfBirth).Where(mainModel.PreformTestA.User.Id.Eq(5)).Exec(tx)
	assert.Nil(t, err)
	//assert.Equal(t, int64(1), affected)

	userLog, err := mainModel.PreformTestA.UserLog.Select().Where(mainModel.PreformTestA.UserLog.Id.Eq(parentUuid)).GetOne()
	assert.Nil(t, err)
	assert.False(t, userLog.RelatedLogId.Valid)
	assert.Equal(t, "test1", userLog.Detail["UserAgent"])
	userLog.Detail["UserAgent"] = "test1-1"
	_, err = userLog.Update(preform.UpdateConfig{Tx: tx})
	assert.Nil(t, err)
	assert.Nil(t, tx.Commit())

	time.Sleep(time.Second)

	user, err = mainModel.PreformTestA.User.Select().Where(mainModel.PreformTestA.User.Id.Eq(5)).GetOne()
	assert.Nil(t, err)
	assert.Equal(t, "test5-1", user.Name)
	assert.Equal(t, []int32{1, 2, 3}, user.ManagerIds)
	assert.Equal(t, int32(2), user.CreatedBy)
	assert.Equal(t, int32(5), user.Id)
	assert.Equal(t, int32(12), user.DetailAge[0])
	userLog, err = mainModel.PreformTestA.UserLog.GetOne(parentUuid)
	assert.Nil(t, err)
	assert.Equal(t, "test1-1", userLog.Detail["UserAgent"])

	_, err = mainModel.PreformTestA.User.Update().Set(mainModel.PreformTestA.User.LoginedAt, time.Now()).Where(mainModel.PreformTestA.User.Id.Gt(2)).Exec()
	assert.Nil(t, err)
	//assert.Equal(t, int64(3), affected)
}

func TestUserDelete(t *testing.T) {
	tx, err := mainModel.PreformTestA.BeginTx(context.Background())
	_, err = mainModel.PreformTestA.User.Delete().Where(mainModel.PreformTestA.User.Id.Gt(4)).Exec(tx)
	assert.Nil(t, err)

	user, err := mainModel.PreformTestA.User.GetOne(3)
	assert.Nil(t, err)
	assert.Equal(t, "test3", user.Name)

	_, err = user.Delete(preform.EditConfig{Tx: tx})
	assert.Nil(t, err)

	assert.Nil(t, tx.Commit())

	time.Sleep(time.Second)

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
