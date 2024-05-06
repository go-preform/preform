package mainModel

import (
	"github.com/go-preform/preform"
)

var fooInit = preform.InitFactory[*FactoryFoo, FooBody](func(s *PreformTestASchema) {
	preform.SetColumn(s.Foo.Id.Column).AutoIncrement()
	s.Foo.SetTableName("foo")
})

type FactoryFoo struct {
	preform.Factory[*FactoryFoo, FooBody]
	Id *preform.PrimaryKey[int32] `db:"id" json:"Id" dataType:"int4" autoKey:"true"`
	Fk1 *preform.Column[int32] `db:"fk1" json:"Fk1" dataType:"int4"`
	Fk2 *preform.Column[int32] `db:"fk2" json:"Fk2" dataType:"int4"`
	Bars *preform.ToMany[*FooBody, *FactoryBar, BarBody]
}

func (f FactoryFoo) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryFoo, FooBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int32] )
	ff.Fk1 = cols[1].(*preform.Column[int32] )
	ff.Fk2 = cols[2].(*preform.Column[int32] )
	return ff.Factory.Definition
}


type FooBody struct {
	preform.Body[FooBody,*FactoryFoo]
	Id int32 `db:"id" json:"Id" dataType:"int4" autoKey:"true"`
	Fk1 int32 `db:"fk1" json:"Fk1" dataType:"int4"`
	Fk2 int32 `db:"fk2" json:"Fk2" dataType:"int4"`
	Bars []*BarBody
}

func (m FooBody) Factory() *FactoryFoo { return m.Body.Factory(PreformTestA.Foo) }

func (m *FooBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.Foo.Insert(m, cfg...) }

func (m *FooBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.Foo.UpdateByPk(m, cfg...) }

func (m *FooBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.Foo.DeleteByPk(m, cfg...) }

func (m FooBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.Fk1, &m.Fk2} }

func (m *FooBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.Fk1
		case 2: return &m.Fk2
	}
	return nil
}

func (m *FooBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.Fk1, &m.Fk2}
}

func (m *FooBody) RelatedValuePtrs() []any { return []any{&m.Bars} }


func (m *FooBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.Bars
	}
	return nil
}


func (m *FooBody) LoadBars(noCache ...bool) ([]*BarBody, error) {
	if len(m.Bars) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.Foo.Bars.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.Bars, nil
}

