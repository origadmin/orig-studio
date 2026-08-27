/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"io"
	"time"

	"origadmin/application/origstudio/internal/dal/enums"
)

// Storage defines the interface for media storage operations.
type Storage interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	DownloadToFile(ctx context.Context, key string, localPath string) error
	Delete(ctx context.Context, key string) error
	GetURL(ctx context.Context, key string) (string, error)

	StorePart(ctx context.Context, uploadID string, partNumber int, r io.Reader, size int64) (string, error)
	MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error
	DeleteParts(ctx context.Context, uploadID string) error

	PromoteToOriginal(ctx context.Context, tempPath string) (string, error)
	CleanupTempParts(ctx context.Context, userID, uploadID string) error

	UploadDir(ctx context.Context, localDir string, keyPrefix string) error
	DeletePrefix(ctx context.Context, keyPrefix string) error

	SyncStatus(ctx context.Context, key string) (enums.SyncStatus, error)
}

// contextKey is an unexported type for context keys defined in this package.
type contextKey int

const (
	userIDCtxKey contextKey = iota
	sessionCreateTimeCtxKey
)

// ContextWithUserID returns a context with the userID set for storage path generation.
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

// ContextWithSessionCreateTime stores the upload session creation time in context.
// Used by storage layer to generate time-safe paths without calling time.Now().
func ContextWithSessionCreateTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, sessionCreateTimeCtxKey, t)
}

// SessionCreateTimeFromContext extracts the session creation time from context.
// Returns (zero time, false) if not set; callers should fall back to time.Now().
func SessionCreateTimeFromContext(ctx context.Context) (time.Time, bool) {
	v, ok := ctx.Value(sessionCreateTimeCtxKey).(time.Time)
	return v, ok
}
