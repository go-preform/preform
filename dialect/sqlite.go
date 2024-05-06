package dialect

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"github.com/iancoleman/strcase"
	"regexp"
	"strings"
)

type sqliteDialect struct {
	basicSqlDialect
}

func NewSqliteDialect() *sqliteDialect {
	return &sqliteDialect{basicSqlDialect: basicSqlDialect{
		quoteTpl:           `"%s"`,
		lastInsertIdMethod: LastInsertIdMethodByRes,
	}}
}

func (d sqliteDialect) DefaultValueExpr() squirrel.Sqlizer {
	return nil
}

func (d sqliteDialect) GetStructure(db *sql.DB, schemasEmptyIsAll ...string) []*preformShare.Scheme {

	var (
		schemes                         []*preformShare.Scheme
		scheme                          *preformShare.Scheme
		schemaName, tableName, tableSql string
		table                           *preformShare.Table
		tableByName                     = make(map[string]*preformShare.Table)
		allTables                       = []string{}
	)
	if len(schemasEmptyIsAll) == 0 {
		schemasEmptyIsAll = append(schemasEmptyIsAll, "main")
	}
	for _, schemaName = range schemasEmptyIsAll {
		rows, err := db.Query(fmt.Sprintf("SELECT tbl_name,sql FROM %s.sqlite_master where \"type\"='table';", schemaName))
		if err != nil {
			panic(err)
		}
		scheme = &preformShare.Scheme{Name: schemaName, Imports: map[string]struct{}{`"github.com/go-preform/preform/preformBuilder"`: {}}}
		schemes = append(schemes, scheme)
		for rows.Next() {
			err = rows.Scan(&tableName, &tableSql)
			if err != nil {
				panic(err)
			}
			switch tableName {
			case "sqlite_sequence":
				continue
			}
			table = &preformShare.Table{Name: tableName, Scheme: scheme, Imports: map[string]struct{}{}, Sql: tableSql, ForeignKeys: map[string]*preformShare.ForeignKey{}, ColumnByName: map[string]*preformShare.Column{}}
			tableByName[fmt.Sprintf("%s.%s", schemaName, tableName)] = table
			allTables = append(allTables, tableName)
			scheme.Tables = append(scheme.Tables, table)

		}
		_ = rows.Close()
	}
	d.getTableDetails(db, tableByName, schemasEmptyIsAll, allTables)
	return schemes
}

func (d sqliteDialect) getTableDetails(db *sql.DB, tableByName map[string]*preformShare.Table, schemaNames, tableNames []string) {
	var (
		dummyInt, fkId, isPk                                                                    int
		schemaName, colName, fkColName, dataType, dataTypeSchema, dummyStr, fkName, fkTableName string
		colDefault, colComment                                                                  sql.NullString
		notNullable                                                                             int
		ok                                                                                      bool
		table, fkTable                                                                          *preformShare.Table
		col, fkCol                                                                              *preformShare.Column
		schemas                                                                                 = map[string]struct{}{}
		fk                                                                                      *preformShare.ForeignKey
	)
	for _, schemaName = range schemaNames {
		schemas[schemaName] = struct{}{}
	}
	for _, table = range tableByName {
		tableSql := regexp.MustCompile("[\\s\\t]+").ReplaceAllString(strings.ToUpper(table.Sql), " ")
		rows, err := db.Query(fmt.Sprintf("PRAGMA \"%s\".table_info(\"%s\")", table.Scheme.Name, table.Name))
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			err = rows.Scan(&dummyInt, &colName, &dataType, &notNullable, &colDefault, &isPk)
			if err != nil {
				panic(err)
			}
			if col, ok = table.ColumnByName[colName]; !ok {
				col = &preformShare.Column{Name: colName, Table: table}
				table.Columns = append(table.Columns, col)
				table.ColumnByName[colName] = col
			}
			if col.Type == "" {
				if isPk == 1 {
					col.IsPrimaryKey = true
					if dataType == "INTEGER" && strings.Contains(tableSql, "PRIMARY KEY AUTOINCREMENT") {
						col.IsAutoKey = true
					}
				}
				if colDefault.Valid {
					col.DefaultValue = colDefault
				}
				if notNullable == 0 {
					col.Nullable = true
				}
				col.Type = dataType
				if matches := regexp.MustCompile(fmt.Sprintf("%s\\\"* [^\\r\\n]+[\\s\\t]*--([^\\r\\n]+)[\\r\\n]", colName)).FindStringSubmatch(table.Sql); len(matches) > 0 {
					col.Comment = matches[1]
					parts := strings.Split(col.Comment, ";")
				loopCommentParts:
					for i, part := range parts {
						if strings.HasPrefix(part, "type:") {
							col.Type = strings.Split(part, ":")[1]
						} else if strings.HasPrefix(part, "fk:") {
							settingParts := strings.Split(part, ":")
							parts := strings.Split(settingParts[1], ".")
							if len(parts) == 2 {
								parts = append([]string{table.Scheme.Name}, parts...)
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
										continue loopCommentParts
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
				sqliteCalcGoType(col, dataTypeSchema)
			}

		}
		rows, err = db.Query(fmt.Sprintf("PRAGMA \"%s\".foreign_key_list(\"%s\")", table.Scheme.Name, table.Name))
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			err = rows.Scan(&fkId, &dummyInt, &fkTableName, &colName, &fkColName, &dummyStr, &dummyStr, &dummyStr)
			if err != nil {
				panic(err)
			}
			if col, ok = table.ColumnByName[colName]; ok {
				if fkTable, ok = tableByName[fmt.Sprintf("%s.%s", table.Scheme.Name, fkTableName)]; ok {
					fkName = fmt.Sprintf("fk_%s_%s_%s", table.Name, colName, fkColName)
					if fkCol, ok = fkTable.ColumnByName[fkColName]; !ok {
						fkCol = &preformShare.Column{Name: fkColName, Table: fkTable}
						fkTable.Columns = append(fkTable.Columns, fkCol)
						fkTable.ColumnByName[fkColName] = fkCol
					}
					if fk, ok = table.ForeignKeys[fkName]; !ok {
						fk = &preformShare.ForeignKey{Name: fkName}
						table.ForeignKeys[fk.Name] = fk
						col.ForeignKeys = append(col.ForeignKeys, fk)
					}

					fk.LocalKeys = append(fk.LocalKeys, col)
					fk.ForeignKeys = append(fk.ForeignKeys, fkCol)
				}
			}

		}
	}

}

func sqliteCalcGoType(col *preformShare.Column, schemaName string) {
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
	switch t {
	case "date", "datetime":
		col.GoType += "preformTypes.SqliteTime"
		col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
		col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
	case "BLOB":
		col.GoType += "[]byte"
	case "INTEGER", "INT":
		col.GoType += "int64"
	case "TEXT":
		col.GoType += "string"
	case "jsonb", "json":
		col.GoType += "preformTypes.JsonRaw[any]"
		col.Table.Scheme.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
		col.Table.Imports[`"github.com/go-preform/preform/types"`] = struct{}{}
		col.IsScanner = true
	case "REAL":
		col.GoType += "float64"
	default:
		col.GoType = "any"
	}
}

func (d sqliteDialect) ParseCustomTypeScan(src any) (dst []string, err error) {
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
