/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"
	"time"

	"origadmin/application/origstudio/internal/dal/enums"
)

/// UploadSession represents an upload session for multipart uploads.
//
// A类修复: 新增 TempPath/FinalPath 字段，在 InitiateMultipartUpload 时一次计算并持久化，
// CompleteMultipartUpload 使用 session 中记录的固定路径，解决跨日跨月路径不一致问题。
type UploadSession struct {
	UploadID     string             `json:"upload_id"`
	Filename     string             `json:"filename"`
	FileSize     int64              `json:"file_size"`
	ContentType  string             `json:"content_type"`
	TotalParts   int                `json:"total_parts"`
	ChunkSize    int                `json:"chunk_size"`
	UploadedSize int64              `json:"uploaded_size"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	CategoryID   *int64             `json:"category_id"`
	ChannelID    *string            `json:"channel_id"`
	Tags         []string           `json:"tags"`
	UserID       *string            `json:"user_id"`
	Status       enums.UploadStatus `json:"status"`
	Thumbnail    string             `json:"thumbnail"`
	Parts        map[int]string     `json:"parts"` // part_number -> etag
	Sha256       string             `json:"sha256"`
	StoragePath  string             `json:"storage_path"`
	TempDir      string             `json:"temp_dir"`
	// TempPath 记录分片合并后的临时文件相对路径key，Complete时不再重新计算
	TempPath     string             `json:"temp_path"`
	// FinalPath 记录最终存储的相对路径key（temp/→originals/提升后），Complete时直接使用
	FinalPath    string             `json:"final_path"`
	ExpiresAt    time.Time          `json:"expires_at"`
	CreateTime   time.Time          `json:"create_time"`
	UpdateTime   time.Time          `json:"update_time"`
}

// UploadRepo defines the storage operations for upload sessions.
type UploadRepo interface {
	CreateSession(ctx context.Context, session *UploadSession) error
	GetSession(ctx context.Context, uploadID string) (*UploadSession, error)
	UpdateSession(ctx context.Context, session *UploadSession) error
	DeleteSession(ctx context.Context, uploadID string) error
	ListSessions(
		ctx context.Context,
		userID string,
		status enums.UploadStatus,
		page, pageSize int,
	) ([]*UploadSession, int, error)
	// DeleteExpiredSessions finds and deletes sessions that have expired.
	// Returns the list of upload IDs deleted.
	DeleteExpiredSessions(ctx context.Context, now time.Time) ([]string, error)
}
