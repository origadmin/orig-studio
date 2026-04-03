package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EncodingTask holds the schema definition for the EncodingTask entity.
type EncodingTask struct {
	ent.Schema
}

// Fields of the EncodingTask.
func (EncodingTask) Fields() []ent.Field {
	return []ent.Field{
		field.Int("media_id"),
		field.Int("profile_id"),
		field.String("status").MaxLen(20).Default("pending"), // pending, processing, success, failed
		field.Int("progress").Default(0),                     // 0-100
		field.String("output_path").MaxLen(512).Optional(),
		field.Text("error_message").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the EncodingTask.
func (EncodingTask) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("tasks").Field("media_id").Required().Unique(),
		edge.From("profile", EncodeProfile.Type).Ref("tasks").Field("profile_id").Required().Unique(),
	}
}

// Indexes of the EncodingTask.
func (EncodingTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
