package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TagName struct {
	ent.Schema
}

func (TagName) Fields() []ent.Field {
	return []ent.Field{
		field.Int("tag_id"),
		field.String("language").MaxLen(10),
		field.String("text").MaxLen(255),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (TagName) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tag_id", "language").Unique(),
		index.Fields("language", "text"),
	}
}

func (TagName) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_tag_names"),
	}
}

func (TagName) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tag", Tag.Type).Ref("names").Field("tag_id").Unique().Required(),
	}
}
