/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// UploadSession entity - tracks multipart upload sessions for large file uploads.
// Supports chunked upload, resume from interruption, and progress tracking.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UploadSession holds the schema definition for the UploadSession entity.
type UploadSession struct {
	ent.Schema
}

// Fields of the UploadSession.
func (UploadSession) Fields() []ent.Field {
	return []ent.Field{
		// Unique upload session identifier (UUID)
		field.String("upload_id").Unique().NotEmpty().MaxLen(64).
			Comment("Unique upload session identifier"),
		// Original filename
		field.String("filename").NotEmpty().MaxLen(255).
			Comment("Original filename"),
		// Total file size in bytes
		field.Int64("file_size").Positive().
			Comment("Total file size in bytes"),
		// MIME type
		field.String("content_type").Optional().MaxLen(128).
			Comment("MIME type of the file"),
		// Total number of parts expected
		field.Int("total_parts").Positive().
			Comment("Total number of parts expected"),
		// Chunk size in bytes (default 2MB)
		field.Int("chunk_size").Default(2 * 1024 * 1024).
			Comment("Chunk size in bytes"),
		// Bytes uploaded so far
		field.Int64("uploaded_size").Default(0).
			Comment("Bytes uploaded so far"),
		// Media title (optional)
		field.String("title").Optional().MaxLen(255).
			Comment("Media title"),
		// Media description (optional)
		field.Text("description").Optional().
			Comment("Media description"),
		// Category ID (optional)
		field.Int64("category_id").Optional().Nillable().
			Comment("Category ID"),
		// Tags (optional)
		field.JSON("tags", []string{}).Optional().
			Comment("Tags for the media"),
		// User ID who initiated the upload
		field.Int64("user_id").Optional().Nillable().
			Comment("User ID who initiated the upload"),
		// Session status: pending, uploading, completed, aborted
		field.String("status").MaxLen(20).Default("pending").
			Comment("Session status: pending, uploading, completed, aborted"),
		// Map of part_number -> etag for uploaded parts
		field.JSON("parts", map[int]string{}).Optional().
			Comment("Map of part_number -> etag for uploaded parts"),
		// SHA-256 hash of the file (for integrity verification)
		field.String("sha256").Optional().MaxLen(64).
			Comment("SHA-256 hash of the file"),
		// Storage path for the final file
		field.String("storage_path").Optional().MaxLen(512).
			Comment("Storage path for the final file"),
		// Temporary directory for storing parts
		field.String("temp_dir").Optional().MaxLen(512).
			Comment("Temporary directory for storing parts"),
		// Session expiration time (default 24 hours)
		field.Time("expires_at").Default(func() time.Time {
			return time.Now().Add(24 * time.Hour)
		}).
			Comment("Session expiration time"),
		// Creation time
		field.Time("created_at").Default(time.Now).
			Comment("Creation time"),
		// Last update time
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).
			Comment("Last update time"),
	}
}

// Indexes of the UploadSession.
func (UploadSession) Indexes() []ent.Index {
	return []ent.Index{
		// Unique index on upload_id
		index.Fields("upload_id").Unique(),
		// Index on user_id for querying user's uploads
		index.Fields("user_id"),
		// Index on status for filtering
		index.Fields("status"),
		// Index on expires_at for cleanup jobs
		index.Fields("expires_at"),
	}
}
