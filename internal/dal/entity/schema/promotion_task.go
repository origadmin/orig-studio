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

type PromotionTask struct {
	ent.Schema
}

func (PromotionTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("promotion_tasks"),
	}
}

func (PromotionTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("channel_id").MaxLen(36).NotEmpty(),
		field.String("template_id").MaxLen(36).Optional(),
		field.Enum("task_type").Values(
			"single", "digest", "scheduled",
		).Default("single"),
		field.Enum("status").Values(
			"pending", "processing", "completed", "failed", "cancelled",
		).Default("pending"),
		field.JSON("media_ids", []string{}).Optional(),
		field.JSON("payload", map[string]interface{}{}).Optional(),
		field.String("rendered_subject").MaxLen(512).Optional(),
		field.Text("rendered_body").Optional(),
		field.Int("retry_count").Default(0),
		field.Int("max_retries").Default(3),
		field.String("error_message").Optional(),
		field.Time("scheduled_at").Optional().Nillable(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.String("tenant_id").MaxLen(36).Default("default").
			Comment("Tenant isolation key"),
		field.Time("create_time").Default(time.Now).Immutable(),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PromotionTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id"),
		index.Fields("status"),
		index.Fields("task_type"),
		index.Fields("scheduled_at"),
		index.Fields("tenant_id"),
	}
}

func (PromotionTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", PromotionChannel.Type).
			Ref("tasks").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("template", PromotionTemplate.Type).
			Ref("tasks").
			Field("template_id").
			Unique(),
		edge.To("logs", PromotionLog.Type),
	}
}