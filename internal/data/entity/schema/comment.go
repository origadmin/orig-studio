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

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Comment struct {
	ent.Schema
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.Text("text"),
		field.Time("add_date").Default(time.Now),
		field.String("media_id").StorageKey("media_comments"),
		field.String("user_id").StorageKey("user_comments"),
		field.Enum("status").Values("PENDING", "APPROVED", "REJECTED").Default("PENDING"),
		field.Int("report_count").Default(0),
		field.String("moderated_by").Optional(),
		field.Time("moderated_at").Optional(),
	}
}

func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("add_date"),
		index.Fields("media_id"),
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("media_id", "status"),
	}
}

func (Comment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_comments"),
		entsql.WithComments(true),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).
			Ref("comments").
			Field("media_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("comments").
			Field("user_id").
			Unique().
			Required(),
		edge.To("replies", Comment.Type).
			From("parent").
			Unique(),
		edge.To("comment_likes", CommentLike.Type),
		edge.To("reports", CommentReport.Type),
		edge.From("moderator", User.Type).
			Ref("moderated_comments").
			Field("moderated_by").
			Unique(),
	}
}
