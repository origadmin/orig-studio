package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EncodeProfile holds the schema definition for the EncodeProfile entity.
type EncodeProfile struct {
	ent.Schema
}

// Fields of the EncodeProfile.
func (EncodeProfile) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").MaxLen(100).Unique(),
		field.String("description").Optional(),
		field.String("extension").MaxLen(20).Default("mp4"), // mp4, webm, etc.
		field.String("resolution").MaxLen(20),               // e.g., 720, 1080, 1280x720
		field.String("video_codec").MaxLen(50).Default("h264"),
		field.String("video_bitrate").MaxLen(20).Optional(),
		field.String("audio_codec").MaxLen(50).Default("aac"),
		field.String("audio_bitrate").MaxLen(20).Optional(),
		field.Text("bento_parameters").Optional(), // Bento4 mp4hls extra args
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the EncodeProfile.
func (EncodeProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tasks", EncodingTask.Type),
	}
}

// Indexes of the EncodeProfile.
func (EncodeProfile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("is_active"),
	}
}
