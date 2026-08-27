package dto

import "time"

type LiveRoomDTO struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	StreamKey       string    `json:"stream_key,omitempty"`
	RtmpURL         string    `json:"rtmp_url"`
	HlsURL          string    `json:"hls_url"`
	Status          string    `json:"status"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
	MaxViewers      int       `json:"max_viewers"`
	CurrentViewers  int       `json:"current_viewers"`
	PeakViewers     int       `json:"peak_viewers"`
	Thumbnail       string    `json:"thumbnail"`
	Category        string    `json:"category"`
	Tags            []string  `json:"tags"`
	UserID          string    `json:"user_id"`
	CreateTime      time.Time `json:"create_time"`
	UpdateTime      time.Time `json:"update_time"`
}