package mainModel

import (
	"github.com/go-preform/preform"
)

var barInit = preform.InitFactory[*FactoryBar, BarBody](func(s *PreformTestBSchema, preformTestA *PreformTestASchema) {
	s.Bar.Foo.InitRelation(s.Bar.Id1, preformTestA.Foo.Fk1, s.Bar.Id2, preformTestA.Foo.Fk2)
	preformTestA.Foo.Bars.InitRelation(preformTestA.Foo.Fk1, s.Bar.Id1, preformTestA.Foo.Fk2, s.Bar.Id2)
	s.Bar.SetTableName("bar")
})

type FactoryBar struct {
	preform.Factory[*FactoryBar, BarBody]
	Id1 *preform.PrimaryKey[int32] `db:"id1" json:"Id1" dataType:"int"`
	Id2 *preform.PrimaryKey[int32] `db:"id2" json:"Id2" dataType:"int"`
	
	//relations
	Foo *preform.ToOne[*BarBody, *FactoryFoo, FooBody]
}

func (f FactoryBar) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryBar, BarBody])
	ff.Factory.Definition = &ff
	ff.Id1 = cols[0].(*preform.PrimaryKey[int32] )
	ff.Id2 = cols[1].(*preform.PrimaryKey[int32] )
	return ff.Factory.Definition
}


type BarBody struct {
	preform.Body[BarBody,*FactoryBar]
	Id1 int32 `db:"id1" json:"Id1" dataType:"int"`
	Id2 int32 `db:"id2" json:"Id2" dataType:"int"`
	
	Foo *FooBody
}

func (m BarBody) Factory() *FactoryBar { return m.Body.Factory(PreformTestB.Bar) }

func (m *BarBody) Insert(cfg ... preform.EditConfig) error { return PreformTestB.Bar.Insert(m, cfg...) }

func (m *BarBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestB.Bar.UpdateByPk(m, cfg...) }

func (m *BarBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestB.Bar.DeleteByPk(m, cfg...) }

func (m BarBody) FieldValueImmutablePtrs() []any { return []any{&m.Id1, &m.Id2} }

func (m *BarBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id1
		case 1: return &m.Id2
	}
	return nil
}

func (m *BarBody) FieldValuePtrs() []any { 
	return []any{&m.Id1, &m.Id2}
}

func (m *BarBody) RelatedValuePtrs() []any { return []any{&m.Foo} }


func (m *BarBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.Foo
	}
	return nil
}


func (m *BarBody) LoadFoo(noCache ...bool) (*FooBody, error) {
	if m.Foo == nil || len(noCache) != 0 && noCache[0] {
		err := PreformTestB.Bar.Foo.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.Foo, nil
}

