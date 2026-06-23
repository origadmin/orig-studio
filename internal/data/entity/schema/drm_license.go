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

type DrmLicense struct {
	ent.Schema
}

func (DrmLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("drm_licenses"),
		entsql.WithComments(true),
	}
}

func (DrmLicense) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.Enum("status").Values("active", "expired", "revoked").Default("active"),
		field.String("device_id").Optional().MaxLen(256),
		field.JSON("device_info", map[string]any{}).Optional(),
		field.Time("issued_at").Default(time.Now),
		field.Time("expires_at").Optional(),
		field.String("key_id").NotEmpty().MaxLen(36),
		field.String("user_id").NotEmpty().MaxLen(36),
	}
}

func (DrmLicense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("device_id"),
		index.Fields("key_id"),
		index.Fields("user_id"),
		index.Fields("expires_at"),
		index.Fields("user_id", "status"),
		index.Fields("user_id", "key_id"),
	}
}

func (DrmLicense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("key", DrmKey.Type).
			Ref("licenses").
			Field("key_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("drm_licenses").
			Field("user_id").
			Unique().
			Required(),
	}
}
