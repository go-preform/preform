package mainModel

import (
	"github.com/go-preform/preform"
	"github.com/satori/go.uuid"
	"github.com/go-preform/preform/types"
	"time"
)

var UserAndLog = preform.IniPrebuildQueryFactory[*UserAndLogFactory, UserAndLogBody](func(d *UserAndLogFactory) {
	d.User = d.PreformTestASchema.User.SetAlias("User").(*FactoryUser)
	d.UserLog = d.PreformTestASchema.UserLog.SetAlias("UserLog").(*FactoryUserLog)
	d.SetSrc(d.User).
		Join("Inner", d.UserLog, d.PreformTestASchema.UserLog.UserId.Eq(d.PreformTestASchema.User.Id)).DefineCols(
		preform.SetPrebuildQueryCol(d, d.UserLog.Id.SetAlias("UserLogId"), d.UserLogId),
		preform.SetPrebuildQueryCol(d, d.UserLog.RelatedLogId.SetAlias("UserLogRelatedLogId"), d.UserLogRelatedLogId),
		preform.SetPrebuildQueryCol(d, d.User.LoginedAt.SetAlias("UserLoginedAt"), d.UserLoginedAt),
		preform.SetPrebuildQueryCol(d, d.User.Name.SetAlias("UserName"), d.UserName),
		preform.SetPrebuildQueryCol(d, d.User.Detail.SetAlias("UserDetail"), d.UserDetail),
		preform.SetPrebuildQueryCol(d, d.User.Config.SetAlias("UserConfig"), d.UserConfig),
		preform.SetPrebuildQueryCol(d, d.User.ExtraConfig.SetAlias("UserExtraConfig"), d.UserExtraConfig),
		preform.SetPrebuildQueryCol(d, d.User.CreatedAt.SetAlias("UserCreatedAt"), d.UserCreatedAt),
		preform.SetPrebuildQueryCol(d, d.User.ManagerIds.SetAlias("UserManagerIds"), d.UserManagerIds),
		preform.SetPrebuildQueryCol(d, d.UserLog.Type.SetAlias("UserLogType"), d.UserLogType),
		preform.SetPrebuildQueryCol(d, d.UserLog.Detail.SetAlias("UserLogDetail"), d.UserLogDetail),
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
	UserLogId *preform.PrebuildQueryCol[uuid.UUID, preform.NoAggregation]
	UserLogRelatedLogId *preform.PrebuildQueryCol[preformTypes.Null[uuid.UUID], preform.NoAggregation]
	UserLoginedAt *preform.PrebuildQueryCol[preformTypes.Null[time.Time], preform.NoAggregation]
	UserName *preform.PrebuildQueryCol[string, preform.NoAggregation]
	UserDetail *preform.PrebuildQueryCol[preformTypes.Null[preformTypes.JsonRaw[interface {}]], preform.NoAggregation]
	UserConfig *preform.PrebuildQueryCol[preformTypes.JsonRaw[interface {}], preform.NoAggregation]
	UserExtraConfig *preform.PrebuildQueryCol[preformTypes.JsonRaw[interface {}], preform.NoAggregation]
	UserCreatedAt *preform.PrebuildQueryCol[time.Time, preform.NoAggregation]
	UserManagerIds *preform.PrebuildQueryCol[preformTypes.Array[int32], preform.NoAggregation]
	UserLogType *preform.PrebuildQueryCol[PreformTestALogType, preform.NoAggregation]
	UserLogDetail *preform.PrebuildQueryCol[PreformTestALogDetail, preform.NoAggregation]
	UserId *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserCreatedBy *preform.PrebuildQueryCol[int32, preform.NoAggregation]
	UserLogUserId *preform.PrebuildQueryCol[int32, preform.NoAggregation]
}

type UserAndLogBody struct {
	preform.QueryBody[UserAndLogBody, *UserAndLogFactory]
	UserLogId uuid.UUID `db:"UserLogId" json:"Id" dataType:"uuid"`
	UserLogRelatedLogId preformTypes.Null[uuid.UUID] `db:"UserLogRelatedLogId" json:"RelatedLogId" dataType:"uuid"`
	UserLoginedAt preformTypes.Null[time.Time] `db:"UserLoginedAt" json:"LoginedAt" dataType:"timestamptz"`
	UserName string `db:"UserName" json:"Name" dataType:"varchar"`
	UserDetail preformTypes.Null[preformTypes.JsonRaw[interface {}]] `db:"UserDetail" json:"Detail" dataType:"jsonb"`
	UserConfig preformTypes.JsonRaw[interface {}] `db:"UserConfig" json:"Config" dataType:"jsonb" defaultValue:"'{}'::json"`
	UserExtraConfig preformTypes.JsonRaw[interface {}] `db:"UserExtraConfig" json:"ExtraConfig" dataType:"jsonb" defaultValue:"'{}'::json"`
	UserCreatedAt time.Time `db:"UserCreatedAt" json:"CreatedAt" dataType:"timestamptz"`
	UserManagerIds preformTypes.Array[int32] `db:"UserManagerIds" json:"ManagerIds" dataType:"_int4"`
	UserLogType PreformTestALogType `db:"UserLogType" json:"Type" dataType:"log_type"`
	UserLogDetail PreformTestALogDetail `db:"UserLogDetail" json:"Detail" dataType:"log_detail"`
	UserId int32 `db:"UserId" json:"Id" dataType:"int4" autoKey:"true"`
	UserCreatedBy int32 `db:"UserCreatedBy" json:"CreatedBy" dataType:"int4"`
	UserLogUserId int32 `db:"UserLogUserId" json:"UserId" dataType:"int4"`
}

func (m UserAndLogBody) Factory() *UserAndLogFactory { return UserAndLog }

func (m *UserAndLogBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.UserLogId
		case 1: return &m.UserLogRelatedLogId
		case 2: return &m.UserLoginedAt
		case 3: return &m.UserName
		case 4: return &m.UserDetail
		case 5: return &m.UserConfig
		case 6: return &m.UserExtraConfig
		case 7: return &m.UserCreatedAt
		case 8: return &m.UserManagerIds
		case 9: return &m.UserLogType
		case 10: return &m.UserLogDetail
		case 11: return &m.UserId
		case 12: return &m.UserCreatedBy
		case 13: return &m.UserLogUserId
	}
	return nil
}

func (m *UserAndLogBody) FieldValuePtrs() []any { 
	return []any{&m.UserLogId, &m.UserLogRelatedLogId, &m.UserLoginedAt, &m.UserName, &m.UserDetail, &m.UserConfig, &m.UserExtraConfig, &m.UserCreatedAt, &m.UserManagerIds, &m.UserLogType, &m.UserLogDetail, &m.UserId, &m.UserCreatedBy, &m.UserLogUserId}
}


