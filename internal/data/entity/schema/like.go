/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Like model - corresponds to Django files.Like model
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

type Like struct {
	ent.Schema
}

func (Like) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now),
	}
}

func (Like) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_like"),
		entsql.WithComments(true),
	}
}

func (Like) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).
			Ref("likes").
			Unique(),
		edge.To("user", User.Type),
	}
}
