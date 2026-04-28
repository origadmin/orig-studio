/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * MediaPlaylist - Many-to-Many relationship between Media and Playlist
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type MediaPlaylist struct {
	ent.Schema
}

func (MediaPlaylist) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("playlist_id"),
		field.String("media_id"),
		field.Int("ordering").Default(1),
		field.Time("action_date").Default(time.Now),
	}
}

func (MediaPlaylist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_playlist_media"),
		entsql.WithComments(true),
	}
}

func (MediaPlaylist) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type).Field("media_id").Unique().Required(),
		edge.To("playlist", Playlist.Type).Field("playlist_id").Unique().Required(),
	}
}
