package enums

import (
	"strings"
)

// Status defines a general-purpose status for various entities.
type Status string

const (
	StatusUnknown  Status = "unknown"
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
	StatusInvalid  Status = StatusUnknown
)

// EncodingTaskStatus represents encoding task status
type EncodingTaskStatus string

const (
	EncodingTaskStatusUnknown    EncodingTaskStatus = "unknown"
	EncodingTaskStatusPending    EncodingTaskStatus = "pending"
	EncodingTaskStatusProcessing EncodingTaskStatus = "processing"
	EncodingTaskStatusSuccess    EncodingTaskStatus = "success"
	EncodingTaskStatusFailed     EncodingTaskStatus = "failed"
	EncodingTaskStatusSkipped    EncodingTaskStatus = "skipped"
	EncodingTaskStatusPartial    EncodingTaskStatus = "partial"
	EncodingTaskStatusInvalid    EncodingTaskStatus = EncodingTaskStatusUnknown
)

// UploadStatus represents upload session status
type UploadStatus string

const (
	UploadStatusUnknown   UploadStatus = "unknown"
	UploadStatusPending   UploadStatus = "pending"
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusAborted   UploadStatus = "aborted"
	UploadStatusInvalid   UploadStatus = UploadStatusUnknown
)

// MediaEncodingStatus represents media encoding status
type MediaEncodingStatus string

const (
	MediaEncodingStatusUnknown    MediaEncodingStatus = "unknown"
	MediaEncodingStatusPending    MediaEncodingStatus = "pending"
	MediaEncodingStatusProcessing MediaEncodingStatus = "processing"
	MediaEncodingStatusSuccess    MediaEncodingStatus = "success"
	MediaEncodingStatusFailed     MediaEncodingStatus = "failed"
	MediaEncodingStatusPartial    MediaEncodingStatus = "partial"
	MediaEncodingStatusInvalid    MediaEncodingStatus = MediaEncodingStatusUnknown
)

// ParseEncodingTaskStatus parses encoding task status strings
func ParseEncodingTaskStatus(from string) EncodingTaskStatus {
	switch strings.ToLower(from) {
	case string(EncodingTaskStatusPending):
		return EncodingTaskStatusPending
	case string(EncodingTaskStatusProcessing):
		return EncodingTaskStatusProcessing
	case string(EncodingTaskStatusSuccess):
		return EncodingTaskStatusSuccess
	case string(EncodingTaskStatusFailed):
		return EncodingTaskStatusFailed
	case string(EncodingTaskStatusSkipped):
		return EncodingTaskStatusSkipped
	case string(EncodingTaskStatusPartial):
		return EncodingTaskStatusPartial
	default:
		return EncodingTaskStatusUnknown
	}
}

// ParseUploadStatus parses upload status strings
func ParseUploadStatus(from string) UploadStatus {
	switch strings.ToLower(from) {
	case string(UploadStatusPending):
		return UploadStatusPending
	case string(UploadStatusUploading):
		return UploadStatusUploading
	case string(UploadStatusCompleted):
		return UploadStatusCompleted
	case string(UploadStatusAborted):
		return UploadStatusAborted
	default:
		return UploadStatusUnknown
	}
}

// ParseMediaEncodingStatus parses media encoding status strings
func ParseMediaEncodingStatus(from string) MediaEncodingStatus {
	switch strings.ToLower(from) {
	case string(MediaEncodingStatusPending):
		return MediaEncodingStatusPending
	case string(MediaEncodingStatusProcessing):
		return MediaEncodingStatusProcessing
	case string(MediaEncodingStatusSuccess):
		return MediaEncodingStatusSuccess
	case string(MediaEncodingStatusFailed):
		return MediaEncodingStatusFailed
	case string(MediaEncodingStatusPartial):
		return MediaEncodingStatusPartial
	default:
		return MediaEncodingStatusUnknown
	}
}

// SyncStatus represents the S3 sync status of a media file.
type SyncStatus string

const (
	SyncStatusLocalOnly SyncStatus = "local_only"
	SyncStatusSyncing   SyncStatus = "syncing"
	SyncStatusSynced    SyncStatus = "synced"
	SyncStatusFailed    SyncStatus = "failed"
)

// ParseSyncStatus parses sync status strings.
func ParseSyncStatus(from string) SyncStatus {
	switch strings.ToLower(from) {
	case string(SyncStatusLocalOnly):
		return SyncStatusLocalOnly
	case string(SyncStatusSyncing):
		return SyncStatusSyncing
	case string(SyncStatusSynced):
		return SyncStatusSynced
	case string(SyncStatusFailed):
		return SyncStatusFailed
	default:
		return SyncStatusLocalOnly
	}
}

// Values returns all possible values for SyncStatus.
func (SyncStatus) Values() []string {
	return []string{
		string(SyncStatusLocalOnly),
		string(SyncStatusSyncing),
		string(SyncStatusSynced),
		string(SyncStatusFailed),
	}
}

func (s SyncStatus) String() string {
	return string(s)
}

// ParseStatus parses general status strings for enable/disable statuses
func ParseStatus(from string) Status {
	switch strings.ToLower(from) {
	case string(StatusEnabled):
		return StatusEnabled
	case string(StatusDisabled):
		return StatusDisabled
	default:
		return StatusUnknown
	}
}

// Values returns all possible values for Status
func (Status) Values() []string {
	return []string{
		string(StatusUnknown),
		string(StatusEnabled),
		string(StatusDisabled),
	}
}

// Values returns all possible values for EncodingTaskStatus
func (EncodingTaskStatus) Values() []string {
	return []string{
		string(EncodingTaskStatusUnknown),
		string(EncodingTaskStatusPending),
		string(EncodingTaskStatusProcessing),
		string(EncodingTaskStatusSuccess),
		string(EncodingTaskStatusFailed),
		string(EncodingTaskStatusSkipped),
		string(EncodingTaskStatusPartial),
	}
}

func (s EncodingTaskStatus) String() string {
	return string(s)
}

// Values returns all possible values for UploadStatus
func (UploadStatus) Values() []string {
	return []string{
		string(UploadStatusUnknown),
		string(UploadStatusPending),
		string(UploadStatusUploading),
		string(UploadStatusCompleted),
		string(UploadStatusAborted),
	}
}

// Values returns all possible values for MediaEncodingStatus
func (MediaEncodingStatus) Values() []string {
	return []string{
		string(MediaEncodingStatusUnknown),
		string(MediaEncodingStatusPending),
		string(MediaEncodingStatusProcessing),
		string(MediaEncodingStatusSuccess),
		string(MediaEncodingStatusFailed),
		string(MediaEncodingStatusPartial),
	}
}
