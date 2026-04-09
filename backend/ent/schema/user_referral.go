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

// UserReferral holds the schema definition for the UserReferral entity.
//
// 邀请归因记录：记录用户之间的邀请关系及奖励发放情况。
//
// 删除策略：硬删除
type UserReferral struct {
	ent.Schema
}

func (UserReferral) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_referrals"},
	}
}

func (UserReferral) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("inviter_id").
			Comment("邀请人用户ID"),
		field.Int64("invitee_id").
			Comment("被邀请人用户ID"),
		field.String("code").
			MaxLen(16).
			NotEmpty().
			Comment("使用的推荐码快照"),
		field.Float("inviter_rewarded").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("邀请人实际获得奖励金额"),
		field.Float("invitee_rewarded").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("被邀请人额外获得奖励金额"),
		field.Float("inviter_reward_snapshot").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("注册时邀请人奖励金额快照"),
		field.Float("invitee_reward_snapshot").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("注册时被邀请人奖励金额快照"),
		field.Time("reward_granted_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("奖励发放时间，NULL 表示待激活"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (UserReferral) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("inviter", User.Type).
			Ref("referrals_as_inviter").
			Field("inviter_id").
			Unique().
			Required(),
		edge.From("invitee", User.Type).
			Ref("referrals_as_invitee").
			Field("invitee_id").
			Unique().
			Required(),
	}
}

func (UserReferral) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("inviter_id"),
		index.Fields("invitee_id").Unique(),
		index.Fields("code"),
	}
}
