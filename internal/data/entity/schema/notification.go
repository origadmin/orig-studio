/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Notification model - corresponds to Django users.Notification model
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("action").MaxLen(30),
		field.Bool("notify").Default(false),
		field.String("method").MaxLen(20).Default("email"),
		field.Int("user_id"),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
	}
}

func (Notification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("users_notification"),
		entsql.WithComments(true),
	}
}

func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("notifications").Required(),
	}
}
