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

type LiveStream struct {
	ent.Schema
}

func (LiveStream) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("stream_key").Unique().NotEmpty().MaxLen(64),
		field.String("rtmp_url").Optional().MaxLen(512),
		field.String("hls_url").Optional().MaxLen(512),
		field.Enum("status").Values("pending", "live", "ended").Default("pending"),
		field.Time("started_at").Optional(),
		field.Time("ended_at").Optional(),
		field.Int("peak_viewers").Default(0),
		field.Int64("duration").Default(0),
		field.String("room_id"),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LiveStream) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("stream_key").Unique(),
		index.Fields("status"),
		index.Fields("room_id"),
		index.Fields("started_at"),
		index.Fields("create_time"),
		index.Fields("room_id", "status"),
	}
}

func (LiveStream) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("live_streams"),
	}
}

func (LiveStream) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("room", LiveRoom.Type).Ref("streams").Field("room_id").Required().Unique(),
		edge.To("recordings", LiveRecording.Type),
	}
}
