package preformTypes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFloatString(t *testing.T) {
	r := NewRatFromFloat64(12.5)
	assert.Equal(t, "12.5", r.FloatString(-1))
	r = NewRatFromFloat64(1.251)
	assert.Equal(t, "1.251", r.FloatString(-1))
}

func BenchmarkNeg1(b *testing.B) {
	r := NewRatFromFloat64(1.251)
	for i := 0; i < b.N; i++ {
		_ = r.FloatString(-1)
	}
}

func Benchmark5(b *testing.B) {
	r := NewRatFromFloat64(1.251)
	for i := 0; i < b.N; i++ {
		_ = r.FloatString(5)
	}
}
