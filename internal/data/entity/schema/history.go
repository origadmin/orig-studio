/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * History model - watch history for multi-content-type support
 */

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type History struct {
	ent.Schema
}

func (History) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("user_id").Comment("Owner user ID"),
		field.String("content_id").Comment("Content ID (UUID of media/article)"),
		field.Enum("content_type").
			Values("video", "article", "audio").
			Default("video").
			Comment("Content type for polymorphic content_id"),
		field.Int("progress_seconds").Default(0).Comment("Watched/read seconds"),
		field.Int("duration_seconds").Default(0).Comment("Total duration in seconds"),
		field.Bool("is_finished").Default(false).Comment("Whether content is fully consumed (progress >= duration * threshold)"),
		field.String("title").Default("").Comment("Denormalized content title at watch time"),
		field.String("thumbnail").Default("").Comment("Denormalized content thumbnail URL at watch time"),
		field.String("short_token").Default("").Comment("Denormalized content short_token for frontend links"),
		field.Time("last_watched_at").Default(time.Now).Comment("Last watch timestamp"),
		field.Time("create_time").Default(time.Now).Comment("Record creation time (first watch)"),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now).Comment("Record update time"),
	}
}

func (History) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint: one history record per user per content
		index.Fields("user_id", "content_id", "content_type").Unique(),
		// Query optimization: list history sorted by last watched
		index.Fields("user_id", "last_watched_at"),
		// Query optimization: filter by content type
		index.Fields("user_id", "content_type", "last_watched_at"),
	}
}

func (History) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("user_history"),
		entsql.WithComments(true),
	}
}

func (History) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("history").
			Field("user_id").
			Unique().
			Required(),
		// NOTE: No edge to Media/Article.
		// Reason: content_id is polymorphic (points to different tables based on content_type).
		// ent does not support polymorphic foreign keys.
		// Denormalized fields (title, thumbnail, short_token) are stored directly
		// to avoid JOINs and preserve data when content is deleted.
	}
}
