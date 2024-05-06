package preform

import (
	"context"
	"errors"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"strings"
)

type UpdateConfig struct {
	Tx        *Tx
	Cascading bool
	Cols      []ICol
	Ctx       context.Context
}

// UpdateByPk update body by pk, can cascade, will skip pk columns
func (f *Factory[FPtr, B]) UpdateByPk(body any, cfgs ...UpdateConfig) (updated int64, err error) {
	var (
		db            = f.Db()
		exec       DB = db
		query         = db.sqStmtBuilder.UpdateFast(strings.Split(f.fromClause(), " AS ")[0])
		cfg        UpdateConfig
		ctx        = db.ctx
		isPk       bool
		cols       = f.columns
		mBody      = body.(iModelBody)
		bodyValues = mBody.FieldValuePtrs()
	)
	if len(cfgs) != 0 {
		cfg = cfgs[0]
	}
	if cfg.Tx != nil {
		exec = cfg.Tx
	}
	if cfg.Ctx != nil {
		ctx = cfg.Ctx
	}
	if cfg.Cols != nil {
		cols = cfg.Cols
	}
	for _, col := range cols {
		_, _, isPk, _ = col.properties()
		if !isPk {
			query = query.Set(db.dialect.QuoteIdentifier(col.DbName()), col.unwrapPtrForUpdate(bodyValues[col.GetPos()]))
		}
	}
	if f.fixCond != nil {
		query = query.Where(f.fixCond)
	}
	query = query.Where(f.PkCondByValues(bodyValues))
	q, args, err := db.dialect.UpdateSqlizer(query)
	if err != nil {
		return 0, err
	}
	res, err := exec.RelatedFactory([]preformShare.IQueryFactory{f}).ExecContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	updated, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if cfg.Cascading {
		if len(f.relations) != 0 {
			var (
				bodyAsiModelRelatedBody = body.(iModelRelatedBody)
				relatedBodies           []any
				relatedBody             any
			)
			for _, rel := range f.relations {
				relatedBodies = rel.unwrapPtrBodyToTargetBodies(bodyAsiModelRelatedBody.RelatedValuePtrs()[rel.Index()])
				if relatedBodies != nil {
					for _, relatedBody = range relatedBodies {
						if relatedBody != nil {
							if _, err = rel.TargetFactory().UpdateByPk(relatedBody, cfgs...); err != nil {
								return 0, err
							}
						}
					}
				}
			}
		}
	}
	return
}

type UpdateBuilder[B any] struct {
	Builder  preformShare.UpdateBuilder
	factory  IFactory
	execer   preformShare.QueryRunner
	hasWhere bool
	bodyCols []ICol
	bodies   []*B
	ctx      context.Context
}

func (f *Factory[FPtr, B]) Update() *UpdateBuilder[B] {
	b := f.Db().sqStmtBuilder.UpdateFast(strings.Split(f.fromClause(), " AS ")[0])
	if f.fixCond != nil {
		b = b.Where(f.fixCond)
	}
	return &UpdateBuilder[B]{Builder: b, factory: f.Definition, execer: f.Db()}
}

func (b *UpdateBuilder[B]) Ctx(ctx context.Context) *UpdateBuilder[B] {
	b.ctx = ctx
	return b
}

func (b *UpdateBuilder[B]) Where(cond ICond) *UpdateBuilder[B] {
	b.Builder = b.Builder.Where(noTableCodeWhere(cond))
	b.hasWhere = true
	return b
}

func noTableCodeWhere(cond ICond) ICond {
	switch cond.(type) {
	case IColConditioner:
		return cond.(IColConditioner).NoParentCode()
	case squirrel.And:
		condAnd := cond.(squirrel.And)
		for i := range condAnd {
			condAnd[i] = noTableCodeWhere(condAnd[i])
		}
		return condAnd
	case squirrel.Or:
		condOr := cond.(squirrel.Or)
		for i := range condOr {
			condOr[i] = noTableCodeWhere(condOr[i])
		}
		return condOr
	}
	return cond
}

func (b *UpdateBuilder[B]) Columns(cols ...ICol) *UpdateBuilder[B] {
	for _, col := range cols {
		if col.QueryFactory() == b.factory {
			b.bodyCols = append(b.bodyCols, col)
		}
	}
	return b
}

func (b *UpdateBuilder[B]) SetBodies(body ...*B) *UpdateBuilder[B] {
	b.bodies = body
	return b
}

func (b *UpdateBuilder[B]) Set(col any, value any) *UpdateBuilder[B] {
	switch col.(type) {
	case ICol:
		b.Builder = b.Builder.Set(b.factory.Db().GetDialect().QuoteIdentifier(col.(ICol).DbName()), value)
	case string:
		b.Builder = b.Builder.Set(col.(string), value)
	}
	return b
}

func (b *UpdateBuilder[B]) SetMap(clause map[string]any) *UpdateBuilder[B] {
	b.Builder = b.Builder.SetMap(clause)
	return b
}

func (b *UpdateBuilder[B]) LimitOffset(limit, offset uint64) *UpdateBuilder[B] {
	b.Builder = b.Builder.Limit(limit).Offset(offset)
	return b
}

func (b UpdateBuilder[B]) ToSql() (string, []any, error) {
	var db = b.factory.Db()
	if l := len(b.bodies); l != 0 {
		var (
			cols       = b.bodyCols
			pk         = b.factory.Pks()
			i          int
			col        ICol
			bodyValues []any
			pkColName  = pk[0].DbName()
			pkColPos   = pk[0].GetPos()
			pkValues   = make([]any, l)
		)
		if len(pk) != 1 && l > 1 {
			return "", nil, errors.New("update by bodies must have one pk")
		}
		if len(cols) == 0 {
			cols = b.factory.Columns()
		}
		if l == 1 {
			bodyValues = any(b.bodies[0]).(iModelBody).FieldValuePtrs()
			pkValues[0] = bodyValues[pkColPos]
			for _, col = range cols {
				b.Builder = b.Builder.Set(db.GetDialect().QuoteIdentifier(col.DbName()), col.unwrapPtrForUpdate(bodyValues[col.GetPos()]))
			}
		} else {
			var (
				cases = make([]squirrel.CaseBuilder, len(cols))
			)
			for i = range cols {
				cases[i] = squirrel.Case(db.GetDialect().QuoteIdentifier(pkColName))
			}
			for _, body := range b.bodies {
				for i, col = range cols {
					bodyValues = any(body).(iModelBody).FieldValuePtrs()
					pkValues[i] = bodyValues[pkColPos]
					cases[i] = cases[i].When(bodyValues[pkColPos], col.unwrapPtrForUpdate(bodyValues[col.GetPos()]))
				}
			}
		}
		if !b.hasWhere {
			b.Builder = b.Builder.Where(squirrel.Eq{pkColName: pkValues})
		}
	}
	return db.dialect.UpdateSqlizer(b.Builder)
}

func (b UpdateBuilder[B]) Exec(tx ...preformShare.QueryRunner) (int64, error) {
	q, args, err := b.ToSql()
	if err != nil {
		return 0, err
	}
	if b.ctx == nil {
		b.ctx = b.factory.Db().ctx
	}
	if len(tx) == 0 {
		tx = []preformShare.QueryRunner{b.execer}
	}
	res, err := tx[0].RelatedFactory([]preformShare.IQueryFactory{b.factory}).ExecContext(b.ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
