/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * Package hashtag provides hashtag parsing and slug generation utilities.
 * It extracts #hashtags from text and generates URL-friendly slugs using
 * a one-size-fits-all strategy: pure ASCII names are slugified, while
 * names containing any non-ASCII characters are Base58-encoded entirely.
 */

package hashtag

import (
	"math/big"
	"regexp"
	"strings"
)

const (
	// MaxHashtags is the maximum number of hashtags extracted from a single text.
	MaxHashtags = 20
	// MaxSlugLength is the maximum length of a generated slug.
	MaxSlugLength = 100
	// FallbackSlug is the default slug when generation yields an empty result.
	FallbackSlug = "tag"
)

// base58Alphabet is the Bitcoin Base58 alphabet (no 0/O/I/l).
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	// hashtagRegex matches #tag patterns. Supports Unicode letters, numbers,
	// underscores, and hyphens. Single-character tags are also matched.
	hashtagRegex = regexp.MustCompile(`#([\p{L}\p{N}_][\p{L}\p{N}_\-]{0,98}[\p{L}\p{N}_])|#([\p{L}\p{N}_])`)

	// excludePatterns are patterns that should not be parsed as hashtags.
	excludePatterns = []*regexp.Regexp{
		regexp.MustCompile(`https?://\S*#\S*`),    // URL with fragment
		regexp.MustCompile(`&#[a-zA-Z0-9]+;`),     // HTML entities
	}

	// colorCodeRegex matches CSS color codes like #fff, #ff0000, #ffffff00.
	colorCodeRegex = regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)

	// nonAlphaNumRegex matches sequences of non-alphanumeric ASCII characters.
	nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

	// multiHyphenRegex matches consecutive hyphens.
	multiHyphenRegex = regexp.MustCompile(`-{2,}`)
)

// ParseHashtags extracts hashtag names from the given text.
// It excludes URL fragments, color codes, and HTML entities.
// Results are deduplicated (case-insensitive) and limited to MaxHashtags.
func ParseHashtags(text string) []string {
	if text == "" {
		return nil
	}

	// Pre-process: replace excluded patterns with spaces
	cleaned := text
	for _, p := range excludePatterns {
		cleaned = p.ReplaceAllString(cleaned, " ")
	}
	// Replace color codes with spaces
	cleaned = colorCodeRegex.ReplaceAllString(cleaned, " ")

	// Extract hashtag names
	matches := hashtagRegex.FindAllStringSubmatch(cleaned, -1)

	seen := make(map[string]bool)
	var result []string

	for _, match := range matches {
		var name string
		if match[1] != "" {
			name = match[1]
		} else if match[2] != "" {
			name = match[2]
		}
		if name == "" {
			continue
		}

		// Trim leading/trailing hyphens
		name = strings.Trim(name, "-")

		if name == "" {
			continue
		}

		// Deduplicate (case-insensitive)
		lower := strings.ToLower(name)
		if seen[lower] {
			continue
		}
		seen[lower] = true

		result = append(result, name)
		if len(result) >= MaxHashtags {
			break
		}
	}

	return result
}

// GenerateTagSlug generates a URL-friendly slug from a tag name.
// Strategy (one-size-fits-all):
//   - Pure ASCII names: slugify (lowercase, replace non-alphanum with hyphens)
//   - Names with any non-ASCII: Base58-encode the entire name (UTF-8 bytes)
//
// Empty results fall back to FallbackSlug ("tag").
func GenerateTagSlug(name string) string {
	if name == "" {
		return FallbackSlug
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return FallbackSlug
	}

	if isASCII(name) {
		return slugify(name)
	}
	return base58Encode([]byte(name))
}

// isASCII checks if all bytes in the string are within the ASCII range (0-127).
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// slugify converts a pure-ASCII name to a URL-friendly slug.
func slugify(name string) string {
	slug := strings.ToLower(name)
	slug = nonAlphaNumRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	slug = multiHyphenRegex.ReplaceAllString(slug, "-")

	if len(slug) > MaxSlugLength {
		slug = slug[:MaxSlugLength]
		// Trim trailing hyphen if truncation left one
		slug = strings.TrimRight(slug, "-")
	}

	if slug == "" {
		return FallbackSlug
	}
	return slug
}

// base58Encode encodes a byte slice using the Bitcoin Base58 alphabet.
// The result is deterministic: the same input always produces the same output.
func base58Encode(data []byte) string {
	if len(data) == 0 {
		return FallbackSlug
	}

	// Count leading zero bytes (they become leading '1's in Base58)
	leadingZeros := 0
	for _, b := range data {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	// Convert to big integer
	num := new(big.Int).SetBytes(data)

	// Encode
	alphabet := base58Alphabet
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	var encoded []byte
	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		encoded = append(encoded, alphabet[mod.Int64()])
	}

	// Add leading '1's for each leading zero byte
	for i := 0; i < leadingZeros; i++ {
		encoded = append(encoded, alphabet[0])
	}

	// Reverse the encoded bytes
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}

	result := string(encoded)

	// Truncate to max slug length
	if len(result) > MaxSlugLength {
		result = result[:MaxSlugLength]
	}

	if result == "" {
		return FallbackSlug
	}
	return result
}
