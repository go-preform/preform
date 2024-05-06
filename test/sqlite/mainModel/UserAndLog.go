package mainModel

import (
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/types"
)

var UserAndLog = preform.IniPrebuildQueryFactory[*UserAndLogFactory, UserAndLogBody](func(d *UserAndLogFactory) {
	d.User = d.PreformTestASchema.User.SetAlias("User").(*FactoryUser)
	d.UserLog = d.PreformTestASchema.UserLog.SetAlias("UserLog").(*FactoryUserLog)
	d.SetSrc(d.User).
		Join("Inner", d.UserLog, d.PreformTestASchema.UserLog.UserId.Eq(d.PreformTestASchema.User.Id)).DefineCols(
		preform.SetPrebuildQueryCol(d, d.User.Id.SetAlias("UserId"), d.UserId),
		preform.SetPrebuildQueryCol(d, d.User.Name.SetAlias("UserName"), d.UserName),
		preform.SetPrebuildQueryCol(d, d.User.CreatedBy.SetAlias("UserCreatedBy"), d.UserCreatedBy),
		preform.SetPrebuildQueryCol(d, d.User.CreatedAt.SetAlias("UserCreatedAt"), d.UserCreatedAt),
		preform.SetPrebuildQueryCol(d, d.User.LoginedAt.SetAlias("UserLoginedAt"), d.UserLoginedAt),
		preform.SetPrebuildQueryCol(d, d.User.Detail.SetAlias("UserDetail"), d.UserDetail),
		preform.SetPrebuildQueryCol(d, d.User.Config.SetAlias("UserConfig"), d.UserConfig),
		preform.SetPrebuildQueryCol(d, d.User.ExtraConfig.SetAlias("UserExtraConfig"), d.UserExtraConfig),
		preform.SetPrebuildQueryCol(d, d.UserLog.Id.SetAlias("UserLogId"), d.UserLogId),
		preform.SetPrebuildQueryCol(d, d.UserLog.UserId.SetAlias("UserLogUserId"), d.UserLogUserId),
		preform.SetPrebuildQueryCol(d, d.UserLog.RelatedLogId.SetAlias("UserLogRelatedLogId"), d.UserLogRelatedLogId),
		preform.SetPrebuildQueryCol(d, d.UserLog.Type.SetAlias("UserLogType"), d.UserLogType),
		preform.SetPrebuildQueryCol(d, d.UserLog.Detail.SetAlias("UserLogDetail"), d.UserLogDetail),
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
	UserId *preform.PrebuildQueryCol[int64, preform.NoAggregation]
	UserName *preform.PrebuildQueryCol[string, preform.NoAggregation]
	UserCreatedBy *preform.PrebuildQueryCol[int64, preform.NoAggregation]
	UserCreatedAt *preform.PrebuildQueryCol[preformTypes.SqliteTime, preform.NoAggregation]
	UserLoginedAt *preform.PrebuildQueryCol[preformTypes.Null[preformTypes.SqliteTime], preform.NoAggregation]
	UserDetail *preform.PrebuildQueryCol[preformTypes.Null[preformTypes.JsonRaw[interface {}]], preform.NoAggregation]
	UserConfig *preform.PrebuildQueryCol[preformTypes.JsonRaw[interface {}], preform.NoAggregation]
	UserExtraConfig *preform.PrebuildQueryCol[preformTypes.JsonRaw[interface {}], preform.NoAggregation]
	UserLogId *preform.PrebuildQueryCol[int64, preform.NoAggregation]
	UserLogUserId *preform.PrebuildQueryCol[int64, preform.NoAggregation]
	UserLogRelatedLogId *preform.PrebuildQueryCol[preformTypes.Null[int64], preform.NoAggregation]
	UserLogType *preform.PrebuildQueryCol[int64, preform.NoAggregation]
	UserLogDetail *preform.PrebuildQueryCol[preformTypes.JsonRaw[interface {}], preform.NoAggregation]
}

type UserAndLogBody struct {
	preform.QueryBody[UserAndLogBody, *UserAndLogFactory]
	UserId int64 `db:"UserId" json:"Id" dataType:"INTEGER" autoKey:"true"`
	UserName string `db:"UserName" json:"Name" dataType:"TEXT"`
	UserCreatedBy int64 `db:"UserCreatedBy" json:"CreatedBy" dataType:"INTEGER"`
	UserCreatedAt preformTypes.SqliteTime `db:"UserCreatedAt" json:"CreatedAt" dataType:"datetime" comment:"type:datetime"`
	UserLoginedAt preformTypes.Null[preformTypes.SqliteTime] `db:"UserLoginedAt" json:"LoginedAt" dataType:"datetime" comment:"type:datetime"`
	UserDetail preformTypes.Null[preformTypes.JsonRaw[interface {}]] `db:"UserDetail" json:"Detail" dataType:"jsonb" defaultValue:"NULL"`
	UserConfig preformTypes.JsonRaw[interface {}] `db:"UserConfig" json:"Config" dataType:"jsonb"`
	UserExtraConfig preformTypes.JsonRaw[interface {}] `db:"UserExtraConfig" json:"ExtraConfig" dataType:"jsonb"`
	UserLogId int64 `db:"UserLogId" json:"Id" dataType:"INTEGER" autoKey:"true"`
	UserLogUserId int64 `db:"UserLogUserId" json:"UserId" dataType:"INTEGER"`
	UserLogRelatedLogId preformTypes.Null[int64] `db:"UserLogRelatedLogId" json:"RelatedLogId" dataType:"INTEGER"`
	UserLogType int64 `db:"UserLogType" json:"Type" dataType:"INTEGER"`
	UserLogDetail preformTypes.JsonRaw[interface {}] `db:"UserLogDetail" json:"Detail" dataType:"jsonb"`
}

func (m UserAndLogBody) Factory() *UserAndLogFactory { return UserAndLog }

func (m *UserAndLogBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.UserId
		case 1: return &m.UserName
		case 2: return &m.UserCreatedBy
		case 3: return &m.UserCreatedAt
		case 4: return &m.UserLoginedAt
		case 5: return &m.UserDetail
		case 6: return &m.UserConfig
		case 7: return &m.UserExtraConfig
		case 8: return &m.UserLogId
		case 9: return &m.UserLogUserId
		case 10: return &m.UserLogRelatedLogId
		case 11: return &m.UserLogType
		case 12: return &m.UserLogDetail
	}
	return nil
}

func (m *UserAndLogBody) FieldValuePtrs() []any { 
	return []any{&m.UserId, &m.UserName, &m.UserCreatedBy, &m.UserCreatedAt, &m.UserLoginedAt, &m.UserDetail, &m.UserConfig, &m.UserExtraConfig, &m.UserLogId, &m.UserLogUserId, &m.UserLogRelatedLogId, &m.UserLogType, &m.UserLogDetail}
}


