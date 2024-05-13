package preformBuilder

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gertd/go-pluralize"
	"github.com/go-preform/preform"
	preformShare "github.com/go-preform/preform/share"
	"github.com/iancoleman/strcase"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func BuildModel(conn *sql.DB, modelPkgName, outputFiles string, schemasEmptyIsAll ...string) {
	db := preform.DbFromNative(conn)

	schemas := db.Dialect().GetStructure(db.DB.DB, schemasEmptyIsAll...)

	_ = os.Mkdir(outputFiles, 0755)
	_ = os.Mkdir(outputFiles+"/src", 0755)
	//
	//modelPkgNames := strings.Split(outputFiles, "/")
	//modelPkgName := modelPkgNames[len(modelPkgNames)-1]
	//
	//fmt.Println(modelPkgName, schemas[0])

	var (
		schemaInits []string
		schemaNames []string
		tableDefs   = make([][]preformShare.IFactoryBuilder, len(schemas))
	)

	for i, schema := range schemas {
		var (
			imports   []string
			models    []string
			settings  []string
			vars      []string
			arrs      []string
			enumTypes []string

			schemaName = strcase.ToCamel(schema.Name)
		)
		if len(schema.Tables) == 0 {
			continue
		}
		for lib := range schema.Imports {
			imports = append(imports, lib)
		}
		for enumName := range schema.Enums {
			enumTypes = append(enumTypes, fmt.Sprintf(`type Enum_%s_%s string`, strcase.ToCamel(schemaName), strcase.ToCamel(enumName)))
		}

		for customTypeName, ct := range schema.CustomTypes {
			attrs := make([]string, 0, len(ct.Attr))
			for _, c := range ct.Attr {
				attrs = append(attrs, fmt.Sprintf(`%s %s`, strcase.ToCamel(c.Name), c.Type))
			}
			enumTypes = append(enumTypes, fmt.Sprintf(`type CustomType_%s_%s struct{
	%s
}`, strcase.ToCamel(schemaName), strcase.ToCamel(customTypeName), strings.Join(attrs, "\n\t")))
		}
	lookForMiddleTable:
		for _, table := range schema.Tables {
			for _, col := range table.Columns {
				if !col.IsPrimaryKey {
					continue lookForMiddleTable
				}
			}
			if len(table.ForeignKeys) == 2 {
				var (
					k1 *preformShare.ForeignKey
				)
				for _, fk := range table.ForeignKeys {
					if k1 == nil {
						k1 = fk
					} else {
						for _, k1k := range k1.LocalKeys {
							for _, k2k := range fk.LocalKeys {
								if k1k == k2k {
									continue lookForMiddleTable
								}
							}
						}
						if len(k1.ForeignKeys)+len(fk.ForeignKeys) != len(table.Columns) {
							continue lookForMiddleTable
						}
						k1.ForeignKeys[0].ForeignKeys = append(k1.ForeignKeys[0].ForeignKeys, k1)
						if k1.ForeignKeys[0].Table != fk.ForeignKeys[0].Table {
							fk.ForeignKeys[0].ForeignKeys = append(fk.ForeignKeys[0].ForeignKeys, fk)
						}
						k1.AssociatedFk = fk
						fk.AssociatedFk = k1
						table.IsMiddleTable = true
					}
				}
			}
		}
		for _, table := range schema.Tables {
			var (
				fields        []string
				field         string
				tableSettings = []string{fmt.Sprintf(`d.SetTableName("%s")`, table.Name)}
				tableName     = strcase.ToLowerCamel(table.Name)
			)
			for _, col := range table.Columns {
				tags := []string{fmt.Sprintf(`db:"%s"`, col.Name), fmt.Sprintf(`json:"%s"`, strcase.ToCamel(col.Name)), fmt.Sprintf(`dataType:"%s"`, col.Type)}
				thisSettings := []string{}
				if col.IsPrimaryKey {
					field = fmt.Sprintf("%s	preformBuilder.PrimaryKeyDef[%s]", strcase.ToCamel(col.Name), col.GoType)
					if !table.IsMiddleTable && len(col.ForeignKeys) != 0 {
						for _, fk := range col.ForeignKeys {
							if fk.LocalKeys[0] == col {
								if len(fk.LocalKeys) != 1 {
									colSetting := make([]string, len(fk.LocalKeys)-1)
									for k := range fk.LocalKeys[1:] {
										colSetting[k] = fmt.Sprintf("%sBuilder.FkComposite(d.%s, %s.%s.%s)", pkgName, strcase.ToCamel(fk.LocalKeys[k+1].Name), strcase.ToCamel(fk.ForeignKeys[k+1].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[k+1].Table.Name), strcase.ToCamel(fk.ForeignKeys[k+1].Name))
									}
									thisSettings = append(thisSettings, fmt.Sprintf(`SetAssociatedKey(%s.%s.%s, %sBuilder.FkName("%s"), %s)`, strcase.ToCamel(fk.ForeignKeys[0].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[0].Table.Name), strcase.ToCamel(fk.ForeignKeys[0].Name), pkgName, fk.Name, strings.Join(colSetting, ", ")))

								} else {
									thisSettings = append(thisSettings, fmt.Sprintf(`SetAssociatedKey(%s.%s.%s, %sBuilder.FkName("%s"))`, strcase.ToCamel(fk.ForeignKeys[0].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[0].Table.Name), strcase.ToCamel(fk.ForeignKeys[0].Name), pkgName, fk.Name))
								}
							} else if fk.LocalKeys[0].Table != table && fk.LocalKeys[0].Table.IsMiddleTable {
								var (
									localCols    = make([]string, len(fk.LocalKeys))
									localMtCols  = make([]string, len(fk.LocalKeys))
									targetCols   = make([]string, len(fk.AssociatedFk.ForeignKeys))
									targetMtCols = make([]string, len(fk.AssociatedFk.ForeignKeys))
								)

								for k := range fk.LocalKeys {
									localCols[k] = fmt.Sprintf(`%s.%s.%s`, strcase.ToCamel(fk.ForeignKeys[k].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[k].Table.Name), strcase.ToCamel(fk.ForeignKeys[k].Name))
									localMtCols[k] = fmt.Sprintf(`%s.%s.%s`, strcase.ToCamel(fk.LocalKeys[k].Table.Scheme.Name), strcase.ToLowerCamel(fk.LocalKeys[k].Table.Name), strcase.ToCamel(fk.LocalKeys[k].Name))
								}
								for k := range fk.AssociatedFk.LocalKeys {
									targetCols[k] = fmt.Sprintf(`%s.%s.%s`, strcase.ToCamel(fk.AssociatedFk.ForeignKeys[k].Table.Scheme.Name), strcase.ToLowerCamel(fk.AssociatedFk.ForeignKeys[k].Table.Name), strcase.ToCamel(fk.AssociatedFk.ForeignKeys[k].Name))
									targetMtCols[k] = fmt.Sprintf(`%s.%s.%s`, strcase.ToCamel(fk.AssociatedFk.LocalKeys[k].Table.Scheme.Name), strcase.ToLowerCamel(fk.AssociatedFk.LocalKeys[k].Table.Name), strcase.ToCamel(fk.AssociatedFk.LocalKeys[k].Name))
								}
								thisSettings = append(thisSettings, fmt.Sprintf(
									`SetAssociatedKey(%s.%s.%s, preformBuilder.FkMiddleTable(%s.%s, []preformShare.IColDef{%s}, []preformShare.IColDef{%s}, []preformShare.IColDef{%s}, []preformShare.IColDef{%s}))`,
									strcase.ToCamel(fk.LocalKeys[0].Table.Scheme.Name),
									strcase.ToLowerCamel(fk.LocalKeys[0].Table.Name),
									strcase.ToCamel(fk.LocalKeys[0].Name),
									strcase.ToCamel(fk.LocalKeys[0].Table.Scheme.Name),
									strcase.ToLowerCamel(fk.LocalKeys[0].Table.Name),
									strings.Join(localCols, ", "),
									strings.Join(localMtCols, ", "),
									strings.Join(targetCols, ", "),
									strings.Join(targetMtCols, ", "),
								))
							} else if fk.ForeignKeys[0].Table != fk.LocalKeys[0].Table {
								for i := range fk.ForeignKeys {
									if fk.ForeignKeys[i] == col {
										thisSettings = append(thisSettings, fmt.Sprintf(`RelatedFk(&%s.%s.%s)`, strcase.ToCamel(fk.LocalKeys[i].Table.Scheme.Name), strcase.ToLowerCamel(fk.LocalKeys[i].Table.Name), strcase.ToCamel(fk.LocalKeys[i].Name)))
										break
									}
								}
							}
						}
					}
				} else if !table.IsMiddleTable && len(col.ForeignKeys) != 0 {
					added := false
					for _, fk := range col.ForeignKeys {
						if fk.LocalKeys[0] == col {
							if len(fk.LocalKeys) != 1 {
								colSetting := make([]string, len(fk.LocalKeys)-1)
								for k := range fk.LocalKeys[1:] {
									colSetting[k] = fmt.Sprintf("%sBuilder.FkComposite(d.%s, %s.%s.%s)", pkgName, strcase.ToCamel(fk.LocalKeys[k+1].Name), strcase.ToCamel(fk.ForeignKeys[k+1].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[k+1].Table.Name), strcase.ToCamel(fk.ForeignKeys[k+1].Name))
								}
								thisSettings = append(thisSettings, fmt.Sprintf(`SetAssociatedKey(%s.%s.%s, %sBuilder.FkName("%s"), %s)`, strcase.ToCamel(fk.ForeignKeys[0].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[0].Table.Name), strcase.ToCamel(fk.ForeignKeys[0].Name), pkgName, fk.Name, strings.Join(colSetting, ", ")))

							} else {
								thisSettings = append(thisSettings, fmt.Sprintf(`SetAssociatedKey(%s.%s.%s, %sBuilder.FkName("%s"))`, strcase.ToCamel(fk.ForeignKeys[0].Table.Scheme.Name), strcase.ToLowerCamel(fk.ForeignKeys[0].Table.Name), strcase.ToCamel(fk.ForeignKeys[0].Name), pkgName, fk.Name))
							}
							added = true
						} else if fk.ForeignKeys[0].Table != fk.LocalKeys[0].Table {
							for i := range fk.ForeignKeys {
								if fk.ForeignKeys[i] == col {
									thisSettings = append(thisSettings, fmt.Sprintf(`RelatedFk(&%s.%s.%s)`, strcase.ToCamel(fk.LocalKeys[i].Table.Scheme.Name), strcase.ToLowerCamel(fk.LocalKeys[i].Table.Name), strcase.ToCamel(fk.LocalKeys[i].Name)))
									break
								}
							}
						}
					}
					if added {
						field = fmt.Sprintf("%s	preformBuilder.ForeignKeyDef[%s]", strcase.ToCamel(col.Name), col.GoType)
					} else {
						field = fmt.Sprintf("%s	preformBuilder.ColumnDef[%s]", strcase.ToCamel(col.Name), col.GoType)
					}
				} else {
					field = fmt.Sprintf("%s	preformBuilder.ColumnDef[%s]", strcase.ToCamel(col.Name), col.GoType)
				}
				if col.IsAutoKey {
					tags = append(tags, `autoKey:"true"`)
				}
				if col.Comment != "" {
					tags = append(tags, fmt.Sprintf(`comment:"%s"`, url.PathEscape(col.Comment)))
				}
				if col.DefaultValue.Valid {
					tags = append(tags, fmt.Sprintf(`defaultValue:"%s"`, col.DefaultValue.String))
				}
				fields = append(fields, fmt.Sprintf("%s `%s`", field, strings.Join(tags, " ")))
				if len(thisSettings) != 0 {
					tableSettings = append(tableSettings, fmt.Sprintf(`d.%s.%s`, strcase.ToCamel(col.Name), strings.Join(thisSettings, ".")))
				}
			}
			if table.IsView {
				tableSettings = append(tableSettings, `d.IsView()`)
			}
			models = append(models, fmt.Sprintf(`type %s_%s struct {
	preformBuilder.FactoryBuilder[*%s_%s]
	%s
}`, schemaName, tableName, schemaName, tableName, strings.Join(fields, "\n\t")))
			vars = append(vars, fmt.Sprintf(`%s *%s_%s`, tableName, schemaName, tableName))
			arrs = append(arrs, fmt.Sprintf(`%s.%s`, schemaName, tableName))
			settings = append(settings, fmt.Sprintf(`
	%s.%s = preformBuilder.InitFactoryBuilder(%s.name, func(d *%s_%s) {
		%s
	})`, schemaName, tableName, schemaName, schemaName, tableName, strings.Join(tableSettings, "\n\t\t")))
			for _, inherit := range table.Inheritors {
				settings = append(settings, fmt.Sprintf(`%s.%s.AddInheritor("%s") //key:"%s" bound:"%s"`, schemaName, tableName, inherit[0], inherit[1], inherit[2]))
			}

			factory := InitFactoryFromTable(table, table.Imports)
			tableDefs[i] = append(tableDefs[i], factory)

		}
		enumJson, _ := json.Marshal(schema.Enums)
		enumObj := strings.Replace(strings.Replace(string(enumJson), "[", "{", -1), "]", "}", -1)
		if enumObj == "null" {
			enumObj = "{}"
		}
		json.Unmarshal(enumJson, &schema.Enums)
		customTypeJson, _ := json.Marshal(schema.CustomTypes)
		customTypeObj := strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(
						strings.Replace(
							strings.Replace(
								strings.Replace(string(customTypeJson), `"Imports":`, "Imports: map[string]struct{}", -1), `"IsScanner":`, "IsScanner:", -1), `"NotNull":`, "NotNull:", -1), `"Type":`, "Type:", -1), `"Name":`, "Name:", -1), `}],`, "}},", -1), `"Attr":[`, "Attr: []*preformShare.CustomTypeAttr{", -1)
		if customTypeObj == "null" {
			customTypeObj = "{}"
		}
		err := os.WriteFile(fmt.Sprintf("%s/src/%s.go", outputFiles, strcase.ToCamel(schema.Name)), []byte(fmt.Sprintf(
			`package main
import (
	preformShare "github.com/go-preform/preform/share"
	%s
)

%s

%s

type %sSchema struct {
	name string
	%s
}

var (
	%s = %sSchema{name: "%s"}
)

func init%s() (string, []preformShare.IFactoryBuilder, *%sSchema, map[string][]string, map[string]*preformShare.CustomType) {

	//implement IFactoryBuilderWithSetup in a new file if you need to customize the factory
	%s

	return "%s",
		[]preformShare.IFactoryBuilder{
			%s,
		},
		&%s,
		map[string][]string%s,
        map[string]*preformShare.CustomType%s
}`,
			strings.Join(imports, "\n\t"),
			strings.Join(enumTypes, "\n"),
			strings.Join(models, "\n\n"),
			schemaName,
			strings.Join(vars, "\n\t"),
			schemaName,
			schemaName,
			strcase.ToCamel(schema.Name),
			strcase.ToCamel(schema.Name),
			schemaName,
			strings.Join(settings, "\n\t"),
			schema.Name,
			strings.Join(arrs, ",\n\t\t\t"),
			schemaName,
			enumObj,
			customTypeObj,
		)), 0777)
		if err != nil {
			panic(err)
		}
		schemaNames = append(schemaNames, schema.Name)
		schemaInits = append(schemaInits, fmt.Sprintf(`{
		name, factories, schema, enums, customTypes := init%s()
		enumBySchema[name] = enums
		customTypesBySchema[name] = customTypes
		preformShare.BuildingSchemas[reflect.TypeOf(schema)] = schema
		deferPrepareFns = append(deferPrepareFns, func(){preformBuilder.PrepareSchema("%s", "..", name, schema.name, factories)})
		deferBuildFns = append(deferBuildFns, func(){preformBuilder.BuildSchema("%s", "..", name, schema.name, factories, enums, customTypes)})
		schemas = append(schemas, schema.name)
	}`, strcase.ToCamel(schema.Name), modelPkgName, modelPkgName))

	}
	err := os.WriteFile(fmt.Sprintf("%s/src/main.go", outputFiles), []byte(fmt.Sprintf(`package main
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
	%s

	preformBuilder.BuildEnum("%s", "../", enumBySchema)
	preformBuilder.BuildCustomType("%s", "../", customTypesBySchema)
	for _, fn := range deferPrepareFns {
		fn()
	}
	for _, fn := range deferBuildFns {
		fn()
	}
	preformBuilder.BuildDbMainFile("%s", "../", PrebuildQueries, schemas...)
}
`, strings.Join(schemaInits, "\n\t"), modelPkgName, modelPkgName, modelPkgName)), 0777)
	if err != nil {
		panic(err)
	}
	err = trgGoBuild(outputFiles)
	if err != nil {
		var (
			enumBySchema = map[string]map[string][]string{}
			customTypes  = map[string]map[string]*preformShare.CustomType{}
		)
		BuildDbMainFile(modelPkgName, outputFiles, []preformShare.IQueryBuilder{}, schemaNames...)
		for i, schema := range schemas {
			PrepareSchema(modelPkgName, outputFiles, schema.Name, strcase.ToCamel(schema.Name), tableDefs[i])
		}
		for i, schema := range schemas {
			enumBySchema[schema.Name] = schema.Enums
			customTypes[schema.Name] = schema.CustomTypes
			BuildSchema(modelPkgName, outputFiles, schema.Name, strcase.ToCamel(schema.Name), tableDefs[i], schema.Enums, schema.CustomTypes)
		}
		BuildEnum(modelPkgName, outputFiles, enumBySchema)
		BuildCustomType(modelPkgName, outputFiles, customTypes)
	}

}

func BuildCustomType(modelPkgName, path string, customTypesBySchema map[string]map[string]*preformShare.CustomType) {
	if len(customTypesBySchema) == 0 {
		return
	}
	var (
		ctCodes     []string
		ctAttrCodes []string

		typeName    string
		typeDetail  *preformShare.CustomType
		imports     = map[string]struct{}{}
		importCodes []string
	)
	for schemaName, customTypes := range customTypesBySchema {
		for typeName, typeDetail = range customTypes {
			ctAttrCodes = make([]string, len(typeDetail.Attr))
			attAssignCodes := make([]string, len(typeDetail.Attr))
			attNameCodes := make([]string, len(typeDetail.Attr))
			for i, ctAttr := range typeDetail.Attr {
				ctAttrCodes[i] = fmt.Sprintf("%s %s", strcase.ToCamel(ctAttr.Name), ctAttr.Type)
				attNameCodes[i] = fmt.Sprintf(`ct.%s`, strcase.ToCamel(ctAttr.Name))
				if ctAttr.IsScanner {
					attAssignCodes[i] = fmt.Sprintf(`
	err = ct.%s.Scan(inputs[%d])
	if err != nil {
		return err
	}
`, strcase.ToCamel(ctAttr.Name), i)
				} else {
					attAssignCodes[i] = fmt.Sprintf(`
	err = preformTypes.GenericScan(&ct.%s, inputs[%d])
	if err != nil {
		return err
	}
`, strcase.ToCamel(ctAttr.Name), i)
					if _, ok := imports[`"github.com/go-preform/preform/types"`]; !ok {
						imports[`"github.com/go-preform/preform/types"`] = struct{}{}
						importCodes = append(importCodes, `"github.com/go-preform/preform/types"`)
					}
				}
			}
			for lib := range typeDetail.Imports {
				if _, ok := imports[lib]; !ok {
					imports[lib] = struct{}{}
					importCodes = append(importCodes, fmt.Sprintf(`%s`, lib))
				}
			}
			ctName := strcase.ToCamel(schemaName + " " + typeName)
			ctCodes = append(ctCodes, fmt.Sprintf(
				`type %s struct{
	%s
}

func (ct *%s) Scan(src any) error {
	inputs, err := %s.DB.GetDialect().ParseCustomTypeScan(src)
	if err != nil {
		return err
	}
	%s
	return nil
}

func (ct %s) Value() (driver.Value, error) {
	return %s.DB.GetDialect().ParseCustomTypeValue("%s", %s)
}

`,
				ctName,
				strings.Join(ctAttrCodes, "\n\t"),
				ctName,
				strcase.ToCamel(schemaName),
				strings.Join(attAssignCodes, ""),
				ctName,
				strcase.ToCamel(schemaName),
				typeName,
				strings.Join(attNameCodes, ", "),
			))
		}
	}

	if len(ctCodes) == 0 {
		return
	}

	if len(importCodes) != 0 {
		modelPkgName += `

import (
	` + strings.Join(importCodes, "\n\t") + `
	"database/sql/driver"
)`
	} else {

		modelPkgName += `

import (
	"database/sql/driver"
)`
	}

	err := os.WriteFile(fmt.Sprintf("%s/CustomTypes.go", path), []byte(fmt.Sprintf(`package %s
%s

`, modelPkgName, strings.Join(ctCodes, "\n\n"))), 0777)
	if err != nil {
		panic(err)
	}
}

func BuildEnum(modelPkgName, path string, enumsBySchema map[string]map[string][]string) {
	var (
		enumTypes      []string
		enumCodes      []string
		enumDefCodes   []string
		enumValueCodes []string
		enumTypeName   string

		enumName   string
		enumValues []string

		pluralize = pluralize.NewClient()
	)
	for schemaName, enums := range enumsBySchema {
		for enumName, enumValues = range enums {
			enumTypeName = strcase.ToCamel(schemaName + " " + enumName)
			enumTypes = append(enumTypes, fmt.Sprintf(`type %s string`, enumTypeName))
			enumDefCodes = make([]string, len(enumValues))
			enumValueCodes = make([]string, len(enumValues))
			for i, enumValue := range enumValues {
				enumDefCodes[i] = fmt.Sprintf("%s %s", strcase.ToCamel(enumValue), enumTypeName)
				enumValueCodes[i] = fmt.Sprintf(`%s: "%s",`, strcase.ToCamel(enumValue), enumValue)
			}
			enumCodes = append(enumCodes, fmt.Sprintf(`%s = struct{
		%s
	}{
		%s
	}`, strcase.ToCamel(schemaName+" "+pluralize.Plural(enumName)), strings.Join(enumDefCodes, "\n\t\t"), strings.Join(enumValueCodes, "\n\t\t")))
		}
	}

	if len(enumTypes) == 0 {
		return
	}

	err := os.WriteFile(fmt.Sprintf("%s/Enum.go", path), []byte(fmt.Sprintf(`package %s
%s

var (
	%s
)
`, modelPkgName, strings.Join(enumTypes, "\n\n"), strings.Join(enumCodes, "\n\n\t"))), 0777)
	if err != nil {
		panic(err)
	}
}

var (
	schemaNameReserveWords = map[string]bool{
		"break":       true,
		"default":     true,
		"func":        true,
		"interface":   true,
		"select":      true,
		"case":        true,
		"defer":       true,
		"go":          true,
		"any":         true,
		"chan":        true,
		"else":        true,
		"goto":        true,
		"map":         true,
		"struct":      true,
		"bool":        true,
		"const":       true,
		"fallthrough": true,
		"if":          true,
		"package":     true,
		"switch":      true,
		"true":        true,
		"false":       true,
		"import":      true,
		"range":       true,
		"type":        true,
	}
)

func BuildDbMainFile(modelPkgName, path string, queries []preformShare.IQueryBuilder, schemas ...string) {
	var (
		initCodes        []string
		cloneParams      []string
		cloneReturns     []string
		cloneCode        []string
		cloneInheritCode []string
	)

	for _, v := range queries {
		//name, schemaField, factoryName, defCode, modelCode, importPaths := v.(*QueryBuilder).GenerateCode()
		//schemaFields = append(schemaFields, fmt.Sprintf(`%s *%s`, schemaField, factoryName))
		//schemaFieldPtrs = append(schemaFieldPtrs, fmt.Sprintf(`s.%s`, schemaField))
		//schemaSetup = append(schemaSetup, fmt.Sprintf(`%s.%s = %s`, schemaName, schemaField, schemaField))
		//schemaFieldPtrPos = append(schemaFieldPtrPos, fmt.Sprintf(`&ptrs[%d]`, len(schemaFieldPtrs)+1))
		name, _, _, defCode, modelCode, importPaths := v.(*QueryBuilder).GenerateCode("")
		importPaths = append([]string{fmt.Sprintf(`"%s"`, pkgPath)}, importPaths...)
		err := os.WriteFile(path+"/"+name+".go", []byte(fmt.Sprintf(`package %s

import (
	%s
)

%s

%s
`, modelPkgName, strings.Join(importPaths, "\n\t"), defCode, modelCode)), 0777)
		if err != nil {
			fmt.Println(err)
		}
	}
	sort.Strings(schemas)
	for _, schema := range schemas {
		initCodes = append(initCodes, fmt.Sprintf(`schemas = append(schemas, init%s(conn, "", queryRunnerForTest...))`, strcase.ToCamel(schema)))
		cloneParams = append(cloneParams, fmt.Sprintf(`%sName string`, strcase.ToLowerCamel(schema)))
		cloneReturns = append(cloneReturns, fmt.Sprintf(`%s *%sSchema`, strcase.ToLowerCamel(schema), strcase.ToCamel(schema)))
		cloneCode = append(cloneCode, fmt.Sprintf(`%s = %s.clone(%sName, db...).(*%sSchema)`, strcase.ToLowerCamel(schema), strcase.ToCamel(schema), strcase.ToLowerCamel(schema), strcase.ToCamel(schema)))
		cloneInheritCode = append(cloneInheritCode, fmt.Sprintf(`%s.Inherit(%s)`, strcase.ToLowerCamel(schema), strcase.ToCamel(schema)))
	}

	err := os.WriteFile(fmt.Sprintf("%s/Db.go", path), []byte(fmt.Sprintf(`package %s
import (
	"database/sql"
	"github.com/go-preform/preform"
	"github.com/go-preform/preform/share"
)

func Init(conn *sql.DB, queryRunnerForTest ... preformShare.QueryRunner) {
	schemas := []%s.ISchema{}
	%s
	preform.PrepareQueriesAndRelation(schemas...)
}

func CloneAll(%s, db ... *sql.DB) (%s) {
	%s
	preform.PrepareQueriesAndRelation(%s)
	%s
	return
}

`, modelPkgName, pkgName, strings.Join(initCodes, "\n\t"),
		strings.Join(cloneParams, ", "),
		strings.Join(cloneReturns, ", "),
		strings.Join(cloneCode, "\n\t"),
		strings.Replace(strings.Join(cloneParams, ", "), "Name string", "", -1),
		strings.Join(cloneInheritCode, "\n\t"))), 0777)
	if err != nil {
		panic(err)
	}
}

func trgGoBuild(path string) error {
	fmt.Println("go build----------------------------")
	d, _ := os.Getwd()
	p := fmt.Sprintf("%s/%s/src", d, path)
	cmd := exec.Command("go", "run", p)
	cmd.Dir = p
	out, err := cmd.CombinedOutput()
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println(string(out), err)
	return err
}
