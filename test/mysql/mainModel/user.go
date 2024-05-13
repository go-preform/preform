package mainModel

import (
	"github.com/go-preform/preform"
	"time"
	"github.com/go-preform/preform/types"
)

var userInit = preform.InitFactory[*FactoryUser, UserBody](func(s *PreformTestASchema) {
	preform.SetColumn(s.User.Id.Column).AutoIncrement()
	s.User.UserByUserManagerManagerId.InitMtRelation(PreformTestA.UserManager, []preform.IColFromFactory{s.User.Id}, []preform.IColFromFactory{PreformTestA.UserManager.UserId}, []preform.IColFromFactory{PreformTestA.User.Id}, []preform.IColFromFactory{PreformTestA.UserManager.ManagerId})
	s.User.UserByUserManagerUserId.InitMtRelation(PreformTestA.UserManager, []preform.IColFromFactory{PreformTestA.User.Id}, []preform.IColFromFactory{PreformTestA.UserManager.ManagerId}, []preform.IColFromFactory{s.User.Id}, []preform.IColFromFactory{PreformTestA.UserManager.UserId})
	s.User.UserByUserFk.InitRelation(s.User.CreatedBy, s.User.Id)
	s.User.UsersByUserFk.InitRelation(s.User.Id, s.User.CreatedBy)
	s.User.SetTableName("user")
})

type FactoryUser struct {
	preform.Factory[*FactoryUser, UserBody]
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"varchar"`
	CreatedAt *preform.Column[time.Time] `db:"created_at" json:"CreatedAt" dataType:"datetime"`
	LoginedAt *preform.Column[preformTypes.Null[time.Time]] `db:"logined_at" json:"LoginedAt" dataType:"datetime"`
	Id *preform.PrimaryKey[int32] `db:"id" json:"Id" dataType:"int" autoKey:"true"`
	CreatedBy *preform.ForeignKey[int32] `db:"created_by" json:"CreatedBy" dataType:"int"`
	
	//relations
	UserByUserManagerManagerId *preform.MiddleTable[*UserBody, *FactoryUser, UserBody, UserManagerBody]
	UserByUserFk *preform.ToOne[*UserBody, *FactoryUser, UserBody]
	UserByUserManagerUserId *preform.MiddleTable[*UserBody, *FactoryUser, UserBody, UserManagerBody]
	UsersByUserFk *preform.ToMany[*UserBody, *FactoryUser, UserBody]
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
	ff.Name = cols[0].(*preform.Column[string] )
	ff.CreatedAt = cols[1].(*preform.Column[time.Time] )
	ff.LoginedAt = cols[2].(*preform.Column[preformTypes.Null[time.Time]] )
	ff.Id = cols[3].(*preform.PrimaryKey[int32] )
	ff.CreatedBy = cols[4].(*preform.ForeignKey[int32] )
	return ff.Factory.Definition
}


type UserBody struct {
	preform.Body[UserBody,*FactoryUser]
	Name string `db:"name" json:"Name" dataType:"varchar"`
	CreatedAt time.Time `db:"created_at" json:"CreatedAt" dataType:"datetime"`
	LoginedAt preformTypes.Null[time.Time] `db:"logined_at" json:"LoginedAt" dataType:"datetime"`
	Id int32 `db:"id" json:"Id" dataType:"int" autoKey:"true"`
	CreatedBy int32 `db:"created_by" json:"CreatedBy" dataType:"int"`
	
	UserByUserManagerManagerId []*UserBody
	UserByUserFk *UserBody
	UserByUserManagerUserId []*UserBody
	UsersByUserFk []*UserBody
	UserLogs []*UserLogBody
	UserLogsByUserLogUserFkRegister []*UserLogBody
}

func (m UserBody) Factory() *FactoryUser { return m.Body.Factory(PreformTestA.User) }

func (m *UserBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.User.Insert(m, cfg...) }

func (m *UserBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.User.UpdateByPk(m, cfg...) }

func (m *UserBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.User.DeleteByPk(m, cfg...) }

func (m UserBody) FieldValueImmutablePtrs() []any { return []any{&m.Name, &m.CreatedAt, &m.LoginedAt, &m.Id, &m.CreatedBy} }

func (m *UserBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Name
		case 1: return &m.CreatedAt
		case 2: return &m.LoginedAt
		case 3: return &m.Id
		case 4: return &m.CreatedBy
	}
	return nil
}

func (m *UserBody) FieldValuePtrs() []any { 
	return []any{&m.Name, &m.CreatedAt, &m.LoginedAt, &m.Id, &m.CreatedBy}
}

func (m *UserBody) RelatedValuePtrs() []any { return []any{&m.UserByUserManagerManagerId, &m.UserByUserFk, &m.UserByUserManagerUserId, &m.UsersByUserFk, &m.UserLogs, &m.UserLogsByUserLogUserFkRegister} }


func (m *UserBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.UserByUserManagerManagerId
			case 1: return &m.UserByUserFk
			case 2: return &m.UserByUserManagerUserId
			case 3: return &m.UsersByUserFk
			case 4: return &m.UserLogs
			case 5: return &m.UserLogsByUserLogUserFkRegister
	}
	return nil
}


func (m *UserBody) LoadUserByUserManagerManagerId(noCache ...bool) ([]*UserBody, error) {
	if len(m.UserByUserManagerManagerId) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserByUserManagerManagerId.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByUserManagerManagerId, nil
}

func (m *UserBody) LoadUserByUserFk(noCache ...bool) (*UserBody, error) {
	if m.UserByUserFk == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserByUserFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByUserFk, nil
}

func (m *UserBody) LoadUserByUserManagerUserId(noCache ...bool) ([]*UserBody, error) {
	if len(m.UserByUserManagerUserId) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserByUserManagerUserId.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByUserManagerUserId, nil
}

func (m *UserBody) LoadUsersByUserFk(noCache ...bool) ([]*UserBody, error) {
	if len(m.UsersByUserFk) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UsersByUserFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UsersByUserFk, nil
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

