package preform

import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

const (
	INSERT_CHUNK_SIZE = 1000
)

var (
	modelFactoryByName = map[string]IFactory{}

	modelFactoryByType = map[reflect.Type]IFactory{}
)

type slice[T any] []T

func (s slice[T]) Slice() any {
	return s
}

func (s slice[T]) Set(i int, v any) {
	s[i] = v.(T)
}

type iSlice interface {
	Slice() any
	Set(int, any)
}

type IQuery interface {
	preformShare.IQueryFactory
	CodeName() string
	TableName() string
	tableNameWithParent() string
	fromClause() string
	selectQuery() preformShare.SelectBuilder
	Alias() string
	SetAlias(alias string) IQuery
	Db() *db
	BodyType() reflect.Type
	Prepare(s ...ISchema)
	Columns() []ICol
	ColumnsByName() map[string]ICol
	NewBodyPtr() any
	NewBody() any
	newBodyPtrSlice(int) iSlice
}

type IFactory interface {
	IQuery
	Schema() ISchema
	Pks() []ICol
	PkCondByBody(body iModelBody) squirrel.Eq
	Insert(body any, cfg ...EditConfig) error
	UpdateByPk(body any, cfg ...UpdateConfig) (int64, error)
	DeleteByPk(body any, cfg ...EditConfig) (int64, error)
	setSchema(schema ISchema) (setOk bool)
	SetSchema(schema ISchema) IFactory
	Relations() map[string]IRelation
	SelectAny(cols ...any) *SelectQuery[any]
	Clone() IFactory
	CloneInstance(factory IFactory) IFactory
	//Inherit(dbName, codeName string) IFactory
	factoryPtr() any
	FixedCondition() ICond
	IModelScanner() any
	setModelScanner(any)
}

type factory[FPtr IFactory, B any] struct {
	Definition          FPtr
	columns             []ICol
	columnsByName       map[string]ICol
	relations           []IRelation
	relationsByName     map[string]IRelation
	bodyType            reflect.Type
	tableName, codeName string
	primaryKeys         []ICol
	autoPk              ICol
	setter              any //func(s ISchema)
	fieldsByName        map[string]preformShare.IField
	modelScanner        IModelScanner[B]
}

func (f factory[FPtr, B]) NewBody() any {
	var (
		mm    B
		mmPtr any = &mm
	)
	mmPtr.(iModelBody).setFactory(f.Definition)
	return mm
}

func (f factory[FPtr, B]) NewBodyPtr() any {
	var (
		mm    B
		mmPtr any = &mm
	)
	mmPtr.(iModelBody).setFactory(f.Definition)
	return mmPtr
}

func (f factory[FPtr, B]) newBodyPtrSlice(l int) iSlice {
	var (
		mm = make([]*B, l)
	)
	return slice[*B](mm)
}

func (f factory[FPtr, B]) CloneInstance(factory IFactory) IFactory {
	return f.Definition.CloneInstance(factory)
}

func (f factory[FPtr, B]) Columns() []ICol {
	return f.columns
}

func (f factory[FPtr, B]) FieldsByNameNotSafeForClone() map[string]preformShare.IField {
	return f.fieldsByName
}

func (f factory[FPtr, B]) ColumnsByName() map[string]ICol {
	return f.columnsByName
}

func (f *factory[FPtr, B]) addRelation(r IRelation) uint32 {
	f.relations = append(f.relations, r)
	return uint32(len(f.relations)) - 1
}

func (f factory[FPtr, B]) Relations() map[string]IRelation {
	return f.relationsByName
}

func PrepareFactories(s ISchema) {
	for _, f := range modelFactoryByName {
		f.Prepare(s)
	}
}

func InitFactory[FPtr IFactory, B any](setter any) func() FPtr {
	return func() FPtr {
		d, factory := initFactoryFields[FPtr, B](setter)
		reflect.ValueOf(d).Elem().Field(0).Set(reflect.ValueOf(*newFactory(factory, factory.codeName, factory.tableName, nil)))

		return d
	}
}

func initFactoryFields[FPtr IFactory, B any](setter any) (FPtr, *factory[FPtr, B]) {

	var (
		body       B
		dd         FPtr
		defTypeRef = reflect.TypeOf(dd).Elem()
		defRef     = reflect.New(defTypeRef)
		d          = defRef.Interface().(FPtr)
		factory    = &factory[FPtr, B]{
			Definition:      d,
			codeName:        strings.Replace(defTypeRef.Name(), "Factory", "", 1),
			bodyType:        reflect.TypeOf(body),
			setter:          setter,
			relationsByName: map[string]IRelation{},
			columnsByName:   map[string]ICol{},
			fieldsByName:    map[string]preformShare.IField{},
		}
	)
	factory.tableName = strcase.ToSnake(factory.codeName)
	defRef = defRef.Elem()

	var (
		colDbName string
		colPos    = 0
	)

	for i := 0; i < defTypeRef.NumField(); i++ {
		fieldRef := defTypeRef.Field(i)
		if fieldRef.IsExported() {
			if _, ok := defRef.Field(i).Interface().(ICol); ok {
				c := reflect.New(fieldRef.Type.Elem()).Interface().(ICol)
				if colDbName = fieldRef.Tag.Get("db"); colDbName == "" {
					colDbName = strcase.ToSnake(fieldRef.Name)
				}
				c.initCol(fieldRef, colDbName, d, colPos)
				factory.columns = append(factory.columns, c)
				factory.columnsByName[fieldRef.Name] = c
				factory.fieldsByName[fieldRef.Name] = c
				defRef.Field(i).Set(reflect.ValueOf(c))
				colPos++
			} else if _, ok := defRef.Field(i).Interface().(IRelation); ok {
				c := reflect.New(fieldRef.Type.Elem()).Interface().(IRelation)
				c.prepare(fieldRef.Name, uint32(len(factory.relations)))
				defRef.Field(i).Set(reflect.ValueOf(c))
				factory.relationsByName[fieldRef.Name] = c
				factory.relations = append(factory.relations, c)
			}
		}
	}

	modelFactoryByName[factory.codeName] = d
	modelFactoryByType[defTypeRef] = d
	factory.modelScanner = &modelScanner[B]{bodyCreator: factory.NewModel}
	return d, factory
}

func (f *factory[FPtr, B]) setModelScanner(scanner any) {
	f.SetModelScanner(scanner.(IModelScanner[B]))
}

func (f *factory[FPtr, B]) SetModelScanner(scanner IModelScanner[B]) {
	if scanner != nil {
		f.modelScanner = scanner
	} else {
		f.modelScanner = &modelScanner[B]{bodyCreator: f.NewModel}
	}
}

func (f factory[FPtr, B]) NewModel() B {
	var (
		mm B
	)
	return mm
}

func (f factory[FPtr, B]) Pks() []ICol {
	return f.primaryKeys
}

func (f factory[FPtr, B]) PkCondByValues(values []any) squirrel.Eq {
	var (
		eq = squirrel.Eq{}
	)
	for _, pk := range f.primaryKeys {
		eq[f.Definition.Db().dialect.QuoteIdentifier(pk.DbName())] = values[pk.GetPos()]
	}
	return eq
}

func (f factory[FPtr, B]) PkCondByBody(body iModelBody) squirrel.Eq {
	var (
		eq = squirrel.Eq{}
	)
	for _, pk := range f.primaryKeys {
		eq[f.Definition.Db().dialect.QuoteIdentifier(pk.DbName())] = body.FieldValuePtr(pk.GetPos())
	}
	return eq
}

func (f factory[FPtr, B]) BodyType() reflect.Type {
	return f.bodyType
}

func (f factory[FPtr, B]) CodeName() string {
	return f.codeName
}

func (f factory[FPtr, B]) TableName() string {
	return f.tableName
}

func (f *factory[FPtr, B]) Prepare(s ...ISchema) {
	setterV := reflect.ValueOf(f.setter)
	if setterV.Kind() != reflect.Func {
		panic("factory setter must be a function")
	}
	var (
		inT      reflect.Type
		setterT  = setterV.Type()
		setterIn []reflect.Value
		ss       ISchema
	)
	for i := 0; i < setterT.NumIn(); i++ {
		inT = setterT.In(i)
		for _, ss = range s {
			if inT == reflect.TypeOf(ss) {
				setterIn = append(setterIn, reflect.ValueOf(ss))
				break
			}
		}
	}
	setterV.Call(setterIn)
	var (
		fV           = reflect.ValueOf(f.Definition).Elem()
		fT           = fV.Type()
		relationIT   = reflect.TypeOf((*IRelation)(nil)).Elem()
		isPk, isAuto bool
	)
	for i, l := 0, fT.NumField(); i < l; i++ {
		if fT.Field(i).Type.Implements(relationIT) {
			f.relationsByName[fT.Field(i).Name] = fV.Field(i).Interface().(IRelation)
		}
	}
	for _, col := range f.columns {
		col.setFactory(f.Definition)
		if _, _, isPk, isAuto = col.properties(); isAuto {
			f.autoPk = col
			f.primaryKeys = append(f.primaryKeys, col)
		} else if isPk {
			f.primaryKeys = append(f.primaryKeys, col)
		}
	}
}

// syntax sugar to avoid long call chain
func (f factory[FPtr, B]) Use(fn func(factory FPtr)) {
	fn(f.Definition)
}

type iForceBodyScan interface {
	ForceBodyScan()
}

func SelectByFactory[B any](f IQuery, modelScanner IModelScanner[B], cols ...any) *SelectQuery[B] {
	var (
		db   = f.Db()
		body B
		q    = &SelectQuery[B]{
			db:                       db,
			selectBuilder:            f.selectQuery().PlaceholderFormat(db.sqPlaceholderFormat),
			queryFactory:             f,
			relatedFactoriesForCache: []preformShare.IQueryFactory{f},
			ctx:                      db.ctx,
			scanner:                  modelScanner,
		}
	)
	_, q.forceBodyScan = any(body).(iForceBodyScan)
	if len(cols) > 0 {
		q = q.Columns(cols...)
	}
	return q
}
