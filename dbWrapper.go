package preform

import (
	"context"
	"github.com/go-preform/preform/dialect"
	preformShare "github.com/go-preform/preform/share"
)

type queryRunnerWrap struct {
	preformShare.DbQueryRunner
}

func (d queryRunnerWrap) InsertAndReturnAutoId(ctx context.Context, lastIdMethod preformShare.SqlDialectLastInsertIdMethod, query string, args ...interface{}) (lastId int64, err error) {
	if lastIdMethod == dialect.LastInsertIdMethodByRes {
		res, err := d.ExecContext(ctx, query, args...)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	} else {
		err = d.QueryRowContext(ctx, query, args...).Scan(&lastId)
		return
	}
}

func (d queryRunnerWrap) BaseRunner() preformShare.DbQueryRunner {
	return d.DbQueryRunner
}

func (d queryRunnerWrap) RelatedFactory([]preformShare.IQueryFactory) preformShare.QueryRunner {
	return d
}
