package schema

import (
	"origadmin/application/origcms/internal/helpers/idutil"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Setting struct {
	ent.Schema
}

func (Setting) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()),
		field.String("key").NotEmpty().Unique().MaxLen(200),
		field.Text("value").Default(""),
		field.Enum("type").Values("string", "int", "bool", "json").Default("string"),
		field.Enum("category").Values("general", "upload", "review", "email").Default("general"),
		field.Text("description").Optional(),
		field.Bool("is_sensitive").Default(false),
		field.Text("fallback_value").Optional(),
		field.Bool("is_builtin").Default(true),
		field.Time("create_time").Default(time.Now).Immutable(),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Setting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key").Unique(),
		index.Fields("category"),
	}
}

func (Setting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("system_settings"),
		entsql.WithComments(true),
	}
}
