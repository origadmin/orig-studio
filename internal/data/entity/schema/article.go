/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Article entity for content management

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Article struct {
	ent.Schema
}

func (Article) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7), // UUIDv7 for distributed system
		field.String("title").MaxLen(255).NotEmpty(),
		field.Text("content").NotEmpty(),
		field.Text("summary").Optional(),
		field.String("slug").MaxLen(150).Unique().Optional(),
		field.String("state").MaxLen(20).Default("draft"), // draft / published / archived
		field.Int64("view_count").Default(0),
		field.Int64("comment_count").Default(0),
		field.Bool("featured").Default(false),
		field.JSON("tags", []string{}).Optional(),
		field.String("user_id"),
		field.Int64("category_id").Optional().StructTag(`json:"category_id,omitempty"`),
		field.Time("published_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Article) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("slug"),
		index.Fields("state"),
		index.Fields("featured"),
		index.Fields("view_count"),
		index.Fields("created_at"),
		index.Fields("published_at"),
		index.Fields("user_id"),
	}
}

func (Article) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_articles"),
	}
}

func (Article) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("articles").Field("user_id").Required().Unique(),
		edge.From("category", Category.Type).Ref("articles").Field("category_id").Unique(),
		edge.To("comments", Comment.Type),
	}
}
