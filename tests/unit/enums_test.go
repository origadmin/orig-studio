package unit

import (
	"testing"

	"origadmin/application/origstudio/internal/dal/enums"
)

// TestEncodingTaskStatus tests the EncodingTaskStatus enum
func TestEncodingTaskStatus(t *testing.T) {
	// Test string representation
	status := enums.EncodingTaskStatusPending
	if string(status) != "pending" {
		t.Errorf("Expected 'pending', got '%s'", string(status))
	}

	// Test parsing
	parsed := enums.ParseEncodingTaskStatus("pending")
	if parsed != enums.EncodingTaskStatusPending {
		t.Errorf("Expected EncodingTaskStatusPending, got %v", parsed)
	}

	// Test parsing with case insensitivity
	parsed = enums.ParseEncodingTaskStatus("PENDING")
	if parsed != enums.EncodingTaskStatusPending {
		t.Errorf("Expected EncodingTaskStatusPending, got %v", parsed)
	}

	// Test parsing unknown status
	parsed = enums.ParseEncodingTaskStatus("unknown_status")
	if parsed != enums.EncodingTaskStatusUnknown {
		t.Errorf("Expected EncodingTaskStatusUnknown, got %v", parsed)
	}
}

// TestUploadStatus tests the UploadStatus enum
func TestUploadStatus(t *testing.T) {
	// Test string representation
	status := enums.UploadStatusPending
	if string(status) != "pending" {
		t.Errorf("Expected 'pending', got '%s'", string(status))
	}

	// Test parsing
	parsed := enums.ParseUploadStatus("pending")
	if parsed != enums.UploadStatusPending {
		t.Errorf("Expected UploadStatusPending, got %v", parsed)
	}

	// Test parsing with case insensitivity
	parsed = enums.ParseUploadStatus("PENDING")
	if parsed != enums.UploadStatusPending {
		t.Errorf("Expected UploadStatusPending, got %v", parsed)
	}

	// Test parsing unknown status
	parsed = enums.ParseUploadStatus("unknown_status")
	if parsed != enums.UploadStatusUnknown {
		t.Errorf("Expected UploadStatusUnknown, got %v", parsed)
	}
}

// TestMediaEncodingStatus tests the MediaEncodingStatus enum
func TestMediaEncodingStatus(t *testing.T) {
	// Test string representation
	status := enums.MediaEncodingStatusPending
	if string(status) != "pending" {
		t.Errorf("Expected 'pending', got '%s'", string(status))
	}

	// Test parsing
	parsed := enums.ParseMediaEncodingStatus("pending")
	if parsed != enums.MediaEncodingStatusPending {
		t.Errorf("Expected MediaEncodingStatusPending, got %v", parsed)
	}

	// Test parsing with case insensitivity
	parsed = enums.ParseMediaEncodingStatus("PENDING")
	if parsed != enums.MediaEncodingStatusPending {
		t.Errorf("Expected MediaEncodingStatusPending, got %v", parsed)
	}

	// Test parsing unknown status
	parsed = enums.ParseMediaEncodingStatus("unknown_status")
	if parsed != enums.MediaEncodingStatusUnknown {
		t.Errorf("Expected MediaEncodingStatusUnknown, got %v", parsed)
	}
}
