/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"origadmin/application/origstudio/internal/dal/enums"
)

// EncodingTask represents a transcoding sub-task for a specific media and profile.
type EncodingTask struct {
	Id           string                   `json:"id"`
	MediaId      string                   `json:"media_id"`
	ProfileId    int                      `json:"profile_id"`
	Status       enums.EncodingTaskStatus `json:"status"` // pending, processing, success, failed
	OutputPath   string                   `json:"output_path"`
	ErrorMessage string                   `json:"error_message"`
	Chunk        bool                     `json:"chunk"` // is chunk? (视频分段转码标识)
	CreateTime   string                   `json:"create_time,omitempty"`
	UpdateTime   string                   `json:"update_time,omitempty"`
}

// StatusCounts holds per-media-status counts.
type StatusCounts struct {
	Processing int `json:"processing"`
	Pending    int `json:"pending"`
	Partial    int `json:"partial"`
	Failed     int `json:"failed"`
	Success    int `json:"success"`
}
