package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type PortalBanner struct {
	ent.Schema
}

func (PortalBanner) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("portal_banners"),
	}
}

func (PortalBanner) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("title").MaxLen(255).NotEmpty(),
		field.JSON("title_i18n", map[string]string{}).Optional(),
		field.String("subtitle").MaxLen(255).Optional(),
		field.JSON("subtitle_i18n", map[string]string{}).Optional(),
		field.String("badge_text").MaxLen(64).Optional(),
		field.String("image_url").MaxLen(512).Optional(),
		field.String("image_mobile_url").MaxLen(512).Optional(),
		field.String("bg_color_start").MaxLen(32).Optional(),
		field.String("bg_color_end").MaxLen(32).Optional(),
		field.Float("bg_overlay_opacity").Default(0),
		field.String("primary_btn_text").MaxLen(64).Optional(),
		field.String("primary_btn_url").MaxLen(512).Optional(),
		field.String("secondary_btn_text").MaxLen(64).Optional(),
		field.String("secondary_btn_url").MaxLen(512).Optional(),
		field.Int("sequence").Default(0),
		field.Bool("is_active").Default(true),
		field.Time("start_at").Optional(),
		field.Time("end_at").Optional(),
		field.Int("auto_slide_interval").Default(5000),
	}
}

func (PortalBanner) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sequence"),
		index.Fields("is_active"),
	}
}
