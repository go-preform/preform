package mainModel

import (
	"github.com/go-preform/preform"
)

var userManagerInit = preform.InitFactory[*FactoryUserManager, UserManagerBody](func(s *PreformTestASchema) {
	s.UserManager.SetTableName("user_manager")
})

type FactoryUserManager struct {
	preform.Factory[*FactoryUserManager, UserManagerBody]
	UserId *preform.PrimaryKey[int32] `db:"user_id" json:"UserId" dataType:"int"`
	ManagerId *preform.PrimaryKey[int32] `db:"manager_id" json:"ManagerId" dataType:"int"`
}

func (f FactoryUserManager) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryUserManager, UserManagerBody])
	ff.Factory.Definition = &ff
	ff.UserId = cols[0].(*preform.PrimaryKey[int32] )
	ff.ManagerId = cols[1].(*preform.PrimaryKey[int32] )
	return ff.Factory.Definition
}


type UserManagerBody struct {
	preform.Body[UserManagerBody,*FactoryUserManager]
	UserId int32 `db:"user_id" json:"UserId" dataType:"int"`
	ManagerId int32 `db:"manager_id" json:"ManagerId" dataType:"int"`
}

func (m UserManagerBody) Factory() *FactoryUserManager { return m.Body.Factory(PreformTestA.UserManager) }

func (m *UserManagerBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.UserManager.Insert(m, cfg...) }

func (m *UserManagerBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.UserManager.UpdateByPk(m, cfg...) }

func (m *UserManagerBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.UserManager.DeleteByPk(m, cfg...) }

func (m UserManagerBody) FieldValueImmutablePtrs() []any { return []any{&m.UserId, &m.ManagerId} }

func (m *UserManagerBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.UserId
		case 1: return &m.ManagerId
	}
	return nil
}

func (m *UserManagerBody) FieldValuePtrs() []any { 
	return []any{&m.UserId, &m.ManagerId}
}

func (m *UserManagerBody) RelatedValuePtrs() []any { return []any{} }


func (m *UserManagerBody) RelatedByPos(pos uint32, toSet ...any) bool {
	return false
}




