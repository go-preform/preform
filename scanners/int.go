package scanners

import (
	"strconv"
)

type Uint struct {
	Value    *uint
	ValueAny *any
}

func (s *Uint) Ptr(p any) Interface {
	s.Value = p.(*uint)
	return s
}

func (s *Uint) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(uint)
	return s
}

func (s *Uint) Scan(src any) (err error) {
	switch src.(type) {
	case uint64:
		*s.Value = uint(src.(uint64))

	case uint32:
		*s.Value = uint(src.(uint32))

	case uint16:
		*s.Value = uint(src.(uint16))

	case uint8:
		*s.Value = uint(src.(uint8))

	case uint:
		*s.Value = src.(uint)

	case string:
		var (
			u64 uint64
		)
		u64, err = strconv.ParseUint(src.(string), 10, 64)
		*s.Value = uint(u64)

	case []byte:
		var (
			u64 uint64
		)
		u64, err = strconv.ParseUint(string(src.([]byte)), 10, 64)
		*s.Value = uint(u64)

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err

}

type Uint16 struct {
	Value    *uint16
	ValueAny *any
}

func (s *Uint16) Ptr(p any) Interface {
	s.Value = p.(*uint16)
	return s
}

func (s *Uint16) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(uint16)
	return s
}

func (s *Uint16) Scan(src any) (err error) {
	switch src.(type) {
	case uint64:
		*s.Value = uint16(src.(uint64))

	case uint32:
		*s.Value = uint16(src.(uint32))

	case uint:
		*s.Value = uint16(src.(uint))

	case uint8:
		*s.Value = uint16(src.(uint8))

	case uint16:
		*s.Value = src.(uint16)

	case string:
		var (
			u64 uint64
		)
		u64, err = strconv.ParseUint(src.(string), 10, 64)
		*s.Value = uint16(u64)

	case []byte:
		var (
			u64 uint64
		)
		u64, err = strconv.ParseUint(string(src.([]byte)), 10, 64)
		*s.Value = uint16(u64)

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err

}

type Uint32 struct {
	Value    *uint32
	ValueAny *any
}

func (s *Uint32) Ptr(p any) Interface {
	s.Value = p.(*uint32)
	return s
}

func (s *Uint32) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(uint32)
	return s
}

func (s *Uint32) Scan(src any) (err error) {
	switch src.(type) {
	case uint64:
		*s.Value = uint32(src.(uint64))

	case uint:
		*s.Value = uint32(src.(uint))

	case uint16:
		*s.Value = uint32(src.(uint16))

	case uint8:
		*s.Value = uint32(src.(uint8))

	case uint32:
		*s.Value = src.(uint32)

	case string:
		var (
			u64 uint64
		)
		u64, err = strconv.ParseUint(src.(string), 10, 64)
		*s.Value = uint32(u64)

	case []byte:
		var (
			u64 uint64
		)
		u64, err = strconv.ParseUint(string(src.([]byte)), 10, 64)
		*s.Value = uint32(u64)

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err

}

type Uint64 struct {
	Value    *uint64
	ValueAny *any
}

func (s *Uint64) Ptr(p any) Interface {
	s.Value = p.(*uint64)
	return s
}

func (s *Uint64) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(uint64)
	return s
}

func (s *Uint64) Scan(src any) (err error) {

	switch src.(type) {
	case uint:
		*s.Value = uint64(src.(uint))

	case uint32:
		*s.Value = uint64(src.(uint32))

	case uint16:
		*s.Value = uint64(src.(uint16))

	case uint8:
		*s.Value = uint64(src.(uint8))

	case uint64:
		*s.Value = src.(uint64)

	case string:

		*s.Value, err = strconv.ParseUint(src.(string), 10, 64)

	case []byte:

		*s.Value, err = strconv.ParseUint(string(src.([]byte)), 10, 64)

	}

	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}

type Int struct {
	Value    *int
	ValueAny *any
}

func (s *Int) Ptr(p any) Interface {
	s.Value = p.(*int)
	return s
}

func (s *Int) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(int)
	return s
}

func (s *Int) Scan(src any) (err error) {

	switch src.(type) {
	case int64:
		*s.Value = int(src.(int64))

	case int32:
		*s.Value = int(src.(int32))

	case int16:
		*s.Value = int(src.(int16))

	case int8:
		*s.Value = int(src.(int8))

	case int:
		*s.Value = src.(int)

	case string:
		var (
			i64 int64
		)
		i64, err = strconv.ParseInt(src.(string), 10, 64)
		*s.Value = int(i64)

	case []byte:
		var (
			i64 int64
		)
		i64, err = strconv.ParseInt(string(src.([]byte)), 10, 64)
		*s.Value = int(i64)

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err

}

type Int16 struct {
	Value    *int16
	ValueAny *any
}

func (s *Int16) Ptr(p any) Interface {
	s.Value = p.(*int16)
	return s
}

func (s *Int16) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(int16)
	return s
}

func (s *Int16) Scan(src any) (err error) {

	switch src.(type) {
	case int64:
		*s.Value = int16(src.(int64))

	case int32:
		*s.Value = int16(src.(int32))

	case int:
		*s.Value = int16(src.(int))

	case int8:
		*s.Value = int16(src.(int8))

	case int16:
		*s.Value = src.(int16)

	case string:
		var (
			i64 int64
		)
		i64, err = strconv.ParseInt(src.(string), 10, 64)
		*s.Value = int16(i64)

	case []byte:
		var (
			i64 int64
		)
		i64, err = strconv.ParseInt(string(src.([]byte)), 10, 64)
		*s.Value = int16(i64)

	}

	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}

type Int32 struct {
	Value    *int32
	ValueAny *any
}

func (s *Int32) Ptr(p any) Interface {
	s.Value = p.(*int32)
	return s
}

func (s *Int32) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(int32)
	return s
}

func (s *Int32) Scan(src any) (err error) {

	switch src.(type) {
	case int64:
		*s.Value = int32(src.(int64))

	case int:
		*s.Value = int32(src.(int))

	case int16:
		*s.Value = int32(src.(int16))

	case int8:
		*s.Value = int32(src.(int8))

	case int32:
		*s.Value = src.(int32)

	case string:
		var (
			i64 int64
		)
		i64, err = strconv.ParseInt(src.(string), 10, 64)
		*s.Value = int32(i64)

	case []byte:
		var (
			i64 int64
		)
		i64, err = strconv.ParseInt(string(src.([]byte)), 10, 64)
		*s.Value = int32(i64)

	}

	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}

type Int64 struct {
	Value    *int64
	ValueAny *any
}

func (s *Int64) Ptr(p any) Interface {
	s.Value = p.(*int64)
	return s
}

func (s *Int64) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(int64)
	return s
}

func (s *Int64) Scan(src any) (err error) {

	switch src.(type) {
	case int:
		*s.Value = int64(src.(int))

	case int32:
		*s.Value = int64(src.(int32))

	case int16:
		*s.Value = int64(src.(int16))

	case int8:
		*s.Value = int64(src.(int8))

	case int64:
		*s.Value = src.(int64)

	case string:
		*s.Value, err = strconv.ParseInt(src.(string), 10, 64)

	case []byte:
		*s.Value, err = strconv.ParseInt(string(src.([]byte)), 10, 64)

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err

}
