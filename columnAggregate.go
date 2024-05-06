package preform

import (
	"fmt"
	"github.com/go-preform/preform/dialect"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
)

var Aggr = aggr{}

type aggr struct{}

// implements ICol
type AggregateCol[T any] struct {
	ICol
	Aggregator preformShare.Aggregator
	body       any
	params     []any
	alias      string
	dialect    preformShare.IDialect
}

type iTypedCol interface {
	GetRawPtrScanner() (v any, toScanner func(*any) any)
}

type iAggregateCol interface {
	ICol
	WithDialect(iDialect preformShare.IDialect) squirrel.Sqlizer
}

func (a *AggregateCol[T]) SetAlias(alias string) ICol {
	a.alias = alias
	return a
}

func (a *AggregateCol[T]) GetRawPtrScanner() (vv any, toScanner func(*any) any) {
	var (
		v        T
		toParse  any
		dummyCol = &column[T]{}
	)
	switch a.Aggregator {
	case dialect.AggSum, dialect.AggAvg, dialect.AggMax, dialect.AggMin, dialect.AggMean, dialect.AggMedian, dialect.AggMode, dialect.AggStdDev:
		prepareColumnTypeFunc[T](dummyCol, &v)
	case dialect.AggCount, dialect.AggCountDistinct:
		u64 := uint64(0)
		toParse = &u64
		prepareColumnTypeFunc[T](dummyCol, toParse)
	case dialect.AggGroupConcat:
		str := ""
		toParse = &str
		prepareColumnTypeFunc[T](dummyCol, toParse)
	case dialect.AggArray:
		var arr []T
		toParse = &arr
		prepareColumnTypeFunc[T](dummyCol, toParse, v)
	}
	if dummyCol.sqlScanner == nil {
		if dummyCol.isScanner {
			if toParse == nil {
				toParse = &v
			}
			return toParse, func(a *any) any {
				return toParse
			}
		}
		return v, func(a *any) any {
			return a
		}
	}
	return v, func(a *any) any {
		return dummyCol.sqlScanner().PtrAny(a)
	}
}

func (a AggregateCol[T]) WithDialect(d preformShare.IDialect) ICond {
	a.dialect = d
	return &a
}

func (a AggregateCol[T]) ToSql() (string, []any, error) {
	q, args, err := a.dialect.Aggregate(a.Aggregator, a.body, a.params...).ToSql()
	if a.alias != "" {
		q = fmt.Sprintf("%s AS %s", q, a.dialect.QuoteIdentifier(a.alias))
	} else if a.ICol != nil {
		q = fmt.Sprintf("%s AS %s%s", q, a.DbName(), a.Aggregator)
	}
	return q, args, err
}

func Aggregate[T any](fn preformShare.Aggregator, colOrString any, params ...any) iAggregateCol {
	c := &AggregateCol[T]{Aggregator: fn, body: colOrString, params: params}
	if col, ok := colOrString.(ICol); ok {
		c.ICol = col
		c.alias = col.Alias()
		c.dialect = col.QueryFactory().Db().dialect
	} else {
		c.dialect = DefaultDB.dialect
	}
	return c
}

func (a aggr) Sum(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggSum, colOrString)
}

func (a aggr) Avg(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggAvg, colOrString)
}

func (a aggr) Max(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggMax, colOrString)
}

func (a aggr) Min(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggMin, colOrString)
}

func (a aggr) Count(colOrString any) iAggregateCol {
	return Aggregate[int64](dialect.AggCount, colOrString)
}

func (a aggr) CountDistinct(colOrString any) iAggregateCol {
	return Aggregate[int64](dialect.AggCountDistinct, colOrString)
}

func (a aggr) Mean(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggMean, colOrString)
}

func (a aggr) Median(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggMedian, colOrString)
}

func (a aggr) Mode(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggMode, colOrString)
}

func (a aggr) StdDev(colOrString any) iAggregateCol {
	return Aggregate[float64](dialect.AggStdDev, colOrString)
}
func (a aggr) GroupConcat(colOrString any, splitter string) iAggregateCol {
	return Aggregate[string](dialect.AggGroupConcat, colOrString, splitter)
}

func (a aggr) JsonAgg(colOrString any) iAggregateCol {
	return Aggregate[string](dialect.AggJson, colOrString)
}

func (a aggr) ArrayAgg(colOrString any) iAggregateCol {
	return Aggregate[string](dialect.AggArray, colOrString)
}

func (a Column[T]) Aggregate(fn preformShare.Aggregator, params ...any) iAggregateCol {
	return &AggregateCol[T]{ICol: &a, Aggregator: fn, body: a.GetCode(), params: params, alias: a.Alias()}
}

func (a Column[T]) Sum() iAggregateCol {
	return Aggr.Sum(&a)
}

func (a Column[T]) Avg() iAggregateCol {
	return Aggr.Avg(&a)
}

func (a Column[T]) Max() iAggregateCol {
	return Aggr.Max(&a)
}

func (a Column[T]) Min() iAggregateCol {
	return Aggr.Min(&a)
}

func (a Column[T]) CountDistinct() iAggregateCol {
	return Aggr.CountDistinct(&a)
}

func (a Column[T]) Count() iAggregateCol {
	return Aggr.Count(&a)
}

func (a Column[T]) Mean() iAggregateCol {
	return Aggr.Mean(&a)
}

func (a Column[T]) Median() iAggregateCol {
	return Aggr.Median(&a)
}

func (a Column[T]) Mode() iAggregateCol {
	return Aggr.Mode(&a)
}

func (a Column[T]) JsonAgg() iAggregateCol {
	return Aggr.JsonAgg(&a)
}

func (a Column[T]) ArrayAgg() iAggregateCol {
	return Aggr.ArrayAgg(&a)
}

func (a Column[T]) StdDev() iAggregateCol {
	return Aggr.StdDev(&a)
}
func (a Column[T]) GroupConcat(splitter string) iAggregateCol {
	return Aggr.GroupConcat(&a, splitter)
}
