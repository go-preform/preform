package dialect

import (
	"database/sql"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"github.com/iancoleman/strcase"
	"strings"
)

type mysqlDialect struct {
	basicSqlDialect
}

func NewMysqlDialect() *mysqlDialect {
	return &mysqlDialect{basicSqlDialect: basicSqlDialect{
		quoteTpl:           "`%s`",
		lastInsertIdMethod: LastInsertIdMethodByRes,
	}}
}

func (d mysqlDialect) GetStructure(db *sql.DB, schemasEmptyIsAll ...string) []*preformShare.Scheme {
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
	schemaQ := squirrel.Select("TABLE_SCHEMA", "TABLE_NAME", "TABLE_COMMENT").From("information_schema.TABLES")
	if len(schemasEmptyIsAll) != 0 {
		schemaQ = schemaQ.Where(squirrel.Eq{"TABLE_SCHEMA": schemasEmptyIsAll})
	}
	schemaQ = schemaQ.Where(squirrel.And{squirrel.Eq{"TABLE_TYPE": "BASE TABLE"}, squirrel.NotEq{"TABLE_SCHEMA": []string{"performance_schema", "information_schema", "mysql"}}})
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

func (d mysqlDialect) getTableDetails(db *sql.DB, tableByName map[string]*preformShare.Table, schemaNames, tableNames []string) {
	var (
		schemaName, tableName, colName, dataType, colComment, extra string
		colDefault, keyName, fkSchemaName, fkTableName, fkName      sql.NullString
		isNullable                                                  string
		ok                                                          bool
		table, fkTable                                              *preformShare.Table
		col, fkCol                                                  *preformShare.Column
		schemas                                                     = map[string]struct{}{}
		fk                                                          *preformShare.ForeignKey
	)
	for _, schemaName = range schemaNames {
		schemas[schemaName] = struct{}{}
	}
	q := squirrel.Select(
		"c.TABLE_SCHEMA",
		"c.TABLE_NAME",
		"c.COLUMN_NAME",
		"if(c.DATA_TYPE='enum',c.COLUMN_TYPE,c.DATA_TYPE)",
		"c.IS_NULLABLE",
		"c.COLUMN_DEFAULT",
		"k.CONSTRAINT_NAME",
		"k.REFERENCED_TABLE_SCHEMA",
		"k.REFERENCED_TABLE_NAME",
		"k.REFERENCED_COLUMN_NAME",
		"c.COLUMN_COMMENT",
		"c.EXTRA",
	).From("information_schema.COLUMNS c").
		LeftJoin("information_schema.KEY_COLUMN_USAGE k ON k.TABLE_SCHEMA = c.TABLE_SCHEMA AND k.TABLE_NAME = c.TABLE_NAME AND k.COLUMN_NAME = c.COLUMN_NAME").
		Where(squirrel.Eq{"c.TABLE_SCHEMA": schemaNames, "c.TABLE_NAME": tableNames}).OrderBy("c.TABLE_SCHEMA", "c.TABLE_NAME", "c.ORDINAL_POSITION").GroupBy("c.TABLE_SCHEMA",
		"c.TABLE_NAME",
		"c.COLUMN_NAME",
		"c.DATA_TYPE",
		"c.COLUMN_TYPE",
		"c.IS_NULLABLE",
		"c.COLUMN_DEFAULT",
		"k.CONSTRAINT_NAME",
		"k.REFERENCED_TABLE_SCHEMA",
		"k.REFERENCED_TABLE_NAME",
		"k.REFERENCED_COLUMN_NAME",
		"c.COLUMN_COMMENT",
		"c.EXTRA", "c.ORDINAL_POSITION")
	rows, err := q.RunWith(db).Query()
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&schemaName, &tableName, &colName, &dataType, &isNullable, &colDefault, &keyName, &fkSchemaName, &fkTableName, &fkName, &colComment, &extra)
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
				if strings.ToLower(extra) == "auto_increment" {
					col.IsAutoKey = true
				}
				col.DefaultValue = colDefault
				if isNullable == "YES" {
					col.Nullable = true
				}
				col.Type = dataType
				col.Comment = colComment
				enums := mysqlCalcGoType(col)
				if enums != nil {
					for k, v := range enums {
						if table.Scheme.Enums == nil {
							table.Scheme.Enums = make(map[string][]string)
						}
						table.Scheme.Enums[k] = v
					}
				}
			}
			if keyName.Valid {
				if keyName.String == "PRIMARY" {
					col.IsPrimaryKey = true
				} else if fkSchemaName.Valid && fkTableName.Valid && fkName.Valid {
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
									break
								}
							}
						}
					}
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

}

func mysqlCalcGoType(col *preformShare.Column) (enum map[string][]string) {
	var (
		t        = col.Type
		nullable = col.Nullable
	)
	if nullable {
		defer func() {
			col.GoType = fmt.Sprintf("preformTypes.Null[%s]", col.GoType)
			col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
			col.IsScanner = true
		}()
	}
	if strings.HasPrefix(t, "enum(") {
		tt := strings.Split(strings.Trim(t[4:], "()"), ",")
		var enumVals []string
		for _, ttt := range tt {
			enumVals = append(enumVals, strings.Trim(strings.Split(ttt, "=")[0], "' "))
		}
		col.GoType = fmt.Sprintf("Enum_%s_%s%s", strcase.ToCamel(col.Table.Scheme.Name), strcase.ToCamel(col.Table.Name), strcase.ToCamel(col.Name))
		enum = map[string][]string{fmt.Sprintf("%s%s", strcase.ToCamel(col.Table.Name), strcase.ToCamel(col.Name)): enumVals}
		return
	}
	switch t {
	case "double":
		col.GoType += "float64"
	case "float":
		col.GoType += "float32"
	case "bigint":
		col.GoType += "int64"
	case "int":
		col.GoType += "int32"
	case "tinyint", "smallint", "mediumint":
		col.GoType += "int16"
	case "date", "datetime", "timestamp", "time":
		col.GoType += "time.Time"
		col.Table.Scheme.Imports[`"time"`] = struct{}{}
		col.Table.Imports[`"time"`] = struct{}{}
	case "blob", "longblob":
		col.GoType += "[]byte"
	case "text", "varchar", "char", "longtext", "mediumtext", "tinytext":
		col.GoType += "string"
	case "decimal", "numeric":
		col.GoType += "preformTypes.Rat"
		col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
		col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
	default:
		col.GoType = "any"
	}
	return
}
