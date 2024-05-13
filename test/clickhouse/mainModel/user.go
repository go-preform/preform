package mainModel

import (
	"github.com/go-preform/preform"
	"time"
)

var userInit = preform.InitFactory[*FactoryUser, UserBody](func(s *PreformTestASchema) {
	s.User.UsersByCommentManagerIds01.InitRelation(s.User.ManagerIds, s.User.Id)
	s.User.UsersByCommentManagerIds02.InitRelation(s.User.Id, s.User.ManagerIds)
	s.User.UserByCommentCreatedBy0.InitRelation(s.User.CreatedBy, s.User.Id)
	s.User.UsersByCommentCreatedBy0.InitRelation(s.User.Id, s.User.CreatedBy)
	s.User.SetTableName("user")
})

type FactoryUser struct {
	preform.Factory[*FactoryUser, UserBody]
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"String"`
	ManagerIds *preform.ForeignKey[[]int32] `db:"manager_ids" json:"ManagerIds" dataType:"Array(Int32)" comment:"fk:preform_test_a.user.id"`
	CreatedAt *preform.PrimaryKey[time.Time] `db:"created_at" json:"CreatedAt" dataType:"DateTime"`
	LoginedAt *preform.Column[*time.Time] `db:"logined_at" json:"LoginedAt" dataType:"Nullable(DateTime)"`
	DetailAge *preform.Column[[]int32] `db:"detail.age" json:"DetailAge" dataType:"Array(Int32)"`
	DetailDateOfBirth *preform.Column[[]time.Time] `db:"detail.date_of_birth" json:"DetailDateOfBirth" dataType:"Array(Date)"`
	Id *preform.PrimaryKey[int32] `db:"id" json:"Id" dataType:"Int32"`
	CreatedBy *preform.ForeignKey[int32] `db:"created_by" json:"CreatedBy" dataType:"Int32" comment:"fk:preform_test_a.user.id"`
	
	//relations
	UsersByCommentManagerIds01 *preform.ToMany[*UserBody, *FactoryUser, UserBody]
	UserByCommentCreatedBy0 *preform.ToOne[*UserBody, *FactoryUser, UserBody]
	UsersByCommentManagerIds02 *preform.ToMany[*UserBody, *FactoryUser, UserBody]
	UsersByCommentCreatedBy0 *preform.ToMany[*UserBody, *FactoryUser, UserBody]
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
	ff.ManagerIds = cols[1].(*preform.ForeignKey[[]int32] )
	ff.CreatedAt = cols[2].(*preform.PrimaryKey[time.Time] )
	ff.LoginedAt = cols[3].(*preform.Column[*time.Time] )
	ff.DetailAge = cols[4].(*preform.Column[[]int32] )
	ff.DetailDateOfBirth = cols[5].(*preform.Column[[]time.Time] )
	ff.Id = cols[6].(*preform.PrimaryKey[int32] )
	ff.CreatedBy = cols[7].(*preform.ForeignKey[int32] )
	return ff.Factory.Definition
}


type UserBody struct {
	preform.Body[UserBody,*FactoryUser]
	Name string `db:"name" json:"Name" dataType:"String"`
	ManagerIds []int32 `db:"manager_ids" json:"ManagerIds" dataType:"Array(Int32)" comment:"fk:preform_test_a.user.id"`
	CreatedAt time.Time `db:"created_at" json:"CreatedAt" dataType:"DateTime"`
	LoginedAt *time.Time `db:"logined_at" json:"LoginedAt" dataType:"Nullable(DateTime)"`
	DetailAge []int32 `db:"detail.age" json:"DetailAge" dataType:"Array(Int32)"`
	DetailDateOfBirth []time.Time `db:"detail.date_of_birth" json:"DetailDateOfBirth" dataType:"Array(Date)"`
	Id int32 `db:"id" json:"Id" dataType:"Int32"`
	CreatedBy int32 `db:"created_by" json:"CreatedBy" dataType:"Int32" comment:"fk:preform_test_a.user.id"`
	
	UsersByCommentManagerIds01 []*UserBody
	UserByCommentCreatedBy0 *UserBody
	UsersByCommentManagerIds02 []*UserBody
	UsersByCommentCreatedBy0 []*UserBody
	UserLogs []*UserLogBody
	UserLogsByUserLogUserFkRegister []*UserLogBody
}

func (m UserBody) Factory() *FactoryUser { return m.Body.Factory(PreformTestA.User) }

func (m *UserBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.User.Insert(m, cfg...) }

func (m *UserBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.User.UpdateByPk(m, cfg...) }

func (m *UserBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.User.DeleteByPk(m, cfg...) }

func (m UserBody) FieldValueImmutablePtrs() []any { return []any{&m.Name, &m.ManagerIds, &m.CreatedAt, &m.LoginedAt, &m.DetailAge, &m.DetailDateOfBirth, &m.Id, &m.CreatedBy} }

func (m *UserBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Name
		case 1: return &m.ManagerIds
		case 2: return &m.CreatedAt
		case 3: return &m.LoginedAt
		case 4: return &m.DetailAge
		case 5: return &m.DetailDateOfBirth
		case 6: return &m.Id
		case 7: return &m.CreatedBy
	}
	return nil
}

func (m *UserBody) FieldValuePtrs() []any { 
	return []any{&m.Name, &m.ManagerIds, &m.CreatedAt, &m.LoginedAt, &m.DetailAge, &m.DetailDateOfBirth, &m.Id, &m.CreatedBy}
}

func (m *UserBody) RelatedValuePtrs() []any { return []any{&m.UsersByCommentManagerIds01, &m.UserByCommentCreatedBy0, &m.UsersByCommentManagerIds02, &m.UsersByCommentCreatedBy0, &m.UserLogs, &m.UserLogsByUserLogUserFkRegister} }


func (m *UserBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.UsersByCommentManagerIds01
			case 1: return &m.UserByCommentCreatedBy0
			case 2: return &m.UsersByCommentManagerIds02
			case 3: return &m.UsersByCommentCreatedBy0
			case 4: return &m.UserLogs
			case 5: return &m.UserLogsByUserLogUserFkRegister
	}
	return nil
}


func (m *UserBody) LoadUsersByCommentManagerIds01(noCache ...bool) ([]*UserBody, error) {
	if len(m.UsersByCommentManagerIds01) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UsersByCommentManagerIds01.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UsersByCommentManagerIds01, nil
}

func (m *UserBody) LoadUserByCommentCreatedBy0(noCache ...bool) (*UserBody, error) {
	if m.UserByCommentCreatedBy0 == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UserByCommentCreatedBy0.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByCommentCreatedBy0, nil
}

func (m *UserBody) LoadUsersByCommentManagerIds02(noCache ...bool) ([]*UserBody, error) {
	if len(m.UsersByCommentManagerIds02) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UsersByCommentManagerIds02.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UsersByCommentManagerIds02, nil
}

func (m *UserBody) LoadUsersByCommentCreatedBy0(noCache ...bool) ([]*UserBody, error) {
	if len(m.UsersByCommentCreatedBy0) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.User.UsersByCommentCreatedBy0.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UsersByCommentCreatedBy0, nil
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

