// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/go-preform/preform/benchmark/ent/predicate"
	"github.com/go-preform/preform/benchmark/ent/testa"
)

// TestADelete is the builder for deleting a TestA entity.
type TestADelete struct {
	config
	hooks    []Hook
	mutation *TestAMutation
}

// Where appends a list predicates to the TestADelete builder.
func (ta *TestADelete) Where(ps ...predicate.TestA) *TestADelete {
	ta.mutation.Where(ps...)
	return ta
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ta *TestADelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ta.sqlExec, ta.mutation, ta.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ta *TestADelete) ExecX(ctx context.Context) int {
	n, err := ta.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ta *TestADelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(testa.Table, sqlgraph.NewFieldSpec(testa.FieldID, field.TypeInt))
	if ps := ta.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ta.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ta.mutation.done = true
	return affected, err
}

// TestADeleteOne is the builder for deleting a single TestA entity.
type TestADeleteOne struct {
	ta *TestADelete
}

// Where appends a list predicates to the TestADelete builder.
func (tao *TestADeleteOne) Where(ps ...predicate.TestA) *TestADeleteOne {
	tao.ta.mutation.Where(ps...)
	return tao
}

// Exec executes the deletion query.
func (tao *TestADeleteOne) Exec(ctx context.Context) error {
	n, err := tao.ta.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{testa.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (tao *TestADeleteOne) ExecX(ctx context.Context) {
	if err := tao.Exec(ctx); err != nil {
		panic(err)
	}
}
