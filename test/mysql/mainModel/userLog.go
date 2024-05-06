package mainModel

import (
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/types"
)

var userLogInit = preform.InitFactory[*FactoryUserLog, UserLogBody](func(s *PreformTestASchema) {
	preform.SetColumn(s.UserLog.Id.Column).AutoIncrement()
	s.UserLog.UserByUserLogUserFk.InitRelation(s.UserLog.UserId, s.User.Id)
	s.User.UserLogs.InitRelation(s.User.Id, s.UserLog.UserId)
	s.UserLog.UserByUserLogUserFkRegister.InitRelation(s.UserLog.UserId, s.User.Id)
	s.User.UserLogsByUserLogUserFkRegister.InitRelation(s.User.Id, s.UserLog.UserId).ExtraCond(s.UserLog.Type.Eq(1))
	s.UserLog.UserLogByUserLogUserLogFk.InitRelation(s.UserLog.RelatedLogId, s.UserLog.Id)
	s.UserLog.UserLogsByUserLogUserLogFk.InitRelation(s.UserLog.Id, s.UserLog.RelatedLogId)
	s.UserLog.SetTableName("user_log")
})

type FactoryUserLog struct {
	preform.Factory[*FactoryUserLog, UserLogBody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"bigint" autoKey:"true"`
	UserId *preform.ForeignKey[int32] `db:"user_id" json:"UserId" dataType:"int"`
	RelatedLogId *preform.ForeignKey[preformTypes.Null[int64]] `db:"related_log_id" json:"RelatedLogId" dataType:"bigint"`
	Type *preform.Column[PreformTestAUserLogType] `db:"type" json:"Type" dataType:"enum('Register','login')"`
	
	//relations
	UserByUserLogUserFk *preform.ToOne[*UserLogBody, *FactoryUser, UserBody]
	UserByUserLogUserFkRegister *preform.ToOne[*UserLogBody, *FactoryUser, UserBody]
	UserLogByUserLogUserLogFk *preform.ToOne[*UserLogBody, *FactoryUserLog, UserLogBody]
	UserLogsByUserLogUserLogFk *preform.ToMany[*UserLogBody, *FactoryUserLog, UserLogBody]
}

func (f FactoryUserLog) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryUserLog, UserLogBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	ff.UserId = cols[1].(*preform.ForeignKey[int32] )
	ff.RelatedLogId = cols[2].(*preform.ForeignKey[preformTypes.Null[int64]] )
	ff.Type = cols[3].(*preform.Column[PreformTestAUserLogType] )
	return ff.Factory.Definition
}


type UserLogBody struct {
	preform.Body[UserLogBody,*FactoryUserLog]
	Id int64 `db:"id" json:"Id" dataType:"bigint" autoKey:"true"`
	UserId int32 `db:"user_id" json:"UserId" dataType:"int"`
	RelatedLogId preformTypes.Null[int64] `db:"related_log_id" json:"RelatedLogId" dataType:"bigint"`
	Type PreformTestAUserLogType `db:"type" json:"Type" dataType:"enum('Register','login')"`
	
	UserByUserLogUserFk *UserBody
	UserByUserLogUserFkRegister *UserBody
	UserLogByUserLogUserLogFk *UserLogBody
	UserLogsByUserLogUserLogFk []*UserLogBody
}

func (m UserLogBody) Factory() *FactoryUserLog { return m.Body.Factory(PreformTestA.UserLog) }

func (m *UserLogBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.UserLog.Insert(m, cfg...) }

func (m *UserLogBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.UserLog.UpdateByPk(m, cfg...) }

func (m *UserLogBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.UserLog.DeleteByPk(m, cfg...) }

func (m UserLogBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.UserId, &m.RelatedLogId, &m.Type} }

func (m *UserLogBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.UserId
		case 2: return &m.RelatedLogId
		case 3: return &m.Type
	}
	return nil
}

func (m *UserLogBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.UserId, &m.RelatedLogId, &m.Type}
}

func (m *UserLogBody) RelatedValuePtrs() []any { return []any{&m.UserByUserLogUserFk, &m.UserByUserLogUserFkRegister, &m.UserLogByUserLogUserLogFk, &m.UserLogsByUserLogUserLogFk} }


func (m *UserLogBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.UserByUserLogUserFk
			case 1: return &m.UserByUserLogUserFkRegister
			case 2: return &m.UserLogByUserLogUserLogFk
			case 3: return &m.UserLogsByUserLogUserLogFk
	}
	return nil
}


func (m *UserLogBody) LoadUserByUserLogUserFk(noCache ...bool) (*UserBody, error) {
	if m.UserByUserLogUserFk == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.UserLog.UserByUserLogUserFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByUserLogUserFk, nil
}

func (m *UserLogBody) LoadUserByUserLogUserFkRegister(noCache ...bool) (*UserBody, error) {
	if m.UserByUserLogUserFkRegister == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.UserLog.UserByUserLogUserFkRegister.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserByUserLogUserFkRegister, nil
}

func (m *UserLogBody) LoadUserLogByUserLogUserLogFk(noCache ...bool) (*UserLogBody, error) {
	if m.UserLogByUserLogUserLogFk == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.UserLog.UserLogByUserLogUserLogFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserLogByUserLogUserLogFk, nil
}

func (m *UserLogBody) LoadUserLogsByUserLogUserLogFk(noCache ...bool) ([]*UserLogBody, error) {
	if len(m.UserLogsByUserLogUserLogFk) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.UserLog.UserLogsByUserLogUserLogFk.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.UserLogsByUserLogUserLogFk, nil
}

