/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

// EncodeProfile 编码配置
type EncodeProfile struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Extension       string `json:"extension"`
	Resolution      string `json:"resolution"`
	VideoCodec      string `json:"video_codec"`
	VideoBitrate    string `json:"video_bitrate"`
	AudioCodec      string `json:"audio_codec"`
	AudioBitrate    string `json:"audio_bitrate"`
	BentoParameters string `json:"bento_parameters"`
	IsActive        bool   `json:"is_active"`
	CommandPreview  string `json:"command_preview,omitempty"`
}
