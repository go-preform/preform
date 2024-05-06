package preformModel

import (
	"github.com/go-preform/preform"
	"time"
)

var testAInit = preform.InitFactory[*FactoryTestA, TestABody](func(s *PreformBenchmarkSchema) {
	s.TestA.SetTableName("test_a")
})

type FactoryTestA struct {
	preform.Factory[*FactoryTestA, TestABody]
	Id *preform.PrimaryKey[int64] `db:"id" json:"Id" dataType:"int8"`
	Name *preform.Column[string] `db:"name" json:"Name" dataType:"varchar"`
	Int4 *preform.Column[int32] `db:"int4" json:"Int4" dataType:"int4"`
	Int8 *preform.Column[int64] `db:"int8" json:"Int8" dataType:"int8"`
	Float4 *preform.Column[float32] `db:"float4" json:"Float4" dataType:"float4"`
	Float8 *preform.Column[float64] `db:"float8" json:"Float8" dataType:"float8"`
	Bool *preform.Column[bool] `db:"bool" json:"Bool" dataType:"bool"`
	Text *preform.Column[string] `db:"text" json:"Text" dataType:"text"`
	Time *preform.Column[time.Time] `db:"time" json:"Time" dataType:"timestamptz"`
	TestBS *preform.ToMany[*TestABody, *FactoryTestB, TestBBody]
}

func (f FactoryTestA) CloneInstance(factory preform.IFactory) preform.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.Factory = *factory.(*preform.Factory[*FactoryTestA, TestABody])
	ff.Factory.Definition = &ff
	ff.Id = cols[0].(*preform.PrimaryKey[int64] )
	ff.Name = cols[1].(*preform.Column[string] )
	ff.Int4 = cols[2].(*preform.Column[int32] )
	ff.Int8 = cols[3].(*preform.Column[int64] )
	ff.Float4 = cols[4].(*preform.Column[float32] )
	ff.Float8 = cols[5].(*preform.Column[float64] )
	ff.Bool = cols[6].(*preform.Column[bool] )
	ff.Text = cols[7].(*preform.Column[string] )
	ff.Time = cols[8].(*preform.Column[time.Time] )
	return ff.Factory.Definition
}


type TestABody struct {
	preform.Body[TestABody,*FactoryTestA]
	Id int64 `db:"id" json:"Id" dataType:"int8"`
	Name string `db:"name" json:"Name" dataType:"varchar"`
	Int4 int32 `db:"int4" json:"Int4" dataType:"int4"`
	Int8 int64 `db:"int8" json:"Int8" dataType:"int8"`
	Float4 float32 `db:"float4" json:"Float4" dataType:"float4"`
	Float8 float64 `db:"float8" json:"Float8" dataType:"float8"`
	Bool bool `db:"bool" json:"Bool" dataType:"bool"`
	Text string `db:"text" json:"Text" dataType:"text"`
	Time time.Time `db:"time" json:"Time" dataType:"timestamptz"`
	TestBS []*TestBBody
}

func (m TestABody) Factory() *FactoryTestA { return m.Body.Factory(PreformBenchmark.TestA) }

func (m *TestABody) Insert(cfg ... preform.EditConfig) error { return PreformBenchmark.TestA.Insert(m, cfg...) }

func (m *TestABody) Update(cfg ... preform.UpdateConfig) (affected int64, err error) { return PreformBenchmark.TestA.UpdateByPk(m, cfg...) }

func (m *TestABody) Delete(cfg ... preform.EditConfig) (affected int64, err error) { return PreformBenchmark.TestA.DeleteByPk(m, cfg...) }

func (m TestABody) FieldValueImmutablePtrs() []any { return []any{&m.Id, &m.Name, &m.Int4, &m.Int8, &m.Float4, &m.Float8, &m.Bool, &m.Text, &m.Time} }

func (m *TestABody) FieldValuePtr(pos int) any { 
	switch pos {
		case 0: return &m.Id
		case 1: return &m.Name
		case 2: return &m.Int4
		case 3: return &m.Int8
		case 4: return &m.Float4
		case 5: return &m.Float8
		case 6: return &m.Bool
		case 7: return &m.Text
		case 8: return &m.Time
	}
	return nil
}

func (m *TestABody) FieldValuePtrs() []any { 
	return []any{&m.Id, &m.Name, &m.Int4, &m.Int8, &m.Float4, &m.Float8, &m.Bool, &m.Text, &m.Time}
}

func (m *TestABody) RelatedValuePtrs() []any { return []any{&m.TestBS} }


func (m *TestABody) RelatedByPos(pos uint32) any {
	switch pos {
			case 0: return &m.TestBS
	}
	return nil
}


func (m *TestABody) LoadTestBS(noCache ...bool) ([]*TestBBody, error) {
	if len(m.TestBS) == 0 || len(noCache) != 0 && noCache[0] {
		err := PreformBenchmark.TestA.TestBS.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.TestBS, nil
}

