package preformTypes

import (
	"database/sql/driver"
	"encoding/json"
)

type JsonRaw[T any] struct {
	unmarshalled bool
	value        T
	src          []byte
}

func NewJsonRaw[T any](v T) JsonRaw[T] {
	var (
		j = JsonRaw[T]{}
	)
	j.Set(v)
	return j
}

func (j JsonRaw[T]) Src() []byte {
	return j.src
}

func (j *JsonRaw[T]) Set(v T) {
	j.unmarshalled = true
	j.value = v
	j.src, _ = json.Marshal(v)
}

func (j *JsonRaw[T]) Get() T {
	if !j.unmarshalled {
		if json.Unmarshal(j.src, &j.value) == nil {
			j.unmarshalled = true
		}
	}
	return j.value
}

func (j JsonRaw[T]) Value() (driver.Value, error) {
	if j.unmarshalled {
		j.src, _ = json.Marshal(j.value)
	}
	return string(j.src), nil
}

func (j *JsonRaw[T]) V() (driver.Value, error) {
	if j.unmarshalled {
		return j.value, nil
	}
	if j.src == nil || len(j.src) == 0 {
		return nil, nil
	}
	j.unmarshalled = true
	return j.value, json.Unmarshal(j.src, &j.value)
}

func (j *JsonRaw[T]) Scan(src interface{}) (err error) {
	if src == nil {
		return
	}
	switch src.(type) {
	case []byte:
		j.src = src.([]byte)
	case string:
		j.src = []byte(src.(string))
	default:
		j.value = src.(T)
		j.src, err = json.Marshal(src)
	}
	return
}

func (j JsonRaw[T]) T() any {
	var (
		t T
	)
	return t
}

func (j JsonRaw[T]) TypeForExport() any {
	var (
		t T
	)
	return t
}

func (j *JsonRaw[T]) SetSrc(src []byte) {
	j.src = src
}

func (j *JsonRaw[T]) UnmarshalJSON(src []byte) error {
	j.src = src
	return json.Unmarshal(src, &j.value)
}

func (j JsonRaw[T]) MarshalJSON() ([]byte, error) {
	if j.unmarshalled {
		j.src, _ = json.Marshal(j.value)
	}
	if len(j.src) != 0 && string(j.src) != "null" {
		return j.src, nil
	}
	return []byte("{}"), nil
}

func (j JsonRaw[T]) String() string {
	return string(j.src)
}
