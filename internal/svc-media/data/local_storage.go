/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath string) *LocalStorage {
	// 确保基础目录存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic(fmt.Sprintf("failed to create base directory: %v", err))
	}
	
	// 确保临时目录存在
	tempDir := filepath.Join(basePath, ".temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create temp directory: %v", err))
	}
	
	return &LocalStorage{
		basePath: basePath,
	}
}

// StorePart 存储分片
func (s *LocalStorage) StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error) {
	tempDir := filepath.Join(s.basePath, ".temp", uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	
	partPath := filepath.Join(tempDir, fmt.Sprintf("part_%05d", partNumber))
	if err := os.WriteFile(partPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write part: %v", err)
	}
	
	// 计算ETag
	hash := sha256.Sum256(data)
	etag := hex.EncodeToString(hash[:])
	
	return etag, nil
}

// MergeParts 合并分片
func (s *LocalStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	tempDir := filepath.Join(s.basePath, ".temp", uploadID)
	finalFilePath := filepath.Join(s.basePath, finalPath)
	
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(finalFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create final directory: %v", err)
	}
	
	// 创建目标文件
	dst, err := os.Create(finalFilePath)
	if err != nil {
		return fmt.Errorf("failed to create final file: %v", err)
	}
	defer dst.Close()
	
	// 按顺序读取分片并写入目标文件
	for i := 1; i <= totalParts; i++ {
		partPath := filepath.Join(tempDir, fmt.Sprintf("part_%05d", i))
		src, err := os.Open(partPath)
		if err != nil {
			return fmt.Errorf("failed to open part %d: %v", i, err)
		}
		
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			return fmt.Errorf("failed to copy part %d: %v", i, err)
		}
		
		src.Close()
	}
	
	return nil
}

// DeleteParts 删除分片
func (s *LocalStorage) DeleteParts(ctx context.Context, uploadID string) error {
	tempDir := filepath.Join(s.basePath, ".temp", uploadID)
	return os.RemoveAll(tempDir)
}

// GetFile 获取文件
func (s *LocalStorage) GetFile(ctx context.Context, path string) ([]byte, error) {
	filePath := filepath.Join(s.basePath, path)
	return os.ReadFile(filePath)
}

// PutFile 存储文件
func (s *LocalStorage) PutFile(ctx context.Context, path string, data []byte) error {
	filePath := filepath.Join(s.basePath, path)
	
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	return os.WriteFile(filePath, data, 0644)
}

// DeleteFile 删除文件
func (s *LocalStorage) DeleteFile(ctx context.Context, path string) error {
	filePath := filepath.Join(s.basePath, path)
	return os.Remove(filePath)
}

// Exists 检查文件是否存在
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	filePath := filepath.Join(s.basePath, path)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Upload 上传文件
func (s *LocalStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(s.basePath, key)
	
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}
	
	return key, nil
}

// Download 下载文件
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	filePath := filepath.Join(s.basePath, key)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Delete 删除文件
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	filePath := filepath.Join(s.basePath, key)
	return os.Remove(filePath)
}

// GetURL 获取文件 URL
func (s *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return "http://localhost:8080/" + key, nil
}
