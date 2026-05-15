package biz

import (
	"context"
	"strconv"
	"sync"
	"time"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/setting"
	systemdal "origadmin/application/origstudio/internal/features/system/dal"
)

type ConfigProvider interface {
	Get(ctx context.Context, key string) string
	GetBool(ctx context.Context, key string) bool
	GetInt(ctx context.Context, key string) int
	GetAll(ctx context.Context) map[string]string
}

type cacheEntry struct {
	value     map[string]*entity.Setting
	expiredAt time.Time
}

type SettingUseCase struct {
	repo     *systemdal.SettingRepo
	cache    *cacheEntry
	cacheMu  sync.RWMutex
	cacheTTL time.Duration
}

func NewSettingUseCase(repo *systemdal.SettingRepo) *SettingUseCase {
	return &SettingUseCase{
		repo:     repo,
		cacheTTL: 60 * time.Second,
	}
}

func (uc *SettingUseCase) GetByKey(ctx context.Context, key string) (*entity.Setting, error) {
	return uc.repo.GetByKey(ctx, key)
}

func (uc *SettingUseCase) ListByCategory(
	ctx context.Context,
	category string,
) ([]*entity.Setting, error) {
	return uc.repo.ListByCategory(ctx, category)
}

func (uc *SettingUseCase) ListAll(ctx context.Context) ([]*entity.Setting, error) {
	return uc.repo.ListAll(ctx)
}

func (uc *SettingUseCase) Upsert(ctx context.Context, s *entity.Setting) (*entity.Setting, error) {
	result, err := uc.repo.Upsert(ctx, s)
	if err != nil {
		return nil, err
	}
	uc.InvalidateCache()
	return result, nil
}

func (uc *SettingUseCase) BatchUpsert(ctx context.Context, settings []*entity.Setting) error {
	for _, s := range settings {
		_, err := uc.Upsert(ctx, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (uc *SettingUseCase) Delete(ctx context.Context, id string) error {
	err := uc.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	uc.InvalidateCache()
	return nil
}

func (uc *SettingUseCase) ResetToDefault(ctx context.Context, key string) (*entity.Setting, error) {
	s, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if s.FallbackValue == "" {
		return s, nil
	}
	updated, err := s.Update().
		SetValue(s.FallbackValue).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	uc.InvalidateCache()
	return updated, nil
}

func (uc *SettingUseCase) SeedDefaults(ctx context.Context) error {
	return uc.repo.SeedDefaults(ctx, DefaultSettings())
}

func (uc *SettingUseCase) InvalidateCache() {
	uc.cacheMu.Lock()
	uc.cache = nil
	uc.cacheMu.Unlock()
}

func (uc *SettingUseCase) loadCache(ctx context.Context) (map[string]*entity.Setting, error) {
	uc.cacheMu.RLock()
	if uc.cache != nil && time.Now().Before(uc.cache.expiredAt) {
		val := uc.cache.value
		uc.cacheMu.RUnlock()
		return val, nil
	}
	uc.cacheMu.RUnlock()

	uc.cacheMu.Lock()
	defer uc.cacheMu.Unlock()

	if uc.cache != nil && time.Now().Before(uc.cache.expiredAt) {
		return uc.cache.value, nil
	}

	items, err := uc.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*entity.Setting, len(items))
	for _, item := range items {
		m[item.Key] = item
	}

	uc.cache = &cacheEntry{
		value:     m,
		expiredAt: time.Now().Add(uc.cacheTTL),
	}
	return m, nil
}

func (uc *SettingUseCase) Get(ctx context.Context, key string) string {
	m, err := uc.loadCache(ctx)
	if err != nil {
		return ""
	}
	if s, ok := m[key]; ok {
		return s.Value
	}
	return ""
}

func (uc *SettingUseCase) GetBool(ctx context.Context, key string) bool {
	val := uc.Get(ctx, key)
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return b
}

func (uc *SettingUseCase) GetInt(ctx context.Context, key string) int {
	val := uc.Get(ctx, key)
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return n
}

func (uc *SettingUseCase) GetAll(ctx context.Context) map[string]string {
	m, err := uc.loadCache(ctx)
	if err != nil {
		return map[string]string{}
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v.Value
	}
	return result
}

func (uc *SettingUseCase) GetPublicSettings(ctx context.Context) map[string]string {
	publicKeys := map[string]bool{
		"site_name":          true,
		"site_description":   true,
		"primary_url":        true,
		"allow_registration": true,
		"allow_upload":       true,
		"module_articles":    true,
		"module_videos":      true,
		"module_music":       true,
		"homepage_layout":    true,
	}
	m, err := uc.loadCache(ctx)
	if err != nil {
		return map[string]string{}
	}
	result := make(map[string]string)
	for k, v := range m {
		if publicKeys[k] && !v.IsSensitive {
			result[k] = v.Value
		}
	}
	return result
}

func (uc *SettingUseCase) MaskSensitive(s *entity.Setting) *entity.Setting {
	if s.IsSensitive {
		masked := *s
		masked.Value = "******"
		return &masked
	}
	return s
}

func DefaultSettings() []*entity.Setting {
	return []*entity.Setting{
		{
			Key:           "site_name",
			Value:         "OrigStudio",
			Type:          setting.TypeString,
			Category:      setting.CategoryGeneral,
			Description:   "Site name",
			FallbackValue: "OrigStudio",
			IsBuiltin:     true,
		},
		{
			Key:           "site_description",
			Value:         "A modern media CMS",
			Type:          setting.TypeString,
			Category:      setting.CategoryGeneral,
			Description:   "Site description",
			FallbackValue: "A modern media CMS",
			IsBuiltin:     true,
		},
		{
			Key:           "base_url",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryGeneral,
			Description:   "Base URL of the site",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "allow_registration",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryGeneral,
			Description:   "Allow new user registration",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "allow_upload",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryUpload,
			Description:   "Allow media upload",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "max_upload_size_video",
			Value:         "5368709120",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Max video upload size in bytes (5GB)",
			FallbackValue: "5368709120",
			IsBuiltin:     true,
		},
		{
			Key:           "max_upload_size_image",
			Value:         "52428800",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Max image upload size in bytes (50MB)",
			FallbackValue: "52428800",
			IsBuiltin:     true,
		},
		{
			Key:           "auto_approve",
			Value:         "false",
			Type:          setting.TypeBool,
			Category:      setting.CategoryReview,
			Description:   "Auto approve submitted content",
			FallbackValue: "false",
			IsBuiltin:     true,
		},
		{
			Key:           "require_review",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryReview,
			Description:   "Require review before publishing",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "smtp_host",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryEmail,
			Description:   "SMTP server host",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "smtp_port",
			Value:         "587",
			Type:          setting.TypeInt,
			Category:      setting.CategoryEmail,
			Description:   "SMTP server port",
			FallbackValue: "587",
			IsBuiltin:     true,
		},
		{
			Key:           "smtp_user",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryEmail,
			Description:   "SMTP username",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "smtp_password",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryEmail,
			Description:   "SMTP password",
			FallbackValue: "",
			IsSensitive:   true,
			IsBuiltin:     true,
		},
		{
			Key:           "sprite_frame_interval",
			Value:         "10",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Sprite frame interval in seconds",
			FallbackValue: "10",
			IsBuiltin:     true,
		},
		{
			Key:           "sprite_columns",
			Value:         "5",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Number of columns in sprite image",
			FallbackValue: "5",
			IsBuiltin:     true,
		},
		{
			Key:           "sprite_frame_width",
			Value:         "160",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Width of each sprite frame in pixels",
			FallbackValue: "160",
			IsBuiltin:     true,
		},
		{
			Key:           "sprite_frame_height",
			Value:         "90",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Height of each sprite frame in pixels",
			FallbackValue: "90",
			IsBuiltin:     true,
		},
		{
			Key:           "sprite_max_frames",
			Value:         "100",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Maximum number of frames in sprite",
			FallbackValue: "100",
			IsBuiltin:     true,
		},
		{
			Key:           "thumbnail_quality",
			Value:         "2",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Thumbnail quality (1-31, lower is better)",
			FallbackValue: "2",
			IsBuiltin:     true,
		},
		{
			Key:           "thumbnail_resolution",
			Value:         "1280x720",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Thumbnail resolution",
			FallbackValue: "1280x720",
			IsBuiltin:     true,
		},
		{
			Key:           "thumbnail_position",
			Value:         "0.2",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Thumbnail position (0.0-1.0, percentage of video duration)",
			FallbackValue: "0.2",
			IsBuiltin:     true,
		},
		{
			Key:           "module_articles",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryModule,
			Description:   "Enable Articles content module",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "module_videos",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryModule,
			Description:   "Enable Videos content module",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "module_music",
			Value:         "false",
			Type:          setting.TypeBool,
			Category:      setting.CategoryModule,
			Description:   "Enable Music content module",
			FallbackValue: "false",
			IsBuiltin:     true,
		},
		{
			Key:           "homepage_layout",
			Value:         "auto",
			Type:          setting.TypeString,
			Category:      setting.CategoryModule,
			Description:   "Homepage layout mode: auto|video|article|mixed|welcome",
			FallbackValue: "auto",
			IsBuiltin:     true,
		},
		{
			Key:           "storage_base_path",
			Value:         "./data/uploads",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Base path for media storage (subdirectories are auto-created)",
			FallbackValue: "./data/uploads",
			IsBuiltin:     true,
		},
		{
			Key:           "storage_type",
			Value:         "local",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Storage backend type: local, s3, hybrid",
			FallbackValue: "local",
			IsBuiltin:     true,
		},
		{
			Key:           "s3_endpoint",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "S3/MinIO endpoint URL",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "s3_region",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "S3 region",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "s3_bucket",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "S3 bucket name",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "s3_access_key",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "S3 access key ID",
			FallbackValue: "",
			IsSensitive:   true,
			IsBuiltin:     true,
		},
		{
			Key:           "s3_secret_key",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "S3 secret access key",
			FallbackValue: "",
			IsSensitive:   true,
			IsBuiltin:     true,
		},
		{
			Key:           "s3_use_path_style",
			Value:         "false",
			Type:          setting.TypeBool,
			Category:      setting.CategoryUpload,
			Description:   "Use path-style S3 URLs (true for MinIO)",
			FallbackValue: "false",
			IsBuiltin:     true,
		},
		{
			Key:           "base_urls",
			Value:         `[""]`,
			Type:          setting.TypeJSON,
			Category:      setting.CategoryGeneral,
			Description:   "Allowed site URLs (JSON array, for CORS and validation)",
			FallbackValue: `[""]`,
			IsBuiltin:     true,
		},
		{
			Key:           "primary_url",
			Value:         "",
			Type:          setting.TypeString,
			Category:      setting.CategoryGeneral,
			Description:   "Primary site URL for emails, SEO canonical URLs",
			FallbackValue: "",
			IsBuiltin:     true,
		},
		{
			Key:           "auto_transcode",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryUpload,
			Description:   "Automatically transcode uploaded videos",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "transcode_method",
			Value:         "ffmpeg",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Transcode engine: ffmpeg",
			FallbackValue: "ffmpeg",
			IsBuiltin:     true,
		},
		{
			Key:           "allowed_video_formats",
			Value:         "mp4,webm,mkv,avi,mov",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Allowed video file extensions (comma-separated)",
			FallbackValue: "mp4,webm,mkv,avi,mov",
			IsBuiltin:     true,
		},
		{
			Key:           "allowed_image_formats",
			Value:         "jpg,png,gif,webp",
			Type:          setting.TypeString,
			Category:      setting.CategoryUpload,
			Description:   "Allowed image file extensions (comma-separated)",
			FallbackValue: "jpg,png,gif,webp",
			IsBuiltin:     true,
		},
		{
			Key:           "max_video_duration",
			Value:         "7200",
			Type:          setting.TypeInt,
			Category:      setting.CategoryUpload,
			Description:   "Maximum video duration in seconds (0=unlimited)",
			FallbackValue: "7200",
			IsBuiltin:     true,
		},
		{
			Key:           "smtp_sender_name",
			Value:         "OrigStudio",
			Type:          setting.TypeString,
			Category:      setting.CategoryEmail,
			Description:   "Sender display name for outgoing emails",
			FallbackValue: "OrigStudio",
			IsBuiltin:     true,
		},
		{
			Key:           "smtp_use_tls",
			Value:         "true",
			Type:          setting.TypeBool,
			Category:      setting.CategoryEmail,
			Description:   "Use TLS for SMTP connection",
			FallbackValue: "true",
			IsBuiltin:     true,
		},
		{
			Key:           "min_password_length",
			Value:         "8",
			Type:          setting.TypeInt,
			Category:      setting.CategoryGeneral,
			Description:   "Minimum password length for registration",
			FallbackValue: "8",
			IsBuiltin:     true,
		},
		{
			Key:           "require_email_verification",
			Value:         "false",
			Type:          setting.TypeBool,
			Category:      setting.CategoryGeneral,
			Description:   "Require email verification after registration",
			FallbackValue: "false",
			IsBuiltin:     true,
		},
		{
			Key:           "api_rate_limit",
			Value:         "60",
			Type:          setting.TypeInt,
			Category:      setting.CategoryGeneral,
			Description:   "API rate limit (requests per minute per IP, 0=unlimited)",
			FallbackValue: "60",
			IsBuiltin:     true,
		},
	}
}
