package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type AdPlacement struct {
	ent.Schema
}

func (AdPlacement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("ad_placements"),
	}
}

func (AdPlacement) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("name").MaxLen(128).NotEmpty(),
		field.String("slug").MaxLen(64).Unique().NotEmpty(),
		field.String("type").MaxLen(32).NotEmpty(),
		field.String("description").MaxLen(512).Optional(),
		field.Int("width").Default(0),
		field.Int("height").Default(0),
		field.Int("max_ads").Default(1),
		field.Bool("is_active").Default(true),
		field.Int("sequence").Default(0),
	}
}

func (AdPlacement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug"),
		index.Fields("type"),
		index.Fields("is_active"),
		index.Fields("sequence"),
	}
}
