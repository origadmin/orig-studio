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
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
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
// DEPRECATED: Use TranscodeToHLS instead for direct HLS output (no intermediate MP4 needed).
// Kept for backward compatibility with non-HLS use cases.
func TranscodeToMP4(
	ctx context.Context,
	inputPath, outputPath string,
	resolution string,
	videoCodec string,
	audioCodec string,
	videoBitrate string,
	audioBitrate string,
) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	size := ResolutionToSize(resolution)

	// Select video codec
	vcodec := "libx264"
	if videoCodec == "h265" || videoCodec == "hevc" {
		vcodec = "libx265"
	}

	args := []string{
		"-i", inputPath,
		"-s", size,
		"-c:v", vcodec,
	}

	// Apply video bitrate if specified and not "auto"
	if videoBitrate != "" && videoBitrate != "auto" {
		args = append(args, "-b:v", videoBitrate)
	}

	// Audio codec: always aac for MP4 compatibility
	args = append(args, "-c:a", "aac")

	// Apply audio bitrate if specified
	if audioBitrate != "" {
		args = append(args, "-b:a", audioBitrate)
	}

	args = append(args,
		"-movflags", "faststart+frag_keyframe+empty_moov",
		"-y",
		outputPath,
	)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcoding to MP4 failed: %w, output: %s", err, string(output))
	}

	return nil
}

// TranscodeToHLS transcodes the input file directly to HLS segments in one pass.
// No intermediate MP4 is created — this is more efficient than TranscodeToMP4 + MP4HLS.
//
// Output structure:
//
//	{outputDir}/          (e.g., hls/{uuid}/{profile_name}/)
//	  index.m3u8          (variant playlist)
//	  segment_001.ts
//	  segment_002.ts
//	  ...
func TranscodeToHLS(
	ctx context.Context,
	inputPath string,
	outputDir string,
	profileName string,
	resolution string,
	videoCodec string,
	audioCodec string,
	videoBitrate string,
	audioBitrate string,
) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create HLS output directory: %w", err)
	}

	size := ResolutionToSize(resolution)

	// Select video codec
	vcodec := "libx264"
	if videoCodec == "h265" || videoCodec == "hevc" {
		vcodec = "libx265"
	}

	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, "index.m3u8")

	args := []string{
		"-i", inputPath,
		"-s", size,
		"-c:v", vcodec,
		"-c:a", "aac",
		// HLS muxer settings
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
	}

	// Apply bitrates if specified
	if videoBitrate != "" && videoBitrate != "auto" {
		args = append(args, "-b:v", videoBitrate)
	}
	if audioBitrate != "" {
		args = append(args, "-b:a", audioBitrate)
	}

	args = append(args, "-y", playlistPath)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg direct HLS transcoding failed for profile %s: %w, output: %s", profileName, err, string(output))
	}

	return nil
}

// GenerateGIFPreview creates an animated GIF preview from the source video.
// Used for hover thumbnails on progress bars and media list previews.
// Output is written to outputPath (e.g., previews/{uuid}.gif).
func GenerateGIFPreview(ctx context.Context, inputPath, outputPath string, scale string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create preview output directory: %w", err)
	}

	// ffmpeg palette generation for high quality GIF
	palettePath := outputPath + ".png"

	// Step 1: Generate color palette
	paletteArgs := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf("fps=10,scale=%s:flags=lanczos,palettegen", scale),
		"-t", "5", // first 5 seconds only
		"-y",
		palettePath,
	}
	paletteCmd := exec.CommandContext(ctx, ffmpegPath, paletteArgs...)
	if output, err := paletteCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg palette generation failed: %w, output: %s", err, string(output))
	}

	// Step 2: Use palette to create GIF
	gifArgs := []string{
		"-i", inputPath,
		"-i", palettePath,
		"-lavfi", fmt.Sprintf("fps=10,scale=%s:flags=lanczos [x]; [x][1:v] paletteuse", scale),
		"-t", "5",
		"-y",
		outputPath,
	}
	gifCmd := exec.CommandContext(ctx, ffmpegPath, gifArgs...)
	if output, err := gifCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg GIF preview failed: %w, output: %s", err, string(output))
	}

	// Clean up temporary palette file
	os.Remove(palettePath)

	return nil
}
