/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import "time"

// EncodeProfile 编码配置
type EncodeProfile struct {
	Id              int       `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Extension       string    `json:"extension"`
	Resolution      string    `json:"resolution"`
	VideoCodec      string    `json:"video_codec"`
	VideoBitrate    string    `json:"video_bitrate"`
	AudioCodec      string    `json:"audio_codec"`
	AudioBitrate    string    `json:"audio_bitrate"`
	BentoParameters string    `json:"bento_parameters"`
	IsActive        bool      `json:"is_active"`
	CreateTime      time.Time `json:"create_time"`
	UpdateTime      time.Time `json:"update_time"`
	CommandPreview  string    `json:"command_preview,omitempty"`
}
