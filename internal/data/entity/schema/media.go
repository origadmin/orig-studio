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
)

type Media struct {
	ent.Schema
}

func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").MaxLen(255),
		field.Text("description").Optional(),
		field.String("friendly_token").Unique().MaxLen(150).Optional(),
		field.String("type").MaxLen(20).Default("video"), // video / image / audio
		field.String("url").MaxLen(512),                  // original file path
		field.String("hls_file").MaxLen(1024).Optional(), // HLS master playlist path
		field.String("thumbnail").MaxLen(512).Optional(),
		field.String("poster").MaxLen(512).Optional(),
		field.String("preview_file_path").MaxLen(512).Optional(),
		field.Int("duration").Default(0),
		field.String("size").MaxLen(32).Optional(),
		field.Int("width").Default(0),
		field.Int("height").Default(0),
		field.String("mime_type").MaxLen(128).Optional(),
		field.String("md5sum").MaxLen(64).Optional(),
		field.String("extension").MaxLen(32).Optional(),
		field.Int("privacy").Default(1), // 1: public, 2: private, 3: unlisted
		// encoding_status: pending / processing / success / failed
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
		field.Bool("is_reviewed").Default(true),
		field.Int("reported_times").Default(0),
		field.JSON("tags", []string{}).Optional(),
		field.Int("user_id"),
		// category_id 由 edge 自动管理，不需显式定义
		field.Time("published_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Media) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("friendly_token"),
		index.Fields("title"),
		index.Fields("type"),
		index.Fields("state"),
		index.Fields("encoding_status"),
		index.Fields("featured"),
		index.Fields("view_count"),
		index.Fields("created_at"),
		index.Fields("user_id"),
		// category_id 由 edge 自动管理，索引通过 edge 生成
	}
}

func (Media) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("media").Required(),
		edge.From("category", Category.Type).Ref("media").Unique(),
		edge.To("comments", Comment.Type),
		edge.From("channels", Channel.Type).Ref("media"),
		edge.To("playlists", MediaPlaylist.Type),
		edge.To("tags_rel", MediaTag.Type),
		edge.To("favorites", Favorite.Type),
		edge.To("likes", Like.Type),
	}
}
