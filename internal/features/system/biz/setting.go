package biz

import (
	"context"
	"strconv"
	"sync"
	"time"

	systemdal "origadmin/application/origstudio/internal/features/system/dal"
	systemdto "origadmin/application/origstudio/internal/features/system/dto"
)

type ConfigProvider interface {
	Get(ctx context.Context, key string) string
	GetBool(ctx context.Context, key string) bool
	GetInt(ctx context.Context, key string) int
	GetAll(ctx context.Context) map[string]string
}

type cacheEntry struct {
	value     map[string]*systemdto.SettingDTO
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

func (uc *SettingUseCase) GetByKey(ctx context.Context, key string) (*systemdto.SettingDTO, error) {
	return uc.repo.GetByKey(ctx, key)
}

func (uc *SettingUseCase) ListByCategory(
	ctx context.Context,
	category string,
) ([]*systemdto.SettingDTO, error) {
	return uc.repo.ListByCategory(ctx, category)
}

func (uc *SettingUseCase) ListAll(ctx context.Context) ([]*systemdto.SettingDTO, error) {
	return uc.repo.ListAll(ctx)
}

func (uc *SettingUseCase) Upsert(ctx context.Context, s *systemdto.SettingDTO) (*systemdto.SettingDTO, error) {
	result, err := uc.repo.Upsert(ctx, s)
	if err != nil {
		return nil, err
	}
	uc.InvalidateCache()
	return result, nil
}

func (uc *SettingUseCase) BatchUpsert(ctx context.Context, settings []*systemdto.SettingDTO) error {
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

func (uc *SettingUseCase) ResetToDefault(ctx context.Context, key string) (*systemdto.SettingDTO, error) {
	result, err := uc.repo.ResetToDefault(ctx, key)
	if err != nil {
		return nil, err
	}
	uc.InvalidateCache()
	return result, nil
}

func (uc *SettingUseCase) SeedDefaults(ctx context.Context) error {
	return uc.repo.SeedDefaults(ctx, systemdal.DefaultSettings())
}

func (uc *SettingUseCase) InvalidateCache() {
	uc.cacheMu.Lock()
	uc.cache = nil
	uc.cacheMu.Unlock()
}

func (uc *SettingUseCase) loadCache(ctx context.Context) (map[string]*systemdto.SettingDTO, error) {
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

	m := make(map[string]*systemdto.SettingDTO, len(items))
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

func (uc *SettingUseCase) MaskSensitive(s *systemdto.SettingDTO) *systemdto.SettingDTO {
	if s.IsSensitive {
		masked := *s
		masked.Value = "******"
		return &masked
	}
	return s
}
