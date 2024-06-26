// Code generated by entimport, DO NOT EDIT.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type TestC struct {
	ent.Schema
}

func (TestC) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.Int("b_id").Optional(), field.String("name"), field.Int32("int4"), field.Int("int8"), field.Float32("float4"), field.Float("float8"), field.Bool("bool"), field.String("text"), field.Time("time")}
}
func (TestC) Edges() []ent.Edge {
	return []ent.Edge{edge.From("test_b", TestB.Type).Ref("test_cs").Unique().Field("b_id")}
}
func (TestC) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "test_c"}}
}
