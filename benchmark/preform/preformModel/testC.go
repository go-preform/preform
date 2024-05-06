package preformModel

import (
	"github.com/go-preform/preform"
	"time"
)

var testCInit = preform.InitFactory[*FactoryTestC, TestCBody](func(s *PreformBenchmarkSchema) {
	s.TestC.TestB.InitRelation(s.TestC.BId, s.TestB.Id)
	s.TestB.TestCS.InitRelation(s.TestB.Id, s.TestC.BId)
	s.TestC.SetTableName("test_c")
})

type FactoryTestC struct {
	preform.Factory[*FactoryTestC, TestCBody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"int8"`
	BId *preform.ForeignKey[int64] `db:"b_id" json:"BId" dataType:"int8"`
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"varchar"`
	Int4 *preform.Column[int32] `db:"int4" json:"Int4" dataType:"int4"`
	Int8 *preform.Column[int64] `db:"int8" json:"Int8" dataType:"int8"`
	Float4 *preform.Column[float32] `db:"float4" json:"Float4" dataType:"float4"`
	Float8 *preform.Column[float64] `db:"float8" json:"Float8" dataType:"float8"`
	Bool *preform.Column[bool] `db:"bool" json:"Bool" dataType:"bool"`
	Text *preform.Column[string] `db:"text" json:"Text" dataType:"text"`
	Time *preform.Column[time.Time] `db:"time" json:"Time" dataType:"timestamptz"`
	
	//relations
	TestB *preform.ToOne[*TestCBody, *FactoryTestB, TestBBody]
}

func (f FactoryTestC) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryTestC, TestCBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	ff.BId = cols[1].(*preform.ForeignKey[int64] )
	ff.Name = cols[2].(*preform.Column[string] )
	ff.Int4 = cols[3].(*preform.Column[int32] )
	ff.Int8 = cols[4].(*preform.Column[int64] )
	ff.Float4 = cols[5].(*preform.Column[float32] )
	ff.Float8 = cols[6].(*preform.Column[float64] )
	ff.Bool = cols[7].(*preform.Column[bool] )
	ff.Text = cols[8].(*preform.Column[string] )
	ff.Time = cols[9].(*preform.Column[time.Time] )
	return ff.Factory.Definition
}


type TestCBody struct {
	preform.Body[TestCBody,*FactoryTestC]
	Id int64 `db:"id" json:"Id" dataType:"int8"`
	BId int64 `db:"b_id" json:"BId" dataType:"int8"`
	Name string `db:"name" json:"Name" dataType:"varchar"`
	Int4 int32 `db:"int4" json:"Int4" dataType:"int4"`
	Int8 int64 `db:"int8" json:"Int8" dataType:"int8"`
	Float4 float32 `db:"float4" json:"Float4" dataType:"float4"`
	Float8 float64 `db:"float8" json:"Float8" dataType:"float8"`
	Bool bool `db:"bool" json:"Bool" dataType:"bool"`
	Text string `db:"text" json:"Text" dataType:"text"`
	Time time.Time `db:"time" json:"Time" dataType:"timestamptz"`
	
	TestB *TestBBody
}

func (m TestCBody) Factory() *FactoryTestC { return m.Body.Factory(PreformBenchmark.TestC) }

func (m *TestCBody) Insert(cfg ... preform.EditConfig) error { return PreformBenchmark.TestC.Insert(m, cfg...) }

func (m *TestCBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformBenchmark.TestC.UpdateByPk(m, cfg...) }

func (m *TestCBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformBenchmark.TestC.DeleteByPk(m, cfg...) }

func (m TestCBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.BId, &m.Name, &m.Int4, &m.Int8, &m.Float4, &m.Float8, &m.Bool, &m.Text, &m.Time} }

func (m *TestCBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.BId
		case 2: return &m.Name
		case 3: return &m.Int4
		case 4: return &m.Int8
		case 5: return &m.Float4
		case 6: return &m.Float8
		case 7: return &m.Bool
		case 8: return &m.Text
		case 9: return &m.Time
	}
	return nil
}

func (m *TestCBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.BId, &m.Name, &m.Int4, &m.Int8, &m.Float4, &m.Float8, &m.Bool, &m.Text, &m.Time}
}

func (m *TestCBody) RelatedValuePtrs() []any { return []any{&m.TestB} }


func (m *TestCBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.TestB
	}
	return nil
}


func (m *TestCBody) LoadTestB(noCache ...bool) (*TestBBody, error) {
	if m.TestB == nil || len(noCache) != 0 && noCache[0] {
		err := PreformBenchmark.TestC.TestB.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.TestB, nil
}

