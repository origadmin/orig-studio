/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Playlist model - corresponds to Django files.Playlist model
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

	"github.com/google/uuid"
)

type Playlist struct {
	ent.Schema
}

func (Playlist) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().MaxLen(100),
		field.Text("description"),
		field.String("friendly_token").MaxLen(12).Unique(),
		field.UUID("uid", uuid.New()).Unique(),
		field.Int("user_id"),
		field.Time("add_date").Default(time.Now),
	}
}

func (Playlist) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("friendly_token"),
		index.Fields("user_id"),
		index.Fields("add_date"),
	}
}

func (Playlist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_playlist"),
		entsql.WithComments(true),
	}
}

func (Playlist) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("playlists").Required(),
	}
}
