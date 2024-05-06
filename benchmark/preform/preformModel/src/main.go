package main
import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/preform/preformBuilder"
	"reflect"
)

var (
	PrebuildQueries = []preformShare.IQueryBuilder{}
)

func main() {
	var (
		schemas = []string{}
		enumBySchema = map[string]map[string][]string{}
		customTypesBySchema = map[string]map[string]*preformShare.CustomType{}
		deferPrepareFns = []func(){}
		deferBuildFns = []func(){}
	)
	{
		name, factories, schema, enums, customTypes := initPreformBenchmark()
		enumBySchema[name] = enums
		customTypesBySchema[name] = customTypes
		preformShare.BuildingSchemas[reflect.TypeOf(schema)] = schema
		deferPrepareFns = append(deferPrepareFns, func(){preformBuilder.PrepareSchema("preformModel", "..", name, schema.name, factories)})
		deferBuildFns = append(deferBuildFns, func(){preformBuilder.BuildSchema("preformModel", "..", name, schema.name, factories, enums, customTypes)})
		schemas = append(schemas, schema.name)
	}

	preformBuilder.BuildEnum("preformModel", "../", enumBySchema)
	preformBuilder.BuildCustomType("preformModel", "../", customTypesBySchema)
	for _, fn := range deferPrepareFns {
		fn()
	}
	for _, fn := range deferBuildFns {
		fn()
	}
	preformBuilder.BuildDbMainFile("preformModel", "../", PrebuildQueries, schemas...)
}
