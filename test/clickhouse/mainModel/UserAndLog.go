package mainModel

import (
	"github.com/go-preform/preform"
	"github.com/satori/go.uuid"
	"time"
)

var UserAndLog = preform.IniPrebuildQueryFactory[*UserAndLogFactory, UserAndLogBody](func(d *UserAndLogFactory) {
	d.User = d.PreformTestASchema.User.SetAlias("User").(*FactoryUser)
	d.UserLog = d.PreformTestASchema.UserLog.SetAlias("UserLog").(*FactoryUserLog)
	d.SetSrc(d.User).
		Join("Inner", d.UserLog, d.PreformTestASchema.UserLog.UserId.Eq(d.PreformTestASchema.User.Id)).DefineCols(
		preform.SetPrebuildQueryCol(d, d.UserLog.Id.SetAlias("UserLogId"), d.UserLogId),
		preform.SetPrebuildQueryCol(d, d.UserLog.Type.SetAlias("UserLogType"), d.UserLogType),
		preform.SetPrebuildQueryCol(d, d.User.CreatedAt.SetAlias("UserCreatedAt"), d.UserCreatedAt),
		preform.SetPrebuildQueryCol(d, d.User.LoginedAt.SetAlias("UserLoginedAt"), d.UserLoginedAt),
		preform.SetPrebuildQueryCol(d, d.User.DetailAge.SetAlias("UserDetailAge"), d.UserDetailAge),
		preform.SetPrebuildQueryCol(d, d.User.DetailDateOfBirth.SetAlias("UserDetailDateOfBirth"), d.UserDetailDateOfBirth),
		preform.SetPrebuildQueryCol(d, d.User.ManagerIds.SetAlias("UserManagerIds"), d.UserManagerIds),
		preform.SetPrebuildQueryCol(d, d.User.Name.SetAlias("UserName"), d.UserName),
		preform.SetPrebuildQueryCol(d, d.UserLog.Detail.SetAlias("UserLogDetail"), d.UserLogDetail),
		preform.SetPrebuildQueryCol(d, d.User.Id.SetAlias("UserId"), d.UserId),
		preform.SetPrebuildQueryCol(d, d.User.CreatedBy.SetAlias("UserCreatedBy"), d.UserCreatedBy),
		preform.SetPrebuildQueryCol(d, d.UserLog.UserId.SetAlias("UserLogUserId"), d.UserLogUserId),
		preform.SetPrebuildQueryCol(d, d.UserLog.RelatedLogId.SetAlias("UserLogRelatedLogId"), d.UserLogRelatedLogId),
	).
	PreSetWhere(d.PreformTestASchema.UserLog.UserId.NotEq(2))
})

type UserAndLogFactory struct {
	preform.PrebuildQueryFactory[*UserAndLogFactory, UserAndLogBody]
	//schema src
	PreformTestASchema *PreformTestASchema

	//factory src
	User *FactoryUser
	UserLog *FactoryUserLog
	
	//columns
	UserLogId *preform.PrebuildQueryCol[uuid.UUID, preform.NoAggregation]
	UserLogType *preform.PrebuildQueryCol[PreformTestAUserLogType, preform.NoAggregation]
	UserCreatedAt *preform.PrebuildQueryCol[time.Time, preform.NoAggregation]
	UserLoginedAt *preform.PrebuildQueryCol[*time.Time, preform.NoAggregation]
	UserDetailAge *preform.PrebuildQueryCol[[]int32, preform.NoAggregation]
	UserDetailDateOfBirth *preform.PrebuildQueryCol[[]time.Time, preform.NoAggregation]
	UserManagerIds *preform.PrebuildQueryCol[[]int32, preform.NoAggregation]
	UserName *preform.PrebuildQueryCol[string, preform.NoAggregation]
	UserLogDetail *preform.PrebuildQueryCol[map[string]string, preform.NoAggregation]
	UserId *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserCreatedBy *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserLogUserId *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserLogRelatedLogId *preform.PrebuildQueryCol[uuid.NullUUID, preform.NoAggregation]
}

type UserAndLogBody struct {
	preform.QueryBody[UserAndLogBody, *UserAndLogFactory]
	UserLogId uuid.UUID `db:"UserLogId" json:"Id" dataType:"UUID"`
	UserLogType PreformTestAUserLogType `db:"UserLogType" json:"Type" dataType:"Enum8('Register' = 1, 'Login' = 2)"`
	UserCreatedAt time.Time `db:"UserCreatedAt" json:"CreatedAt" dataType:"DateTime"`
	UserLoginedAt *time.Time `db:"UserLoginedAt" json:"LoginedAt" dataType:"Nullable(DateTime)"`
	UserDetailAge []int32 `db:"UserDetailAge" json:"DetailAge" dataType:"Array(Int32)"`
	UserDetailDateOfBirth []time.Time `db:"UserDetailDateOfBirth" json:"DetailDateOfBirth" dataType:"Array(Date)"`
	UserManagerIds []int32 `db:"UserManagerIds" json:"ManagerIds" dataType:"Array(Int32)" comment:"fk:preform_test_a.user.id"`
	UserName string `db:"UserName" json:"Name" dataType:"String"`
	UserLogDetail map[string]string `db:"UserLogDetail" json:"Detail" dataType:"Map(String, String)"`
	UserId int32 `db:"UserId" json:"Id" dataType:"Int32"`
	UserCreatedBy int32 `db:"UserCreatedBy" json:"CreatedBy" dataType:"Int32" comment:"fk:preform_test_a.user.id"`
	UserLogUserId int32 `db:"UserLogUserId" json:"UserId" dataType:"Int32" comment:"fk:preform_test_a.user.id"`
	UserLogRelatedLogId uuid.NullUUID `db:"UserLogRelatedLogId" json:"RelatedLogId" dataType:"Nullable(UUID)" comment:"fk:preform_test_a.user_log.id"`
}

func (m UserAndLogBody) Factory() *UserAndLogFactory { return UserAndLog }

func (m *UserAndLogBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.UserLogId
		case 1: return &m.UserLogType
		case 2: return &m.UserCreatedAt
		case 3: return &m.UserLoginedAt
		case 4: return &m.UserDetailAge
		case 5: return &m.UserDetailDateOfBirth
		case 6: return &m.UserManagerIds
		case 7: return &m.UserName
		case 8: return &m.UserLogDetail
		case 9: return &m.UserId
		case 10: return &m.UserCreatedBy
		case 11: return &m.UserLogUserId
		case 12: return &m.UserLogRelatedLogId
	}
	return nil
}

func (m *UserAndLogBody) FieldValuePtrs() []any { 
	return []any{&m.UserLogId, &m.UserLogType, &m.UserCreatedAt, &m.UserLoginedAt, &m.UserDetailAge, &m.UserDetailDateOfBirth, &m.UserManagerIds, &m.UserName, &m.UserLogDetail, &m.UserId, &m.UserCreatedBy, &m.UserLogUserId, &m.UserLogRelatedLogId}
}


