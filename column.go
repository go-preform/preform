package preform

import (
	"database/sql"
	"fmt"
	"github.com/go-preform/preform/scanners"
	preformShare "github.com/go-preform/preform/share"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type iPtrUnwrapper interface {
	UnwrapPtr() any
}
type Column[T any] struct {
	*column[T]
	alias         string
	factory       IFactory
	code          string
	codeWithAlias string
	colConditioner
}

type column[T any] struct {
	isPk              bool
	isAuto            bool
	isPtr             bool
	isArray           bool
	isScanner         bool
	pos               int
	NewPtr            func() *T
	dbName            string
	name              string
	dbType            string
	sqlScanner        func() scanners.Interface
	insertValueParser func(*T) any
	valueParser       func(*T) any
	defaultVal        string
	isPtrUnwrapper    bool
	valueCompareLess  func(any, any) bool
	stringfier        func(any) string
}

type ColumnSetter[T any] struct {
	*Column[T]
}
type ICol interface {
	preformShare.ICol
	iTypedCol
	IColConditioner
	initCol(ref reflect.StructField, dbName string, factory IFactory, pos int)
	setFactory(factory IFactory)
	clone(f IFactory) ICol
	properties() (isArray bool, isPtr bool, isPk bool, isAuto bool)
	SetAlias(alias string) ICol
	Alias() string
	QueryFactory() IQuery
	unwrapPtr(any) any
	unwrapPtrForInsert(any) any //care default value
	unwrapPtrForUpdate(any) any
	getValueFromBodyFlatten(body iModelBody) []any
	getValueFromBody(body iModelBody) any
	getValueFromBodiesAndLess(iModelBody, iModelBody) bool
	setValueToBody(body iModelBody, value any)
	wrapScanner(any) any
	valueToString(any) string
	flatten(any) []any

	Asc() string
	Desc() string

	Aggregate(fn preformShare.Aggregator, params ...any) iAggregateCol
	Sum() iAggregateCol
	Avg() iAggregateCol
	Max() iAggregateCol
	Min() iAggregateCol
	Count() iAggregateCol
	CountDistinct() iAggregateCol
	Mean() iAggregateCol
	Median() iAggregateCol
	Mode() iAggregateCol
	StdDev() iAggregateCol
	JsonAgg() iAggregateCol
	ArrayAgg() iAggregateCol
	GroupConcat(splitter string) iAggregateCol
	IsSame(ICol) bool
}

type IColFromFactory interface {
	ICol
	Factory() IFactory
}

// todo refactor init timing
func (c *Column[T]) initCol(ref reflect.StructField, dbName string, factory IFactory, pos int) {
	c.column = &column[T]{pos: pos, dbName: dbName, name: ref.Name}
	c.alias = ref.Name
	c.defaultVal = ref.Tag.Get("defaultValue")
	c.dbType = ref.Tag.Get("dataType")
	c.factory = factory
	c.colConditioner = colConditioner{col: c}
	if factory != nil && factory.Db() != nil {
		c.colConditioner.dialect = factory.Db().dialect
	}
	var (
		v     T
		vType = reflect.TypeOf(v)
	)
	if (vType.Kind() == reflect.Slice || vType.Kind() == reflect.Array) && vType.Elem().Kind() != reflect.Uint8 {
		c.isArray = true
	}
	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
		c.column.NewPtr = func() *T {
			return reflect.New(vType).Interface().(*T)
		}
		c.column.isPtr = true
	} else {
		c.column.NewPtr = c.column.newPtr
		prepareColumnTypeFunc[T](c.column, &v)
	}
}

func (c *Column[T]) setFactory(factory IFactory) {
	c.factory = factory
	c.colConditioner.dialect = factory.Db().dialect
	c.setValueParser()
}

func prepareColumnTypeFunc[T any](col *column[T], v any, arrEl ...any) {
	col.stringfier = func(v any) string {
		return fmt.Sprintf("%v", v)
	}
	switch v.(type) {
	case *int:
		col.stringfier = func(v any) string {
			return strconv.FormatInt(int64(v.(int)), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Int{} }, false, func(a, b any) bool {
			return a.(int) < b.(int)
		}
	case *int16:
		col.stringfier = func(v any) string {
			return strconv.FormatInt(int64(v.(int16)), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Int16{} }, false, func(a, b any) bool {
			return a.(int16) < b.(int16)
		}
	case *int32:
		col.stringfier = func(v any) string {
			return strconv.FormatInt(int64(v.(int32)), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Int32{} }, false, func(a, b any) bool {
			return a.(int32) < b.(int32)
		}
	case *int64:
		col.stringfier = func(v any) string {
			return strconv.FormatInt(v.(int64), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Int64{} }, false, func(a, b any) bool {
			return a.(int64) < b.(int64)
		}
	case *uint:
		col.stringfier = func(v any) string {
			return strconv.FormatUint(uint64(v.(uint)), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Uint{} }, false, func(a, b any) bool {
			return a.(uint) < b.(uint)
		}
	case *uint16:
		col.stringfier = func(v any) string {
			return strconv.FormatUint(uint64(v.(uint16)), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Uint16{} }, false, func(a, b any) bool {
			return a.(uint16) < b.(uint16)
		}
	case *uint32:
		col.stringfier = func(v any) string {
			return strconv.FormatUint(uint64(v.(uint32)), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Uint32{} }, false, func(a, b any) bool {
			return a.(uint32) < b.(uint32)
		}
	case *uint64:
		col.stringfier = func(v any) string {
			return strconv.FormatUint(v.(uint64), 10)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Uint64{} }, false, func(a, b any) bool {
			return a.(uint64) < b.(uint64)
		}
	case *float32:
		col.stringfier = func(v any) string {
			return strconv.FormatFloat(float64(v.(float32)), 'f', -1, 64)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Float32{} }, false, func(a, b any) bool {
			return a.(float32) < b.(float32)
		}
	case *float64:
		col.stringfier = func(v any) string {
			return strconv.FormatFloat(v.(float64), 'f', -1, 64)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Float64{} }, false, func(a, b any) bool {
			return a.(float64) < b.(float64)
		}
	case *string:
		col.stringfier = func(v any) string {
			return v.(string)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = nil, false, func(a, b any) bool {
			return a.(string) < b.(string)
		}
	case *bool:
		col.stringfier = func(v any) string {
			if v.(bool) {
				return "true"
			} else {
				return "false"
			}
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = nil, false, nil
	case *[]byte:
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Bytes{} }, false, func(a, b any) bool {
			return string(a.([]byte)) < string(b.([]byte))
		}
	case *time.Time:
		col.stringfier = func(v any) string {
			return v.(time.Time).Format(time.RFC3339Nano)
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = nil, false, func(a, b any) bool {
			return a.(time.Time).Before(b.(time.Time))
		}
	case *[]int32:
		col.stringfier = func(v any) string {
			var (
				vv = v.([]int32)
				s  = make([]string, len(vv))
			)
			for i := range vv {
				s[i] = strconv.FormatInt(int64(vv[i]), 10)
			}
			return "[" + strings.Join(s, ",") + "]"
		}
		col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[int32]{} }, false, nil
	default:
		if len(arrEl) != 0 {
			switch arrEl[0].(type) {
			case int:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]int)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatInt(int64(vv[i]), 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[int]{} }, false, nil
			case int16:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]int16)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatInt(int64(vv[i]), 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[int16]{} }, false, nil
			case int32:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]int32)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatInt(int64(vv[i]), 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[int32]{} }, false, nil
			case int64:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]int64)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatInt(vv[i], 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[int64]{} }, false, nil
			case uint:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]uint)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatUint(uint64(vv[i]), 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[uint]{} }, false, nil
			case uint16:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]uint16)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatUint(uint64(vv[i]), 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[uint16]{} }, false, nil
			case uint32:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]uint32)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatUint(uint64(vv[i]), 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[uint32]{} }, false, nil
			case uint64:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]uint64)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatUint(vv[i], 10)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[uint64]{} }, false, nil
			case float32:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]float32)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatFloat(float64(vv[i]), 'f', -1, 64)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[float32]{} }, false, nil
			case float64:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]float64)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = strconv.FormatFloat(vv[i], 'f', -1, 64)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[float64]{} }, false, nil
			case string:
				col.stringfier = func(v any) string {
					return "[" + strings.Join(v.([]string), ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[string]{} }, false, nil
			case bool:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]bool)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						if vv[i] {
							s[i] = "true"
						} else {
							s[i] = "false"
						}
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[bool]{} }, false, nil
			case []byte:
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[[]byte]{} }, false, nil
			case time.Time:
				col.stringfier = func(v any) string {
					var (
						vv = v.([]time.Time)
						s  = make([]string, len(vv))
					)
					for i := range vv {
						s[i] = vv[i].Format(time.RFC3339Nano)
					}
					return "[" + strings.Join(s, ",") + "]"
				}
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[time.Time]{} }, false, nil
			default:
				col.sqlScanner, col.isScanner, col.valueCompareLess = func() scanners.Interface { return &scanners.Array[any]{} }, false, nil

			}
		}
		if _, ok := v.(sql.Scanner); ok {
			col.sqlScanner, col.isScanner, col.valueCompareLess = nil, true, nil
		}
	}
}

func (c *Column[T]) flatten(arr any) []any {
	var (
		src = arr.([]T)
		dst = make([]any, len(src))
	)
	for i := range src {
		dst[i] = src[i]
	}
	return dst
}
func (c Column[T]) clone(f IFactory) ICol {
	var (
		cc   = c
		cPtr = &cc
	)
	cc.factory = f
	cc.code = ""
	cc.codeWithAlias = ""
	cc.colConditioner = colConditioner{col: cPtr, dialect: f.Db().dialect}
	c.setValueParser()
	return cPtr
}
func (c *Column[T]) SetAlias(alias string) ICol {
	return newColWrap(c, c.factory, alias)
}

func (c *Column[T]) GetCodeWithAlias() string {
	if c.codeWithAlias == "" {
		if c.dbName != c.alias {
			c.codeWithAlias = c.
				factory.
				Db().
				dialect.QuoteIdentifier(c.factory.Alias()) + "." + c.factory.Db().dialect.QuoteIdentifier(c.dbName) + " AS " + c.factory.Db().dialect.QuoteIdentifier(c.alias)
		} else {
			c.codeWithAlias = c.factory.Db().dialect.QuoteIdentifier(c.factory.Alias()) + "." + c.factory.Db().dialect.QuoteIdentifier(c.dbName)
		}
	}
	return c.codeWithAlias
}

func (c column[T]) getValueFromBodiesAndLess(bodyA, bodyB iModelBody) bool {
	if c.valueCompareLess == nil {
		return false
	}
	v1, v2 := bodyA.FieldValuePtr(c.pos), bodyB.FieldValuePtr(c.pos)
	return c.valueCompareLess(*(v1.(*T)), *(v2.(*T)))
}

func (c column[T]) getValueFromBody(body iModelBody) any {
	v := body.FieldValuePtr(c.pos)
	if c.isArray {
		if iter, ok := v.(preformShare.IArrayTypes); ok {
			return iter.IterAny()
		} else if c.isPtr {
			return iterAny(**(v.(**[]T)))
		} else {
			return iterAny(*(v.(*[]T)))
		}
	}
	return *(v.(*T))
}

func (c column[T]) getValueFromBodyFlatten(body iModelBody) []any {
	v := body.FieldValuePtr(c.pos)
	if c.isArray {
		if iter, ok := v.(preformShare.IArrayTypes); ok {
			return iter.IterAny()
		} else if c.isPtr {
			return iterAny(**(v.(**[]T)))
		} else {
			return iterAny(*(v.(*[]T)))
		}
	}
	return []any{*(v.(*T))}
}

func (c column[T]) setValueToBody(body iModelBody, value any) {
	*(body.FieldValuePtr(c.pos).(*T)) = value.(T)
}

func (c column[T]) valueToString(value any) string {
	return c.stringfier(value)
}

func (c column[T]) properties() (isArray bool, isPtr bool, isPk bool, isAuto bool) {
	isPtr = c.isPtr
	isArray = c.isArray
	isPk = c.isPk
	isAuto = c.isAuto
	return
}
func (c column[T]) GetRawPtrScanner() (vv any, toScanner func(*any) any) {
	var (
		v T
	)
	if c.sqlScanner == nil {
		if c.isScanner {
			// deprecated
			return v, func(a *any) any {
				return (&scanners.ScannerAny[T]{}).PtrAny(a)
			}
		}
		return v, func(a *any) any {
			return a
		}
	}
	return v, func(a *any) any {
		return c.sqlScanner().PtrAny(a)
	}
}

func (c column[T]) wrapScanner(ptr any) any {
	if c.sqlScanner != nil {
		s := c.sqlScanner()
		s.Ptr(ptr)
		return s
	}
	return ptr
}

func (c column[T]) unwrapPtr(ptr any) any {
	if c.isPtrUnwrapper {
		return ptr.(iPtrUnwrapper).UnwrapPtr()
	}
	return *(ptr.(*T))
}

func (c column[T]) unwrapPtrForUpdate(ptr any) any {
	return c.valueParser(ptr.(*T))
}

func (c column[T]) unwrapPtrForInsert(ptr any) any {
	return c.insertValueParser(ptr.(*T))
}

func (c *Column[T]) GetCode() string {
	if c.code == "" {
		c.code = c.factory.
			Db().
			dialect.
			QuoteIdentifier(c.factory.Alias()) + "." + c.factory.
			Db().
			dialect.
			QuoteIdentifier(c.dbName)
	}
	return c.code
}

func (c *column[T]) SetValue(ptr, value any) {
	*(ptr.(iModelBody).FieldValuePtr(c.pos).(*T)) =
		value.(T)
}

func (c *column[T]) NewValue() any {
	var (
		v T
	)
	return v
}

func (c *Column[T]) Asc() string {
	return c.GetCode() + " ASC"
}

func (c *Column[T]) Desc() string {
	return c.GetCode() + " DESC"
}

func (c Column[T]) ToSql() (string, []interface{}, error) {
	return c.GetCodeWithAlias(), nil, nil
}

func (c Column[T]) Alias() string {
	return c.alias
}

func SetColumn[T any](c *Column[T]) *ColumnSetter[T] {
	return &ColumnSetter[T]{c}
}

func (c *ColumnSetter[T]) GetCol() ICol {
	return c.col
}

func (c *ColumnSetter[T]) SetName(name string) *ColumnSetter[T] {
	c.dbName = name
	return c
}

func (c *ColumnSetter[T]) AutoIncrement() *ColumnSetter[T] {
	c.isAuto = true
	return c
}

func (c *ColumnSetter[T]) PK() *ColumnSetter[T] {
	c.isPk = true
	return c
}

func (c Column[T]) QueryFactory() IQuery {
	return c.factory
}

func (c Column[T]) ParentModel() any {
	return c.factory.NewBody()
}

func (c Column[T]) Factory() IFactory {
	return c.factory
}

type sliceUnsafe struct {
	Array unsafe.Pointer
	cap   int
	len   int
}

func (c *Column[T]) setValueParser() {
	if db := c.factory.Db(); db == nil {
		return
	} else {
		var (
			zeroVal T
			zeroPtr any = &zeroVal
			vt          = reflect.TypeOf(zeroVal)
			parsers     = db.dialect.ValueParsers()
		)
		_, c.isPtrUnwrapper = zeroPtr.(iPtrUnwrapper)
		if c.isPtrUnwrapper {
			c.valueParser = func(v *T) any {
				return any(v).(iPtrUnwrapper).UnwrapPtr()
			}
		} else {
			if parser, ok := parsers[vt]; ok {
				c.valueParser = parser.(func(string, bool) func(*T) any)(c.dbType, false)
			} else if vt.Kind() == reflect.Slice {
				var zeroArr = reflect.New(reflect.ArrayOf(0, vt.Elem())).Elem().Interface()
				c.valueParser = func(v *T) any {
					if (*sliceUnsafe)(unsafe.Pointer(v)).Array == nil {
						return zeroArr
					}
					return *v
				}
			} else {
				c.valueParser = func(v *T) any {
					return *v
				}
			}
		}
		if c.defaultVal != "" || c.isAuto {
			if parser, ok := parsers[vt]; ok {
				c.insertValueParser = parser.(func(string, bool) func(*T) any)(c.dbType, true)
			} else {
				c.insertValueParser = makeDefaultParser(zeroVal)
			}
		} else {
			c.insertValueParser = c.valueParser
		}
	}
}

func (c column[T]) Name() string {
	return c.name
}

func (c column[T]) DbName() string {
	return c.dbName
}

func (c column[T]) GetPos() int {
	return c.pos
}

func (c *column[T]) NewZero() T {
	var v T
	return v
}

func (c *column[T]) newPtr() *T {
	var v T
	return &v
}

func (c *column[T]) new() *T {
	var v T
	return any(v).(*T)
}

func (c Column[T]) IsSame(cc ICol) bool {
	return c.factory.CodeName() == cc.QueryFactory().CodeName() && c.Name() == cc.Name()
}

func makeDefaultParser[T any](zeroVal T) func(*T) any {

	switch any(zeroVal).(type) {
	case []byte:
		return func(v *T) any {
			if any(*v).([]byte) == nil {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case float32:
		return func(v *T) any {
			if any(*v).(float32) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case float64:
		return func(v *T) any {
			if any(*v).(float64) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case bool:
		return func(v *T) any {
			if any(*v).(bool) == false {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case string:
		return func(v *T) any {
			if any(*v).(string) == "" {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case uint:
		return func(v *T) any {
			if any(*v).(uint) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case uint16:
		return func(v *T) any {
			if any(*v).(uint16) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case uint32:
		return func(v *T) any {
			if any(*v).(uint32) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case uint64:
		return func(v *T) any {
			if any(*v).(uint64) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case int:
		return func(v *T) any {
			if any(*v).(int) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case int16:
		return func(v *T) any {
			if any(*v).(int16) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case int32:
		return func(v *T) any {
			if any(*v).(int32) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case int64:
		return func(v *T) any {
			if any(*v).(int64) == 0 {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	case time.Time:
		return func(v *T) any {
			if any(*v).(time.Time).IsZero() {
				return preformShare.DEFAULT_VALUE
			}
			return *v
		}
	}
	return func(v *T) any {
		if reflect.ValueOf(*v).IsZero() {
			return preformShare.DEFAULT_VALUE
		}
		return *v
	}
}
