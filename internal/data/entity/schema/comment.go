/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Comment model - corresponds to Django files.Comment model
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
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
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
		edge.From("media", Media.Type).Ref("comments").Required().Unique(),
		edge.From("user", User.Type).Ref("comments").Required().Unique(),
		edge.To("replies", Comment.Type).From("parent").Unique(),
	}
}
