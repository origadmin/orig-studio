/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel model - corresponds to Django users.Channel model
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

type Channel struct {
	ent.Schema
}

func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7), // UUIDv7 for distributed system
		field.String("user_id"),
		field.String("title").NotEmpty().MaxLen(90),
		field.String("slug").MaxLen(100).Unique(),
		field.Text("description"),
		field.String("short_token").MaxLen(12).Unique().DefaultFunc(idutil.GenShortID),
		field.String("banner_logo").MaxLen(500),
		field.Bool("is_public").Default(true),
		field.Time("add_date").Default(time.Now),
	}
}

func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("title"),
		index.Fields("slug"),
		index.Fields("short_token"),
		index.Fields("add_date"),
	}
}

func (Channel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("users_channel"),
		entsql.WithComments(true),
	}
}

func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("channels").Field("user_id").Unique().Required(),
		edge.To("media", Media.Type),
	}
}
