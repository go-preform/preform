package preform

import (
	"context"
	"database/sql"
	"github.com/go-preform/preform/dialect"
	preformShare "github.com/go-preform/preform/share"
	"github.com/jmoiron/sqlx"
)

// implements QueryRunner
type dbWithTracer struct {
	preformShare.QueryRunner
	tracer     ITracer
	driverName string
	txId       string
}

func (d *dbWithTracer) QueryRowContext(ctx context.Context, query string, args ...interface{}) (row *sql.Row) {
	ctx, end := d.tracer.Trace(ctx, d.driverName, query, d.txId, args...)
	row = d.QueryRunner.QueryRowContext(ctx, query, args...)
	end(row.Err())(false, nil)
	return
}

func (d *dbWithTracer) QueryRow(query string, args ...interface{}) (row *sql.Row) {
	return d.QueryRowContext(context.Background(), query, args...)
}

func (d *dbWithTracer) QueryRowxContext(ctx context.Context, query string, args ...interface{}) (row *sqlx.Row) {
	ctx, end := d.tracer.Trace(ctx, d.driverName, query, d.txId, args...)
	row = d.QueryRunner.QueryRowxContext(ctx, query, args...)
	end(row.Err())(false, nil)
	return
}

func (d *dbWithTracer) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return d.QueryRowxContext(context.Background(), query, args...)
}

func (d *dbWithTracer) QueryxContext(ctx context.Context, query string, args ...interface{}) (rows *sqlx.Rows, err error) {
	ctx, end := d.tracer.Trace(ctx, d.driverName, query, d.txId, args...)
	rows, err = d.QueryRunner.QueryxContext(ctx, query, args...)
	end(err)(false, nil)
	return
}

func (d *dbWithTracer) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return d.QueryxContext(context.Background(), query, args...)
}

func (d *dbWithTracer) ExecContext(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	ctx, end := d.tracer.TraceExec(ctx, d.driverName, query, d.txId, args...)
	res, err = d.QueryRunner.ExecContext(ctx, query, args...)
	end(res, err)
	return
}

func (d *dbWithTracer) InsertAndReturnAutoId(ctx context.Context, lastIdMethod preformShare.SqlDialectLastInsertIdMethod, query string, args ...interface{}) (lastId int64, err error) {
	if lastIdMethod == dialect.LastInsertIdMethodByRes {
		res, err := d.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	} else {
		ctx, end := d.tracer.TraceExec(ctx, d.driverName, query, d.txId, args...)
		err = d.QueryRunner.QueryRowContext(ctx, query, args...).Scan(&lastId)
		end(dummySqlResult{err: err}, err)
		return
	}
}

func (d *dbWithTracer) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.ExecContext(context.Background(), query, args...)
}

func (d *dbWithTracer) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	ctx, end := d.tracer.Trace(ctx, d.driverName, query, d.txId, args...)
	rows, err = d.QueryRunner.QueryContext(ctx, query, args...)
	end(err)(false, nil)
	return
}

func (d *dbWithTracer) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	return d.QueryContext(context.Background(), query, args...)
}

type RowsWithTrace struct {
	*sql.Rows
	finisher func(bool, error)
}

func (r RowsWithTrace) Next() bool {
	n := r.Rows.Next()
	if !n {
		r.finisher(true, nil)
		r.finisher = nil
	}
	return n
}

func (r *RowsWithTrace) Close() error {
	if r.finisher != nil {
		r.finisher(true, nil)
		r.finisher = nil
	}
	return r.Rows.Close()
}

func (d *dbWithTracer) QueryTraceScan(ctx context.Context, query string, args ...interface{}) (rows IRows, err error) {
	ctx, end := d.tracer.Trace(ctx, d.driverName, query, d.txId, args...)
	var (
		r *sql.Rows
	)
	r, err = d.QueryRunner.QueryContext(ctx, query, args...)
	rows = &RowsWithTrace{Rows: r, finisher: end(err)}
	return
}

func (d *dbWithTracer) RelatedFactory(fs []preformShare.IQueryFactory) preformShare.QueryRunner {
	d.QueryRunner = d.QueryRunner.RelatedFactory(fs)
	return d
}

type StmtWithTrace struct {
	*sql.Stmt
	tracer              ITracer
	driver, query, txId string
}

func (s *StmtWithTrace) Query(args ...interface{}) (rows IRows, err error) {
	return s.QueryContext(context.Background(), args...)
}

func (s *StmtWithTrace) QueryRow(args ...interface{}) (row *sql.Row) {
	return s.QueryRowContext(context.Background(), args...)
}

func (s *StmtWithTrace) Exec(args ...interface{}) (res sql.Result, err error) {
	return s.ExecContext(context.Background(), args...)
}

func (s *StmtWithTrace) QueryContext(ctx context.Context, args ...interface{}) (rows IRows, err error) {
	ctx, end := s.tracer.Trace(ctx, s.driver, s.query, s.txId, args...)
	var (
		r *sql.Rows
	)
	r, err = s.Stmt.QueryContext(ctx, args...)
	rows = &RowsWithTrace{Rows: r, finisher: end(err)}
	return
}

func (s *StmtWithTrace) QueryRowContext(ctx context.Context, args ...interface{}) (row *sql.Row) {
	ctx, end := s.tracer.Trace(ctx, s.driver, s.query, s.txId, args...)
	row = s.Stmt.QueryRowContext(ctx, args...)
	end(row.Err())(false, nil)
	return
}

func (s *StmtWithTrace) ExecContext(ctx context.Context, args ...interface{}) (res sql.Result, err error) {
	ctx, end := s.tracer.TraceExec(ctx, s.driver, s.query, s.txId, args...)
	res, err = s.Stmt.ExecContext(ctx, args...)
	end(res, err)
	return
}

func (d *dbWithTracer) PrepareTrace(ctx context.Context, query string) (stmt IStmt, err error) {
	var (
		s *sql.Stmt
	)
	s, err = d.QueryRunner.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	return &StmtWithTrace{Stmt: s, tracer: d.tracer, driver: d.driverName, query: query, txId: d.txId}, nil
}

type dummySqlResult struct {
	err error
}

func (d dummySqlResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (d dummySqlResult) RowsAffected() (int64, error) {
	if d.err == sql.ErrNoRows {
		return 0, nil
	}
	return 1, nil
}
