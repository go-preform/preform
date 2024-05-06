package scanners

import "strconv"

type Float32 struct {
	Value    *float32
	ValueAny *any
}

func (s *Float32) Ptr(p any) Interface {
	s.Value = p.(*float32)
	return s
}

func (s *Float32) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(float32)
	return s
}

func (s *Float32) Scan(src any) (err error) {
	switch src.(type) {
	case float64:
		*s.Value = float32(src.(float64))
	case float32:
		*s.Value = src.(float32)
	case []byte:
		var (
			f64 float64
		)
		f64, err = strconv.ParseFloat(string(src.([]byte)), 64)
		*s.Value = float32(f64)
	case string:
		var (
			f64 float64
		)
		f64, err = strconv.ParseFloat(src.(string), 64)
		*s.Value = float32(f64)
	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}

type Float64 struct {
	Value    *float64
	ValueAny *any
}

func (s *Float64) Ptr(p any) Interface {
	s.Value = p.(*float64)
	return s
}

func (s *Float64) PtrAny(p *any) Interface {
	s.ValueAny = p
	s.Value = new(float64)
	return s
}

func (s *Float64) Scan(src any) (err error) {
	switch src.(type) {
	case float64:
		*s.Value = src.(float64)
	case float32:
		*s.Value = float64(src.(float32))
	case []byte:
		*s.Value, err = strconv.ParseFloat(string(src.([]byte)), 64)
		return err
	case string:
		*s.Value, err = strconv.ParseFloat(src.(string), 64)
		return err
	}
	if s.ValueAny != nil {
		*s.ValueAny = *s.Value
	}
	return err
}
