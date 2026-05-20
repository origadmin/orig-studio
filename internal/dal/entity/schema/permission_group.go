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

type PermissionGroup struct {
	ent.Schema
}

func (PermissionGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("name").NotEmpty().MaxLen(128).Unique(),
		field.Text("description").Optional(),
		field.JSON("permissions", []string{}),
		field.JSON("category_scope", []string{}).Optional(),
		field.Bool("is_active").Default(true),
		field.String("created_by").Optional(),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PermissionGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
		index.Fields("is_active"),
	}
}

func (PermissionGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("auth_permission_groups"),
		entsql.WithComments(true),
	}
}

func (PermissionGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", GroupMember.Type),
		edge.From("creator", User.Type).Ref("created_groups").Field("created_by").Unique(),
	}
}
