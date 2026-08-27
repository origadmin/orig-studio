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

type DrmPolicy struct {
	ent.Schema
}

func (DrmPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("drm_policies"),
		entsql.WithComments(true),
	}
}

func (DrmPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("name").NotEmpty().MaxLen(128),
		field.Enum("type").Values("hls_aes128", "widevine", "fairplay", "multi").Default("hls_aes128"),
		field.String("hls_key_url").Optional().MaxLen(512),
		field.String("widevine_pssh").Optional().MaxLen(2048),
		field.String("fairplay_cert_url").Optional().MaxLen(512),
		field.Bool("is_default").Default(false),
		field.String("description").Optional().MaxLen(500),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (DrmPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("is_default"),
		index.Fields("create_time"),
	}
}

func (DrmPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("keys", DrmKey.Type),
		edge.To("media_policies", MediaDrmPolicy.Type),
	}
}
