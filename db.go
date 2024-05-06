package preform

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-preform/preform/dialect"
	preformShare "github.com/go-preform/preform/share"
	"github.com/go-preform/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lann/builder"
	"math/rand"
	"reflect"
	"strconv"
)

type db struct {
	preformShare.QueryRunner
	DB                  *sqlx.DB
	dialect             preformShare.IDialect
	ctx                 context.Context
	sqPlaceholderFormat squirrel.PlaceholderFormat
	sqStmtBuilder       squirrel.StatementBuilderType
	driverName          string
	healthCtxCancel     context.CancelFunc
	errorLogger         func(driver, msg string, err error)
	queryTraceScan      func(ctx context.Context, query string, args ...interface{}) (rows IRows, err error)
	prepareTrace        func(ctx context.Context, query string) (IStmt, error)
	tracer              ITracer
}
type DB interface {
	preformShare.QueryRunner
	GetDialect() preformShare.IDialect
	Db() *db
	Error(msg string, err error)

	QueryTraceScan(ctx context.Context, query string, args ...interface{}) (rows IRows, err error)
	PrepareTrace(ctx context.Context, query string) (IStmt, error)
}

var (
	DefaultDB *db
)

func DbFromNative(d *sql.DB, queryRunner ...preformShare.QueryRunner) *db {
	dd := &db{
		dialect:             dialect.NewBasicSqlDialect("`%s`"),
		driverName:          "postgres",
		ctx:                 context.TODO(),
		sqPlaceholderFormat: squirrel.Dollar,
		sqStmtBuilder:       squirrel.StatementBuilderType(builder.EmptyBuilder).PlaceholderFormat(squirrel.Dollar),

		errorLogger: errorLog,
	}

	if d != nil {
		if testDriver, ok := d.Driver().(preformShare.ITestSqlConnectorDriver); ok {
			dd.driverName = testDriver.TestDriverName()
			switch dd.driverName {
			case "sqlite3", "sqlite":
				dd.dialect = dialect.NewSqliteDialect()
			case "postgres":
				dd.dialect = dialect.NewPostgresqlDialect()
			case "mysql":
				dd.sqPlaceholderFormat = squirrel.Question
				dd.sqStmtBuilder = squirrel.StatementBuilderType(builder.EmptyBuilder).PlaceholderFormat(squirrel.Question)
				dd.dialect = dialect.NewMysqlDialect()
			case "clickhouse":
				dd.dialect = dialect.NewClickhouseDialect()
			}
			dd.DB = sqlx.NewDb(d, dd.driverName)
		} else {
			//some drivers are not exported
			switch reflect.TypeOf(d.Driver()).String() {
			case "*sqlite3.SQLiteDriver":
				dd.driverName = "sqlite3"
				dd.dialect = dialect.NewSqliteDialect()
			case "*pq.Driver", "*stdlib.Driver":
				dd.driverName = "postgres"
				dd.dialect = dialect.NewPostgresqlDialect()
			case "*mysql.MySQLDriver":
				dd.driverName = "mysql"
				dd.sqPlaceholderFormat = squirrel.Question
				dd.sqStmtBuilder = squirrel.StatementBuilderType(builder.EmptyBuilder).PlaceholderFormat(squirrel.Question)
				dd.dialect = dialect.NewMysqlDialect()
			case "*clickhouse.stdDriver":
				dd.driverName = "clickhouse"
				dd.dialect = dialect.NewClickhouseDialect()
			}
			dd.DB = sqlx.NewDb(d, dd.driverName)
			dd.QueryRunner = queryRunnerWrap{dd.DB}
		}
	}
	if len(queryRunner) != 0 && queryRunner[0] != nil {
		dd.QueryRunner = queryRunner[0]
	}
	dd.queryTraceScan = dd._queryTraceScan
	dd.prepareTrace = dd._prepareTrace
	if DefaultDB == nil {
		DefaultDB = dd
	}
	return dd
}

func (d *db) Dialect() preformShare.IDialect {
	return d.dialect
}
func (d *db) DriverName() string {
	return d.driverName
}

func (d *db) QueryTraceScan(ctx context.Context, query string, args ...interface{}) (rows IRows, err error) {
	return d.queryTraceScan(ctx, query, args...)
}

func (d *db) PrepareTrace(ctx context.Context, query string) (IStmt, error) {
	return d.prepareTrace(ctx, query)
}

func (d *db) _prepareTrace(ctx context.Context, query string) (IStmt, error) {
	return nil, errors.New("no tracer")
}

func (d *db) _queryTraceScan(ctx context.Context, query string, args ...interface{}) (rows IRows, err error) {
	return nil, errors.New("no tracer")
}

func errorLog(driver, msg string, err error) {
	fmt.Printf("Preform Error %s %s %v\n", driver, msg, err)
}

func (d *db) SetTracer(tracerNilToOff ITracer) {
	if d.healthCtxCancel != nil {
		d.healthCtxCancel()
	}
	if tracerNilToOff == nil {
		if traceRunner, ok := d.QueryRunner.(*dbWithTracer); ok {
			d.QueryRunner = traceRunner.QueryRunner
		} else {
			d.QueryRunner = queryRunnerWrap{d.DB}
		}
		d.errorLogger = errorLog
		d.queryTraceScan = d._queryTraceScan
		d.prepareTrace = d._prepareTrace
	} else {
		dwt := &dbWithTracer{QueryRunner: d.QueryRunner, driverName: d.driverName, tracer: tracerNilToOff}
		d.QueryRunner = dwt
		d.tracer = tracerNilToOff
		d.errorLogger = tracerNilToOff.Error
		d.queryTraceScan = dwt.QueryTraceScan
		d.prepareTrace = dwt.PrepareTrace
		var healthCtx context.Context
		healthCtx, d.healthCtxCancel = context.WithCancel(d.ctx)
		go tracerNilToOff.HealthLoop(healthCtx, d)
	}
}

func (d *db) Db() *db {
	return d
}

func (d *db) BeginTx(ctx context.Context, opt ...*sql.TxOptions) (*Tx, error) {
	var (
		tx  *sqlx.Tx
		err error
		dwt *dbWithTracer
		ok  bool
	)
	if len(opt) == 0 {
		tx, err = d.DB.BeginTxx(ctx, nil)
	} else {
		tx, err = d.DB.BeginTxx(ctx, opt[0])
	}
	if err != nil {
		return nil, err
	}
	if dwt, ok = d.QueryRunner.(*dbWithTracer); ok {
		dwt = &dbWithTracer{QueryRunner: queryRunnerWrap{tx}, driverName: d.driverName, tracer: d.tracer, txId: strconv.FormatInt(rand.Int63(), 36)}
		return &Tx{QueryRunner: queryRunnerWrap{dwt}, tx: tx, db: d, queryTraceScan: dwt.QueryTraceScan, prepareTrace: dwt.PrepareTrace}, nil
	} else if _, ok := d.QueryRunner.BaseRunner().(preformShare.ITestQueryRunner); ok {
		return &Tx{QueryRunner: d.QueryRunner, tx: tx, db: d, queryTraceScan: d.queryTraceScan, prepareTrace: d.prepareTrace}, nil
	} else {
		return &Tx{QueryRunner: queryRunnerWrap{tx}, tx: tx, db: d, prepareTrace: d.prepareTrace, queryTraceScan: d.queryTraceScan}, nil
	}
}

func (d *db) GetDialect() preformShare.IDialect {
	return d.dialect
}

func (d *db) Error(msg string, err error) {
	d.errorLogger(d.driverName, msg, err)
}
