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

type LiveRoom struct {
	ent.Schema
}

func (LiveRoom) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("title").NotEmpty().MaxLen(256),
		field.Text("description").Optional().MaxLen(2000),
		field.String("stream_key").Unique().NotEmpty().MaxLen(64).DefaultFunc(idutil.GenUUIDv7).Sensitive(),
		field.String("rtmp_url").Optional().MaxLen(512),
		field.String("hls_url").Optional().MaxLen(512),
		field.Enum("status").Values("idle", "preparing", "live", "ended", "offline").Default("idle"),
		field.Time("scheduled_at").Optional(),
		field.Time("started_at").Optional(),
		field.Time("ended_at").Optional(),
		field.Int("max_viewers").Default(0),
		field.Int("current_viewers").Default(0),
		field.Int("peak_viewers").Default(0),
		field.String("thumbnail").Optional().MaxLen(512),
		field.String("category").Optional().MaxLen(128),
		field.JSON("tags", []string{}).Optional(),
		field.String("user_id"),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LiveRoom) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("user_id"),
		index.Fields("stream_key").Unique(),
		index.Fields("scheduled_at"),
		index.Fields("category"),
		index.Fields("create_time"),
		index.Fields("user_id", "status"),
	}
}

func (LiveRoom) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("live_rooms"),
	}
}

func (LiveRoom) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("live_rooms").Field("user_id").Required().Unique(),
		edge.To("streams", LiveStream.Type),
		edge.To("chat_messages", LiveChatMessage.Type),
		edge.To("schedules", LiveSchedule.Type),
	}
}
