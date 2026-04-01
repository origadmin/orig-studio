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
)

type Channel struct {
	ent.Schema
}

func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().MaxLen(90),
		field.Text("description"),
		field.String("friendly_token").MaxLen(12).Unique(),
		field.String("banner_logo").MaxLen(500),
		field.Time("add_date").Default(time.Now),
	}
}

func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("friendly_token"),
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
		edge.From("user", User.Type).Ref("channels").Unique().Required(),
		edge.To("media", Media.Type),
	}
}
