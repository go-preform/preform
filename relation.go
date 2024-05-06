package preform

import (
	"context"
	"github.com/go-preform/squirrel"
)

type IRelation interface {
	InitRelation(localAndForeignKeyPairs ...IColFromFactory) IRelation
	ExtraCond(cond ICond) IRelation
	TargetFactory() IFactory
	LocalFactory() IFactory
	LocalKeys() []IColFromFactory
	ForeignKeys() []IColFromFactory
	Index() uint32
	clone(targetFactory ...IFactory) IRelation
	getRelation() any
	eagerLoader() IEagerLoader
	Eager(IRelation) IEagerLoader
	Columns(...any) IEagerLoader
	Where(...ICond) IEagerLoader
	Limit(uint32) IEagerLoader
	Offset(uint32) IEagerLoader
	OrderBy(any) IEagerLoader
	unwrapPtrBodyToTargetBodies(ptr any) []any
	setForeignKey(srcBody any, targetBodies ...any) error
	Name() string
	prepare(name string, pos uint32)
	JoinClause() ForeignKeyJoin
	IsMiddleTable() bool
}

type iRelation[TargetBody any] interface {
	TargetBody() TargetBody
}

type IEagerLoader interface {
	IRelation
	loadEagers(ctx context.Context, srcModels any, useFast bool) error
	loadEager(ctx context.Context, srcModels any, useFast bool) (relatedModels any, err error)
	addChain(...IEagerLoader) IEagerLoader
}

type iModelRelatedBody interface {
	iModelBody
	RelatedByPos(pos uint32) any
	RelatedValuePtrs() []any
}

type relation[SrcBody iModelRelatedBody, TargetFactory IFactory, TargetBody any] struct {
	localKeys     []IColFromFactory
	foreignKeys   []IColFromFactory
	index         uint32
	isMany        bool
	relationPtr   IRelation
	cond          squirrel.And
	name          string
	targetFactory TargetFactory
	hasArrayKeys  bool
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) InitRelation(rr IRelation, localAndForeignKeyPairs ...IColFromFactory) {
	var (
		l       = len(localAndForeignKeyPairs)
		isArray bool
	)
	if l == 0 {
		panic("localAndForeignKeyPairs must not be empty")
	}
	if l%2 != 0 {
		panic("localAndForeignKeyPairs must be even")
	}
	r.targetFactory = localAndForeignKeyPairs[1].Factory().SetAlias(r.name).(TargetFactory)
	for i := 0; i < l; i += 2 {
		r.localKeys = append(r.localKeys, localAndForeignKeyPairs[i])
		r.foreignKeys = append(r.foreignKeys, r.targetFactory.Columns()[localAndForeignKeyPairs[i+1].GetPos()].(IColFromFactory))
		isArray, _, _, _ = localAndForeignKeyPairs[i].properties()
		if isArray {
			r.hasArrayKeys = true
		} else {
			isArray, _, _, _ = localAndForeignKeyPairs[i+1].properties()
			if isArray {
				r.hasArrayKeys = true
			}
		}
	}
	r.relationPtr = rr
}

func (r relation[SrcBody, TargetFactory, TargetBody]) IsMiddleTable() bool {
	return false

}

func (r relation[SrcBody, TargetFactory, TargetBody]) unwrapPtrBodyToTargetBodies(ptr any) []any {
	var (
		res  []any
		tNil *TargetBody
	)
	if r.isMany {
		arr := *(ptr.(*[]*TargetBody))
		res = make([]any, len(arr))
		for i, body := range arr {
			res[i] = body
		}
		return res
	}
	v := *(ptr.(**TargetBody))
	if v == tNil {
		return []any{nil}
	}
	return []any{v}
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Index() uint32 {
	return r.index
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Name() string {
	return r.name
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) prepare(name string, pos uint32) {
	r.name = name
	r.index = pos
}

func (r relation[SrcBody, TargetFactory, TargetBody]) TargetFactory() IFactory {
	return r.targetFactory
}

func (r relation[SrcBody, TargetFactory, TargetBody]) LocalFactory() IFactory {
	return r.localKeys[0].Factory()
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) eagerLoader() IEagerLoader {
	return &EagerLoader[SrcBody, TargetFactory, TargetBody]{IRelation: r.relationPtr}
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) Eager(relation IRelation) IEagerLoader {
	localFactory := r.TargetFactory()
	targetFactory := relation.LocalFactory()
	if localFactory.Schema() != targetFactory.Schema() || localFactory.TableName() != targetFactory.TableName() {
		r.TargetFactory().Db().Error("Eagar relation must be in same factory", nil)
		return nil
	}
	rr := relation.clone()
	return r.eagerLoader().addChain(rr.eagerLoader())
}

func (r relation[SrcBody, TargetFactory, TargetBody]) LoadQuery(body SrcBody) *SelectQuery[TargetBody] {
	var (
		cond = make(squirrel.And, len(r.foreignKeys))
	)
	for i, c := range r.foreignKeys {
		cond[i] = c.Eq(r.localKeys[i].unwrapPtr(body.FieldValuePtr(r.localKeys[i].GetPos())))
	}
	return SelectByFactory[TargetBody](r.targetFactory, r.targetFactory.IModelScanner().(IModelScanner[TargetBody])).
		Where(append(cond, r.cond...))
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Load(body SrcBody) error {
	var (
		related any
		err     error
		q       = r.LoadQuery(body)
	)
	if r.isMany {
		var (
			relateds    []TargetBody
			relatedPtrs []*TargetBody
		)
		relateds, err = q.GetAll()
		if err != nil {
			return err
		}
		relatedPtrs = make([]*TargetBody, len(relateds))
		for i := range relateds {
			relatedPtrs[i] = &relateds[i]
		}
		related = relatedPtrs
	} else {
		related, err = q.GetOne()
	}
	if err != nil {
		return err
	}
	if r.isMany {
		*body.RelatedByPos(r.index).(*[]*TargetBody) = related.([]*TargetBody)
	} else {
		*body.RelatedByPos(r.index).(**TargetBody) = related.(*TargetBody)
	}
	return nil
}

func (r relation[SrcBody, TargetFactory, TargetBody]) LocalKeys() []IColFromFactory {
	return r.localKeys
}

func (r relation[SrcBody, TargetFactory, TargetBody]) ForeignKeys() []IColFromFactory {
	return r.foreignKeys
}

func (r relation[SrcBody, TargetFactory, TargetBody]) JoinClause() ForeignKeyJoin {
	return ForeignKeyJoin{r.foreignKeys, r.localKeys, r.cond}
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) getRelation() any {
	return r
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Columns(cols ...any) IEagerLoader {
	return r.eagerLoader().Columns(cols...)
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Where(conds ...ICond) IEagerLoader {
	return r.eagerLoader().Where(conds...)
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Limit(limit uint32) IEagerLoader {
	return r.eagerLoader().Limit(limit)
}

func (r relation[SrcBody, TargetFactory, TargetBody]) Offset(offset uint32) IEagerLoader {
	return r.eagerLoader().Offset(offset)
}

func (r relation[SrcBody, TargetFactory, TargetBody]) OrderBy(orderBy any) IEagerLoader {
	return r.eagerLoader().OrderBy(orderBy)
}

func (r relation[SrcBody, TargetFactory, TargetBody]) setForeignKey(srcBody any, targetBodies ...any) error {
	var (
		sBody = srcBody.(SrcBody)
		j     int
	)
	for _, targetBody := range targetBodies {
		for j = range r.foreignKeys {
			r.foreignKeys[j].SetValue(targetBody,
				r.localKeys[j].getValueFromBody(sBody))
		}
	}
	return nil
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) ExtraCond(cond ICond) *relation[SrcBody, TargetFactory, TargetBody] {
	if c, ok := cond.(colConditioner); ok {
		if c.col.QueryFactory().tableNameWithParent() == r.targetFactory.tableNameWithParent() {
			c.col = r.targetFactory.Columns()[c.col.GetPos()].(IColFromFactory)
			cond = c
		} else if c.col.QueryFactory().tableNameWithParent() == r.localKeys[0].Factory().tableNameWithParent() {
			c.col = r.localKeys[0].Factory().Columns()[c.col.GetPos()].(IColFromFactory)
		}
	}
	r.cond = append(r.cond, cond)
	return r
}

func (r *relation[SrcBody, TargetFactory, TargetBody]) TargetBody() TargetBody {
	var (
		body TargetBody
	)
	return body
}

type ToOne[SB iModelRelatedBody, TF IFactory, TB any] struct {
	relation[SB, TF, TB]
}

func (r *ToOne[SrcBody, TargetFactory, TargetBody]) InitRelation(localAndForeignKeyPairs ...IColFromFactory) IRelation {
	r.relation.InitRelation(r, localAndForeignKeyPairs...)
	r.isMany = false
	return r
}

func (r *ToOne[SrcBody, TargetFactory, TargetBody]) ExtraCond(cond ICond) IRelation {
	r.relation.ExtraCond(cond)
	return r
}

func (r ToOne[SrcBody, TargetFactory, TargetBody]) clone(targetFactory ...IFactory) IRelation {
	rr := &ToOne[SrcBody, TargetFactory, TargetBody]{
		relation: r.relation,
	}
	if len(targetFactory) != 0 && targetFactory[0].Alias() != r.name {
		rr.targetFactory = targetFactory[0].SetAlias(r.name).(TargetFactory)
	}
	rr.relationPtr = rr
	return rr
}

type ToMany[SB iModelRelatedBody, TF IFactory, TB any] struct {
	relation[SB, TF, TB]
}

func (r *ToMany[SrcBody, TargetFactory, TargetBody]) InitRelation(localAndForeignKeyPairs ...IColFromFactory) IRelation {
	r.relation.InitRelation(r, localAndForeignKeyPairs...)
	r.isMany = true
	return r
}

func (r ToMany[SrcBody, TargetFactory, TargetBody]) clone(targetFactory ...IFactory) IRelation {
	rr := &ToMany[SrcBody, TargetFactory, TargetBody]{
		relation: r.relation,
	}
	if len(targetFactory) != 0 && targetFactory[0].Alias() != r.name {
		rr.targetFactory = targetFactory[0].SetAlias(r.name).(TargetFactory)
	}
	rr.relationPtr = rr
	return rr
}

func (r *ToMany[SrcBody, TargetFactory, TargetBody]) ExtraCond(cond ICond) IRelation {
	r.relation.ExtraCond(cond)
	return r
}

type EagerLoader[SB iModelRelatedBody, TF IFactory, TB any] struct {
	IRelation
	eagerChain []IEagerLoader
	columns    []any
	conds      []ICond
	limit      uint32
	offset     uint32
	orderBy    any
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) eagerLoader() IEagerLoader {
	return loader
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) Eager(relation IRelation) IEagerLoader {
	rr := relation.clone()
	return loader.addChain(rr.eagerLoader())
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) addChain(loaders ...IEagerLoader) IEagerLoader {
	loader.eagerChain = append(loader.eagerChain, loaders...)
	return loader
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) Columns(cols ...any) IEagerLoader {
	var (
		ok   bool
		icol ICol
	)
	for _, col := range cols {
		if icol, ok = col.(ICol); ok && icol.QueryFactory() != loader.TargetFactory() {
			loader.TargetFactory().Db().Error("Eagar relation must be in same factory", nil)
			return nil
		}
	}
	loader.columns = append(loader.columns, cols...)
	return loader
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) Where(conds ...ICond) IEagerLoader {
	loader.conds = append(loader.conds, conds...)
	return loader
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) Limit(limit uint32) IEagerLoader {
	loader.limit = limit
	return loader
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) Offset(offset uint32) IEagerLoader {
	loader.offset = offset
	return loader
}

func (loader *EagerLoader[SrcBody, TargetFactory, TargetBody]) OrderBy(orderBy any) IEagerLoader {
	loader.orderBy = orderBy
	return loader
}
