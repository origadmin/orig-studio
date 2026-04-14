/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Subscription model - for user subscriptions/follows
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Subscription struct {
	ent.Schema
}

func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("subscriber_id"),
		field.String("channel_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (Subscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subscriber_id", "channel_id").
			Unique(),
		index.Fields("subscriber_id"),
		index.Fields("channel_id"),
		index.Fields("created_at"),
	}
}

func (Subscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("subscriptions_subscription"),
		entsql.WithComments(true),
	}
}

func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("subscriber", User.Type).
			Ref("subscriptions").
			Field("subscriber_id").
			Unique().
			Required(),
		edge.From("channel", User.Type).
			Ref("subscribers").
			Field("channel_id").
			Unique().
			Required(),
	}
}
