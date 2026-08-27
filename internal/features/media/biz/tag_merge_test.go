/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * Unit tests for the unified tag merge (NormalizeAndMerge / NormalizeTag /
 * ExtractHashtags). These encode the acceptance matrix the user requested:
 * three sources (manual / title-hashtag / intro-hashtag) × hit / miss, plus the
 * normalization edge cases (case folding, '#'-prefix, control chars, unicode).
 */

package biz

import (
	"reflect"
	"testing"
)

// contains reports whether want is present in got (order-independent check for a
// single tag). Kept local to avoid depending on test helpers that may not exist.
func contains(got []string, want string) bool {
	for _, g := range got {
		if g == want {
			return true
		}
	}
	return false
}

func TestNormalizeTag(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"plain", "golang", "golang", true},
		{"uppercase folds", "Golang", "golang", true},
		{"leading hash", "#golang", "golang", true},
		{"space padded hash", "  #Golang  ", "golang", true},
		{"control char stripped", "go\rlang", "golang", true},
		{"internal space dropped", "go lang", "golang", true},
		{"empty", "", "", false},
		{"whitespace only", "   ", "", false},
		{"hash only", "#", "", false},
		{"leading dash dropped", "-tag", "tag", true},
		{"unicode cjk", "教程", "教程", true},
		{"unicode mixed", "#Go语言", "go语言", true},
		{"invalid symbols dropped", "go@lang!", "golang", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := NormalizeTag(c.in)
			if ok != c.ok {
				t.Fatalf("NormalizeTag(%q) ok = %v, want %v", c.in, ok, c.ok)
			}
			if got != c.want {
				t.Fatalf("NormalizeTag(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestExtractHashtags(t *testing.T) {
	// NOTE: '+' is not in the frontend hashtag class [\p{L}\p{N}_-], so "#C++"
	// extracts only "C" — this intentionally mirrors web/src/lib/utils/hashtag.ts
	// (HASHTAG_REGEX) to keep front/back behaviour in lock-step. Extraction is
	// case-preserving; lowercasing happens later in NormalizeTag/NormalizeAndMerge.
	got := ExtractHashtags("Learn #golang and #C++ basics in #教程")
	want := []string{"golang", "C", "教程"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExtractHashtags = %v, want %v", got, want)
	}
	if ExtractHashtags("") != nil {
		t.Fatalf("ExtractHashtags('') should be nil")
	}
	if ExtractHashtags("no tags here") != nil {
		t.Fatalf("ExtractHashtags('no tags here') should be nil")
	}
}

// --- Source: TITLE hashtag ---

func TestMerge_Title_Hit(t *testing.T) {
	got := NormalizeAndMerge("Intro to #golang", "", nil)
	if !contains(got, "golang") {
		t.Fatalf("title hashtag not merged: %v", got)
	}
}

func TestMerge_Title_Miss(t *testing.T) {
	got := NormalizeAndMerge("Intro to programming", "", nil)
	if contains(got, "golang") {
		t.Fatalf("unexpected tag from title: %v", got)
	}
}

// --- Source: INTRO (description) hashtag ---

func TestMerge_Intro_Hit(t *testing.T) {
	got := NormalizeAndMerge("", "A #tutorial about loops", nil)
	if !contains(got, "tutorial") {
		t.Fatalf("intro hashtag not merged: %v", got)
	}
}

func TestMerge_Intro_Miss(t *testing.T) {
	got := NormalizeAndMerge("", "A video about loops", nil)
	if contains(got, "tutorial") {
		t.Fatalf("unexpected tag from intro: %v", got)
	}
}

// --- Source: MANUAL tag ---

func TestMerge_Manual_Hit(t *testing.T) {
	got := NormalizeAndMerge("", "", []string{"golang"})
	if !contains(got, "golang") {
		t.Fatalf("manual tag not merged: %v", got)
	}
}

func TestMerge_Manual_Miss(t *testing.T) {
	got := NormalizeAndMerge("", "", []string{"rust"})
	if contains(got, "golang") {
		t.Fatalf("unexpected manual tag: %v", got)
	}
}

// --- Cross-source dedup + case folding ---

func TestMerge_CrossSourceDedup(t *testing.T) {
	// manual + title + intro all carry "golang" (different cases) → single entry.
	got := NormalizeAndMerge("Watch #Golang", "See #GOLANG now", []string{"golang"})
	if len(got) != 1 || got[0] != "golang" {
		t.Fatalf("cross-source dedup failed: %v (len %d)", got, len(got))
	}
}

func TestMerge_CaseFold(t *testing.T) {
	got := NormalizeAndMerge("", "", []string{"Golang", "golang"})
	if len(got) != 1 || got[0] != "golang" {
		t.Fatalf("case folding failed: %v", got)
	}
}

func TestMerge_HashPrefixInManual(t *testing.T) {
	got := NormalizeAndMerge("", "", []string{"#golang", "golang"})
	if len(got) != 1 || got[0] != "golang" {
		t.Fatalf("'#golang' and 'golang' should collapse: %v", got)
	}
}

// --- All three sources together ---

func TestMerge_AllThreeSources(t *testing.T) {
	got := NormalizeAndMerge(
		"#titleTag",
		"desc with #introTag",
		[]string{"manualTag"},
	)
	for _, want := range []string{"manualtag", "titletag", "introtag"} {
		if !contains(got, want) {
			t.Fatalf("missing %q in %v", want, got)
		}
	}
}

// --- Empty / degenerate inputs ---

func TestMerge_AllEmpty(t *testing.T) {
	got := NormalizeAndMerge("", "", nil)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

func TestMerge_IgnoresJunkManual(t *testing.T) {
	got := NormalizeAndMerge("", "", []string{"", "  ", "#", "@@@"})
	if len(got) != 0 {
		t.Fatalf("junk manual tags should be dropped, got %v", got)
	}
}

// --- Parity with frontend mergeTagsWithHashtags ordering ---
// Frontend order: existingTags(manual) → title → description; first occurrence
// wins. With lowercase canonicalization the first occurrence is the lowercased
// manual entry.
func TestMerge_OrderParity(t *testing.T) {
	got := NormalizeAndMerge("#Beta", "#Gamma", []string{"alpha"})
	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}
