package preformSqlizer

import (
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
)

type ArrayEq struct {
	ArrColA any
	ArrColB any
}

func (h ArrayEq) WithDialect(d preformShare.IDialect) squirrel.Sqlizer {
	return preformShare.SqlizerWithDialectWrapper{Sqlizer: func() (string, []any, error) {
		return d.ArrayEq(h.ArrColA, h.ArrColB)
	}}
}

type ArrayHasAny struct {
	ArrColA any
	ArrColB any
}

func (h ArrayHasAny) WithDialect(d preformShare.IDialect) squirrel.Sqlizer {
	return preformShare.SqlizerWithDialectWrapper{Sqlizer: func() (string, []any, error) {
		return d.ArrayHasAny(h.ArrColA, h.ArrColB)
	}}
}

func (h ArrayHasAny) ToSql(dialect preformShare.IDialect) (query string, args []any, err error) {
	return dialect.ArrayHasAny(h.ArrColA, h.ArrColB)
}

type ArrayAny struct {
	ArrCol any
	Value  any
}

func (h ArrayAny) WithDialect(d preformShare.IDialect) squirrel.Sqlizer {
	return preformShare.SqlizerWithDialectWrapper{Sqlizer: func() (string, []any, error) {
		return d.ArrayAny(h.ArrCol, h.Value)
	}}
}

func (h ArrayAny) ToSql(dialect preformShare.IDialect) (query string, args []any, err error) {
	return dialect.ArrayAny(h.ArrCol, h.Value)
}

type ArrayContainsBy struct {
	ArrColA any
	ArrColB any
}

func (h ArrayContainsBy) WithDialect(d preformShare.IDialect) squirrel.Sqlizer {
	return preformShare.SqlizerWithDialectWrapper{Sqlizer: func() (string, []any, error) {
		return d.ArrayContainsBy(h.ArrColA, h.ArrColB)
	}}
}

func (h ArrayContainsBy) ToSql(dialect preformShare.IDialect) (query string, args []any, err error) {
	return dialect.ArrayContainsBy(h.ArrColA, h.ArrColB)
}

type ArrayContains struct {
	ArrColA any
	ArrColB any
}

func (h ArrayContains) WithDialect(d preformShare.IDialect) squirrel.Sqlizer {
	return preformShare.SqlizerWithDialectWrapper{Sqlizer: func() (string, []any, error) {
		return d.ArrayContains(h.ArrColA, h.ArrColB)
	}}
}

func (h ArrayContains) ToSql(dialect preformShare.IDialect) (query string, args []any, err error) {
	return dialect.ArrayContains(h.ArrColA, h.ArrColB)
}

type ArrayConcat struct {
	ArrColA any
	ArrColB any
}

func (h ArrayConcat) WithDialect(d preformShare.IDialect) squirrel.Sqlizer {
	return preformShare.SqlizerWithDialectWrapper{Sqlizer: func() (string, []any, error) {
		return d.ArrayConcat(h.ArrColA, h.ArrColB)
	}}
}

func (h ArrayConcat) ToSql(dialect preformShare.IDialect) (query string, args []any, err error) {
	return dialect.ArrayConcat(h.ArrColA, h.ArrColB)
}
