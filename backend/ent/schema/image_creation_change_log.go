package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ImageCreationChangeLog struct {
	ent.Schema
}

func (ImageCreationChangeLog) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "image_creation_change_logs"}}
}

func (ImageCreationChangeLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("actor_user_id"),
		field.String("action").MaxLen(32),
		field.String("target_type").MaxLen(16),
		field.String("target_id").MaxLen(64),
		field.JSON("metadata", map[string]any{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ImageCreationChangeLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("actor", User.Type).Ref("image_creation_change_logs").Field("actor_user_id").Unique().Required(),
	}
}

func (ImageCreationChangeLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("actor_user_id", "created_at"),
		index.Fields("target_type", "target_id", "created_at"),
	}
}
