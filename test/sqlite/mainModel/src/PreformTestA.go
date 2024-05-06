package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/types"
)



type PreformTestA_user struct {
	preformBuilder.FactoryBuilder[*PreformTestA_user]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"TEXT"`
	CreatedBy	preformBuilder.ForeignKeyDef[int64] `db:"created_by" json:"CreatedBy" dataType:"INTEGER"`
	CreatedAt	preformBuilder.ColumnDef[preformTypes.SqliteTime] `db:"created_at" json:"CreatedAt" dataType:"datetime" comment:"type:datetime"`
	LoginedAt	preformBuilder.ColumnDef[preformTypes.Null[preformTypes.SqliteTime]] `db:"logined_at" json:"LoginedAt" dataType:"datetime" comment:"type:datetime"`
	Detail	preformBuilder.ColumnDef[preformTypes.Null[preformTypes.JsonRaw[any]]] `db:"detail" json:"Detail" dataType:"jsonb" defaultValue:"NULL"`
	Config	preformBuilder.ColumnDef[preformTypes.JsonRaw[any]] `db:"config" json:"Config" dataType:"jsonb"`
	ExtraConfig	preformBuilder.ColumnDef[preformTypes.JsonRaw[any]] `db:"extra_config" json:"ExtraConfig" dataType:"jsonb"`
}

type PreformTestA_userLog struct {
	preformBuilder.FactoryBuilder[*PreformTestA_userLog]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	UserId	preformBuilder.ForeignKeyDef[int64] `db:"user_id" json:"UserId" dataType:"INTEGER"`
	RelatedLogId	preformBuilder.ForeignKeyDef[preformTypes.Null[int64]] `db:"related_log_id" json:"RelatedLogId" dataType:"INTEGER"`
	Type	preformBuilder.ColumnDef[int64] `db:"type" json:"Type" dataType:"INTEGER"`
	Detail	preformBuilder.ColumnDef[preformTypes.JsonRaw[any]] `db:"detail" json:"Detail" dataType:"jsonb"`
}

type PreformTestA_foo struct {
	preformBuilder.FactoryBuilder[*PreformTestA_foo]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	Fk	preformBuilder.ForeignKeyDef[int64] `db:"fk" json:"Fk" dataType:"INTEGER" comment:"fk:preform_test_b.bar.id"`
}

type PreformTestASchema struct {
	name string
	user *PreformTestA_user
	userLog *PreformTestA_userLog
	foo *PreformTestA_foo
}

var (
	PreformTestA = PreformTestASchema{name: "PreformTestA"}
)

func initPreformTestA() (string, []preformShare.IFactoryBuilder, *PreformTestASchema, map[string][]string, map[string]*preformShare.CustomType) {

	//implement IFactoryBuilderWithSetup in a new file if you need to customize the factory
	
	PreformTestA.user = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_user) {
		d.SetTableName("user")
		d.CreatedBy.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("fk_user_created_by_id"))
	})
	
	PreformTestA.userLog = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_userLog) {
		d.SetTableName("user_log")
		d.UserId.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("fk_user_log_user_id_id"))
		d.RelatedLogId.SetAssociatedKey(PreformTestA.userLog.Id, preformBuilder.FkName("fk_user_log_related_log_id_id"))
	})
	
	PreformTestA.foo = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_foo) {
		d.SetTableName("foo")
		d.Fk.SetAssociatedKey(PreformTestB.bar.Id, preformBuilder.FkName("comment_fk_0"))
	})

	return "preform_test_a",
		[]preformShare.IFactoryBuilder{
			PreformTestA.user,
			PreformTestA.userLog,
			PreformTestA.foo,
		},
		&PreformTestA,
		map[string][]string{},
        map[string]*preformShare.CustomType{}
}