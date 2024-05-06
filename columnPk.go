package preform

import "reflect"

type PrimaryKey[T any] struct {
	*ForeignKey[T]
}

func (c *PrimaryKey[T]) initCol(ref reflect.StructField, dbName string, factory IFactory, pos int) {
	c.ForeignKey = &ForeignKey[T]{}
	c.ForeignKey.initCol(ref, dbName, factory, pos)
	c.isPk = true
}

func (c PrimaryKey[T]) clone(f IFactory) ICol {
	var (
		cc   = c
		cPtr = &cc
	)
	cc.ForeignKey = cc.ForeignKey.clone(f).(*ForeignKey[T])
	return cPtr
}
