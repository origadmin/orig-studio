package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/dal/enums"
	"origadmin/application/origstudio/internal/pkg/idutil"
)

// EncodingTask holds the schema definition for the EncodingTask entity.
type EncodingTask struct {
	ent.Schema
}

// Fields of the EncodingTask.
func (EncodingTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.DefaultUUIDv7()), // UUIDv7 for distributed system
		field.String("media_id").MaxLen(36),    // UUID for distributed system
		field.Int("profile_id"),
		field.Enum("status").
			GoType(enums.EncodingTaskStatusUnknown).
			Default(string(enums.EncodingTaskStatusPending)),
		field.String("output_path").MaxLen(512).Optional(),
		field.Text("error_message").Optional(),
		field.Bool("chunk").Default(false),
		field.Int("progress").Default(0),
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Annotations of the EncodingTask.
func (EncodingTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("system_encoding_tasks"),
	}
}

// Indexes of the EncodingTask.
func (EncodingTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("status"),
		index.Fields("create_time"),
	}
}
