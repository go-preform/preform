package mainModel

import (
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/types"
	"time"
	SrcTypes "github.com/go-preform/preform/test/pg/mainModel/src/types"
)

var userInit = preform.InitFactory[*FactoryUser, UserBody](func(s *PreformTestASchema) {
	preform.SetColumn(s.User.Id.Column).AutoIncrement()
	s.User.UserByUserUserFk.InitRelation(s.User.CreatedBy, s.User.Id)
	s.User.UsersByUserUserFk.InitRelation(s.User.Id, s.User.CreatedBy)
	s.User.SetTableName("user")
})

type FactoryUser struct {
	preform.Factory[*FactoryUser, UserBody]
	Id *preform.PrimaryKey[int32] `db:"id" json:"Id" dataType:"int4" autoKey:"true"`
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"varchar"`
	ManagerIds *preform.Column[preformTypes.Array[int32]] `db:"manager_ids" json:"ManagerIds" dataType:"_int4"`
	CreatedBy *preform.ForeignKey[int32] `db:"created_by" json:"CreatedBy" dataType:"int4"`
	CreatedAt *preform.Column[time.Time] `db:"created_at" json:"CreatedAt" dataType:"timestamptz"`
	LoginedAt *preform.Column[preformTypes.Null[time.Time]] `db:"logined_at" json:"LoginedAt" dataType:"timestamptz"`
	Detail *preform.Column[preformTypes.Null[preformTypes.JsonRaw[*SrcTypes.UserDetail]]] `db:"detail" json:"Detail" dataType:"jsonb"`
	Config *preform.Column[preformTypes.JsonRaw[*SrcTypes.UserConfig]] `db:"config" json:"Config" dataType:"jsonb" defaultValue:"'{}'::json"`
	ExtraConfig *preform.Column[preformTypes.JsonRaw[SrcTypes.UserConfig]] `db:"extra_config" json:"ExtraConfig" dataType:"jsonb" defaultValue:"'{}'::json"`
	
	//relations
	UserByUserUserFk *preform.ToOne[*UserBody, *FactoryUser, UserBody]
	UsersByUserUserFk *preform.ToMany[*UserBody, *FactoryUser, UserBody]
	UserLogs *preform.ToMany[*UserBody, *FactoryUserLog, UserLogBody]
	UserLogsByUserLogUserFkRegister *preform.ToMany[*UserBody, *FactoryUserLog, UserLogBody]
}

func (f FactoryUser) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryUser, UserBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int32] )
	ff.Name = cols[1].(*preform.Column[string] )
	ff.ManagerIds = cols[2].(*preform.Column[preformTypes.Array[int32]] )
	ff.CreatedBy = cols[3].(*preform.ForeignKey[int32] )
	ff.CreatedAt = cols[4].(*preform.Column[time.Time] )
	ff.LoginedAt = cols[5].(*preform.Column[preformTypes.Null[time.Time]] )
	ff.Detail = cols[6].(*preform.Column[preformTypes.Null[preformTypes.JsonRaw[*SrcTypes.UserDetail]]] )
	ff.Config = cols[7].(*preform.Column[preformTypes.JsonRaw[*SrcTypes.UserConfig]] )
	ff.ExtraConfig = cols[8].(*preform.Column[preformTypes.JsonRaw[SrcTypes.UserConfig]] )
	return ff.Factory.Definition
}


type UserBody struct {
	preform.Body[UserBody,*FactoryUser]
	Id int32 `db:"id" json:"Id" dataType:"int4" autoKey:"true"`
	Name string `db:"name" json:"Name" dataType:"varchar"`
	ManagerIds preformTypes.Array[int32] `db:"manager_ids" json:"ManagerIds" dataType:"_int4"`
	CreatedBy int32 `db:"created_by" json:"CreatedBy" dataType:"int4"`
	CreatedAt time.Time `db:"created_at" json:"CreatedAt" dataType:"timestamptz"`
	LoginedAt preformTypes.Null[time.Time] `db:"logined_at" json:"LoginedAt" dataType:"timestamptz"`
	Detail preformTypes.Null[preformTypes.JsonRaw[*SrcTypes.UserDetail]] `db:"detail" json:"Detail" dataType:"jsonb"`
	Config preformTypes.JsonRaw[*SrcTypes.UserConfig] `db:"config" json:"Config" dataType:"jsonb" defaultValue:"'{}'::json"`
	ExtraConfig preformTypes.JsonRaw[SrcTypes.UserConfig] `db:"extra_config" json:"ExtraConfig" dataType:"jsonb" defaultValue:"'{}'::json"`
	
	UserByUserUserFk *UserBody
	UsersByUserUserFk []*UserBody
	UserLogs []*UserLogBody
	UserLogsByUserLogUserFkRegister []*UserLogBody
}

func (m UserBody) Factory() *FactoryUser { return m.Body.Factory(PreformTestA.User) }

func (m *UserBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.User.Insert(m, cfg...) }

func (m *UserBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.User.UpdateByPk(m, cfg...) }

func (m *UserBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.User.DeleteByPk(m, cfg...) }

func (m UserBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.Name, &m.ManagerIds, &m.CreatedBy, &m.CreatedAt, &m.LoginedAt, &m.Detail, &m.Config, &m.ExtraConfig} }

func (m *UserBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.Name
		case 2: return &m.ManagerIds
		case 3: return &m.CreatedBy
		case 4: return &m.CreatedAt
		case 5: return &m.LoginedAt
		case 6: return &m.Detail
		case 7: return &m.Config
		case 8: return &m.ExtraConfig
	}
	return nil
}

func (m *UserBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.Name, &m.ManagerIds, &m.CreatedBy, &m.CreatedAt, &m.LoginedAt, &m.Detail, &m.Config, &m.ExtraConfig}
}

func (m *UserBody) RelatedValuePtrs() []any { return []any{&m.UserByUserUserFk, &m.UsersByUserUserFk, &m.UserLogs, &m.UserLogsByUserLogUserFkRegister} }


func (m *UserBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.UserByUserUserFk
			case 1: return &m.UsersByUserUserFk
			case 2: return &m.UserLogs
			case 3: return &m.UserLogsByUserLogUserFkRegister
	}
	return nil
}


func (m *UserBody) LoadUserByUserUserFk(noCache ...bool) (*UserBody, error) {
	if m.UserByUserUserFk == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserByUserUserFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByUserUserFk, nil
}

func (m *UserBody) LoadUsersByUserUserFk(noCache ...bool) ([]*UserBody, error) {
	if len(m.UsersByUserUserFk) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UsersByUserUserFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UsersByUserUserFk, nil
}

func (m *UserBody) LoadUserLogs(noCache ...bool) ([]*UserLogBody, error) {
	if len(m.UserLogs) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserLogs.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserLogs, nil
}

func (m *UserBody) LoadUserLogsByUserLogUserFkRegister(noCache ...bool) ([]*UserLogBody, error) {
	if len(m.UserLogsByUserLogUserFkRegister) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserLogsByUserLogUserFkRegister.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserLogsByUserLogUserFkRegister, nil
}

