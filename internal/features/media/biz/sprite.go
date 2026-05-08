package biz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/conf"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
	"origadmin/application/origcms/internal/helpers/ffmpeg"
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
		Position:   0.2,
	}
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
}

func NewSpriteUseCase(
	mediaRepo MediaRepo,
	configProvider systembiz.ConfigProvider,
	paths *conf.StoragePaths,
	logger log.Logger,
) *SpriteUseCase {
	return &SpriteUseCase{
		mediaRepo:      mediaRepo,
		configProvider: configProvider,
		logger:         log.NewHelper(log.With(logger, "module", "media.sprite")),
		paths:          paths,
	}
}

func (uc *SpriteUseCase) GenerateSpriteAndVTT(ctx context.Context, mediaID string) error {
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
	fullSpritePath := uc.paths.SpriteImageAbsPath(mediaID)
	fullVttPath := uc.paths.SpriteVTTAbsPath(mediaID)

	frameCount, err := ffmpeg.GenerateSpriteSheet(ctx, fullInputPath, fullSpritePath, cfg.FrameInterval, cfg.FrameWidth, cfg.FrameHeight, cfg.Columns)
	if err != nil {
		uc.mediaRepo.UpdateSpriteFields(ctx, mediaID, "failed", "", "")
		return fmt.Errorf("generate sprite sheet for media %s: %w", mediaID, err)
	}

	duration, _ := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)
	spriteImageRef := "sprite.jpg"

	if err := ffmpeg.GenerateWebVTT(fullVttPath, spriteImageRef, frameCount, float64(cfg.FrameInterval), cfg.Columns, cfg.FrameWidth, cfg.FrameHeight, duration); err != nil {
		uc.mediaRepo.UpdateSpriteFields(ctx, mediaID, "failed", "", "")
		return fmt.Errorf("generate webvtt for media %s: %w", mediaID, err)
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

	if timestamp <= 0 {
		duration, err := ffmpeg.GetVideoDurationSeconds(ctx, uc.paths.FullPath(m.Url))
		if err != nil {
			return fmt.Errorf("get video duration for media %s: %w", mediaID, err)
		}
		timestamp = duration * cfg.Position
	}

	thumbPath := uc.paths.RelativeThumbnail(mediaID)
	fullInputPath := uc.paths.FullPath(m.Url)
	fullThumbPath := uc.paths.ThumbnailAbsPath(mediaID)

	duration, _ := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)
	position := timestamp / duration
	if position > 1.0 {
		position = cfg.Position
	}

	actualTimestamp, err := ffmpeg.ExtractThumbnailAtPosition(ctx, fullInputPath, fullThumbPath, duration, position, cfg.Quality, cfg.Resolution)
	if err != nil {
		return fmt.Errorf("extract thumbnail for media %s: %w", mediaID, err)
	}

	if err := uc.mediaRepo.UpdateThumbnailFields(ctx, mediaID, thumbPath, actualTimestamp); err != nil {
		uc.logger.Warnf("failed to update thumbnail fields for media %s: %v", mediaID, err)
	}

	uc.logger.Infof("thumbnail regenerated for media %s: path=%s timestamp=%.3f", mediaID, thumbPath, actualTimestamp)
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
	fullGifPath := uc.paths.PreviewAbsPath(mediaID)

	duration, _ := ffmpeg.GetVideoDurationSeconds(ctx, fullInputPath)

	generated, err := ffmpeg.GenerateGIFPreviewConditional(ctx, fullInputPath, fullGifPath, duration, 0, cfg.MaxDuration, cfg.Fps, 320)
	if err != nil {
		return fmt.Errorf("generate GIF preview for media %s: %w", mediaID, err)
	}

	if generated {
		if err := uc.mediaRepo.UpdatePreviewFilePath(ctx, mediaID, gifPath); err != nil {
			uc.logger.Warnf("failed to update preview path for media %s: %v", mediaID, err)
		}
		uc.logger.Infof("GIF preview generated for media %s: %s", mediaID, gifPath)
	}

	return nil
}
