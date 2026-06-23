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

type Payment struct {
	ent.Schema
}

func (Payment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_payments"),
		entsql.WithComments(true),
	}
}

func (Payment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.Enum("channel").Values("stripe", "paypal", "alipay", "wallet"),
		field.String("transaction_id").Unique().Optional().MaxLen(256),
		field.Float("amount"),
		field.String("currency").Default("USD").MaxLen(3),
		field.Enum("status").Values("pending", "success", "failed", "refunded").Default("pending"),
		field.Time("paid_at").Optional(),
		field.JSON("metadata", map[string]any{}).Optional(),
		field.String("order_id").NotEmpty(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Payment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("transaction_id").Unique(),
		index.Fields("order_id"),
		index.Fields("channel"),
		index.Fields("status"),
		index.Fields("create_time"),
		index.Fields("channel", "status"),
	}
}

func (Payment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("payments").
			Field("order_id").
			Unique().
			Required(),
	}
}
