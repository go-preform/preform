package preform

import (
	"context"
	preformShare "github.com/go-preform/preform/share"
	"strings"
)

func (f *Factory[FPtr, B]) DeleteByPk(body any, cfgs ...EditConfig) (Deleted int64, err error) {
	var (
		db       = f.Db()
		exec  DB = db
		query    = db.sqStmtBuilder.DeleteFast(strings.Split(f.fromClause(), " AS ")[0])
		cfg   EditConfig
		ctx   = db.ctx
		mBody = body.(iModelBody)
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
	if f.fixCond != nil {
		query = query.Where(noTableCodeWhere(f.fixCond))
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
							if _, err = rel.TargetFactory().DeleteByPk(relatedBody, cfgs...); err != nil {
								return 0, err
							}
						}
					}
				}
			}
		}
	}
	query = query.Where(f.PkCondByBody(mBody))
	q, args, err := db.dialect.DeleteSqlizer(query)
	if err != nil {
		return 0, err
	}
	res, err := exec.RelatedFactory([]preformShare.IQueryFactory{f}).ExecContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	Deleted, err = res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return
}

type DeleteBuilder[B any] struct {
	Builder  preformShare.DeleteBuilder
	factory  IFactory
	execer   DB
	hasWhere bool
	bodyCols []ICol
	bodies   []*B
	ctx      context.Context
}

func (f *Factory[FPtr, B]) Delete() *DeleteBuilder[B] {
	b := f.Db().sqStmtBuilder.DeleteFast(strings.Split(f.fromClause(), " AS ")[0])
	if f.fixCond != nil {
		b = b.Where(f.fixCond)
	}
	return &DeleteBuilder[B]{Builder: b, factory: f.Definition, execer: f.Db()}
}

func (b *DeleteBuilder[B]) Ctx(ctx context.Context) *DeleteBuilder[B] {
	b.ctx = ctx
	return b
}

func (b *DeleteBuilder[B]) Where(cond ICond) *DeleteBuilder[B] {
	b.Builder = b.Builder.Where(noTableCodeWhere(cond))
	b.hasWhere = true
	return b
}

func (b *DeleteBuilder[B]) LimitOffset(limit, offset uint64) *DeleteBuilder[B] {
	b.Builder = b.Builder.Limit(limit).Offset(offset)
	return b
}

func (b DeleteBuilder[B]) ToSql() (string, []any, error) {
	return b.factory.Db().dialect.DeleteSqlizer(b.Builder)
}

func (b DeleteBuilder[B]) Exec(tx ...preformShare.QueryRunner) (int64, error) {
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
