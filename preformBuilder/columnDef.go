package preformBuilder

import (
	"fmt"
	"github.com/go-preform/preform/dialect"
	preformShare "github.com/go-preform/preform/share"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type ColumnDef[T any] struct {
	*columnDef[T]
	overwriteType preformShare.IColDef
}

type columnDefTagOpt struct {
	PK bool
}

type iColDef interface {
	preformShare.IColDef
	InitCol(ref *reflect.StructField, builder iFactoryBuilder)
}

type columnDef[T any] struct {
	fieldRef       *reflect.StructField
	settings       [][]string
	alias          string
	name           string
	codeName       string
	oColName       string
	typeName       string
	factoryBuilder iFactoryBuilder
	ConditionForBuilder[T]
	relatedFks []preformShare.IColDef
	tags       string
}

func (c *columnDef[T]) RelatedFk(def preformShare.IColDef) *columnDef[T] {
	c.relatedFks = append(c.relatedFks, def)
	return c
}

func (c columnDef[T]) RelatedFks() []preformShare.IColDef {
	return c.relatedFks
}

func (c *columnDef[T]) Factory() preformShare.IFactoryBuilder {
	return c.factoryBuilder
}

func (c columnDef[T]) SrcName() string {
	return c.codeName
}

func (c columnDef[T]) CodeName() string {
	if c.alias != "" {
		return strcase.ToCamel(c.alias)
	} else {
		return c.codeName
	}
}

func InitColBySchema(col *preformShare.Column, builder iFactoryBuilder) iColDef {
	c := &ColumnDef[any]{}
	c.columnDef = &columnDef[any]{name: col.Name, typeName: col.GoType, codeName: strcase.ToCamel(col.Name), factoryBuilder: builder, ConditionForBuilder: NewConditionForBuilder[any](c)}
	c.oColName = c.name
	c.alias = c.name
	col.IColDef = c
	if len(col.ForeignKeys) != 0 {
		return &ForeignKeyDef[any]{ColumnDef: c, associated: map[preformShare.IFactoryBuilder][]*foreignKeyCnf{}}
	}
	if col.IsPrimaryKey {
		if col.IsAutoKey {
			return c.PK().AutoIncrement()
		}
		return c.PK()
	}
	return c
}

func (c *ColumnDef[T]) InitCol(ref *reflect.StructField, builder iFactoryBuilder) {
	c.columnDef = &columnDef[T]{fieldRef: ref, name: ref.Tag.Get("db"), codeName: ref.Name, factoryBuilder: builder, ConditionForBuilder: NewConditionForBuilder[T](c)}
	if c.name == "" {
		c.name = strcase.ToSnake(ref.Name)
	}
	c.tags = strings.Replace(string(ref.Tag), fmt.Sprintf(`db:"%s"`, ref.Tag.Get("db")), "", -1)
	c.oColName = c.name
	c.alias = c.name
}

func (c ColumnDef[T]) ColDef() preformShare.IColDef {
	return c
}

// OverwriteType
func (c *ColumnDef[T]) OverwriteType(col preformShare.IColDef) *ColumnDef[T] {
	c.overwriteType = col
	return c
}

func (c ColumnDef[T]) GetType() reflect.Type {
	if c.overwriteType != nil {
		return c.overwriteType.GetType()
	}
	if c.columnDef != nil {
		return c.columnDef.GetType()
	}
	var (
		t  T
		rt = reflect.TypeOf(t)
	)
	//if rt.Kind() == reflect.Ptr {
	//	return rt.Elem()
	//}
	return rt
}

func (c columnDef[T]) GetType() reflect.Type {
	var (
		t  T
		rt = reflect.TypeOf(t)
	)
	//if rt != nil && rt.Kind() == reflect.Ptr {
	//	return rt.Elem()
	//}
	return rt
}

func (c *columnDef[T]) setAlias(alias string) *columnDef[T] {
	cc := *c
	cc.alias = alias
	cc.settings = [][]string{{"setAlias", fmt.Sprintf(`"%s"`, alias)}}
	return &cc
}

func (c *ColumnDef[T]) SetAlias(alias string) *ColumnDef[T] {
	return &ColumnDef[T]{columnDef: c.setAlias(alias)}
}

func (c ColumnDef[T]) SetAliasI(alias string) preformShare.IColDef {
	return c.SetAlias(alias)
}

func (c columnDef[T]) Alias() string {
	if c.alias != "" {
		return c.alias
	} else {
		return c.name
	}
}

func (c columnDef[T]) OColName() string {
	return c.oColName
}

func (c *ColumnDef[T]) SetName(name string) *ColumnDef[T] {
	c.name = name
	c.settings = append(c.settings, []string{"SetName", fmt.Sprintf(`"%s"`, name)})
	return c
}

type PrimaryKeyDef[T any] struct {
	*ForeignKeyDef[T]
}

func (c *PrimaryKeyDef[T]) InitCol(ref *reflect.StructField, builder iFactoryBuilder) {
	c.ForeignKeyDef = &ForeignKeyDef[T]{}
	c.ForeignKeyDef.InitCol(ref, builder)
	if ref.Tag.Get("autoKey") == "true" {
		c.AutoIncrement()
	}
}

func (c PrimaryKeyDef[T]) GenerateCode(schemaName string, fromQuery bool) (importPath []string, defColCode []string, modelColCode []string, settingCodes []string, extraFunc string) {
	importPath, defColCode, modelColCode, settingCodes, extraFunc = c.ForeignKeyDef.GenerateCode(schemaName, fromQuery)
	if !fromQuery {
		if c.typeName == "" {
			var (
				ct     = c.GetType()
				ctName string
			)
			if ct == nil {
				ctName = "any"
			} else {
				ctName = ct.String()
			}

			defColCode[0] = fmt.Sprintf("%s *%s.PrimaryKey[%s] `db:\"%s\"%s`", c.CodeName(), pkgName, ctName, c.alias, c.tags)
		} else {
			defColCode[0] = fmt.Sprintf("%s *%s.PrimaryKey[%s] `db:\"%s\"%s`", c.CodeName(), pkgName, c.typeName, c.alias, c.tags)
		}
	}
	if len(settingCodes) != 0 {
		if strings.Contains(settingCodes[0], ".SetForeignKey(s") {
			settingCodes[0] = strings.Replace(settingCodes[0], fmt.Sprintf(".SetForeignKey(s.%s.%s)", c.Factory().CodeName(), c.CodeName()), fmt.Sprintf(".SetForeignKey(s.%s.%s.ForeignKey)", c.Factory().CodeName(), c.CodeName()), 1)

		} else if strings.Contains(settingCodes[0], ".SetColumn(s") {
			settingCodes[0] = strings.Replace(settingCodes[0], fmt.Sprintf(".SetColumn(s.%s.%s)", c.Factory().CodeName(), c.CodeName()), fmt.Sprintf(".SetColumn(s.%s.%s.Column)", c.Factory().CodeName(), c.CodeName()), 1)
		}
	}
	//	if c.middleTables != nil {
	//		for name, mt := range c.middleTables {
	//			if mt.cnf.relationName != "" {
	//				name = mt.cnf.relationName
	//			}
	//			defColCode = append(defColCode, fmt.Sprintf("%s *%s.MiddleTable[*%sBody, *Factory%s, %sBody]", name, pkgName, c.Factory().FullCodeName(), mt.middleTableTarget.FullCodeName(), mt.middleTableTarget.FullCodeName()))
	//			modelColCode = append(modelColCode, fmt.Sprintf("%s []*%sBody", name, mt.middleTableTarget.FullCodeName()))
	//			c.setRelationCodes = append(c.setRelationCodes, []string{
	//				fmt.Sprintf("case %%v: m.%s = toSet[0].([]*%sBody)", name, mt.middleTableTarget.FullCodeName()),
	//				name,
	//				fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) ([]*%sBody, error) {
	//	if len(m.%s) == 0 || len(noCache) != 0 && noCache[0] {
	//		err := %s.%s.%s.Load(m)
	//		if err != nil {
	//			return nil, err
	//		}
	//	}
	//	return m.%s, nil
	//}`, c.Factory().FullCodeName(), name, mt.middleTableTarget.FullCodeName(), name, schemaName, c.Factory().CodeName(), name, name),
	//			})
	//		}
	//		setRelationCodesTemp = append(setRelationCodesTemp, c.setRelationCodes...)
	//	}
	//if c.middleTableLocalKeys != nil {
	//	var (
	//		middleTableName           string
	//		middleTableLocalColName   string
	//		middleTableForeignColName string
	//	)
	//	if c.middleTableName != "" {
	//		middleTableName = c.middleTableName
	//	} else {
	//		middleTableName = fmt.Sprintf("%s_%s", c.factoryBuilder.TableName(), c.middleTableLocalKeys.Factory().TableName())
	//	}
	//	if c.middleTableLocalColName != "" {
	//		middleTableLocalColName = c.middleTableLocalColName
	//	} else {
	//		middleTableLocalColName = fmt.Sprintf("%s_id", c.factoryBuilder.TableName())
	//	}
	//	if c.middleTableLocalKeyNames != "" {
	//		middleTableForeignColName = c.middleTableLocalKeyNames
	//	} else {
	//		middleTableForeignColName = fmt.Sprintf("%s_id", c.middleTableLocalKeys.Factory().TableName())
	//	}
	//	settingCodes[0] += fmt.Sprintf(".MiddleTable(%s.%s, %s, %s, %s)", c.middleTableLocalKeys.Factory().CodeName(), c.middleTableLocalKeys.CodeName(), middleTableName, middleTableLocalColName, middleTableForeignColName)
	//}
	return
}

func (c *ColumnDef[T]) PK() *PrimaryKeyDef[T] {
	c.settings = append(c.settings, []string{"PK"})
	c.factoryBuilder.setPk(c)
	return &PrimaryKeyDef[T]{ForeignKeyDef: &ForeignKeyDef[T]{ColumnDef: c}}
}

func (c *PrimaryKeyDef[T]) AutoIncrement() *PrimaryKeyDef[T] {
	c.settings = append(c.settings, []string{"AutoIncrement"})
	return c
}

func (c ColumnDef[T]) GenerateCode(schemaName string, fromQuery bool) (importPath []string, defColCode []string, modelColCode []string, settingCodes []string, extraFunc string) {
	var (
		tType           = c.GetType()
		typeName        string
		tmpSettingCodes []string
		name            = c.CodeName()
		defCode         []string
		bodyCode        []string
	)
	if c.typeName != "" {
		typeName = c.typeName
		if strings.Contains(typeName, fmt.Sprintf("Enum_%s", strcase.ToCamel(schemaName))) {
			typeName = strings.Replace(typeName, fmt.Sprintf("Enum_%s", strcase.ToCamel(schemaName)), strcase.ToCamel(schemaName), 1)
		}
		if strings.Contains(typeName, fmt.Sprintf("CustomType_%s", strcase.ToCamel(schemaName))) {
			typeName = strings.Replace(typeName, fmt.Sprintf("CustomType_%s", strcase.ToCamel(schemaName)), strcase.ToCamel(schemaName), 1)
		}
		if strings.Contains(typeName, "preformTypes.") {
			importPath = append(importPath, "github.com/go-preform/preform/types")
		}
		if strings.Contains(typeName, "sql.") || strings.Contains(typeName, "sql.") {
			importPath = append(importPath, "database/sql")
		}
		if strings.Contains(typeName, "time.") || strings.Contains(typeName, "[time.") {
			importPath = append(importPath, "time")
		}
		if strings.Contains(typeName, "uuid.") {
			importPath = append(importPath, "github.com/satori/go.uuid")
		}
	} else if tType != nil {
		var imports []string
		typeName, imports = parseColType(tType, name, schemaName)
		importPath = append(importPath, imports...)
	}
	if fromQuery {
		if c.OColName() == c.Alias() {
			name = strcase.ToCamel(fmt.Sprintf("%s %s", c.Factory().CodeName(), c.CodeName()))
		}
		defCode = []string{fmt.Sprintf("%s *%s.PrebuildQueryCol[%s, %s.NoAggregation]", name, pkgName, typeName, pkgName)}
		bodyCode = []string{fmt.Sprintf("%s %s `db:\"%s\"%s`", name, typeName, c.alias, c.tags)}
	} else {
		//if strings.HasPrefix(typeName, "[]") && typeName != "[]byte" {
		//	defCode = []string{fmt.Sprintf("%s *%s.ArrayColumn[%s]", name, pkgName, strings.Replace(typeName, "[]", "", 1))}
		//} else {
		defCode = []string{fmt.Sprintf("%s *%s.Column[%s] `db:\"%s\"%s`", name, pkgName, typeName, c.alias, c.tags)}
		//}
		bodyCode = []string{fmt.Sprintf("%s %s `db:\"%s\"%s`", name, typeName, c.alias, c.tags)}
	}
	for _, setting := range c.settings {
		if fromQuery && (setting[0] == "setAlias" || setting[0] == "AutoIncrement") {
			continue
		}
		tmpSettingCodes = append(tmpSettingCodes, fmt.Sprintf("%s(%s)", setting[0], strings.Join(setting[1:], ",")))
	}
	if len(tmpSettingCodes) != 0 {
		settingCodes = []string{fmt.Sprintf("%s.SetColumn(s.%s.%s).%s", pkgName, c.Factory().CodeName(), name, strings.Join(tmpSettingCodes, "."))}
	}
	return importPath,
		defCode,
		bodyCode,
		settingCodes, ""
}

type aggregatedCol[T any] struct {
	*columnDef[T]
	aggregateSetting []string
	aggregatedType   reflect.Type
}

func (c aggregatedCol[T]) ColDef() preformShare.IColDef {
	return c
}

func (c aggregatedCol[T]) GenerateCode(schemaName string, fromQuery bool) (importPath []string, defColCode []string, modelColCode []string, settingCodes []string, extraFunc string) {
	var (
		t               *T
		tType           = reflect.TypeOf(t).Elem()
		tmpSettingCodes []string
		name            = c.CodeName()
		defCode         []string
		bodyCode        []string
	)
	if c.aggregatedType != nil {
		tType = c.aggregatedType
	}
	defCode = []string{fmt.Sprintf("%s *%s.PrebuildQueryCol[%s, %s.NoAggregation]", name, pkgName, tType.String(), pkgName)}
	bodyCode = []string{fmt.Sprintf("%s %s `db:\"%s\"%s`", name, tType.String(), c.alias, c.tags)}
	for _, setting := range c.settings {
		if fromQuery && setting[0] == "setAlias" {
			tmpSettingCodes = append(tmpSettingCodes, fmt.Sprintf("SetAlias(%s)", strings.Join(setting[1:], ",")))
		} else {
			tmpSettingCodes = append(tmpSettingCodes, fmt.Sprintf("%s(%s)", setting[0], strings.Join(setting[1:], ",")))
		}
	}
	settingCodes = []string{fmt.Sprintf("%s.SetColumn(s.%s.%s).%s(%s)", pkgName, c.Factory().CodeName(), name, c.aggregateSetting[0], strings.Join(c.aggregateSetting[1:], ","))}
	if len(tmpSettingCodes) != 0 {
		settingCodes[0] += fmt.Sprintf(".%s", strings.Join(tmpSettingCodes, "."))
	}
	return []string{tType.PkgPath()},
		defCode,
		bodyCode,
		settingCodes, ""
}

func (a columnDef[T]) Aggregate(fn preformShare.Aggregator, aggregatedType reflect.Type, params ...any) *aggregatedCol[T] {
	var (
		paramsCode = make([]string, len(params))
		aa         = a.setAlias(strcase.ToCamel(a.CodeName() + " " + string(fn)))
	)
	if len(params) == 0 {
		return &aggregatedCol[T]{aa, []string{"Aggregate", fmt.Sprintf(`"%s"`, fn)}, aggregatedType}
	}
	for i := range params {
		switch params[i].(type) {
		case string:
			paramsCode[i] = fmt.Sprintf(`"%s"`, params[i])
		default:
			paramsCode[i] = fmt.Sprintf(`%v`, params[i])
		}
	}
	return &aggregatedCol[T]{aa, []string{"Aggregate", fmt.Sprintf(`"%s", %s`, fn, strings.Join(paramsCode, `,`))}, aggregatedType}

}

func (a columnDef[T]) Sum() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggSum, nil)
}

func (a columnDef[T]) Avg() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggAvg, nil)
}

func (a columnDef[T]) Max() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggMax, nil)
}

func (a columnDef[T]) Min() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggMin, nil)
}

func (a columnDef[T]) Count() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggCount, reflect.TypeOf(0))
}

func (a columnDef[T]) CountDistinct() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggCountDistinct, reflect.TypeOf(0))
}

func (a columnDef[T]) Mean() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggMean, nil)
}

func (a columnDef[T]) Median() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggMedian, nil)
}

func (a columnDef[T]) Mode() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggMode, nil)
}

func (a columnDef[T]) StdDev() *aggregatedCol[T] {
	return a.Aggregate(dialect.AggStdDev, nil)
}
func (a columnDef[T]) GroupConcat(splitter string) *aggregatedCol[T] {
	return a.Aggregate(dialect.AggGroupConcat, reflect.TypeOf(""), splitter)
}

// SetAlias
func (a *aggregatedCol[T]) SetAlias(alias string) *aggregatedCol[T] {
	a.columnDef = a.columnDef.setAlias(alias)
	return a
}
func (a aggregatedCol[T]) SetAliasI(alias string) preformShare.IColDef {
	return a.SetAlias(alias)
}
