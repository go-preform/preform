package preform

import (
	"fmt"
	"github.com/go-preform/squirrel"
	"reflect"
)

type IForeignKey interface {
	ICol
	AssociatedKeys() []IColFromFactory
}

type ForeignKey[T any] struct {
	*Column[T]
	associatedKeys []IColFromFactory
}

type ForeignKeySetter[T any] struct {
	*ColumnSetter[T]
	fk *ForeignKey[T]
}

func (c *ForeignKey[T]) initCol(ref reflect.StructField, dbName string, factory IFactory, pos int) {
	c.Column = &Column[T]{}
	c.Column.initCol(ref, dbName, factory, pos)
	c.alias = ref.Name
}

func (c ForeignKey[T]) clone(f IFactory) ICol {
	var (
		cc   = c
		cPtr = &cc
	)
	cc.Column = c.Column.clone(f).(*Column[T])
	return cPtr
}

func (c ForeignKey[T]) AssociatedKeys() []IColFromFactory {
	return c.associatedKeys
}

func (c *ForeignKey[T]) SetAlias(alias string) ICol {
	return newColWrap(c, c.factory, alias)
}

func (c *ForeignKey[T]) Join(col ...IColFromFactory) iForeignKeyJoin {
	if len(col) == 0 {
		if len(c.associatedKeys) == 0 {
			panic(fmt.Errorf("no associated keys for %s", c.GetCode()))
		}
		return ForeignKeyJoin{[]IColFromFactory{c.associatedKeys[0]}, []IColFromFactory{c}, nil}
	}
	return ForeignKeyJoin{[]IColFromFactory{col[0]}, []IColFromFactory{c}, nil}
}

func (c ForeignKey[T]) ToSql() (string, []interface{}, error) {
	return c.GetCode(), []any{}, nil
}

func (c ForeignKey[T]) ToJoinSql(queryFactory IQuery) (IQuery, string, []interface{}, error) {
	return c.Join().ToJoinSql(queryFactory)
}

func (c *ForeignKey[T]) GetCode() string {
	if c.code == "" {
		c.code = c.factory.Db().dialect.QuoteIdentifier(c.factory.Alias()) + "." + c.factory.Db().dialect.QuoteIdentifier(c.dbName)
	}
	return c.code
}
func (c *ForeignKey[T]) GetCodeWithAlias() string {
	if c.codeWithAlias == "" {
		c.codeWithAlias = c.factory.Db().dialect.QuoteIdentifier(c.factory.Alias()) + "." + c.factory.Db().dialect.QuoteIdentifier(c.dbName) + " AS " + c.factory.Db().dialect.QuoteIdentifier(c.alias)
	}
	return c.codeWithAlias
}

func (c ForeignKey[T]) Alias() string {
	return c.alias
}

func (c ForeignKey[T]) TargetFactory() IFactory {
	return c.associatedKeys[0].Factory()
}

func SetForeignKey[T any](c *ForeignKey[T]) *ForeignKeySetter[T] {
	return &ForeignKeySetter[T]{&ColumnSetter[T]{c.Column}, c}
}

func (s *ForeignKeySetter[T]) SetRelation(related IRelation, col ...IColFromFactory) *ForeignKeySetter[T] {
	related.InitRelation(col...)
	s.fk.associatedKeys = append(s.fk.associatedKeys, col[1])
	return s
}

type iForeignKeyJoin interface {
	ToJoinSql(queryFactory IQuery) (IQuery, string, []any, error)
	Join(col ...IColFromFactory) iForeignKeyJoin
	TargetFactory() IFactory
}

type ForeignKeyJoin struct {
	src             []IColFromFactory
	target          []IColFromFactory
	extraConditions []ICond
}

func (c ForeignKeyJoin) Join(col ...IColFromFactory) iForeignKeyJoin {
	return c
}
func (c ForeignKeyJoin) TargetFactory() IFactory {
	return c.target[0].Factory()
}

func (c ForeignKeyJoin) ToJoinSql(queryFactory IQuery) (IQuery, string, []any, error) {
	f := c.src[0].Factory()
	var (
		fromCols = c.src
		toCols   = c.target
	)
	if f == queryFactory {
		f = toCols[0].Factory()
		fromCols, toCols = toCols, fromCols
	}
	var (
		fixedCond    = f.FixedCondition()
		extraCondLen = len(c.extraConditions)
		cond         = make(squirrel.And, len(fromCols)+extraCondLen)
		toCol        IColFromFactory
	)
	copy(cond, c.extraConditions)
	for i, fromCol := range fromCols {
		toCol = toCols[i]
		if isArray, _, _, _ := fromCol.properties(); isArray {
			if isArray, _, _, _ := toCol.properties(); isArray {
				cond[i+extraCondLen] = toCol.HasAny(fromCol)
			} else {
				cond[i+extraCondLen] = fromCol.Any(toCol)
			}
		} else if isArray, _, _, _ := toCol.properties(); isArray {
			cond[i+extraCondLen] = toCol.Any(fromCol)
		} else {
			cond[i+extraCondLen] = fromCol.Eq(toCol)
		}
	}
	if fixedCond != nil {
		cond = append(cond, f.FixedCondition())
	}

	q, args, err := cond.ToSql()
	return f, fmt.Sprintf("%s ON %s", f.fromClause(), q), args, err
}
