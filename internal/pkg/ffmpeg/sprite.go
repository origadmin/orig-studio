package ffmpeg

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func GetVideoDurationSeconds(ctx context.Context, inputPath string) (float64, error) {
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
	seconds, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return seconds, nil
}

func GenerateSpriteSheet(ctx context.Context, inputPath string, outputPath string, frameInterval int, frameWidth int, frameHeight int, columns int) (frameCount int, err error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, fmt.Errorf("failed to create sprite output directory: %w", err)
	}

	os.Remove(outputPath)

	duration, err := GetVideoDurationSeconds(ctx, inputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration: %w", err)
	}

	totalFrames := int(duration/float64(frameInterval)) + 1
	if totalFrames > 1000 {
		frameInterval = int(duration/1000) + 1
		totalFrames = int(duration/float64(frameInterval)) + 1
	}

	rows := (totalFrames + columns - 1) / columns

	tileFilter := fmt.Sprintf("fps=1/%d,scale=%d:%d,tile=%dx%d", frameInterval, frameWidth, frameHeight, columns, rows)
	args := []string{
		"-i", inputPath,
		"-vf", tileFilter,
		"-frames:v", "1",
		"-q:v", "3",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return totalFrames, nil
	}

	tmpDir, tmpErr := os.MkdirTemp("", "sprite_frames_*")
	if tmpErr != nil {
		return 0, fmt.Errorf("ffmpeg tile filter failed and temp dir creation failed: tile error: %w, output: %s; temp dir error: %w", err, string(output), tmpErr)
	}
	defer os.RemoveAll(tmpDir)

	framePattern := filepath.Join(tmpDir, "frame_%04d.jpg")
	extractArgs := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf("fps=1/%d,scale=%d:%d", frameInterval, frameWidth, frameHeight),
		"-q:v", "3",
		"-y",
		framePattern,
	}

	extractCmd := exec.CommandContext(ctx, ffmpegPath, extractArgs...)
	extractOutput, extractErr := extractCmd.CombinedOutput()
	if extractErr != nil {
		return 0, fmt.Errorf("ffmpeg frame extraction failed: %w, output: %s", extractErr, string(extractOutput))
	}

	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		return 0, fmt.Errorf("failed to read temp frame directory: %w", readErr)
	}

	var frames []image.Image
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".jpg") {
			continue
		}
		f, openErr := os.Open(filepath.Join(tmpDir, entry.Name()))
		if openErr != nil {
			f.Close()
			continue
		}
		img, decodeErr := jpeg.Decode(f)
		f.Close()
		if decodeErr != nil {
			continue
		}
		frames = append(frames, img)
	}

	if len(frames) == 0 {
		return 0, fmt.Errorf("no frames extracted from video")
	}

	actualFrames := len(frames)
	actualRows := (actualFrames + columns - 1) / columns
	spriteWidth := columns * frameWidth
	spriteHeight := actualRows * frameHeight

	spriteImg := image.NewRGBA(image.Rect(0, 0, spriteWidth, spriteHeight))

	for i, frame := range frames {
		col := i % columns
		row := i / columns
		x := col * frameWidth
		y := row * frameHeight
		for dy := 0; dy < frame.Bounds().Dy() && dy < frameHeight; dy++ {
			for dx := 0; dx < frame.Bounds().Dx() && dx < frameWidth; dx++ {
				spriteImg.Set(x+dx, y+dy, frame.At(frame.Bounds().Min.X+dx, frame.Bounds().Min.Y+dy))
			}
		}
	}

	outFile, createErr := os.Create(outputPath)
	if createErr != nil {
		return 0, fmt.Errorf("failed to create sprite sheet file: %w", createErr)
	}
	defer outFile.Close()

	if encodeErr := jpeg.Encode(outFile, spriteImg, &jpeg.Options{Quality: 85}); encodeErr != nil {
		return 0, fmt.Errorf("failed to encode sprite sheet: %w", encodeErr)
	}

	return actualFrames, nil
}

func GenerateWebVTT(outputPath string, spriteImagePath string, frameCount int, frameInterval float64, columns int, frameWidth int, frameHeight int, videoDuration float64) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create VTT output directory: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")

	for i := 0; i < frameCount; i++ {
		startTime := float64(i) * frameInterval
		endTime := math.Min(float64(i+1)*frameInterval, videoDuration)

		col := i % columns
		row := i / columns
		x := col * frameWidth
		y := row * frameHeight

		sb.WriteString(fmt.Sprintf("%s --> %s\n", formatVTTTime(startTime), formatVTTTime(endTime)))
		sb.WriteString(fmt.Sprintf("%s#xywh=%d,%d,%d,%d\n\n", spriteImagePath, x, y, frameWidth, frameHeight))
	}

	if err := os.WriteFile(outputPath, []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write VTT file: %w", err)
	}

	return nil
}

func formatVTTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}

func ExtractThumbnailAtPosition(ctx context.Context, inputPath string, outputPath string, duration float64, position float64, quality int, resolution string) (timestamp float64, err error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	timestamp = duration * position
	if timestamp < 1.0 {
		timestamp = 1.0
	}

	var width, height int
	parts := strings.Split(resolution, "x")
	if len(parts) == 2 {
		width, _ = strconv.Atoi(parts[0])
		height, _ = strconv.Atoi(parts[1])
	}
	if width == 0 || height == 0 {
		width = 1280
		height = 720
	}

	scaleFilter := fmt.Sprintf("scale=%d:%d", width, height)
	args := []string{
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", inputPath,
		"-vframes", "1",
		"-vf", scaleFilter,
		"-q:v", strconv.Itoa(quality),
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return timestamp, nil
	}

	fallbackTimestamp := 0.0
	fallbackArgs := []string{
		"-ss", fmt.Sprintf("%.3f", fallbackTimestamp),
		"-i", inputPath,
		"-vframes", "1",
		"-vf", scaleFilter,
		"-q:v", strconv.Itoa(quality),
		"-y",
		outputPath,
	}

	fallbackCmd := exec.CommandContext(ctx, ffmpegPath, fallbackArgs...)
	fallbackOutput, fallbackErr := fallbackCmd.CombinedOutput()
	if fallbackErr != nil {
		return 0, fmt.Errorf("ffmpeg thumbnail extraction failed at position %.3f: %w (output: %s); fallback at 0s also failed: %w (output: %s)", timestamp, err, string(output), fallbackErr, string(fallbackOutput))
	}

	return fallbackTimestamp, nil
}

func GenerateGIFPreviewConditional(ctx context.Context, inputPath string, outputPath string, videoDuration float64, threshold float64, maxDuration float64, fps int, width int) (bool, error) {
	if videoDuration >= threshold {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return false, fmt.Errorf("failed to create GIF output directory: %w", err)
	}

	gifDuration := math.Min(videoDuration, maxDuration)

	palettePath := outputPath + ".png"

	paletteArgs := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf("fps=%d,scale=%d:-1:flags=lanczos,palettegen", fps, width),
		"-t", fmt.Sprintf("%.3f", gifDuration),
		"-y",
		palettePath,
	}

	paletteCmd := exec.CommandContext(ctx, ffmpegPath, paletteArgs...)
	paletteOutput, err := paletteCmd.CombinedOutput()
	if err != nil {
		os.Remove(palettePath)
		return false, fmt.Errorf("ffmpeg palette generation failed: %w, output: %s", err, string(paletteOutput))
	}

	gifArgs := []string{
		"-i", inputPath,
		"-i", palettePath,
		"-lavfi", fmt.Sprintf("fps=%d,scale=%d:-1:flags=lanczos[x];[x][1:v]paletteuse", fps, width),
		"-t", fmt.Sprintf("%.3f", gifDuration),
		"-y",
		outputPath,
	}

	gifCmd := exec.CommandContext(ctx, ffmpegPath, gifArgs...)
	gifOutput, err := gifCmd.CombinedOutput()
	os.Remove(palettePath)
	if err != nil {
		return false, fmt.Errorf("ffmpeg GIF preview failed: %w, output: %s", err, string(gifOutput))
	}

	return true, nil
}
