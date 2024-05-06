package preform

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"reflect"
	"strings"
)

type iPrebuildQueryFactory interface {
	IQuery
	Prepare(s ...ISchema)
	addCol(col ICol) int
}

type PrebuildQueryFactory[FPtr iPrebuildQueryFactory, B any] struct {
	bodyType                reflect.Type
	Query                   preformShare.SelectBuilder
	def                     FPtr
	defTypeRef              reflect.Type
	codeName, alias, parent string
	setter                  func(d FPtr)
	allCols                 []ICol
	columnsByName           map[string]ICol
	preDefinedCols          []any
	SrcByAlias              map[string]IFactory
	srcs                    []IFactory
	schemas                 []ISchema
	db                      *db
	modelScanner            IModelScanner[B]
}

func (f PrebuildQueryFactory[FPtr, B]) TableName() string {
	return ""
}
func (f PrebuildQueryFactory[FPtr, B]) TableNames() []string {
	return []string{}
}

func (f PrebuildQueryFactory[FPtr, B]) tableNameWithParent() string {
	return ""
}

func (f PrebuildQueryFactory[FPtr, B]) fromClause() string {
	return ""
}

func (f PrebuildQueryFactory[FPtr, B]) selectQuery() preformShare.SelectBuilder {
	return f.Query
}

func (f PrebuildQueryFactory[FPtr, B]) AddRelation(r IRelation) uint32 {
	return 0
}

func (f PrebuildQueryFactory[FPtr, B]) Columns() []ICol {
	return f.allCols
}

func (f PrebuildQueryFactory[FPtr, B]) ColumnsByName() map[string]ICol {
	return f.columnsByName
}

func (f PrebuildQueryFactory[FPtr, B]) Pks() []ICol {
	return []ICol{}
}

func (f PrebuildQueryFactory[FPtr, B]) PkAndValues(body iModelBody) squirrel.Eq {
	return squirrel.Eq{}
}

func (f PrebuildQueryFactory[FPtr, B]) BodyType() reflect.Type {
	return f.bodyType
}

func (f PrebuildQueryFactory[FPtr, B]) CodeName() string {
	return f.codeName
}

func (f *PrebuildQueryFactory[FPtr, B]) clone() *PrebuildQueryFactory[FPtr, B] {
	var (
		d   = reflect.ValueOf(f.def).Elem()
		dt  = d.Type()
		dd  = reflect.New(dt)
		ddd = dd.Elem()
		ff  = *f
	)
	for i, l := 1, dt.NumField(); i < l; i++ {
		ddd.Field(i).Set(d.Field(i))
	}
	ff.def = dd.Interface().(FPtr)
	ddd.Field(0).Set(reflect.ValueOf(ff))
	return &ff
}

func (f PrebuildQueryFactory[FPtr, B]) Parent() string {
	return f.alias
}

func (f PrebuildQueryFactory[FPtr, B]) Alias() string {
	return f.alias
}

func (f *PrebuildQueryFactory[FPtr, B]) SetAlias(alias string) IQuery {
	ff := f.clone()
	ff.alias = alias
	return ff
}

func (f PrebuildQueryFactory[FPtr, B]) NewBody() any {
	var (
		b B
	)
	return b
}

func (f PrebuildQueryFactory[FPtr, B]) NewBodyPtr() any {
	var (
		b B
	)
	return &b
}
func (f PrebuildQueryFactory[FPtr, B]) newBodyPtrSlice(l int) iSlice {
	var (
		mm = make([]*B, l)
	)
	return slice[*B](mm)
}

var (
	queries = []iPrebuildQueryFactory{}
)

func PrepareQueriesAndRelation(s ...ISchema) {
	for _, ss := range s {
		ss.PrepareFactories(s)
	}
	for _, q := range queries {
		q.Prepare(s...)
	}
}

func IniPrebuildQueryFactory[FPtr iPrebuildQueryFactory, B any](setter func(d FPtr)) FPtr {
	var (
		body       B
		dd         FPtr
		defTypeRef = reflect.TypeOf(dd).Elem()
		defRef     = reflect.New(defTypeRef)
		d          = defRef.Interface().(FPtr)
		de         = defRef.Elem()
		f          = PrebuildQueryFactory[FPtr, B]{
			codeName:      defTypeRef.Name(),
			setter:        setter,
			def:           d,
			defTypeRef:    defTypeRef,
			SrcByAlias:    map[string]IFactory{},
			bodyType:      reflect.TypeOf(body),
			columnsByName: map[string]ICol{},
		}
	)

	for i := 0; i < f.defTypeRef.NumField(); i++ {
		fieldRef := f.defTypeRef.Field(i)
		if fieldRef.IsExported() {
			if _, ok := de.Field(i).Interface().(ICol); ok {
				c := reflect.New(fieldRef.Type.Elem()).Interface().(ICol)
				de.Field(i).Set(reflect.ValueOf(c))
			}
		}
	}
	f.alias = f.codeName
	f.modelScanner = &modelScanner[B]{bodyCreator: f.NewModel}
	reflect.ValueOf(d).Elem().Field(0).Set(reflect.ValueOf(f))
	queries = append(queries, any(d).(iPrebuildQueryFactory))
	return d
}

func (f *PrebuildQueryFactory[FPtr, B]) Db() *db {
	return f.schemas[0].Db()
}

func (f *PrebuildQueryFactory[FPtr, B]) NewModel() B {
	var (
		mm B
	)
	return mm
}

func (f *PrebuildQueryFactory[FPtr, B]) Prepare(schemas ...ISchema) {
	var (
		defRef   = reflect.ValueOf(f.def).Elem()
		iSchemaT = reflect.TypeOf((*ISchema)(nil)).Elem()
	)

	for i := 0; i < f.defTypeRef.NumField(); i++ {
		fieldRef := f.defTypeRef.Field(i)
		if !fieldRef.IsExported() {
			continue
		}
		if fieldRef.Type.Kind() == reflect.Ptr {
			//table src will set in setter
			if fieldRef.Type.Implements(iSchemaT) {
				for _, s := range schemas {
					if reflect.TypeOf(s) == fieldRef.Type {
						defRef.Field(i).Set(reflect.ValueOf(s))
						f.schemas = append(f.schemas, s)
						break
					}
				}
			}
		}
	}

	f.Query = squirrel.SelectFast()
	f.setter(f.def)
	for _, col := range f.allCols {
		f.columnsByName[col.DbName()] = col
	}
}

func (f *PrebuildQueryFactory[FPtr, B]) SetSrc(factory IFactory) *PrebuildQueryFactory[FPtr, B] {
	f.Query = f.Query.From(factory.fromClause())
	return f
}

func (f *PrebuildQueryFactory[FPtr, B]) Join(join string, factory IFactory, cond ...ICond) *PrebuildQueryFactory[FPtr, B] {
	joinSql := factory.fromClause()
	var (
		condSql  string
		tmpArgs  []any
		condArgs []any
	)
	if l := len(cond); l != 0 {
		if l > 1 {
			cond[0] = any(cond).(squirrel.And)
		}
		condSql, tmpArgs, _ = cond[0].ToSql()
		joinSql += " ON " + condSql
		for _, arg := range tmpArgs {
			if _, ok := arg.(ICond); ok {
				sql, subArgs, _ := arg.(ICond).ToSql()
				joinSql = strings.Replace(joinSql, "?", sql, 1)
				condArgs = append(condArgs, subArgs...)
			} else {
				condArgs = append(condArgs, arg)
			}
		}
	}
	switch join {
	case "Left":
		f.Query = f.Query.LeftJoin(joinSql, condArgs...)
	case "Right":
		f.Query = f.Query.RightJoin(joinSql, condArgs...)
	case "Inner":
		f.Query = f.Query.Join(joinSql, condArgs...)
	case "CrossJoin":
		f.Query = f.Query.CrossJoin(joinSql, condArgs...)
	}
	return f
}

func (f *PrebuildQueryFactory[FPtr, B]) DefineCols(cols ...any) *PrebuildQueryFactory[FPtr, B] {
	f.preDefinedCols = cols
	return f
}

func (f *PrebuildQueryFactory[FPtr, B]) addCol(col ICol) int {
	f.allCols = append(f.allCols, col)
	return len(f.allCols) - 1
}

func (f *PrebuildQueryFactory[FPtr, B]) Select(overwrittenCols ...any) *SelectQuery[B] {
	return SelectByFactory[B](f, f.modelScanner, overwrittenCols...)
}

func (f *PrebuildQueryFactory[FPtr, B]) PreSetWhere(cond any, extra ...any) *PrebuildQueryFactory[FPtr, B] {
	f.Query = f.Query.Where(cond, extra...)
	return f
}

// having
func (f *PrebuildQueryFactory[FPtr, B]) Having(cond any, extra ...any) *PrebuildQueryFactory[FPtr, B] {
	f.Query = f.Query.Having(cond, extra...)
	return f
}

// group by
func (f *PrebuildQueryFactory[FPtr, B]) GroupBy(group ...any) *PrebuildQueryFactory[FPtr, B] {
	var (
		groupStrs = make([]string, len(group))
		col       ICol
		ok        bool
		s         squirrel.Sqlizer
		args      []any
	)
	for i, g := range group {
		if col, ok = g.(ICol); ok {
			groupStrs[i] = col.GetCode()
		} else if s, ok = g.(squirrel.Sqlizer); ok {
			groupStrs[i], args, _ = s.ToSql()
			if len(args) != 0 {
				groupStrs[i], args, _ = preformShare.NestSql(groupStrs[i], args, f.Db().dialect)
				for _, arg := range args {
					if _, ok = arg.(string); ok {
						groupStrs[i] = strings.Replace(groupStrs[i], "?", fmt.Sprintf(`"%s"`, arg.(string)), 1)
					} else {
						groupStrs[i] = strings.Replace(groupStrs[i], "?", fmt.Sprintf(`%v`, arg), 1)
					}
				}
			}
		}
	}
	f.Query = f.Query.GroupBy(groupStrs...)
	return f
}

func (f *PrebuildQueryFactory[FPtr, B]) SetModelScanner(scanner IModelScanner[B]) {
	if scanner != nil {
		f.modelScanner = scanner
	} else {
		f.modelScanner = &modelScanner[B]{bodyCreator: f.NewModel}
	}
}

// order by
func (f *PrebuildQueryFactory[FPtr, B]) OrderBy(order ...string) *PrebuildQueryFactory[FPtr, B] {
	f.Query = f.Query.OrderBy(order...)
	return f
}

// deprecated
func (f *PrebuildQueryFactory[FPtr, B]) Insert(body any, cfg ...EditConfig) error {
	return nil
}
