package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	ffmpegPath  = "ffmpeg"
	ffprobePath = "ffprobe"
)

func init() {
	// 默认探索路径
	defaultToolsDir := `tools\bin`
	// 如果在 Windows 下且当前目录下有 tools\bin
	if _, err := os.Stat(filepath.Join(defaultToolsDir, "ffmpeg.exe")); err == nil {
		ffmpegPath, _ = filepath.Abs(filepath.Join(defaultToolsDir, "ffmpeg.exe"))
		ffprobePath, _ = filepath.Abs(filepath.Join(defaultToolsDir, "ffprobe.exe"))
	}
}

// SetFFmpegPath sets the path to the ffmpeg executable.
func SetFFmpegPath(path string) {
	ffmpegPath = path
}

// SetFFprobePath sets the path to the ffprobe executable.
func SetFFprobePath(path string) {
	ffprobePath = path
}

// ExtractThumbnail extracts a frame from the video at the given timestamp and saves it as an image.
func ExtractThumbnail(ctx context.Context, inputPath, outputPath string, timestamp string) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// ffmpeg -ss [timestamp] -i [input] -vframes 1 -q:v 2 [output]
	// -ss before -i is faster because it seeks to the nearest keyframe.
	args := []string{
		"-ss", timestamp,
		"-i", inputPath,
		"-vframes", "1",
		"-q:v", "2", // quality 2-31, lower is better
		"-y", // overwrite output
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail extraction failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GetVideoDuration returns the duration of the video using ffprobe.
func GetVideoDuration(ctx context.Context, inputPath string) (time.Duration, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(output))
	}

	durationStr := strings.TrimSpace(string(output))
	var seconds float64
	_, err = fmt.Sscanf(durationStr, "%f", &seconds)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// TranscodeToMP4 transcodes the input file to a standard MP4 file with specific resolution and codec.
func TranscodeToMP4(ctx context.Context, inputPath, outputPath string, resolution string, videoCodec string, audioCodec string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Basic ffmpeg command for standard MP4 output
	// ffmpeg -i [input] -s [resolution] -c:v [vcodec] -c:a [acodec] -f mp4 -y [output]
	args := []string{
		"-i", inputPath,
		"-s", resolution,
		"-c:v", "libx264", // default to libx264 for now, could be mapped from videoCodec
		"-c:a", "aac",
		"-movflags", "faststart+frag_keyframe+empty_moov", // optimized for streaming/dash/hls inputs
		"-y",
		outputPath,
	}

	// Simple codec mapping if needed
	if videoCodec == "h265" || videoCodec == "hevc" {
		args[5] = "libx265"
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcoding to MP4 failed: %w, output: %s", err, string(output))
	}

	return nil
}
