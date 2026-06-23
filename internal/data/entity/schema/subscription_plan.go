/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type SubscriptionPlan struct {
	ent.Schema
}

func (SubscriptionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_subscription_plans"),
		entsql.WithComments(true),
	}
}

func (SubscriptionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("name").NotEmpty().MaxLen(150),
		field.String("description").Optional().MaxLen(1000),
		field.Float("price"),
		field.String("currency").Default("USD").MaxLen(3),
		field.Int("duration_days"),
		field.JSON("features", map[string]any{}).Optional(),
		field.Bool("is_active").Default(true),
		field.Int("sort_order").Default(0),
	}
}

func (SubscriptionPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
		index.Fields("sort_order"),
		index.Fields("duration_days"),
	}
}

func (SubscriptionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("subscriptions", UserSubscription.Type),
		edge.To("plan_orders", Order.Type),
	}
}
