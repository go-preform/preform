package dialect

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	preformTypes "github.com/go-preform/preform/types"
	"github.com/go-preform/squirrel"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
	"time"
)

type postgresqlDialect struct {
	basicSqlDialect
}

func NewPostgresqlDialect() *postgresqlDialect {
	return &postgresqlDialect{basicSqlDialect: basicSqlDialect{
		quoteTpl:           `"%s"`,
		lastInsertIdMethod: LastInsertIdMethodBySuffix,
		lastInsertIdSuffix: func(col string) squirrel.Sqlizer {
			return squirrel.Expr(fmt.Sprintf("RETURNING %s", col))
		},
	}}
}

func (d postgresqlDialect) Aggregate(fn preformShare.Aggregator, body any, params ...any) squirrel.Sqlizer {
	var (
		bodyStr string
		args    []any
	)
	switch body.(type) {
	case string:
		bodyStr = body.(string)
	case preformShare.ICol:
		bodyStr = body.(preformShare.ICol).GetCode()
	case squirrel.Sqlizer:
		s := body.(squirrel.Sqlizer)
		bodyStr, args, _ = s.ToSql()
		bodyStr, args, _ = preformShare.NestSql(bodyStr, args)
	}
	switch fn {
	case AggGroupConcat:
		return squirrel.Expr(fmt.Sprintf("STRING_AGG(%s, ?)", bodyStr), append(args, params[0])...)
	case AggCountDistinct:
		return squirrel.Expr(fmt.Sprintf("COUNT(DISTINCT %s)", bodyStr), args...)
	default:
		if l := len(params); l != 0 {
			return squirrel.Expr(fmt.Sprintf("%s(%s%s)", fn, bodyStr, strings.Repeat(",?", l)), append(args, params...)...)
		}
		return squirrel.Expr(fmt.Sprintf("%s(%s)", fn, bodyStr), args...)
	}
}

func (d postgresqlDialect) GetStructure(db *sql.DB, schemasEmptyIsAll ...string) []*preformShare.Scheme {

	var (
		schemes                                                        []*preformShare.Scheme
		scheme                                                         *preformShare.Scheme
		ok                                                             bool
		schemaByName                                                   = make(map[string]*preformShare.Scheme)
		schemaName, tableName, inheritTable, inheritKey, inheritBounds string
		enumName, enumValue                                            string
		enumsBySchema                                                  = map[string]map[string][]string{}
		enums                                                          map[string][]string
		lastCtName, lastCtSchema, ctName, ctAttrName, ctAttrType       string
		ctAttrTypeSchema                                               sql.NullString
		ctBySchema                                                     = map[string]map[string]*preformShare.CustomType{}
		tableComment                                                   sql.NullString
		table                                                          *preformShare.Table
		tableByName                                                    = make(map[string]*preformShare.Table)
		allSchemas                                                     = []string{}
		allTables                                                      = []string{}
		inherits                                                       = map[string][][3]string{}
		inherited                                                      = map[string]string{}
	)

	//get partitioned tables
	rows, err := squirrel.Select("inhparent.relnamespace::regnamespace::text as schema",
		"inhparent.relname as table_name",
		"inhrel.relname as part_name",
		"pg_get_partkeydef(inhparent.oid) as partition_key",
		"pg_get_expr(inhrel.relpartbound, inhrel.oid) AS bounds").PlaceholderFormat(squirrel.Dollar).
		From("pg_class inhparent").
		Join("pg_inherits AS i on i.inhparent  = inhparent.oid").
		Join("pg_class as inhrel on i.inhrelid = inhrel.oid").
		Where("inhparent.relkind = 'p'").RunWith(db).Query()

	if err != nil {
		panic(err)
	}
	for rows.Next() {
		err = rows.Scan(&schemaName, &tableName, &inheritTable, &inheritKey, &inheritBounds)
		if err != nil {
			panic(err)
		}
		if parentInheritTable, ok := inherited[fmt.Sprintf("%s.%s", schemaName, tableName)]; ok {
			tableName = parentInheritTable
		} else if _, ok = inherits[fmt.Sprintf("%s.%s", schemaName, tableName)]; !ok {
			inherits[fmt.Sprintf("%s.%s", schemaName, tableName)] = [][3]string{}
		}
		inherits[fmt.Sprintf("%s.%s", schemaName, tableName)] = append(inherits[fmt.Sprintf("%s.%s", schemaName, tableName)], [3]string{inheritTable, inheritKey, inheritBounds})
		inherited[fmt.Sprintf("%s.%s", schemaName, inheritTable)] = tableName
	}

	//get enums
	rows, err = squirrel.Select("pn.nspname", "pt.typname", "pe.enumlabel").
		PlaceholderFormat(squirrel.Dollar).
		From("pg_catalog.pg_namespace pn").
		Join("pg_catalog.pg_type pt on pt.typnamespace = pn.oid").
		Join("pg_catalog.pg_enum pe on pe.enumtypid = pt.oid").OrderBy("pt.oid, pe.enumsortorder").RunWith(db).Query()

	if err != nil {
		panic(err)
	}
	for rows.Next() {
		err = rows.Scan(&schemaName, &enumName, &enumValue)
		if err != nil {
			panic(err)
		}
		if enums, ok = enumsBySchema[schemaName]; !ok {
			enums = map[string][]string{}
			enumsBySchema[schemaName] = enums
		}
		if _, ok = enums[enumName]; !ok {
			enums[enumName] = []string{}
		}
		enums[enumName] = append(enums[enumName], enumValue)
	}

	//get types
	rows, err = squirrel.Select("pn.nspname", "pt.typname", "pa.attname", "attT.typname", "attTn.nspname").
		PlaceholderFormat(squirrel.Dollar).
		From("pg_catalog.pg_type pt").
		Join("pg_catalog.pg_attribute pa on pt.typrelid = pa.attrelid").
		Join("pg_catalog.pg_namespace pn ON pn.oid = pt.typnamespace").
		Join("pg_catalog.pg_type attT on attT.\"oid\" = pa.atttypid").
		LeftJoin("pg_catalog.pg_namespace attTn on attTn.oid = attT.typnamespace").
		LeftJoin("information_schema.tables t on t.table_name =pt.typname and t.table_schema =pn.nspname").
		Where(squirrel.And{
			squirrel.Eq{"pt.typcategory": "C"},
			squirrel.NotEq{"pn.nspname": []string{"information_schema", "pg_catalog"}},
			squirrel.Gt{"pa.attnum": 0},
			squirrel.Eq{"t.table_name": nil},
		}).
		OrderBy("pn.nspname", "pt.typname", "pa.attnum").
		RunWith(db).Query()

	if err != nil {
		panic(err)
	}
	for rows.Next() {
		err = rows.Scan(&schemaName, &ctName, &ctAttrName, &ctAttrType, &ctAttrTypeSchema)
		if err != nil {
			panic(err)
		}
		imports := map[string]struct{}{}
		dummyCol := &preformShare.Column{Type: ctAttrType, Nullable: false, Table: &preformShare.Table{Imports: imports, Scheme: &preformShare.Scheme{Imports: imports}}}
		pgCalcGoType(dummyCol, schemaName, map[string]map[string][]string{}, map[string]map[string]*preformShare.CustomType{})
		if ctAttrType != "json" && ctAttrType != "jsonb" && ctAttrType != "_json" && ctAttrType != "_jsonb" {
			for {
				if ctAttrTypeSchema.Valid && ctAttrTypeSchema.String != "pg_catalog" && ctAttrTypeSchema.String != "information_schema" {
					if enums, ok = enumsBySchema[ctAttrTypeSchema.String]; ok {
						if _, ok = enums[ctAttrType]; ok {
							dummyCol.GoType = fmt.Sprintf("%s%s", strcase.ToCamel(ctAttrTypeSchema.String), strcase.ToCamel(ctAttrType))
							break
						}
					}
				}
				if strings.Contains(dummyCol.GoType, "[any]") {
					dummyCol.GoType = strings.Replace(dummyCol.GoType, "[any]", fmt.Sprintf("[%s%s]", strcase.ToCamel(schemaName), strcase.ToCamel(ctAttrType)), 1)
				} else if dummyCol.GoType == "any" {
					dummyCol.GoType = fmt.Sprintf("%s%s", strcase.ToCamel(schemaName), strcase.ToCamel(ctAttrType))
				}
				break
			}
		}
		if schemaName != lastCtSchema || ctName != lastCtName {
			lastCtSchema = schemaName
			lastCtName = ctName
			if _, ok = ctBySchema[schemaName]; !ok {
				ctBySchema[schemaName] = map[string]*preformShare.CustomType{}
			}
			ctBySchema[schemaName][ctName] = &preformShare.CustomType{Name: ctName, Attr: []*preformShare.CustomTypeAttr{{Name: ctAttrName, Type: dummyCol.GoType, NotNull: true, IsScanner: dummyCol.IsScanner}}, Imports: imports}
		} else {
			ctBySchema[schemaName][ctName].Attr = append(ctBySchema[schemaName][ctName].Attr, &preformShare.CustomTypeAttr{Name: ctAttrName, Type: dummyCol.GoType, NotNull: true, IsScanner: dummyCol.IsScanner})
			for imp := range imports {
				ctBySchema[schemaName][ctName].Imports[imp] = struct{}{}
			}
		}
	}

	schemaQ := squirrel.Select("table_schema", "table_name", "obj_description(concat(table_schema, '.', table_name)::regclass)").PlaceholderFormat(squirrel.Dollar).From("information_schema.tables")
	if len(schemasEmptyIsAll) != 0 {
		schemaQ = schemaQ.Where(squirrel.Eq{"table_schema": schemasEmptyIsAll})
	}

	prepareTable := func(schemaName, tableName string, tableComment sql.NullString, isView bool) bool {
		if _, ok = inherited[fmt.Sprintf("%s.%s", schemaName, tableName)]; ok {
			return false
		}
		if scheme, ok = schemaByName[schemaName]; !ok {
			scheme = &preformShare.Scheme{Name: schemaName, Imports: map[string]struct{}{`"github.com/go-preform/preform/preformBuilder"`: {}}, Enums: map[string][]string{}, CustomTypes: map[string]*preformShare.CustomType{}}
			schemaByName[schemaName] = scheme
			if customTypes, ok := ctBySchema[schemaName]; ok {
				scheme.CustomTypes = customTypes
			}
			if enums, ok := enumsBySchema[schemaName]; ok {
				scheme.Enums = enums
			}
			schemes = append(schemes, scheme)
			allSchemas = append(allSchemas, schemaName)
		}
		table = &preformShare.Table{Name: tableName, Scheme: scheme, ColumnByName: make(map[string]*preformShare.Column), Comment: tableComment.String, Imports: map[string]struct{}{}, ForeignKeys: map[string]*preformShare.ForeignKey{}}
		scheme.Tables = append(scheme.Tables, table)
		tableByName[fmt.Sprintf("%s.%s", schemaName, tableName)] = table
		allTables = append(allTables, tableName)
		if inherits[fmt.Sprintf("%s.%s", schemaName, tableName)] != nil {
			table.Inheritors = inherits[fmt.Sprintf("%s.%s", schemaName, tableName)]
		}
		return true
	}

	schemaQ = squirrel.Select("table_schema", "table_name", "obj_description(concat(table_schema, '.', table_name)::regclass)").PlaceholderFormat(squirrel.Dollar).From("information_schema.tables")
	if len(schemasEmptyIsAll) != 0 {
		schemaQ = schemaQ.Where(squirrel.Eq{"table_schema": schemasEmptyIsAll})
	}
	schemaQ = schemaQ.Where(squirrel.And{squirrel.Eq{"table_type": "BASE TABLE"}, squirrel.NotEq{"table_schema": []string{"pg_catalog", "information_schema"}}})
	rows, err = schemaQ.RunWith(db).Query()
	if err != nil {
		q, args, _ := schemaQ.ToSql()
		fmt.Println(q, args)
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&schemaName, &tableName, &tableComment)
		if err != nil {
			panic(err)
		}
		prepareTable(schemaName, tableName, tableComment, false)
	}

	schemaQ = squirrel.Select("table_schema", "table_name").PlaceholderFormat(squirrel.Dollar).From("information_schema.views")
	if len(schemasEmptyIsAll) != 0 {
		schemaQ = schemaQ.Where(squirrel.Eq{"table_schema": schemasEmptyIsAll})
	}
	schemaQ = schemaQ.Where(squirrel.NotEq{"table_schema": []string{"pg_catalog", "information_schema"}})
	rows, err = schemaQ.RunWith(db).Query()
	if err != nil {
		q, args, _ := schemaQ.ToSql()
		fmt.Println(q, args)
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&schemaName, &tableName)
		prepareTable(schemaName, tableName, sql.NullString{}, true)
	}
	d.getTableDetails(db, tableByName, allSchemas, allTables, enumsBySchema, ctBySchema)
	return schemes
}

func (d postgresqlDialect) getTableDetails(db *sql.DB, tableByName map[string]*preformShare.Table, schemaNames, tableNames []string, enumsBySchema map[string]map[string][]string, ctBySchema map[string]map[string]*preformShare.CustomType) {
	var (
		schemaName, tableName, colName, dataType, dataTypeSchema           string
		colDefault, keyName, fkSchemaName, fkTableName, fkName, colComment sql.NullString
		isNullable                                                         string
		ok                                                                 bool
		table, fkTable                                                     *preformShare.Table
		col, fkCol                                                         *preformShare.Column
		keyPos, fkPos                                                      sql.NullInt64
		schemas                                                            = map[string]struct{}{}
		constraintType                                                     sql.NullString
		fk                                                                 *preformShare.ForeignKey
	)
	for _, schemaName = range schemaNames {
		schemas[schemaName] = struct{}{}
	}
	q := squirrel.Select(
		"c.table_schema",
		"c.table_name",
		"c.column_name",
		"c.udt_name",
		"c.udt_schema",
		"c.is_nullable",
		"c.column_default",
		"k.constraint_name",
		"ccu.table_schema",
		"ccu.table_name",
		"ccu.column_name",
		"pg_catalog.col_description(concat(c.table_schema, '.', c.table_name)::regclass,c.dtd_identifier::int) \"comment\"",
		"k.ordinal_position",
		"k.position_in_unique_constraint",
		"tc.constraint_type",
	).PlaceholderFormat(squirrel.Dollar).From("information_schema.columns c").
		LeftJoin("information_schema.key_column_usage k ON k.table_schema = c.table_schema AND k.table_name = c.table_name AND k.column_name = c.column_name").
		LeftJoin("information_schema.constraint_column_usage ccu ON ccu.constraint_schema = c.table_schema AND ccu.constraint_name = k.constraint_name").
		LeftJoin("information_schema.table_constraints tc ON tc.constraint_schema = c.table_schema AND tc.constraint_name = k.constraint_name").
		Where(squirrel.Eq{"c.table_schema": schemaNames, "c.table_name": tableNames}).
		GroupBy("c.table_schema",
			"c.table_name",
			"c.column_name",
			"c.udt_name",
			"c.udt_schema",
			"c.is_nullable",
			"c.column_default",
			"k.constraint_name",
			"ccu.table_schema",
			"ccu.table_name",
			"ccu.column_name",
			"\"comment\"",
			"k.ordinal_position",
			"k.position_in_unique_constraint", "c.ordinal_position", "tc.constraint_type").
		OrderBy("c.table_schema", "c.table_name", "c.ordinal_position")
	rows, err := q.RunWith(db).Query()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&schemaName, &tableName, &colName, &dataType, &dataTypeSchema, &isNullable, &colDefault, &keyName, &fkSchemaName, &fkTableName, &fkName, &colComment, &keyPos, &fkPos, &constraintType)
		if err != nil {
			panic(err)
		}
		if table, ok = tableByName[fmt.Sprintf("%s.%s", schemaName, tableName)]; ok {
			if col, ok = table.ColumnByName[colName]; !ok {
				col = &preformShare.Column{Name: colName, Table: table}
				table.Columns = append(table.Columns, col)
				table.ColumnByName[colName] = col
			}
			if col.Type == "" {
				if strings.HasPrefix(colDefault.String, "nextval(") || strings.HasPrefix(colDefault.String, "gen_random_uuid()") {
					col.IsAutoKey = true
				} else {
					col.DefaultValue = colDefault
				}
				if isNullable == "YES" {
					col.Nullable = true
				}
				col.Type = dataType
				col.Comment = colComment.String
				pgCalcGoType(col, dataTypeSchema, enumsBySchema, ctBySchema)
			}
			if keyName.Valid {
				if fkTableName.String == tableName && fkName.String == colName && fkSchemaName.String == schemaName {
					if constraintType.String == "PRIMARY KEY" {
						col.IsPrimaryKey = true
						col.PkPos = keyPos.Int64
					}
				} else if constraintType.String == "FOREIGN KEY" {
					if _, ok = schemas[fkSchemaName.String]; ok {
						if fkTable, ok = tableByName[fmt.Sprintf("%s.%s", fkSchemaName.String, fkTableName.String)]; ok {
							if fkCol, ok = fkTable.ColumnByName[fkName.String]; !ok {
								fkCol = &preformShare.Column{Name: fkName.String, Table: fkTable}
								fkTable.Columns = append(fkTable.Columns, fkCol)
								fkTable.ColumnByName[fkName.String] = fkCol
							}
							if fk, ok = table.ForeignKeys[keyName.String]; !ok {
								fk = &preformShare.ForeignKey{Name: keyName.String}
								table.ForeignKeys[keyName.String] = fk
								col.ForeignKeys = append(col.ForeignKeys, fk)
							}
							if len(fk.LocalKeys) == 0 {
								fk.LocalKeys = append(fk.LocalKeys, col)
								fk.ForeignKeys = append(fk.ForeignKeys, fkCol)
								fkCol.ForeignKeys = append(fkCol.ForeignKeys, fk)
							} else if fk.LocalKeys[len(fk.LocalKeys)-1] != col && fk.ForeignKeys[len(fk.ForeignKeys)-1] != fkCol {
							checkComposite:
								for {
									for i := range fk.LocalKeys {
										if fk.LocalKeys[i] == col {
											break checkComposite
										}
									}
									for i := range fk.ForeignKeys {
										if fk.ForeignKeys[i] == fkCol {
											break checkComposite
										}
									}
									fk.LocalKeys = append(fk.LocalKeys, col)
									fk.ForeignKeys = append(fk.ForeignKeys, fkCol)
									fkCol.ForeignKeys = append(fkCol.ForeignKeys, fk)
									break
								}
							}
						}
					}
				}
			}
			if colComment.Valid {
				fkParts := strings.Split(colComment.String, ";")
			loopFkParts:
				for i, fkPart := range fkParts {
					if strings.HasPrefix(fkPart, "fk:") {
						settingParts := strings.Split(fkPart, ":")
						parts := strings.Split(settingParts[1], ".")
						if len(parts) == 2 {
							parts = append([]string{schemaName}, parts...)
						} else if len(parts) != 3 {
							fmt.Println("ignore illegal fk comment:", colComment.String)
							continue
						}
						if _, ok = schemas[parts[0]]; ok {
							if fkTable, ok = tableByName[fmt.Sprintf("%s.%s", parts[0], parts[1])]; ok {
								if fkCol, ok = fkTable.ColumnByName[parts[2]]; !ok {
									fkCol = &preformShare.Column{Name: parts[2], Table: fkTable}
									fkTable.Columns = append(fkTable.Columns, fkCol)
									fkTable.ColumnByName[parts[2]] = fkCol
								}
								if _, ok = table.ForeignKeys[fmt.Sprintf("comment_%s_%d", col.Name, i)]; ok {
									continue loopFkParts
								} else {
									fk = &preformShare.ForeignKey{Name: fmt.Sprintf("comment_%s_%d", col.Name, i)}
									table.ForeignKeys[fk.Name] = fk
									col.ForeignKeys = append(col.ForeignKeys, fk)
								}
								fk.LocalKeys = append(fk.LocalKeys, col)
								fk.ForeignKeys = append(fk.ForeignKeys, fkCol)
								if len(settingParts) == 4 {
									fk.RelationName, fk.ReverseName = strcase.ToCamel(settingParts[2]), strcase.ToCamel(settingParts[3])
								} else if len(settingParts) == 3 {
									fk.RelationName = strcase.ToCamel(settingParts[2])
								}
							}
						}
					}
				}

			}
		}
	}

}

func pgCalcGoType(col *preformShare.Column, schemaName string, enumsBySchema map[string]map[string][]string, ctBySchema map[string]map[string]*preformShare.CustomType) {
	var (
		t = col.Type
	)
	if col.Nullable {
		defer func() {
			col.GoType = fmt.Sprintf("preformTypes.Null[%s]", col.GoType)
			col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.IsScanner = true
		}()
	}
	if strings.HasPrefix(t, "_") {
		//switch t {
		//case "_int2", "_int4", "_int8", "_varchar", "_text", "_char", "_float4", "_float8":
		//	if !col.Nullable {
		//		col.GoType = "[]"
		//		goto gogo
		//	}
		//}
		defer func() {
			col.GoType = fmt.Sprintf("preformTypes.Array[%s]", col.GoType)
			col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.IsScanner = true
		}()
		//gogo:
		t = t[1:]
	}

	if strings.HasPrefix(t, "int") {
		col.GoType += "int"
		switch strings.Replace(t, "int", "", 1) {
		case "2":
			col.GoType += "16"
		case "4":
			col.GoType += "32"
		default:
			col.GoType += "64"
		}
	} else if strings.HasPrefix(t, "float") {
		col.GoType += "float"
		switch strings.Replace(t, "float", "", 1) {
		case "4":
			col.GoType += "32"
		default:
			col.GoType += "64"
		}
	} else if strings.HasPrefix(t, "time") {
		col.GoType += "time.Time"
		col.Table.Scheme.Imports[`"time"`] = struct{}{}
		col.Table.Imports[`"time"`] = struct{}{}
	} else {
		switch t {
		case "date":
			col.GoType += "time.Time"
			col.Table.Scheme.Imports[`"time"`] = struct{}{}
			col.Table.Imports[`"time"`] = struct{}{}
		case "bool":
			col.GoType += "bool"
		case "bytea":
			col.GoType += "[]byte"
		case "text", "varchar", "char":
			col.GoType += "string"
		case "uuid":
			col.GoType += "uuid.UUID"
			col.Table.Scheme.Imports[`"github.com/satori/go.uuid"`] = struct{}{}
			col.Table.Imports[`"github.com/satori/go.uuid"`] = struct{}{}
			col.IsScanner = true
		case "jsonb", "json":
			col.GoType += "preformTypes.JsonRaw[any]"
			col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.IsScanner = true
		case "numeric":
			col.GoType += "preformTypes.Rat"
			col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.IsScanner = true
		default:
			if enums, ok := enumsBySchema[schemaName]; ok {
				if _, ok := enums[t]; ok {
					col.GoType = fmt.Sprintf("Enum_%s_%s", strcase.ToCamel(schemaName), strcase.ToCamel(t))
					return
				}
			}
			if customTypes, ok := ctBySchema[schemaName]; ok {
				if _, ok := customTypes[t]; ok {
					col.GoType = fmt.Sprintf("CustomType_%s_%s", strcase.ToCamel(schemaName), strcase.ToCamel(t))
					col.IsScanner = true
					return
				}
			}
			col.GoType = "any"
		}
	}
}

func (d postgresqlDialect) ParseCustomTypeScan(src any) (dst []string, err error) {
	if src == nil {
		return
	}
	var (
		inputData []byte
	)
	switch src.(type) {
	case string:
		inputData = []byte(src.(string))
	case []byte:
		inputData = src.([]byte)
	}
	if len(inputData) < 3 {
		return
	}
	dst, err = csv.NewReader(bytes.NewReader(inputData[1 : len(inputData)-1])).Read()
	return
}

func (d postgresqlDialect) ParseCustomTypeValue(typeName string, values ...any) (dst string, err error) {
	var (
		valuer driver.Valuer
		ok     bool
		v      driver.Value
		strs   = make([]string, len(values))
	)
	for i, value := range values {
		if valuer, ok = value.(driver.Valuer); ok {
			v, err = valuer.Value()
			if err != nil {
				return
			}
			strs[i] = fmt.Sprintf("%v", v)
			if strings.Contains(strs[i], "::") {
				strs[i] = strings.Split(strs[i], "::")[0][1:]
				strs[i] = strs[i][:len(strs[i])-1]
			}
		} else {
			switch value.(type) {
			case time.Time:
				strs[i] = value.(time.Time).Format("2006-01-02T15:04:05Z")
			default:
				strs[i] = fmt.Sprintf("%v", value)
			}
		}
		strs[i] = fmt.Sprintf(`"%s"`, strings.Replace(strs[i], `"`, `""`, -1))
	}
	return fmt.Sprintf(`(%s)`, strings.Join(strs, ",")), nil
}

type isJson interface {
	typeExported
	MarshalJSON() ([]byte, error)
	Src() []byte
	String() string
}

type typeExported interface {
	TypeForExport() any
}

func (d postgresqlDialect) CaseStmtToSql(builder squirrel.CaseBuilder, col preformShare.ICol) (string, []any, error) {
	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return sqlStr, args, err
	}
	var (
		v         = col.NewValue()
		wrapTypes func(v any)
	)
	wrapTypes = func(v any) {
		switch v.(type) {
		case time.Time:
			sqlStr = fmt.Sprintf("(%s)::timestamp", sqlStr)
			for i := range args {
				switch args[i].(type) {
				case time.Time:
					args[i] = args[i].(time.Time).Format(time.RFC3339)
				}
			}
		case int32, int64, int, uint32, uint64, uint, float32, float64, preformTypes.Rat:
			sqlStr = fmt.Sprintf("(%s)::numeric", sqlStr)
			for i := range args {
				args[i] = fmt.Sprintf("%v", args[i])
			}
		case isJson:
			sqlStr = fmt.Sprintf("(%s)::json", sqlStr)
		default:
			vt := reflect.TypeOf(v)
			switch vt.Kind() {
			case reflect.Slice, reflect.Array:
				switch vt.Elem().Kind() {
				case reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Float32, reflect.Float64:
					sqlStr = fmt.Sprintf("(%s)::_numeric", sqlStr)
				case reflect.String:
					sqlStr = fmt.Sprintf("(%s)::_text", sqlStr)
				case reflect.Struct:
					if vt.Elem().String() == "time.Time" {
						sqlStr = fmt.Sprintf("(%s)::_timestamp", sqlStr)
					} else if vt.Elem().String() == "preformTypes.Rat" {
						sqlStr = fmt.Sprintf("(%s)::_numeric", sqlStr)
					}
				}
				for i := range args {
					if vv, ok := args[i].(driver.Valuer); ok {
						if vvv, e := vv.Value(); e == nil {
							args[i] = fmt.Sprintf("%v", vvv)
						}
					} else {
						args[i] = fmt.Sprintf("%v", args[i])
					}
				}
			default:
				if e, ok := v.(typeExported); ok {
					wrapTypes(e.TypeForExport())
				}
			}
		}
	}
	wrapTypes(v)

	return sqlStr, args, err
}
