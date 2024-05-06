package mainModel
// Code generated by preform. DO NOT EDIT.
import (
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/share"
)

type PreformTestASchema struct {
	preform.Schema[*PreformTestASchema, PreformTestASchema]
	User *FactoryUser
	UserLog *FactoryUserLog
}

func (s *PreformTestASchema) Factories() []preform.IFactory {
	return []preform.IFactory{s.User, s.UserLog} 
}

func (s *PreformTestASchema) Clone(name string, db...*sql.DB) preform.ISchema {
	ss := s.clone(name, db...)
	ss.PrepareFactories([]preform.ISchema{})
	ss.Inherit(s)
	return ss
}

func (s *PreformTestASchema) clone(name string, db...*sql.DB) preform.ISchema {
	db = append(db, s.Db().DB.DB)
	var queryRunners []preformShare.QueryRunner
	if s.Db().QueryRunner.BaseRunner() != s.Db().DB {
		queryRunners = append(queryRunners, s.Db().QueryRunner)
	}
	if name == "" {
		name = s.Name()
	}
	ss := initPreformTestA(db[0], name, queryRunners...)
	return ss
}

var (
	PreformTestA *PreformTestASchema
)

func initPreformTestA(conn *sql.DB, name string, queryRunnerForTest ... preformShare.QueryRunner) preform.ISchema {
	s := &PreformTestASchema{}
	if PreformTestA == nil {
		PreformTestA = s
	}
	s.User = userInit()
	s.UserLog = userLogInit()
	if name == "" {
		name = "preform_test_a"
	}
	s.Init(name, s, conn, queryRunnerForTest...)
	return s
}
