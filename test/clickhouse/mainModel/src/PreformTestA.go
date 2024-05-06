package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
	"time"
	"github.com/satori/go.uuid"
)

type Enum_PreformTestA_UserLogType string

type PreformTestA_user struct {
	preformBuilder.FactoryBuilder[*PreformTestA_user]
	Id	preformBuilder.PrimaryKeyDef[int32] `db:"id" json:"Id" dataType:"Int32"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"String"`
	ManagerIds	preformBuilder.ForeignKeyDef[[]int32] `db:"manager_ids" json:"ManagerIds" dataType:"Array(Int32)" comment:"fk:preform_test_a.user.id"`
	CreatedBy	preformBuilder.ForeignKeyDef[int32] `db:"created_by" json:"CreatedBy" dataType:"Int32" comment:"fk:preform_test_a.user.id"`
	CreatedAt	preformBuilder.PrimaryKeyDef[time.Time] `db:"created_at" json:"CreatedAt" dataType:"DateTime"`
	LoginedAt	preformBuilder.ColumnDef[*time.Time] `db:"logined_at" json:"LoginedAt" dataType:"Nullable(DateTime)"`
	DetailAge	preformBuilder.ColumnDef[[]int32] `db:"detail.age" json:"DetailAge" dataType:"Array(Int32)"`
	DetailDateOfBirth	preformBuilder.ColumnDef[[]time.Time] `db:"detail.date_of_birth" json:"DetailDateOfBirth" dataType:"Array(Date)"`
}

type PreformTestA_userLog struct {
	preformBuilder.FactoryBuilder[*PreformTestA_userLog]
	Id	preformBuilder.PrimaryKeyDef[uuid.UUID] `db:"id" json:"Id" dataType:"UUID"`
	UserId	preformBuilder.PrimaryKeyDef[int32] `db:"user_id" json:"UserId" dataType:"Int32" comment:"fk:preform_test_a.user.id"`
	RelatedLogId	preformBuilder.ForeignKeyDef[uuid.NullUUID] `db:"related_log_id" json:"RelatedLogId" dataType:"Nullable(UUID)" comment:"fk:preform_test_a.user_log.id"`
	Type	preformBuilder.ColumnDef[Enum_PreformTestA_UserLogType] `db:"type" json:"Type" dataType:"Enum8('Register' = 1, 'Login' = 2)"`
	Detail	preformBuilder.ColumnDef[map[string]string] `db:"detail" json:"Detail" dataType:"Map(String, String)"`
}

type PreformTestASchema struct {
	name string
	user *PreformTestA_user
	userLog *PreformTestA_userLog
}

var (
	PreformTestA = PreformTestASchema{name: "PreformTestA"}
)

func initPreformTestA() (string, []preformShare.IFactoryBuilder, *PreformTestASchema, map[string][]string, map[string]*preformShare.CustomType) {

	//implement IFactoryBuilderWithSetup in a new file if you need to customize the factory
	
	PreformTestA.user = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_user) {
		d.SetTableName("user")
		d.ManagerIds.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("comment_manager_ids_0"))
		d.CreatedBy.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("comment_created_by_0"))
	})
	
	PreformTestA.userLog = preformBuilder.InitFactoryBuilder(PreformTestA.name, func(d *PreformTestA_userLog) {
		d.SetTableName("user_log")
		d.UserId.SetAssociatedKey(PreformTestA.user.Id, preformBuilder.FkName("comment_user_id_0"))
		d.RelatedLogId.SetAssociatedKey(PreformTestA.userLog.Id, preformBuilder.FkName("comment_related_log_id_0"))
	})

	return "preform_test_a",
		[]preformShare.IFactoryBuilder{
			PreformTestA.user,
			PreformTestA.userLog,
		},
		&PreformTestA,
		map[string][]string{"UserLogType":{"Register","Login"}},
        map[string]*preformShare.CustomType{}
}