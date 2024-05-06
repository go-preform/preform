package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
	"github.com/go-preform/preform/types"
	"time"
	"github.com/satori/go.uuid"
)

type Enum_PreformTestA_LogType string
type CustomType_PreformTestA_LogDetail struct{}

type PreformTestA_user struct {
	preformBuilder.FactoryBuilder[*PreformTestA_user]
	Id	preformBuilder.PrimaryKeyDef[int32] `db:"id" json:"Id" dataType:"int4" autoKey:"true"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"varchar"`
	ManagerIds	preformBuilder.ColumnDef[preformTypes.Array[int32]] `db:"manager_ids" json:"ManagerIds" dataType:"_int4"`
	CreatedBy	preformBuilder.ForeignKeyDef[int32] `db:"created_by" json:"CreatedBy" dataType:"int4"`
	CreatedAt	preformBuilder.ColumnDef[time.Time] `db:"created_at" json:"CreatedAt" dataType:"timestamptz"`
	LoginedAt	preformBuilder.ColumnDef[preformTypes.Null[time.Time]] `db:"logined_at" json:"LoginedAt" dataType:"timestamptz"`
	Detail	preformBuilder.ColumnDef[preformTypes.Null[preformTypes.JsonRaw[any]]] `db:"detail" json:"Detail" dataType:"jsonb"`
	Config	preformBuilder.ColumnDef[preformTypes.JsonRaw[any]] `db:"config" json:"Config" dataType:"jsonb" defaultValue:"'{}'::json"`
	ExtraConfig	preformBuilder.ColumnDef[preformTypes.JsonRaw[any]] `db:"extra_config" json:"ExtraConfig" dataType:"jsonb" defaultValue:"'{}'::json"`
}

type PreformTestA_userLog struct {
	preformBuilder.FactoryBuilder[*PreformTestA_userLog]
	Id	preformBuilder.PrimaryKeyDef[uuid.UUID] `db:"id" json:"Id" dataType:"uuid"`
	UserId	preformBuilder.ForeignKeyDef[int32] `db:"user_id" json:"UserId" dataType:"int4"`
	RelatedLogId	preformBuilder.ForeignKeyDef[preformTypes.Null[uuid.UUID]] `db:"related_log_id" json:"RelatedLogId" dataType:"uuid"`
	Type	preformBuilder.ColumnDef[Enum_PreformTestA_LogType] `db:"type" json:"Type" dataType:"log_type"`
	Detail	preformBuilder.ColumnDef[CustomType_PreformTestA_LogDetail] `db:"detail" json:"Detail" dataType:"log_detail"`
}

type PreformTestA_foo struct {
	preformBuilder.FactoryBuilder[*PreformTestA_foo]
	Id	preformBuilder.PrimaryKeyDef[int32] `db:"id" json:"Id" dataType:"int4" autoKey:"true"`
	Fk1	preformBuilder.ColumnDef[int32] `db:"fk1" json:"Fk1" dataType:"int4"`
	Fk2	preformBuilder.ColumnDef[int32] `db:"fk2" json:"Fk2" dataType:"int4"`
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
		d.Id.RelatedFk(&PreformTestA.userLog.UserId)
		d.CreatedBy.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("user_user_fk"))
	})
	
	PreformTestA.userLog = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_userLog) {
		d.SetTableName("user_log")
		d.UserId.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("user_log_user_fk"))
		d.RelatedLogId.SetAssociatedKey(PreformTestA.userLog.Id, preformBuilder.FkName("user_log_user_log_fk"))
	})
	
	PreformTestA.foo = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_foo) {
		d.SetTableName("foo")
		d.Fk1.RelatedFk(&PreformTestB.bar.Id1)
		d.Fk2.RelatedFk(&PreformTestB.bar.Id2)
	})

	return "preform_test_a",
		[]preformShare.IFactoryBuilder{
			PreformTestA.user,
			PreformTestA.userLog,
			PreformTestA.foo,
		},
		&PreformTestA,
		map[string][]string{"log_type":{"Register","Login"}},
        map[string]*preformShare.CustomType{"log_detail":{Name:"log_detail",Attr: []*preformShare.CustomTypeAttr{{Name:"user_agent",Type:"string",NotNull:true,IsScanner:false},{Name:"session_id",Type:"uuid.UUID",NotNull:true,IsScanner:true},{Name:"last_login",Type:"time.Time",NotNull:true,IsScanner:false}},Imports: map[string]struct{}{"\"github.com/satori/go.uuid\"":{},"\"time\"":{}}}}
}