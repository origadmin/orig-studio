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
	"entgo.io/ent/schema/index"
	"time"
)

type Like struct {
	ent.Schema
}

func (Like) Fields() []ent.Field {
	return []ent.Field{
		field.Int("media_id"),
		field.Int("user_id"),
		field.String("like_type").MaxLen(10).Default("like"), // like or dislike
		field.Time("created_at").Default(time.Now),
	}
}

func (Like) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "media_id").Unique(),
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
			Field("media_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("likes").
			Field("user_id").
			Unique().
			Required(),
	}
}
