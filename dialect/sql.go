package dialect

import (
	"database/sql"
	"errors"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	preformSqlizer "github.com/go-preform/preform/sqlizer"
	"github.com/go-preform/squirrel"
	"reflect"
	"strings"
)

var (
	ErrorNotSupport = errors.New("driver not support")
)

const (
	LastInsertIdMethodByRes preformShare.SqlDialectLastInsertIdMethod = iota
	LastInsertIdMethodBySuffix
	LastInsertIdMethodNone
)

type basicSqlDialect struct {
	quoteTpl           string
	lastInsertIdMethod preformShare.SqlDialectLastInsertIdMethod
	lastInsertIdSuffix func(col string) squirrel.Sqlizer
}

func (d basicSqlDialect) QuoteIdentifier(s string) string {
	return fmt.Sprintf(d.quoteTpl, s)
}

func (d basicSqlDialect) LastInsertIdMethod() (method preformShare.SqlDialectLastInsertIdMethod, suffix func(col string) squirrel.Sqlizer) {
	return d.lastInsertIdMethod, d.lastInsertIdSuffix
}

func NewBasicSqlDialect(quoteTpl string) *basicSqlDialect {
	return &basicSqlDialect{quoteTpl: quoteTpl}
}

func (d basicSqlDialect) GetStructure(db *sql.DB, schemasEmptyIsAll ...string) []*preformShare.Scheme {
	return nil
}

func (d basicSqlDialect) Aggregate(fn preformShare.Aggregator, body any, params ...any) squirrel.Sqlizer {
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
		return squirrel.Expr(fmt.Sprintf("GROUP_CONCAT(%s, ?)", bodyStr), append(args, params[0])...)
	case AggCountDistinct:
		return squirrel.Expr(fmt.Sprintf("COUNT(DISTINCT %s)", bodyStr), args...)
	default:
		if l := len(params); l != 0 {
			return squirrel.Expr(fmt.Sprintf("%s(%s%s)", fn, bodyStr, strings.Repeat(",?", l)), append(args, params...)...)
		}
		return squirrel.Expr(fmt.Sprintf("%s(%s)", fn, bodyStr), args...)
	}
}

func (d basicSqlDialect) ArrayAny(arrCol any, value any) (query string, args []any, err error) {
	return "", nil, ErrorNotSupport
}

func (d basicSqlDialect) ArrayHasAny(arrColA any, arrColB any) (query string, args []any, err error) {
	return "", nil, ErrorNotSupport
}

func (d basicSqlDialect) ArrayConcat(arrColA any, arrColB any) (query string, args []any, err error) {
	return "", nil, ErrorNotSupport
}
func (d basicSqlDialect) ArrayContains(arrColA any, arrColB any) (query string, args []any, err error) {
	return "", nil, ErrorNotSupport
}
func (d basicSqlDialect) ArrayContainsBy(arrColA any, arrColB any) (query string, args []any, err error) {
	return "", nil, ErrorNotSupport
}
func (d basicSqlDialect) ArrayEq(arrColA any, arrColB any) (query string, args []any, err error) {
	return "", nil, ErrorNotSupport
}

func (d basicSqlDialect) UpdateSqlizer(q preformShare.UpdateBuilder) (string, []any, error) {
	return q.ToSql()
}
func (d basicSqlDialect) DeleteSqlizer(q preformShare.DeleteBuilder) (string, []any, error) {
	return q.ToSql()
}

func (d basicSqlDialect) ValueParsers() map[reflect.Type]any {
	return map[reflect.Type]any{}
}

func (d basicSqlDialect) DefaultValueExpr() squirrel.Sqlizer {
	return squirrel.Expr("DEFAULT")
}

func (d basicSqlDialect) ParseCustomTypeScan(src any) (dst []string, err error) {
	return nil, ErrorNotSupport
}

func (d basicSqlDialect) ParseCustomTypeValue(typeName string, values ...any) (dst string, err error) {
	return "", ErrorNotSupport
}

func (d basicSqlDialect) Eq(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.Eq{col.GetCode(): v}
}

func (d basicSqlDialect) NotEq(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.NotEq{col.GetCode(): v}
}

func (d basicSqlDialect) Like(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.Like{col.GetCode(): v}
}

func (d basicSqlDialect) Gt(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.Gt{col.GetCode(): v}
}

func (d basicSqlDialect) GtOrEq(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.GtOrEq{col.GetCode(): v}
}

func (d basicSqlDialect) Lt(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.Lt{col.GetCode(): v}
}

func (d basicSqlDialect) LtOrEq(col preformShare.ICol, v any) squirrel.Sqlizer {
	return squirrel.LtOrEq{col.GetCode(): v}
}

func (d basicSqlDialect) Between(col preformShare.ICol, v1, v2 any) squirrel.Sqlizer {
	return preformSqlizer.Between{col.GetCode(): [2]any{v1, v2}}
}
