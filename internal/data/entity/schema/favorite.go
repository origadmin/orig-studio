/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Favorite model - corresponds to Django files.Favorite model
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

type Favorite struct {
	ent.Schema
}

func (Favorite) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now),
	}
}

func (Favorite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_favorite"),
		entsql.WithComments(true),
	}
}

func (Favorite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).
			Ref("favorites").
			Unique(),
		edge.To("user", User.Type),
	}
}
