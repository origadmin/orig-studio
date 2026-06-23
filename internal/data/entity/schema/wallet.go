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

type Wallet struct {
	ent.Schema
}

func (Wallet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_wallets"),
		entsql.WithComments(true),
	}
}

func (Wallet) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.Float("balance").Default(0),
		field.Float("frozen").Default(0),
		field.String("currency").Default("USD").MaxLen(3),
		field.String("user_id").Unique().NotEmpty(),
	}
}

func (Wallet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").Unique(),
	}
}

func (Wallet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_wallet").
			Field("user_id").
			Unique().
			Required(),
		edge.To("transactions", WalletTransaction.Type),
	}
}
