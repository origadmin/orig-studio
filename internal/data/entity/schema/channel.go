/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Channel model - represents a content channel owned by a user.
 * Per A009: channels are on-demand, not auto-created on registration.
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

// ChannelLink represents an external link associated with a channel.
type ChannelLink struct {
	Type     string `json:"type"`     // "website", "social", "custom"
	Platform string `json:"platform"` // "twitter", "github", etc.
	URL      string `json:"url"`
	Title    string `json:"title"`
}

type Channel struct {
	ent.Schema
}

func (Channel) Fields() []ent.Field {
	return []ent.Field{
		// Identity
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("user_id"),
		field.String("name").NotEmpty().MaxLen(150),          // Display name (required)
		field.String("title").NotEmpty().MaxLen(90),           // SEO title (kept for backward compat)
		field.String("slug").MaxLen(150).Unique().Optional(),  // URL slug (auto-generated from name)
		field.String("handle").MaxLen(150).Unique().NotEmpty(), // @handle identifier (immutable after creation)
		field.String("short_token").MaxLen(12).Unique().DefaultFunc(idutil.GenShortID),

		// Description
		field.Text("description"),

		// Visual
		field.String("avatar").MaxLen(500).Optional(),      // Channel avatar image URL
		field.String("banner").MaxLen(500).Optional(),      // Banner image URL (replaces banner_logo)
		field.String("banner_logo").MaxLen(500).Optional(), // DEPRECATED: kept for migration

		// Classification
		field.Enum("status").Values("ACTIVE", "INACTIVE", "SUSPENDED", "PENDING_REVIEW").Default("ACTIVE"),
		field.Enum("privacy").Values("PUBLIC", "PRIVATE", "UNLISTED", "PAID", "SUBSCRIBERS_ONLY").Default("PUBLIC"),
		field.JSON("tags", []string{}).Optional(),          // Channel topic tags (0-10)
		field.Int64("category_id").Optional(),              // Primary category FK

		// Flags
		// NOTE: is_default REMOVED (v2) -- per A009, no default channel auto-creation
		field.Bool("is_verified").Default(false), // Verified badge

		// Denormalized Counts
		field.Int64("subscriber_count").Default(0),
		field.Int("media_count").Default(0),
		field.Int("article_count").Default(0),
		field.Int64("total_views").Default(0),

		// External Links
		field.JSON("links", []ChannelLink{}).Optional(),

		// Timestamps
		field.Time("add_date").Default(time.Now),                      // DEPRECATED: kept for backward compat
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("name"),
		index.Fields("slug"),
		index.Fields("handle"),
		index.Fields("short_token"),
		index.Fields("status"),
		index.Fields("category_id"),
		index.Fields("create_time"),
		index.Fields("add_date"),
		// Composite: user's channel count check
		index.Fields("user_id", "status"),
	}
}

func (Channel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("user_channels"),
		entsql.WithComments(true),
	}
}

func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("channels").Field("user_id").Unique().Required(),
		edge.To("media", Media.Type),
		edge.To("articles", Article.Type),                                  // NEW: Channel has articles
		edge.From("category", Category.Type).Ref("channels").Field("category_id").Unique(), // NEW: Channel belongs to category
	}
}
