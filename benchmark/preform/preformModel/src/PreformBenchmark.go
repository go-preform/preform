package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
)



type PreformBenchmark_testA struct {
	preformBuilder.FactoryBuilder[*PreformBenchmark_testA]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"int8"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"varchar"`
	Int4	preformBuilder.ColumnDef[int32] `db:"int4" json:"Int4" dataType:"int4"`
	Int8	preformBuilder.ColumnDef[int64] `db:"int8" json:"Int8" dataType:"int8"`
	Float4	preformBuilder.ColumnDef[float32] `db:"float4" json:"Float4" dataType:"float4"`
	Float8	preformBuilder.ColumnDef[float64] `db:"float8" json:"Float8" dataType:"float8"`
	Bool	preformBuilder.ColumnDef[bool] `db:"bool" json:"Bool" dataType:"bool"`
	Text	preformBuilder.ColumnDef[string] `db:"text" json:"Text" dataType:"text"`
}

type PreformBenchmark_testB struct {
	preformBuilder.FactoryBuilder[*PreformBenchmark_testB]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"int8"`
	AId	preformBuilder.ForeignKeyDef[int64] `db:"a_id" json:"AId" dataType:"int8"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"varchar"`
	Int4	preformBuilder.ColumnDef[int32] `db:"int4" json:"Int4" dataType:"int4"`
	Int8	preformBuilder.ColumnDef[int64] `db:"int8" json:"Int8" dataType:"int8"`
	Float4	preformBuilder.ColumnDef[float32] `db:"float4" json:"Float4" dataType:"float4"`
	Float8	preformBuilder.ColumnDef[float64] `db:"float8" json:"Float8" dataType:"float8"`
	Bool	preformBuilder.ColumnDef[bool] `db:"bool" json:"Bool" dataType:"bool"`
	Text	preformBuilder.ColumnDef[string] `db:"text" json:"Text" dataType:"text"`
}

type PreformBenchmark_testC struct {
	preformBuilder.FactoryBuilder[*PreformBenchmark_testC]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"int8"`
	BId	preformBuilder.ForeignKeyDef[int64] `db:"b_id" json:"BId" dataType:"int8"`
	Name	preformBuilder.ColumnDef[string] `db:"name" json:"Name" dataType:"varchar"`
	Int4	preformBuilder.ColumnDef[int32] `db:"int4" json:"Int4" dataType:"int4"`
	Int8	preformBuilder.ColumnDef[int64] `db:"int8" json:"Int8" dataType:"int8"`
	Float4	preformBuilder.ColumnDef[float32] `db:"float4" json:"Float4" dataType:"float4"`
	Float8	preformBuilder.ColumnDef[float64] `db:"float8" json:"Float8" dataType:"float8"`
	Bool	preformBuilder.ColumnDef[bool] `db:"bool" json:"Bool" dataType:"bool"`
	Text	preformBuilder.ColumnDef[string] `db:"text" json:"Text" dataType:"text"`
}

type PreformBenchmarkSchema struct {
	name string
	testA *PreformBenchmark_testA
	testB *PreformBenchmark_testB
	testC *PreformBenchmark_testC
}

var (
	PreformBenchmark = PreformBenchmarkSchema{name: "PreformBenchmark"}
)

func initPreformBenchmark() (string, []preformShare.IFactoryBuilder, *PreformBenchmarkSchema, map[string][]string, map[string]*preformShare.CustomType) {

	//implement IFactoryBuilderWithSetup in a new file if you need to customize the factory
	
	PreformBenchmark.testA = preformBuilder.InitFactoryBuilder(PreformBenchmark.name, func(d *PreformBenchmark_testA) {
		d.SetTableName("test_a")
		d.Id.RelatedFk(&PreformBenchmark.testB.AId)
	})
	
	PreformBenchmark.testB = preformBuilder.InitFactoryBuilder(PreformBenchmark.name, func(d *PreformBenchmark_testB) {
		d.SetTableName("test_b")
		d.Id.RelatedFk(&PreformBenchmark.testC.BId)
		d.AId.SetAssociatedKey(PreformBenchmark.testA.Id, preformBuilder.FkName("test_b_pk_fk"))
	})
	
	PreformBenchmark.testC = preformBuilder.InitFactoryBuilder(PreformBenchmark.name, func(d *PreformBenchmark_testC) {
		d.SetTableName("test_c")
		d.BId.SetAssociatedKey(PreformBenchmark.testB.Id, preformBuilder.FkName("test_c_pk_fk"))
	})

	return "preform_benchmark",
		[]preformShare.IFactoryBuilder{
			PreformBenchmark.testA,
			PreformBenchmark.testB,
			PreformBenchmark.testC,
		},
		&PreformBenchmark,
		map[string][]string{},
        map[string]*preformShare.CustomType{}
}