package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

// AdCreative 广告创意库：一次定义、可被多个广告位复用（解耦创意与广告位）。
type AdCreative struct {
	ent.Schema
}

func (AdCreative) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("ad_creatives"),
	}
}

func (AdCreative) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("title").MaxLen(255).NotEmpty(),
		field.JSON("title_i18n", map[string]string{}).Optional(),
		field.String("image_url").MaxLen(512).Optional(),
		field.String("image_mobile_url").MaxLen(512).Optional(),
		field.String("link_url").MaxLen(512).Optional(),
		field.String("link_target").MaxLen(16).Default("_blank"),
		field.String("badge_text").MaxLen(64).Optional(),
		field.Bool("is_active").Default(true),
		field.Int("priority").Default(0),
		field.Int64("impressions").Default(0),
		field.Int64("clicks").Default(0),
	}
}

func (AdCreative) Edges() []ent.Edge {
	return []ent.Edge{
		// 多对多：一个创意可被多个广告位引用
		edge.From("placements", AdPlacement.Type).
			Ref("creatives"),
	}
}

func (AdCreative) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
		index.Fields("priority"),
	}
}
