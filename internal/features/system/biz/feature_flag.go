package biz

import (
	"context"
	"strings"
	"sync"
	"time"

	systemdto "origadmin/application/origstudio/internal/features/system/dto"
)

type FeatureFlagUseCase struct {
	settingUC *SettingUseCase
	cache     map[string]bool
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
	expiredAt time.Time
}

func NewFeatureFlagUseCase(settingUC *SettingUseCase) *FeatureFlagUseCase {
	return &FeatureFlagUseCase{
		settingUC: settingUC,
		cacheTTL:  30 * time.Second,
	}
}

func (uc *FeatureFlagUseCase) IsEnabled(ctx context.Context, flag string) bool {
	flags := uc.GetAll(ctx)
	return flags[flag]
}

func (uc *FeatureFlagUseCase) GetAll(ctx context.Context) map[string]bool {
	uc.cacheMu.RLock()
	if uc.cache != nil && time.Now().Before(uc.expiredAt) {
		result := uc.cache
		uc.cacheMu.RUnlock()
		return result
	}
	uc.cacheMu.RUnlock()

	uc.cacheMu.Lock()
	defer uc.cacheMu.Unlock()

	if uc.cache != nil && time.Now().Before(uc.expiredAt) {
		return uc.cache
	}

	flags := uc.loadFromSettings(ctx)
	uc.cache = flags
	uc.expiredAt = time.Now().Add(uc.cacheTTL)
	return flags
}

func (uc *FeatureFlagUseCase) loadFromSettings(ctx context.Context) map[string]bool {
	flags := make(map[string]bool)

	settings, err := uc.settingUC.ListByCategory(ctx, string(systemdto.SettingCategoryFeature))
	if err != nil {
		return flags
	}

	for _, s := range settings {
		flagName := strings.TrimPrefix(s.Key, "feature_")
		flagName = toCamelCase(flagName)
		flags[flagName] = s.Value == "true"
	}

	return flags
}

func (uc *FeatureFlagUseCase) InvalidateCache() {
	uc.cacheMu.Lock()
	uc.cache = nil
	uc.cacheMu.Unlock()
}

func (uc *FeatureFlagUseCase) SetFlag(ctx context.Context, flag string, enabled bool) error {
	key := "feature_" + toSnakeCase(flag)
	val := "false"
	if enabled {
		val = "true"
	}

	existing, err := uc.settingUC.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	if existing != nil {
		existing.Value = val
		_, err = uc.settingUC.Upsert(ctx, existing)
	} else {
		_, err = uc.settingUC.Upsert(ctx, &systemdto.SettingDTO{
			Key:           key,
			Value:         val,
			Type:          systemdto.SettingTypeBool,
			Category:      systemdto.SettingCategoryFeature,
			Description:   "Feature flag: " + flag,
			FallbackValue: "false",
			IsBuiltin:     false,
		})
	}

	if err != nil {
		return err
	}

	uc.InvalidateCache()
	uc.settingUC.InvalidateCache()
	return nil
}

func toCamelCase(snake string) string {
	parts := strings.Split(snake, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func toSnakeCase(camel string) string {
	var result []byte
	for i, ch := range camel {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, byte(ch))
	}
	return strings.ToLower(string(result))
}
