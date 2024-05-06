package mainModel
// Code generated by preform. DO NOT EDIT.
import (
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/share"
)

type PreformTestBSchema struct {
	preform.Schema[*PreformTestBSchema, PreformTestBSchema]
	Bar *FactoryBar
}

func (s *PreformTestBSchema) Factories() []preform.IFactory {
	return []preform.IFactory{s.Bar} 
}

func (s *PreformTestBSchema) Clone(name string, db...*sql.DB) preform.ISchema {
	ss := s.clone(name, db...)
	ss.PrepareFactories([]preform.ISchema{})
	ss.Inherit(s)
	return ss
}

func (s *PreformTestBSchema) clone(name string, db...*sql.DB) preform.ISchema {
	db = append(db, s.Db().DB.DB)
	var queryRunners []preformShare.QueryRunner
	if s.Db().QueryRunner.BaseRunner() != s.Db().DB {
		queryRunners = append(queryRunners, s.Db().QueryRunner)
	}
	if name == "" {
		name = s.Name()
	}
	ss := initPreformTestB(db[0], name, queryRunners...)
	return ss
}

var (
	PreformTestB *PreformTestBSchema
)

func initPreformTestB(conn *sql.DB, name string, queryRunnerForTest ... preformShare.QueryRunner) preform.ISchema {
	s := &PreformTestBSchema{}
	if PreformTestB == nil {
		PreformTestB = s
	}
	s.Bar = barInit()
	if name == "" {
		name = "preform_test_b"
	}
	s.Init(name, s, conn, queryRunnerForTest...)
	return s
}
