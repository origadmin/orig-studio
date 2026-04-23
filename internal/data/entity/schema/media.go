/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Media entity - updated to use clean field names (M1 unification)
// Replaces the Django-specific field names with clean Go-style names.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type Media struct {
	ent.Schema
}

func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7), // UUIDv7 for distributed system
		field.String("title").MaxLen(255),
		field.Text("description").Optional(),
		field.String("short_token").MaxLen(150).DefaultFunc(idutil.GenShortID).Unique().NotEmpty(),
		field.String("type").MaxLen(20).Default("video"), // video / image / audio
		field.String("url").MaxLen(512),                  // original file path
		field.String("hls_file").MaxLen(1024).Optional(), // HLS master playlist path
		field.String("thumbnail").MaxLen(512).Optional(), // thumbnail path
		field.String("poster").MaxLen(512).Optional(),
		field.String("preview_file_path").MaxLen(512).Optional(), // GIF preview path
		field.Int("duration").Default(0),
		field.String("size").MaxLen(32).Optional(),
		field.Int("width").Default(0),
		field.Int("height").Default(0),
		field.String("mime_type").MaxLen(128).Optional(),
		field.String("md5sum").MaxLen(64).Optional(),
		field.String("extension").MaxLen(32).Optional(),
		field.Int("privacy").Default(1), // 1: public, 2: private, 3: unlisted
		// encoding_status: pending / processing / success / partial / failed
		field.String("encoding_status").MaxLen(20).Default("pending"),
		// state: draft / active / deleted
		field.String("state").MaxLen(20).Default("active"),
		field.Int64("view_count").Default(0),
		field.Int64("like_count").Default(0),
		field.Int64("dislike_count").Default(0),
		field.Int64("comment_count").Default(0),
		field.Int64("favorite_count").Default(0),
		field.Int64("download_count").Default(0),
		field.Bool("allow_download").Default(true),
		field.Bool("enable_comments").Default(true),
		field.Bool("featured").Default(false),
		field.String("review_status").MaxLen(20).Default("pending_review"),
		field.Bool("listable").Default(false),
		field.Int("reported_times").Default(0),
		field.JSON("tags", []string{}).Optional(),
		field.String("user_id"),
		field.String("category_id").Optional(),
		field.String("channel_id").Optional(),
		field.Time("published_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Media) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("title"),
		index.Fields("type"),
		index.Fields("state"),
		index.Fields("encoding_status"),
		index.Fields("featured"),
		index.Fields("view_count"),
		index.Fields("created_at"),
		index.Fields("user_id"),
		index.Fields("short_token").Unique(),
		index.Fields("review_status", "listable", "state"),
		index.Fields("listable"),
	}
}

func (Media) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("media").Field("user_id").Required().Unique(),
		edge.From("category", Category.Type).Ref("media").Field("category_id").Unique(),
		edge.To("comments", Comment.Type),
		edge.From("channel", Channel.Type).Ref("media").Field("channel_id").Unique(),
		edge.To("playlists", MediaPlaylist.Type),
		edge.To("tags_rel", MediaTag.Type),
		edge.To("favorites", Favorite.Type),
		edge.To("likes", Like.Type),
		// edge.To("tasks", EncodingTask.Type),
	}
}
