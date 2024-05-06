package preformTestUtil

import (
	"context"
	"database/sql"
	"database/sql/driver"
)

func NewTestDB(driver string) *sql.DB {
	return sql.OpenDB(&TestSqlConnector{driverName: driver})
}

type TestSqlConnector struct {
	driverName string
}

func (t TestSqlConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return &testSqlConnectorConn{}, nil
}

func (t TestSqlConnector) Driver() driver.Driver {
	return &TestSqlConnectorDriver{name: t.driverName}
}

type TestSqlConnectorDriver struct {
	name string
}

func (t TestSqlConnectorDriver) TestDriverName() string {
	return t.name
}

func (t TestSqlConnectorDriver) Open(name string) (driver.Conn, error) {
	return &testSqlConnectorConn{}, nil
}

type testSqlConnectorConn struct{}

func (t testSqlConnectorConn) Prepare(query string) (driver.Stmt, error) {
	return &testSqlConnectorStmt{}, nil
}

func (t testSqlConnectorConn) Close() error {
	return nil
}

func (t testSqlConnectorConn) Begin() (driver.Tx, error) {
	return &testSqlConnectorTx{}, nil
}

type testSqlConnectorStmt struct{}

func (t testSqlConnectorStmt) Close() error {
	return nil
}

func (t testSqlConnectorStmt) NumInput() int {
	return 0
}

func (t testSqlConnectorStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &testSqlConnectorResult{}, nil
}

func (t testSqlConnectorStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &testSqlConnectorRows{}, nil
}

type testSqlConnectorResult struct{}

func (t testSqlConnectorResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (t testSqlConnectorResult) RowsAffected() (int64, error) {
	return 0, nil
}

type testSqlConnectorRows struct{}

func (t testSqlConnectorRows) Columns() []string {
	return []string{}
}

func (t testSqlConnectorRows) Close() error {
	return nil
}

func (t testSqlConnectorRows) Next(dest []driver.Value) error {
	return nil
}

type testSqlConnectorTx struct{}

func (t testSqlConnectorTx) Commit() error {
	return nil
}

func (t testSqlConnectorTx) Rollback() error {
	return nil
}
