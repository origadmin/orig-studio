package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type AdClickLog struct {
	ent.Schema
}

func (AdClickLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("ad_click_logs"),
	}
}

func (AdClickLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("ad_id").MaxLen(36).NotEmpty(),
		field.String("placement_id").MaxLen(36).NotEmpty(),
		field.String("ip").MaxLen(64).Optional(),
		field.String("user_agent").MaxLen(512).Optional(),
		field.String("user_id").MaxLen(36).Optional(),
		field.String("referer").MaxLen(512).Optional(),
	}
}

func (AdClickLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ad_id"),
		index.Fields("placement_id"),
		index.Fields("user_id"),
	}
}
