package main

import (
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/test/sqlite/mainModel/src/types"
	preformTypes "github.com/go-preform/preform/types"
)

func init() {
	PrebuildQueries = append(PrebuildQueries, preformBuilder.BuildQuery("user_and_log", func(builder *preformBuilder.QueryBuilder, pta *PreformTestASchema) *preformBuilder.QueryBuilder {
		builder.From(pta.user).InnerJoinByForeignKey(pta.userLog.UserId).Where(pta.userLog.UserId.NotEq(2))
		return builder
	}))
}

func (p *PreformTestA_user) Setup() (skipAutoSetter bool) {
	p.Detail.OverwriteType(preformBuilder.ColumnDef[preformTypes.Null[preformTypes.JsonRaw[*types.UserDetail]]]{})
	p.Config.OverwriteType(preformBuilder.ColumnDef[preformTypes.JsonRaw[*types.UserConfig]]{})
	p.ExtraConfig.OverwriteType(preformBuilder.ColumnDef[preformTypes.JsonRaw[types.UserConfig]]{})
	return false
}

func (p *PreformTestA_userLog) Setup() (skipAutoSetter bool) {
	p.SetTableName("user_log")
	p.Id.RelatedFk(&PreformTestA.userLog.RelatedLogId)
	p.UserId.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("user_log_user_fk"), preformBuilder.FkReverseName("UserLogs"))
	p.UserId.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkCond(nil, p.Type.Eq(1)), preformBuilder.FkName("user_log_user_fk_register"))
	p.RelatedLogId.SetAssociatedKey(PreformTestA.userLog.Id, preformBuilder.FkName("user_log_user_log_fk"))
	return true
}
