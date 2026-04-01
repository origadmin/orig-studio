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
)

type MediaPlaylist struct {
	ent.Schema
}

func (MediaPlaylist) Fields() []ent.Field {
	return []ent.Field{
		field.Int("ordering").Default(1),
		field.Time("action_date").Default(time.Now),
	}
}

func (MediaPlaylist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("files_playlistmedia"),
		entsql.WithComments(true),
	}
}

func (MediaPlaylist) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type),
		edge.To("playlist", Playlist.Type),
	}
}
