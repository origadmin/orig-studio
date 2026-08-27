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

type WalletTransaction struct {
	ent.Schema
}

func (WalletTransaction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("payment_wallet_transactions"),
		entsql.WithComments(true),
	}
}

func (WalletTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.Enum("type").Values("deposit", "withdraw", "consume", "refund", "freeze", "unfreeze"),
		field.Float("amount"),
		field.Float("balance_before"),
		field.Float("balance_after"),
		field.String("reference").Optional().MaxLen(128),
		field.String("description").Optional().MaxLen(500),
		field.String("wallet_id").NotEmpty(),
		field.Time("create_time").Default(time.Now),
	}
}

func (WalletTransaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("wallet_id"),
		index.Fields("type"),
		index.Fields("reference"),
		index.Fields("create_time"),
		index.Fields("wallet_id", "type"),
	}
}

func (WalletTransaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("wallet", Wallet.Type).
			Ref("transactions").
			Field("wallet_id").
			Unique().
			Required(),
	}
}
