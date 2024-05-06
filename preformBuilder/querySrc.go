package preformBuilder

import (
	preformShare "github.com/go-preform/preform/share"
)

type iQueryBuilderSrc interface {
	toSrc() *queryBuilderSrc
}

type QueryBuilderSrc[T preformShare.IFactoryBuilder] struct {
	queryBuilderSrc
}

func NewQueryBuilderSrc[T preformShare.IFactoryBuilder](model T) *QueryBuilderSrc[T] {
	return &QueryBuilderSrc[T]{queryBuilderSrc{src: model}}
}

func (v *QueryBuilderSrc[T]) toSrc() *queryBuilderSrc {
	return &v.queryBuilderSrc
}

func (v *QueryBuilderSrc[T]) JoinCond(cond preformShare.ICondForBuilder) *QueryBuilderSrc[T] {
	v.joinCond = cond
	return v
}

type queryBuilderSrc struct {
	src           preformShare.IFactoryBuilder
	joinDirection string
	joinCond      preformShare.ICondForBuilder
}
