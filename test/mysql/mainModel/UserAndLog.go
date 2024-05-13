package mainModel

import (
	"github.com/go-preform/preform"
	"time"
	"github.com/go-preform/preform/types"
)

var UserAndLog = preform.IniPrebuildQueryFactory[*UserAndLogFactory, UserAndLogBody](func(d *UserAndLogFactory) {
	d.User = d.PreformTestASchema.User.SetAlias("User").(*FactoryUser)
	d.UserLog = d.PreformTestASchema.UserLog.SetAlias("UserLog").(*FactoryUserLog)
	d.SetSrc(d.User).
		Join("Inner", d.UserLog, d.PreformTestASchema.UserLog.UserId.Eq(d.PreformTestASchema.User.Id)).DefineCols(
		preform.SetPrebuildQueryCol(d, d.User.Name.SetAlias("UserName"), d.UserName),
		preform.SetPrebuildQueryCol(d, d.User.CreatedAt.SetAlias("UserCreatedAt"), d.UserCreatedAt),
		preform.SetPrebuildQueryCol(d, d.User.LoginedAt.SetAlias("UserLoginedAt"), d.UserLoginedAt),
		preform.SetPrebuildQueryCol(d, d.UserLog.Id.SetAlias("UserLogId"), d.UserLogId),
		preform.SetPrebuildQueryCol(d, d.UserLog.RelatedLogId.SetAlias("UserLogRelatedLogId"), d.UserLogRelatedLogId),
		preform.SetPrebuildQueryCol(d, d.UserLog.Type.SetAlias("UserLogType"), d.UserLogType),
		preform.SetPrebuildQueryCol(d, d.User.Id.SetAlias("UserId"), d.UserId),
		preform.SetPrebuildQueryCol(d, d.User.CreatedBy.SetAlias("UserCreatedBy"), d.UserCreatedBy),
		preform.SetPrebuildQueryCol(d, d.UserLog.UserId.SetAlias("UserLogUserId"), d.UserLogUserId),
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
	UserName *preform.PrebuildQueryCol[string, preform.NoAggregation]
	UserCreatedAt *preform.PrebuildQueryCol[time.Time, preform.NoAggregation]
	UserLoginedAt *preform.PrebuildQueryCol[preformTypes.Null[time.Time], preform.NoAggregation]
	UserLogId *preform.PrebuildQueryCol[int64, preform.NoAggregation]
	UserLogRelatedLogId *preform.PrebuildQueryCol[preformTypes.Null[int64], preform.NoAggregation]
	UserLogType *preform.PrebuildQueryCol[PreformTestAUserLogType, preform.NoAggregation]
	UserId *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserCreatedBy *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserLogUserId *preform.PrebuildQueryCol[int32, preform.NoAggregation]
}

type UserAndLogBody struct {
	preform.QueryBody[UserAndLogBody, *UserAndLogFactory]
	UserName string `db:"UserName" json:"Name" dataType:"varchar"`
	UserCreatedAt time.Time `db:"UserCreatedAt" json:"CreatedAt" dataType:"datetime"`
	UserLoginedAt preformTypes.Null[time.Time] `db:"UserLoginedAt" json:"LoginedAt" dataType:"datetime"`
	UserLogId int64 `db:"UserLogId" json:"Id" dataType:"bigint" autoKey:"true"`
	UserLogRelatedLogId preformTypes.Null[int64] `db:"UserLogRelatedLogId" json:"RelatedLogId" dataType:"bigint"`
	UserLogType PreformTestAUserLogType `db:"UserLogType" json:"Type" dataType:"enum('Register','login')"`
	UserId int32 `db:"UserId" json:"Id" dataType:"int" autoKey:"true"`
	UserCreatedBy int32 `db:"UserCreatedBy" json:"CreatedBy" dataType:"int"`
	UserLogUserId int32 `db:"UserLogUserId" json:"UserId" dataType:"int"`
}

func (m UserAndLogBody) Factory() *UserAndLogFactory { return UserAndLog }

func (m *UserAndLogBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.UserName
		case 1: return &m.UserCreatedAt
		case 2: return &m.UserLoginedAt
		case 3: return &m.UserLogId
		case 4: return &m.UserLogRelatedLogId
		case 5: return &m.UserLogType
		case 6: return &m.UserId
		case 7: return &m.UserCreatedBy
		case 8: return &m.UserLogUserId
	}
	return nil
}

func (m *UserAndLogBody) FieldValuePtrs() []any { 
	return []any{&m.UserName, &m.UserCreatedAt, &m.UserLoginedAt, &m.UserLogId, &m.UserLogRelatedLogId, &m.UserLogType, &m.UserId, &m.UserCreatedBy, &m.UserLogUserId}
}


