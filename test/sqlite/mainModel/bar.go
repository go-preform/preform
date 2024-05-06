package mainModel

import (
	"github.com/go-preform/preform"
)

var barInit = preform.InitFactory[*FactoryBar, BarBody](func(s *PreformTestBSchema) {
	s.Bar.SetTableName("bar")
})

type FactoryBar struct {
	preform.Factory[*FactoryBar, BarBody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"INTEGER"`
	Foos *preform.ToMany[*BarBody, *FactoryFoo, FooBody]
}

func (f FactoryBar) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryBar, BarBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	return ff.Factory.Definition
}


type BarBody struct {
	preform.Body[BarBody,*FactoryBar]
	Id int64 `db:"id" json:"Id" dataType:"INTEGER"`
	Foos []*FooBody
}

func (m BarBody) Factory() *FactoryBar { return m.Body.Factory(PreformTestB.Bar) }

func (m *BarBody) Insert(cfg ... preform.EditConfig) error { return PreformTestB.Bar.Insert(m, cfg...) }

func (m *BarBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformTestB.Bar.UpdateByPk(m, cfg...) }

func (m *BarBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformTestB.Bar.DeleteByPk(m, cfg...) }

func (m BarBody) FieldValueImmutablePtrs() []any { return []any{&m.Id} }

func (m *BarBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
	}
	return nil
}

func (m *BarBody) FieldValuePtrs() []any { 
	return []any{&m.Id}
}

func (m *BarBody) RelatedValuePtrs() []any { return []any{&m.Foos} }


func (m *BarBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.Foos
	}
	return nil
}


func (m *BarBody) LoadFoos(noCache ...bool) ([]*FooBody, error) {
	if len(m.Foos) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformTestB.Bar.Foos.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.Foos, nil
}

