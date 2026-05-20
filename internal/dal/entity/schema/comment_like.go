package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type CommentLike struct {
	ent.Schema
}

func (CommentLike) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("comment_id"),
		field.String("user_id"),
		field.String("like_type").MaxLen(10).Default("like"),
		field.Time("create_time").Default(time.Now),
	}
}

func (CommentLike) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "comment_id").Unique(),
		index.Fields("comment_id"),
	}
}

func (CommentLike) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_comment_likes"),
		entsql.WithComments(true),
	}
}

func (CommentLike) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("comment", Comment.Type).
			Ref("comment_likes").
			Field("comment_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("comment_likes").
			Field("user_id").
			Unique().
			Required(),
	}
}
