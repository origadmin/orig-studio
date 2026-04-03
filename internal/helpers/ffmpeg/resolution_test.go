package ffmpeg

import "testing"

func TestResolutionToSize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"240", "426x240"},
		{"360", "640x360"},
		{"480", "854x480"},
		{"720", "1280x720"},
		{"1080", "1920x1080"},
		{"1440", "2560x1440"},
		{"2160", "3840x2160"},
		{"1280x720", "1280x720"}, // already WxH
		{"1920x1080", "1920x1080"},
		{"unknown", "unknown"}, // passthrough
	}

	for _, tt := range tests {
		got := ResolutionToSize(tt.input)
		if got != tt.expected {
			t.Errorf("ResolutionToSize(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsSkipResolution(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"-", true},
		{"", true},
		{"720", false},
		{"1080", false},
		{"1280x720", false},
	}

	for _, tt := range tests {
		got := IsSkipResolution(tt.input)
		if got != tt.expected {
			t.Errorf("IsSkipResolution(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
