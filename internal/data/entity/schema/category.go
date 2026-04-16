/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Category entity - updated to use clean field names (M1 unification)

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("name").MaxLen(128).Unique(),
		field.String("slug").MaxLen(128).Unique().Optional(),
		field.Text("description").Optional(),
		field.String("thumbnail").MaxLen(512).Optional(),
		field.String("listings_thumbnail").MaxLen(512).Optional(),
		field.String("icon").MaxLen(255).Optional(),
		field.String("color").MaxLen(32).Optional(),
		field.String("parent_id").Optional(),
		field.Int("sequence").Default(0),
		field.Int("status").Default(1), // 1: active, 2: inactive
		field.Int("media_count").Default(0),
		field.Bool("is_global").Default(false),
		field.Bool("is_rbac_category").Default(false),
		field.String("identity_provider").Optional(),
		field.String("user_id").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("slug"),
		index.Fields("parent_id"),
		index.Fields("status"),
		index.Fields("is_global"),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("categories").Field("user_id").Unique(),
		edge.To("media", Media.Type),
		edge.To("articles", Article.Type),
		edge.To("children", Category.Type).From("parent").Unique().Field("parent_id"),
	}
}
