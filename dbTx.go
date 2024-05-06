package preform

import (
	"context"
	preformShare "github.com/go-preform/preform/share"
	"github.com/jmoiron/sqlx"
)

type Tx struct {
	preformShare.QueryRunner
	tx             *sqlx.Tx
	db             *db
	queryTraceScan func(ctx context.Context, query string, args ...interface{}) (rows IRows, err error)
	prepareTrace   func(ctx context.Context, query string) (IStmt, error)
}

func (t *Tx) GetDialect() preformShare.IDialect {
	return t.db.dialect
}

func (t *Tx) Db() *db {
	return t.db
}

func (t *Tx) Error(msg string, err error) {
	t.db.errorLogger(t.db.driverName, msg, err)
}

func (t *Tx) QueryTraceScan(ctx context.Context, query string, args ...interface{}) (rows IRows, err error) {
	return t.queryTraceScan(ctx, query, args...)
}

func (t *Tx) PrepareTrace(ctx context.Context, query string) (IStmt, error) {
	return t.prepareTrace(ctx, query)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}
