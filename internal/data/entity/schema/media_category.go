/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * MediaCategory - Many-to-Many relationship between Media and Category
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
)

type MediaCategory struct {
	ent.Schema
}

func (MediaCategory) Fields() []ent.Field {
	return []ent.Field{}
}

func (MediaCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_media_category"),
		entsql.WithComments(true),
	}
}

func (MediaCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type),
		edge.To("category", Category.Type),
	}
}
