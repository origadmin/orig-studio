/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Comment model - corresponds to Django files.Comment model
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

	"github.com/google/uuid"
)

type Comment struct {
	ent.Schema
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Text("text"),
		field.UUID("uid", uuid.New()).Unique(),
		field.Time("add_date").Default(time.Now),
		field.Int("media_id"),
		field.Int("user_id"),
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("user_id"),
		index.Fields("add_date"),
	}
}

func (Comment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_comment"),
		entsql.WithComments(true),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("comments").Required(),
		edge.From("user", User.Type).Ref("comments").Required(),
		edge.To("replies", Comment.Type),
	}
}
