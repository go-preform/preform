package preform

type iAggregation interface {
}

type NoAggregation struct{}

type PrebuildQueryCol[T any, A iAggregation] struct {
	ICol
	query       IQuery
	pos         int
	isAggregate bool
}

func SetPrebuildQueryCol[T any, A iAggregation](f iPrebuildQueryFactory, src ICol, c *PrebuildQueryCol[T, A]) *PrebuildQueryCol[T, A] {
	*c = PrebuildQueryCol[T, A]{ICol: src, query: f}
	c.pos = f.addCol(c)
	_, c.isAggregate = src.(iAggregateCol)
	return c
}

func (c PrebuildQueryCol[T, A]) GetCodeWithAlias() string {
	if c.isAggregate {
		return sqlizerToString(c.ICol)
	} else {
		return c.ICol.GetCodeWithAlias()
	}
}

func (c PrebuildQueryCol[T, A]) QueryFactory() IQuery {
	return c.query
}

func (c PrebuildQueryCol[T, A]) GetPos() int {
	return c.pos
}

func (c PrebuildQueryCol[T, A]) Factory() IFactory {
	return nil //just to fulfill interface
}
