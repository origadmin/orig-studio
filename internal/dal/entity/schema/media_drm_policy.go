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

type MediaDrmPolicy struct {
	ent.Schema
}

func (MediaDrmPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("media_drm_policies"),
		entsql.WithComments(true),
	}
}

func (MediaDrmPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("media_id").NotEmpty().MaxLen(36),
		field.String("policy_id").NotEmpty().MaxLen(36),
		field.Time("create_time").Default(time.Now),
	}
}

func (MediaDrmPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("policy_id"),
		index.Fields("media_id", "policy_id").Unique(),
	}
}

func (MediaDrmPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).
			Ref("drm_policies").
			Field("media_id").
			Unique().
			Required(),
		edge.From("policy", DrmPolicy.Type).
			Ref("media_policies").
			Field("policy_id").
			Unique().
			Required(),
	}
}
