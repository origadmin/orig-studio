package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

// ProgressCallback is a callback function for reporting transcoding progress.
// progress: 0-100 percentage
// frame: current frame number
// fps: frames per second
// speed: transcoding speed (e.g., "1.5x")
// time: current time in seconds
type ProgressCallback func(progress int, frame int64, fps float64, speed string, time float64)

func init() {
	// Default search path
	defaultToolsDir := `tools\bin`
	// If on Windows and tools\bin exists in current directory
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
	} else if videoCodec == "vp9" {
		vcodec = "libvpx-vp9"
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
	} else if videoCodec == "vp9" {
		vcodec = "libvpx-vp9"
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

// TranscodeToHLSWithProgress transcodes the input file to HLS with real-time progress reporting.
// This is an enhanced version of TranscodeToHLS that provides progress updates via callback.
func TranscodeToHLSWithProgress(
	ctx context.Context,
	inputPath string,
	outputDir string,
	profileName string,
	resolution string,
	videoCodec string,
	audioCodec string,
	videoBitrate string,
	audioBitrate string,
	duration float64,
	progressCb ProgressCallback,
) error {
	// Clean up existing files in output directory to avoid segment residue
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}
	
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create HLS output directory: %w", err)
	}

	size := ResolutionToSize(resolution)

	vcodec := "libx264"
	if videoCodec == "h265" || videoCodec == "hevc" {
		vcodec = "libx265"
	} else if videoCodec == "vp9" {
		vcodec = "libvpx-vp9"
	}

	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, "index.m3u8")

	args := []string{
		"-progress", "pipe:1",
		"-i", inputPath,
		"-s", size,
		"-c:v", vcodec,
		"-c:a", "aac",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
	}

	if videoBitrate != "" && videoBitrate != "auto" {
		args = append(args, "-b:v", videoBitrate)
	}
	if audioBitrate != "" {
		args = append(args, "-b:a", audioBitrate)
	}

	args = append(args, "-y", playlistPath)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		parseProgressOutput(stdout, duration, progressCb)
	}()

	errOutput, _ := io.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg direct HLS transcoding failed for profile %s: %w, output: %s", profileName, err, string(errOutput))
	}

	<-progressDone

	return nil
}

// parseProgressOutput parses ffmpeg progress output from the reader.
// ffmpeg progress format:
//
//	frame=123
//	fps=30.5
//	total_time=4.05
//	out_time_ms=4050000
//	speed=1.23x
//	progress=continue
//	(empty line)
func parseProgressOutput(reader io.Reader, duration float64, progressCb ProgressCallback) {
	if progressCb == nil {
		return
	}

	scanner := bufio.NewScanner(reader)
	progressData := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			if len(progressData) > 0 {
				frame := parseInt64(progressData["frame"])
				fps := parseFloat64(progressData["fps"])
				speed := progressData["speed"]
				outTimeMs := parseInt64(progressData["out_time_ms"])
				time := float64(outTimeMs) / 1000000.0

				var progress int
				if duration > 0 {
					progress = int((time / duration) * 100)
					if progress > 100 {
						progress = 100
					}
				} else {
					progress = 0
				}

				progressCb(progress, frame, fps, speed, time)

				progressData = make(map[string]string)
			}
		} else {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				progressData[parts[0]] = parts[1]
			}
		}
	}
}

func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

func parseFloat64(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
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
		"-vf", fmt.Sprintf("fps=5,scale=%s:flags=lanczos,palettegen", scale),
		"-t", "3", // first 3 seconds only for faster generation
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
		"-lavfi", fmt.Sprintf("fps=5,scale=%s:flags=lanczos [x]; [x][1:v] paletteuse", scale),
		"-t", "3",
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


