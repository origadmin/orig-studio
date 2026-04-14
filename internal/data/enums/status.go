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
	EncodingTaskStatusUnknown   EncodingTaskStatus = "unknown"
	EncodingTaskStatusPending   EncodingTaskStatus = "pending"
	EncodingTaskStatusProcessing EncodingTaskStatus = "processing"
	EncodingTaskStatusSuccess   EncodingTaskStatus = "success"
	EncodingTaskStatusFailed    EncodingTaskStatus = "failed"
	EncodingTaskStatusSkipped   EncodingTaskStatus = "skipped"
	EncodingTaskStatusPartial   EncodingTaskStatus = "partial"
	EncodingTaskStatusInvalid   EncodingTaskStatus = EncodingTaskStatusUnknown
)

// UploadStatus represents upload session status
type UploadStatus string

const (
	UploadStatusUnknown    UploadStatus = "unknown"
	UploadStatusPending    UploadStatus = "pending"
	UploadStatusUploading  UploadStatus = "uploading"
	UploadStatusCompleted  UploadStatus = "completed"
	UploadStatusAborted    UploadStatus = "aborted"
	UploadStatusInvalid    UploadStatus = UploadStatusUnknown
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
	case "pending":
		return EncodingTaskStatusPending
	case "processing":
		return EncodingTaskStatusProcessing
	case "success":
		return EncodingTaskStatusSuccess
	case "failed":
		return EncodingTaskStatusFailed
	case "skipped":
		return EncodingTaskStatusSkipped
	case "partial":
		return EncodingTaskStatusPartial
	default:
		return EncodingTaskStatusUnknown
	}
}

// ParseUploadStatus parses upload status strings
func ParseUploadStatus(from string) UploadStatus {
	switch strings.ToLower(from) {
	case "pending":
		return UploadStatusPending
	case "uploading":
		return UploadStatusUploading
	case "completed":
		return UploadStatusCompleted
	case "aborted":
		return UploadStatusAborted
	default:
		return UploadStatusUnknown
	}
}

// ParseMediaEncodingStatus parses media encoding status strings
func ParseMediaEncodingStatus(from string) MediaEncodingStatus {
	switch strings.ToLower(from) {
	case "pending":
		return MediaEncodingStatusPending
	case "processing":
		return MediaEncodingStatusProcessing
	case "success":
		return MediaEncodingStatusSuccess
	case "failed":
		return MediaEncodingStatusFailed
	case "partial":
		return MediaEncodingStatusPartial
	default:
		return MediaEncodingStatusUnknown
	}
}

// ParseStatus parses general status strings for enable/disable statuses
func ParseStatus(from string) Status {
	switch strings.ToLower(from) {
	case "enabled":
		return StatusEnabled
	case "disabled":
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
