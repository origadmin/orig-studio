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

type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_orders"),
		entsql.WithComments(true),
	}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("order_no").Unique().NotEmpty().MaxLen(64),
		field.Float("amount"),
		field.String("currency").Default("USD").MaxLen(3),
		field.Enum("status").Values("pending", "paid", "completed", "refunded", "expired", "cancelled").Default("pending"),
		field.String("payment_method").Optional().MaxLen(32),
		field.Time("paid_at").Optional(),
		field.Time("expired_at").Optional(),
		field.String("user_id").NotEmpty(),
		field.String("plan_id").Optional(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_no").Unique(),
		index.Fields("user_id"),
		index.Fields("plan_id"),
		index.Fields("status"),
		index.Fields("payment_method"),
		index.Fields("create_time"),
		index.Fields("user_id", "status"),
		index.Fields("expired_at"),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("plan", SubscriptionPlan.Type).
			Ref("plan_orders").
			Field("plan_id").
			Unique(),
		edge.To("payments", Payment.Type),
		edge.To("refunds", Refund.Type),
	}
}
