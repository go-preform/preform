package preform

import (
	"context"
	"database/sql"
	preformShare "github.com/go-preform/preform/share"
	"reflect"
)

var (
	initialSchemas = map[reflect.Type][]ISchema{}
)

type Schema[TPtr ISchema, T any] struct {
	DB
	db                  *db
	SchemaName          string
	instance            TPtr
	factoriesByName     map[string]IFactory
	factoriesByBodyType map[reflect.Type]IFactory
}

type ISchema interface {
	Name() string
	Clone(name string, db ...*sql.DB) ISchema
	setName(name string)
	Factories() []IFactory
	FactoriesByName() map[string]IFactory
	Init(schemaName string, instance any, db *sql.DB, queryRunner ...preformShare.QueryRunner)
	PrepareFactories(schemas []ISchema)
	Inherit(s ISchema)
	DbName() string
	setExecer(e DB)
	Db() *db
	SetTracerToDb(tracer ITracer)
}

func (s *Schema[TPtr, T]) Init(schemaName string, instance any, conn *sql.DB, queryRunner ...preformShare.QueryRunner) {
	s.instance = instance.(TPtr)
	s.SetConn(conn, queryRunner...)
	s.SchemaName = schemaName
	s.factoriesByName = make(map[string]IFactory)
	s.factoriesByBodyType = make(map[reflect.Type]IFactory)

}

func (s *Schema[TPtr, T]) PrepareFactories(schemas []ISchema) {
	var (
		factories = s.instance.Factories()
		tT        = reflect.TypeOf(s.instance)
	)
	if len(schemas) == 0 {
		schemas = initialSchemas[tT]
	} else if len(initialSchemas[tT]) == 0 {
		initialSchemas[tT] = schemas
	}
	schemas = append([]ISchema{s.instance}, schemas...)
	for i, f := range factories {
		//todo optimize
		if reflect.ValueOf(f).IsNil() {
			factories = s.instance.Factories()
			f = factories[i]
		}
		f.Prepare(schemas...)
		s.factoriesByName[f.Alias()] = f
		if s.factoriesByBodyType[f.BodyType()] == nil {
			s.factoriesByBodyType[f.BodyType()] = f
		}
	}
}

func (s *Schema[TPtr, T]) Inherit(schema ISchema) {
	if _, ok := schema.(TPtr); ok {
		var (
			factories  = s.instance.Factories()
			oFactories = schema.Factories()
		)
		for i, f := range factories {
			f.setModelScanner(oFactories[i].IModelScanner())
		}
	}
}

func (s Schema[TPtr, T]) FactoriesByName() map[string]IFactory {
	return s.factoriesByName
}

func (s Schema[TPtr, T]) FactoriesByBodyType() map[reflect.Type]IFactory {
	return s.factoriesByBodyType
}

func (s *Schema[TPtr, T]) SetTracerToDb(tracerNilToOff ITracer) {
	s.db.SetTracer(tracerNilToOff)
}

func (s *Schema[TPtr, T]) setExecer(e DB) {
	s.DB = e
}

func (s *Schema[TPtr, T]) SetConn(conn *sql.DB, queryRunner ...preformShare.QueryRunner) {
	s.db = DbFromNative(conn, queryRunner...)
	s.DB = s.db
}

func (s *Schema[TPtr, T]) Db() *db {
	return s.db
}

func (s *Schema[TPtr, T]) setName(name string) {
	s.SchemaName = name
}

func (s Schema[TPtr, T]) Name() string {
	return s.SchemaName
}

func (s *Schema[TPtr, T]) DbName() string {
	return s.SchemaName
}

func (s *Schema[TPtr, T]) Use(fn func(s TPtr)) {
	fn(s.instance)
}

func (s Schema[TPtr, T]) BeginTx(ctx context.Context) (*Tx, error) {
	if ctx == nil {
		ctx = s.db.ctx
	}
	return s.db.BeginTx(ctx)
}

// deprecated, slow
func (s Schema[TPtr, T]) RunTx(fn func(t TPtr) error) error {
	var (
		tPtr = s.instance.Clone(s.SchemaName).(TPtr)
	)
	tx, err := s.db.BeginTx(s.db.ctx)
	if err != nil {
		return err
	}
	tPtr.setExecer(tx)
	err = fn(tPtr)
	if err == nil {
		return tx.Commit()
	}
	_ = tx.Rollback()
	return err
}
