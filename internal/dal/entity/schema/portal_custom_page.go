package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type PortalCustomPage struct {
	ent.Schema
}

func (PortalCustomPage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("portal_custom_pages"),
	}
}

func (PortalCustomPage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("title").MaxLen(255).NotEmpty(),
		field.String("slug").MaxLen(150).Unique().NotEmpty(),
		field.String("type").MaxLen(32).Default("custom"),
		field.String("content_format").MaxLen(32).Default("markdown"),
		field.Text("content").Optional(),
		field.String("layout").MaxLen(32).Default("full"),
		field.Bool("is_published").Default(false),
		field.Time("published_at").Optional(),
		field.String("seo_title").MaxLen(255).Optional(),
		field.String("seo_description").MaxLen(512).Optional(),
		field.String("featured_image").MaxLen(512).Optional(),
		field.Int64("view_count").Default(0),
	}
}

func (PortalCustomPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("is_published"),
	}
}
