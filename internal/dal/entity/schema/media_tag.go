/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * MediaTag - Many-to-Many relationship between Media and Tag
 */

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MediaTag struct {
	ent.Schema
}

// Fields declares the join columns explicitly (edge-schema pattern, mirroring
// GroupMember). Declaring them as real fields is what makes the columns
// NOT NULL and gives them readable names (media_id / tag_id) instead of the
// ent-derived defaults (media_tags_rel / tag_media_tags).
func (MediaTag) Fields() []ent.Field {
	return []ent.Field{
		// Matches content_media.id -> varchar(36) UUIDv7.
		field.String("media_id").MaxLen(36),
		// Matches content_tags.id -> auto-increment int.
		field.Int("tag_id"),
	}
}

func (MediaTag) Indexes() []ent.Index {
	return []ent.Index{
		// A media may carry a given tag at most once.
		index.Fields("media_id", "tag_id").Unique(),
		// media -> tags lookup (media detail / list projection).
		index.Fields("media_id"),
		// tag -> medias lookup (GET /api/v1/tags/{slug}/medias).
		index.Fields("tag_id"),
	}
}

func (MediaTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_media_tags"),
		entsql.WithComments(true),
	}
}

func (MediaTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("tags_rel").Field("media_id").Unique().Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("tag", Tag.Type).Ref("media_tags").Field("tag_id").Unique().Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
