package cachedQueryRunner

import (
	"database/sql/driver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnsafeScan(t *testing.T) {
	r := &cachedQueryRunner{}
	data := [][]driver.Value{
		{[]string{"a", "b", "c", "d"}},
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10, 11, 12},
	}
	rows := r.cachedRows(data)
	i := 1
	cols, err := rows.Columns()
	assert.Nil(t, err)
	assert.Equal(t, cols, data[0][0])
	_, err = rows.ColumnTypes()
	rows.NextResultSet()
	assert.Nil(t, err)
	for rows.Next() {
		var (
			a int
			b int
			c int
			d int
		)
		err = rows.Scan(&a, &b, &c, &d)
		assert.Nil(t, err)
		assert.Equal(t, a, data[i][0])
		assert.Equal(t, b, data[i][1])
		assert.Equal(t, c, data[i][2])
		assert.Equal(t, d, data[i][3])
		i++
	}
}
