/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * Tag merging for the unified tag model (BUG-131 / category_tag strategy).
 *
 * Three tag sources must converge into one canonical, deduplicated set before a
 * media record is persisted:
 *   1. manual tags   — typed explicitly by the editor
 *   2. title tags    — #hashtags parsed from the title
 *   3. intro tags    — #hashtags parsed from the description
 *
 * Normalization (trim / '#'-strip / lowercase / validation) happens BEFORE
 * merging, so the sources are comparable and "Golang" / "golang" collapse to a
 * single entry. The result is what gets written to both the jsonb projection
 * (content_media.tags, DTO only) and the authoritative M2M table
 * (content_media_tags) once BUG-132 is fixed.
 *
 * The hashtag extraction regex mirrors the frontend rule in
 * web/src/lib/utils/hashtag.ts (HASHTAG_REGEX), keeping front/back behaviour
 * in lock-step.
 */

package biz

import (
	"regexp"
	"strings"
	"unicode"
)

// hashtagRegex extracts hashtags of the form #word where word starts with a
// Unicode letter/number/underscore and continues with letters/numbers/_/-.
// The `u` flag in the frontend is implicit in Go's RE2 engine (Unicode-aware).
var hashtagRegex = regexp.MustCompile(`#[\p{L}\p{N}_][\p{L}\p{N}_\-]*`)

// isTagRune reports whether r is allowed inside a canonical tag token.
// Mirrors the frontend class [\p{L}\p{N}_-].
func isTagRune(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsNumber(r)
}

// NormalizeTag canonicalizes a single raw tag token:
//   - trims surrounding whitespace
//   - strips a single leading '#'
//   - lowercases
//   - keeps only Unicode letters/numbers/_/- (drops control chars, spaces, …)
//   - drops a result that is empty or starts with '-'
//
// Returns ("", false) when the input carries no valid token.
func NormalizeTag(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false
	}
	if strings.HasPrefix(s, "#") {
		s = strings.TrimSpace(s[1:])
	}
	s = strings.ToLower(s)

	var b strings.Builder
	for _, r := range s {
		if isTagRune(r) {
			b.WriteRune(r)
		}
	}
	out := strings.TrimLeft(b.String(), "-")
	if out == "" {
		return "", false
	}
	return out, true
}

// ExtractHashtags returns the #hashtag tokens found in text (without the '#').
// Returns nil for empty input.
func ExtractHashtags(text string) []string {
	if text == "" {
		return nil
	}
	matches := hashtagRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		// m is guaranteed to be at least 2 runes ("#" + 1 class char).
		out = append(out, m[1:])
	}
	return out
}

// NormalizeAndMerge merges three tag sources — manual tags, hashtags parsed from
// the title, and hashtags parsed from the description — into a single canonical,
// deduplicated set.
//
// Source precedence (first occurrence wins, matching the frontend
// mergeTagsWithHashtags): manual → title → description.
func NormalizeAndMerge(title, description string, manual []string) []string {
	sources := [][]string{manual, ExtractHashtags(title), ExtractHashtags(description)}

	seen := make(map[string]struct{}, len(manual))
	out := make([]string, 0, len(manual))
	for _, src := range sources {
		for _, raw := range src {
			norm, ok := NormalizeTag(raw)
			if !ok {
				continue
			}
			if _, dup := seen[norm]; dup {
				continue
			}
			seen[norm] = struct{}{}
			out = append(out, norm)
		}
	}
	return out
}
