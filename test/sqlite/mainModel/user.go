package mainModel

import (
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/types"
	SrcTypes "github.com/go-preform/preform/test/sqlite/mainModel/src/types"
)

var userInit = preform.InitFactory[*FactoryUser, UserBody](func(s *PreformTestASchema) {
	preform.SetColumn(s.User.Id.Column).AutoIncrement()
	s.User.UserByFkUserCreatedById.InitRelation(s.User.CreatedBy, s.User.Id)
	s.User.UsersByFkUserCreatedById.InitRelation(s.User.Id, s.User.CreatedBy)
	s.User.SetTableName("user")
})

type FactoryUser struct {
	preform.Factory[*FactoryUser, UserBody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"TEXT"`
	CreatedBy *preform.ForeignKey[int64] `db:"created_by" json:"CreatedBy" dataType:"INTEGER"`
	CreatedAt *preform.Column[preformTypes.SqliteTime] `db:"created_at" json:"CreatedAt" dataType:"datetime" comment:"type:datetime"`
	LoginedAt *preform.Column[preformTypes.Null[preformTypes.SqliteTime]] `db:"logined_at" json:"LoginedAt" dataType:"datetime" comment:"type:datetime"`
	Detail *preform.Column[preformTypes.Null[preformTypes.JsonRaw[*SrcTypes.UserDetail]]] `db:"detail" json:"Detail" dataType:"jsonb" defaultValue:"NULL"`
	Config *preform.Column[preformTypes.JsonRaw[*SrcTypes.UserConfig]] `db:"config" json:"Config" dataType:"jsonb"`
	ExtraConfig *preform.Column[preformTypes.JsonRaw[SrcTypes.UserConfig]] `db:"extra_config" json:"ExtraConfig" dataType:"jsonb"`
	
	//relations
	UserByFkUserCreatedById *preform.ToOne[*UserBody, *FactoryUser, UserBody]
	UsersByFkUserCreatedById *preform.ToMany[*UserBody, *FactoryUser, UserBody]
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
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	ff.Name = cols[1].(*preform.Column[string] )
	ff.CreatedBy = cols[2].(*preform.ForeignKey[int64] )
	ff.CreatedAt = cols[3].(*preform.Column[preformTypes.SqliteTime] )
	ff.LoginedAt = cols[4].(*preform.Column[preformTypes.Null[preformTypes.SqliteTime]] )
	ff.Detail = cols[5].(*preform.Column[preformTypes.Null[preformTypes.JsonRaw[*SrcTypes.UserDetail]]] )
	ff.Config = cols[6].(*preform.Column[preformTypes.JsonRaw[*SrcTypes.UserConfig]] )
	ff.ExtraConfig = cols[7].(*preform.Column[preformTypes.JsonRaw[SrcTypes.UserConfig]] )
	return ff.Factory.Definition
}


type UserBody struct {
	preform.Body[UserBody,*FactoryUser]
	Id int64 `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	Name string `db:"name" json:"Name" dataType:"TEXT"`
	CreatedBy int64 `db:"created_by" json:"CreatedBy" dataType:"INTEGER"`
	CreatedAt preformTypes.SqliteTime `db:"created_at" json:"CreatedAt" dataType:"datetime" comment:"type:datetime"`
	LoginedAt preformTypes.Null[preformTypes.SqliteTime] `db:"logined_at" json:"LoginedAt" dataType:"datetime" comment:"type:datetime"`
	Detail preformTypes.Null[preformTypes.JsonRaw[*SrcTypes.UserDetail]] `db:"detail" json:"Detail" dataType:"jsonb" defaultValue:"NULL"`
	Config preformTypes.JsonRaw[*SrcTypes.UserConfig] `db:"config" json:"Config" dataType:"jsonb"`
	ExtraConfig preformTypes.JsonRaw[SrcTypes.UserConfig] `db:"extra_config" json:"ExtraConfig" dataType:"jsonb"`
	
	UserByFkUserCreatedById *UserBody
	UsersByFkUserCreatedById []*UserBody
	UserLogs []*UserLogBody
	UserLogsByUserLogUserFkRegister []*UserLogBody
}

func (m UserBody) Factory() *FactoryUser { return m.Body.Factory(PreformTestA.User) }

func (m *UserBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.User.Insert(m, cfg...) }

func (m *UserBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.User.UpdateByPk(m, cfg...) }

func (m *UserBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.User.DeleteByPk(m, cfg...) }

func (m UserBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.Name, &m.CreatedBy, &m.CreatedAt, &m.LoginedAt, &m.Detail, &m.Config, &m.ExtraConfig} }

func (m *UserBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.Name
		case 2: return &m.CreatedBy
		case 3: return &m.CreatedAt
		case 4: return &m.LoginedAt
		case 5: return &m.Detail
		case 6: return &m.Config
		case 7: return &m.ExtraConfig
	}
	return nil
}

func (m *UserBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.Name, &m.CreatedBy, &m.CreatedAt, &m.LoginedAt, &m.Detail, &m.Config, &m.ExtraConfig}
}

func (m *UserBody) RelatedValuePtrs() []any { return []any{&m.UserByFkUserCreatedById, &m.UsersByFkUserCreatedById, &m.UserLogs, &m.UserLogsByUserLogUserFkRegister} }


func (m *UserBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.UserByFkUserCreatedById
			case 1: return &m.UsersByFkUserCreatedById
			case 2: return &m.UserLogs
			case 3: return &m.UserLogsByUserLogUserFkRegister
	}
	return nil
}


func (m *UserBody) LoadUserByFkUserCreatedById(noCache ...bool) (*UserBody, error) {
	if m.UserByFkUserCreatedById == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserByFkUserCreatedById.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByFkUserCreatedById, nil
}

func (m *UserBody) LoadUsersByFkUserCreatedById(noCache ...bool) ([]*UserBody, error) {
	if len(m.UsersByFkUserCreatedById) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UsersByFkUserCreatedById.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UsersByFkUserCreatedById, nil
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

