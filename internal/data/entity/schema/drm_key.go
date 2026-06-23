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

type DrmKey struct {
	ent.Schema
}

func (DrmKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("drm_keys"),
		entsql.WithComments(true),
	}
}

func (DrmKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("content_id").NotEmpty().MaxLen(36),
		field.String("key_value").NotEmpty().Sensitive().MaxLen(512),
		field.String("key_id").Optional().MaxLen(256),
		field.String("iv").Optional().MaxLen(64),
		field.String("policy_id").NotEmpty().MaxLen(36),
		field.Time("created_at").Default(time.Now),
		field.Time("expires_at").Optional(),
	}
}

func (DrmKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("content_id"),
		index.Fields("key_id"),
		index.Fields("policy_id"),
		index.Fields("expires_at"),
		index.Fields("content_id", "policy_id"),
	}
}

func (DrmKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", DrmPolicy.Type).
			Ref("keys").
			Field("policy_id").
			Unique().
			Required(),
		edge.To("licenses", DrmLicense.Type),
	}
}
