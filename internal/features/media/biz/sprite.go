package biz

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
	"origadmin/application/origstudio/internal/features/media/ffmpeg"
)

type SpriteConfig struct {
	FrameInterval int `json:"frame_interval"`
	Columns       int `json:"columns"`
	FrameWidth    int `json:"frame_width"`
	FrameHeight   int `json:"frame_height"`
	MaxFrames     int `json:"max_frames"`
}

type ThumbnailConfig struct {
	Quality    int     `json:"quality"`
	Resolution string  `json:"resolution"`
	Position   float64 `json:"position"`
}

func DefaultSpriteConfig() SpriteConfig {
	return SpriteConfig{
		FrameInterval: 10,
		Columns:       5,
		FrameWidth:    160,
		FrameHeight:   90,
		MaxFrames:     100,
	}
}

func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		Quality:    2,
		Resolution: "1280x720",
		Position:   0.05,
	}
}

func parseResolution(resolution string) (width, height int) {
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	w, err1 := strconv.Atoi(parts[0])
	h, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return w, h
}

func LoadSpriteConfig(ctx context.Context, cp systembiz.ConfigProvider) SpriteConfig {
	cfg := DefaultSpriteConfig()
	if v := cp.GetInt(ctx, "sprite_frame_interval"); v > 0 {
		cfg.FrameInterval = v
	}
	if v := cp.GetInt(ctx, "sprite_columns"); v > 0 {
		cfg.Columns = v
	}
	if v := cp.GetInt(ctx, "sprite_frame_width"); v > 0 {
		cfg.FrameWidth = v
	}
	if v := cp.GetInt(ctx, "sprite_frame_height"); v > 0 {
		cfg.FrameHeight = v
	}
	if v := cp.GetInt(ctx, "sprite_max_frames"); v > 0 {
		cfg.MaxFrames = v
	}
	return cfg
}

func LoadThumbnailConfig(ctx context.Context, cp systembiz.ConfigProvider) ThumbnailConfig {
	cfg := DefaultThumbnailConfig()
	if v := cp.GetInt(ctx, "thumbnail_quality"); v > 0 {
		cfg.Quality = v
	}
	if v := cp.Get(ctx, "thumbnail_resolution"); v != "" {
		cfg.Resolution = v
	}
	if v := cp.Get(ctx, "thumbnail_position"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 && f <= 1.0 {
			cfg.Position = f
		}
	}
	return cfg
}

type SpriteUseCase struct {
	mediaRepo      MediaRepo
	configProvider systembiz.ConfigProvider
	logger         *log.Helper
	paths          *conf.StoragePaths
	storage        Storage
}

func NewSpriteUseCase(
	mediaRepo MediaRepo,
	configProvider systembiz.ConfigProvider,
	paths *conf.StoragePaths,
	storage Storage,
	logger log.Logger,
) *SpriteUseCase {
	return &SpriteUseCase{
		mediaRepo:      mediaRepo,
		configProvider: configProvider,
		logger:         log.NewHelper(log.With(logger, "module", "media.sprite")),
		paths:          paths,
		storage:        storage,
	}
}

func (uc *SpriteUseCase) GenerateSpriteAndVTT(ctx context.Context, mediaID string) error {
	defer func() {
		if r := recover(); r != nil {
			uc.mediaRepo.UpdateSpriteFields(context.Background(), mediaID, "failed", "", "")
			uc.logger.Errorf("GenerateSpriteAndVTT panicked for media %s: %v", mediaID, r)
		}
	}()

	m, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}

	if !strings.HasPrefix(m.MimeType, "video/") {
		return fmt.Errorf("media %s is not a video (type: %s)", mediaID, m.MimeType)
	}

	if err := uc.mediaRepo.UpdateSpriteFields(ctx, mediaID, "processing", "", ""); err != nil {
		uc.logger.Warnf("failed to set sprite_status to processing for media %s: %v", mediaID, err)
	}

	cfg := LoadSpriteConfig(ctx, uc.configProvider)

	spritePath := uc.paths.RelativeSpriteImage(mediaID)
	vttPath := uc.paths.RelativeSpriteVTT(mediaID)

	fullInputPath := uc.paths.FullPath(m.Url)
	var workDir *LocalWorkDir
	if _, err := os.Stat(fullInputPath); err != nil {
		workDir, err = DownloadToLocalWorkDir(ctx, uc.storage, m.Url, uc.paths.FullPath)
		if err != nil {
			return fmt.Errorf("download source for sprite: %w", err)
		}
		if workDir != nil {
			fullInputPath = workDir.LocalPath
		}
	}
	defer workDir.Cleanup()

	fullSpritePath := uc.paths.SpriteImageAbsPath(mediaID)
	fullVttPath := uc.paths.SpriteVTTAbsPath(mediaID)

	frameW, frameH := cfg.FrameWidth, cfg.FrameHeight
	vw, vh, probeErr := ffmpeg.GetVideoDisplaySize(ctx, fullInputPath)
	if probeErr != nil {
		uc.logger.Warnf("failed to detect video size for %s, using default dimensions: %v", mediaID, probeErr)
	} else if vh > vw {
		frameW, frameH = cfg.FrameHeight, cfg.FrameWidth
	}

	frameCount, _, _, tileCols, err := ffmpeg.GenerateSpriteSheet(ctx, fullInputPath, fullSpritePath, cfg.FrameInterval, frameW, frameH, cfg.Columns)
	if err != nil {
		uc.mediaRepo.UpdateSpriteFields(ctx, mediaID, "failed", "", "")
		return fmt.Errorf("generate sprite sheet for media %s: %w", mediaID, err)
	}

	duration, _ := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)
	spriteImageRef := "sprite.jpg"

	if err := ffmpeg.GenerateWebVTT(fullVttPath, spriteImageRef, frameCount, float64(cfg.FrameInterval), tileCols, frameW, frameH, duration); err != nil {
		uc.mediaRepo.UpdateSpriteFields(ctx, mediaID, "failed", "", "")
		return fmt.Errorf("generate webvtt for media %s: %w", mediaID, err)
	}

	// Always sync the locally-generated sprite sheet + VTT to storage (S3) so
	// they are served consistently across nodes and via CDN / multi-node
	// deployments. Previously this block was guarded by `if workDir != nil`,
	// which skipped the upload whenever the source video was already local -
	// leaving the sprite only on the local volume and breaking multi-node / CDN
	// setups. This is the same root cause already fixed for thumbnails in
	// RegenerateThumbnail below. The sprite files are written to spriteDir above
	// (GenerateSpriteSheet / GenerateWebVTT) regardless of whether a download
	// was needed, so unconditional upload is safe.
	spriteDir := uc.paths.SpriteDirAbs(mediaID)
	if _, err := os.Stat(spriteDir); err == nil {
		if err := uc.storage.UploadDir(ctx, spriteDir, uc.paths.Relative("sprites", mediaID)); err != nil {
			uc.logger.Warnf("failed to upload sprite dir to S3 for media %s: %v", mediaID, err)
		}
	}

	if err := uc.mediaRepo.UpdateSpriteFields(ctx, mediaID, "success", spritePath, vttPath); err != nil {
		uc.logger.Warnf("failed to update sprite fields for media %s: %v", mediaID, err)
	}

	uc.logger.Infof("sprite and VTT generated for media %s: sprite=%s vtt=%s frames=%d", mediaID, spritePath, vttPath, frameCount)
	return nil
}

func (uc *SpriteUseCase) RegenerateThumbnail(ctx context.Context, mediaID string, timestamp float64) error {
	m, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}

	cfg := LoadThumbnailConfig(ctx, uc.configProvider)

	fullInputPath := uc.paths.FullPath(m.Url)
	var workDir *LocalWorkDir
	if _, err := os.Stat(fullInputPath); err != nil {
		workDir, err = DownloadToLocalWorkDir(ctx, uc.storage, m.Url, uc.paths.FullPath)
		if err != nil {
			return fmt.Errorf("download source for thumbnail: %w", err)
		}
		if workDir != nil {
			fullInputPath = workDir.LocalPath
		}
	}
	defer workDir.Cleanup()

	if timestamp <= 0 {
		duration, err := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)
		if err != nil {
			return fmt.Errorf("get video duration for media %s: %w", mediaID, err)
		}
		timestamp = duration * cfg.Position
	}

	thumbPath := uc.paths.RelativeThumbnail(mediaID)
	fullThumbPath := uc.paths.ThumbnailAbsPath(mediaID)

	duration, _ := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)
	position := timestamp / duration
	if position > 1.0 {
		position = cfg.Position
	}

	thumbResolution := cfg.Resolution
	vw, vh, probeErr := ffmpeg.GetVideoDisplaySize(ctx, fullInputPath)
	if probeErr != nil {
		uc.logger.Warnf("failed to detect video size for thumbnail %s, using default resolution: %v", mediaID, probeErr)
	} else if vh > vw {
		w, h := parseResolution(cfg.Resolution)
		if w > 0 && h > 0 {
			thumbResolution = fmt.Sprintf("%dx%d", h, w)
		}
	}

	actualTimestamp, err := ffmpeg.ExtractThumbnailAtPosition(ctx, fullInputPath, fullThumbPath, duration, position, cfg.Quality, thumbResolution)
	if err != nil {
		return fmt.Errorf("extract thumbnail for media %s: %w", mediaID, err)
	}

	// Always sync the locally-extracted thumbnail to storage (S3) so it is
	// served consistently across nodes and via CDN. Previously this block was
	// guarded by `if workDir != nil`, which skipped the upload whenever the
	// source video was already local - leaving the regenerated thumbnail only
	// on the local volume and breaking multi-node / CDN setups. The thumbnail
	// file (fullThumbPath) is already written above (line 253) regardless of
	// whether a download was needed, so unconditional upload is safe.
	relKey := uc.paths.RelativeThumbnail(mediaID)
	data, fErr := os.ReadFile(fullThumbPath)
	if fErr == nil {
		if _, err := uc.storage.Upload(ctx, relKey, bytes.NewReader(data), int64(len(data)), ""); err != nil {
			uc.logger.Warnf("failed to upload thumbnail %s to storage: %v", relKey, err)
		}
	}

	if err := uc.mediaRepo.UpdateThumbnailFields(ctx, mediaID, thumbPath, actualTimestamp); err != nil {
		uc.logger.Warnf("failed to update thumbnail fields for media %s: %v", mediaID, err)
	}

	uc.logger.Infof("thumbnail regenerated for media %s: path=%s timestamp=%.3f", mediaID, thumbPath, actualTimestamp)
	return nil
}

func (uc *SpriteUseCase) SetCustomThumbnail(ctx context.Context, mediaID string, imagePath string) error {
	_, err := os.Stat(imagePath)
	if err != nil {
		return fmt.Errorf("custom thumbnail image not found: %w", err)
	}

	cfg := LoadThumbnailConfig(ctx, uc.configProvider)
	thumbPath := uc.paths.RelativeThumbnail(mediaID)
	fullThumbPath := uc.paths.ThumbnailAbsPath(mediaID)

	if err := ffmpeg.ProcessImageToThumbnail(ctx, imagePath, fullThumbPath, cfg.Quality, cfg.Resolution); err != nil {
		return fmt.Errorf("process custom thumbnail for media %s: %w", mediaID, err)
	}

	data, fErr := os.ReadFile(fullThumbPath)
	if fErr == nil {
		relKey := uc.paths.RelativeThumbnail(mediaID)
		if _, err := uc.storage.Upload(ctx, relKey, bytes.NewReader(data), int64(len(data)), ""); err != nil {
			uc.logger.Warnf("failed to upload custom thumbnail %s to storage: %v", relKey, err)
		}
	}

	if err := uc.mediaRepo.UpdateThumbnailFields(ctx, mediaID, thumbPath, 0); err != nil {
		uc.logger.Warnf("failed to update thumbnail fields for media %s: %v", mediaID, err)
	}

	uc.logger.Infof("custom thumbnail set for media %s: path=%s", mediaID, thumbPath)
	return nil
}

// SetSpriteSheetThumbnail sets the media COVER to the WHOLE generated sprite
// sheet image (per product requirement: "整体雪碧图应作为第一张图" — the entire
// sprite strip is the first/cover image, NOT a single sampled frame).
//
// The previously-shipped behaviour re-sampled a video frame when the whole
// sprite sheet was chosen, which produced a single-frame thumbnail that looked
// identical to an existing frame ("选择整体雪碧图没有效果" / "整体雪碧图的逻辑
// 完全错误"). This method instead points the cover directly at the sprite sheet
// asset (m.SpritePath), so the cover becomes the full thumbnail strip.
func (uc *SpriteUseCase) SetSpriteSheetThumbnail(ctx context.Context, mediaID string) error {
	m, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}
	if m.SpritePath == "" {
		return fmt.Errorf("sprite sheet not available for media %s (status=%s)", mediaID, m.SpriteStatus)
	}
	if err := uc.mediaRepo.UpdateThumbnailFields(ctx, mediaID, m.SpritePath, 0); err != nil {
		uc.logger.Warnf("failed to set sprite-sheet cover for media %s: %v", mediaID, err)
	}
	uc.logger.Infof("sprite-sheet cover set for media %s: thumbnail=%s", mediaID, m.SpritePath)
	return nil
}

func (uc *SpriteUseCase) ProcessPostTranscode(ctx context.Context, mediaID string) error {
	m, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}

	if !strings.HasPrefix(m.MimeType, "video/") {
		return nil
	}

	if err := uc.GenerateSpriteAndVTT(ctx, mediaID); err != nil {
		uc.logger.Warnf("sprite generation failed for media %s: %v", mediaID, err)
	}

	if err := uc.GenerateGIFPreview(ctx, mediaID); err != nil {
		uc.logger.Warnf("GIF preview generation failed for media %s: %v", mediaID, err)
	}

	if m.Thumbnail == "" {
		if err := uc.RegenerateThumbnail(ctx, mediaID, 0); err != nil {
			uc.logger.Warnf("thumbnail regeneration failed for media %s: %v", mediaID, err)
		}
	}

	return nil
}

// GifPreviewConfig holds configuration for GIF preview generation.
type GifPreviewConfig struct {
	Fps         int     `json:"fps"`
	Scale       string  `json:"scale"`
	MaxDuration float64 `json:"max_duration"`
}

// DefaultGifPreviewConfig returns the default GIF preview configuration.
func DefaultGifPreviewConfig() GifPreviewConfig {
	return GifPreviewConfig{
		Fps:         5,
		Scale:       "320:-1",
		MaxDuration: 3.0,
	}
}

// LoadGifPreviewConfig loads GIF preview configuration from the config provider.
func LoadGifPreviewConfig(ctx context.Context, cp systembiz.ConfigProvider) GifPreviewConfig {
	cfg := DefaultGifPreviewConfig()
	if v := cp.GetInt(ctx, "gif_preview_fps"); v > 0 {
		cfg.Fps = v
	}
	if v := cp.Get(ctx, "gif_preview_scale"); v != "" {
		cfg.Scale = v
	}
	if v := cp.Get(ctx, "gif_preview_max_duration"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cfg.MaxDuration = f
		}
	}
	return cfg
}

// GenerateGIFPreview generates an animated GIF preview for the given media.
func (uc *SpriteUseCase) GenerateGIFPreview(ctx context.Context, mediaID string) error {
	m, err := uc.mediaRepo.Get(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("get media %s: %w", mediaID, err)
	}

	if !strings.HasPrefix(m.MimeType, "video/") {
		return nil
	}

	cfg := LoadGifPreviewConfig(ctx, uc.configProvider)

	gifPath := uc.paths.RelativePreview(mediaID)
	fullInputPath := uc.paths.FullPath(m.Url)
	var workDir *LocalWorkDir
	if _, err := os.Stat(fullInputPath); err != nil {
		workDir, err = DownloadToLocalWorkDir(ctx, uc.storage, m.Url, uc.paths.FullPath)
		if err != nil {
			return fmt.Errorf("download source for gif preview: %w", err)
		}
		if workDir != nil {
			fullInputPath = workDir.LocalPath
		}
	}
	defer workDir.Cleanup()

	fullGifPath := uc.paths.PreviewAbsPath(mediaID)

	duration, _ := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)

	generated, err := ffmpeg.GenerateGIFPreviewConditional(ctx, fullInputPath, fullGifPath, duration, 0, cfg.MaxDuration, cfg.Fps, 320)
	if err != nil {
		return fmt.Errorf("generate GIF preview for media %s: %w", mediaID, err)
	}

	if generated {
		if workDir != nil {
			relKey := uc.paths.RelativePreview(mediaID)
			// Read into memory first to avoid source/destination path conflict
			data, fErr := os.ReadFile(fullGifPath)
			if fErr == nil {
				if _, err := uc.storage.Upload(ctx, relKey, bytes.NewReader(data), int64(len(data)), ""); err != nil {
					uc.logger.Warnf("failed to upload preview %s to storage: %v", relKey, err)
				}
			}
		}

		if err := uc.mediaRepo.UpdatePreviewFilePath(ctx, mediaID, gifPath); err != nil {
			uc.logger.Warnf("failed to update preview path for media %s: %v", mediaID, err)
		}
		uc.logger.Infof("GIF preview generated for media %s: %s", mediaID, gifPath)
	}

	return nil
}
