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

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Playlist struct {
	ent.Schema
}

func (Playlist) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("title").NotEmpty().MaxLen(100),
		field.Text("description"),
		field.String("short_token").MaxLen(12).Unique().DefaultFunc(idutil.DefaultShortID()),
		field.String("user_id"),
		field.Int("privacy").Default(1), // 1=public, 2=unlisted, 3=private
		field.Time("add_date").Default(time.Now),
	}
}

func (Playlist) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("short_token"),
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
