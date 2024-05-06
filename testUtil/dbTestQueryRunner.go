package preformTestUtil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	preformShare "github.com/go-preform/preform/share"
	"github.com/jmoiron/sqlx"
)

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

	contextDone atomic.Pointer[error] // error that awaitDone saw; set before close attempt

	// closemu prevents Rows from closing while there
	// is an active streaming result. It is held for read during non-close operations
	// and exclusively during close.
	//
	// closemu guards lasterr and closed.
	closemu sync.RWMutex
	closed  bool
	lasterr error // non-nil only if closed is true

	// lastcols is only used in Scan, Next, and NextResultSet which are expected
	// not to be called concurrently.
	lastcols []driver.Value

	// closemuScanHold is whether the previous call to Scan kept closemu RLock'ed
	// without unlocking it. It does that when the user passes a *RawBytes scan
	// target. In that case, we need to prevent awaitDone from closing the Rows
	// while the user's still using the memory. See go.dev/issue/60304.
	//
	// It is only used by Scan, Next, and NextResultSet which are expected
	// not to be called concurrently.
	closemuScanHold bool

	// hitEOF is whether Next hit the end of the rows without
	// encountering an error. It's set in Next before
	// returning. It's only used by Next and Err which are
	// expected not to be called concurrently.
	hitEOF bool
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

type TestQueryRunner struct {
	ErrorQueue    []error
	LastIdQueue   []int64
	AffectedQueue []int64
	StmtQueue     []*sql.Stmt
	QueryRows     []*sql.Rows
	QueryxRows    []*sqlx.Rows
}

func NewTestQueryRunner() *TestQueryRunner {
	return &TestQueryRunner{}
}

func (t *TestQueryRunner) popError() error {
	if len(t.ErrorQueue) == 0 {
		return nil
	}
	err := t.ErrorQueue[0]
	t.ErrorQueue = t.ErrorQueue[1:]
	return err
}

func (t *TestQueryRunner) popLastId() int64 {
	if len(t.LastIdQueue) == 0 {
		return 0
	}
	id := t.LastIdQueue[0]
	t.LastIdQueue = t.LastIdQueue[1:]
	return id
}

func (t *TestQueryRunner) popAffected() int64 {
	if len(t.AffectedQueue) == 0 {
		return 0
	}
	id := t.AffectedQueue[0]
	t.AffectedQueue = t.AffectedQueue[1:]
	return id
}

func (t *TestQueryRunner) AddToQueryRows(value [][]driver.Value, columns ...[]string) {
	if len(value) == 0 || len(value[0]) == 0 {
		return
	}
	if len(columns) > 0 {
		value = append([][]driver.Value{{columns[0]}}, value...)
	}
	rows := &Rows{dc: &driverConn{}, releaseConn: func(error) {}, rowsi: &driverRows{value: value, size: len(value[0][0].([]string)), pos: 1, len: len(value)}, closemu: sync.RWMutex{}}
	t.QueryRows = append(t.QueryRows, (*sql.Rows)(unsafe.Pointer(rows)))
}

func (t *TestQueryRunner) AddToQueryxRows(value [][]driver.Value, columns ...[]string) {
	if len(value) == 0 || len(value[0]) == 0 {
		return
	}
	if len(columns) > 0 {
		value = append([][]driver.Value{{columns[0]}}, value...)
	}
	rows := &Rows{dc: &driverConn{}, releaseConn: func(error) {}, rowsi: &driverRows{value: value, size: len(value[0][0].([]string)), pos: 1, len: len(value)}, closemu: sync.RWMutex{}}
	t.QueryxRows = append(t.QueryxRows, (*sqlx.Rows)(unsafe.Pointer(rows)))
}

func (t *TestQueryRunner) Prepare(query string) (*sql.Stmt, error) {
	return t.PrepareContext(context.Background(), query)
}

func (t *TestQueryRunner) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//preform will skip nil prepared statements
	return nil, t.popError()
}

func (t *TestQueryRunner) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.ExecContext(context.Background(), query, args...)
}

type testQueryRunnerResult struct {
	parent *TestQueryRunner
}

func (t *testQueryRunnerResult) LastInsertId() (int64, error) {
	return t.parent.popLastId(), t.parent.popError()
}

func (t *testQueryRunnerResult) RowsAffected() (int64, error) {
	return t.parent.popAffected(), t.parent.popError()
}

func (t *TestQueryRunner) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return &testQueryRunnerResult{parent: t}, t.popError()
}

func (t *TestQueryRunner) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.QueryContext(context.Background(), query, args...)
}

func (t *TestQueryRunner) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// handle fake rows in test IModelScanner
	if len(t.QueryRows) == 0 {
		return &sql.Rows{}, t.popError()
	}
	rows := t.QueryRows[0]
	t.QueryRows = t.QueryRows[1:]
	return rows, t.popError()
}

func (t *TestQueryRunner) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.QueryRowContext(context.Background(), query, args...)
}

type Row struct {
	err  error
	rows *sql.Rows
}

func (t *TestQueryRunner) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// handle fake rows in test IModelScanner
	rows := &sql.Rows{}
	if len(t.QueryRows) > 0 {
		rows = t.QueryRows[0]
		t.QueryRows = t.QueryRows[1:]
	}
	return (*sql.Row)(unsafe.Pointer(&Row{err: t.popError(), rows: rows}))
}

func (t *TestQueryRunner) Preparex(query string) (*sqlx.Stmt, error) {
	return t.PreparexContext(context.Background(), query)
}

func (t *TestQueryRunner) PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error) {
	return nil, t.popError()
}

func (t *TestQueryRunner) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return t.QueryxContext(context.Background(), query, args...)
}

func (t *TestQueryRunner) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	// handle fake rows in test IModelScanner
	if len(t.QueryxRows) == 0 {
		return &sqlx.Rows{}, t.popError()
	}
	rows := t.QueryxRows[0]
	t.QueryxRows = t.QueryxRows[1:]
	return rows, t.popError()
}

func (t *TestQueryRunner) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return t.QueryRowxContext(context.Background(), query, args...)
}

func (t *TestQueryRunner) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return &sqlx.Row{}
}

func (t *TestQueryRunner) InsertAndReturnAutoId(ctx context.Context, lastIdMethod preformShare.SqlDialectLastInsertIdMethod, query string, args ...any) (int64, error) {
	return t.popLastId(), t.popError()
}

func (t *TestQueryRunner) BaseRunner() preformShare.DbQueryRunner {
	return t
}

func (t *TestQueryRunner) RelatedFactory([]preformShare.IQueryFactory) preformShare.QueryRunner {
	return t
}

func (t *TestQueryRunner) IsTester() {}
