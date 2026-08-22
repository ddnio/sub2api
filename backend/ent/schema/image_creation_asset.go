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

type ImageCreationAsset struct {
	ent.Schema
}

func (ImageCreationAsset) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "image_creation_assets"}}
}

func (ImageCreationAsset) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").MaxLen(64).Immutable(),
		field.Bytes("content").SchemaType(map[string]string{dialect.Postgres: "bytea"}),
		field.String("content_type").MaxLen(32),
		field.Int("byte_size").Positive(),
		field.Int("width").Positive(),
		field.Int("height").Positive(),
		field.String("source_type").MaxLen(16),
		field.String("source_provider").Optional().Nillable().MaxLen(64),
		field.String("source_model").Optional().Nillable().MaxLen(120),
		field.Int64("created_by"),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ImageCreationAsset) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("creator", User.Type).Ref("image_creation_assets").Field("created_by").Unique().Required(),
		edge.To("draft_templates", ImageCreationTemplate.Type),
		edge.To("published_templates", ImageCreationTemplate.Type),
	}
}

func (ImageCreationAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_by", "created_at"),
	}
}
