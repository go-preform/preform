package preform

import (
	"github.com/go-preform/squirrel"
	"strings"
)

type iJoiner[SB iModelRelatedBody, TF IFactory, TB any] interface {
	join(relatedBodies []TB) []*TB
}

type bfJoiner[SB iModelRelatedBody, TF IFactory, TB any] struct {
	relation  *relation[SB, TF, TB]
	bodies    []SB
	toCompare [][]func(any) bool
}

func (joiner bfJoiner[SrcBody, TargetFactory, TargetBody]) join(relatedBodies []TargetBody) []*TargetBody {
	var (
		i, j, k int
		fk      = joiner.relation.foreignKeys
		tc      = joiner.toCompare
	)
	relatedBodiesPtr := make([]iModelBody, len(relatedBodies))
	res := make([]*TargetBody, len(relatedBodies))
	if joiner.relation.isMany {
		var (
			ptrs []*TargetBody
		)
		for k = range tc {
			ptrs = nil
		loopCompares:
			for i = range relatedBodies {
				if k == 0 {
					res[i] = &relatedBodies[i]
					relatedBodiesPtr[i] = any(res[i]).(iModelBody)
				}
				for j = range fk {
					if !tc[k][j](fk[j].getValueFromBody(relatedBodiesPtr[i])) {
						continue loopCompares
					}
				}
				ptrs = append(ptrs, res[i])
			}
			if len(ptrs) != 0 {
				*joiner.bodies[k].RelatedByPos(joiner.relation.index).(*[]*TargetBody) = ptrs
			}
		}
	} else {
		for k = range tc {
		loopCompare:
			for i = range relatedBodies {
				if k == 0 {
					res[i] = &relatedBodies[i]
					relatedBodiesPtr[i] = any(res[i]).(iModelBody)
				}
				for j = range fk {
					if !tc[k][j](fk[j].getValueFromBody(any(relatedBodiesPtr[i]).(iModelBody))) {
						continue loopCompare
					}
				}
				*joiner.bodies[k].RelatedByPos(joiner.relation.index).(**TargetBody) = res[i]
			}
		}
	}
	return res
}

func prepareBFCompare[SrcBody iModelRelatedBody, TargetFactory IFactory, TargetBody any](r *relation[SrcBody, TargetFactory, TargetBody], bodies []SrcBody) (iJoiner[SrcBody, TargetFactory, TargetBody], squirrel.And) {
	var (
		toCompare            = make([][]func(any) bool, len(bodies))
		thisValues           any
		lkl                  = len(r.localKeys)
		toCond               []any
		cond                 = make(squirrel.And, lkl)
		lkIsArray, fkIsArray = make([]bool, lkl), make([]bool, lkl)
		i, j, k              int
	)
	for i = range r.localKeys {
		toCond = nil
		lkIsArray[i], _, _, _ = r.localKeys[i].properties()
		fkIsArray[i], _, _, _ = r.foreignKeys[i].properties()
		for j = range bodies {
			if i == 0 {
				toCompare[j] = make([]func(any) bool, lkl)
			}
			thisValues = r.localKeys[i].getValueFromBody(bodies[j])
			if lkIsArray[i] {
				if fkIsArray[i] {
					toCompare[j][i] = joinCompareArrayEq(thisValues.([]any))
					toCond = append(toCond, thisValues)
				} else {
					toCompare[j][i] = joinCompareArrayHasAny(thisValues.([]any))
					toCond = append(toCond, r.localKeys[i].flatten(thisValues)...)
				}
			} else {
				toCompare[j][i] = joinCompareEq(thisValues)
				toCond = append(toCond, thisValues)
			}
		}
		if lkIsArray[i] && fkIsArray[i] {
			or := make(squirrel.Or, len(toCond))
			for k = range toCond {
				or = append(or, r.foreignKeys[i].Eq(toCond[k]))
			}
			cond[i] = or
		} else if lkIsArray[i] {
			cond[i] = r.foreignKeys[i].HasAny(toCond)
		} else {
			cond[i] = r.foreignKeys[i].Eq(toCond)
		}
	}
	return bfJoiner[SrcBody, TargetFactory, TargetBody]{relation: r, bodies: bodies, toCompare: toCompare}, cond
}

func joinCompareArrayHasAny(arr []any) func(v any) bool {
	return func(v any) bool {
		for _, a := range arr {
			if a == v {
				return true
			}
		}
		return false
	}
}

func joinCompareArrayEq(arrA []any) func(v any) bool {
	return func(v any) bool {
		var (
			arrB = v.([]any)
		)
		if len(arrA) != len(arrB) {
			return false
		}
		for i := range arrB {
			if arrA[i] != arrB[i] {
				return false
			}
		}
		return true
	}
}

func joinCompareEq(arr any) func(v any) bool {
	return func(v any) bool {
		return arr == v
	}
}

type hashJoiner[SB iModelRelatedBody, TF IFactory, TB any] struct {
	relation  *relation[SB, TF, TB]
	bodies    []SB
	toCompare map[string]*hashJoinerNode[SB, TB]
}

type hashJoinerNode[SB iModelRelatedBody, TB any] struct {
	parents []SB
}

func prepareHashCompare[SrcBody iModelRelatedBody, TargetFactory IFactory, TargetBody any](r *relation[SrcBody, TargetFactory, TargetBody], bodies []SrcBody) (iJoiner[SrcBody, TargetFactory, TargetBody], squirrel.And) {
	var (
		toCompare  map[string]*hashJoinerNode[SrcBody, TargetBody]
		bl         = len(bodies)
		lk         = r.localKeys
		lkl        = len(lk)
		cond       = make(squirrel.And, lkl)
		i, j       int
		ok         bool
		joinerNode *hashJoinerNode[SrcBody, TargetBody]
	)
	if lkl == 1 {
		_, _, ok, _ = lk[0].properties()
		ok = ok && len(lk[0].Factory().Pks()) == 1
	}
	if ok {
		colValues := make([]any, bl)
		toCompare = make(map[string]*hashJoinerNode[SrcBody, TargetBody], bl)
		for j = range bodies {
			colValues[j] = lk[0].getValueFromBody(bodies[j])
			joinerNode = &hashJoinerNode[SrcBody, TargetBody]{parents: []SrcBody{bodies[j]}}
			toCompare[lk[0].valueToString(colValues[j])] = joinerNode
		}
		cond[0] = r.foreignKeys[i].Eq(colValues)
	} else {
		var (
			thisValues any
			key        string
			bodyKey    = make([]string, lkl)
			colValues  = make([][]any, lkl)
		)
		toCompare = make(map[string]*hashJoinerNode[SrcBody, TargetBody])
		for j = range bodies {
			for i = range lk {
				thisValues = lk[i].getValueFromBody(bodies[j])
				if j == 0 {
					colValues[i] = make([]any, bl)
				}
				colValues[i][j] = thisValues
				bodyKey[i] = lk[i].valueToString(thisValues)
			}
			key = strings.Join(bodyKey, ")(")
			if joinerNode, ok = toCompare[key]; !ok {
				joinerNode = &hashJoinerNode[SrcBody, TargetBody]{parents: []SrcBody{bodies[j]}}
				toCompare[key] = joinerNode
			} else {
				joinerNode.parents = append(joinerNode.parents, bodies[j])
			}
		}
		for i = range lk {
			cond[i] = r.foreignKeys[i].Eq(colValues[i])
		}
	}
	return hashJoiner[SrcBody, TargetFactory, TargetBody]{relation: r, bodies: bodies, toCompare: toCompare}, cond
}

func (joiner hashJoiner[SrcBody, TargetFactory, TargetBody]) join(relatedBodies []TargetBody) []*TargetBody {
	var (
		i, j       int
		fk         = joiner.relation.foreignKeys
		fkl        = len(fk)
		bodyKey    = make([]string, fkl)
		key        string
		joinerNode *hashJoinerNode[SrcBody, TargetBody]
		ok         bool
		parent     SrcBody
	)
	relatedBodiesPtr := make([]*TargetBody, len(relatedBodies))
	if fkl == 1 {
		if joiner.relation.isMany {
			var children *[]*TargetBody
			for i = range relatedBodies {
				relatedBodiesPtr[i] = &relatedBodies[i]
				if joinerNode, ok = joiner.toCompare[fk[0].valueToString(fk[0].getValueFromBody(any(relatedBodiesPtr[i]).(iModelBody)))]; ok {
					for _, parent = range joinerNode.parents {
						children = parent.RelatedByPos(joiner.relation.index).(*[]*TargetBody)
						*children = append(*children, relatedBodiesPtr[i])
					}
				}
			}
		} else {
			for i = range relatedBodies {
				relatedBodiesPtr[i] = &relatedBodies[i]
				if joinerNode, ok = joiner.toCompare[fk[0].valueToString(fk[0].getValueFromBody(any(relatedBodiesPtr[i]).(iModelBody)))]; ok {
					for _, parent = range joinerNode.parents {
						*parent.RelatedByPos(joiner.relation.index).(**TargetBody) = relatedBodiesPtr[i]
					}
				}
			}
		}
	} else {
		if joiner.relation.isMany {
			var children *[]*TargetBody
			for i = range relatedBodies {
				relatedBodiesPtr[i] = &relatedBodies[i]
				for j = range fk {
					bodyKey[j] = fk[j].valueToString(fk[j].getValueFromBody(any(relatedBodiesPtr[i]).(iModelBody)))
				}
				key = strings.Join(bodyKey, ")(")
				if joinerNode, ok = joiner.toCompare[key]; ok {
					children = parent.RelatedByPos(joiner.relation.index).(*[]*TargetBody)
					*children = append(*children, relatedBodiesPtr[i])
				}
			}
		} else {
			for i = range relatedBodies {
				relatedBodiesPtr[i] = &relatedBodies[i]
				for j = range fk {
					bodyKey[j] = fk[j].valueToString(fk[j].getValueFromBody(any(relatedBodiesPtr[i]).(iModelBody)))
				}
				key = strings.Join(bodyKey, ")(")
				if joinerNode, ok = joiner.toCompare[key]; ok {
					*parent.RelatedByPos(joiner.relation.index).(**TargetBody) = relatedBodiesPtr[i]
				}
			}
		}
	}
	return relatedBodiesPtr
}
