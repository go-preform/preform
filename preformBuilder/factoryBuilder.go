package preformBuilder

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/iancoleman/strcase"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

const (
	pkgName = "preform"
	pkgPath = "github.com/go-preform/preform"
)

type iFactoryBuilder interface {
	preformShare.IFactoryBuilder
	preGenerateCode(schemaName string)
	generateCode(schemaName string) (name, schemaField, factoryName, defCode, modelCode string, importPaths, inheritors []string)
	addToColSet(col preformShare.IColDef)
	define(def iFactoryBuilder, codeName, schemaName string)
	setSetter(setter any)
	setPk(col preformShare.IColDef)
	setSchemaPrefix(bool)
	AllFieldNames() map[string]struct{}
	addAssociated(builder preformShare.IFactoryBuilder)
	associatedModelTimes(model preformShare.IFactoryBuilder) int
}
type FactoryBuilder[D iFactoryBuilder] struct {
	Definition       D
	setter           func(D)
	codeName         string
	needSchemaPrefix bool
	alias, name      string
	colSet           map[string]preformShare.IColDef
	primaryKeys      []preformShare.IColDef
	columns          []preformShare.IColDef
	schema           string

	allFieldNames map[string]struct{}

	defColCodes      []string
	modelColCodes    []string
	settingCodes     []string
	importPaths      []string
	ptrsCode         []string
	extraFuncs       []string
	setRelationCodes [][]string
	cloneColCodes    []string

	inheritors           []string
	inheritorByCondition map[string]preformShare.ICondForBuilder

	customFields       [][2]any
	associatedModelCnt map[string]int
	isView             bool

	disableColAlign bool
}

func (f *FactoryBuilder[D]) setSchemaPrefix(b bool) {
	f.needSchemaPrefix = b
}

func (f *FactoryBuilder[D]) AddCode(def, model, setting, extraFuncs []string, setRelationCodes [][]string) {
	if def != nil {
		f.defColCodes = append(f.defColCodes, def...)
	}
	if model != nil {
		f.modelColCodes = append(f.modelColCodes, model...)
	}
	if setting != nil {
		f.settingCodes = append(f.settingCodes, setting...)
	}
	if extraFuncs != nil {
		f.extraFuncs = append(f.extraFuncs, extraFuncs...)
	}
	if setRelationCodes != nil {
		f.setRelationCodes = append(f.setRelationCodes, setRelationCodes...)
	}
}

func (f *FactoryBuilder[D]) addAssociated(builder preformShare.IFactoryBuilder) {
	if _, ok := f.associatedModelCnt[builder.CodeName()]; !ok {
		f.associatedModelCnt[builder.CodeName()] = 1
	} else {
		f.associatedModelCnt[builder.CodeName()]++
	}
}

func (f *FactoryBuilder[D]) associatedModelTimes(model preformShare.IFactoryBuilder) int {
	return f.associatedModelCnt[model.CodeName()]
}

func (f *FactoryBuilder[D]) AddCustomField(name string, valueType reflect.Type) {
	f.customFields = append(f.customFields, [2]any{name, valueType})
}

func (f *FactoryBuilder[D]) setSetter(setter any) {
	f.setter = setter.(func(D))
}

func (f *FactoryBuilder[D]) Clone() preformShare.IFactoryBuilder {
	ff := *f
	ff.Definition = f.Definition
	v := reflect.ValueOf(ff.Definition)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v.Field(0).Set(reflect.ValueOf(ff))
	return v.Addr().Interface().(preformShare.IFactoryBuilder)
}

func (f *FactoryBuilder[D]) AddInheritor(tableName string) preformShare.IFactoryBuilder {
	f.inheritors = append(f.inheritors, tableName)
	return f
}

func (f *FactoryBuilder[D]) AddInheritorByCondition(codeName string, cond preformShare.ICondForBuilder) preformShare.IFactoryBuilder {
	if f.inheritorByCondition == nil {
		f.inheritorByCondition = map[string]preformShare.ICondForBuilder{}
	}
	f.inheritorByCondition[codeName] = cond
	return f
}

func (f *FactoryBuilder[D]) SetAlias(alias string) preformShare.IFactoryBuilder {
	f.alias = alias
	return f
}

func (f *FactoryBuilder[D]) AddAssociated(builder preformShare.IFactoryBuilder) {
	if _, ok := f.associatedModelCnt[builder.CodeName()]; !ok {
		f.associatedModelCnt[builder.CodeName()] = 1
	} else {
		f.associatedModelCnt[builder.CodeName()]++
	}
}

func (f *FactoryBuilder[D]) AssociatedModelTimes(model preformShare.IFactoryBuilder) int {
	return f.associatedModelCnt[model.CodeName()]
}

func (f FactoryBuilder[D]) FactoryType() reflect.Type {
	return reflect.TypeOf(f.Definition)
}

func (f FactoryBuilder[D]) SchemaName() string {
	return f.schema
}

func (f FactoryBuilder[D]) Alias() string {
	if f.alias != "" {
		return f.alias
	}
	return f.codeName
}

func (f FactoryBuilder[D]) ColSet() map[string]preformShare.IColDef {
	return f.colSet
}

func (f FactoryBuilder[D]) Cols() []preformShare.IColDef {
	return f.columns
}

func (f *FactoryBuilder[D]) define(def iFactoryBuilder, codeName, schemaName string) {
	if def != nil {
		f.Definition = def.(D)
	}
	if strings.Contains(codeName, "_") {
		f.codeName = strcase.ToCamel(strings.Split(codeName, "_")[1])
	} else {
		f.codeName = strcase.ToCamel(codeName)
	}

	f.associatedModelCnt = map[string]int{}
	f.name = strcase.ToSnake(f.codeName)
	f.schema = strcase.ToCamel(schemaName)
	f.allFieldNames = map[string]struct{}{}
}

// setPk
func (f *FactoryBuilder[D]) setPk(col preformShare.IColDef) {
	f.primaryKeys = append(f.primaryKeys, col)
}

func (f *FactoryBuilder[D]) addToColSet(col preformShare.IColDef) {
	if f.colSet == nil {
		f.colSet = map[string]preformShare.IColDef{}
	}
	f.columns = append(f.columns, col)
	f.colSet[col.CodeName()] = col
}

func (f *FactoryBuilder[D]) IsView() {
	f.isView = true
	if !strings.Contains(f.codeName, "View") {
		f.codeName = f.codeName + "View"
	}
}

// PK
func (f *FactoryBuilder[D]) PK() []preformShare.IColDef {
	if len(f.primaryKeys) == 0 {
		f.primaryKeys = append(f.primaryKeys, f.columns[0])
	}
	return f.primaryKeys
}

func (f FactoryBuilder[D]) CodeName() string {
	return f.codeName
}

func (f FactoryBuilder[D]) TableName() string {
	return f.name
}

func (f FactoryBuilder[D]) AllFieldNames() map[string]struct{} {
	return f.allFieldNames
}

func InitFactoryFromTable(table *preformShare.Table, importPaths map[string]struct{}) *FactoryBuilder[iFactoryBuilder] {
	var (
		factory = &FactoryBuilder[iFactoryBuilder]{colSet: make(map[string]preformShare.IColDef), allFieldNames: map[string]struct{}{}, associatedModelCnt: map[string]int{}}
		setters []func(iFactoryBuilder)
	)
	factory.define(nil, strcase.ToCamel(table.Name), strcase.ToCamel(table.Scheme.Name))
	for importPath := range importPaths {
		factory.importPaths = append(factory.importPaths, importPath)
	}

	for _, col := range table.Columns {
		c := InitColBySchema(col, factory)
		factory.addToColSet(c)
		factory.allFieldNames[c.CodeName()] = struct{}{}

		if len(col.ForeignKeys) != 0 {
			setters = append(setters, func(col *preformShare.Column, c *ForeignKeyDef[any]) func(f iFactoryBuilder) {
				return func(f iFactoryBuilder) {
					for _, cc := range col.ForeignKeys {
						settings := []foreignKeyCnfSetter{}
						if cc.RelationName != "" {
							settings = append(settings, FkRelationName(cc.RelationName))
						}
						if cc.ReverseName != "" {
							settings = append(settings, FkReverseName(cc.ReverseName))
						}
						c.SetAssociatedKey(cc.LocalKeys[0].IColDef, settings...)
					}
				}
			}(col, c.(*ForeignKeyDef[any])))
		}
	}
	factory.setSetter(func(f iFactoryBuilder) {
		for _, setter := range setters {
			setter(f)
		}
	})

	if ff, ok := preformShare.BuildingModelsByName[factory.CodeName()]; ok {
		factory.setSchemaPrefix(true)
		ff.(iFactoryBuilder).setSchemaPrefix(true)
	} else {
		preformShare.BuildingModelsByName[factory.CodeName()] = factory
	}
	return factory
}

func InitFactoryBuilder[D iFactoryBuilder](schemaName string, setter func(D)) D {
	var (
		dd         D
		dv         = reflect.New(reflect.TypeOf(dd).Elem())
		d          = dv.Interface().(D)
		defTypeRef = dv.Elem().Type()
		factory    = dv.Elem().Field(0).Addr().Interface().(iFactoryBuilder)
		defRef     reflect.Value
	)
	factory.define(d, defTypeRef.Name(), schemaName)
	factory.setSetter(setter)

	defRef = dv.Elem()
	for i := 0; i < defTypeRef.NumField(); i++ {
		fieldRef := defTypeRef.Field(i)
		if fieldRef.Type.Kind() == reflect.Struct && fieldRef.Type.NumField() != 0 {
			if c, ok := defRef.Field(i).Addr().Interface().(iColDef); ok {
				c.InitCol(&fieldRef, d)
				defRef.Field(i).Set(reflect.ValueOf(c).Elem())
				factory.addToColSet(c)
				factory.AllFieldNames()[c.CodeName()] = struct{}{}
			}
		}
	}
	preformShare.BuildingModels[defTypeRef] = d

	if ff, ok := preformShare.BuildingModelsByName[factory.CodeName()]; ok {
		factory.setSchemaPrefix(true)
		ff.(iFactoryBuilder).setSchemaPrefix(true)
	} else {
		preformShare.BuildingModelsByName[factory.CodeName()] = d
	}
	return d
}

func (f *FactoryBuilder[D]) SetTableName(name string) *FactoryBuilder[D] {
	f.name = name
	return f
}

func (f *FactoryBuilder[D]) FullCodeName() string {
	if f.needSchemaPrefix {
		return fmt.Sprintf("%s_%s", f.schema, f.codeName)
	}
	return f.codeName
}

func (f *FactoryBuilder[D]) SetColAlign(enable bool) *FactoryBuilder[D] {
	f.disableColAlign = !enable
	return f
}

func (f *FactoryBuilder[D]) preGenerateCode(schemaName string) {
	var (
		//defRef     = reflect.ValueOf(f.Definition).Elem()
		//defTypeRef = defRef.FactoryType()
		//modelName        = f.CodeName()
		defColCode          []string
		modelColCode        []string
		importPath          []string
		importedPaths       = map[string]struct{}{}
		colSettingCodes     []string
		extraFunc           string
		defColCodes         []string
		modelColCodes       []string
		defColCodesSuffix   []string
		modelColCodesSuffix []string
		ptrsCode            []string
		importPaths         []string
		extraFuncs          []string
		settingCodes        []string
		setRelationCodes    [][]string
		cloneColCodes       []string
		tagRx               = regexp.MustCompile("([^\\s]+)[\\s\\t]+(.+)[\\s\\t]*(`[^`]+`)")
	)
	for _, importPath := range f.importPaths {
		importedPaths[importPath] = struct{}{}
	}
	if fs, ok := any(f.Definition).(preformShare.IFactoryBuilderWithSetup); ok {
		if !fs.Setup() {
			f.setter(f.Definition)
		}
	} else {
		f.setter(f.Definition)
	}
	if f.name != "" {
		f.settingCodes = append(f.settingCodes, fmt.Sprintf(`s.%s.SetTableName("%s")`, f.codeName, f.name))
	}
	if !f.disableColAlign {
		f.alignColumns()
	}
	//fmt.Println("preGenerateCode", modelName)
	for i, c := range f.columns {
		if _, ok := any(c).(iPk); ok {
			f.primaryKeys = append(f.primaryKeys, c)
		}
		importPath, defColCode, modelColCode, colSettingCodes, extraFunc = c.GenerateCode(schemaName, false)

		defColCodeParts := tagRx.FindStringSubmatch(defColCode[0])
		cloneColCodes = append(cloneColCodes, fmt.Sprintf(`ff.%s = cols[%d].(%s)`, defColCodeParts[1], i, defColCodeParts[2]))
		if cc, ok := c.(iForeignKeyDef); ok {
			setRelationCodes = append(setRelationCodes, cc.GetSetRelationCodes()...)
		}
		defColCodes = append(defColCodes, defColCode[0])
		modelColCodes = append(modelColCodes, modelColCode[0])
		if len(defColCode) != 1 {
			defColCodesSuffix = append(defColCodesSuffix, defColCode[1:]...)
			modelColCodesSuffix = append(modelColCodesSuffix, modelColCode[1:]...)
		}
		ptrsCode = append(ptrsCode, fmt.Sprintf(`m.%s`, c.CodeName()))
		if len(importPath) != 0 {
			for _, p := range importPath {
				p = strings.TrimSpace(p)
				pathParts := strings.Split(p, " ")
				if _, ok := importedPaths[p]; !ok && p != "" {
					if len(pathParts) == 2 {
						if _, ok := importedPaths[pathParts[1]]; ok {
							continue
						}
					}
					importedPaths[p] = struct{}{}
					if len(pathParts) == 1 {
						importPaths = append(importPaths, fmt.Sprintf(`"%s"`, p))
					} else {
						importPaths = append(importPaths, fmt.Sprintf(`%s "%s"`, pathParts[0], pathParts[1]))
					}
				}
			}
		}
		if extraFunc != "" {
			extraFuncs = append(extraFuncs, extraFunc)
		}
		if len(colSettingCodes) != 0 {
			settingCodes = append(settingCodes, colSettingCodes...)
		}
		//		}
		//	}
		//}
	}

	if f.customFields != nil {
		defColCodes = append(defColCodes, "")
		modelColCodes = append(modelColCodes, "")
		for _, f := range f.customFields {
			typeName, thisImports := parseColType(f[1].(reflect.Type), f[0].(string), schemaName)
			importPaths = append(importPaths, thisImports...)
			modelColCodes = append(modelColCodes, fmt.Sprintf("%s %s `db:\"-\"`", f[0].(string), typeName))
			defColCodes = append(defColCodes, fmt.Sprintf(`%s *%s.CustomField[%s]`, f[0].(string), pkgName, typeName))
		}
	}
	if len(modelColCodesSuffix) != 0 {
		defColCodesSuffix = append([]string{"", "//relations"}, defColCodesSuffix...)
		defColCodes = append(defColCodes, defColCodesSuffix...)
		modelColCodesSuffix = append([]string{""}, modelColCodesSuffix...)
		modelColCodes = append(modelColCodes, modelColCodesSuffix...)
	}
	f.defColCodes = append(defColCodes, f.defColCodes...)
	f.modelColCodes = append(modelColCodes, f.modelColCodes...)
	f.ptrsCode = append(ptrsCode, f.ptrsCode...)
	f.importPaths = importPaths
	f.extraFuncs = append(extraFuncs, f.extraFuncs...)
	f.settingCodes = append(settingCodes, f.settingCodes...)
	f.setRelationCodes = append(setRelationCodes, f.setRelationCodes...)
	f.cloneColCodes = append(cloneColCodes, f.cloneColCodes...)
}

func (f *FactoryBuilder[D]) generateCode(schemaName string) (name, schemaField, factoryName, defCode, modelCode string, importPaths, inheritors []string) {
	var (
		modelName        = strcase.ToLowerCamel(f.CodeName())
		exportModelName  = strcase.ToCamel(modelName)
		schemaFieldName  = exportModelName
		setRelationCode  string
		setRelationCodes []string
		getRelatedCode   []string
		relatedPtrsCode  []string
		schemaInputs     = []string{fmt.Sprintf(`s *%sSchema`, schemaName)}
		schemaInputNames = map[string]struct{}{schemaInputs[0]: {}}
		relatedByPosCode string
	)
	f.schema = schemaName
	fmt.Println("generateCode", modelName)
	if f.needSchemaPrefix {
		modelName = strcase.ToLowerCamel(schemaName) + "_" + modelName
		exportModelName = schemaName + "_" + exportModelName
	}
	for i, c := range f.setRelationCodes {
		setRelationCode += fmt.Sprintf("\t\t\t%s\n", fmt.Sprintf(c[0], i))
		relatedPtrsCode = append(relatedPtrsCode, fmt.Sprintf("&m.%s", c[1]))
		getRelatedCode = append(getRelatedCode, c[2])
		if strings.Contains(c[0], "[]") {
			setRelationCodes = append(setRelationCodes, fmt.Sprintf("\t\tcase %v: return len(m.%s) != 0", i, c[1]))
		} else {
			setRelationCodes = append(setRelationCodes, fmt.Sprintf("\t\tcase %v: return m.%s != nil", i, c[1]))
		}
		if len(c) > 3 {
			if _, ok := schemaInputNames[c[3]]; !ok {
				schemaInputNames[c[3]] = struct{}{}
				schemaInputs = append(schemaInputs, c[3])
			}
		}
	}
	for _, iTable := range f.inheritors {
		f.settingCodes = append(f.settingCodes, fmt.Sprintf(`s.%s = any(s.%s.Clone().(*Factory%s).SetTableName("%s").Definition).(*Factory%s)`, strcase.ToCamel(iTable), strcase.ToCamel(f.CodeName()), exportModelName, iTable, exportModelName))
	}
	for codeName, cond := range f.inheritorByCondition {
		f.inheritors = append(f.inheritors, codeName)
		f.settingCodes = append(f.settingCodes, fmt.Sprintf(`s.%s = any(s.%s.Clone().(*Factory%s).SetFixedCondition(s.%s.%s).Definition).(*Factory%s)`, strcase.ToCamel(codeName), strcase.ToCamel(f.CodeName()), exportModelName, strcase.ToCamel(f.CodeName()), cond.ToCondCode(), exportModelName))
	}
	ptrsSwitchCode := make([]string, len(f.ptrsCode))
	for i, ptr := range f.ptrsCode {
		ptrsSwitchCode[i] = fmt.Sprintf("case %d: return &%s", i, ptr)

	}
	if len(setRelationCodes) != 0 {
		relatedByPosCode = fmt.Sprintf(`
func (m *%sBody) RelatedByPos(pos uint32) any {
	switch pos {
%s	}
	return nil
}
`,
			exportModelName,
			setRelationCode)
	} else {

		relatedByPosCode = fmt.Sprintf(`
func (m *%sBody) RelatedByPos(pos uint32, toSet ...any) bool {
	return false
}
`,
			exportModelName)
	}
	factoryTypeName := "Factory"
	if f.isView {
		factoryTypeName = "ViewFactory"
	}
	return modelName, schemaFieldName, fmt.Sprintf("Factory%s", exportModelName), fmt.Sprintf(
			`var %sInit = %s.Init%s[*Factory%s, %sBody](func(%s) {
	%s
})

type Factory%s struct {
	%s.%s[*Factory%s, %sBody]
	%s
}

func (f Factory%s) CloneInstance(factory %s.IFactory) %s.IFactory {
	var (
		ff = f
		cols = factory.Columns()
	)
	ff.%s = *factory.(*%s.%s[*Factory%s, %sBody])
	ff.%s.Definition = &ff
	%s
	return ff.%s.Definition
}
`,
			modelName,
			pkgName,
			factoryTypeName,
			exportModelName,
			exportModelName,
			strings.Join(schemaInputs, ", "),
			strings.Join(f.settingCodes, "\n\t"),
			exportModelName,
			pkgName,
			factoryTypeName,
			exportModelName,
			exportModelName,
			strings.Join(f.defColCodes, "\n\t"),
			exportModelName,
			pkgName,
			pkgName,
			factoryTypeName,
			pkgName,
			factoryTypeName,
			exportModelName,
			exportModelName,
			factoryTypeName,
			strings.Join(f.cloneColCodes, "\n\t"),
			factoryTypeName,
		), fmt.Sprintf(
			`type %sBody struct {
	%s.Body[%sBody,*Factory%s]
	%s
}

func (m %sBody) Factory() *Factory%s { return m.Body.Factory(%s) }

func (m *%sBody) Insert(cfg ... %s.EditConfig) error { return %s.Insert(m, cfg...) }

func (m *%sBody) Update(cfg ... %s.UpdateConfig) (affected int64, err error) { return %s.UpdateByPk(m, cfg...) }

func (m *%sBody) Delete(cfg ... %s.EditConfig) (affected int64, err error) { return %s.DeleteByPk(m, cfg...) }

func (m %sBody) FieldValueImmutablePtrs() []any { return []any{&%s} }

func (m *%sBody) FieldValuePtr(pos int) any { 
	switch pos {
		%s
	}
	return nil
}

func (m *%sBody) FieldValuePtrs() []any { 
	return []any{&%s}
}

func (m *%sBody) RelatedValuePtrs() []any { return []any{%s} }

%s

%s
%s`,
			exportModelName,
			pkgName,
			exportModelName,
			exportModelName,
			strings.Join(f.modelColCodes, "\n\t"),
			exportModelName,
			exportModelName,
			schemaName+"."+f.CodeName(),
			exportModelName,
			pkgName,
			schemaName+"."+f.CodeName(),
			exportModelName,
			pkgName,
			schemaName+"."+f.CodeName(),
			exportModelName,
			pkgName,
			schemaName+"."+f.CodeName(),
			exportModelName,
			strings.Join(f.ptrsCode, ", &"),
			exportModelName,
			strings.Join(ptrsSwitchCode, "\n\t\t"),
			exportModelName,
			strings.Join(f.ptrsCode, ", &"),
			exportModelName,
			strings.Join(relatedPtrsCode, ", "),
			relatedByPosCode,
			strings.Join(getRelatedCode, "\n\n"),
			strings.Join(f.extraFuncs, "\n\n"),
		),
		f.importPaths, f.inheritors
}

type iTypeForExport interface {
	TypeForExport() any
}

func (f *FactoryBuilder[D]) alignColumns() {
	type colAlign struct {
		col   preformShare.IColDef
		align int
	}
	var (
		aligns = make([]colAlign, len(f.columns))
	)
	for i, c := range f.columns {
		v := c.NewValue()
		for {
			if t, ok := v.(iTypeForExport); ok {
				v = t.TypeForExport()
			} else {
				break
			}
		}
		if rt := reflect.TypeOf(v); rt == nil {
			aligns[i] = colAlign{c, 8}
		} else {
			if rt.Kind() == reflect.Array {
				aligns[i] = colAlign{c, rt.Align() * rt.Len()}
			} else if rt.Kind() == reflect.Struct {
				aligns[i] = colAlign{c, rt.FieldAlign()}
			} else {
				aligns[i] = colAlign{c, rt.Align()}
			}
		}
	}
	sort.Slice(aligns, func(i, j int) bool {
		return aligns[i].align > aligns[j].align
	})
	for i, c := range aligns {
		f.columns[i] = c.col
	}
}
