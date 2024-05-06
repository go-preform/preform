// Code generated by entimport, DO NOT EDIT.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type TestA struct {
	ent.Schema
}

func (TestA) Fields() []ent.Field {
	return []ent.Field{field.Int("id"), field.String("name"), field.Int32("int4"), field.Int("int8"), field.Float32("float4"), field.Float("float8"), field.Bool("bool"), field.String("text"), field.Time("time")}
}
func (TestA) Edges() []ent.Edge {
	return []ent.Edge{edge.To("test_bs", TestB.Type)}
}
func (TestA) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "test_a"}}
}