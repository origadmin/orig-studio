package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
)

type ChannelTag struct {
	ent.Schema
}

func (ChannelTag) Fields() []ent.Field {
	return []ent.Field{}
}

func (ChannelTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("content_channel_tags"),
		entsql.WithComments(true),
	}
}

func (ChannelTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).Ref("tags_rel").Unique(),
		edge.From("tag", Tag.Type).Ref("channel_tags").Unique(),
	}
}
