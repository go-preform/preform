package preformShare

import "reflect"

var (
	BuildingModels       = map[reflect.Type]IFactoryBuilder{}
	BuildingSchemas      = map[reflect.Type]any{}
	BuildingModelsByName = map[string]IFactoryBuilder{}
	BuildingQueries      = map[string]IQueryBuilder{}

	DEFAULT_VALUE = &struct{}{}

	CTX_LOGGER = &struct{}{}
)
