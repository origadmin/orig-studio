/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * User model - corresponds to Django users.User model
 */

package schema

import (
	"origadmin/application/origcms/internal/helpers/idutil"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("username").NotEmpty().Unique().MaxLen(150),
		field.String("email").NotEmpty().Unique().MaxLen(254),
		field.String("password").MaxLen(256),
		field.String("name").MaxLen(250).SchemaType(map[string]string{"postgres": "VARCHAR(250)"}),
		field.String("slug").MaxLen(64).Unique().Optional(),
		field.String("first_name").Optional().MaxLen(150),
		field.String("last_name").Optional().MaxLen(150),
		field.Enum("status").Values("PENDING", "ACTIVE", "INACTIVE", "SUSPENDED", "REJECTED").Default("ACTIVE"),
		field.Bool("is_staff").Default(false),
		field.Enum("role").Values("user", "admin", "editor").Default("user"),
		field.Bool("is_superuser").Default(false),
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
		field.String("nickname").MaxLen(150).Optional(),
		field.String("phone").MaxLen(32).Optional(),
		field.String("avatar").MaxLen(500).Optional(),
		field.String("last_login_ip").MaxLen(64).Optional(),
		field.String("login_ip").MaxLen(64).Optional(),
		field.Time("last_login_time").Optional(),
		field.Time("login_time").Optional(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
		// Audit author fields - store user ID who created/updated the record (UUIDv7 string matching users.id)
		field.String("create_author").Default("").Comment("Create author user ID"),
		field.String("update_author").Default("").Comment("Update author user ID"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username"),
		index.Fields("email"),
		index.Fields("slug"),
		index.Fields("status"),
		index.Fields("is_staff"),
		index.Fields("date_added"),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("users"),
		entsql.WithComments(true),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type),
		edge.To("articles", Article.Type),
		edge.To("channels", Channel.Type),
		edge.To("playlists", Playlist.Type),
		edge.To("comments", Comment.Type),
		edge.To("notifications", Notification.Type).StorageKey(edge.Table("user_notification_mappings")),
		edge.To("categories", Category.Type),
		edge.To("tags", Tag.Type),
		edge.To("favorites", Favorite.Type),
		edge.To("likes", Like.Type),
		edge.To("comment_likes", CommentLike.Type),
		edge.To("subscriptions", Subscription.Type),
		edge.To("subscribers", Subscription.Type),
		edge.To("review_logs", MediaReviewLog.Type),
		edge.To("comment_reports", CommentReport.Type),
		edge.To("moderated_comments", Comment.Type),
		edge.To("group_memberships", GroupMember.Type),
		edge.To("created_groups", PermissionGroup.Type),
	}
}
