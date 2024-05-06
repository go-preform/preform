package mainModel

import (
	"github.com/go-preform/preform"
)

var fooInit = preform.InitFactory[*FactoryFoo, FooBody](func(s *PreformTestASchema, preformTestB *PreformTestBSchema) {
	preform.SetColumn(s.Foo.Id.Column).AutoIncrement()
	s.Foo.Bar.InitRelation(s.Foo.Fk, preformTestB.Bar.Id)
	preformTestB.Bar.Foos.InitRelation(preformTestB.Bar.Id, s.Foo.Fk)
	s.Foo.SetTableName("foo")
})

type FactoryFoo struct {
	preform.Factory[*FactoryFoo, FooBody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	Fk *preform.ForeignKey[int64] `db:"fk" json:"Fk" dataType:"INTEGER" comment:"fk:preform_test_b.bar.id"`
	
	//relations
	Bar *preform.ToOne[*FooBody, *FactoryBar, BarBody]
}

func (f FactoryFoo) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryFoo, FooBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	ff.Fk = cols[1].(*preform.ForeignKey[int64] )
	return ff.Factory.Definition
}


type FooBody struct {
	preform.Body[FooBody,*FactoryFoo]
	Id int64 `db:"id" json:"Id" dataType:"INTEGER" autoKey:"true"`
	Fk int64 `db:"fk" json:"Fk" dataType:"INTEGER" comment:"fk:preform_test_b.bar.id"`
	
	Bar *BarBody
}

func (m FooBody) Factory() *FactoryFoo { return m.Body.Factory(PreformTestA.Foo) }

func (m *FooBody) Insert(cfg ... preform.EditConfig) error { return PreformTestA.Foo.Insert(m, cfg...) }

func (m *FooBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestA.Foo.UpdateByPk(m, cfg...) }

func (m *FooBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestA.Foo.DeleteByPk(m, cfg...) }

func (m FooBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.Fk} }

func (m *FooBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.Fk
	}
	return nil
}

func (m *FooBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.Fk}
}

func (m *FooBody) RelatedValuePtrs() []any { return []any{&m.Bar} }


func (m *FooBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.Bar
	}
	return nil
}


func (m *FooBody) LoadBar(noCache ...bool) (*BarBody, error) {
	if m.Bar == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestA.Foo.Bar.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.Bar, nil
}

