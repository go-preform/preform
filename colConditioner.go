package preform

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	preformSqlizer "github.com/go-preform/preform/sqlizer"
	"github.com/go-preform/squirrel"
)

type ICond = squirrel.Sqlizer

type colConditioner struct {
	col        ICol
	dialect    preformShare.IDialect
	conditions []func(ICol, preformShare.IDialect) ICond //prevent render before init
	isOr       bool
}

type IColConditioner interface {
	ICond
	Col() ICol
	Expr(exprWithFmt string, args ...any) colConditioner
	Eq(v any) colConditioner
	NotEq(v any) colConditioner
	Like(v any) colConditioner
	Gt(v any) colConditioner
	GtOrEq(v any) colConditioner
	Lt(v any) colConditioner
	LtOrEq(v any) colConditioner
	Between(v1, v2 any) colConditioner
	And(cond ...ICond) colConditioner
	Or(cond ...ICond) colConditioner
	Contains(arrCol any) colConditioner
	ContainsBy(arrCol any) colConditioner
	HasAny(arrCol any) colConditioner
	Any(v any) colConditioner
	Concat(arrCol any) colConditioner
	NoParentCode() colConditioner
}

// ToSql
func (c colConditioner) ToSql() (query string, args []any, err error) {
	if len(c.conditions) == 1 {
		query, args, err = c.conditions[0](c.col, c.dialect).ToSql()
	} else {
		conds := make([]ICond, len(c.conditions))
		for i, cond := range c.conditions {
			conds[i] = cond(c.col, c.dialect)
		}
		if c.isOr {
			query, args, err = squirrel.Or(conds).ToSql()
		} else {
			query, args, err = squirrel.And(conds).ToSql()
		}
	}
	if err != nil {
		return
	}
	return preformShare.NestCondSql(query, args)
}

func (c colConditioner) NoParentCode() colConditioner {
	f := c.col.QueryFactory().(IFactory)
	cc := newColWrap(c.col, f, "")
	cc.code = f.Db().GetDialect().QuoteIdentifier(c.col.DbName())
	c.col = cc
	return c
}

func (c colConditioner) Col() ICol {
	return c.col
}

func (c colConditioner) Expr(exprWithFmt string, args ...any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return squirrel.Expr(fmt.Sprintf(exprWithFmt, c.GetCode()), args...)
	})
	return c
}

// Eq
func (c colConditioner) Eq(v any) colConditioner {
	if isArray, _, _, _ := c.col.properties(); isArray {
		c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
			return preformSqlizer.ArrayEq{ArrColA: c, ArrColB: v}.WithDialect(d)
		})
	} else {
		c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.Eq(c, v) })
	}
	return c
}

// Neq
func (c colConditioner) NotEq(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.NotEq(c, v) })
	return c
}

// like
func (c colConditioner) Like(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.Like(c, v) })
	return c
}

// Gt
func (c colConditioner) Gt(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.Gt(c, v) })
	return c
}

// Gte
func (c colConditioner) GtOrEq(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.GtOrEq(c, v) })
	return c
}

// Lt
func (c colConditioner) Lt(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.Lt(c, v) })
	return c
}

// Lte
func (c colConditioner) LtOrEq(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return d.LtOrEq(c, v) })
	return c
}

// Between
func (c colConditioner) Between(v1, v2 any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return d.Between(c, v1, v2)
	})
	return c
}

func (c colConditioner) And(cond ...ICond) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return squirrel.And(cond) })
	return c
}

func (c colConditioner) Or(cond ...ICond) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond { return squirrel.Or(cond) })
	return c
}

func (c colConditioner) Any(v any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return preformSqlizer.ArrayAny{ArrCol: c, Value: v}.
			WithDialect(d)
	})
	return c
}

func (c colConditioner) HasAny(arrayCol any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return preformSqlizer.ArrayHasAny{ArrColA: c, ArrColB: arrayCol}.WithDialect(d)
	})
	return c
}

func (c colConditioner) Concat(arrayCol any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return preformSqlizer.ArrayConcat{ArrColA: c, ArrColB: arrayCol}.WithDialect(d)
	})
	return c
}

func (c colConditioner) Contains(arrayCol any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return preformSqlizer.ArrayContains{ArrColA: c, ArrColB: arrayCol}.WithDialect(d)
	})
	return c
}

func (c colConditioner) ContainsBy(arrayCol any) colConditioner {
	c.conditions = append(c.conditions, func(c ICol, d preformShare.IDialect) ICond {
		return preformSqlizer.ArrayContainsBy{ArrColA: c, ArrColB: arrayCol}.WithDialect(d)
	})
	return c
}
