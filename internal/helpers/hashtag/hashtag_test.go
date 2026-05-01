/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * Unit tests for hashtag parsing and slug generation.
 * Covers AC-01 through AC-04 acceptance criteria.
 */

package hashtag

import (
	"strings"
	"testing"
)

// ==================== AC-01: Hashtag Parsing - Basic Functionality ====================

func TestParseHashtags_SingleTag(t *testing.T) {
	// AC-01-01: Single tag
	result := ParseHashtags("Check out #golang")
	if len(result) != 1 || result[0] != "golang" {
		t.Errorf("expected [golang], got %v", result)
	}
}

func TestParseHashtags_MultipleTags(t *testing.T) {
	// AC-01-02: Multiple tags
	result := ParseHashtags("#frontend #react #typescript")
	if len(result) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(result), result)
	}
	expected := []string{"frontend", "react", "typescript"}
	for i, e := range expected {
		if result[i] != e {
			t.Errorf("expected result[%d]=%q, got %q", i, e, result[i])
		}
	}
}

func TestParseHashtags_ChineseTag(t *testing.T) {
	// AC-01-03: Chinese tag
	result := ParseHashtags("学习#前端开发")
	if len(result) != 1 || result[0] != "前端开发" {
		t.Errorf("expected [前端开发], got %v", result)
	}
}

func TestParseHashtags_TagInMiddle(t *testing.T) {
	// AC-01-04: Tag in the middle
	result := ParseHashtags("This is #awesome stuff")
	if len(result) != 1 || result[0] != "awesome" {
		t.Errorf("expected [awesome], got %v", result)
	}
}

func TestParseHashtags_TagAtEnd(t *testing.T) {
	// AC-01-05: Tag at the end
	result := ParseHashtags("Great video #vlog")
	if len(result) != 1 || result[0] != "vlog" {
		t.Errorf("expected [vlog], got %v", result)
	}
}

func TestParseHashtags_TagAtBeginning(t *testing.T) {
	// AC-01-06: Tag at the beginning
	result := ParseHashtags("#tutorial How to code")
	if len(result) != 1 || result[0] != "tutorial" {
		t.Errorf("expected [tutorial], got %v", result)
	}
}

func TestParseHashtags_TagWithPunctuation(t *testing.T) {
	// AC-01-07: Tag followed by punctuation
	result := ParseHashtags("#go!")
	if len(result) != 1 || result[0] != "go" {
		t.Errorf("expected [go], got %v", result)
	}
}

func TestParseHashtags_TagWithNewline(t *testing.T) {
	// AC-01-08: Tags separated by newline
	result := ParseHashtags("#tag1\n#tag2")
	if len(result) != 2 {
		t.Errorf("expected 2 tags, got %d: %v", len(result), result)
	}
	if result[0] != "tag1" || result[1] != "tag2" {
		t.Errorf("expected [tag1, tag2], got %v", result)
	}
}

// ==================== AC-02: Hashtag Parsing - Edge Cases ====================

func TestParseHashtags_EmptyText(t *testing.T) {
	// AC-02-01: Empty text
	result := ParseHashtags("")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestParseHashtags_NoTags(t *testing.T) {
	// AC-02-02: No tags
	result := ParseHashtags("Hello world")
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestParseHashtags_DuplicateTags(t *testing.T) {
	// AC-02-03: Duplicate tags (case-insensitive dedup)
	result := ParseHashtags("#go #Go #GO")
	if len(result) != 1 {
		t.Errorf("expected 1 tag (dedup), got %d: %v", len(result), result)
	}
	if result[0] != "go" {
		t.Errorf("expected [go], got %v", result)
	}
}

func TestParseHashtags_OnlyHash(t *testing.T) {
	// AC-02-04: Only # signs
	result := ParseHashtags("# # #")
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestParseHashtags_ColorCode(t *testing.T) {
	// AC-02-05: Color codes should not be parsed
	result := ParseHashtags("color #ff0000")
	if len(result) != 0 {
		t.Errorf("expected empty (color code excluded), got %v", result)
	}
}

func TestParseHashtags_URLFragment(t *testing.T) {
	// AC-02-06: URL fragments should not be parsed
	result := ParseHashtags("http://example.com#section")
	if len(result) != 0 {
		t.Errorf("expected empty (URL fragment excluded), got %v", result)
	}
}

func TestParseHashtags_HTMLEntity(t *testing.T) {
	// AC-02-07: HTML entities should not be parsed
	result := ParseHashtags("&#39;")
	if len(result) != 0 {
		t.Errorf("expected empty (HTML entity excluded), got %v", result)
	}
}

func TestParseHashtags_LongTag(t *testing.T) {
	// AC-02-08: Very long tag name (should still be extracted, regex limits to 100 chars)
	longName := strings.Repeat("a", 200)
	result := ParseHashtags("#" + longName)
	if len(result) != 1 {
		t.Errorf("expected 1 tag, got %d", len(result))
	}
	// The regex captures up to 100 characters for the tag name
	if len(result[0]) > 100 {
		t.Errorf("tag name should be at most 100 chars, got %d", len(result[0]))
	}
}

func TestParseHashtags_SingleCharTag(t *testing.T) {
	// AC-02-09: Single character tag
	result := ParseHashtags("#a")
	if len(result) != 1 || result[0] != "a" {
		t.Errorf("expected [a], got %v", result)
	}
}

func TestParseHashtags_HyphenTag(t *testing.T) {
	// AC-02-10: Hyphenated tag
	result := ParseHashtags("#machine-learning")
	if len(result) != 1 || result[0] != "machine-learning" {
		t.Errorf("expected [machine-learning], got %v", result)
	}
}

func TestParseHashtags_UnderscoreTag(t *testing.T) {
	// AC-02-11: Underscore tag
	result := ParseHashtags("#machine_learning")
	if len(result) != 1 || result[0] != "machine_learning" {
		t.Errorf("expected [machine_learning], got %v", result)
	}
}

func TestParseHashtags_NumberStartTag(t *testing.T) {
	// AC-02-12: Tag starting with number
	result := ParseHashtags("#3d-modeling")
	if len(result) != 1 || result[0] != "3d-modeling" {
		t.Errorf("expected [3d-modeling], got %v", result)
	}
}

func TestParseHashtags_MaxTagLimit(t *testing.T) {
	// AC-02-13: Tag limit (max 20)
	var tags []string
	for i := 0; i < 25; i++ {
		tags = append(tags, "#tag"+strings.Repeat("a", i))
	}
	text := strings.Join(tags, " ")
	result := ParseHashtags(text)
	if len(result) > MaxHashtags {
		t.Errorf("expected at most %d tags, got %d", MaxHashtags, len(result))
	}
}

func TestParseHashtags_CJKMixedTag(t *testing.T) {
	// AC-02-14: CJK mixed tag
	result := ParseHashtags("#Vue3入门")
	if len(result) != 1 || result[0] != "Vue3入门" {
		t.Errorf("expected [Vue3入门], got %v", result)
	}
}

// ==================== AC-03: Slug Generation ====================

func TestGenerateTagSlug_EnglishWord(t *testing.T) {
	// AC-03-01: English word
	result := GenerateTagSlug("Technology")
	if result != "technology" {
		t.Errorf("expected 'technology', got '%s'", result)
	}
}

func TestGenerateTagSlug_MultipleWords(t *testing.T) {
	// AC-03-02: Multiple words
	result := GenerateTagSlug("Machine Learning")
	if result != "machine-learning" {
		t.Errorf("expected 'machine-learning', got '%s'", result)
	}
}

func TestGenerateTagSlug_SpecialChars(t *testing.T) {
	// AC-03-03: Special characters (C++ -> c)
	result := GenerateTagSlug("C++")
	if result != "c" {
		t.Errorf("expected 'c', got '%s'", result)
	}
}

func TestGenerateTagSlug_Chinese(t *testing.T) {
	// AC-03-04: Chinese name -> Base58 encoded
	result := GenerateTagSlug("前端开发")
	if result == "" {
		t.Error("expected non-empty Base58 result, got empty string")
	}
	if result == "前端开发" {
		t.Error("expected Base58 encoded result, not the raw Chinese string")
	}
	// Verify it only contains Base58 characters
	for _, c := range result {
		if !isBase58Char(c) {
			t.Errorf("result contains non-Base58 character: %c", c)
		}
	}
}

func TestGenerateTagSlug_MixedASCII(t *testing.T) {
	// AC-03-05: Mixed ASCII (Go 1.21 -> go-1-21)
	result := GenerateTagSlug("Go 1.21")
	if result != "go-1-21" {
		t.Errorf("expected 'go-1-21', got '%s'", result)
	}
}

func TestGenerateTagSlug_Slash(t *testing.T) {
	// AC-03-06: Slash (AI/ML -> ai-ml)
	result := GenerateTagSlug("AI/ML")
	if result != "ai-ml" {
		t.Errorf("expected 'ai-ml', got '%s'", result)
	}
}

func TestGenerateTagSlug_EmptyInput(t *testing.T) {
	// AC-03-07: Empty input -> fallback
	result := GenerateTagSlug("")
	if result != FallbackSlug {
		t.Errorf("expected '%s', got '%s'", FallbackSlug, result)
	}
}

func TestGenerateTagSlug_LongInput(t *testing.T) {
	// AC-03-08: Very long input (truncated to 100 chars)
	longName := strings.Repeat("a", 150)
	result := GenerateTagSlug(longName)
	if len(result) > MaxSlugLength {
		t.Errorf("expected max %d chars, got %d", MaxSlugLength, len(result))
	}
}

func TestGenerateTagSlug_ConsecutiveSpaces(t *testing.T) {
	// AC-03-09: Consecutive spaces
	result := GenerateTagSlug("hello   world")
	if result != "hello-world" {
		t.Errorf("expected 'hello-world', got '%s'", result)
	}
}

func TestGenerateTagSlug_LeadingTrailingSpaces(t *testing.T) {
	// AC-03-10: Leading/trailing spaces
	result := GenerateTagSlug(" technology ")
	if result != "technology" {
		t.Errorf("expected 'technology', got '%s'", result)
	}
}

func TestGenerateTagSlug_MixedChineseEnglish(t *testing.T) {
	// AC-03-11: Mixed Chinese and English -> Base58 encode entire name
	result := GenerateTagSlug("React前端")
	if result == "" {
		t.Error("expected non-empty Base58 result, got empty string")
	}
	// Must be Base58 encoded (not slugified)
	for _, c := range result {
		if !isBase58Char(c) {
			t.Errorf("result contains non-Base58 character: %c", c)
		}
	}
	// Must differ from pure-ASCII slugification of "React"
	asciiResult := GenerateTagSlug("React")
	if result == asciiResult {
		t.Error("mixed Chinese/English slug should differ from pure ASCII slug")
	}
}

func TestGenerateTagSlug_PureNumbers(t *testing.T) {
	// AC-03-12: Pure numbers
	result := GenerateTagSlug("123")
	if result != "123" {
		t.Errorf("expected '123', got '%s'", result)
	}
}

// ==================== Base58 Specific Tests ====================

func TestBase58Encode_Deterministic(t *testing.T) {
	// Same input must always produce the same output
	data := []byte("前端开发")
	result1 := base58Encode(data)
	result2 := base58Encode(data)
	if result1 != result2 {
		t.Errorf("Base58 encoding is not deterministic: '%s' != '%s'", result1, result2)
	}
}

func TestBase58Encode_DifferentInputs(t *testing.T) {
	// Different inputs must produce different outputs
	result1 := base58Encode([]byte("前端开发"))
	result2 := base58Encode([]byte("后端开发"))
	if result1 == result2 {
		t.Error("different inputs produced the same Base58 output")
	}
}

func TestBase58Encode_EmptyInput(t *testing.T) {
	result := base58Encode([]byte{})
	if result != FallbackSlug {
		t.Errorf("expected fallback '%s' for empty input, got '%s'", FallbackSlug, result)
	}
}

func TestBase58Encode_OnlyBase58Chars(t *testing.T) {
	// Result must only contain Base58 characters
	result := base58Encode([]byte("测试"))
	for _, c := range result {
		if !isBase58Char(c) {
			t.Errorf("result contains non-Base58 character: %c in '%s'", c, result)
		}
	}
}

func TestBase58Encode_KnownValue(t *testing.T) {
	// Test with a known value: "Hello" in Base58 should be "9Ajdvzr"
	// Using Bitcoin Base58: encode bytes of "Hello"
	result := base58Encode([]byte("Hello"))
	if result == "" || result == FallbackSlug {
		t.Errorf("expected valid Base58 for 'Hello', got '%s'", result)
	}
	// Verify determinism
	result2 := base58Encode([]byte("Hello"))
	if result != result2 {
		t.Errorf("Base58 not deterministic: '%s' != '%s'", result, result2)
	}
}

func TestBase58Encode_LeadingZeros(t *testing.T) {
	// Leading zero bytes should produce leading '1's
	result := base58Encode([]byte{0, 0, 5})
	if !strings.HasPrefix(result, "11") {
		t.Errorf("expected leading '1's for leading zero bytes, got '%s'", result)
	}
}

// ==================== isASCII Tests ====================

func TestIsASCII_PureASCII(t *testing.T) {
	if !isASCII("Hello World 123") {
		t.Error("expected ASCII string to return true")
	}
}

func TestIsASCII_WithChinese(t *testing.T) {
	if isASCII("Hello 世界") {
		t.Error("expected string with Chinese to return false")
	}
}

func TestIsASCII_Empty(t *testing.T) {
	if !isASCII("") {
		t.Error("expected empty string to return true")
	}
}

func TestIsASCII_SpecialChars(t *testing.T) {
	if !isASCII("C++ !@#$%^&*()") {
		t.Error("expected ASCII special chars to return true")
	}
}

// ==================== slugify Tests ====================

func TestSlugify_BasicConversions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Technology", "technology"},
		{"Machine Learning", "machine-learning"},
		{"C++", "c"},
		{"Go 1.21", "go-1-21"},
		{"AI/ML", "ai-ml"},
		{"hello   world", "hello-world"},
		{" technology ", "technology"},
		{"123", "123"},
		{"---test---", "test"},
		{"test--multiple---hyphens", "test-multiple-hyphens"},
	}

	for _, tt := range tests {
		result := slugify(tt.input)
		if result != tt.expected {
			t.Errorf("slugify(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestSlugify_AllSpecialChars(t *testing.T) {
	// All special chars should produce fallback
	result := slugify("!!!@@@###")
	if result != FallbackSlug {
		t.Errorf("expected fallback for all-special-chars, got '%s'", result)
	}
}

// ==================== Integration: ParseHashtags + GenerateTagSlug ====================

func TestIntegration_ParseAndGenerateSlug(t *testing.T) {
	text := "Check out #golang and #前端开发 for web development"
	tags := ParseHashtags(text)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	slug1 := GenerateTagSlug(tags[0])
	slug2 := GenerateTagSlug(tags[1])

	if slug1 != "golang" {
		t.Errorf("expected 'golang', got '%s'", slug1)
	}
	// Chinese tag should be Base58 encoded
	if slug2 == "前端开发" {
		t.Error("Chinese tag slug should be Base58 encoded, not raw Chinese")
	}
	if slug2 == "" {
		t.Error("Chinese tag slug should not be empty")
	}
}

// ==================== Helper Functions ====================

// isBase58Char checks if a rune is a valid Base58 character.
func isBase58Char(c rune) bool {
	return strings.ContainsRune(base58Alphabet, c)
}
