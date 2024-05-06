package scanners

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Array[T any] struct {
	Value    *[]T
	ValueAny *any
}

func (s *Array[T]) Ptr(p any) Interface {
	s.Value = p.(*[]T)
	return s
}

func (s *Array[T]) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new([]T)
	return s
}

func (s *Array[T]) Scan(src any) (err error) {
	if src == nil {
		s.Value = nil
		return nil
	}
	var (
		v       T
		vPtr    any = &v
		scanner Interface
	)
	switch src.(type) {
	case []T:
		*s.Value = src.([]T)
		return nil
	case string:
		src = []byte(src.(string))
	}
	if _, ok := vPtr.(sql.Scanner); ok {
		scanner = &Scanner{}
	} else {
		switch vPtr.(type) {
		case *string:
			scanner = &String{}
		case *int:
			scanner = &Int{}
		case *int16:
			scanner = &Int16{}
		case *int32:
			scanner = &Int32{}
		case *int64:
			scanner = &Int64{}
		case *float32:
			scanner = &Float32{}
		case *float64:
			scanner = &Float64{}
		case *bool:
			scanner = &Bool{}
		case *[]byte:
			scanner = &Bytes{}
		case *time.Time:
			scanner = &Time{}
		default:
			scanner = &Json{}
		}
	}
	dims, elems, err := parseArray(src.([]byte), []byte(","))
	if err != nil {
		return err
	}
	if len(dims) > 1 {
		return fmt.Errorf("pq: scanning from multidimensional ARRAY%s is not implemented",
			strings.Replace(fmt.Sprint(dims), " ", "][", -1))
	}
	if len(dims) == 0 {
		dims = append(dims, 0)
	}
	tmp := make([]T, len(elems))
	for i, e := range elems {
		scanner.Ptr(&tmp[i])
		err = scanner.Scan(e)
		if err != nil {
			return err
		}
	}
	*s.Value = tmp
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}

// parseArray extracts the dimensions and elements of an array represented in
// text format. Only representations emitted by the backend are supported.
// Notably, whitespace around brackets and delimiters is significant, and NULL
// is case-sensitive.
//
// See http://www.postgresql.org/docs/current/static/arrays.html#ARRAYS-IO
func parseArray(src, del []byte) (dims []int, elems [][]byte, err error) {
	var depth, i int

	if len(src) < 1 || src[0] != '{' {
		return nil, nil, fmt.Errorf("pq: unable to parse array; expected %q at offset %d", '{', 0)
	}

Open:
	for i < len(src) {
		switch src[i] {
		case '{':
			depth++
			i++
		case '}':
			elems = make([][]byte, 0)
			goto Close
		default:
			break Open
		}
	}
	dims = make([]int, i)

Element:
	for i < len(src) {
		switch src[i] {
		case '{':
			if depth == len(dims) {
				break Element
			}
			depth++
			dims[depth-1] = 0
			i++
		case '"':
			var elem = []byte{}
			var escape bool
			for i++; i < len(src); i++ {
				if escape {
					elem = append(elem, src[i])
					escape = false
				} else {
					switch src[i] {
					default:
						elem = append(elem, src[i])
					case '\\':
						escape = true
					case '"':
						elems = append(elems, elem)
						i++
						break Element
					}
				}
			}
		default:
			for start := i; i < len(src); i++ {
				if bytes.HasPrefix(src[i:], del) || src[i] == '}' {
					elem := src[start:i]
					if len(elem) == 0 {
						return nil, nil, fmt.Errorf("pq: unable to parse array; unexpected %q at offset %d", src[i], i)
					}
					if bytes.Equal(elem, []byte("NULL")) {
						elem = nil
					}
					elems = append(elems, elem)
					break Element
				}
			}
		}
	}

	for i < len(src) {
		if bytes.HasPrefix(src[i:], del) && depth > 0 {
			dims[depth-1]++
			i += len(del)
			goto Element
		} else if src[i] == '}' && depth > 0 {
			dims[depth-1]++
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("pq: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}

Close:
	for i < len(src) {
		if src[i] == '}' && depth > 0 {
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("pq: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}
	if depth > 0 {
		err = fmt.Errorf("pq: unable to parse array; expected %q at offset %d", '}', i)
	}
	if err == nil {
		for _, d := range dims {
			if (len(elems) % d) != 0 {
				err = fmt.Errorf("pq: multidimensional arrays must have elements with matching dimensions")
			}
		}
	}
	return
}
