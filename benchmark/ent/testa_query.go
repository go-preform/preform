// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/go-preform/preform/benchmark/ent/predicate"
	"github.com/go-preform/preform/benchmark/ent/testa"
	"github.com/go-preform/preform/benchmark/ent/testb"
)

// TestAQuery is the builder for querying TestA entities.
type TestAQuery struct {
	config
	ctx        *QueryContext
	order      []testa.OrderOption
	inters     []Interceptor
	predicates []predicate.TestA
	withTestBs *TestBQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the TestAQuery builder.
func (ta *TestAQuery) Where(ps ...predicate.TestA) *TestAQuery {
	ta.predicates = append(ta.predicates, ps...)
	return ta
}

// Limit the number of records to be returned by this query.
func (ta *TestAQuery) Limit(limit int) *TestAQuery {
	ta.ctx.Limit = &limit
	return ta
}

// Offset to start from.
func (ta *TestAQuery) Offset(offset int) *TestAQuery {
	ta.ctx.Offset = &offset
	return ta
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (ta *TestAQuery) Unique(unique bool) *TestAQuery {
	ta.ctx.Unique = &unique
	return ta
}

// Order specifies how the records should be ordered.
func (ta *TestAQuery) Order(o ...testa.OrderOption) *TestAQuery {
	ta.order = append(ta.order, o...)
	return ta
}

// QueryTestBs chains the current query on the "test_bs" edge.
func (ta *TestAQuery) QueryTestBs() *TestBQuery {
	query := (&TestBClient{config: ta.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := ta.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := ta.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(testa.Table, testa.FieldID, selector),
			sqlgraph.To(testb.Table, testb.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, testa.TestBsTable, testa.TestBsColumn),
		)
		fromU = sqlgraph.SetNeighbors(ta.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first TestA entity from the query.
// Returns a *NotFoundError when no TestA was found.
func (ta *TestAQuery) First(ctx context.Context) (*TestA, error) {
	nodes, err := ta.Limit(1).All(setContextOp(ctx, ta.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{testa.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (ta *TestAQuery) FirstX(ctx context.Context) *TestA {
	node, err := ta.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first TestA ID from the query.
// Returns a *NotFoundError when no TestA ID was found.
func (ta *TestAQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = ta.Limit(1).IDs(setContextOp(ctx, ta.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{testa.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (ta *TestAQuery) FirstIDX(ctx context.Context) int {
	id, err := ta.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single TestA entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one TestA entity is found.
// Returns a *NotFoundError when no TestA entities are found.
func (ta *TestAQuery) Only(ctx context.Context) (*TestA, error) {
	nodes, err := ta.Limit(2).All(setContextOp(ctx, ta.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{testa.Label}
	default:
		return nil, &NotSingularError{testa.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (ta *TestAQuery) OnlyX(ctx context.Context) *TestA {
	node, err := ta.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only TestA ID in the query.
// Returns a *NotSingularError when more than one TestA ID is found.
// Returns a *NotFoundError when no entities are found.
func (ta *TestAQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = ta.Limit(2).IDs(setContextOp(ctx, ta.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{testa.Label}
	default:
		err = &NotSingularError{testa.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (ta *TestAQuery) OnlyIDX(ctx context.Context) int {
	id, err := ta.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of TestAs.
func (ta *TestAQuery) All(ctx context.Context) ([]*TestA, error) {
	ctx = setContextOp(ctx, ta.ctx, "All")
	if err := ta.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*TestA, *TestAQuery]()
	return withInterceptors[[]*TestA](ctx, ta, qr, ta.inters)
}

// AllX is like All, but panics if an error occurs.
func (ta *TestAQuery) AllX(ctx context.Context) []*TestA {
	nodes, err := ta.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of TestA IDs.
func (ta *TestAQuery) IDs(ctx context.Context) (ids []int, err error) {
	if ta.ctx.Unique == nil && ta.path != nil {
		ta.Unique(true)
	}
	ctx = setContextOp(ctx, ta.ctx, "IDs")
	if err = ta.Select(testa.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (ta *TestAQuery) IDsX(ctx context.Context) []int {
	ids, err := ta.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (ta *TestAQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, ta.ctx, "Count")
	if err := ta.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, ta, querierCount[*TestAQuery](), ta.inters)
}

// CountX is like Count, but panics if an error occurs.
func (ta *TestAQuery) CountX(ctx context.Context) int {
	count, err := ta.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (ta *TestAQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, ta.ctx, "Exist")
	switch _, err := ta.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (ta *TestAQuery) ExistX(ctx context.Context) bool {
	exist, err := ta.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the TestAQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (ta *TestAQuery) Clone() *TestAQuery {
	if ta == nil {
		return nil
	}
	return &TestAQuery{
		config:     ta.config,
		ctx:        ta.ctx.Clone(),
		order:      append([]testa.OrderOption{}, ta.order...),
		inters:     append([]Interceptor{}, ta.inters...),
		predicates: append([]predicate.TestA{}, ta.predicates...),
		withTestBs: ta.withTestBs.Clone(),
		// clone intermediate query.
		sql:  ta.sql.Clone(),
		path: ta.path,
	}
}

// WithTestBs tells the query-builder to eager-load the nodes that are connected to
// the "test_bs" edge. The optional arguments are used to configure the query builder of the edge.
func (ta *TestAQuery) WithTestBs(opts ...func(*TestBQuery)) *TestAQuery {
	query := (&TestBClient{config: ta.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	ta.withTestBs = query
	return ta
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Name string `json:"name,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.TestA.Query().
//		GroupBy(testa.FieldName).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (ta *TestAQuery) GroupBy(field string, fields ...string) *TestAGroupBy {
	ta.ctx.Fields = append([]string{field}, fields...)
	grbuild := &TestAGroupBy{build: ta}
	grbuild.flds = &ta.ctx.Fields
	grbuild.label = testa.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Name string `json:"name,omitempty"`
//	}
//
//	client.TestA.Query().
//		Select(testa.FieldName).
//		Scan(ctx, &v)
func (ta *TestAQuery) Select(fields ...string) *TestASelect {
	ta.ctx.Fields = append(ta.ctx.Fields, fields...)
	sbuild := &TestASelect{TestAQuery: ta}
	sbuild.label = testa.Label
	sbuild.flds, sbuild.scan = &ta.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a TestASelect configured with the given aggregations.
func (ta *TestAQuery) Aggregate(fns ...AggregateFunc) *TestASelect {
	return ta.Select().Aggregate(fns...)
}

func (ta *TestAQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range ta.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, ta); err != nil {
				return err
			}
		}
	}
	for _, f := range ta.ctx.Fields {
		if !testa.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if ta.path != nil {
		prev, err := ta.path(ctx)
		if err != nil {
			return err
		}
		ta.sql = prev
	}
	return nil
}

func (ta *TestAQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*TestA, error) {
	var (
		nodes       = []*TestA{}
		_spec       = ta.querySpec()
		loadedTypes = [1]bool{
			ta.withTestBs != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*TestA).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &TestA{config: ta.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, ta.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := ta.withTestBs; query != nil {
		if err := ta.loadTestBs(ctx, query, nodes,
			func(n *TestA) { n.Edges.TestBs = []*TestB{} },
			func(n *TestA, e *TestB) { n.Edges.TestBs = append(n.Edges.TestBs, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (ta *TestAQuery) loadTestBs(ctx context.Context, query *TestBQuery, nodes []*TestA, init func(*TestA), assign func(*TestA, *TestB)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*TestA)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(testb.FieldAID)
	}
	query.Where(predicate.TestB(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(testa.TestBsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.AID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "a_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (ta *TestAQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := ta.querySpec()
	_spec.Node.Columns = ta.ctx.Fields
	if len(ta.ctx.Fields) > 0 {
		_spec.Unique = ta.ctx.Unique != nil && *ta.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, ta.driver, _spec)
}

func (ta *TestAQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(testa.Table, testa.Columns, sqlgraph.NewFieldSpec(testa.FieldID, field.TypeInt))
	_spec.From = ta.sql
	if unique := ta.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if ta.path != nil {
		_spec.Unique = true
	}
	if fields := ta.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, testa.FieldID)
		for i := range fields {
			if fields[i] != testa.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := ta.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := ta.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := ta.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := ta.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (ta *TestAQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(ta.driver.Dialect())
	t1 := builder.Table(testa.Table)
	columns := ta.ctx.Fields
	if len(columns) == 0 {
		columns = testa.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if ta.sql != nil {
		selector = ta.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if ta.ctx.Unique != nil && *ta.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range ta.predicates {
		p(selector)
	}
	for _, p := range ta.order {
		p(selector)
	}
	if offset := ta.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := ta.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// TestAGroupBy is the group-by builder for TestA entities.
type TestAGroupBy struct {
	selector
	build *TestAQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (tab *TestAGroupBy) Aggregate(fns ...AggregateFunc) *TestAGroupBy {
	tab.fns = append(tab.fns, fns...)
	return tab
}

// Scan applies the selector query and scans the result into the given value.
func (tab *TestAGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tab.build.ctx, "GroupBy")
	if err := tab.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TestAQuery, *TestAGroupBy](ctx, tab.build, tab, tab.build.inters, v)
}

func (tab *TestAGroupBy) sqlScan(ctx context.Context, root *TestAQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(tab.fns))
	for _, fn := range tab.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*tab.flds)+len(tab.fns))
		for _, f := range *tab.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*tab.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tab.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// TestASelect is the builder for selecting fields of TestA entities.
type TestASelect struct {
	*TestAQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ta *TestASelect) Aggregate(fns ...AggregateFunc) *TestASelect {
	ta.fns = append(ta.fns, fns...)
	return ta
}

// Scan applies the selector query and scans the result into the given value.
func (ta *TestASelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ta.ctx, "Select")
	if err := ta.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TestAQuery, *TestASelect](ctx, ta.TestAQuery, ta, ta.inters, v)
}

func (ta *TestASelect) sqlScan(ctx context.Context, root *TestAQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ta.fns))
	for _, fn := range ta.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ta.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ta.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
