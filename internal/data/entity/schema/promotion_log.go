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

type PromotionLog struct {
	ent.Schema
}

func (PromotionLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("promotion_logs"),
	}
}

func (PromotionLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("task_id").MaxLen(36).NotEmpty(),
		field.String("channel_id").MaxLen(36).NotEmpty(),
		field.String("platform").MaxLen(32).NotEmpty(),
		field.Enum("status").Values(
			"success", "failed", "rate_limited", "skipped",
		).Default("success"),
		field.String("platform_message_id").MaxLen(256).Optional(),
		field.Text("request_payload").Optional(),
		field.Text("response_body").Optional(),
		field.Int("http_status").Default(0),
		field.String("error_message").Optional(),
		field.Int("duration_ms").Default(0),
		field.String("tenant_id").MaxLen(36).Default("default").
			Comment("Tenant isolation key"),
		field.Time("create_time").Default(time.Now).Immutable(),
	}
}

func (PromotionLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_id"),
		index.Fields("channel_id"),
		index.Fields("platform"),
		index.Fields("status"),
		index.Fields("create_time"),
		index.Fields("tenant_id"),
	}
}

func (PromotionLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", PromotionChannel.Type).
			Ref("logs").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("task", PromotionTask.Type).
			Ref("logs").
			Field("task_id").
			Required().
			Unique(),
	}
}