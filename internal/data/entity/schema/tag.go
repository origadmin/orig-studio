/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Tag model - corresponds to Django files.Tag model
 */

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

type Tag struct {
	ent.Schema
}

func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().Unique().MaxLen(100),
		field.String("slug").MaxLen(100).Unique().Optional(),
		field.Int("media_count").Default(0),
		field.String("listings_thumbnail").MaxLen(400).Optional().Default(""),
		field.Enum("status").Values("ACTIVE", "INACTIVE").Default("ACTIVE"),
		field.String("description").Optional().MaxLen(500),
		field.String("color").MaxLen(32).Optional(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("slug"),
		index.Fields("media_count"),
	}
}

func (Tag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_tags"),
		entsql.WithComments(true),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("tags"),
	}
}
