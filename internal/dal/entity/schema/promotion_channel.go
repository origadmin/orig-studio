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

type PromotionChannel struct {
	ent.Schema
}

func (PromotionChannel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("promotion_channels"),
	}
}

func (PromotionChannel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("name").MaxLen(128).NotEmpty(),
		field.String("slug").MaxLen(64).Unique().NotEmpty(),
		field.Enum("platform").Values(
			"telegram", "discord", "slack", "twitter",
			"youtube", "mastodon", "wechat", "weibo",
			"email", "webhook", "rss",
		).Default("telegram"),
		field.Enum("push_mode").Values(
			"realtime", "daily_digest", "manual",
		).Default("realtime"),
		field.Text("config").Optional().
			Comment("AES-GCM encrypted platform config, base64 encoded"),
		field.Bool("is_active").Default(false),
		field.Int("daily_limit").Default(0),
		field.Int("sent_today").Default(0),
		field.Time("limit_reset_at").Optional().Nillable(),
		field.JSON("rate_limit", map[string]interface{}{}).Optional(),
		field.Int("sequence").Default(0),
		field.String("tenant_id").MaxLen(36).Default("default").
			Comment("Tenant isolation key"),
		field.Time("create_time").Default(time.Now).Immutable(),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PromotionChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug"),
		index.Fields("platform"),
		index.Fields("is_active"),
		index.Fields("tenant_id"),
	}
}

func (PromotionChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("templates", PromotionTemplate.Type),
		edge.To("tasks", PromotionTask.Type),
		edge.To("logs", PromotionLog.Type),
	}
}