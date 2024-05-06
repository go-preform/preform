package preform

import (
	"context"
	sql "database/sql"
	preformShare "github.com/go-preform/preform/share"
)

type ITracer interface {
	Error(driver, msg string, err error)
	Trace(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(err error) func(fetched bool, err error))
	TraceExec(ctx context.Context, driver, query, txId string, args ...any) (context.Context, func(res sql.Result, err error))
	HealthLoop(ctx context.Context, db DB)
	SetLv(logLv preformShare.LogLv)
}
