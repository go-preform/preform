package preformTypes

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-preform/preform/scanners"
	"strconv"
	"strings"
)

type Array[T any] []T

func (a Array[T]) TypeForExport() any {
	return []T{}
}

func (a Array[T]) IterAny() []any {
	var (
		dst = make([]any, len(a))
	)
	for i, v := range a {
		dst[i] = v
	}
	return dst
}

func (a *Array[T]) UnwrapPtr() any {
	return a.IterAny()
}

func (a Array[T]) ValueParsers(v Array[T]) any {
	if v == nil {
		return [0]T{}
	}
	return a
}

func (a *Array[T]) Scan(src any) error {
	var (
		arr []T
	)
	scanner := &scanners.Array[T]{}
	scanner.Ptr(&arr)
	err := scanner.Scan(src)
	if err != nil {
		return err
	}
	*a = arr
	return nil
}

func (a Array[T]) Value() (driver.Value, error) {
	if n := len(a); n > 0 {
		var (
			val          driver.Value
			err          error
			valueStrings = make([]string, n)
			a0Any        = any(a[0])
			str          string
		)
		if _, ok := a0Any.(driver.Valuer); ok {
			for i, v := range a {
				val, err = any(v).(driver.Valuer).Value()
				if err != nil {
					return nil, err
				}
				if str, ok = val.(string); ok {
					if strings.Contains(str, "::") {
						parts := strings.Split(str, "::")
						valueStrings[i] = fmt.Sprintf(`"%s"`, strings.Replace(parts[0][1:len(parts[0])-1], `"`, `\"`, -1))
						continue
					}
				}
				valueStrings[i] = quoteValueToString(val)
			}
		} else {
			switch a0Any.(type) {
			case int:
				for i, v := range a {
					valueStrings[i] = strconv.FormatInt(int64(any(v).(int)), 10)
				}
			case int32:
				for i, v := range a {
					valueStrings[i] = strconv.FormatInt(int64(any(v).(int32)), 10)
				}
			case int64:
				for i, v := range a {
					valueStrings[i] = strconv.FormatInt(any(v).(int64), 10)
				}
			case float32:
				for i, v := range a {
					valueStrings[i] = strconv.FormatFloat(float64(any(v).(float32)), 'f', -1, 32)
				}
			case float64:
				for i, v := range a {
					valueStrings[i] = strconv.FormatFloat(any(v).(float64), 'f', -1, 64)
				}
			case bool:
				for i, v := range a {
					valueStrings[i] = strconv.FormatBool(any(v).(bool))
				}
			case string:
				for i, v := range a {
					valueStrings[i] = quoteValueToString(v)
				}
			case []byte:
				size := 1 + 6*n
				for _, x := range a {
					size += hex.EncodedLen(len(any(x).([]byte)))
				}

				b := make([]byte, size)

				for i, s := 0, b; i < n; i++ {
					o := copy(s, `,"\\x`)
					o += hex.Encode(s[o:], any(a[i]).([]byte))
					s[o] = '"'
					s = s[o+1:]
				}

				b = []byte(fmt.Sprintf(`ARRAY[%s]`, string(b[1:size-1])))
				return string(b), nil
			default:
				data, err := json.Marshal(a)
				if err != nil {
					return nil, err
				}
				if data[0] == '[' {
					data[0], data[len(data)-1] = '{', '}'
					return string(data), nil
				} else {
					return nil, fmt.Errorf("pq: unable to parse array; unexpected %q at offset %d", data[0], 0)
				}

			}
		}
		return "{" + strings.Join(valueStrings, ",") + "}", nil
	}

	return "NULL", nil
}

func quoteValueToString(v driver.Value) string {
	switch v.(type) {
	case string:
		return "\"" + strings.Replace(v.(string), "\"", "\\\"", -1) + "\""
	case []byte:
		return "\"" + strings.Replace(string(v.([]byte)), "\"", "\\\"", -1) + "\""
	default:
		return "\"" + strings.Replace(fmt.Sprintf("%v", v), "\"", "\\\"", -1) + "\""
	}
}
