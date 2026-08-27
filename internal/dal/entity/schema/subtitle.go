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

// Subtitle entity — BUG-186: real subtitle storage replacing the stub endpoints.
//
// Decisions (G5 2026-08-19):
//   - uploaded srt/vtt are normalized to .vtt before storage (single format;
//     srt -> vtt is a pure text conversion, zero quality loss)
//   - status machine: processing (uploading/converting) -> active | failed
//     (failed carries error_message for the UI to surface "格式错误/解析失败")
//   - one row per language track; a media can hold many subtitle tracks
type Subtitle struct {
	ent.Schema
}

func (Subtitle) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("media_id").MaxLen(36),
		field.String("language").MaxLen(16).NotEmpty(), // ISO 639-1: zh / en / ja / ko
		field.String("label").MaxLen(64).Optional(),    // 展示名: 中文 / English
		field.String("file_url").MaxLen(512).Optional(), // SeaweedFS 可读 URL（统一 .vtt）
		field.String("status").MaxLen(20).Default("processing"), // active / processing / failed
		field.String("error_message").MaxLen(512).Optional(),    // failed 原因（格式错误/解析失败）
		field.Time("create_time").Default(time.Now),
		field.Time("update_time").Default(time.Now),
	}
}

func (Subtitle) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("subtitles"),
	}
}

func (Subtitle) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("media_id"),
		index.Fields("language"),
		index.Fields("create_time"),
	}
}

func (Subtitle) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("media", Media.Type).Ref("subtitles").Field("media_id").Unique().Required(),
	}
}
