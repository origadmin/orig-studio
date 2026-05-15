/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"io"

	"origadmin/application/origstudio/internal/data/enums"
)

// Storage defines the interface for media storage operations.
type Storage interface {
	// Direct upload/download
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)

	// Multipart upload
	StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error)
	MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error
	DeleteParts(ctx context.Context, uploadID string) error

	// File promotion and temp cleanup
	PromoteToOriginal(ctx context.Context, tempPath string) (string, error)
	CleanupTempParts(ctx context.Context, userID, uploadID string) error

	// Sync status tracking (for hybrid/S3 storage)
	SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error)
}

// contextKey is an unexported type for context keys defined in this package.
type contextKey int

const (
	userIDCtxKey contextKey = iota
)

// ContextWithUserID returns a context with the userID set for storage path generation.
// The dal layer reads this value to determine user-isolated storage paths.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

// UserIDFromContext extracts the userID from the context.
// Falls back to "_system" if not set.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userIDCtxKey).(string); ok && v != "" {
		return v
	}
	return "_system"
}
