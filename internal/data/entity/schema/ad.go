package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type Ad struct {
	ent.Schema
}

func (Ad) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("ads"),
	}
}

func (Ad) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("placement_id").MaxLen(36).NotEmpty(),
		field.String("title").MaxLen(255).NotEmpty(),
		field.JSON("title_i18n", map[string]string{}).Optional(),
		field.String("image_url").MaxLen(512).Optional(),
		field.String("image_mobile_url").MaxLen(512).Optional(),
		field.String("link_url").MaxLen(512).Optional(),
		field.String("link_target").MaxLen(16).Default("_blank"),
		field.String("badge_text").MaxLen(64).Optional(),
		field.Int("priority").Default(0),
		field.Bool("is_active").Default(true),
		field.Time("start_at").Optional(),
		field.Time("end_at").Optional(),
		field.Int64("impressions").Default(0),
		field.Int64("clicks").Default(0),
	}
}

func (Ad) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("placement_id"),
		index.Fields("is_active"),
		index.Fields("priority"),
		index.Fields("start_at", "end_at"),
	}
}
