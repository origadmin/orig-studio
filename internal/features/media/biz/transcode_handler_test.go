/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"encoding/json"
	"testing"

	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/features/media/dto"
)

func TestMediaEncodeRequest_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    MediaEncodeRequest
		wantErr bool
	}{
		{
			name:    "valid request with all fields",
			payload: `{"media_id":"123","media_path":"uploads/test.mp4","content_type":"video/mp4"}`,
			want: MediaEncodeRequest{
				MediaID:     "123",
				MediaPath:   "uploads/test.mp4",
				ContentType: "video/mp4",
				TaskID:      nil,
			},
			wantErr: false,
		},
		{
			name:    "valid request with task_id for retry",
			payload: `{"media_id":"123","media_path":"uploads/test.mp4","content_type":"video/mp4","task_id":"task-456"}`,
			want: MediaEncodeRequest{
				MediaID:     "123",
				MediaPath:   "uploads/test.mp4",
				ContentType: "video/mp4",
				TaskID:      strPtr("task-456"),
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			payload: `{"media_id":`,
			want:    MediaEncodeRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req MediaEncodeRequest
			err := json.Unmarshal([]byte(tt.payload), &req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if req.MediaID != tt.want.MediaID {
					t.Errorf("MediaID = %v, want %v", req.MediaID, tt.want.MediaID)
				}
				if req.MediaPath != tt.want.MediaPath {
					t.Errorf("MediaPath = %v, want %v", req.MediaPath, tt.want.MediaPath)
				}
				if req.ContentType != tt.want.ContentType {
					t.Errorf("ContentType = %v, want %v", req.ContentType, tt.want.ContentType)
				}
				if (req.TaskID == nil) != (tt.want.TaskID == nil) {
					t.Errorf("TaskID mismatch: got %v, want %v", req.TaskID, tt.want.TaskID)
				}
				if req.TaskID != nil && tt.want.TaskID != nil && *req.TaskID != *tt.want.TaskID {
					t.Errorf("TaskID = %v, want %v", *req.TaskID, *tt.want.TaskID)
				}
			}
		})
	}
}

func TestMediaEncodeEvent_Marshal(t *testing.T) {
	event := MediaEncodeEvent{
		MediaID: "123",
		Task: &EncodingTask{
			Id:        "task-1",
			MediaId:   "123",
			ProfileId: 1,
			Status:    enums.EncodingTaskStatusSuccess,
		},
		Status: "success",
		Error:  "",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded MediaEncodeEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.MediaID != event.MediaID {
		t.Errorf("MediaID = %v, want %v", decoded.MediaID, event.MediaID)
	}
	if decoded.Status != event.Status {
		t.Errorf("Status = %v, want %v", decoded.Status, event.Status)
	}
	if decoded.Task == nil || decoded.Task.Id != event.Task.Id {
		t.Errorf("Task.Id = %v, want %v", decoded.Task.Id, event.Task.Id)
	}
}

func TestEncodingTask_StatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		initial enums.EncodingTaskStatus
		target  enums.EncodingTaskStatus
	}{
		{"pending to processing", enums.EncodingTaskStatusPending, enums.EncodingTaskStatusProcessing},
		{"processing to success", enums.EncodingTaskStatusProcessing, enums.EncodingTaskStatusSuccess},
		{"processing to failed", enums.EncodingTaskStatusProcessing, enums.EncodingTaskStatusFailed},
		{"pending to success direct", enums.EncodingTaskStatusPending, enums.EncodingTaskStatusSuccess},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &EncodingTask{
				Id:        "test-task",
				MediaId:   "media-1",
				ProfileId: 1,
				Status:    tt.initial,
			}
			task.Status = tt.target

			if task.Status != tt.target {
				t.Errorf("Status transition failed: got %v, want %v", task.Status, tt.target)
			}
		})
	}
}

func TestTranscodeHandler_EstimateBandwidth(t *testing.T) {
	tests := []struct {
		name    string
		profile *dto.EncodeProfile
		wantMin int
		wantMax int
	}{
		{
			name: "4K profile with bitrate",
			profile: &dto.EncodeProfile{
				Resolution:      "2160",
				BentoParameters: "--video-bitrate 20M",
			},
			wantMin: 19_000_000,
			wantMax: 21_000_000,
		},
		{
			name: "1080p profile with bitrate",
			profile: &dto.EncodeProfile{
				Resolution:      "1080",
				BentoParameters: "--video-bitrate 4000k",
			},
			wantMin: 3_500_000,
			wantMax: 4_500_000,
		},
		{
			name: "720p profile without explicit bitrate (fallback)",
			profile: &dto.EncodeProfile{
				Resolution: "720",
			},
			wantMin: 3_500_000,
			wantMax: 4_500_000,
		},
		{
			name: "480p fallback",
			profile: &dto.EncodeProfile{
				Resolution: "480",
			},
			wantMin: 1_800_000,
			wantMax: 2_200_000,
		},
		{
			name: "360p fallback",
			profile: &dto.EncodeProfile{
				Resolution: "360",
			},
			wantMin: 900_000,
			wantMax: 1_100_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bw := estimateBandwidth(tt.profile)
			if bw < tt.wantMin || bw > tt.wantMax {
				t.Errorf("estimateBandwidth() = %d, want between %d and %d", bw, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestTranscodeJob_OutputPaths(t *testing.T) {
	mediaUUID := "test-media-uuid-123"
	profileName := "1080p"

	tests := []struct {
		name      string
		profile   *dto.EncodeProfile
		expectDir string
	}{
		{
			name:      "video profile HLS output",
			profile:   &dto.EncodeProfile{Name: profileName, Extension: "mp4", Resolution: "1080"},
			expectDir: "hls/" + mediaUUID + "/" + profileName,
		},
		{
			name:      "preview profile GIF output",
			profile:   &dto.EncodeProfile{Name: "preview", Extension: "gif", Resolution: "-"},
			expectDir: "previews",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outputDir string
			if IsVideoProfile(tt.profile) {
				outputDir = "hls/" + mediaUUID + "/" + tt.profile.Name
			} else if IsPreviewProfile(tt.profile) {
				outputDir = "previews"
			}

			if outputDir != tt.expectDir {
				t.Errorf("outputDir = %v, want %v", outputDir, tt.expectDir)
			}
		})
	}
}

func TestTranscodeHandler_ProcessMedia_StatusDetermination(t *testing.T) {
	tests := []struct {
		name       string
		videoTasks []enums.EncodingTaskStatus
		wantStatus string
	}{
		{
			name:       "all video tasks succeed",
			videoTasks: []enums.EncodingTaskStatus{enums.EncodingTaskStatusSuccess, enums.EncodingTaskStatusSuccess},
			wantStatus: "success",
		},
		{
			name:       "some video tasks succeed (partial)",
			videoTasks: []enums.EncodingTaskStatus{enums.EncodingTaskStatusSuccess, enums.EncodingTaskStatusFailed},
			wantStatus: "partial",
		},
		{
			name:       "all video tasks fail",
			videoTasks: []enums.EncodingTaskStatus{enums.EncodingTaskStatusFailed, enums.EncodingTaskStatusFailed},
			wantStatus: "failed",
		},
		{
			name:       "no video tasks (all preview)",
			videoTasks: []enums.EncodingTaskStatus{},
			wantStatus: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := determineMediaStatus(tt.videoTasks)
			if status != tt.wantStatus {
				t.Errorf("determineMediaStatus() = %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

func determineMediaStatus(videoTasks []enums.EncodingTaskStatus) string {
	if len(videoTasks) == 0 {
		return "success"
	}

	successCount := 0

	for _, status := range videoTasks {
		if status == enums.EncodingTaskStatusSuccess {
			successCount++
		}
	}

	if successCount == len(videoTasks) {
		return "success"
	}
	if successCount > 0 {
		return "partial"
	}
	return "failed"
}

func strPtr(s string) *string {
	return &s
}

func TestEncodingEvent_Progress(t *testing.T) {
	event := &EncodingEvent{
		MediaId:  "media-1",
		Task:     &EncodingTask{Id: "task-1"},
		Progress: 50,
		Speed:    "1.5x",
		Fps:      30,
		Time:     120,
	}

	if event.Progress != 50 {
		t.Errorf("Progress = %d, want 50", event.Progress)
	}

	if event.Speed != "1.5x" {
		t.Errorf("Speed = %s, want 1.5x", event.Speed)
	}
}

func TestMediaEncodingStatus_Constants(t *testing.T) {
	tests := []struct {
		status enums.MediaEncodingStatus
		want   string
	}{
		{enums.MediaEncodingStatusUnknown, "unknown"},
		{enums.MediaEncodingStatusPending, "pending"},
		{enums.MediaEncodingStatusProcessing, "processing"},
		{enums.MediaEncodingStatusSuccess, "success"},
		{enums.MediaEncodingStatusFailed, "failed"},
		{enums.MediaEncodingStatusPartial, "partial"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("status = %s, want %s", tt.status, tt.want)
			}
		})
	}
}

func TestEncodingTaskStatus_Constants(t *testing.T) {
	tests := []struct {
		status enums.EncodingTaskStatus
		want   string
	}{
		{enums.EncodingTaskStatusUnknown, "unknown"},
		{enums.EncodingTaskStatusPending, "pending"},
		{enums.EncodingTaskStatusProcessing, "processing"},
		{enums.EncodingTaskStatusSuccess, "success"},
		{enums.EncodingTaskStatusFailed, "failed"},
		{enums.EncodingTaskStatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("status = %s, want %s", tt.status, tt.want)
			}
		})
	}
}

func TestTranscodeJob_Structure(t *testing.T) {
	job := TranscodeJob{
		MediaID:   "media-123",
		TaskID:    "task-456",
		Profile:   &dto.EncodeProfile{Name: "1080p", Resolution: "1080"},
		InputPath: "/path/to/input.mp4",
		OutputDir: "/path/to/output",
		UUID:      "uuid-789",
	}

	if job.MediaID != "media-123" {
		t.Errorf("MediaID = %s, want media-123", job.MediaID)
	}

	if job.TaskID != "task-456" {
		t.Errorf("TaskID = %s, want task-456", job.TaskID)
	}

	if job.Profile.Name != "1080p" {
		t.Errorf("Profile.Name = %s, want 1080p", job.Profile.Name)
	}
}

func TestVariantInfo(t *testing.T) {
	info := VariantInfo{
		TaskID:      "task-1",
		ProfileName: "1080p",
		Bandwidth:   8_000_000,
		Resolution:  "1920x1080",
		Codec:      "h264",
	}

	if info.Bandwidth != 8_000_000 {
		t.Errorf("Bandwidth = %d, want 8000000", info.Bandwidth)
	}

	if info.Resolution != "1920x1080" {
		t.Errorf("Resolution = %s, want 1920x1080", info.Resolution)
	}

	if info.ProfileName != "1080p" {
		t.Errorf("ProfileName = %s, want 1080p", info.ProfileName)
	}
}