package ffmpeg

import (
	"strconv"
	"strings"
	"testing"
)

func TestMapCodecName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"h264", "libx264"},
		{"h265", "libx265"},
		{"hevc", "libx265"},
		{"vp9", "libvpx-vp9"},
		{"unknown", "libx264"},
		{"", "libx264"},
	}

	for _, tt := range tests {
		got := MapCodecName(tt.input)
		if got != tt.expected {
			t.Errorf("MapCodecName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapAudioCodec(t *testing.T) {
	tests := []struct {
		audioCodec   string
		videoEncoder string
		expected     string
	}{
		// Explicit audio codec
		{"aac", "libx264", "aac"},
		{"libaac", "libx264", "aac"},
		{"opus", "libx264", "libopus"},
		{"libopus", "libx265", "libopus"},
		// Empty or "-" audio codec: default by video encoder
		{"", "libx264", "aac"},
		{"", "libx265", "aac"},
		{"", "libvpx-vp9", "libopus"},
		{"-", "libx264", "aac"},
		{"-", "libvpx-vp9", "libopus"},
	}

	for _, tt := range tests {
		got := MapAudioCodec(tt.audioCodec, tt.videoEncoder)
		if got != tt.expected {
			t.Errorf("MapAudioCodec(%q, %q) = %q, want %q", tt.audioCodec, tt.videoEncoder, got, tt.expected)
		}
	}
}

func TestLevelForResolution(t *testing.T) {
	tests := []struct {
		resolution string
		expected   string
	}{
		// Short-form resolutions
		{"240", "3.0"},
		{"360", "3.0"},
		{"480", "3.0"},
		{"720", "4.1"},
		{"1080", "4.2"},
		{"1440", "5.1"},
		{"2160", "5.2"},
		// WxH format
		{"1280x720", "4.1"},
		{"1920x1080", "4.2"},
		{"3840x2160", "5.2"},
		// Unknown defaults to 4.1
		{"9999", "4.1"},
	}

	for _, tt := range tests {
		got := levelForResolution(tt.resolution)
		if got != tt.expected {
			t.Errorf("levelForResolution(%q) = %q, want %q", tt.resolution, got, tt.expected)
		}
	}
}

func TestBuildVideoArgs_H264(t *testing.T) {
	args := buildVideoArgs("h264", "720", "2500k", "aac", "128k")

	argsStr := strings.Join(args, " ")

	// Must contain scale filter with aspect ratio preservation
	if !strings.Contains(argsStr, "force_original_aspect_ratio=decrease") {
		t.Error("H.264 args missing force_original_aspect_ratio=decrease in scale filter")
	}
	if !strings.Contains(argsStr, "force_divisible_by=2") {
		t.Error("H.264 args missing force_divisible_by=2 in scale filter")
	}
	if !strings.Contains(argsStr, "flags=lanczos") {
		t.Error("H.264 args missing flags=lanczos in scale filter")
	}
	if !strings.Contains(argsStr, "scale=1280:720") {
		t.Error("H.264 args missing scale=1280:720")
	}

	// Must contain CRF
	if !containsPair(args, "-crf", "23") {
		t.Error("H.264 args missing -crf 23")
	}

	// Must contain preset
	if !containsPair(args, "-preset", "medium") {
		t.Error("H.264 args missing -preset medium")
	}

	// Must contain profile and level
	if !containsPair(args, "-profile:v", "main") {
		t.Error("H.264 args missing -profile:v main")
	}
	if !containsPair(args, "-level", "4.1") {
		t.Error("H.264 args missing -level 4.1 for 720p")
	}

	// Must contain pix_fmt
	if !containsPair(args, "-pix_fmt", "yuv420p") {
		t.Error("H.264 args missing -pix_fmt yuv420p")
	}

	// Must contain key frame control
	if !containsPair(args, "-force_key_frames", "expr:gte(t,n_forced*4)") {
		t.Error("H.264 args missing -force_key_frames")
	}
	if !containsPair(args, "-x264-params", "keyint=240:keyint_min=120") {
		t.Error("H.264 args missing -x264-params")
	}

	// Must contain maxrate/bufsize from videoBitrate
	if !containsPair(args, "-maxrate", "2500k") {
		t.Error("H.264 args missing -maxrate 2500k")
	}
	if !containsPair(args, "-bufsize", "2500k") {
		t.Error("H.264 args missing -bufsize 2500k")
	}

	// Must contain audio codec and bitrate
	if !containsPair(args, "-c:a", "aac") {
		t.Error("H.264 args missing -c:a aac")
	}
	if !containsPair(args, "-b:a", "128k") {
		t.Error("H.264 args missing -b:a 128k")
	}
}

func TestBuildVideoArgs_H265(t *testing.T) {
	args := buildVideoArgs("h265", "1080", "5000k", "aac", "192k")

	argsStr := strings.Join(args, " ")

	// Must contain scale filter
	if !strings.Contains(argsStr, "scale=1920:1080") {
		t.Error("H.265 args missing scale=1920:1080")
	}

	// Must contain CRF (28 for H.265)
	if !containsPair(args, "-crf", "28") {
		t.Error("H.265 args missing -crf 28")
	}

	// Must contain preset
	if !containsPair(args, "-preset", "medium") {
		t.Error("H.265 args missing -preset medium")
	}

	// Must contain profile
	if !containsPair(args, "-profile:v", "main") {
		t.Error("H.265 args missing -profile:v main")
	}

	// Must contain level for 1080p
	if !containsPair(args, "-level", "4.2") {
		t.Error("H.265 args missing -level 4.2 for 1080p")
	}

	// Must contain x265-params
	if !containsPair(args, "-x265-params", "keyint=240:keyint_min=120") {
		t.Error("H.265 args missing -x265-params")
	}

	// Must contain pix_fmt
	if !containsPair(args, "-pix_fmt", "yuv420p") {
		t.Error("H.265 args missing -pix_fmt yuv420p")
	}
}

func TestBuildVideoArgs_VP9(t *testing.T) {
	args := buildVideoArgs("vp9", "720", "2500k", "", "128k")

	argsStr := strings.Join(args, " ")

	// Must contain scale filter
	if !strings.Contains(argsStr, "scale=1280:720") {
		t.Error("VP9 args missing scale=1280:720")
	}

	// Must contain CRF (31 for VP9)
	if !containsPair(args, "-crf", "31") {
		t.Error("VP9 args missing -crf 31")
	}

	// Must contain -b:v 0 for VP9 CRF mode
	if !containsPair(args, "-b:v", "0") {
		t.Error("VP9 args missing -b:v 0 for CRF mode")
	}

	// Must contain quality good
	if !containsPair(args, "-quality", "good") {
		t.Error("VP9 args missing -quality good")
	}

	// Must contain cpu-used
	if !containsPair(args, "-cpu-used", "2") {
		t.Error("VP9 args missing -cpu-used 2")
	}

	// VP9 should NOT have pix_fmt yuv420p (VP9 handles pixel format internally)
	if containsPair(args, "-pix_fmt", "yuv420p") {
		t.Error("VP9 args should not contain -pix_fmt yuv420p")
	}

	// VP9 should NOT have preset
	if containsPair(args, "-preset", "medium") {
		t.Error("VP9 args should not contain -preset")
	}

	// VP9 default audio should be libopus
	if !containsPair(args, "-c:a", "libopus") {
		t.Error("VP9 args missing -c:a libopus (default for VP9)")
	}
}

func TestBuildVideoArgs_NoBitrate(t *testing.T) {
	args := buildVideoArgs("h264", "720", "", "aac", "")

	argsStr := strings.Join(args, " ")

	// Should NOT contain maxrate/bufsize when videoBitrate is empty
	if strings.Contains(argsStr, "-maxrate") {
		t.Error("H.264 args should not contain -maxrate when videoBitrate is empty")
	}
	if strings.Contains(argsStr, "-bufsize") {
		t.Error("H.264 args should not contain -bufsize when videoBitrate is empty")
	}

	// Should NOT contain -b:a when audioBitrate is empty
	if strings.Contains(argsStr, "-b:a") {
		t.Error("H.264 args should not contain -b:a when audioBitrate is empty")
	}
}

func TestBuildVideoArgs_AutoBitrate(t *testing.T) {
	args := buildVideoArgs("h264", "720", "auto", "aac", "128k")

	argsStr := strings.Join(args, " ")

	// Should NOT contain maxrate/bufsize when videoBitrate is "auto"
	if strings.Contains(argsStr, "-maxrate") {
		t.Error("H.264 args should not contain -maxrate when videoBitrate is auto")
	}
	if strings.Contains(argsStr, "-bufsize") {
		t.Error("H.264 args should not contain -bufsize when videoBitrate is auto")
	}
}

func TestMapCodecName_InvalidLongNames(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"libx264", "libx264"},
		{"libx265", "libx264"},
		{"libvpx-vp9", "libx264"},
	}

	for _, tt := range tests {
		got := MapCodecName(tt.input)
		if got != tt.expected {
			t.Errorf("MapCodecName(%q) = %q, want %q (long names should fallback to default)", tt.input, got, tt.expected)
		}
	}
}

func TestResolutionToWidthHeight(t *testing.T) {
	tests := []struct {
		resolution    string
		expectedW     string
		expectedH     string
	}{
		{"720", "1280", "720"},
		{"1080", "1920", "1080"},
		{"240", "426", "240"},
		{"1280x720", "1280", "720"},
		{"1920x1080", "1920", "1080"},
	}

	for _, tt := range tests {
		w, h := ResolutionToWidthHeight(tt.resolution)
		if w != tt.expectedW || h != tt.expectedH {
			t.Errorf("ResolutionToWidthHeight(%q) = (%q, %q), want (%q, %q)",
				tt.resolution, w, h, tt.expectedW, tt.expectedH)
		}
	}
}

// containsPair checks if the args slice contains a consecutive key-value pair.
func containsPair(args []string, key, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}

// --- Preview Command Tests (TDD Red Phase) ---

func TestPreviewHLSCommand_H264(t *testing.T) {
	cmd := PreviewHLSCommand("<input_file>", "<output_dir>/720p", "720p", "720", "h264", "aac", "2500k", "128k")

	// Must start with ffmpeg
	if !strings.HasPrefix(cmd, "ffmpeg ") {
		t.Error("HLS preview command must start with 'ffmpeg '")
	}

	// Must contain input file
	if !strings.Contains(cmd, "-i <input_file>") {
		t.Error("HLS preview command missing '-i <input_file>'")
	}

	// Must contain HLS format flags
	if !strings.Contains(cmd, "-f hls") {
		t.Error("HLS preview command missing '-f hls'")
	}
	if !strings.Contains(cmd, "-hls_time 6") {
		t.Error("HLS preview command missing '-hls_time 6'")
	}
	if !strings.Contains(cmd, "-hls_list_size 0") {
		t.Error("HLS preview command missing '-hls_list_size 0'")
	}

	// Must contain H.264 specific args
	if !strings.Contains(cmd, "-c:v libx264") {
		t.Error("HLS preview command missing '-c:v libx264'")
	}
	if !strings.Contains(cmd, "-crf 23") {
		t.Error("HLS preview command missing '-crf 23'")
	}

	// Must contain maxrate/bufsize
	if !strings.Contains(cmd, "-maxrate 2500k") {
		t.Error("HLS preview command missing '-maxrate 2500k'")
	}

	// Must contain audio codec
	if !strings.Contains(cmd, "-c:a aac") {
		t.Error("HLS preview command missing '-c:a aac'")
	}
	if !strings.Contains(cmd, "-b:a 128k") {
		t.Error("HLS preview command missing '-b:a 128k'")
	}

	// Must contain output paths
	if !strings.Contains(cmd, "segment_%03d.ts") {
		t.Error("HLS preview command missing segment pattern")
	}
	if !strings.Contains(cmd, "index.m3u8") {
		t.Error("HLS preview command missing index.m3u8")
	}

	// Must contain -y flag
	if !strings.Contains(cmd, "-y") {
		t.Error("HLS preview command missing '-y'")
	}
}

func TestPreviewHLSCommand_VP9(t *testing.T) {
	cmd := PreviewHLSCommand("<input_file>", "<output_dir>/720p", "720p", "720", "vp9", "", "2500k", "128k")

	// Must contain VP9 specific args
	if !strings.Contains(cmd, "-c:v libvpx-vp9") {
		t.Error("VP9 HLS preview command missing '-c:v libvpx-vp9'")
	}
	if !strings.Contains(cmd, "-crf 31") {
		t.Error("VP9 HLS preview command missing '-crf 31'")
	}

	// VP9 default audio should be libopus
	if !strings.Contains(cmd, "-c:a libopus") {
		t.Error("VP9 HLS preview command missing '-c:a libopus'")
	}
}

func TestPreviewHLSCommand_NoBitrate(t *testing.T) {
	cmd := PreviewHLSCommand("<input_file>", "<output_dir>/720p", "720p", "720", "h264", "aac", "", "")

	// Should NOT contain maxrate/bufsize when videoBitrate is empty
	if strings.Contains(cmd, "-maxrate") {
		t.Error("HLS preview should not contain -maxrate when videoBitrate is empty")
	}
	if strings.Contains(cmd, "-bufsize") {
		t.Error("HLS preview should not contain -bufsize when videoBitrate is empty")
	}

	// Should NOT contain -b:a when audioBitrate is empty
	if strings.Contains(cmd, "-b:a") {
		t.Error("HLS preview should not contain -b:a when audioBitrate is empty")
	}
}

func TestPreviewGIFCommand(t *testing.T) {
	cmd := PreviewGIFCommand("<input_file>", "<output_dir>/preview.gif", "320:-1")

	// Must contain two commands joined by &&
	parts := strings.Split(cmd, " && ")
	if len(parts) != 2 {
		t.Fatalf("GIF preview command must contain two commands joined by ' && ', got %d parts", len(parts))
	}

	// First command: palette generation
	paletteCmd := parts[0]
	if !strings.HasPrefix(paletteCmd, "ffmpeg ") {
		t.Error("Palette command must start with 'ffmpeg '")
	}
	if !strings.Contains(paletteCmd, "-i <input_file>") {
		t.Error("Palette command missing '-i <input_file>'")
	}
	if !strings.Contains(paletteCmd, "palettegen") {
		t.Error("Palette command missing 'palettegen'")
	}
	if !strings.Contains(paletteCmd, "scale=320:-1") {
		t.Error("Palette command missing 'scale=320:-1'")
	}
	if !strings.Contains(paletteCmd, "-t 3") {
		t.Error("Palette command missing '-t 3'")
	}

	// Second command: GIF generation
	gifCmd := parts[1]
	if !strings.HasPrefix(gifCmd, "ffmpeg ") {
		t.Error("GIF command must start with 'ffmpeg '")
	}
	if !strings.Contains(gifCmd, "paletteuse") {
		t.Error("GIF command missing 'paletteuse'")
	}
	if !strings.Contains(gifCmd, "<output_dir>/preview.gif") {
		t.Error("GIF command missing output path")
	}
}

func TestGetVideoResolution_ParseOutput(t *testing.T) {
	tests := []struct {
		input    string
		w, h     int
		hasError bool
	}{
		{"1920x1080\n", 1920, 1080, false},
		{"1280x720\n", 1280, 720, false},
		{"640x360\n", 640, 360, false},
		{"invalid\n", 0, 0, true},
		{"1920\n", 0, 0, true},
		{"ax720\n", 0, 0, true},
	}

	for _, tt := range tests {
		parts := strings.Split(strings.TrimSpace(tt.input), "x")
		if len(parts) != 2 {
			if !tt.hasError {
				t.Errorf("expected no error for %q but got wrong parts count", tt.input)
			}
			continue
		}
		w, errW := strconv.Atoi(parts[0])
		h, errH := strconv.Atoi(parts[1])
		if (errW != nil || errH != nil) != tt.hasError {
			t.Errorf("input=%q: expected hasError=%v, got errW=%v errH=%v", tt.input, tt.hasError, errW, errH)
		}
		if !tt.hasError && (w != tt.w || h != tt.h) {
			t.Errorf("input=%q: expected %dx%d, got %dx%d", tt.input, tt.w, tt.h, w, h)
		}
	}
}

