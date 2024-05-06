package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
)



type PreformTestB_bar struct {
	preformBuilder.FactoryBuilder[*PreformTestB_bar]
	Id	preformBuilder.PrimaryKeyDef[int64] `db:"id" json:"Id" dataType:"INTEGER"`
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
	})

	return "preform_test_b",
		[]preformShare.IFactoryBuilder{
			PreformTestB.bar,
		},
		&PreformTestB,
		map[string][]string{},
        map[string]*preformShare.CustomType{}
}