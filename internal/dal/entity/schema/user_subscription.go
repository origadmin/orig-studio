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

type UserSubscription struct {
	ent.Schema
}

func (UserSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_user_subscriptions"),
		entsql.WithComments(true),
	}
}

func (UserSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.Enum("status").Values("active", "expired", "cancelled", "past_due").Default("active"),
		field.Time("started_at"),
		field.Time("expires_at"),
		field.Bool("auto_renew").Default(true),
		field.Time("cancelled_at").Optional(),
		field.String("user_id").NotEmpty(),
		field.String("plan_id").NotEmpty(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (UserSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("plan_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
		index.Fields("user_id", "status"),
	}
}

func (UserSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_subscriptions").
			Field("user_id").
			Unique().
			Required(),
		edge.From("plan", SubscriptionPlan.Type).
			Ref("subscriptions").
			Field("plan_id").
			Unique().
			Required(),
		edge.To("orders", Order.Type),
	}
}
