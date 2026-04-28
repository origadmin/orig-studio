/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Favorite model - corresponds to Django files.Favorite model
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Favorite struct {
	ent.Schema
}

func (Favorite) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("media_id"),
		field.String("user_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (Favorite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "media_id").Unique(),
	}
}

func (Favorite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_favorites"),
		entsql.WithComments(true),
	}
}

func (Favorite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).
			Ref("favorites").
			Field("media_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("favorites").
			Field("user_id").
			Unique().
			Required(),
	}
}
