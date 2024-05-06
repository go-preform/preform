package scanners

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"
)

type Interface interface {
	Scan(src any) error
	Ptr(p any) Interface
	PtrAny(p *any) Interface
}

type Scanner struct {
	Value sql.Scanner
}

func (s *Scanner) Ptr(p any) Interface {
	s.Value = p.(sql.Scanner)
	return s
}

func (s *Scanner) PtrAny(p *any) Interface {
	s.Value = (*p).(sql.Scanner)
	return s
}

func (s *Scanner) Scan(src any) (err error) {
	return s.Value.Scan(src)
}

type ScannerAny[T any] struct {
	Value    sql.Scanner
	ValueAny *any
}

func (s *ScannerAny[T]) Ptr(p any) Interface {
	return s
}

func (s *ScannerAny[T]) PtrAny(p *any) Interface {
	var (
		v T
	)
	s.Value = any(&v).(sql.Scanner)
	s.ValueAny = p
	return s
}

func (s *ScannerAny[T]) Scan(src any) (err error) {
	err = s.Value.Scan(src)
	if err == nil {
		*s.ValueAny = *any(s.Value).(*T)
	}
	return err

}

type String struct {
	Value    *string
	ValueAny *any
}

func (s *String) Ptr(p any) Interface {
	s.Value = p.(*string)
	return s
}

func (s *String) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(string)
	return s
}

func (s *String) Scan(src any) error {
	switch src.(type) {
	case time.Time:
		*s.Value = src.(time.Time).Format(time.RFC3339Nano)

	case string:
		*s.Value = src.(string)

	case []byte:
		*s.Value = string(src.([]byte))

	case int, int8, int16, int32, int64:
		*s.Value = strconv.FormatInt(reflect.ValueOf(src).Int(), 10)

	case uint, uint8, uint16, uint32, uint64:
		*s.Value = strconv.FormatUint(reflect.ValueOf(src).Uint(), 10)

	case float32, float64:
		*s.Value = strconv.FormatFloat(reflect.ValueOf(src).Float(), 'f', -1, 64)

	case bool:
		*s.Value = strconv.FormatBool(src.(bool))

	}

	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return nil
}

type Bytes struct {
	Value    *[]byte
	ValueAny *any
}

func (s *Bytes) Ptr(p any) Interface {
	s.Value = p.(*[]byte)
	return s
}

func (s *Bytes) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new([]byte)
	return s
}

func (s *Bytes) Scan(src any) error {
	switch src.(type) {
	case string:
		*s.Value = []byte(src.(string))

	case []byte:
		*s.Value = src.([]byte)

	case int, int8, int16, int32, int64:
		*s.Value = strconv.AppendInt(*s.Value, reflect.ValueOf(src).Int(), 10)

	case uint, uint8, uint16, uint32, uint64:
		*s.Value = strconv.AppendUint(*s.Value, reflect.ValueOf(src).Uint(), 10)

	case float32, float64:
		*s.Value = strconv.AppendFloat(*s.Value, reflect.ValueOf(src).Float(), 'f', -1, 64)

	case bool:
		*s.Value = strconv.AppendBool(*s.Value, reflect.ValueOf(src).Bool())

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return nil
}

type Bool struct {
	Value    *bool
	ValueAny *any
}

func (s *Bool) Ptr(p any) Interface {
	s.Value = p.(*bool)
	return s
}

func (s *Bool) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(bool)
	return s
}

func (s *Bool) Scan(src any) (err error) {
	switch src.(type) {
	case bool:
		*s.Value = src.(bool)

	case string:
		*s.Value, err = strconv.ParseBool(src.(string))

	case []byte:
		*s.Value, err = strconv.ParseBool(string(src.([]byte)))

	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}

type Time struct {
	Value    *time.Time
	ValueAny *any
}

func (s *Time) Ptr(p any) Interface {
	s.Value = p.(*time.Time)
	return s
}

func (s *Time) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(time.Time)
	return s
}

func (s *Time) Scan(src any) (err error) {
	switch src.(type) {
	case time.Time:
		*s.Value = src.(time.Time)
	case string:
		*s.Value, err = time.Parse(time.RFC3339Nano, src.(string))
	default:
		err = errors.New("invalid time format")
	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return nil
}

type Json struct {
	Value    any
	ValueAny *any
}

func (j *Json) Ptr(p any) Interface {
	j.Value = p
	return j
}

func (j *Json) PtrAny(p *any) Interface {
	j.ValueAny = p
	return j
}

func (j *Json) Scan(src any) error {
	err := json.Unmarshal(src.([]byte), &j.Value)
	if err != nil {
		return err
	}
	if j.ValueAny != nil {
		*j.ValueAny = j.Value
	}
	return nil
}
