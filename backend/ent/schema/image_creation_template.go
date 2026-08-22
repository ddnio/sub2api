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
	"github.com/Wei-Shaw/sub2api/internal/domain"
)

type ImageCreationTemplate struct {
	ent.Schema
}

func (ImageCreationTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "image_creation_templates"}}
}

func (ImageCreationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("state").MaxLen(16).Default("draft"),
		field.JSON("draft_data", domain.ImageCreationTemplateDocument{}).SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("published_data", &domain.ImageCreationTemplateDocument{}).Optional().SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Int("revision").Default(1).Positive(),
		field.Int("published_version").Default(0).NonNegative(),
		field.String("draft_cover_asset_id").Optional().Nillable().MaxLen(64),
		field.String("published_cover_asset_id").Optional().Nillable().MaxLen(64),
		field.Int16("home_position").Optional().Nillable(),
		field.Int64("created_by"),
		field.Int64("updated_by"),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("published_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ImageCreationTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("draft_cover", ImageCreationAsset.Type).Ref("draft_templates").Field("draft_cover_asset_id").Unique(),
		edge.From("published_cover", ImageCreationAsset.Type).Ref("published_templates").Field("published_cover_asset_id").Unique(),
		edge.From("creator", User.Type).Ref("created_image_creation_templates").Field("created_by").Unique().Required(),
		edge.From("updater", User.Type).Ref("updated_image_creation_templates").Field("updated_by").Unique().Required(),
	}
}

func (ImageCreationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state", "published_at"),
		index.Fields("updated_at"),
		index.Fields("home_position").Unique().Annotations(entsql.IndexWhere("home_position IS NOT NULL")),
	}
}
