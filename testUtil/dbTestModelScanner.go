package preformTestUtil

import "github.com/go-preform/preform"

type TestModelScanner[B any] struct {
	ErrorQueue  []error
	CountQueue  []uint64
	AnyQueue    []any
	BodyQueue   []*B
	BodiesQueue [][]B
	RawQueue    []*preform.RowsWithCols
}

func NewTestModelScanner[B any]() *TestModelScanner[B] {
	return &TestModelScanner[B]{}
}

func (t *TestModelScanner[B]) popError() error {
	if len(t.ErrorQueue) == 0 {
		return nil
	}
	err := t.ErrorQueue[0]
	t.ErrorQueue = t.ErrorQueue[1:]
	return err
}

func (t *TestModelScanner[B]) popCount() uint64 {
	if len(t.CountQueue) == 0 {
		return 0
	}
	cnt := t.CountQueue[0]
	t.CountQueue = t.CountQueue[1:]
	return cnt
}

func (t *TestModelScanner[B]) popAny() any {
	if len(t.AnyQueue) == 0 {
		return nil
	}
	any := t.AnyQueue[0]
	t.AnyQueue = t.AnyQueue[1:]
	return any
}

func (t *TestModelScanner[B]) popBody() *B {
	if len(t.BodyQueue) == 0 {
		return nil
	}
	body := t.BodyQueue[0]
	t.BodyQueue = t.BodyQueue[1:]
	return body
}

func (t *TestModelScanner[B]) popBodies() []B {
	if len(t.BodiesQueue) == 0 {
		return nil
	}
	bodies := t.BodiesQueue[0]
	t.BodiesQueue = t.BodiesQueue[1:]
	return bodies
}

func (t *TestModelScanner[B]) popRaw() *preform.RowsWithCols {
	if len(t.RawQueue) == 0 {
		return nil
	}
	raw := t.RawQueue[0]
	t.RawQueue = t.RawQueue[1:]
	return raw
}

func (t *TestModelScanner[B]) ScanCount(rows preform.IRows) (uint64, error) {
	err := t.popError()
	if err != nil {
		return 0, err
	}
	return t.popCount(), err
}

func (t *TestModelScanner[B]) ScanAny(rows preform.IRows, cols []preform.ICol, max uint64) ([]any, error) {
	err := t.popError()
	if err != nil {
		return nil, err
	}
	return []any{t.popAny()}, err
}

func (t *TestModelScanner[B]) ScanBodies(rows preform.IRows, cols []preform.ICol, max uint64) ([]B, error) {
	err := t.popError()
	if err != nil {
		return nil, err
	}
	return t.popBodies(), err
}

func (t *TestModelScanner[B]) ScanBodiesFast(rows preform.IRows, cols []preform.ICol, max uint64) ([]B, error) {
	err := t.popError()
	if err != nil {
		return nil, err
	}
	return t.popBodies(), err
}

func (t *TestModelScanner[B]) ScanBody(row preform.IRow, cols []preform.ICol) (*B, error) {
	err := t.popError()
	if err != nil {
		return nil, err
	}
	return t.popBody(), err
}

func (t *TestModelScanner[B]) ScanRaw(rows preform.IRows, colValTpl []func() (any, func(*any) any), max uint64) (*preform.RowsWithCols, error) {
	err := t.popError()
	if err != nil {
		return nil, err
	}
	return t.popRaw(), err
}

func (t *TestModelScanner[B]) ScanStructs(rows preform.IRows, s any) error {
	ss := s.(*B)
	err := t.popError()
	if err != nil {
		return err
	}
	bPtr := t.popBody()
	if bPtr != nil {
		*ss = *bPtr
	}
	return nil
}

func (t *TestModelScanner[B]) ToAnyScanner() preform.IModelScanner[any] {
	return &TestModelScanner[any]{}
}
