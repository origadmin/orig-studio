/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
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

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type LiveRecording struct {
	ent.Schema
}

func (LiveRecording) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("storage_url").NotEmpty().MaxLen(512),
		field.Int64("duration").Default(0),
		field.Int64("file_size").Default(0),
		field.String("format").Default("mp4").MaxLen(32),
		field.Enum("status").Values("recording", "processing", "completed", "failed").Default("recording"),
		field.String("stream_id"),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LiveRecording) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("stream_id"),
		index.Fields("status"),
		index.Fields("create_time"),
		index.Fields("stream_id", "status"),
	}
}

func (LiveRecording) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("live_recordings"),
	}
}

func (LiveRecording) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("stream", LiveStream.Type).Ref("recordings").Field("stream_id").Required().Unique(),
	}
}
