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

type PromotionTemplate struct {
	ent.Schema
}

func (PromotionTemplate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("promotion_templates"),
	}
}

func (PromotionTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("name").MaxLen(128).NotEmpty(),
		field.String("channel_id").MaxLen(36).NotEmpty(),
		field.Enum("trigger_type").Values(
			"media_published", "media_reviewed", "article_published",
			"daily_digest", "manual",
		).Default("media_published"),
		field.String("subject_tpl").MaxLen(512).Optional(),
		field.Text("body_tpl").NotEmpty(),
		field.JSON("subject_i18n", map[string]string{}).Optional(),
		field.JSON("body_i18n", map[string]string{}).Optional(),
		field.Bool("is_default").Default(false),
		field.Bool("is_active").Default(true),
		field.String("tenant_id").MaxLen(36).Default("default").
			Comment("Tenant isolation key"),
		field.Time("create_time").Default(time.Now).Immutable(),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PromotionTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id"),
		index.Fields("trigger_type"),
		index.Fields("is_default"),
		index.Fields("tenant_id"),
	}
}

func (PromotionTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", PromotionChannel.Type).
			Ref("templates").
			Field("channel_id").
			Required().
			Unique(),
		edge.To("tasks", PromotionTask.Type),
	}
}