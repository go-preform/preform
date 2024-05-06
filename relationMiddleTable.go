package preform

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-preform/squirrel"
	"strings"
)

type MiddleTable[SB iModelRelatedBody, TF IFactory, TB any, MTB any] struct {
	relation[SB, TF, TB]
	middleTable                  IFactory
	localKeyRefs, foreignKeyRefs []IColFromFactory
}

func (r *MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) InitMtRelation(middleTable IFactory, localKeys, localKeyRefs, foreignKey, foreignKeyRefs []IColFromFactory) IRelation {
	r.targetFactory = foreignKey[0].Factory().SetAlias(r.name).(TargetFactory)
	r.middleTable = middleTable
	r.foreignKeys = foreignKey
	r.localKeys = localKeys
	r.localKeyRefs = localKeyRefs
	r.foreignKeyRefs = foreignKeyRefs
	r.relationPtr = r
	r.isMany = true
	return r
}

// deprecated use InitMtRelation
func (r *MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) InitRelation(localAndForeignKeyPairs ...IColFromFactory) IRelation {
	return r
}

func (r *MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) ExtraCond(cond ICond) IRelation {
	r.relation.ExtraCond(cond)
	return r
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) clone(targetFactory ...IFactory) IRelation {
	rr := &MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]{
		relation:       r.relation,
		foreignKeyRefs: r.foreignKeyRefs,
		localKeyRefs:   r.localKeyRefs,
		middleTable:    r.middleTable,
	}
	if len(targetFactory) != 0 && targetFactory[0].Alias() != r.name {
		rr.targetFactory = targetFactory[0].SetAlias(r.name).(TargetFactory)
	}
	rr.relationPtr = rr
	return rr
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) loadQuery(cols []ICol) *SelectQuery[modelWithRelation[TargetBody, MiddleBody]] {
	var (
		onClause = make([]string, len(r.foreignKeys))
		colAny   = make([]any, len(cols), len(cols)+len(r.localKeys))
		scanner  = modelScannerWithRelated[TargetBody, MiddleBody]{refKeyPos: make([]int, 0, len(r.foreignKeyRefs))}
	)
	for i, c := range r.foreignKeys {
		onClause[i] = fmt.Sprintf("%s.%s = %s.%s", r.middleTable.Alias(), r.foreignKeyRefs[i].DbName(), r.targetFactory.Alias(), c.DbName())
	}
	for i, c := range cols {
		colAny[i] = c
	}
	for _, c := range r.localKeyRefs {
		colAny = append(colAny, c)
		scanner.refKeyPos = append(scanner.refKeyPos, c.GetPos())
	}
	return SelectByFactory[modelWithRelation[TargetBody, MiddleBody]](r.targetFactory, scanner, colAny...).From(r.middleTable).
		Join(fmt.Sprintf("%s ON %s", r.targetFactory.fromClause(), strings.Join(onClause, " AND ")))
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) LoadQuery(body SrcBody) *SelectQuery[TargetBody] {
	var (
		cond     = make(squirrel.And, len(r.foreignKeys))
		onClause = make([]string, len(r.foreignKeys))
		cols     = r.targetFactory.Columns()
		colAny   = make([]any, len(cols))
	)
	for i, c := range r.localKeyRefs {
		cond[i] = c.Eq(r.localKeys[i].
			unwrapPtr(body.
				FieldValuePtr(r.localKeys[i].GetPos())))
	}
	for i, c := range r.foreignKeys {
		onClause[i] = fmt.Sprintf("%s.%s = %s.%s", r.middleTable.Alias(), r.foreignKeyRefs[i].DbName(), r.targetFactory.Alias(), c.DbName())
	}
	for i, c := range cols {
		colAny[i] = c
	}
	return SelectByFactory[TargetBody](r.targetFactory, r.targetFactory.IModelScanner().(IModelScanner[TargetBody]), colAny...).From(r.middleTable).
		Join(fmt.Sprintf("%s ON %s", r.targetFactory.fromClause(), strings.Join(onClause, " AND "))).
		Where(append(cond, r.cond...))
}

func (r *MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) eagerLoader() IEagerLoader {
	return &EagerLoaderMt[SrcBody, TargetFactory, TargetBody, MiddleBody]{&EagerLoader[SrcBody, TargetFactory, TargetBody]{IRelation: r}}
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) Eager(relation IRelation) IEagerLoader {
	localFactory := r.TargetFactory()
	targetFactory := relation.LocalFactory()
	if localFactory.Schema() != targetFactory.Schema() || localFactory.TableName() != targetFactory.TableName() {
		r.TargetFactory().Db().Error("Eagar relation must be in same factory", nil)
		return nil
	}
	rr := relation.clone()
	return r.eagerLoader().addChain(rr.eagerLoader())
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) Load(body SrcBody) error {
	var (
		related     any
		err         error
		q           = r.LoadQuery(body)
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

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) JoinClause() ForeignKeyJoin {
	return ForeignKeyJoin{r.foreignKeys, r.localKeys, r.cond}
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) setForeignKey(srcBody any, targetBodies ...any) error {
	if len(targetBodies) != 0 {
		var (
			sBody        = srcBody.(SrcBody)
			middleBodies = make([]MiddleBody, len(targetBodies))
			j            int
		)
		for i, targetBody := range targetBodies {
			for j = range r.localKeyRefs {
				r.localKeyRefs[j].SetValue(&middleBodies[i], r.localKeys[j].getValueFromBody(sBody))
			}
			for j = range r.foreignKeyRefs {
				r.foreignKeyRefs[j].SetValue(&middleBodies[i], r.foreignKeys[j].getValueFromBody(targetBody.(iModelBody)))
			}
		}
		return r.LinkModels(middleBodies...)
	}
	return nil
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) LinkModels(bodies ...MiddleBody) error {
	return r.middleTable.Insert(bodies)
}

func (r MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody]) IsMiddleTable() bool {
	return true
}

type EagerLoaderMt[SB iModelRelatedBody, TF IFactory, TB, MTB any] struct {
	*EagerLoader[SB, TF, TB]
}

func (loader EagerLoaderMt[SrcBody, TargetFactory, TargetBody, MiddleBody]) getRelation() any {
	return loader.IRelation
}

func (loader EagerLoaderMt[SrcBody, TargetFactory, TargetBody, MiddleBody]) loadEagers(ctx context.Context, srcModels any, useFast bool) error {
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

func (loader EagerLoaderMt[SrcBody, TargetFactory, TargetBody, MiddleBody]) loadEager(ctx context.Context, srcModels any, useFast bool) (relatedModels any, err error) {
	if bodies, ok := srcModels.([]SrcBody); ok {
		if len(bodies) == 0 {
			return
		}
		var (
			r                = loader.getRelation().(*MiddleTable[SrcBody, TargetFactory, TargetBody, MiddleBody])
			toCompare        = make([][]func(any) bool, len(bodies))
			thisValue        any
			relatedBodies    []modelWithRelation[TargetBody, MiddleBody]
			relatedBodiesPtr []*TargetBody
			lkl              = len(r.localKeys)
			toCond           []any
			cond             = make(squirrel.And, lkl)
			i, j, k          int
			q                *SelectQuery[modelWithRelation[TargetBody, MiddleBody]]
		)
		for i = range r.localKeys {
			toCond = nil
			for j = range bodies {
				if i == 0 {
					toCompare[j] = make([]func(any) bool, lkl)
				}
				thisValue = r.localKeys[i].getValueFromBody(bodies[j])
				toCond = append(toCond, thisValue)
				toCompare[j][i] = joinCompareEq(thisValue)
			}
			cond[i] = r.localKeyRefs[i].Eq(toCond)
		}
		if len(r.cond) != 0 {
			cond = append(r.cond, cond...)
		}

		if len(loader.columns) != 0 {
			cols := make([]ICol, len(loader.columns))
			for i, col := range loader.columns {
				cols[i] = col.(ICol)
			}
			q = r.loadQuery(cols).Where(cond)
		} else {
			q = r.loadQuery(r.targetFactory.Columns()).Where(cond)
		}
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

		if useFast {
			relatedBodies, err = q.Ctx(ctx).GetAllFast()
		} else {
			relatedBodies, err = q.Ctx(ctx).GetAll()
		}
		if err != nil {
			return
		}
		relatedBodiesPtr = make([]*TargetBody, len(relatedBodies))
		var (
			ptrs []*TargetBody
			mtb  *MiddleBody
		)
		for k = range toCompare {
			ptrs = nil
		loopCompares:
			for i = range relatedBodies {
				mtb = &relatedBodies[i].Related
				relatedBodiesPtr[i] = &relatedBodies[i].Body
				for j = range r.localKeyRefs {
					if !toCompare[k][j](r.localKeyRefs[j].getValueFromBody(any(mtb).(iModelBody))) {
						continue loopCompares
					}
				}
				ptrs = append(ptrs, relatedBodiesPtr[i])
			}
			if len(ptrs) != 0 {
				*bodies[k].RelatedByPos(r.index).(*[]*TargetBody) = ptrs
			}
		}
		return relatedBodiesPtr, nil
	}
	err = errors.New("srcModels is not []SrcBody")
	return
}

type modelScannerWithRelated[B, RELATED any] struct {
	IModelScanner[B]
	refKeyPos []int
}

type modelWithRelation[B, RELATED any] struct {
	Body    B
	Related RELATED
}

func (b modelWithRelation[B, RELATED]) ForceBodyScan() {}

func (b modelScannerWithRelated[B, RELATED]) ScanBodies(rows IRows, cols []ICol, max uint64) ([]modelWithRelation[B, RELATED], error) {
	var (
		model       modelWithRelation[B, RELATED]
		res         = make([]modelWithRelation[B, RELATED], 0, max)
		sortedPtrs  []any
		scanPtrs    []any
		mtbScanPtrs []any
		err         error
		l           = len(cols)
	)
	scanPtrs = any(&model.Body).(hasFieldValuePtrs).FieldValuePtrs()
	mtbScanPtrs = any(&model.Related).(hasFieldValuePtrs).FieldValuePtrs()
	sortedPtrs = make([]any, 0, l)
	for _, col := range cols[:l-len(b.refKeyPos)] {
		sortedPtrs = append(sortedPtrs, scanPtrs[col.GetPos()])
	}
	for _, pos := range b.refKeyPos {
		sortedPtrs = append(sortedPtrs, mtbScanPtrs[pos])
	}
	for rows.Next() {
		err = rows.Scan(sortedPtrs...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		res = append(res, model)
	}
	return res, err
}

func (b modelScannerWithRelated[B, RELATED]) ScanCount(rows IRows) (uint64, error) {
	return 0, nil
}
func (b modelScannerWithRelated[B, RELATED]) ScanAny(rows IRows, cols []ICol, max uint64) ([]any, error) {
	return nil, nil
}
func (b modelScannerWithRelated[B, RELATED]) ScanBodiesFast(rows IRows, cols []ICol, max uint64) ([]modelWithRelation[B, RELATED], error) {
	var (
		model       modelWithRelation[B, RELATED]
		res         = make([]modelWithRelation[B, RELATED], 0, max)
		scanPtrs    []any
		sortedPtrs  []any
		mtbScanPtrs []any
		err         error
		l           = len(cols)
		ll          = l - len(b.refKeyPos)
	)
	scanPtrs = any(&model.Body).(hasFieldValuePtrs).FieldValuePtrs()
	mtbScanPtrs = any(&model.Related).(hasFieldValuePtrs).FieldValuePtrs()
	sortedPtrs = make([]any, 0, l)
	for _, col := range cols[:ll] {
		sortedPtrs = append(sortedPtrs, col.wrapScanner(scanPtrs[col.GetPos()]))
	}
	for i, pos := range b.refKeyPos {
		sortedPtrs = append(sortedPtrs, cols[ll+i].wrapScanner(mtbScanPtrs[pos]))
	}
	for rows.Next() {
		err = rows.Scan(sortedPtrs...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		res = append(res, model)
	}
	return res, err

}
func (b modelScannerWithRelated[B, RELATED]) ScanBody(row IRow, cols []ICol) (*modelWithRelation[B, RELATED], error) {
	return nil, nil
}
func (b modelScannerWithRelated[B, RELATED]) ScanRaw(rows IRows, colValTpl []func() (any, func(*any) any), max uint64) (*RowsWithCols, error) {
	return nil, nil
}
func (b modelScannerWithRelated[B, RELATED]) ScanStructs(rows IRows, s any) error {
	return nil
}
func (b modelScannerWithRelated[B, RELATED]) ToAnyScanner() IModelScanner[any] {
	return nil
}
