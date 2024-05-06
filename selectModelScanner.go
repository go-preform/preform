package preform

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type IModelScanner[B any] interface {
	ScanCount(rows IRows) (uint64, error)
	ScanAny(rows IRows, cols []ICol, max uint64) ([]any, error)
	ScanBodies(rows IRows, cols []ICol, max uint64) ([]B, error)
	ScanBodiesFast(rows IRows, cols []ICol, max uint64) ([]B, error)
	ScanBody(row IRow, cols []ICol) (*B, error)
	ScanRaw(rows IRows, colValTpl []func() (any, func(*any) any), max uint64) (*RowsWithCols, error)
	ScanStructs(rows IRows, s any) error
	ToAnyScanner() IModelScanner[any]
}

type IRows interface {
	Scan(dest ...interface{}) error
	Err() error
	Next() bool
	Close() error
	Columns() ([]string, error)
}

type IStmt interface {
	Exec(args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (IRows, error)
	QueryRow(args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, args ...interface{}) (IRows, error)
	Close() error
}
type IRow interface {
	Scan(...any) error
}

type modelScanner[B any] struct {
	bodyCreator func() B
}

func (b modelScanner[B]) ScanCount(rows IRows) (uint64, error) {
	var (
		count  uint64
		subCnt uint64
		err    error
	)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&subCnt)
		if err != nil {
			return 0, err
		}
		count += subCnt
	}
	return count, err
}

func (b modelScanner[B]) ScanAny(rows IRows, cols []ICol, max uint64) ([]any, error) {
	var (
		body       B = b.bodyCreator()
		res          = make([]any, 0, max)
		sortedPtrs []any
		scanPtrs   []any
		err        error
	)
	//defer rows.Close()
	scanPtrs = any(&body).(hasFieldValuePtrs).FieldValuePtrs()
	if l := len(cols); l == 0 {
		sortedPtrs = scanPtrs
	} else {
		sortedPtrs = make([]any, l)
		for i, col := range cols {
			sortedPtrs[i] = scanPtrs[col.GetPos()]
		}
	}
	for rows.Next() {
		err = rows.Scan(sortedPtrs...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		res = append(res, body)
	}
	return res, err
}

func (b modelScanner[B]) ScanBodies(rows IRows, cols []ICol, max uint64) ([]B, error) {
	var (
		body       = b.bodyCreator()
		res        = make([]B, 0, max)
		sortedPtrs []any
		scanPtrs   []any
		err        error
	)
	//defer rows.Close()
	scanPtrs = any(&body).(hasFieldValuePtrs).FieldValuePtrs()
	if l := len(cols); l == 0 {
		sortedPtrs = scanPtrs
	} else {
		sortedPtrs = make([]any, l)
		for i, col := range cols {
			sortedPtrs[i] = scanPtrs[col.GetPos()]
		}
	}
	for rows.Next() {
		err = rows.Scan(sortedPtrs...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		res = append(res, body)
	}
	return res, err
}

func (b modelScanner[B]) ScanBodiesFast(rows IRows, cols []ICol, max uint64) ([]B, error) {
	var (
		body       B = b.bodyCreator()
		res          = make([]B, 0, max)
		scanPtrs   []any
		sortedPtrs []any
		err        error
	)
	//defer rows.Close()
	scanPtrs = any(&body).(hasFieldValuePtrs).FieldValuePtrs()
	if l := len(cols); l == 0 {
		return nil, errors.New("not support body scan, check your select columns")
	} else {
		sortedPtrs = make([]any, l)
		for i, col := range cols {
			sortedPtrs[i] = col.wrapScanner(scanPtrs[col.GetPos()])
		}
	}
	for rows.Next() {
		err = rows.Scan(sortedPtrs...)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		res = append(res, body)
	}
	return res, err
}

func (b modelScanner[B]) ScanBody(row IRow, cols []ICol) (*B, error) {
	var (
		body       B = b.bodyCreator()
		scanPtrs   []any
		sortedPtrs []any
	)
	scanPtrs = any(&body).(hasFieldValuePtrs).FieldValuePtrs()
	if l := len(cols); l == 0 {
		sortedPtrs = scanPtrs
	} else {
		sortedPtrs = make([]any, l)
		for i, col := range cols {
			sortedPtrs[i] = scanPtrs[col.GetPos()]
		}
	}
	err := row.Scan(sortedPtrs...)
	if err != nil {
		return nil, err
	}
	return &body, err
}

func (b modelScanner[B]) ScanRaw(rows IRows, colValTpl []func() (any, func(*any) any), max uint64) (*RowsWithCols, error) {
	var (
		res     *RowsWithCols
		rowPtr  []any
		row     []any
		cl      int
		err     error
		newPtrs func() ([]any, []any)
	)
	defer rows.Close()

	res, err = newRowsWithCols(rows, max)
	if err != nil {
		return res, err
	}
	if cl = len(res.Columns); cl == len(colValTpl) {
		newPtrs = func() ([]any, []any) {
			var (
				rowPtr  []any
				row     []any
				wrapper func(*any) any
			)
			rowPtr = make([]any, cl)
			row = make([]any, cl)
			for i := range colValTpl {
				row[i], wrapper = colValTpl[i]()
				rowPtr[i] = wrapper(&row[i])
			}
			return rowPtr, row
		}
	} else {
		newPtrs = func() ([]any, []any) {
			rowPtr := make([]any, cl)
			row := make([]any, cl)
			for i := range rowPtr {
				row[i] = new(any)
				rowPtr[i] = &row[i]
			}
			return rowPtr, row
		}
	}
	for rows.Next() {
		rowPtr, row = newPtrs()
		err = rows.Scan(rowPtr...)
		if err != nil {
			return res, err
		}
		res.Rows = append(res.Rows, append([]any{}, row...))
	}
	return res, nil
}

func (b modelScanner[B]) ScanStructs(rows IRows, s any) error {
	defer rows.Close()
	return sqlx.StructScan(rows, s)
}

func (b modelScanner[B]) ToAnyScanner() IModelScanner[any] {
	return modelScanner[any]{}
}
