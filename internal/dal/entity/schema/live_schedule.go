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

type LiveSchedule struct {
	ent.Schema
}

func (LiveSchedule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("title").NotEmpty().MaxLen(256),
		field.Text("description").Optional().MaxLen(1000),
		field.Time("scheduled_at"),
		field.Int("duration").Default(3600),
		field.String("room_id"),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (LiveSchedule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id"),
		index.Fields("scheduled_at"),
		index.Fields("room_id", "scheduled_at"),
		index.Fields("create_time"),
	}
}

func (LiveSchedule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("live_schedules"),
	}
}

func (LiveSchedule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("room", LiveRoom.Type).Ref("schedules").Field("room_id").Required().Unique(),
	}
}
