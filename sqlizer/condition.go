package preformSqlizer

import (
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	"strings"
)

type Between map[string][2]any

func (b Between) ToSql() (query string, args []any, err error) {
	var (
		l       = len(b)
		i       = 0
		queries = make([]string, l)
	)
	args = make([]any, l*2)
	for k, v := range b {
		queries[i] = fmt.Sprintf("%s BETWEEN ? AND ?", k)
		args[i*2], args[i*2+1] = v[0], v[1]
		i++
	}
	query = "(" + strings.Join(queries, " AND ") + ")"
	return
}

type If struct {
	Cond any
	Then any
	Else any
}

func (i If) ToSql() (query string, args []any, err error) {
	var (
		condQ, thenQ, elseQ string
		tempArgs            []any
	)
	switch i.Cond.(type) {
	case string:
		condQ = i.Cond.(string)
	case squirrel.Sqlizer:
		condQ, tempArgs, err = i.Cond.(squirrel.Sqlizer).ToSql()
		if err != nil {
			return
		}
		args = append(args, tempArgs...)
	}
	switch i.Then.(type) {
	case string:
		thenQ = i.Then.(string)
	case squirrel.Sqlizer:
		thenQ, tempArgs, err = i.Then.(squirrel.Sqlizer).ToSql()
		if err != nil {
			return
		}
		args = append(args, tempArgs...)
	}
	switch i.Else.(type) {
	case string:
		elseQ = i.Else.(string)
	case squirrel.Sqlizer:
		elseQ, tempArgs, err = i.Else.(squirrel.Sqlizer).ToSql()
		if err != nil {
			return
		}
		args = append(args, tempArgs...)
	}
	return preformShare.NestCondSql(fmt.Sprintf("IF(%s, %s, %s)", condQ, thenQ, elseQ), args)
}

type IfNull struct {
	Cond any
	Then any
}

func (i IfNull) ToSql() (query string, args []any, err error) {
	var (
		condQ, thenQ string
		tempArgs     []any
	)
	switch i.Cond.(type) {
	case string:
		condQ = i.Cond.(string)
	case squirrel.Sqlizer:
		condQ, tempArgs, err = i.Cond.(squirrel.Sqlizer).ToSql()
		if err != nil {
			return
		}
		args = append(args, tempArgs...)
	}
	switch i.Then.(type) {
	case string:
		thenQ = i.Then.(string)
	case squirrel.Sqlizer:
		thenQ, tempArgs, err = i.Then.(squirrel.Sqlizer).ToSql()
		if err != nil {
			return
		}
		args = append(args, tempArgs...)
	}
	return preformShare.NestCondSql(fmt.Sprintf("IFNULL(%s, %s)", condQ, thenQ), args)
}
