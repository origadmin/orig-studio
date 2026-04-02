package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TranscodeToHLS converts the input video to an HLS master playlist and segments.
// It uses a standard set of parameters for web playback.
func TranscodeToHLS(ctx context.Context, inputPath, outputDir string) (string, error) {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create HLS output directory: %w", err)
	}

	masterPlaylist := "master.m3u8"
	outputPath := filepath.Join(outputDir, masterPlaylist)

	// ffmpeg -i [input] -profile:v baseline -level 3.0 -s 1280x720 -start_number 0 -hls_time 10 -hls_list_size 0 -f hls [output]
	// Using a single quality (720p) for now for simplicity.
	args := []string{
		"-i", inputPath,
		"-profile:v", "main",
		"-level", "4.0",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-ar", "44100",
		"-b:a", "128k",
		"-s", "1280x720", // 720p
		"-start_number", "0",
		"-hls_time", "6", // 6 second segments
		"-hls_list_size", "0", // list all segments in playlist
		"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
		"-f", "hls",
		"-y", // overwrite output
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg HLS transcoding failed: %w, output: %s", err, string(output))
	}

	return masterPlaylist, nil
}
