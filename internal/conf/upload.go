/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"time"
)

// UploadConfig 上传配置
type UploadConfig struct {
	ChunkSize        int           `json:"chunk_size" yaml:"chunk_size"`
	MaxFileSize      int64         `json:"max_file_size" yaml:"max_file_size"`
	SessionExpiry    time.Duration `json:"session_expiry" yaml:"session_expiry"`
	StorageBasePath  string        `json:"storage_base_path" yaml:"storage_base_path"`
	AllowedMimeTypes []string      `json:"allowed_mime_types" yaml:"allowed_mime_types"`
}

// TranscodeConfig 转码配置
type TranscodeConfig struct {
	MaxParallelTasks int           `json:"max_parallel_tasks" yaml:"max_parallel_tasks"`
	TaskTimeout      time.Duration `json:"task_timeout" yaml:"task_timeout"`
	HLSBasePath      string        `json:"hls_base_path" yaml:"hls_base_path"`
	ThumbnailBasePath string       `json:"thumbnail_base_path" yaml:"thumbnail_base_path"`
	PreviewBasePath  string        `json:"preview_base_path" yaml:"preview_base_path"`
}

// DefaultUploadConfig 返回默认上传配置
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		ChunkSize:        5 * 1024 * 1024, // 5MB
		MaxFileSize:      10 * 1024 * 1024 * 1024, // 10GB
		SessionExpiry:    24 * time.Hour,
		StorageBasePath:  "./data/uploads",
		AllowedMimeTypes: []string{
			"video/mp4",
			"video/webm",
			"video/avi",
			"video/mkv",
			"audio/mp3",
			"audio/wav",
			"audio/ogg",
			"image/jpeg",
			"image/png",
			"image/gif",
			"application/pdf",
		},
	}
}

// DefaultTranscodeConfig 返回默认转码配置
func DefaultTranscodeConfig() *TranscodeConfig {
	return &TranscodeConfig{
		MaxParallelTasks: 4,
		TaskTimeout:      4 * time.Hour,
		HLSBasePath:      "./data/hls",
		ThumbnailBasePath: "./data/thumbnails",
		PreviewBasePath:  "./data/previews",
	}
}
