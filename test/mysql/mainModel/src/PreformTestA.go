package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
	"time"
	"github.com/go-preform/preform/types"
)

type Enum_PreformTestA_UserLogType string

type PreformTestA_foo struct {
	preformBuilder.FactoryBuilder[*PreformTestA_foo]
	Id	preformBuilder.PrimaryKeyDef[int32] `db:"id" json:"Id" dataType:"int" autoKey:"true"`
	Fk1	preformBuilder.ColumnDef[int32] `db:"fk1" json:"Fk1" dataType:"int"`
	Fk2	preformBuilder.ColumnDef[int32] `db:"fk2" json:"Fk2" dataType:"int"`
}

type PreformTestA_user struct {
	preformBuilder.FactoryBuilder[*PreformTestA_user]
	Id	preformBuilder.PrimaryKeyDef[int32] `db:"id" json:"Id" dataType:"int" autoKey:"true"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"varchar"`
	CreatedBy	preformBuilder.ForeignKeyDef[int32] `db:"created_by" json:"CreatedBy" dataType:"int"`
	CreatedAt	preformBuilder.ColumnDef[time.Time] `db:"created_at" json:"CreatedAt" dataType:"datetime"`
	LoginedAt	preformBuilder.ColumnDef[preformTypes.Null[time.Time]] `db:"logined_at" json:"LoginedAt" dataType:"datetime"`
}

type PreformTestA_userLog struct {
	preformBuilder.FactoryBuilder[*PreformTestA_userLog]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"bigint" autoKey:"true"`
	UserId	preformBuilder.ForeignKeyDef[int32] `db:"user_id" json:"UserId" dataType:"int"`
	RelatedLogId	preformBuilder.ForeignKeyDef[preformTypes.Null[int64]] `db:"related_log_id" json:"RelatedLogId" dataType:"bigint"`
	Type	preformBuilder.ColumnDef[Enum_PreformTestA_UserLogType] `db:"type" json:"Type" dataType:"enum('Register','login')"`
}

type PreformTestA_userManager struct {
	preformBuilder.FactoryBuilder[*PreformTestA_userManager]
	UserId	preformBuilder.PrimaryKeyDef[int32] `db:"user_id" json:"UserId" dataType:"int"`
	ManagerId	preformBuilder.PrimaryKeyDef[int32] `db:"manager_id" json:"ManagerId" dataType:"int"`
}

type PreformTestASchema struct {
	name string
	foo *PreformTestA_foo
	user *PreformTestA_user
	userLog *PreformTestA_userLog
	userManager *PreformTestA_userManager
}

var (
	PreformTestA = PreformTestASchema{name: "PreformTestA"}
)

func initPreformTestA() (string, []preformShare.IFactoryBuilder, *PreformTestASchema, map[string][]string, map[string]*preformShare.CustomType) {

	//implement IFactoryBuilderWithSetup in a new file if you need to customize the factory
	
	PreformTestA.foo = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_foo) {
		d.SetTableName("foo")
	})
	
	PreformTestA.user = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_user) {
		d.SetTableName("user")
		d.Id.SetAssociatedKey(PreformTestA.userManager.UserId, preformBuilder.FkMiddleTable(PreformTestA.userManager, []preformShare.IColDef{PreformTestA.user.Id}, []preformShare.IColDef{PreformTestA.userManager.UserId}, []preformShare.IColDef{PreformTestA.user.Id}, []preformShare.IColDef{PreformTestA.userManager.ManagerId}))
		d.CreatedBy.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("user_FK"))
	})
	
	PreformTestA.userLog = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_userLog) {
		d.SetTableName("user_log")
		d.UserId.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("user_log_FK"))
		d.RelatedLogId.SetAssociatedKey(PreformTestA.userLog.Id, preformBuilder.FkName("user_log_related_FK"))
	})
	
	PreformTestA.userManager = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_userManager) {
		d.SetTableName("user_manager")
	})

	return "preform_test_a",
		[]preformShare.IFactoryBuilder{
			PreformTestA.foo,
			PreformTestA.user,
			PreformTestA.userLog,
			PreformTestA.userManager,
		},
		&PreformTestA,
		map[string][]string{"UserLogType":{"Register","login"}},
        map[string]*preformShare.CustomType{}
}