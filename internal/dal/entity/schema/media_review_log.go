package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type MediaReviewLog struct {
	ent.Schema
}

func (MediaReviewLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("media_id"),
		field.String("reviewer_id"),
		field.String("action").MaxLen(20),
		field.Text("comment").Optional(),
		field.String("previous_status").MaxLen(20),
		field.String("new_status").MaxLen(20),
		field.Time("create_time").Default(time.Now),
	}
}

func (MediaReviewLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_media_review_logs"),
	}
}

func (MediaReviewLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("reviewer_id"),
		index.Fields("create_time"),
	}
}

func (MediaReviewLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("review_logs").Field("media_id").Unique().Required(),
		edge.From("reviewer", User.Type).Ref("review_logs").Field("reviewer_id").Unique().Required(),
	}
}
