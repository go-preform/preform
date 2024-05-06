package preform

import (
	preformShare "github.com/go-preform/preform/share"
	"reflect"
)

type IColWrap interface {
	preformShare.ICol
	iTypedCol
	initCol(ref reflect.StructField, dbName string, factory IFactory, pos int)
	setFactory(factory IFactory)
	properties() (isArray bool, isPtr bool, isPk bool, isAuto bool)
	SetAlias(alias string) ICol
	QueryFactory() IQuery
	unwrapPtr(any) any
	unwrapPtrForInsert(any) any //care default value
	unwrapPtrForUpdate(any) any
	getValueFromBody(body iModelBody) any
	getValueFromBodyFlatten(body iModelBody) []any
	getValueFromBodiesAndLess(iModelBody, iModelBody) bool
	valueToString(value any) string
	setValueToBody(body iModelBody, value any)
	wrapScanner(any) any
	flatten(any) []any

	Asc() string
	Desc() string

	Aggregate(fn preformShare.Aggregator, params ...any) iAggregateCol
	Sum() iAggregateCol
	Avg() iAggregateCol
	Max() iAggregateCol
	Min() iAggregateCol
	Count() iAggregateCol
	Mean() iAggregateCol
	Median() iAggregateCol
	Mode() iAggregateCol
	StdDev() iAggregateCol
	CountDistinct() iAggregateCol
	JsonAgg() iAggregateCol
	ArrayAgg() iAggregateCol
	GroupConcat(splitter string) iAggregateCol
	IsSame(col ICol) bool
}

type ColumnWrap[C ICol] struct {
	colConditioner
	IColWrap
	col           C
	alias         string
	altFactory    IFactory
	codeWithAlias string
	code          string
}

func newColWrap[C ICol](col C, altFactory IFactory, alias string) *ColumnWrap[C] {
	cc := &ColumnWrap[C]{IColWrap: col, col: col, altFactory: altFactory, alias: alias}
	cc.colConditioner = colConditioner{col: cc, dialect: altFactory.Db().dialect}
	return cc
}

func (c *ColumnWrap[T]) GetCodeWithAlias() string {
	if c.codeWithAlias == "" {
		db := c.altFactory.Db()
		if c.DbName() != c.Alias() {
			c.codeWithAlias = db.
				dialect.QuoteIdentifier(c.altFactory.Alias()) + "." + db.dialect.QuoteIdentifier(c.DbName()) + " AS " + db.dialect.QuoteIdentifier(c.alias)
		} else {
			c.codeWithAlias = db.dialect.QuoteIdentifier(c.altFactory.Alias()) + "." + db.dialect.QuoteIdentifier(c.DbName())
		}
	}
	return c.codeWithAlias
}

func (c *ColumnWrap[T]) GetCode() string {
	if c.code == "" {
		c.code = c.altFactory.
			Db().
			dialect.
			QuoteIdentifier(c.altFactory.Alias()) + "." + c.altFactory.
			Db().
			dialect.
			QuoteIdentifier(c.DbName())
	}
	return c.code
}

func (c ColumnWrap[C]) Alias() string {
	return c.alias
}

func (c ColumnWrap[C]) Factory() IFactory {
	return c.altFactory
}

func (c ColumnWrap[C]) QueryFactory() IQuery {
	return c.altFactory
}

func (c *ColumnWrap[C]) SetAlias(alias string) ICol {
	return newColWrap(c.col, c.altFactory, alias)
}

func (c ColumnWrap[C]) clone(f IFactory) ICol {
	return newColWrap(c.col, f, c.alias)
}

func (c ColumnWrap[C]) Sum() iAggregateCol {
	return Aggr.Sum(&c)
}

func (c ColumnWrap[C]) Avg() iAggregateCol {
	return Aggr.Avg(&c)
}

func (c ColumnWrap[C]) Max() iAggregateCol {
	return Aggr.Max(&c)
}

func (c ColumnWrap[C]) Min() iAggregateCol {
	return Aggr.Min(&c)
}

func (c ColumnWrap[C]) CountDistinct() iAggregateCol {
	return Aggr.CountDistinct(&c)
}

func (c ColumnWrap[C]) Count() iAggregateCol {
	return Aggr.Count(&c)
}

func (c ColumnWrap[C]) Mean() iAggregateCol {
	return Aggr.Mean(&c)
}

func (c ColumnWrap[C]) Median() iAggregateCol {
	return Aggr.Median(&c)
}

func (c ColumnWrap[C]) Mode() iAggregateCol {
	return Aggr.Mode(&c)
}

func (c ColumnWrap[C]) StdDev() iAggregateCol {
	return Aggr.StdDev(&c)
}
func (c ColumnWrap[C]) GroupConcat(splitter string) iAggregateCol {
	return Aggr.GroupConcat(&c, splitter)
}

func (c ColumnWrap[C]) JsonAgg() iAggregateCol {
	return Aggr.JsonAgg(&c)
}

func (c ColumnWrap[C]) ArrayAgg() iAggregateCol {
	return Aggr.ArrayAgg(&c)
}

func (c ColumnWrap[T]) Aggregate(fn preformShare.Aggregator, params ...any) iAggregateCol {
	return &AggregateCol[T]{ICol: &c, Aggregator: fn, body: c.GetCode(), params: params, alias: c.alias, dialect: c.altFactory.Db().dialect}
}
