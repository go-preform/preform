package preformTypes

import (
	"database/sql/driver"
	"errors"
	"math/big"
	"reflect"
	"strings"
)

type Rat struct {
	big.Rat
}

var (
	ZeroRat = NewRat(0, 1)
)

func NewRat(a, b int64) *Rat {
	return &Rat{*big.NewRat(a, b)}
}

func NewRatFromFloat64(f float64) *Rat {
	return NewRat(0, 1).
		SetFloat64(f)
}

func NewRatFromString(s string) *Rat {
	r, _ := NewRat(0, 1).SetString(s)
	return r
}

func (z *Rat) Clone() *Rat {
	return &Rat{Rat: z.Rat}
}

// SetFloat64 sets z to exactly f and returns z.
// If f is not finite, SetFloat returns nil.
func (z *Rat) SetFloat64(f float64) *Rat {
	z.Rat.SetFloat64(f)
	return z
}

// SetFrac sets z to a/b and returns z.
// If b == 0, SetFrac panics.
func (z *Rat) SetFrac(a, b *big.Int) *Rat {
	z.Rat.SetFrac(a, b)
	return z
}

// SetFrac64 sets z to a/b and returns z.
// If b == 0, SetFrac64 panics.
func (z *Rat) SetFrac64(a, b int64) *Rat {
	z.Rat.SetFrac64(a, b)
	return z
}

// SetInt sets z to x (by making a copy of x) and returns z.
func (z *Rat) SetInt(x *big.Int) *Rat {
	z.Rat.SetInt(x)
	return z
}

// SetInt64 sets z to x and returns z.
func (z *Rat) SetInt64(x int64) *Rat {
	z.Rat.SetInt64(x)
	return z
}

// SetUint64 sets z to x and returns z.
func (z *Rat) SetUint64(x uint64) *Rat {
	z.Rat.SetUint64(x)
	return z
}

// SetString sets z to the value of s and returns z and a boolean indicating
// success.
func (z *Rat) SetString(s string) (*Rat, bool) {
	_, ok := z.Rat.SetString(s)
	return z, ok
}

// Set sets z to x (by making a copy of x) and returns z.
func (z *Rat) Set(x *Rat) *Rat {
	z.Rat.Set(&x.Rat)
	return z
}

// Abs sets z to |x| (the absolute value of x) and returns z.
func (z *Rat) Abs(x *Rat) *Rat {
	z.Rat.Abs(&x.Rat)
	return z
}

// Neg sets z to -z and returns z.
func (z *Rat) Neg() *Rat {
	z.Rat.Neg(&z.Rat)
	return z
}

// Inv sets z to 1/x and returns z.
// If x == 0, Inv panics.
func (z *Rat) Inv(x *Rat) *Rat {
	z.Rat.Inv(&x.Rat)
	return z
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (x *Rat) Cmp(y *Rat) int {
	return x.Rat.Cmp(&y.Rat)
}

// Add sets z to the sum x+y and returns z.
func (z *Rat) Add(x, y *Rat) *Rat {
	z.Rat.Add(&x.Rat, &y.Rat)
	return z
}

func (z *Rat) AddX(x *Rat) *Rat {
	z.Rat.Add(&z.Rat, &x.Rat)
	return z
}

// Sub sets z to the difference x-y and returns z.
func (z *Rat) Sub(x, y *Rat) *Rat {
	z.Rat.Sub(&x.Rat, &y.Rat)
	return z
}
func (z *Rat) SubX(x *Rat) *Rat {
	z.Rat.Sub(&z.Rat, &x.Rat)
	return z
}

// Mul sets z to the product x*y and returns z.
func (z *Rat) Mul(x, y *Rat) *Rat {
	z.Rat.Mul(&x.Rat, &y.Rat)
	return z
}
func (z *Rat) MulX(x *Rat) *Rat {
	z.Rat.Mul(&x.Rat, &z.Rat)
	return z
}

// Quo sets z to the quotient x/y and returns z.
// If y == 0, Quo panics.
func (z *Rat) Quo(x, y *Rat) *Rat {
	z.Rat.Quo(&x.Rat, &y.Rat)
	return z
}
func (z *Rat) QuoX(x *Rat) *Rat {
	z.Rat.Quo(&z.Rat, &x.Rat)
	return z
}

func (z *Rat) FloatString(prec int) string {
	if prec <= -1 {
		return strings.TrimRight(z.Rat.FloatString(15), ".0")
	}
	return z.Rat.FloatString(prec)
}

func (z *Rat) String() string {
	return z.FloatString(-1)
}

func (z *Rat) Scan(value interface{}) error {
	if value == nil {
		return errors.New("Cannot scan null value into Rat")
	}
	// first try to see if the data is stored in database as a Numeric datatype
	switch v := value.(type) {

	case float32:
		z.Rat = *(big.NewRat(1, 1).SetFloat64(float64(v)))
		return nil

	case float64:
		// numeric in sqlite3 sends us float64
		z.Rat = *(big.NewRat(1, 1).SetFloat64(v))
		return nil

	case int64:
		// at least in sqlite3 when the value is 0 in db, the data is sent
		// to us as an int64 instead of a float64 ...
		z.Rat = *(big.NewRat(1, 1).SetInt64(v))
		return nil

	case []byte:
		rat, ok := big.NewRat(1, 1).SetString(string(v))
		if !ok {
			return errors.New("Invalid Rat format")
		}
		z.Rat = *rat
		return nil
	case string:
		rat, ok := big.NewRat(1, 1).SetString(v)
		if !ok {
			return errors.New("Invalid Rat format")
		}
		z.Rat = *rat
		return nil
	default:

		return errors.New("Invalid Rat format:" + reflect.TypeOf(v).String())
	}

}

func (z Rat) Value() (driver.Value, error) {
	return z.FloatString(-1), nil
}

func (j *Rat) UnmarshalJSON(src []byte) error {
	if len(src) == 0 {
		return nil
	}
	if src[0] == '"' {
		src = src[1 : len(src)-1]
	}
	r, ok := j.Rat.SetString(string(src))
	if ok {
		j.Rat = *r
		return nil
	} else {
		return errors.New("Invalid Rat format")
	}
}

func (j Rat) MarshalJSON() ([]byte, error) {
	return []byte(j.FloatString(-1)), nil
}

// swagger as string
func (j Rat) TypeForExport() any {
	return ""
}
