/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * User model - corresponds to Django users.User model
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
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").NotEmpty().Unique().MaxLen(150),
		field.String("email").NotEmpty().Unique().MaxLen(254),
		field.String("password").MaxLen(256),
		field.String("name").MaxLen(250).SchemaType(map[string]string{"postgres": "VARCHAR(250)"}),
		field.String("first_name").Optional().MaxLen(150),
		field.String("last_name").Optional().MaxLen(150),
		field.Bool("is_active").Default(true),
		field.Bool("is_staff").Default(false),
		field.Enum("role").Values("user", "admin", "editor").Default("user"),
		field.Bool("is_superuser").Default(false),
		field.Bool("is_approved").Optional(),
		field.Bool("is_featured").Default(false),
		field.Bool("advanced_user").Default(false),
		field.Bool("is_editor").Default(false),
		field.Bool("is_manager").Default(false),
		field.String("title").Optional().MaxLen(250),
		field.Text("description").Optional(),
		field.String("logo").Optional().MaxLen(500),
		field.String("location").Optional().MaxLen(250),
		field.Int("media_count").Default(0),
		field.Bool("notification_on_comments").Default(true),
		field.Bool("allow_contact").Default(false),
		field.Time("date_joined").Default(time.Now),
		field.Time("date_added").Default(time.Now),
		field.Time("last_login").Optional(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username"),
		index.Fields("email"),
		index.Fields("is_active"),
		index.Fields("is_staff"),
		index.Fields("date_added"),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("users_user"),
		entsql.WithComments(true),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type),
		edge.To("channels", Channel.Type),
		edge.To("playlists", Playlist.Type),
		edge.To("comments", Comment.Type),
		edge.To("notifications", Notification.Type),
		edge.To("categories", Category.Type),
		edge.To("tags", Tag.Type),
		edge.To("favorites", Favorite.Type),
		edge.To("likes", Like.Type),
	}
}
