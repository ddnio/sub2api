// backend/ent/schema/payment_plan.go
package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentPlan holds the schema definition for the PaymentPlan entity.
type PaymentPlan struct {
	ent.Schema
}

func (PaymentPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_plans"},
	}
}

func (PaymentPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (PaymentPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Unique(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.String("badge").
			MaxLen(20).
			Optional().
			Nillable(),
		field.Int64("group_id"),
		field.Int("duration_days").
			Positive(),
		field.Float("price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("original_price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Optional().
			Nillable(),
		field.Int("sort_order").
			Default(0).
			Min(0),
		field.Bool("is_active").
			Default(true),
	}
}

func (PaymentPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("payment_plans").
			Field("group_id").
			Required().
			Unique(),
	}
}

func (PaymentPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
		index.Fields("is_active", "sort_order"),
	}
}
