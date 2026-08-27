package schema

import (
	"origadmin/application/origstudio/internal/pkg/idutil"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("name").NotEmpty().MaxLen(200),
		field.String("slug").NotEmpty().Unique().MaxLen(100),
		field.String("domain").Optional().MaxLen(255),
		field.String("logo").Optional().MaxLen(500),
		field.Text("description").Optional(),
		field.Enum("status").Values("active", "suspended", "pending", "deleted").Default("active"),
		field.Enum("plan").Values("free", "pro", "enterprise").Default("free"),
		field.Int("max_users").Default(10),
		field.Int("max_storage_mb").Default(1024),
		field.JSON("config", map[string]interface{}{}).Optional(),
		field.Time("expires_at").Optional(),
		field.Time("create_time").Default(time.Now).Immutable(),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("domain"),
		index.Fields("status"),
	}
}

func (Tenant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("system_tenants"),
		entsql.WithComments(true),
	}
}
