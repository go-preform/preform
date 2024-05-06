package preform

import preformShare "github.com/go-preform/preform/share"

type hasQueryFactory[F IQuery] interface {
	Factory() F
}
type hasFactory[F IFactory] interface {
	Factory() F
}

type iModelBodyReadOnly interface {
	FieldValueImmutablePtrs() []any
}

type iModelBody interface {
	iModelBodyReadOnly
	FieldValuePtrs() []interface{}
	FieldValuePtr(pos int) interface{}
	setFactory(defaultFactory IFactory)
}

type QueryBody[T hasQueryFactory[F], F IQuery] struct{}

func (b QueryBody[T, F]) getQueryFactory() IQuery {
	var (
		bb T
	)
	return bb.Factory()
}

type Body[T hasFactory[F], F IFactory] struct {
	QueryBody[T, F]
	hasFactory bool
	factory    F
}

func (b Body[T, F]) getFactory() IFactory {
	var (
		bb T
	)
	return bb.Factory()
}

func (b *Body[T, F]) setFactory(defaultFactory IFactory) {
	b.factory = defaultFactory.(F)
	b.hasFactory = true
}

func (b Body[T, F]) Factory(defaultFactory F) F {
	if !b.hasFactory {
		return defaultFactory
	}
	return b.factory
}

func (b Body[T, F]) SetCol(body *T, col preformShare.ICol, value any) {
	any(col).(ICol).setValueToBody(any(body).(iModelBody), value)
}

func (b Body[T, F]) Insert(body *T, cfg ...EditConfig) error {
	return (*body).Factory().Insert(body, cfg...)
}

func (b Body[T, F]) UpdateByPk(body *T, cfg ...UpdateConfig) (affected int64, err error) {
	return (*body).Factory().UpdateByPk(body, cfg...)
}

func (b Body[T, F]) Delete(body *T, cfg ...EditConfig) (affected int64, err error) {
	return 0, nil
}
