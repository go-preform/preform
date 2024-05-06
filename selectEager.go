package preform

import (
	"context"
	"errors"
	"github.com/go-preform/squirrel"
)

func (b *SelectQuery[B]) Eager(relation ...IRelation) *SelectQuery[B] {
	for _, r := range relation {
		b.eagerLoaders = append(b.eagerLoaders,
			r.eagerLoader())
	}
	return b
}

func (b *SelectQuery[B]) eagerLoad(ptrs any, fast ...bool) error {
	var (
		err       error
		modelPtrs = ptrs.([]*B)
	)
	for _, eagerLoader := range b.eagerLoaders {
		err = eagerLoader.loadEagers(b.ctx, modelPtrs, len(fast) != 0 && fast[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) loadEagers(ctx context.Context, srcModels any, useFast bool) error {
	var err error
	srcModels, err = loader.loadEager(ctx, srcModels, useFast)
	if err != nil {
		return err
	}
	for _, eagerLoader := range loader.eagerChain {
		srcModels, err = eagerLoader.loadEager(ctx, srcModels, useFast)
		if err != nil {
			return err
		}
		if srcModels == nil {
			return nil
		}
	}
	return nil
}

func (loader EagerLoader[SrcBody, TargetFactory, TargetBody]) loadEager(ctx context.Context, srcModels any, useFast bool) (relatedModels any, err error) {
	if bodies, ok := srcModels.([]SrcBody); ok {
		var (
			bl = len(bodies)
		)
		if bl == 0 {
			return
		}
		var (
			r             = loader.getRelation().(*relation[SrcBody, TargetFactory, TargetBody])
			toCompare     iJoiner[SrcBody, TargetFactory, TargetBody]
			relatedBodies []TargetBody
			cond          squirrel.And
		)
		if bl < 15 || r.hasArrayKeys {
			toCompare, cond = prepareBFCompare(r, bodies)
		} else {
			toCompare, cond = prepareHashCompare(r, bodies)
		}
		if len(r.cond) != 0 {
			cond = append(r.cond, cond...)
		}
		q := SelectByFactory[TargetBody](r.targetFactory, r.targetFactory.IModelScanner().(IModelScanner[TargetBody])).Where(cond)
		if loader.limit != 0 {
			q = q.Limit(uint64(loader.limit))
		}
		if loader.offset != 0 {
			q = q.Offset(uint64(loader.offset))
		}
		if loader.orderBy != nil {
			switch loader.orderBy.(type) {
			case string:
				q = q.OrderBy(loader.orderBy.(string))
			case squirrel.Sqlizer:
				qq, _, _ := loader.orderBy.(squirrel.Sqlizer).ToSql()
				q = q.OrderBy(qq)
			}
		}
		if len(loader.conds) != 0 {
			q = q.Where(squirrel.And(loader.conds))
		}
		if len(loader.columns) != 0 {
			q = q.Columns(loader.columns...)
		}

		if useFast {
			relatedBodies, err = q.Ctx(ctx).GetAllFast()
		} else {
			relatedBodies, err = q.Ctx(ctx).GetAll()
		}
		if err != nil {
			return
		}
		return toCompare.join(relatedBodies), nil
	}
	err = errors.New("srcModels is not []SrcBody")
	return
}
