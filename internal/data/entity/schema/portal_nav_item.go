package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/helpers/idutil"
)

type PortalNavItem struct {
	ent.Schema
}

func (PortalNavItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("portal_nav_items"),
	}
}

func (PortalNavItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("type").MaxLen(32).NotEmpty(),
		field.String("label").MaxLen(128).NotEmpty(),
		field.JSON("label_i18n", map[string]string{}).Optional(),
		field.String("url").MaxLen(512).Optional(),
		field.String("target_type").MaxLen(32).Optional(),
		field.String("target_id").MaxLen(36).Optional(),
		field.String("icon").MaxLen(64).Optional(),
		field.String("color").MaxLen(32).Optional(),
		field.Int("sequence").Default(0),
		field.String("parent_id").MaxLen(36).Optional(),
		field.Bool("is_visible").Default(true),
		field.Bool("open_new_tab").Default(false),
		field.String("css_class").MaxLen(128).Optional(),
	}
}

func (PortalNavItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sequence"),
		index.Fields("is_visible"),
	}
}
