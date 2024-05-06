package preform

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"github.com/iancoleman/strcase"
)

type Factory[FPtr IFactory, B any] struct {
	*factory[FPtr, B]
	alias         string
	schema        ISchema
	parent        string
	fromClauseSql string
	Definition    FPtr
	columns       []ICol
	columnsByName map[string]ICol
	primaryKeys   []ICol
	autoPk        ICol
	tableName     string
	fixCond       ICond
}

func newFactory[FPtr IFactory, B any](factory *factory[FPtr, B], alias, tableName string, schema ISchema, forceNew ...bool) *Factory[FPtr, B] {
	ff := &Factory[FPtr, B]{
		factory:       factory,
		alias:         alias,
		schema:        schema,
		tableName:     tableName,
		columnsByName: map[string]ICol{},
		columns:       make([]ICol, len(factory.columns)),
	}
	if schema != nil || (len(forceNew) != 0 && forceNew[0]) {
		if schema != nil {
			ff.parent = schema.DbName()
		}
		ff.genFromClause()

		var (
			isPk, isAuto bool
		)
		for i, col := range factory.columns {
			ff.columns[i] = col.clone(ff)
			ff.columnsByName[col.Name()] = ff.columns[i]
			if _, _, isPk, isAuto = col.properties(); isPk {
				ff.primaryKeys = append(ff.primaryKeys, ff.columns[i])
				if isAuto {
					ff.autoPk = ff.columns[i]
				}
			}
		}
		ff.Definition = factory.Definition.CloneInstance(ff).(FPtr)
	} else {
		ff.Definition = factory.Definition
		ff.columns = factory.columns
		ff.columnsByName = factory.columnsByName
		ff.primaryKeys = factory.primaryKeys
		ff.autoPk = factory.autoPk
		ff.genFromClause()
	}
	return ff
}

func (f *Factory[FPtr, B]) Prepare(s ...ISchema) {
	if f.setSchema(s[0]) {
		f.factory.Prepare(s...)
		f.columns = f.factory.columns
		f.columnsByName = f.factory.columnsByName
		f.primaryKeys = f.factory.primaryKeys
		f.autoPk = f.factory.autoPk
	}
}

func (f *Factory[FPtr, B]) factoryPtr() any {
	return f
}

//func (f *Factory[FPtr, B]) Inherit(dbName, codeName string) IFactory {
//	new_factory := *f.factory
//	new_factory.codeName = codeName
//	new_factory.tableName = dbName
//	new_factory.inherited = f
//	newF := newFactory(&new_factory, codeName, f.schema, true)
//	return newF.Definition
//}

func (f Factory[FPtr, B]) Alias() string {
	return f.alias
}

func (f Factory[FPtr, B]) SetAlias(alias string) IQuery {
	return newFactory(f.factory, alias, f.tableName, f.schema).Definition
}

func (f Factory[FPtr, B]) tableNameWithParent() string {
	if f.parent != "" {
		return fmt.Sprintf("%s.%s", f.Db().dialect.QuoteIdentifier(f.parent), f.Db().dialect.QuoteIdentifier(f.tableName))
	}
	return f.Db().dialect.QuoteIdentifier(f.tableName)
}

func (f Factory[FPtr, B]) fromClause() string {
	return f.fromClauseSql
}

func (f *Factory[FPtr, B]) genFromClause() {
	var db = f.Db()
	if db == nil {
		return
	}
	if f.parent != "" {
		f.fromClauseSql = fmt.Sprintf("%s.%s AS %s", db.dialect.QuoteIdentifier(f.parent), db.dialect.QuoteIdentifier(f.tableName), db.dialect.QuoteIdentifier(f.alias))
	} else {
		f.fromClauseSql = fmt.Sprintf("%s AS %s", db.dialect.QuoteIdentifier(f.tableName), db.dialect.QuoteIdentifier(f.alias))
	}
}

func (f Factory[FPtr, B]) Columns() []ICol {
	return f.columns[:]
}

func (f Factory[FPtr, B]) ColumnsByName() map[string]ICol {
	var (
		res = map[string]ICol{}
	)
	for k, v := range f.columnsByName {
		res[k] = v
	}
	return res
}

func (f *Factory[FPtr, B]) setSchema(schema ISchema) bool {
	if f.schema != schema {
		if f.schema == nil {
			f.schema = schema
			f.parent = schema.DbName()
			f.genFromClause()
			for _, col := range f.columns {
				col.setFactory(f)
			}
		} else {
			f.schema = schema
			f.parent = schema.DbName()
			f.genFromClause()
		}
		return true
	}
	return false
}

func (f *Factory[FPtr, B]) SetSchema(schema ISchema) IFactory {
	if f.schema == nil {
		f.setSchema(schema)
		return f
	} else {
		return newFactory(f.factory, f.alias, f.tableName, schema).Definition
	}
}

func (f Factory[FPtr, B]) Db() *db {
	if f.schema == nil {
		return nil
	}
	return f.schema.Db()
}

func (f Factory[FPtr, B]) Clone() IFactory {
	return newFactory(f.factory, f.alias, f.tableName, f.schema).Definition
}

func (f Factory[FPtr, B]) Schema() ISchema {
	return f.schema
}

func (f *Factory[FPtr, B]) SetTableName(name string) *Factory[FPtr, B] {
	//if f.inherited != nil {
	//	return f
	//}
	if f.tableName == name {
		return f
	}
	var (
		alias = f.alias
	)
	if alias == strcase.ToCamel(f.tableName) {
		alias = strcase.ToCamel(name)
	}
	f.tableName = name
	f.alias = alias
	f.genFromClause()
	return f
	//var (
	//	alias = f.alias
	//)
	//if alias == strcase.ToCamel(f.tableName) {
	//	alias = strcase.ToCamel(name)
	//}
	//ff := newFactory(f.factory, alias, name, f.schema, true)
	//return ff
}

func (f Factory[FPtr, B]) TableNames() []string {
	return []string{f.tableName}
}

func (f *Factory[FPtr, B]) SetFixedCondition(cond IColConditioner) *Factory[FPtr, B] {
	f.fixCond = cond
	return f
}

func (f Factory[FPtr, B]) FixedCondition() ICond {
	return f.fixCond
}

func (f Factory[FPtr, B]) IModelScanner() any {
	return f.modelScanner
}

func (f *Factory[FPtr, B]) Select(cols ...any) *SelectQuery[B] {
	return SelectByFactory[B](f, f.modelScanner, cols...)
}

func (f *Factory[FPtr, B]) SelectAny(cols ...any) *SelectQuery[any] {
	return SelectByFactory[any](f, f.modelScanner.ToAnyScanner(), cols...)
}

func (f Factory[FPtr, B]) selectQuery() preformShare.SelectBuilder {
	q := f.Db().sqStmtBuilder.SelectFast().From(f.fromClause())
	if f.fixCond != nil {
		q = q.Where(f.fixCond)
	}
	return q
}

func (f Factory[FPtr, B]) GetAll(cond ...ICond) ([]B, error) {
	return f.Select().GetAll(cond...)
}

func (f Factory[FPtr, B]) Count(cond ...ICond) (uint64, error) {
	return f.Select().Where(squirrel.And(cond)).Count()
}

func (f Factory[FPtr, B]) GetOne(pkLookup ...any) (*B, error) {
	return f.Select().GetOne(pkLookup...)
}
