package preformTypes

/*
CREATE TABLE null_test (
	id serial4 NOT NULL,
	int32 int4 NULL,
	int64 int8 NULL,
	str varchar NULL,
	f32 float4 NULL,
	f64 float8 NULL,
	"bool" bool NULL,
	bytes bytea NULL,
	int_array _int4 NULL,
	f32_array _float4 NULL,
	"time" timestamptz NULL,
	CONSTRAINT null_test_pk PRIMARY KEY (id)
);
*/
import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func init() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "", 5432, "postgres", "", "oss")
	conn, err := sql.Open("pgx", psqlconn)
	if err != nil {
		panic(err)
	}
	Init(conn)
}

func TestRead(t *testing.T) {
	allNulls, err := Main.NullTest.Select().GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allNulls))
	assert.Equal(t, int32(1), allNulls[0].Id)
	assert.Equal(t, false, allNulls[0].Str.Valid)
	assert.Equal(t, false, allNulls[0].Int32.Valid)
	assert.Equal(t, false, allNulls[0].Int64.Valid)
	assert.Equal(t, false, allNulls[0].F32.Valid)
	assert.Equal(t, false, allNulls[0].F64.Valid)
	assert.Equal(t, false, allNulls[0].Bool.Valid)
	assert.Equal(t, false, allNulls[0].Bytes.Valid)
	assert.Equal(t, false, allNulls[0].IntArray.Valid)
	assert.Equal(t, false, allNulls[0].F32Array.Valid)
	assert.Equal(t, false, allNulls[0].Time.Valid)

	assert.Equal(t, int32(2), allNulls[1].Id)
	assert.Equal(t, true, allNulls[1].Str.Valid)
	assert.Equal(t, "the quick brown fox jumps over the lazy dog", allNulls[1].Str.V)
	assert.Equal(t, true, allNulls[1].Int32.Valid)
	assert.Equal(t, int32(2), allNulls[1].Int32.V)
	assert.Equal(t, true, allNulls[1].Int64.Valid)
	assert.Equal(t, int64(3), allNulls[1].Int64.V)
	assert.Equal(t, true, allNulls[1].F32.Valid)
	assert.Equal(t, float32(4.5), allNulls[1].F32.V)
	assert.Equal(t, true, allNulls[1].F64.Valid)
	assert.Equal(t, 6.7, allNulls[1].F64.V)
	assert.Equal(t, true, allNulls[1].Bool.Valid)
	assert.Equal(t, true, allNulls[1].Bool.V)
	assert.Equal(t, true, allNulls[1].Bytes.Valid)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 16}, allNulls[1].Bytes.V)
	assert.Equal(t, true, allNulls[1].IntArray.Valid)
	assert.Equal(t, preformTypes.Array[int32]{8}, allNulls[1].IntArray.V)
	assert.Equal(t, true, allNulls[1].F32Array.Valid)
	assert.Equal(t, preformTypes.Array[float32]{9.1}, allNulls[1].F32Array.V)
	assert.Equal(t, true, allNulls[1].Time.Valid)
	assert.Equal(t, time.Date(2024, 1, 4, 0, 0, 0, 0, time.Local), allNulls[1].Time.V)
}
