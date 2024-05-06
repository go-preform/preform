package preform

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"reflect"
)

var (
	errorViewNotWritable = fmt.Errorf("writing view not implemented")
)

func InitViewFactory[FPtr IViewFactory, B any](setter any) func() FPtr {
	return func() FPtr {
		d, factory := initFactoryFields[FPtr, B](setter)
		reflect.ValueOf(d).Elem().Field(0).Set(reflect.ValueOf(*newViewFactory(factory, factory.codeName, factory.tableName, nil)))

		return d
	}
}

type IViewFactory interface {
	IFactory
	IsView()
}

type ViewFactory[FPtr IFactory, B any] struct {
	*Factory[FPtr, B]
}

func newViewFactory[FPtr IFactory, B any](factory *factory[FPtr, B], alias, tableName string, schema ISchema, forceNew ...bool) *ViewFactory[FPtr, B] {
	ff := &ViewFactory[FPtr, B]{
		Factory: newFactory[FPtr, B](factory, alias, tableName, schema, forceNew...),
	}
	return ff
}
func (f ViewFactory[FPtr, B]) SetAlias(alias string) IQuery {
	return newViewFactory(f.factory, alias, f.tableName, f.schema).Definition
}

func (f *ViewFactory[FPtr, B]) SetSchema(schema ISchema) IFactory {
	if f.schema == nil {
		f.setSchema(schema)
		return f
	} else {
		return newViewFactory(f.factory, f.alias, f.tableName, schema).Definition
	}
}

func (f ViewFactory[FPtr, B]) Clone() IFactory {
	return newViewFactory(f.factory, f.alias, f.tableName, f.schema).Definition
}

func (f *ViewFactory[FPtr, B]) SetTableName(name string) *ViewFactory[FPtr, B] {
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
}

func (f *ViewFactory[FPtr, B]) Insert(body any, cfg ...EditConfig) error {
	return errorViewNotWritable
}

func (f *ViewFactory[FPtr, B]) UpdateByPk(body any, cfg ...UpdateConfig) (int64, error) {
	return 0, errorViewNotWritable
}

func (f *ViewFactory[FPtr, B]) DeleteByPk(body any, cfg ...EditConfig) (int64, error) {
	return 0, errorViewNotWritable
}

func (f ViewFactory[FPtr, B]) IsView() {}
