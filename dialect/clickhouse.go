package dialect

import (
	"database/sql"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
	"time"
)

type clickhouseDialect struct {
	basicSqlDialect
}

func NewClickhouseDialect() *clickhouseDialect {
	return &clickhouseDialect{basicSqlDialect: basicSqlDialect{
		quoteTpl:           "`%s`",
		lastInsertIdMethod: LastInsertIdMethodNone,
	}}
}

func (d clickhouseDialect) GetStructure(db *sql.DB, schemasEmptyIsAll ...string) []*preformShare.Scheme {
	var (
		schemes               []*preformShare.Scheme
		scheme                *preformShare.Scheme
		ok                    bool
		schemaByName          = make(map[string]*preformShare.Scheme)
		schemaName, tableName string
		tableComment          sql.NullString
		table                 *preformShare.Table
		tableByName           = make(map[string]*preformShare.Table)
		allSchemas            = []string{}
		allTables             = []string{}
	)
	schemaQ := squirrel.Select("database", "name", "comment").
		From("system.tables")
	if len(schemasEmptyIsAll) != 0 {
		schemaQ = schemaQ.Where(squirrel.Eq{"database": schemasEmptyIsAll})
	}
	schemaQ = schemaQ.Where(squirrel.And{squirrel.Eq{"is_temporary": "0"}, squirrel.NotEq{"database": []string{"informationSchema_columns", "system", "INFORMATION_SCHEMA"}}})

	rows, err := schemaQ.RunWith(db).Query()
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
		if scheme, ok = schemaByName[schemaName]; !ok {
			scheme = &preformShare.Scheme{Name: schemaName, Imports: map[string]struct{}{`"github.com/go-preform/preform/preformBuilder"`: {}}}
			schemaByName[schemaName] = scheme
			schemes = append(schemes, scheme)
			allSchemas = append(allSchemas, schemaName)
		}
		table = &preformShare.Table{Name: tableName, Scheme: scheme, ColumnByName: make(map[string]*preformShare.Column), Comment: tableComment.String, Imports: map[string]struct{}{}, ForeignKeys: map[string]*preformShare.ForeignKey{}}
		scheme.Tables = append(scheme.Tables, table)
		tableByName[fmt.Sprintf("%s.%s", schemaName, tableName)] = table
		allTables = append(allTables, tableName)
	}
	d.getTableDetails(db, tableByName, allSchemas, allTables)
	return schemes
}

func (d clickhouseDialect) getTableDetails(db *sql.DB, tableByName map[string]*preformShare.Table, schemaNames, tableNames []string) {
	var (
		schemaName, tableName, colName, dataType, colDefault, colComment string
		//numPrecision, numScale                                           sql.NullInt64
		isPk           uint8
		ok             bool
		table, fkTable *preformShare.Table
		col, fkCol     *preformShare.Column
		schemas        = map[string]struct{}{}
		fk             *preformShare.ForeignKey
	)
	for _, schemaName = range schemaNames {
		schemas[schemaName] = struct{}{}
	}
	q := squirrel.Select(
		"c.database",
		"c.table",
		"c.name",
		"c.type",
		"c.is_in_primary_key",
		"c.default_expression",
		//"c.numeric_precision",
		//"c.numeric_scale",
		"c.comment",
	).PlaceholderFormat(squirrel.Dollar).From("system.columns c").
		Where(squirrel.Eq{"c.database": schemaNames, "c.table": tableNames}).OrderBy("c.database", "c.table", "c.position")
	rows, err := q.RunWith(db).Query()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(
			&schemaName,
			&tableName,
			&colName,
			&dataType,
			&isPk,
			&colDefault,
			//&numPrecision,
			//&numScale,
			&colComment,
		)
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
				if colDefault != "" {
					col.DefaultValue = sql.NullString{String: colDefault, Valid: colDefault != ""}
				}
				if strings.HasPrefix(dataType, "Nullable") {
					col.Nullable = true
				}
				col.Type = dataType
				col.Comment = colComment
				enums := clickhouseCalcGoType(col)
				if enums != nil {
					for k, v := range enums {
						if table.Scheme.Enums == nil {
							table.Scheme.Enums = make(map[string][]string)
						}
						table.Scheme.Enums[k] = v
					}
				}
			}
			if isPk != 0 {
				col.IsPrimaryKey = true
			}
		}

		if colComment != "" {
			fkParts := strings.Split(colComment, ";")
		loopFkParts:
			for i, fkPart := range fkParts {
				if strings.HasPrefix(fkPart, "fk:") {
					settingParts := strings.Split(fkPart, ":")
					parts := strings.Split(settingParts[1], ".")
					if len(parts) == 2 {
						parts = append([]string{schemaName}, parts...)
					} else if len(parts) != 3 {
						fmt.Println("ignore illegal fk comment:", colComment)
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

func (d clickhouseDialect) UpdateSqlizer(builder preformShare.UpdateBuilder) (string, []any, error) {
	q, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}
	parts := strings.Split(q, " SET ")
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid update query: %s", q)
	}
	return fmt.Sprintf("ALTER TABLE %s UPDATE %s", parts[0][6:], strings.Join(parts[1:], " SET ")), args, nil
}

func (d clickhouseDialect) DeleteSqlizer(builder preformShare.DeleteBuilder) (string, []any, error) {
	return builder.ToSql() //  Lightweight Deletes
	//q, args, err := builder.ToSql()
	//if err != nil {
	//	return "", nil, err
	//}
	//parts := strings.Split(q, " WHERE ")
	//if len(parts) < 2 {
	//	return "", nil, fmt.Errorf("invalid update query: %s", q)
	//}
	//return fmt.Sprintf("ALTER TABLE %s DELETE %s", parts[0][11:], strings.Join(parts[1:], " WHERE ")), args, nil
}

func clickhouseCalcGoType(col *preformShare.Column) (enums map[string][]string) {
	var (
		t        = col.Type
		nullable = col.Nullable
	)
	if nullable {
		defer func() {
			if col.GoType == "uuid.UUID" {
				col.GoType = "uuid.NullUUID"
			} else {
				col.GoType = fmt.Sprintf("*%s", col.GoType)
			}
		}()
		t = t[9:]
		t = t[:len(t)-1]
	}
	//driver not supported
	//if strings.HasPrefix(t, "Nested") {
	//	t = t[7 : len(t)-1]
	//	tt := strings.Split(t, ", ")
	//	ct := &preformShare.CustomType{Name: fmt.Sprintf("%s%s", strcase.ToCamel(col.Table.Name), strcase.ToCamel(col.Name)), Imports: map[string]struct{}{}}
	//	for _, ttt := range tt {
	//		tttt := strings.Split(ttt, " ")
	//		if len(tttt) < 2 {
	//			fmt.Println("ignore illegal nested type:", ttt)
	//			continue
	//		}
	//		dummyCol := &preformShare.Column{Type: tttt[1], Nullable: false, Table: &preformShare.Table{Imports: ct.Imports, Scheme: &preformShare.Scheme{Imports: ct.Imports}}}
	//		clickhouseCalcGoType(dummyCol)
	//		ct.Attr = append(ct.Attr, &preformShare.CustomTypeAttr{Name: tttt[0], Type: dummyCol.GoType, NotNull: !dummyCol.Nullable, IsScanner: dummyCol.IsScanner})
	//	}
	//	if col.Table.Scheme.CustomTypes == nil {
	//		col.Table.Scheme.CustomTypes = make(map[string]*preformShare.CustomType)
	//	}
	//	col.Table.Scheme.CustomTypes[ct.Name] = ct
	//	col.GoType = fmt.Sprintf("CustomType_%s_%s", strcase.ToCamel(col.Table.Scheme.Name), ct.Name)
	//} else {
	col.GoType, enums = clickhouseCalcGoTypeFromString(col, t)
	//}
	return
}

func clickhouseCalcGoTypeFromString(col *preformShare.Column, t string) (goType string, enum map[string][]string) {
	if strings.HasPrefix(t, "Array") {
		defer func() {
			goType = fmt.Sprintf("[]%s", goType)
		}()
		t = t[6:]
		t = t[:len(t)-1]
	}
	if strings.HasPrefix(t, "Enum") {
		tt := strings.Split(strings.Trim(t[4:], "()816"), ",")
		var enumVals []string
		for _, ttt := range tt {
			enumVals = append(enumVals, strings.Trim(strings.Split(ttt, "=")[0], "' "))
		}
		goType = fmt.Sprintf("Enum_%s_%s%s", strcase.ToCamel(col.Table.Scheme.Name), strcase.ToCamel(col.Table.Name), strcase.ToCamel(col.Name))
		enum = map[string][]string{fmt.Sprintf("%s%s", strcase.ToCamel(col.Table.Name), strcase.ToCamel(col.Name)): enumVals}
	} else if strings.HasPrefix(t, "Int") || strings.HasPrefix(t, "UInt") {
		goType += strings.ToLower(t)
	} else if strings.HasPrefix(t, "Float") {
		goType += strings.ToLower(t)
	} else if strings.HasPrefix(t, "Date") {
		goType += "time.Time"
		col.Table.Scheme.Imports[`"time"`] = struct{}{}
		col.Table.Imports[`"time"`] = struct{}{}
	} else if strings.HasPrefix(t, "Decimal") {
		goType += "preformTypes.Rat"
		col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
		col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
	} else if strings.HasPrefix(t, "Map") {
		t = t[4:]
		t = t[:len(t)-1]
		var (
			e1, e2 map[string][]string
		)
		parts := strings.Split(t, ", ")
		parts[0], e1 = clickhouseCalcGoTypeFromString(col, parts[0])
		parts[1], e2 = clickhouseCalcGoTypeFromString(col, parts[1])
		if e1 != nil {
			enum = make(map[string][]string)
			for k, v := range e1 {
				enum[k+"MapKey"] = v
			}
		}
		if e2 != nil {
			if enum == nil {
				enum = make(map[string][]string)
			}
			for k, v := range e2 {
				enum[k+"MapVal"] = v
			}
		}

		goType += fmt.Sprintf("map[%s]%s", parts[0], parts[1])
	} else if strings.HasPrefix(t, "FixedString") {
		goType += "string"
	} else {
		switch t {
		case "Date":
			goType += "time.Time"
			col.Table.Scheme.Imports[`"time"`] = struct{}{}
			col.Table.Imports[`"time"`] = struct{}{}
		case "Bool":
			goType += "bool"
		case "String":
			goType += "string"
		case "UUID":
			goType += "uuid.UUID"
			col.Table.Scheme.Imports[`"github.com/satori/go.uuid"`] = struct{}{}
			col.Table.Imports[`"github.com/satori/go.uuid"`] = struct{}{}
		case "Object('json')":
			goType += "preformTypes.JsonRaw[any]"
			col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
		default:
			fmt.Println("unknown type", t)
			goType = "any"
		}
	}
	return
}

func (d clickhouseDialect) ValueParsers() map[reflect.Type]any {
	var (
		nilTPtr **time.Time
	)
	return map[reflect.Type]any{
		reflect.TypeOf(time.Time{}): func(dbType string, careZero bool) func(v *time.Time) any {
			switch dbType {
			case "Date":
				if careZero {
					return func(v *time.Time) any {
						if v.IsZero() {
							return preformShare.DEFAULT_VALUE
						}
						return v.Format("2006-01-02")
					}
				}
				return func(v *time.Time) any {
					return v.Format("2006-01-02") //v.Format("2006-01-02 15:04:05.000")
				}
			case "DateTime64":
				if careZero {
					return func(v *time.Time) any {
						if v.IsZero() {
							return preformShare.DEFAULT_VALUE
						}
						return v.Format("2006-01-02 15:04:05.000")
					}
				}
				return func(v *time.Time) any {
					return v.Format("2006-01-02 15:04:05.000") //v.Format("2006-01-02 15:04:05.000")
				}
			default:
				if careZero {
					return func(v *time.Time) any {
						if v.IsZero() {
							return preformShare.DEFAULT_VALUE
						}
						return v.Format("2006-01-02 15:04:05")
					}
				}
				return func(v *time.Time) any {
					return v.Format("2006-01-02 15:04:05") //v.Format("2006-01-02 15:04:05.000")
				}
			}
		},
		reflect.TypeOf(&time.Time{}): func(dbType string, careZero bool) func(v **time.Time) any {
			switch dbType {
			case "Date":
				if careZero {
					return func(v **time.Time) any {
						if v == nilTPtr || *v == nil {
							return nil
						}
						if (*v).IsZero() {
							return preformShare.DEFAULT_VALUE
						}
						return (*v).Format("2006-01-02")
					}
				}
				return func(v **time.Time) any {
					if v == nilTPtr || *v == nil {
						return nil
					}
					return (*v).Format("2006-01-02")
				}
			case "DateTime64":
				if careZero {
					return func(v **time.Time) any {
						if v == nilTPtr || *v == nil {
							return nil
						}
						if (*v).IsZero() {
							return preformShare.DEFAULT_VALUE
						}
						return (*v).Format("2006-01-02 15:04:05.000")
					}
				}
				return func(v **time.Time) any {
					if v == nilTPtr || *v == nil {
						return nil
					}
					return (*v).Format("2006-01-02 15:04:05.000")
				}
			default:
				if careZero {
					return func(v **time.Time) any {
						if v == nilTPtr || *v == nil {
							return nil
						}
						if (*v).IsZero() {
							return preformShare.DEFAULT_VALUE
						}
						return (*v).Format("2006-01-02 15:04:05")
					}
				}
				return func(v **time.Time) any {
					if v == nilTPtr || *v == nil {
						return nil
					}
					return (*v).Format("2006-01-02 15:04:05")
				}
			}
		},
	}
}
