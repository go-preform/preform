package preformBuilder

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	preformShare "github.com/go-preform/preform/share"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

var (
	ForeignKeyCnfSetters = foreignKeyCnf{}
	ReverseType          = allReverseType{ToMany: 0, ToOne: 1, NoReverse: 2}
)

type ForeignKeyDef[T any] struct {
	*ColumnDef[T]
	associated       map[preformShare.IFactoryBuilder][]*foreignKeyCnf
	setRelationCodes [][]string
}

type reverseType uint8

type allReverseType struct {
	ToMany    reverseType
	ToOne     reverseType
	NoReverse reverseType
}

type foreignKeyCnf struct {
	toMany        bool
	reverse       reverseType
	name          string
	relationName  string
	reverseName   string
	cond          preformShare.ICondForBuilder
	reverseCond   preformShare.ICondForBuilder
	compositeKeys [][2]preformShare.IColDef
	col           preformShare.IColDef
	middleTable   *middleTableDef
}

type middleTableDef struct {
	middleTable           preformShare.IFactoryBuilder
	LocalKeys             []preformShare.IColDef
	middleTableLocalRefs  []preformShare.IColDef
	targetCols            []preformShare.IColDef
	targetFactory         preformShare.IFactoryBuilder
	middleTableTargetRefs []preformShare.IColDef
}

type foreignKeyCnfSetter func(*foreignKeyCnf)

func FkMiddleTable(table preformShare.IFactoryBuilder, localKeys []preformShare.IColDef, mtLocalKey []preformShare.IColDef, foreignKeys []preformShare.IColDef, mtForeignKey []preformShare.IColDef) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		if len(localKeys) != len(mtLocalKey) || len(foreignKeys) != len(mtForeignKey) || len(localKeys) == 0 || len(foreignKeys) == 0 {
			panic("localKeys and mtLocalKey or foreignKeys and mtForeignKey must have the same length and > 0")
		}
		c.middleTable = &middleTableDef{}
		c.middleTable.middleTable = table
		c.middleTable.LocalKeys = localKeys
		c.middleTable.middleTableLocalRefs = mtLocalKey
		c.middleTable.targetCols = foreignKeys
		c.middleTable.middleTableTargetRefs = mtForeignKey
		c.middleTable.targetFactory = foreignKeys[0].Factory()
		if c.relationName == "" {
			if localKeys[0].Factory() == c.middleTable.targetFactory {
				c.relationName = fmt.Sprintf("%sBy%s%s", foreignKeys[0].Factory().CodeName(), strcase.ToCamel(table.CodeName()), strcase.ToCamel(mtForeignKey[0].CodeName()))
			} else {
				c.relationName = fmt.Sprintf("%sBy%s", foreignKeys[0].Factory().CodeName(), strcase.ToCamel(table.CodeName()))
			}
		}
		if c.reverseName == "" {
			if localKeys[0].Factory() == c.middleTable.targetFactory {
				c.reverseName = fmt.Sprintf("%sBy%s%s", localKeys[0].Factory().CodeName(), strcase.ToCamel(table.CodeName()), strcase.ToCamel(mtLocalKey[0].CodeName()))
			} else {
				c.reverseName = fmt.Sprintf("%sBy%s", localKeys[0].Factory().CodeName(), strcase.ToCamel(table.CodeName()))
			}
		}
	}
}

func FkName(name string) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		c.name = name
	}
}

func FkCond(cond preformShare.ICondForBuilder, reverseCond ...preformShare.ICondForBuilder) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		if cond != nil {
			c.cond = cond
		}
		if len(reverseCond) != 0 {
			c.reverseCond = reverseCond[0]

		}
	}
}

func FkComposite(lk, fk preformShare.IColDef) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		c.compositeKeys = append(c.compositeKeys, [2]preformShare.IColDef{lk, fk})
	}
}

func FkToMany() foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		c.toMany = true
	}
}

func FkReverse(r reverseType) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		c.reverse = r
	}
}

func FkRelationName(name string) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		c.relationName = name
	}
}

func FkReverseName(name string) foreignKeyCnfSetter {
	return func(c *foreignKeyCnf) {
		c.reverseName = name
	}
}

func (c *ForeignKeyDef[T]) InitCol(ref *reflect.StructField, builder iFactoryBuilder) {
	c.ColumnDef = &ColumnDef[T]{}
	c.ColumnDef.InitCol(ref, builder)
	c.codeName = strcase.ToCamel(ref.Name)
	//c.ColumnDef = &ColumnDef[T]{columnDef: &columnDef[T]{fieldRef: ref, codeName: strcase.ToCamel(ref.Name), factoryBuilder: builder, name: ref.Tag.Get("db")}}
	c.associated = map[preformShare.IFactoryBuilder][]*foreignKeyCnf{}
	if c.name == "" {
		c.name = strcase.ToSnake(ref.Name)
	}
}

func (c *ForeignKeyDef[T]) SetName(name string) *ForeignKeyDef[T] {
	c.name = name
	c.settings = append(c.settings, []string{"SetName", fmt.Sprintf(`"%s"`, name)})
	return c
}
func (c ForeignKeyDef[T]) ColDef() preformShare.IColDef {
	return &c
}
func (c ForeignKeyDef[T]) AutoAssociatedCond(factories []preformShare.IFactoryBuilder) preformShare.ICondForBuilder {
	for ff := range c.associated {
		for _, f := range factories {
			//golang bug :P
			if fmt.Sprintf("%p", ff) == fmt.Sprintf("%p", f) {
				for _, pk := range f.PK() {
					if pk.GetType() == c.GetType() {
						return any(c.Eq(pk)).(preformShare.ICondForBuilder)
					}
				}
			}
		}
	}
	return nil
}

func (c *ForeignKeyDef[T]) SetAssociatedKey(col preformShare.IColDef, setters ...foreignKeyCnfSetter) *ForeignKeyDef[T] {
	var (
		vT      = reflect.TypeOf((*T)(nil)).Elem()
		isArray bool
		cnf     = &foreignKeyCnf{col: col}
	)
	if c.typeName != "" {
		isArray = strings.HasPrefix(c.typeName, "[]")
	} else {
		isArray = vT.Kind() == reflect.Slice || vT.Kind() == reflect.Array
	}
	for _, setter := range setters {
		setter(cnf)
	}
	if isArray {
		cnf.toMany = true
	}
	if c.associated == nil {
		c.associated = map[preformShare.IFactoryBuilder][]*foreignKeyCnf{}
	}
	if _, found := c.associated[col.Factory()]; found {
		c.associated[col.Factory()] = append(c.associated[col.Factory()], cnf)
	} else {
		c.associated[col.Factory()] = []*foreignKeyCnf{cnf}
	}
	c.factoryBuilder.addAssociated(col.Factory())
	col.Factory().(iFactoryBuilder).addAssociated(c.factoryBuilder)
	return c
}

func (c *ForeignKeyDef[T]) PK() *ForeignKeyDef[T] {
	c.settings = append(c.settings, []string{"PK"})
	c.factoryBuilder.setPk(c)
	return c
}

func (c *ForeignKeyDef[T]) SetAlias(alias string) *ForeignKeyDef[T] {
	return &ForeignKeyDef[T]{ColumnDef: c.ColumnDef.SetAlias(alias)}
}

func (c ForeignKeyDef[T]) SetAliasI(alias string) preformShare.IColDef {
	return c.SetAlias(alias)
}

var (
	//todo clean this up
	setRelationCodesTemp = [][]string{}
)

func (c ForeignKeyDef[T]) GenerateCode(schemaName string, fromQuery bool) (importPath []string, defColCode []string, modelColCode []string, settingCodes []string, extraFunc string) {
	setRelationCodesTemp = [][]string{}
	if len(c.associated) == 0 {
		return c.ColumnDef.GenerateCode(schemaName, fromQuery)
	}
	var (
		tType            = reflect.TypeOf((*T)(nil)).Elem()
		typeName         = tType.String()
		tmpSettingCodes  []string
		name             = c.CodeName()
		defCode          []string
		bodyCode         []string
		pluralize        = pluralize.NewClient()
		targetSchemaVar  = "s"
		targetSchemaName = fmt.Sprintf("s *%sSchema", schemaName)
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
		if strings.Contains(typeName, "sql.") {
			importPath = append(importPath, "database/sql")
		}
		if strings.Contains(typeName, "time.") || strings.Contains(typeName, "[time.") {
			importPath = append(importPath, "time")
		}
		if strings.Contains(typeName, "uuid.") {
			importPath = append(importPath, "uuid github.com/satori/go.uuid")
		}
	} else if tType != nil {
		var imports []string
		typeName, imports = parseColType(tType, name, schemaName)
		importPath = append(importPath, imports...)
	}
	if fromQuery {
		defCode = append(defCode, fmt.Sprintf("%s %s.PrebuildQueryCol[%s, %s.NoAggregation]", name, pkgName, typeName, pkgName))
		bodyCode = []string{fmt.Sprintf("%s %s `db:\"%s\"%s`", name, typeName, strcase.ToSnake(name), c.tags)}
	} else {
		defCode = append(defCode, fmt.Sprintf("%s *%s.ForeignKey[%s] `db:\"%s\"%s`", name, pkgName, typeName, c.name, c.tags))
		bodyCode = []string{fmt.Sprintf("%s %s `db:\"%s\"%s`", name, typeName, c.name, c.tags)}
	}
	for _, setting := range c.settings {
		if fromQuery && setting[0] == "setAlias" {
			continue
		}
		tmpSettingCodes = append(settingCodes, fmt.Sprintf("%s(%s)", setting[0], strings.Join(setting[1:], ",")))
	}
	var (
		relationName, reverseName string
	)
	if len(tmpSettingCodes) != 0 {
		settingCodes = []string{fmt.Sprintf("%s.SetColumn(%s.%s.%s).%s", pkgName, targetSchemaVar, c.Factory().CodeName(), name, strings.Join(tmpSettingCodes, "."))}
	}
	for aModel, m := range c.associated {
		manyRelated := len(m) != 1
		for _, cnf := range m {
			relatedModelRefName := aModel.CodeName()
			suffix := strcase.ToCamel(cnf.col.CodeName())
			if cnf.name != "" {
				suffix = strcase.ToCamel(cnf.name)
			}
			if _, ok := c.factoryBuilder.AllFieldNames()[relatedModelRefName]; ok {
				manyRelated = true
			}
			if !manyRelated {
				if c.factoryBuilder.associatedModelTimes(aModel) > 1 {
					manyRelated = true
				} else {
					relatedFks := cnf.col.RelatedFks()
					if len(relatedFks) > 1 {
						relatedToThis := 0
						for _, relatedFk := range relatedFks {
							if relatedFk.Factory() == c.Factory() {
								relatedToThis++
								if relatedToThis == 2 {
									manyRelated = true
									break
								}
							}
						}
					}
				}
			}
			if manyRelated {
				if cnf.reverse == ReverseType.ToMany {
					reverseName = fmt.Sprintf("%sBy%s", pluralize.Plural(c.Factory().CodeName()), suffix)
				} else {
					reverseName = fmt.Sprintf("%sBy%s", c.Factory().CodeName(), suffix)
				}
				if cnf.toMany {
					relationName = fmt.Sprintf("%sBy%s", pluralize.Plural(aModel.CodeName()), suffix)
				} else {
					relationName = fmt.Sprintf("%sBy%s", aModel.CodeName(), suffix)
				}
			} else {
				if cnf.reverse == ReverseType.ToMany {
					reverseName = pluralize.Plural(c.Factory().CodeName())
				} else {
					reverseName = c.Factory().CodeName()
				}
				if cnf.toMany {
					relationName = pluralize.Plural(aModel.CodeName())
				} else {
					relationName = aModel.CodeName()
				}
			}
			if reverseName == relationName {
				relationName += "1"
				reverseName += "2"
			}
			if aModel == c.Factory() {
				relatedModelRefName = c.Factory().CodeName()
			} else if aModel.SchemaName() != schemaName {
				targetSchemaVar = strcase.ToLowerCamel(aModel.SchemaName())
				targetSchemaName = fmt.Sprintf("%s *%sSchema", targetSchemaVar, aModel.SchemaName())
			}
			if cnf.relationName != "" {
				relationName = cnf.relationName
				reverseName = fmt.Sprintf("%sBy%s", pluralize.Plural(c.Factory().CodeName()), strcase.ToCamel(relationName))
			}
			if cnf.reverseName != "" {
				reverseName = cnf.reverseName
			}
			if cnf.middleTable != nil {
				defCode = append(defCode, fmt.Sprintf("%s *%s.MiddleTable[*%sBody, *Factory%s, %sBody, %sBody]", relationName, pkgName, c.Factory().FullCodeName(), cnf.middleTable.targetFactory.FullCodeName(), cnf.middleTable.targetFactory.FullCodeName(), cnf.middleTable.middleTable.FullCodeName()))
				bodyCode = append(bodyCode, fmt.Sprintf("%s []*%sBody", relationName, cnf.middleTable.targetFactory.FullCodeName()))
				c.setRelationCodes = append(c.setRelationCodes, []string{
					fmt.Sprintf("case %%v: return &m.%s", relationName),
					relationName,
					fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) ([]*%sBody, error) {
	if len(m.%s) == 0 || len(noCache) != 0 && noCache[0] {
		err := %s.%s.%s.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.%s, nil
}`, c.Factory().FullCodeName(), relationName, cnf.middleTable.targetFactory.FullCodeName(), relationName, schemaName, c.Factory().CodeName(), relationName, relationName),
					targetSchemaName,
				})
			} else if cnf.toMany {
				defCode = append(defCode, fmt.Sprintf("%s *%s.ToMany[*%sBody, *Factory%s, %sBody]", relationName, pkgName, c.Factory().FullCodeName(), aModel.FullCodeName(), aModel.FullCodeName()))
				bodyCode = append(bodyCode, fmt.Sprintf("%s []*%sBody", relationName, aModel.FullCodeName()))
				c.setRelationCodes = append(c.setRelationCodes, []string{
					fmt.Sprintf("case %%v: return &m.%s", relationName),
					relationName,
					fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) ([]*%sBody, error) {
	if len(m.%s) == 0 || len(noCache) != 0 && noCache[0] {
		err := %s.%s.%s.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.%s, nil
}`, c.Factory().FullCodeName(), relationName, aModel.FullCodeName(), relationName, schemaName, c.Factory().CodeName(), relationName, relationName),
					targetSchemaName,
				})
			} else {
				defCode = append(defCode, fmt.Sprintf("%s *%s.ToOne[*%sBody, *Factory%s, %sBody]", relationName, pkgName, c.Factory().FullCodeName(), aModel.FullCodeName(), aModel.FullCodeName()))
				bodyCode = append(bodyCode, fmt.Sprintf("%s *%sBody", relationName, aModel.FullCodeName()))
				c.setRelationCodes = append(c.setRelationCodes, []string{
					fmt.Sprintf("case %%v: return &m.%s", relationName),
					relationName,
					fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) (*%sBody, error) {
	if m.%s == nil || len(noCache) != 0 && noCache[0] {
		err := %s.%s.%s.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.%s, nil
}`, c.Factory().FullCodeName(), relationName, aModel.FullCodeName(), relationName, schemaName, c.Factory().CodeName(), relationName, relationName),
					targetSchemaName,
				})
			}
			if cnf.middleTable == nil {
				compositePairCodes := [2]string{}
				for _, compositePair := range cnf.compositeKeys {
					compositePairCodes[0] += fmt.Sprintf(", s.%s.%s, %s.%s.%s", compositePair[0].Factory().CodeName(), compositePair[0].CodeName(), targetSchemaVar, compositePair[1].Factory().CodeName(), compositePair[1].CodeName())
					compositePairCodes[1] += fmt.Sprintf(", %s.%s.%s, s.%s.%s", targetSchemaVar, compositePair[1].Factory().CodeName(), compositePair[1].CodeName(), compositePair[0].Factory().CodeName(), compositePair[0].CodeName())
				}
				settingCodes = append(settingCodes, fmt.Sprintf("s.%s.%s.InitRelation(s.%s.%s, %s.%s.%s%s)", c.Factory().CodeName(), relationName, c.Factory().CodeName(), c.CodeName(), targetSchemaVar, relatedModelRefName, cnf.col.CodeName(), compositePairCodes[0]))
				settingCodes = append(settingCodes, fmt.Sprintf("%s.%s.%s.InitRelation(%s.%s.%s, s.%s.%s%s)", targetSchemaVar, relatedModelRefName, reverseName, targetSchemaVar, relatedModelRefName, cnf.col.CodeName(), c.Factory().CodeName(), c.CodeName(), compositePairCodes[1]))
			} else {
				var (
					lk   = make([]string, len(cnf.middleTable.LocalKeys))
					mtlk = make([]string, len(cnf.middleTable.middleTableLocalRefs))
					fk   = make([]string, len(cnf.middleTable.targetCols))
					mtfk = make([]string, len(cnf.middleTable.middleTableTargetRefs))
				)
				for i, compositePair := range cnf.middleTable.LocalKeys {
					lk[i] = fmt.Sprintf("s.%s.%s", compositePair.Factory().CodeName(), compositePair.CodeName())
				}
				for i, compositePair := range cnf.middleTable.middleTableLocalRefs {
					mtlk[i] = fmt.Sprintf("%s.%s.%s", cnf.middleTable.middleTable.SchemaName(), cnf.middleTable.middleTable.CodeName(), compositePair.CodeName())
				}
				for i, compositePair := range cnf.middleTable.targetCols {
					fk[i] = fmt.Sprintf("%s.%s.%s", cnf.middleTable.targetFactory.SchemaName(), cnf.middleTable.targetFactory.CodeName(), compositePair.CodeName())
				}
				for i, compositePair := range cnf.middleTable.middleTableTargetRefs {
					mtfk[i] = fmt.Sprintf("%s.%s.%s", cnf.middleTable.middleTable.SchemaName(), cnf.middleTable.middleTable.CodeName(), compositePair.CodeName())
				}
				settingCodes = append(settingCodes, fmt.Sprintf(
					"s.%s.%s.InitMtRelation(%s.%s, []preform.IColFromFactory{%s}, []preform.IColFromFactory{%s}, []preform.IColFromFactory{%s}, []preform.IColFromFactory{%s})",
					c.Factory().CodeName(),
					relationName,
					cnf.middleTable.middleTable.SchemaName(),
					cnf.middleTable.middleTable.CodeName(),
					strings.Join(lk, ", "),
					strings.Join(mtlk, ", "),
					strings.Join(fk, ", "),
					strings.Join(mtfk, ", "),
				))
				settingCodes = append(settingCodes, fmt.Sprintf(
					"%s.%s.%s.InitMtRelation(%s.%s, []preform.IColFromFactory{%s}, []preform.IColFromFactory{%s}, []preform.IColFromFactory{%s}, []preform.IColFromFactory{%s})",
					targetSchemaVar,
					cnf.middleTable.targetFactory.CodeName(),
					reverseName,
					cnf.middleTable.middleTable.SchemaName(),
					cnf.middleTable.middleTable.CodeName(),
					strings.Join(fk, ", "),
					strings.Join(mtfk, ", "),
					strings.Join(lk, ", "),
					strings.Join(mtlk, ", "),
				))

			}
			if cnf.cond != nil {
				settingCodes[len(settingCodes)-2] += fmt.Sprintf(".ExtraCond(s.%s.%s)", c.Factory().CodeName(), cnf.cond.ToCondCode())
			}
			if cnf.reverseCond != nil {
				settingCodes[len(settingCodes)-1] += fmt.Sprintf(".ExtraCond(s.%s.%s)", c.Factory().CodeName(), cnf.reverseCond.ToCondCode())
			}

			if cnf.middleTable != nil {
				cnf.middleTable.targetFactory.AddCode(
					[]string{fmt.Sprintf("%s *%s.MiddleTable[*%sBody, *Factory%s, %sBody, %sBody]", reverseName, pkgName, cnf.middleTable.targetFactory.FullCodeName(), c.Factory().FullCodeName(), c.Factory().FullCodeName(), cnf.middleTable.middleTable.FullCodeName())},
					[]string{fmt.Sprintf("%s []*%sBody", reverseName, c.Factory().FullCodeName())},
					nil,
					nil,
					[][]string{{
						fmt.Sprintf("case %%v: return &m.%s", reverseName),
						reverseName,
						fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) ([]*%sBody, error) {
	if len(m.%s) == 0 || len(noCache) != 0 && noCache[0] {
		err := %s.%s.%s.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.%s, nil
}`, cnf.middleTable.targetFactory.FullCodeName(), reverseName, c.Factory().FullCodeName(), reverseName, cnf.middleTable.targetFactory.SchemaName(), cnf.middleTable.targetFactory.CodeName(), reverseName, reverseName),
					}},
				)
			} else {
				switch cnf.reverse {
				case ReverseType.ToMany:
					aModel.AddCode(
						[]string{fmt.Sprintf("%s *%s.ToMany[*%sBody, *Factory%s, %sBody]", reverseName, pkgName, aModel.FullCodeName(), c.Factory().FullCodeName(), c.Factory().FullCodeName())},
						[]string{fmt.Sprintf("%s []*%sBody", reverseName, c.Factory().FullCodeName())},
						nil,
						nil,
						[][]string{{
							fmt.Sprintf("case %%v: return &m.%s", reverseName),
							reverseName,
							fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) ([]*%sBody, error) {
	if len(m.%s) == 0 || len(noCache) != 0 && noCache[0] {
		err := %s.%s.%s.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.%s, nil
}`, aModel.FullCodeName(), reverseName, c.Factory().FullCodeName(), reverseName, aModel.SchemaName(), aModel.CodeName(), reverseName, reverseName),
						}},
					)
				case ReverseType.ToOne:
					aModel.AddCode(
						[]string{fmt.Sprintf("%s *%s.ToOne[*%sBody, *Factory%s, %sBody]", reverseName, pkgName, aModel.FullCodeName(), c.Factory().FullCodeName(), c.Factory().FullCodeName())},
						[]string{fmt.Sprintf("%s *%sBody", reverseName, c.Factory().FullCodeName())},
						nil,
						nil,
						[][]string{{
							fmt.Sprintf("case %%v: return &m.%s", reverseName),
							reverseName,
							fmt.Sprintf(`func (m *%sBody) Load%s(noCache ...bool) (*%sBody, error) {
	if m.%s == nil || len(noCache) != 0 && noCache[0] {
		err := %s.%s.%s.Load(m)
		if err != nil {
			return nil, err
		}
	}
	return m.%s, nil
}`, aModel.FullCodeName(), reverseName, c.Factory().FullCodeName(), reverseName, aModel.SchemaName(), aModel.CodeName(), reverseName, reverseName),
						}},
					)
				}
			}

		}
	}
	setRelationCodesTemp = c.setRelationCodes
	return importPath,
		defCode,
		bodyCode,
		settingCodes,
		""

}

type iForeignKeyDef interface {
	GetSetRelationCodes() [][]string
}

func (c ForeignKeyDef[T]) GetSetRelationCodes() (codes [][]string) {
	return setRelationCodesTemp
}
