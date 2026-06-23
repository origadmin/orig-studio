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

type Refund struct {
	ent.Schema
}

func (Refund) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_refunds"),
		entsql.WithComments(true),
	}
}

func (Refund) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.Float("amount"),
		field.String("reason").Optional().MaxLen(500),
		field.Enum("status").Values("pending", "approved", "rejected", "completed").Default("pending"),
		field.String("order_id").NotEmpty(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Refund) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("status"),
		index.Fields("create_time"),
	}
}

func (Refund) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("refunds").
			Field("order_id").
			Unique().
			Required(),
	}
}
