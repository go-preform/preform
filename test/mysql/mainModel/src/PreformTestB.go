package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
)



type PreformTestB_bar struct {
	preformBuilder.FactoryBuilder[*PreformTestB_bar]
	Id1	preformBuilder.PrimaryKeyDef[int32] `db:"id1" json:"Id1" dataType:"int"`
	Id2	preformBuilder.PrimaryKeyDef[int32] `db:"id2" json:"Id2" dataType:"int"`
}

type PreformTestBSchema struct {
	name string
	bar *PreformTestB_bar
}

var (
	PreformTestB = PreformTestBSchema{name: "PreformTestB"}
)

func initPreformTestB() (string, []preformShare.IFactoryBuilder, *PreformTestBSchema, map[string][]string, map[string]*preformShare.CustomType) {

	//implement IFactoryBuilderWithSetup in a new file if you need to customize the factory
	
	PreformTestB.bar = preformBuilder.InitFactoryBuilder(PreformTestB.name, func(d *PreformTestB_bar) {
		d.SetTableName("bar")
		d.Id1.SetAssociatedKey(PreformTestA.foo.Fk1, preformBuilder.FkName("bar_FK"), preformBuilder.FkComposite(d.Id2, PreformTestA.foo.Fk2))
	})

	return "preform_test_b",
		[]preformShare.IFactoryBuilder{
			PreformTestB.bar,
		},
		&PreformTestB,
		map[string][]string{},
        map[string]*preformShare.CustomType{}
}