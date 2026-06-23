package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type PromotionSubscription struct {
	ent.Schema
}

func (PromotionSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("promotion_subscriptions"),
	}
}

func (PromotionSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("user_id").MaxLen(36).NotEmpty(),
		field.JSON("channel_ids", []string{}).Optional(),
		field.JSON("tag_ids", []string{}).Optional(),
		field.JSON("platforms", []string{}).Optional(),
		field.Enum("frequency").Values(
			"realtime", "daily", "weekly",
		).Default("daily"),
		field.JSON("quiet_hours", map[string]string{}).Optional(),
		field.Int("daily_limit").Default(20),
		field.Bool("is_active").Default(true),
		field.String("tenant_id").MaxLen(36).Default("default").
			Comment("Tenant isolation key"),
		field.Time("create_time").Default(time.Now).Immutable(),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PromotionSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
		index.Fields("is_active"),
		index.Fields("tenant_id"),
	}
}

func (PromotionSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("promotion_subscriptions").
			Field("user_id").
			Required().
			Unique(),
	}
}