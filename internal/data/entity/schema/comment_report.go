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

type CommentReport struct {
	ent.Schema
}

func (CommentReport) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("comment_id"),
		field.String("reporter_id"),
		field.Enum("reason").Values("SPAM", "HARASSMENT", "INAPPROPRIATE", "OTHER"),
		field.Text("description").Optional(),
		field.Enum("status").Values("PENDING", "REVIEWED", "DISMISSED").Default("PENDING"),
		field.Time("create_time").Default(time.Now),
	}
}

func (CommentReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("reporter_id", "comment_id").Unique(),
		index.Fields("comment_id"),
		index.Fields("reason"),
		index.Fields("create_time"),
		index.Fields("status"),
	}
}

func (CommentReport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_comment_reports"),
		entsql.WithComments(true),
	}
}

func (CommentReport) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("comment", Comment.Type).Ref("reports").Field("comment_id").Unique().Required(),
		edge.From("reporter", User.Type).Ref("comment_reports").Field("reporter_id").Unique().Required(),
	}
}
