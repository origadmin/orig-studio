/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"
)

// Storage 存储接口
type Storage interface {
	// StorePart 存储分片
	StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error)
	// MergeParts 合并分片
	MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error
	// DeleteParts 删除分片
	DeleteParts(ctx context.Context, uploadID string) error
	// GetFile 获取文件
	GetFile(ctx context.Context, path string) ([]byte, error)
	// PutFile 存储文件
	PutFile(ctx context.Context, path string, data []byte) error
	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, path string) error
	// Exists 检查文件是否存在
	Exists(ctx context.Context, path string) (bool, error)
}
