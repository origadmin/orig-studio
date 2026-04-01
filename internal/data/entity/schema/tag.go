/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Tag model - corresponds to Django files.Tag model
 */

package schema

import (
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
		field.Int("media_count").Default(0),
		field.String("listings_thumbnail").MaxLen(400),
	}
}

func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("media_count"),
	}
}

func (Tag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_tag"),
		entsql.WithComments(true),
	}
}

func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("tags"),
	}
}
