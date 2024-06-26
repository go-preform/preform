// Code generated by ent, DO NOT EDIT.

package testc

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/go-preform/preform/benchmark/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldID, id))
}

// BID applies equality check predicate on the "b_id" field. It's identical to BIDEQ.
func BID(v int) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldBID, v))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldName, v))
}

// Int4 applies equality check predicate on the "int4" field. It's identical to Int4EQ.
func Int4(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldInt4, v))
}

// Int8 applies equality check predicate on the "int8" field. It's identical to Int8EQ.
func Int8(v int) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldInt8, v))
}

// Float4 applies equality check predicate on the "float4" field. It's identical to Float4EQ.
func Float4(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldFloat4, v))
}

// Float8 applies equality check predicate on the "float8" field. It's identical to Float8EQ.
func Float8(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldFloat8, v))
}

// Bool applies equality check predicate on the "bool" field. It's identical to BoolEQ.
func Bool(v bool) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldBool, v))
}

// Text applies equality check predicate on the "text" field. It's identical to TextEQ.
func Text(v string) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldText, v))
}

// Time applies equality check predicate on the "time" field. It's identical to TimeEQ.
func Time(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldTime, v))
}

// BIDEQ applies the EQ predicate on the "b_id" field.
func BIDEQ(v int) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldBID, v))
}

// BIDNEQ applies the NEQ predicate on the "b_id" field.
func BIDNEQ(v int) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldBID, v))
}

// BIDIn applies the In predicate on the "b_id" field.
func BIDIn(vs ...int) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldBID, vs...))
}

// BIDNotIn applies the NotIn predicate on the "b_id" field.
func BIDNotIn(vs ...int) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldBID, vs...))
}

// BIDIsNil applies the IsNil predicate on the "b_id" field.
func BIDIsNil() predicate.TestC {
	return predicate.TestC(sql.FieldIsNull(FieldBID))
}

// BIDNotNil applies the NotNil predicate on the "b_id" field.
func BIDNotNil() predicate.TestC {
	return predicate.TestC(sql.FieldNotNull(FieldBID))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.TestC {
	return predicate.TestC(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.TestC {
	return predicate.TestC(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.TestC {
	return predicate.TestC(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.TestC {
	return predicate.TestC(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.TestC {
	return predicate.TestC(sql.FieldContainsFold(FieldName, v))
}

// Int4EQ applies the EQ predicate on the "int4" field.
func Int4EQ(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldInt4, v))
}

// Int4NEQ applies the NEQ predicate on the "int4" field.
func Int4NEQ(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldInt4, v))
}

// Int4In applies the In predicate on the "int4" field.
func Int4In(vs ...int32) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldInt4, vs...))
}

// Int4NotIn applies the NotIn predicate on the "int4" field.
func Int4NotIn(vs ...int32) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldInt4, vs...))
}

// Int4GT applies the GT predicate on the "int4" field.
func Int4GT(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldInt4, v))
}

// Int4GTE applies the GTE predicate on the "int4" field.
func Int4GTE(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldInt4, v))
}

// Int4LT applies the LT predicate on the "int4" field.
func Int4LT(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldInt4, v))
}

// Int4LTE applies the LTE predicate on the "int4" field.
func Int4LTE(v int32) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldInt4, v))
}

// Int8EQ applies the EQ predicate on the "int8" field.
func Int8EQ(v int) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldInt8, v))
}

// Int8NEQ applies the NEQ predicate on the "int8" field.
func Int8NEQ(v int) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldInt8, v))
}

// Int8In applies the In predicate on the "int8" field.
func Int8In(vs ...int) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldInt8, vs...))
}

// Int8NotIn applies the NotIn predicate on the "int8" field.
func Int8NotIn(vs ...int) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldInt8, vs...))
}

// Int8GT applies the GT predicate on the "int8" field.
func Int8GT(v int) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldInt8, v))
}

// Int8GTE applies the GTE predicate on the "int8" field.
func Int8GTE(v int) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldInt8, v))
}

// Int8LT applies the LT predicate on the "int8" field.
func Int8LT(v int) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldInt8, v))
}

// Int8LTE applies the LTE predicate on the "int8" field.
func Int8LTE(v int) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldInt8, v))
}

// Float4EQ applies the EQ predicate on the "float4" field.
func Float4EQ(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldFloat4, v))
}

// Float4NEQ applies the NEQ predicate on the "float4" field.
func Float4NEQ(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldFloat4, v))
}

// Float4In applies the In predicate on the "float4" field.
func Float4In(vs ...float32) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldFloat4, vs...))
}

// Float4NotIn applies the NotIn predicate on the "float4" field.
func Float4NotIn(vs ...float32) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldFloat4, vs...))
}

// Float4GT applies the GT predicate on the "float4" field.
func Float4GT(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldFloat4, v))
}

// Float4GTE applies the GTE predicate on the "float4" field.
func Float4GTE(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldFloat4, v))
}

// Float4LT applies the LT predicate on the "float4" field.
func Float4LT(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldFloat4, v))
}

// Float4LTE applies the LTE predicate on the "float4" field.
func Float4LTE(v float32) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldFloat4, v))
}

// Float8EQ applies the EQ predicate on the "float8" field.
func Float8EQ(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldFloat8, v))
}

// Float8NEQ applies the NEQ predicate on the "float8" field.
func Float8NEQ(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldFloat8, v))
}

// Float8In applies the In predicate on the "float8" field.
func Float8In(vs ...float64) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldFloat8, vs...))
}

// Float8NotIn applies the NotIn predicate on the "float8" field.
func Float8NotIn(vs ...float64) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldFloat8, vs...))
}

// Float8GT applies the GT predicate on the "float8" field.
func Float8GT(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldFloat8, v))
}

// Float8GTE applies the GTE predicate on the "float8" field.
func Float8GTE(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldFloat8, v))
}

// Float8LT applies the LT predicate on the "float8" field.
func Float8LT(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldFloat8, v))
}

// Float8LTE applies the LTE predicate on the "float8" field.
func Float8LTE(v float64) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldFloat8, v))
}

// BoolEQ applies the EQ predicate on the "bool" field.
func BoolEQ(v bool) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldBool, v))
}

// BoolNEQ applies the NEQ predicate on the "bool" field.
func BoolNEQ(v bool) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldBool, v))
}

// TextEQ applies the EQ predicate on the "text" field.
func TextEQ(v string) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldText, v))
}

// TextNEQ applies the NEQ predicate on the "text" field.
func TextNEQ(v string) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldText, v))
}

// TextIn applies the In predicate on the "text" field.
func TextIn(vs ...string) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldText, vs...))
}

// TextNotIn applies the NotIn predicate on the "text" field.
func TextNotIn(vs ...string) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldText, vs...))
}

// TextGT applies the GT predicate on the "text" field.
func TextGT(v string) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldText, v))
}

// TextGTE applies the GTE predicate on the "text" field.
func TextGTE(v string) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldText, v))
}

// TextLT applies the LT predicate on the "text" field.
func TextLT(v string) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldText, v))
}

// TextLTE applies the LTE predicate on the "text" field.
func TextLTE(v string) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldText, v))
}

// TextContains applies the Contains predicate on the "text" field.
func TextContains(v string) predicate.TestC {
	return predicate.TestC(sql.FieldContains(FieldText, v))
}

// TextHasPrefix applies the HasPrefix predicate on the "text" field.
func TextHasPrefix(v string) predicate.TestC {
	return predicate.TestC(sql.FieldHasPrefix(FieldText, v))
}

// TextHasSuffix applies the HasSuffix predicate on the "text" field.
func TextHasSuffix(v string) predicate.TestC {
	return predicate.TestC(sql.FieldHasSuffix(FieldText, v))
}

// TextEqualFold applies the EqualFold predicate on the "text" field.
func TextEqualFold(v string) predicate.TestC {
	return predicate.TestC(sql.FieldEqualFold(FieldText, v))
}

// TextContainsFold applies the ContainsFold predicate on the "text" field.
func TextContainsFold(v string) predicate.TestC {
	return predicate.TestC(sql.FieldContainsFold(FieldText, v))
}

// TimeEQ applies the EQ predicate on the "time" field.
func TimeEQ(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldEQ(FieldTime, v))
}

// TimeNEQ applies the NEQ predicate on the "time" field.
func TimeNEQ(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldNEQ(FieldTime, v))
}

// TimeIn applies the In predicate on the "time" field.
func TimeIn(vs ...time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldIn(FieldTime, vs...))
}

// TimeNotIn applies the NotIn predicate on the "time" field.
func TimeNotIn(vs ...time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldNotIn(FieldTime, vs...))
}

// TimeGT applies the GT predicate on the "time" field.
func TimeGT(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldGT(FieldTime, v))
}

// TimeGTE applies the GTE predicate on the "time" field.
func TimeGTE(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldGTE(FieldTime, v))
}

// TimeLT applies the LT predicate on the "time" field.
func TimeLT(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldLT(FieldTime, v))
}

// TimeLTE applies the LTE predicate on the "time" field.
func TimeLTE(v time.Time) predicate.TestC {
	return predicate.TestC(sql.FieldLTE(FieldTime, v))
}

// HasTestB applies the HasEdge predicate on the "test_b" edge.
func HasTestB() predicate.TestC {
	return predicate.TestC(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, TestBTable, TestBColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTestBWith applies the HasEdge predicate on the "test_b" edge with a given conditions (other predicates).
func HasTestBWith(preds ...predicate.TestB) predicate.TestC {
	return predicate.TestC(func(s *sql.Selector) {
		step := newTestBStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.TestC) predicate.TestC {
	return predicate.TestC(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.TestC) predicate.TestC {
	return predicate.TestC(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.TestC) predicate.TestC {
	return predicate.TestC(sql.NotPredicates(p))
}
