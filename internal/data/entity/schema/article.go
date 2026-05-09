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
		field.String("short_token").MaxLen(150).DefaultFunc(idutil.GenShortID).Unique().NotEmpty(),
		field.String("state").MaxLen(20).Default("draft"), // draft / published / archived
		field.Int64("view_count").Default(0),
		field.Int64("comment_count").Default(0),
		field.Bool("featured").Default(false),
		field.JSON("tags", []string{}).Optional(),
		field.JSON("title_i18n", map[string]string{}).Optional(),
		field.JSON("content_i18n", map[string]string{}).Optional(),
		field.JSON("summary_i18n", map[string]string{}).Optional(),
		field.String("user_id"),
		field.Int64("category_id").Optional().StructTag(`json:"category_id,omitempty"`),
		field.String("media_id").MaxLen(36).Optional().StructTag(`json:"media_id,omitempty"`),
		field.String("thumbnail").MaxLen(512).Optional(),
		field.Time("published_at").Optional(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Article) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("slug"),
		index.Fields("short_token").Unique(),
		index.Fields("state"),
		index.Fields("featured"),
		index.Fields("view_count"),
		index.Fields("create_time"),
		index.Fields("published_at"),
		index.Fields("user_id"),
		index.Fields("media_id"),
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
		edge.From("media", Media.Type).Ref("articles").Field("media_id").Unique(),
		edge.To("comments", Comment.Type),
	}
}
