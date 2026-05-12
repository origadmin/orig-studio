package ffmpeg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// GetVideoResolution returns the width and height of the video using ffprobe.
func GetVideoResolution(ctx context.Context, inputPath string) (width, height int, err error) {
	args := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe resolution failed: %w, output: %s", err, string(output))
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe resolution output: %s", string(output))
	}

	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse width %q: %w", parts[0], err)
	}

	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse height %q: %w", parts[1], err)
	}

	return w, h, nil
}

// MapCodecName converts short codec names to ffmpeg encoder names.
// Only short names are accepted: h264, h265, hevc, vp9.
func MapCodecName(videoCodec string) string {
	switch videoCodec {
	case "h264":
		return "libx264"
	case "h265", "hevc":
		return "libx265"
	case "vp9":
		return "libvpx-vp9"
	default:
		return "libx264"
	}
}

// MapAudioCodec selects the appropriate audio codec based on the requested codec
// and the video encoder in use. When audioCodec is empty or "-", it defaults to
// aac for H.264/H.265 and libopus for VP9.
func MapAudioCodec(audioCodec string, videoEncoder string) string {
	if audioCodec != "" && audioCodec != "-" {
		switch audioCodec {
		case "aac", "libaac":
			return "aac"
		case "opus", "libopus":
			return "libopus"
		}
	}
	// Default: aac for H.264/H.265, libopus for VP9
	if videoEncoder == "libvpx-vp9" {
		return "libopus"
	}
	return "aac"
}

// levelForResolution returns the appropriate H.264 level for a given resolution.
func levelForResolution(resolution string) string {
	height := resolution
	if strings.Contains(resolution, "x") {
		parts := strings.SplitN(resolution, "x", 2)
		height = parts[1]
	}
	switch height {
	case "240", "360", "480":
		return "3.0"
	case "720":
		return "4.1"
	case "1080":
		return "4.2"
	case "1440":
		return "5.1"
	case "2160":
		return "5.2"
	default:
		return "4.1"
	}
}

// buildVideoArgs constructs ffmpeg encoding arguments based on codec, resolution, and bitrate.
// It produces quality-controlled parameters including proper scaling with aspect ratio
// preservation, CRF, maxrate/bufsize, preset, profile/level, key frame control, and pix_fmt.
func buildVideoArgs(videoCodec string, resolution string, videoBitrate string, audioCodec string, audioBitrate string) []string {
	width, height := ResolutionToWidthHeight(resolution)

	// Scale filter with aspect ratio preservation
	scaleFilter := fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=decrease:force_divisible_by=2:flags=lanczos", width, height)

	args := []string{}

	// Video codec and quality
	vcodec := MapCodecName(videoCodec)
	switch vcodec {
	case "libx264":
		args = append(args,
			"-c:v", "libx264",
			"-filter:v", scaleFilter,
			"-pix_fmt", "yuv420p",
			"-crf", "23",
			"-preset", "medium",
			"-profile:v", "main",
			"-level", levelForResolution(resolution),
			"-force_key_frames", "expr:gte(t,n_forced*4)",
			"-x264-params", "keyint=240:keyint_min=120",
		)
	case "libx265":
		args = append(args,
			"-c:v", "libx265",
			"-filter:v", scaleFilter,
			"-pix_fmt", "yuv420p",
			"-crf", "28",
			"-preset", "medium",
			"-profile:v", "main",
			"-level", levelForResolution(resolution),
			"-force_key_frames", "expr:gte(t,n_forced*4)",
			"-x265-params", "keyint=240:keyint_min=120",
		)
	case "libvpx-vp9":
		args = append(args,
			"-c:v", "libvpx-vp9",
			"-filter:v", scaleFilter,
			"-crf", "31",
			"-b:v", "0",
			"-quality", "good",
			"-cpu-used", "2",
		)
	}

	// Maxrate/Bufsize: use video_bitrate from profile if specified
	if videoBitrate != "" && videoBitrate != "auto" {
		args = append(args, "-maxrate", videoBitrate, "-bufsize", videoBitrate)
	}

	// Audio codec
	acodec := MapAudioCodec(audioCodec, vcodec)
	args = append(args, "-c:a", acodec)
	if audioBitrate != "" {
		args = append(args, "-b:a", audioBitrate)
	}

	return args
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

	args := []string{"-i", inputPath}
	args = append(args, buildVideoArgs(videoCodec, resolution, videoBitrate, audioCodec, audioBitrate)...)
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

	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, "index.m3u8")

	args := []string{"-i", inputPath}
	args = append(args, buildVideoArgs(videoCodec, resolution, videoBitrate, audioCodec, audioBitrate)...)
	args = append(args,
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		"-y", playlistPath,
	)

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

	segmentPattern := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, "index.m3u8")

	args := []string{
		"-progress", "pipe:1",
		"-i", inputPath,
	}
	args = append(args, buildVideoArgs(videoCodec, resolution, videoBitrate, audioCodec, audioBitrate)...)
	args = append(args,
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", segmentPattern,
		"-y", playlistPath,
	)

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

// PreviewHLSCommand returns the full ffmpeg command string for HLS transcoding preview.
// This does not execute the command, only generates it for display.
func PreviewHLSCommand(inputPath, outputDir, profileName, resolution, videoCodec, audioCodec, videoBitrate, audioBitrate string) string {
	args := []string{ffmpegPath}
	args = append(args, "-i", inputPath)
	args = append(args, buildVideoArgs(videoCodec, resolution, videoBitrate, audioCodec, audioBitrate)...)
	args = append(args,
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(outputDir, "segment_%03d.ts"),
		"-y", filepath.Join(outputDir, "index.m3u8"),
	)
	return strings.Join(args, " ")
}

// PreviewGIFCommand returns the full ffmpeg command string for GIF preview generation.
// This does not execute the command, only generates it for display.
func PreviewGIFCommand(inputPath, outputPath, scale string) string {
	palettePath := outputPath + ".png"
	paletteCmd := strings.Join([]string{
		ffmpegPath, "-i", inputPath,
		"-vf", fmt.Sprintf("fps=5,scale=%s:flags=lanczos,palettegen", scale),
		"-t", "3", "-y", palettePath,
	}, " ")
	gifCmd := strings.Join([]string{
		ffmpegPath, "-i", inputPath, "-i", palettePath,
		"-lavfi", fmt.Sprintf("fps=5,scale=%s:flags=lanczos[x];[x][1:v]paletteuse", scale),
		"-t", "3", "-y", outputPath,
	}, " ")
	return paletteCmd + " && " + gifCmd
}

// MediaInfo holds comprehensive media file information extracted via ffprobe.
type MediaInfo struct {
	Duration     float64
	Size         int64
	BitRate      int64
	FormatName   string
	Width        int
	Height       int
	VideoCodec   string
	VideoBitRate int64
	FPS          float64
	AudioCodec   string
	AudioBitRate int64
	Channels     int
	SampleRate   int
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType  string `json:"codec_type"`
	CodecName  string `json:"codec_name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	RFrameRate string `json:"r_frame_rate"`
	BitRate    string `json:"bit_rate"`
	Channels   int    `json:"channels"`
	SampleRate string `json:"sample_rate"`
}

type ffprobeFormat struct {
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
	FormatName string `json:"format_name"`
}

// GetMediaInfo extracts comprehensive media information using ffprobe.
func GetMediaInfo(ctx context.Context, inputPath string) (*MediaInfo, error) {
	args := []string{
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	}

	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe media info failed: %w, output: %s", err, string(output))
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("parse ffprobe JSON: %w", err)
	}

	info := &MediaInfo{}

	info.Duration, _ = strconv.ParseFloat(probe.Format.Duration, 64)
	info.Size, _ = strconv.ParseInt(probe.Format.Size, 10, 64)
	info.BitRate, _ = strconv.ParseInt(probe.Format.BitRate, 10, 64)
	info.FormatName = probe.Format.FormatName

	for _, s := range probe.Streams {
		switch s.CodecType {
		case "video":
			info.Width = s.Width
			info.Height = s.Height
			info.VideoCodec = s.CodecName
			info.VideoBitRate, _ = strconv.ParseInt(s.BitRate, 10, 64)
			info.FPS = parseRFrameRate(s.RFrameRate)
		case "audio":
			info.AudioCodec = s.CodecName
			info.AudioBitRate, _ = strconv.ParseInt(s.BitRate, 10, 64)
			info.Channels = s.Channels
			info.SampleRate, _ = strconv.Atoi(s.SampleRate)
		}
	}

	return info, nil
}

func parseRFrameRate(rate string) float64 {
	if rate == "" {
		return 0
	}
	parts := strings.SplitN(rate, "/", 2)
	if len(parts) == 2 {
		num, _ := strconv.ParseFloat(parts[0], 64)
		den, _ := strconv.ParseFloat(parts[1], 64)
		if den > 0 {
			return num / den
		}
	}
	fps, _ := strconv.ParseFloat(rate, 64)
	return fps
}


