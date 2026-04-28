package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origcms/internal/helpers/idutil"
)

type GroupMember struct {
	ent.Schema
}

func (GroupMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("user_id"),
		field.String("group_id"),
		field.Time("joined_at").Default(time.Now),
	}
}

func (GroupMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "group_id").Unique(),
		index.Fields("user_id"),
		index.Fields("group_id"),
	}
}

func (GroupMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("auth_group_members"),
		entsql.WithComments(true),
	}
}

func (GroupMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("group_memberships").Field("user_id").Unique().Required(),
		edge.From("group", PermissionGroup.Type).Ref("members").Field("group_id").Unique().Required(),
	}
}
