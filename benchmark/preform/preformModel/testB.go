package preformModel

import (
	"github.com/go-preform/preform"
	"time"
)

var testBInit = preform.InitFactory[*FactoryTestB, TestBBody](func(s *PreformBenchmarkSchema) {
	s.TestB.TestA.InitRelation(s.TestB.AId, s.TestA.Id)
	s.TestA.TestBS.InitRelation(s.TestA.Id, s.TestB.AId)
	s.TestB.SetTableName("test_b")
})

type FactoryTestB struct {
	preform.Factory[*FactoryTestB, TestBBody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"int8"`
	AId *preform.ForeignKey[int64] `db:"a_id" json:"AId" dataType:"int8"`
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"varchar"`
	Int4 *preform.Column[int32] `db:"int4" json:"Int4" dataType:"int4"`
	Int8 *preform.Column[int64] `db:"int8" json:"Int8" dataType:"int8"`
	Float4 *preform.Column[float32] `db:"float4" json:"Float4" dataType:"float4"`
	Float8 *preform.Column[float64] `db:"float8" json:"Float8" dataType:"float8"`
	Bool *preform.Column[bool] `db:"bool" json:"Bool" dataType:"bool"`
	Text *preform.Column[string] `db:"text" json:"Text" dataType:"text"`
	Time *preform.Column[time.Time] `db:"time" json:"Time" dataType:"timestamptz"`
	
	//relations
	TestA *preform.ToOne[*TestBBody, *FactoryTestA, TestABody]
	TestCS *preform.ToMany[*TestBBody, *FactoryTestC, TestCBody]
}

func (f FactoryTestB) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryTestB, TestBBody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	ff.AId = cols[1].(*preform.ForeignKey[int64] )
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


type TestBBody struct {
	preform.Body[TestBBody,*FactoryTestB]
	Id int64 `db:"id" json:"Id" dataType:"int8"`
	AId int64 `db:"a_id" json:"AId" dataType:"int8"`
	Name string `db:"name" json:"Name" dataType:"varchar"`
	Int4 int32 `db:"int4" json:"Int4" dataType:"int4"`
	Int8 int64 `db:"int8" json:"Int8" dataType:"int8"`
	Float4 float32 `db:"float4" json:"Float4" dataType:"float4"`
	Float8 float64 `db:"float8" json:"Float8" dataType:"float8"`
	Bool bool `db:"bool" json:"Bool" dataType:"bool"`
	Text string `db:"text" json:"Text" dataType:"text"`
	Time time.Time `db:"time" json:"Time" dataType:"timestamptz"`
	
	TestA *TestABody
	TestCS []*TestCBody
}

func (m TestBBody) Factory() *FactoryTestB { return m.Body.Factory(PreformBenchmark.TestB) }

func (m *TestBBody) Insert(cfg ... preform.EditConfig) error { return PreformBenchmark.TestB.Insert(m, cfg...) }

func (m *TestBBody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformBenchmark.TestB.UpdateByPk(m, cfg...) }

func (m *TestBBody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformBenchmark.TestB.DeleteByPk(m, cfg...) }

func (m TestBBody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.AId, &m.Name, &m.Int4, &m.Int8, &m.Float4, &m.Float8, &m.Bool, &m.Text, &m.Time} }

func (m *TestBBody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.AId
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

func (m *TestBBody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.AId, &m.Name, &m.Int4, &m.Int8, &m.Float4, &m.Float8, &m.Bool, &m.Text, &m.Time}
}

func (m *TestBBody) RelatedValuePtrs() []any { return []any{&m.TestA, &m.TestCS} }


func (m *TestBBody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.TestA
			case 1: return &m.TestCS
	}
	return nil
}


func (m *TestBBody) LoadTestA(noCache ...bool) (*TestABody, error) {
	if m.TestA == nil || len(noCache) != 0 && noCache[0] {
		err := PreformBenchmark.TestB.TestA.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.TestA, nil
}

func (m *TestBBody) LoadTestCS(noCache ...bool) ([]*TestCBody, error) {
	if len(m.TestCS) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformBenchmark.TestB.TestCS.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.TestCS, nil
}

