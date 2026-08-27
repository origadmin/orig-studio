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

type LiveChatMessage struct {
	ent.Schema
}

func (LiveChatMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("content").NotEmpty().MaxLen(500),
		field.Enum("type").Values("text", "system", "gift").Default("text"),
		field.Time("sent_at").Default(time.Now),
		field.String("room_id"),
		field.String("user_id").Optional(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LiveChatMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id"),
		index.Fields("user_id"),
		index.Fields("type"),
		index.Fields("sent_at"),
		index.Fields("room_id", "sent_at"),
	}
}

func (LiveChatMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("live_chat_messages"),
	}
}

func (LiveChatMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("room", LiveRoom.Type).Ref("chat_messages").Field("room_id").Required().Unique(),
		edge.From("user", User.Type).Ref("live_chat_messages").Field("user_id").Unique(),
	}
}
