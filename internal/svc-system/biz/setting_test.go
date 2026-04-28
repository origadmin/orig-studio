package biz

import (
	"testing"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/setting"
)

func TestDefaultSettings(t *testing.T) {
	defaults := DefaultSettings()
	if len(defaults) != 13 {
		t.Fatalf("expected 13 default settings, got %d", len(defaults))
	}

	keys := make(map[string]bool)
	for _, s := range defaults {
		if s.Key == "" {
			t.Error("default setting has empty key")
		}
		if keys[s.Key] {
			t.Errorf("duplicate key in default settings: %s", s.Key)
		}
		keys[s.Key] = true
	}
}

func TestDefaultSettingsCategories(t *testing.T) {
	defaults := DefaultSettings()

	categories := map[string]int{
		"general": 0,
		"upload":  0,
		"review":  0,
		"email":   0,
	}
	for _, s := range defaults {
		cat := string(s.Category)
		if _, ok := categories[cat]; !ok {
			t.Errorf("unexpected category: %s", cat)
		}
		categories[cat]++
	}

	if categories["general"] < 2 {
		t.Errorf("expected at least 2 general settings, got %d", categories["general"])
	}
	if categories["upload"] < 1 {
		t.Errorf("expected at least 1 upload setting, got %d", categories["upload"])
	}
	if categories["review"] < 1 {
		t.Errorf("expected at least 1 review setting, got %d", categories["review"])
	}
	if categories["email"] < 1 {
		t.Errorf("expected at least 1 email setting, got %d", categories["email"])
	}
}

func TestDefaultSettingsSensitiveFields(t *testing.T) {
	defaults := DefaultSettings()

	sensitiveKeys := map[string]bool{
		"smtp_password": true,
	}
	for _, s := range defaults {
		if sensitiveKeys[s.Key] && !s.IsSensitive {
			t.Errorf("expected %s to be sensitive", s.Key)
		}
		if !sensitiveKeys[s.Key] && s.IsSensitive {
			t.Errorf("expected %s to not be sensitive", s.Key)
		}
	}
}

func TestDefaultSettingsTypes(t *testing.T) {
	defaults := DefaultSettings()

	boolKeys := map[string]bool{
		"allow_registration": true,
		"allow_upload":       true,
		"auto_approve":       true,
		"require_review":     true,
	}
	intKeys := map[string]bool{
		"max_upload_size_video": true,
		"max_upload_size_image": true,
		"smtp_port":             true,
	}
	for _, s := range defaults {
		if boolKeys[s.Key] && s.Type != setting.TypeBool {
			t.Errorf("expected %s to be bool type, got %s", s.Key, s.Type)
		}
		if intKeys[s.Key] && s.Type != setting.TypeInt {
			t.Errorf("expected %s to be int type, got %s", s.Key, s.Type)
		}
	}
}

func TestMaskSensitive(t *testing.T) {
	uc := &SettingUseCase{}

	sensitive := &entity.Setting{
		Key:         "smtp_password",
		Value:       "secret123",
		IsSensitive: true,
	}
	masked := uc.MaskSensitive(sensitive)
	if masked.Value != "******" {
		t.Errorf("expected masked value '******', got '%s'", masked.Value)
	}
	if masked.Key != "smtp_password" {
		t.Errorf("expected key to remain 'smtp_password', got '%s'", masked.Key)
	}

	normal := &entity.Setting{
		Key:         "site_name",
		Value:       "OrigCMS",
		IsSensitive: false,
	}
	notMasked := uc.MaskSensitive(normal)
	if notMasked.Value != "OrigCMS" {
		t.Errorf("expected value to remain 'OrigCMS', got '%s'", notMasked.Value)
	}
}

func TestMaskSensitiveDoesNotMutateOriginal(t *testing.T) {
	uc := &SettingUseCase{}

	sensitive := &entity.Setting{
		Key:         "smtp_password",
		Value:       "secret123",
		IsSensitive: true,
	}
	masked := uc.MaskSensitive(sensitive)
	if sensitive.Value != "secret123" {
		t.Error("MaskSensitive should not mutate the original setting")
	}
	if masked.Value != "******" {
		t.Errorf("expected masked value '******', got '%s'", masked.Value)
	}
}

func TestConfigProviderInterface(t *testing.T) {
	var _ ConfigProvider = (*SettingUseCase)(nil)
}
