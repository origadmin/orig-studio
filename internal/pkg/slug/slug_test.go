package slug

import (
	"strings"
	"testing"
)

func TestGenerate_PureASCII(t *testing.T) {
	result := Generate("Technology")
	if !strings.HasPrefix(result, "technology-") {
		t.Errorf("expected prefix 'technology-', got '%s'", result)
	}
	suffix := result[len("technology-"):]
	if len(suffix) != DefaultSuffixLen {
		t.Errorf("expected suffix length %d, got %d", DefaultSuffixLen, len(suffix))
	}
}

func TestGenerate_Chinese(t *testing.T) {
	result := Generate("前端开发")
	if result == "" {
		t.Error("expected non-empty slug for Chinese input")
	}
	if result == "前端开发" {
		t.Error("Chinese slug should be Base58 encoded")
	}
	if !strings.Contains(result, "-") {
		t.Error("expected suffix separator '-' in generated slug")
	}
}

func TestGenerate_MixedChineseEnglish(t *testing.T) {
	result := Generate("React前端")
	if result == "" {
		t.Error("expected non-empty slug for mixed input")
	}
	if !strings.Contains(result, "-") {
		t.Error("expected suffix separator '-' in generated slug")
	}
}

func TestGenerate_EmptyInput(t *testing.T) {
	result := Generate("")
	if !strings.HasPrefix(result, FallbackSlug+"-") {
		t.Errorf("expected fallback prefix '%s-', got '%s'", FallbackSlug, result)
	}
}

func TestGenerateWithoutSuffix(t *testing.T) {
	result := GenerateWithoutSuffix("Technology")
	if result != "technology" {
		t.Errorf("expected 'technology', got '%s'", result)
	}
}

func TestGenerateWithoutSuffix_Chinese(t *testing.T) {
	result := GenerateWithoutSuffix("前端开发")
	if result == "" {
		t.Error("expected non-empty slug")
	}
	if strings.Contains(result, "-") {
		t.Error("GenerateWithoutSuffix should not add suffix")
	}
}

func TestGenerateUserSlug_FromNickname(t *testing.T) {
	result := GenerateUserSlug("John Doe", "johndoe")
	if result != "john-doe" {
		t.Errorf("expected 'john-doe', got '%s'", result)
	}
}

func TestGenerateUserSlug_ChineseNickname(t *testing.T) {
	result := GenerateUserSlug("张三", "zhangsan")
	if result == "" {
		t.Error("expected non-empty slug")
	}
	if result == "前端开发" || result == "张三" {
		t.Error("Chinese nickname should be Base58 encoded")
	}
}

func TestGenerateUserSlug_FallbackToUsername(t *testing.T) {
	result := GenerateUserSlug("", "johndoe")
	if result != "johndoe" {
		t.Errorf("expected 'johndoe' from username fallback, got '%s'", result)
	}
}

func TestGenerateUserSlug_ReservedSlug(t *testing.T) {
	result := GenerateUserSlug("admin", "adminuser")
	if result == "admin" {
		t.Error("reserved slug should be skipped")
	}
	if result != "adminuser" {
		t.Errorf("expected fallback to username 'adminuser', got '%s'", result)
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"john-doe", true},
		{"abc", true},
		{"a1b2c3", true},
		{"ab", false},
		{"ABC", false},
		{"a b", false},
		{"-abc", false},
		{"abc-", false},
		{strings.Repeat("a", 65), false},
	}
	for _, tt := range tests {
		result := IsValid(tt.input)
		if result != tt.expected {
			t.Errorf("IsValid(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsReserved(t *testing.T) {
	if !IsReserved("admin") {
		t.Error("'admin' should be reserved")
	}
	if !IsReserved("ADMIN") {
		t.Error("'ADMIN' should be reserved (case-insensitive)")
	}
	if IsReserved("john") {
		t.Error("'john' should not be reserved")
	}
}

func TestEnsureUnique_NoConflict(t *testing.T) {
	existsFn := func(slug string) (bool, error) { return false, nil }
	result, err := EnsureUnique("technology", existsFn)
	if err != nil {
		t.Fatal(err)
	}
	if result != "technology" {
		t.Errorf("expected 'technology', got '%s'", result)
	}
}

func TestEnsureUnique_WithConflict(t *testing.T) {
	callCount := 0
	existsFn := func(slug string) (bool, error) {
		callCount++
		return callCount <= 2, nil
	}
	result, err := EnsureUnique("technology", existsFn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result, "technology-") {
		t.Errorf("expected prefixed slug, got '%s'", result)
	}
}

func TestGenerate_SlugUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s := Generate("test")
		if seen[s] {
			t.Errorf("duplicate slug generated: %s", s)
		}
		seen[s] = true
	}
}
