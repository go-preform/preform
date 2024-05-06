package dialect

import (
	"database/sql/driver"
	"fmt"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
)

func (d clickhouseDialect) ArrayEq(arrColA any, arrColB any) (query string, args []any, err error) {
	return d.arrayFormatter(arrColA, arrColB, "%s = %s")
}

func (d clickhouseDialect) ArrayConcat(arrColA any, arrColB any) (query string, args []any, err error) {
	return d.arrayFormatter(arrColA, arrColB, "arrayConcat(%s, %s)")
}
func (d clickhouseDialect) ArrayContains(arrColA any, arrColB any) (query string, args []any, err error) {
	return d.arrayFormatter(arrColA, arrColB, "hasAll(%s, %s)")
}
func (d clickhouseDialect) ArrayContainsBy(arrColA any, arrColB any) (query string, args []any, err error) {
	return d.arrayFormatter(arrColB, arrColA, "hasAll(%s, %s)")
}

func (d clickhouseDialect) ArrayAny(arrCol any, value any) (query string, args []any, err error) {
	return d.arrayFormatter(arrCol, value, "has(%s, %s)")
}

func (d clickhouseDialect) ArrayHasAny(arrColA any, arrColB any) (query string, args []any, err error) {
	return d.arrayFormatter(arrColA, arrColB, "hasAny(%s, %s)")
}

func (d clickhouseDialect) arrayFormatter(v1 any, v2 any, format string) (query string, args []any, err error) {

	switch v1.(type) {
	case preformShare.ISqlizerWithDialect:
		return d.arrayFormatter(v1.(preformShare.ISqlizerWithDialect).WithDialect(d), v2, format)
	case preformShare.ICol:
		switch v2.(type) {
		case preformShare.ISqlizerWithDialect:
			return d.arrayFormatter(v1, v2.(preformShare.ISqlizerWithDialect).WithDialect(d), format)
		case preformShare.ICol:
			return fmt.Sprintf(format, v1.(preformShare.ICol).GetCode(), v2.(preformShare.ICol).GetCode()), args, nil
		case squirrel.Sqlizer:
			return preformShare.NestCondSql(fmt.Sprintf(format, v1.(preformShare.ICol).GetCode(), "?"), []any{v2}, d)
		default:
			return fmt.Sprintf(format, v1.(preformShare.ICol).GetCode(), "?"), []any{v2}, nil
		}
	case squirrel.Sqlizer:
		switch v2.(type) {
		case preformShare.ISqlizerWithDialect:
			return d.arrayFormatter(v1, v2.(preformShare.ISqlizerWithDialect).WithDialect(d), format)
		case preformShare.ICol:
			return preformShare.NestCondSql(fmt.Sprintf(format, "?", v2.(preformShare.ICol).GetCode()), []any{v1}, d)
		default:
			return preformShare.NestCondSql(fmt.Sprintf(format, "?", "?"), []any{v1, v2}, d)
		}
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []string, []int, []int8, []int16, []int32, []int64, []uint, []uint8, []uint16, []uint32, []uint64, []float32, []float64:
		goto caseAny
	default:
		if _, ok := v1.(driver.Valuer); ok {
			goto caseAny
		}
		return "", nil, fmt.Errorf("Any not support %T %v", v1, v1)
	}
caseAny:
	switch v2.(type) {
	case preformShare.ISqlizerWithDialect:
		return d.arrayFormatter(v1, v2.(preformShare.ISqlizerWithDialect).WithDialect(d), format)
	case preformShare.ICol:
		return fmt.Sprintf(format, "?", v2.(preformShare.ICol).GetCode()), []any{v1}, nil
	case squirrel.Sqlizer:
		return preformShare.NestCondSql(fmt.Sprintf(format, "?", "?"), []any{v1, v2}, d)
	default:
		return fmt.Sprintf(format, "?", "?"), []any{v1, v2}, nil
	}
}
