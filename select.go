package preform

import (
	"context"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"github.com/jmoiron/sqlx"
)

type SelectQuery[B any] struct {
	scanner                  IModelScanner[B]
	selectBuilder            preformShare.SelectBuilder
	limit                    uint64
	db                       DB
	queryFactory             IQuery
	relatedFactoriesForCache []preformShare.IQueryFactory
	noCache                  bool
	Cols                     []ICol
	ColValTpl                []func() (vv any, toScanner func(*any) any)
	ctx                      context.Context
	prepared                 *sqlx.Stmt
	eagerLoaders             []IEagerLoader
	forceBodyScan            bool
}

func (b *SelectQuery[B]) PlaceholderFormat(f squirrel.PlaceholderFormat) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.PlaceholderFormat(f)
	return b
}

func (b SelectQuery[B]) ToSql() (string, []interface{}, error) {
	return b.selectBuilder.ToSql()
}

func (b SelectQuery[B]) MustSql() (string, []interface{}) {
	return b.selectBuilder.MustSql()
}

func (b *SelectQuery[B]) QuoteIdentifier(id string) string {
	return b.db.Db().dialect.QuoteIdentifier(id)
}

func (b *SelectQuery[B]) Prefix(sql string, args ...interface{}) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Prefix(sql, args...)
	return b
}

func (b *SelectQuery[B]) PrefixExpr(expr ICond) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.PrefixExpr(expr)
	return b
}

func (b *SelectQuery[B]) Distinct() *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Distinct()
	return b
}

func (b *SelectQuery[B]) Options(options ...string) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Options(options...)
	return b
}

func (b *SelectQuery[B]) Columns(cols ...any) *SelectQuery[B] {
	return b.convertCols(cols...)
}

func (b *SelectQuery[B]) RemoveColumns() *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.RemoveColumns()
	return b
}

func (b *SelectQuery[B]) Column(column interface{}, args ...interface{}) *SelectQuery[B] {
	if len(args) == 0 {
		return b.convertCols(column)
	}
	b.selectBuilder = b.selectBuilder.Column(column, args...)
	return b
}

func (b *SelectQuery[B]) From(from any) *SelectQuery[B] {
	if f, ok := from.(IQuery); ok {
		b.selectBuilder = b.selectBuilder.From(f.fromClause())
	} else {
		b.selectBuilder = b.selectBuilder.From(from.(string))
	}
	return b
}

func (b *SelectQuery[B]) FromSelect(from preformShare.SelectBuilder, alias string) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.FromSelectFast(from, alias)
	return b
}

func (b *SelectQuery[B]) FromSubQuery(from *subQueryCol, alias string) *SelectQuery[B] {
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, from.relatedFactories...)
	b.selectBuilder = b.selectBuilder.FromSelectFast(from.query, alias)
	return b
}

func (b *SelectQuery[B]) JoinClause(pred interface{}, args ...interface{}) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.JoinClause(pred, args...)
	return b
}

func (b *SelectQuery[B]) JoinRelation(rel IRelation) *SelectQuery[B] {
	q, s, a, _ := rel.JoinClause().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.Join(s, a...)
	return b
}

func (b *SelectQuery[B]) LeftRelation(rel IRelation) *SelectQuery[B] {
	q, s, a, _ := rel.JoinClause().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.LeftJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) RightRelation(rel IRelation) *SelectQuery[B] {
	q, s, a, _ := rel.JoinClause().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.RightJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) InnerRelation(rel IRelation) *SelectQuery[B] {
	q, s, a, _ := rel.JoinClause().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.InnerJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) CrossRelation(rel IRelation) *SelectQuery[B] {
	q, s, a, _ := rel.JoinClause().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.CrossJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) JoinForeignKey(joinKey iForeignKeyJoin) *SelectQuery[B] {
	q, s, a, _ := joinKey.Join().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.Join(s, a...)
	return b
}

func (b *SelectQuery[B]) LeftJoinForeignKey(joinKey iForeignKeyJoin) *SelectQuery[B] {
	q, s, a, _ := joinKey.Join().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.LeftJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) RightJoinForeignKey(joinKey iForeignKeyJoin) *SelectQuery[B] {
	q, s, a, _ := joinKey.Join().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.RightJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) InnerJoinForeignKey(joinKey iForeignKeyJoin) *SelectQuery[B] {
	q, s, a, _ := joinKey.Join().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.InnerJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) CrossJoinForeignKey(joinKey iForeignKeyJoin) *SelectQuery[B] {
	q, s, a, _ := joinKey.Join().ToJoinSql(b.queryFactory)
	b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, q)
	b.selectBuilder = b.selectBuilder.CrossJoin(s, a...)
	return b
}

func (b *SelectQuery[B]) Join(join string, rest ...interface{}) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.Join(join, rest...)
	return b
}

func (b *SelectQuery[B]) LeftJoin(join string, rest ...interface{}) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.LeftJoin(join, rest...)
	return b
}

func (b *SelectQuery[B]) RightJoin(join string, rest ...interface{}) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.RightJoin(join, rest...)
	return b
}

func (b *SelectQuery[B]) InnerJoin(join string, rest ...interface{}) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.InnerJoin(join, rest...)
	return b
}

func (b *SelectQuery[B]) CrossJoin(join string, rest ...interface{}) *SelectQuery[B] {
	b.noCache = true
	b.selectBuilder = b.selectBuilder.CrossJoin(join, rest...)
	return b
}

func (b *SelectQuery[B]) Where(pred interface{}, args ...interface{}) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Where(pred, args...)
	return b
}

func (b *SelectQuery[B]) GroupBy(groupBys ...any) *SelectQuery[B] {
	var (
		groupByStrs = make([]string, len(groupBys))
		s           squirrel.Sqlizer
		ok          bool
	)
	for i, groupBy := range groupBys {
		if s, ok = groupBy.(squirrel.Sqlizer); ok {
			groupByStrs[i], _, _ = s.ToSql()
		} else if _, ok = groupBy.(string); ok {
			groupByStrs[i] = groupBy.(string)
		} else {
			groupByStrs[i] = fmt.Sprintf("%v", groupBy)
		}
	}
	b.selectBuilder = b.selectBuilder.GroupBy(groupByStrs...)
	return b
}

func (b *SelectQuery[B]) Having(pred interface{}, rest ...interface{}) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Having(pred, rest...)
	return b
}

func (b *SelectQuery[B]) OrderByClause(pred interface{}, args ...interface{}) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.OrderByClause(pred, args...)
	return b
}

func (b *SelectQuery[B]) OrderBy(orderBys ...string) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.OrderBy(orderBys...)
	return b
}

func (b *SelectQuery[B]) Limit(limit uint64) *SelectQuery[B] {
	b.limit = limit
	b.selectBuilder = b.selectBuilder.Limit(limit)
	return b
}

func (b *SelectQuery[B]) RemoveLimit() *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.RemoveLimit()
	return b
}

func (b *SelectQuery[B]) Offset(offset uint64) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Offset(offset)
	return b
}

func (b *SelectQuery[B]) RemoveOffset() *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.RemoveOffset()
	return b
}

func (b *SelectQuery[B]) Suffix(sql string, args ...interface{}) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.Suffix(sql, args...)
	return b
}

func (b *SelectQuery[B]) SuffixExpr(expr squirrel.Sqlizer) *SelectQuery[B] {
	b.selectBuilder = b.selectBuilder.SuffixExpr(expr)
	return b
}

func (b *SelectQuery[B]) AsSubQuery(alias string) squirrel.Sqlizer {
	return &subQueryCol{query: b.selectBuilder, alias: alias, relatedFactories: b.relatedFactoriesForCache}
}

type subQueryCol struct {
	query            preformShare.SelectBuilder
	alias            string
	relatedFactories []preformShare.IQueryFactory
}

func (s *subQueryCol) ToSql() (string, []interface{}, error) {
	q, args, err := s.query.ToSql()
	return fmt.Sprintf("(%s) AS %s", q, s.alias), args, err
}

func (b *SelectQuery[B]) convertCols(cols ...any) *SelectQuery[B] {
	var (
		ret       = make([]string, 0, len(cols))
		c         IColFromFactory
		cc        ICond
		cd        preformShare.ISqlizerWithDialect
		tc        iTypedCol
		a         []any
		s         string
		ok        bool
		colValTpl []func() (vv any, toScanner func(*any) any)
		subQ      *subQueryCol
	)
	for _, col := range cols {
		if c, ok = col.(IColFromFactory); ok {
			ret = append(ret, c.GetCodeWithAlias())
			if b.forceBodyScan || (b.queryFactory != nil && (c.QueryFactory() == b.queryFactory || c.QueryFactory().BodyType() == b.queryFactory.BodyType())) {
				b.Cols = append(b.Cols, c)
			} else {
				b.queryFactory = nil
			}
			colValTpl = append(colValTpl, c.GetRawPtrScanner)
		} else {
			if !b.forceBodyScan {
				b.queryFactory = nil
			}
			if tc, ok = col.(iTypedCol); ok {
				if len(ret) != 0 {
					b.selectBuilder = b.selectBuilder.Columns(ret...)
					b.ColValTpl = append(b.ColValTpl, colValTpl...)
					ret = []string{}
					colValTpl = []func() (vv any, toScanner func(*any) any){}
				}
				b.ColValTpl = append(b.ColValTpl, tc.GetRawPtrScanner)
				if cd, ok = col.(preformShare.ISqlizerWithDialect); ok {
					s, a, _ = cd.ToSql(b.db.GetDialect())
					b.selectBuilder = b.selectBuilder.Column(s, a...)
				} else if cc, ok = col.(squirrel.Sqlizer); ok {
					b.selectBuilder = b.selectBuilder.Column(cc)
				}
			} else {
				if s, ok = col.(string); ok {
					b.noCache = true
					colValTpl = append(colValTpl, newAnyPtr)
					ret = append(ret, s)
				} else if cc, ok = col.(squirrel.Sqlizer); ok {
					if subQ, ok = cc.(*subQueryCol); ok {
						b.relatedFactoriesForCache = append(b.relatedFactoriesForCache, subQ.relatedFactories...)
					}
					b.ColValTpl = append(b.ColValTpl, newAnyPtr)
					b.selectBuilder = b.selectBuilder.Column(cc)
				} else {
					b.noCache = true
					colValTpl = append(colValTpl, newAnyPtr)
					ret = append(ret, fmt.Sprintf("%v", col))
				}
			}
		}
	}
	if len(ret) != 0 {
		b.selectBuilder = b.selectBuilder.Columns(ret...)
		b.ColValTpl = append(b.ColValTpl, colValTpl...)
	}
	return b
}

func newAnyPtr() (vv any, toScanner func(*any) any) {
	return new(any), func(a *any) any {
		return a
	}
}
