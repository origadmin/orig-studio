/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"io"
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
}
