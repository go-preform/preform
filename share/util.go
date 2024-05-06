package preformShare

import (
	"errors"
	"fmt"
	"github.com/go-preform/squirrel"
	"strings"
)

func NestCondSql(query string, args []any, dialect ...IDialect) (string, []any, error) {

	var (
		err       error
		finalArgs []any
		tmpArgs   []any
		tmpSql    string
		index     int
	)
	for _, arg := range args {
		if _, ok := arg.(ICol); ok {
			query = strings.Replace(query, "?", arg.(ICol).GetCode(), 1)

		} else if _, ok := arg.(ISqlizerWithDialect); ok {
			if len(dialect) == 0 {
				return "", nil, errors.New("no dialect for ISqlizerWithDialect")
			}
			tmpSql, tmpArgs, err = arg.(ISqlizerWithDialect).ToSql(dialect[0])
			if err != nil {
				return "", nil, err
			}
			tmpSql, tmpArgs, err = NestCondSql(tmpSql, tmpArgs)

			finalArgs = append(finalArgs, tmpArgs...)
			query = strings.Replace(query, "?", tmpSql, 1)

		} else if _, ok := arg.(squirrel.Sqlizer); ok {
			tmpSql, tmpArgs, err = arg.(squirrel.Sqlizer).ToSql()
			if err != nil {
				return "", nil, err
			}
			tmpSql, tmpArgs, err = NestCondSql(tmpSql, tmpArgs)

			finalArgs = append(finalArgs, tmpArgs...)

			if index = strings.Index(query, "?"); index != -1 {
				if index == 0 {
					if query[0:8] == "? = ANY(" {
						query = fmt.Sprintf("(%s)", tmpSql) + query[1:]
						continue
					}
				} else if query[index-1] == ' ' {
					if query[index-2] == '=' {
						if query[index-3] == '!' {
							query = query[0:index-3] + fmt.Sprintf("NOT IN (%s)", tmpSql) + query[index+1:]
						} else {
							query = query[0:index-2] + fmt.Sprintf("IN (%s)", tmpSql) + query[index+1:]
						}
						continue
					} else if query[index-2] == '>' {
						if query[index-3] == '<' {
							query = query[0:index-3] + fmt.Sprintf("NOT BETWEEN %s", tmpSql) + query[index+1:]
							continue
						}
					}
				}
				query = query[0:index] + fmt.Sprintf("(%s)", tmpSql) + query[index+1:]

			}

		} else {
			finalArgs = append(finalArgs, arg)
		}
	}
	return query, finalArgs, nil
}
func NestSql(query string, args []any, dialect ...IDialect) (string, []any, error) {

	var (
		err       error
		finalArgs []any
		tmpArgs   []any
		tmpSql    string
	)
	for _, arg := range args {
		if _, ok := arg.(ISqlizerWithDialect); ok {
			if len(dialect) == 0 {
				return "", nil, errors.New("no dialect for ISqlizerWithDialect")
			}
			tmpSql, tmpArgs, err = arg.(ISqlizerWithDialect).ToSql(dialect[0])
			if err != nil {
				return "", nil, err
			}
			tmpSql, tmpArgs, err = NestSql(tmpSql, tmpArgs)

			finalArgs = append(finalArgs, tmpArgs...)
			query = strings.Replace(query, "?", tmpSql, 1)

		} else if _, ok := arg.(squirrel.Sqlizer); ok {
			tmpSql, tmpArgs, err = arg.(squirrel.Sqlizer).ToSql()
			if err != nil {
				return "", nil, err
			}
			tmpSql, tmpArgs, err = NestSql(tmpSql, tmpArgs)

			finalArgs = append(finalArgs, tmpArgs...)

			query = strings.Replace(query, "?", tmpSql, 1)

		} else {
			finalArgs = append(finalArgs, arg)
		}
	}
	return query, finalArgs, nil
}
