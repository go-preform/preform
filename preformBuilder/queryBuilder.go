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

type QueryBuilder struct {
	name       string
	src        []*queryBuilderSrc
	factories  []preformShare.IFactoryBuilder
	conditions []preformShare.ICondForBuilder
	having     []preformShare.ICondForBuilder
	groupBy    []any
	cols       []preformShare.IColDef
	schemas    []string
	setter     any

	disableColAlign bool
}

func BuildQuery(name string, funcAcceptBuilderPlusModels any) *QueryBuilder {
	return newQueryBuilder(name, funcAcceptBuilderPlusModels)
}

func newQueryBuilder(name string, fn any) *QueryBuilder {
	b := &QueryBuilder{name: name, src: []*queryBuilderSrc{}, setter: fn}
	if _, found := preformShare.BuildingQueries[name]; found {
		panic("queryBuilder " + name + " already exists")
	}
	preformShare.BuildingQueries[name] = b
	return b
}

// build
func (builder *QueryBuilder) prepare() {
	var (
		fv   = reflect.ValueOf(builder.setter)
		ft   = fv.Type()
		l    = ft.NumIn()
		i    int
		args = make([]reflect.Value, l)
	)
	if l == 1 {
		panic("funcAcceptBuilderPlusModels must accept at least one model")
	}
	for i = 1; i < l; i++ {
		if f, found := preformShare.BuildingSchemas[ft.In(i)]; found {
			args[i] = reflect.ValueOf(f)
			builder.schemas = append(builder.schemas, args[i].Elem().FieldByName("name").String())
		} else {
			panic("funcAcceptBuilderPlusModels must accept models that have been registered")
		}
	}
	args[0] = reflect.ValueOf(builder)
	fv.Call(args)
}

func (builder *QueryBuilder) From(f preformShare.IFactoryBuilder) *QueryBuilderSetAlias {
	s := &queryBuilderSrc{src: f}
	builder.src = append(builder.src, s)
	builder.factories = append(builder.factories, f)
	return &QueryBuilderSetAlias{QueryBuilder: builder, src: s}
}

func (builder *QueryBuilder) join(direction string, f preformShare.IFactoryBuilder, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	src := &queryBuilderSrc{src: f, joinDirection: direction}
	if l := len(cond); l != 0 {
		if len(cond) > 1 {
			cond[0] = cond[0].IAnd(cond[1:]...)
		}
		src.joinCond = cond[0]
	}
	builder.src = append(builder.src, src)
	builder.factories = append(builder.factories, f)
	return &QueryBuilderSetAlias{QueryBuilder: builder, src: src}
}

func (builder *QueryBuilder) LeftJoin(f preformShare.IFactoryBuilder, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	return builder.join("Left", f, cond...)
}

func (builder *QueryBuilder) RightJoin(f preformShare.IFactoryBuilder, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	return builder.join("Right", f, cond...)
}

func (builder *QueryBuilder) InnerJoin(f preformShare.IFactoryBuilder, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	return builder.join("Inner", f, cond...)
}

type foreignKeyForJoin interface {
	Factory() preformShare.IFactoryBuilder
	AutoAssociatedCond(factories []preformShare.IFactoryBuilder) preformShare.ICondForBuilder
}

type QueryBuilderSetAlias struct {
	*QueryBuilder
	src *queryBuilderSrc
}

func (builder *QueryBuilderSetAlias) As(alias string) *QueryBuilder {
	builder.src.src.SetAlias(alias)
	return builder.QueryBuilder
}

func (builder *QueryBuilder) joinByForeignKey(direction string, fk foreignKeyForJoin, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	cond = append([]preformShare.ICondForBuilder{fk.AutoAssociatedCond(builder.factories)}, cond...)
	src := &queryBuilderSrc{src: fk.Factory(), joinCond: cond[0], joinDirection: direction}
	if len(cond) > 1 {
		cond[0] = cond[0].IAnd(cond[1:]...)
	}
	src.joinCond = cond[0]
	builder.src = append(builder.src, src)
	return &QueryBuilderSetAlias{QueryBuilder: builder, src: src}
}

func (builder *QueryBuilder) LeftJoinByForeignKey(fk foreignKeyForJoin, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	return builder.joinByForeignKey("Left", fk, cond...)
}

func (builder *QueryBuilder) RightJoinByForeignKey(fk foreignKeyForJoin, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	return builder.joinByForeignKey("Right", fk, cond...)
}

func (builder *QueryBuilder) InnerJoinByForeignKey(fk foreignKeyForJoin, cond ...preformShare.ICondForBuilder) *QueryBuilderSetAlias {
	return builder.joinByForeignKey("Inner", fk, cond...)
}

func (builder *QueryBuilder) Cols(cols ...preformShare.IColDef) *QueryBuilder {
	builder.cols = append(builder.cols, cols...)
	return builder
}

func (builder *QueryBuilder) Where(condition preformShare.ICondForBuilder) *QueryBuilder {
	builder.conditions = append(builder.conditions, condition)
	return builder
}

func (builder *QueryBuilder) GroupBy(stmt ...any) *QueryBuilder {
	builder.groupBy = append(builder.groupBy, stmt...)
	return builder
}

func (builder *QueryBuilder) Having(condition preformShare.ICondForBuilder) *QueryBuilder {
	builder.having = append(builder.having, condition)
	return builder
}

func (f *QueryBuilder) SetColAlign(enable bool) *QueryBuilder {
	f.disableColAlign = !enable
	return f
}

func (builder *QueryBuilder) GenerateCode(schemaName string) (name, schemaField, factoryName, defCode, modelCode string, importPaths []string) {
	var (
		modelName       = strcase.ToCamel(builder.name)
		exportModelName = modelName //strcase.ToCamel(modelName)
		defColCodes     []string
		modelColCodes   []string
		defColCode      []string
		modelColCode    []string
		settingCodes    []string
		importPath      []string
		importedPaths   = map[string]struct{}{}
		schemaInputs    []string
		colSettingCodes []string
		ptrsCode        []string
		ptrsSwitchCode  []string
		extraFunc       string
		extraFuncs      []string
		aliasSet        = map[string]struct{}{}
		subSettingCodes []string
		settingRx       = regexp.MustCompile(pkgName + "\\.(?:SetAssociatedKey|SetColumn)\\([^\\)]+\\)(.+)")
		colName         string
	)
	fmt.Println("generateCode", modelName)
	builder.prepare()
	for _, s := range builder.schemas {
		schemaInputs = append(schemaInputs, fmt.Sprintf(`%sSchema *%sSchema`, s, s))
	}
	for i, src := range builder.src {
		defColCodes = append(defColCodes, fmt.Sprintf(`%s *Factory%s`, src.src.Alias(), src.src.CodeName()))
		settingCodes = append(settingCodes, fmt.Sprintf("d.%s = d.%sSchema.%s.SetAlias(\"%s\").(*Factory%s)", src.src.Alias(), src.src.SchemaName(), src.src.CodeName(), src.src.Alias(), src.src.CodeName()))
		if i == 0 {
			subSettingCodes = append(subSettingCodes, fmt.Sprintf(`d.SetSrc(d.%s)`, src.src.Alias()))
		} else {
			subSettingCodes = append(subSettingCodes, fmt.Sprintf(
				`Join("%s", d.%s, %s)`,
				src.joinDirection,
				src.src.Alias(),
				src.joinCond.ToCode(),
			))
		}
	}
	settingCodes = append(settingCodes, strings.Join(subSettingCodes, ".\n\t\t")+".DefineCols(")
	if len(builder.cols) == 0 {
		for _, src := range builder.src {
			for _, col := range src.src.Cols() {
				builder.cols = append(builder.cols, col.SetAliasI(src.src.CodeName()+col.CodeName()))
			}
		}
	}
	if !builder.disableColAlign {
		builder.alignColumns()
	}
	defColCodes = append(defColCodes, "", "//columns")
	for _, col := range builder.cols {
		importPath, defColCode, modelColCode, colSettingCodes, extraFunc = col.GenerateCode(col.Factory().SchemaName(), true)
		defColCodes = append(defColCodes, defColCode...)
		modelColCodes = append(modelColCodes, modelColCode...)
		if col.Alias() == col.OColName() {
			colName = strcase.ToCamel(fmt.Sprintf(`"%s %s"`, col.Factory().CodeName(), col.CodeName()))
		} else {
			colName = col.CodeName()
		}
		if _, found := aliasSet[colName]; found {
			panic("duplicate column name: " + colName)
		}
		ptrsCode = append(ptrsCode, fmt.Sprintf(`&m.%s`, colName))
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
			settingCodes = append(settingCodes, settingRx.ReplaceAllString(strings.Join(colSettingCodes, "\n\t"), fmt.Sprintf(`	%s.SetPrebuildQueryCol(d, d.%s.%s.SetAlias("%s")$1, d.%s),`, pkgName, col.Factory().Alias(), col.SrcName(), colName, colName)))
		} else {
			settingCodes = append(settingCodes, fmt.Sprintf(`	%s.SetPrebuildQueryCol(d, d.%s.%s.SetAlias("%s"), d.%s),`, pkgName, col.Factory().Alias(), col.SrcName(), colName, colName))
		}
	}
	settingCodes = append(settingCodes, ")")
	if len(builder.conditions) != 0 {
		settingCodes[len(settingCodes)-1] += "."
		subSettingCodes = []string{}
		for _, cond := range builder.conditions {
			subSettingCodes = append(subSettingCodes, fmt.Sprintf(`PreSetWhere(%s)`, cond.ToCode()))
		}
		settingCodes = append(settingCodes, strings.Join(subSettingCodes, ".\n\t\t"))
	}
	if len(builder.having) != 0 {
		settingCodes[len(settingCodes)-1] += "."
		subSettingCodes = []string{}
		for _, cond := range builder.having {
			subSettingCodes = append(subSettingCodes, fmt.Sprintf(`Having(d.%s)`, cond.UseAlias(builder.cols...).ToCondCode()))
		}
		settingCodes = append(settingCodes, strings.Join(subSettingCodes, ".\n\t\t"))
	}
	if len(builder.groupBy) != 0 {
		settingCodes[len(settingCodes)-1] += "."
		stmts := []string{}
		for _, stmt := range builder.groupBy {
			if col, ok := stmt.(preformShare.IColDef); ok {
				useAlias := false
				for _, c := range builder.cols {
					if c == col {
						useAlias = true
						break
					}
				}
				if col.Alias() == col.OColName() {
					if useAlias {
						colName = strcase.ToCamel(fmt.Sprintf(`%s %s`, col.Factory().CodeName(), col.CodeName()))
					} else {
						colName = fmt.Sprintf(`%s.%s`, col.Factory().Alias(), col.SrcName())
					}
				} else {
					colName = col.Alias()
				}
				stmts = append(stmts, "d."+colName)
			} else if cond, ok := stmt.(preformShare.ICondForBuilder); ok {
				stmts = append(stmts, cond.ToCode())
			}
		}
		settingCodes = append(settingCodes, fmt.Sprintf(`GroupBy(%s)`, strings.Join(stmts, ", ")))
	}
	ptrsSwitchCode = make([]string, len(ptrsCode))
	for i := range ptrsCode {
		ptrsSwitchCode[i] = fmt.Sprintf("case %d: return %s", i, ptrsCode[i])
	}
	return modelName, exportModelName, fmt.Sprintf("Factory%s", modelName), fmt.Sprintf(
			`var %s = %s.IniPrebuildQueryFactory[*%s, %sBody](func(d *%s) {
	%s
})

type %s struct {
	%s.PrebuildQueryFactory[*%s, %sBody]
	//schema src
	%s

	//factory src
	%s
}`,
			exportModelName,
			pkgName,
			modelName+"Factory",
			exportModelName,
			modelName+"Factory",
			strings.Join(settingCodes, "\n\t"),
			modelName+"Factory",
			pkgName,
			modelName+"Factory",
			exportModelName,
			strings.Join(schemaInputs, "\n\t"),
			strings.Join(defColCodes, "\n\t"),
		), fmt.Sprintf(
			`type %sBody struct {
	%s.QueryBody[%sBody, *%s]
	%s
}

func (m %sBody) Factory() *%s { return %s }

func (m *%sBody) FieldValuePtr(pos int) any { 
	switch pos {
		%s
	}
	return nil
}

func (m *%sBody) FieldValuePtrs() []any { 
	return []any{%s}
}

%s`,
			exportModelName,
			pkgName,
			exportModelName,
			modelName+"Factory",
			strings.Join(modelColCodes, "\n\t"),
			exportModelName,
			modelName+"Factory",
			exportModelName,
			exportModelName,
			strings.Join(ptrsSwitchCode, "\n\t\t"),
			exportModelName,
			strings.Join(ptrsCode, ", "),
			strings.Join(extraFuncs, "\n\n"),
		),
		importPaths
}

func (f *QueryBuilder) alignColumns() {
	type colAlign struct {
		col   preformShare.IColDef
		align int
	}
	var (
		aligns = make([]colAlign, len(f.cols))
	)
	for i, c := range f.cols {
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
		f.cols[i] = c.col
	}
}
