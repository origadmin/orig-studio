package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/helpers/idutil"
)

type MediaReport struct {
	ent.Schema
}

func (MediaReport) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("media_id"),
		field.String("reporter_id"),
		field.Enum("reason").Values("SPAM", "HARASSMENT", "INAPPROPRIATE", "OTHER"),
		field.Text("description").Optional(),
		field.Enum("status").Values("PENDING", "REVIEWED", "DISMISSED").Default("PENDING"),
		field.Time("create_time").Default(time.Now),
	}
}

func (MediaReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("reporter_id", "media_id").Unique(),
		index.Fields("media_id"),
		index.Fields("reason"),
		index.Fields("create_time"),
		index.Fields("status"),
	}
}

func (MediaReport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_media_reports"),
		entsql.WithComments(true),
	}
}

func (MediaReport) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("reports").Field("media_id").Unique().Required(),
		edge.From("reporter", User.Type).Ref("media_reports").Field("reporter_id").Unique().Required(),
	}
}
