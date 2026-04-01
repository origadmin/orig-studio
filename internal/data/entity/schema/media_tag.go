/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * MediaTag - Many-to-Many relationship between Media and Tag
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
)

type MediaTag struct {
	ent.Schema
}

func (MediaTag) Fields() []ent.Field {
	return []ent.Field{}
}

func (MediaTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_media_tags"),
		entsql.WithComments(true),
	}
}

func (MediaTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type),
		edge.To("tag", Tag.Type),
	}
}
