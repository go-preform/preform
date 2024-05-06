package cachedQueryRunner

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/go-preform/preform/dialect"
	preformShare "github.com/go-preform/preform/share"
	"io"
	"sync"
	"time"
	"unsafe"
)

type iCacher interface {
	Load(key string) (value [][]driver.Value, saver func(value [][]driver.Value, relatedFactories []preformShare.IQueryFactory), ok bool)
	ClearByFactories(factories []preformShare.IQueryFactory)
}

type cachedQueryRunner struct {
	preformShare.DbQueryRunner
	cacher           iCacher
	relatedFactories []preformShare.IQueryFactory
}

func NewUnsafeCachedQueryRunner(DbQueryRunner preformShare.DbQueryRunner, cacher ...iCacher) *cachedQueryRunner {
	if len(cacher) == 0 {
		cacher = append(cacher, NewDefaultCacher(30*time.Minute))
	}
	return &cachedQueryRunner{DbQueryRunner: DbQueryRunner, cacher: cacher[0]}
}

func (r *cachedQueryRunner) RelatedFactory(factories []preformShare.IQueryFactory) preformShare.QueryRunner {
	rr := *r
	rr.relatedFactories = factories
	return &rr
}

func (r *cachedQueryRunner) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	value, saver, ok := r.cacher.Load(r.hashKey(bytes.NewBuffer([]byte(query)), args))
	if !ok {
		rows, err := r.DbQueryRunner.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		value, err = r.rowsToCache(rows)
		if err != nil {
			return nil, err
		}
		saver(value, r.relatedFactories)
	} else {
		if logger := ctx.Value(preformShare.CTX_LOGGER); logger != nil {
			logger.(func(string))(fmt.Sprintf("cached result: %d", len(value)-1))
		}
	}
	return cachedRows(value), nil
}

func (r *cachedQueryRunner) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	value, saver, ok := r.cacher.Load(r.hashKey(bytes.NewBuffer([]byte(query)), args))
	if !ok {
		rows, err := r.DbQueryRunner.QueryContext(ctx, query, args...)
		if err != nil {
			return cachedRow(nil, err)
		}
		value, err = r.rowsToCache(rows)
		if err != nil {
			return cachedRow(nil, err)
		}
		saver(value, r.relatedFactories)
	}
	return cachedRow(value, nil)
}

func (r *cachedQueryRunner) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	res, err := r.DbQueryRunner.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	r.cacher.ClearByFactories(r.relatedFactories)
	return res, nil
}

func (r *cachedQueryRunner) InsertAndReturnAutoId(ctx context.Context, lastIdMethod preformShare.SqlDialectLastInsertIdMethod, query string, args ...interface{}) (lastId int64, err error) {
	if lastIdMethod == dialect.LastInsertIdMethodByRes {
		res, err := r.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, err
		}
		r.cacher.ClearByFactories(r.relatedFactories)
		return res.LastInsertId()
	} else {
		err = r.QueryRowContext(ctx, query, args...).Scan(&lastId)
		return
	}
}

func (r cachedQueryRunner) BaseRunner() preformShare.DbQueryRunner {
	return r.DbQueryRunner
}

func (r cachedQueryRunner) hashKey(buff *bytes.Buffer, args []any) string {
	for _, arg := range args {
		switch arg.(type) {
		case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
			buff.WriteString(fmt.Sprintf("+%v", arg))
		case string:
			buff.WriteString(fmt.Sprintf(":%s", arg.(string)))
		case []byte:
			buff.WriteString(fmt.Sprintf("#%s", arg.(string)))
		default:
			buff.WriteString(fmt.Sprintf("-%v", arg))
		}
	}
	return buff.String()
}

func (r *cachedQueryRunner) rowsToCache(rows *sql.Rows) ([][]driver.Value, error) {
	var (
		err     error
		values  = make([][]driver.Value, 1, 50)
		value   []any
		valPtrs []any
		cols    []string
	)
	cols, err = rows.Columns()
	if err != nil {
		return nil, err
	}
	value = make([]any, len(cols))
	values[0] = []driver.Value{cols}
	for i := range value {
		valPtrs = append(valPtrs, &value[i])
	}
	for rows.Next() {
		if err = rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		values = append(values, *(*[]driver.Value)(unsafe.Pointer(&value)))
	}
	return values, nil
}

type driverConn struct {
	db        *sql.DB
	createdAt time.Time

	sync.Mutex  // guards following
	ci          driver.Conn
	needReset   bool // The connection session should be reset before use if true.
	closed      bool
	finalClosed bool // ci.Close has been called
	openStmt    map[uintptr]bool

	// guarded by db.mu
	inUse      bool
	returnedAt time.Time // Time the connection was created or returned.
	onPut      []func()  // code (with db.mu held) run when conn is next returned
	dbmuClosed bool      // same as closed, but guarded by db.mu, for removeClosedStmtLocked
}

type Rows struct {
	dc          *driverConn // owned; must call releaseConn when closed to release
	releaseConn func(error)
	rowsi       driver.Rows
	cancel      func()  // called when Rows is closed, may be nil.
	closeStmt   uintptr // if non-nil, statement to Close on close

	closemu sync.RWMutex
	closed  bool
	lasterr error // non-nil only if closed is true

	// lastcols is only used in Scan, Next, and NextResultSet which are expected
	// not to be called concurrently.
	lastcols []any
}

func cachedRows(value [][]driver.Value) *sql.Rows {
	rows := &Rows{dc: &driverConn{}, releaseConn: func(error) {}, rowsi: &driverRows{value: value, size: len(value[0][0].([]string)), pos: 1, len: len(value)}}
	return (*sql.Rows)(unsafe.Pointer(rows))
}

type driverRows struct {
	value [][]driver.Value
	size  int
	pos   int
	len   int
}

func (d driverRows) Columns() []string {
	return d.value[0][0].([]string)
}

func (d driverRows) Close() error {
	return nil
}

func (d *driverRows) Next(dest []driver.Value) error {
	if d.pos == d.len {
		return io.EOF
	}
	var (
		l = len(dest)
	)
	if l != d.size {
		return fmt.Errorf("dest size %d not match with row size %d", l, d.size)
	}
	copy(dest, d.value[d.pos])
	d.pos++
	return nil
}

type Row struct {
	err  error
	rows *sql.Rows
}

func cachedRow(value [][]driver.Value, err error) *sql.Row {
	return (*sql.Row)(unsafe.Pointer(&Row{err: err, rows: cachedRows(value)}))
}
