package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/helpers/idutil"
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
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Indexes of the EncodingTask.
func (EncodingTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
