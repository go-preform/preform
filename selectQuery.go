package preform

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

func (b *SelectQuery[B]) Ctx(ctx context.Context) *SelectQuery[B] {
	b.ctx = ctx
	return b
}

type hasFieldValuePtrs interface {
	FieldValuePtrs() []any
}

func (b *SelectQuery[B]) Prepare() (*SelectQuery[B], error) {

	q, _, err := b.ToSql()
	if err != nil {
		return b, err
	}

	b.prepared, err = b.db.RelatedFactory(b.relatedFactoriesForCache).PreparexContext(b.ctx, q)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (b SelectQuery[B]) Query() (*sql.Rows, error) {
	q, a, err := b.ToSql()
	if err != nil {
		return nil, err
	}
	if b.prepared != nil {
		return b.prepared.QueryContext(b.ctx, a...)
	}
	return b.db.RelatedFactory(b.relatedFactoriesForCache).QueryContext(b.ctx, q, a...)
}

func (b SelectQuery[B]) Queryx() (*sqlx.Rows, error) {
	q, a, err := b.ToSql()
	if err != nil {
		return nil, err
	}
	if b.prepared != nil {
		return b.prepared.QueryxContext(b.ctx, a...)
	}
	return b.db.QueryxContext(b.ctx, q, a...)
}

type hasScan interface {
	Scan(...any) error
}

type sqlRow struct {
	hasScan
	err error
}

func (r *sqlRow) Err() error {
	return r.err
}

func (r *sqlRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return r.hasScan.Scan(dest...)
}

func (b SelectQuery[B]) QueryRow() *sqlRow {
	q, a, err := b.ToSql()
	if err != nil {
		return &sqlRow{err: err}
	}
	if b.prepared != nil {
		return &sqlRow{hasScan: b.prepared.QueryRowContext(b.ctx, a...)}
	}
	return &sqlRow{hasScan: b.db.RelatedFactory(b.relatedFactoriesForCache).QueryRowContext(b.ctx, q, a...)}
}

func (b SelectQuery[B]) QueryRowx() *sqlRow {
	q, a, err := b.ToSql()
	if err != nil {
		return &sqlRow{err: err}
	}
	if b.prepared != nil {
		return &sqlRow{hasScan: b.prepared.QueryRowxContext(b.ctx, a...)}
	}
	return &sqlRow{hasScan: b.db.QueryRowxContext(b.ctx, q, a...)}
}

func (b SelectQuery[B]) Count(col ...string) (uint64, error) {
	var (
		rows *sql.Rows
	)
	if len(col) == 0 {
		col = []string{"*"}
	}

	q, a, err := b.RemoveColumns().Column(fmt.Sprintf("COUNT(%s)", col[0])).ToSql()
	if err != nil {
		return 0, err
	}
	rows, err = b.db.QueryContext(b.ctx, q, a...)
	if err != nil {
		return 0, err
	}
	return b.scanner.ScanCount(rows)
}

func (b SelectQuery[B]) QueryAnys() ([]any, error) {
	if b.queryFactory == nil {
		return nil, fmt.Errorf("not support body scan, check your select columns")
	}
	var (
		rows *sql.Rows
		res  []any
	)
	if len(b.Cols) == 0 {
		var (
			cols    = b.queryFactory.Columns()
			colAnys = make([]any, len(cols))
		)
		for i, col := range cols {
			colAnys[i] = col
		}
		b.Columns(colAnys...)
	}
	rows, err := b.Query()
	if err != nil {
		return nil, err
	}
	res, err = b.scanner.ScanAny(rows, b.Cols, b.limit)
	if err != nil {
		return nil, err
	}
	if len(b.eagerLoaders) != 0 {
		var (
			modelPtrs = b.queryFactory.newBodyPtrSlice(len(res))
		)
		for i := range res {
			modelPtrs.Set(i, res[i])
		}
		err = b.eagerLoad(modelPtrs.Slice())
	}
	return res, err
}

func (b SelectQuery[B]) QueryBodies() ([]B, error) {
	if b.queryFactory == nil {
		return nil, fmt.Errorf("not support body scan, check your select columns")
	}
	var (
		rows *sql.Rows
		res  []B
	)
	if len(b.Cols) == 0 {
		var (
			cols    = b.queryFactory.Columns()
			colAnys = make([]any, len(cols))
		)
		for i, col := range cols {
			colAnys[i] = col
		}
		b.Columns(colAnys...)
	}
	rows, err := b.Query()
	if err != nil {
		return nil, err
	}
	res, err = b.scanner.ScanBodies(rows, b.Cols, b.limit)
	if err != nil {
		return nil, err
	}
	if len(b.eagerLoaders) != 0 {
		var (
			modelPtrs = make([]*B, len(res))
		)
		for i := range res {
			modelPtrs[i] = &res[i]
		}
		err = b.eagerLoad(modelPtrs)
	}
	return res, err
}

func (b SelectQuery[B]) QueryBodiesFast() ([]B, error) {
	if b.queryFactory == nil {
		return nil, fmt.Errorf("not support body scan, check your select columns")
	}
	var (
		rows *sql.Rows
		res  []B
	)
	if len(b.Cols) == 0 {
		var (
			cols    = b.queryFactory.Columns()
			colAnys = make([]any, len(cols))
		)
		for i, col := range cols {
			colAnys[i] = col
		}
		b.Columns(colAnys...)
	}
	rows, err := b.Query()
	if err != nil {
		return nil, err
	}
	res, err = b.scanner.ScanBodiesFast(rows, b.Cols, b.limit)
	if err != nil {
		return nil, err
	}
	if len(b.eagerLoaders) != 0 {
		var (
			modelPtrs = make([]*B, len(res))
		)
		for i := range res {
			modelPtrs[i] = &res[i]
		}
		err = b.eagerLoad(modelPtrs, true)
	}
	return res, err
}

func (b SelectQuery[B]) QueryBody() (*B, error) {
	if b.queryFactory == nil {
		return nil, fmt.Errorf("not support body scan, check your select columns")
	}
	var (
		bodyPtr *B
	)
	if len(b.Cols) == 0 {
		var (
			cols    = b.queryFactory.Columns()
			colAnys = make([]any, len(cols))
		)
		for i, col := range cols {
			colAnys[i] = col
		}
		b.Columns(colAnys...)
	}
	q, a, err := b.ToSql()
	if err != nil {
		return nil, err
	}
	bodyPtr, err = b.scanner.ScanBody(b.db.QueryRowContext(b.ctx, q, a...), b.Cols)
	if len(b.eagerLoaders) != 0 {
		if err != nil {
			return nil, err
		}
		return bodyPtr, b.eagerLoad([]*B{bodyPtr})
	}
	return bodyPtr, err
}

func (b SelectQuery[B]) QueryRaw() (*RowsWithCols, error) {
	rows, err := b.Queryx()
	if err != nil {
		return nil, err
	}
	return b.scanner.ScanRaw(rows, b.ColValTpl, b.limit)
}

func (b SelectQuery[B]) QueryStructs(s any) error {
	rows, err := b.Queryx()
	if err != nil {
		return err
	}
	return b.scanner.ScanStructs(rows, s)
}

func (b SelectQuery[B]) GetOne(pkLookup ...any) (*B, error) {
	if len(pkLookup) != 0 {
		if f, ok := b.queryFactory.(IFactory); ok {
			pks := f.Pks()
			for i := range pkLookup {
				b.Where(pks[i].Eq(pkLookup[i]))
			}
		} else {
			return nil, fmt.Errorf("get one pkLookup only work with IFactory")
		}
	}
	return b.Limit(1).QueryBody()
}

func (b SelectQuery[B]) GetAll(cond ...ICond) ([]B, error) {
	for _, c := range cond {
		b.Where(c)
	}
	return b.QueryBodies()
}

func (b SelectQuery[B]) GetAllFast(cond ...ICond) ([]B, error) {
	for _, c := range cond {
		b.Where(c)
	}
	return b.QueryBodiesFast()
}

type RowsWithCols struct {
	Columns    []string
	columnsMap map[string]int
	Rows       [][]any
}

func newRowsWithCols(rows IRows, max uint64) (*RowsWithCols, error) {
	var (
		res = &RowsWithCols{
			Rows: make([][]any, 0, max),
		}
		err error
	)
	res.Columns, err = rows.Columns()
	if err != nil {
		return res, err
	}
	res.columnsMap = make(map[string]int, len(res.Columns))
	for i, col := range res.Columns {
		res.columnsMap[col] = i
	}
	return res, nil
}

func (b RowsWithCols) ToMap() []map[string]any {
	var (
		cl  = len(b.Columns)
		res = make([]map[string]any, len(b.Rows))
		j   int
		col string
	)
	for i, row := range b.Rows {
		res[i] = make(map[string]any, cl)
		for j, col = range b.Columns {
			res[i][col] = row[j]
		}
	}
	return res
}
