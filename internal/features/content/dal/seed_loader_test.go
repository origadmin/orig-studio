/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import "testing"

// TestParseTaxonomyDefault verifies the embedded taxonomy parses to the agreed
// 3-level shape (BUG-162, 2026-08-26 最终定稿「root → L2 分类 → L3 子分类」):
//   - 3 module roots (video/music/article)
//   - video root carries exactly 15 L2 classification children (real content
//     categories; the form/genre axis abstraction was dropped 2026-08-26),
//     each expanding ≥1 L3 leaf
//   - 50 L3 leaves total; no "影视" (film_tv) umbrella node
// This is the regression guard that prevents the silent 2-layer/3-layer
// oscillation seen on 2026-08-19~08-25.
func TestParseTaxonomyDefault(t *testing.T) {
	spec, err := parseTaxonomy(defaultTaxonomyYAML)
	if err != nil {
		t.Fatalf("parse embedded taxonomy: %v", err)
	}

	if len(spec.Categories) != 3 {
		t.Fatalf("want 3 module roots (video/music/article), got %d", len(spec.Categories))
	}
	if spec.Categories[0].Slug != "video" {
		t.Fatalf("first root should be video, got %q", spec.Categories[0].Slug)
	}

	video := spec.Categories[0]
	// 3-level: video has exactly 15 L2 classification children (no flat leaves at L2).
	const wantL2 = 15
	if len(video.Children) != wantL2 {
		t.Fatalf("want %d video L2 classification children, got %d", wantL2, len(video.Children))
	}

	// Every L2 classification must expand into ≥1 L3 leaf.
	totalLeaves := 0
	for _, cat := range video.Children {
		if len(cat.Children) == 0 {
			t.Errorf("3-level: L2 %q must expand ≥1 L3 leaf, got 0", cat.Slug)
			continue
		}
		totalLeaves += len(cat.Children)
	}
	const wantL3 = 50
	if totalLeaves != wantL3 {
		t.Errorf("want %d L3 leaves total, got %d", wantL3, totalLeaves)
	}

	for _, want := range []string{"drama", "movie", "variety", "anime", "mv", "documentary", "gaming", "sports", "entertainment", "tech", "tutorial", "lifestyle", "promo", "ugc", "other"} {
		if !hasChild(video, want) {
			t.Errorf("missing L2 classification %q", want)
		}
	}

	// The deprecated "影视" (film_tv) umbrella must no longer exist — its
	// children were redistributed to drama/movie/variety (BUG-162 2026-08-26).
	if hasChild(video, "film_tv") {
		t.Error("film_tv umbrella must be removed (redistributed to drama/movie/variety)")
	}

	if len(spec.Tags) != 10 {
		t.Errorf("want 10 tags, got %d", len(spec.Tags))
	}
}

// hasChild reports whether a node has a direct child with the given slug.
func hasChild(node *catSpecNode, slug string) bool {
	for _, c := range node.Children {
		if c.Slug == slug {
			return true
		}
	}
	return false
}

// TestParseTaxonomyRejectsNonASCII enforces the §3.7 stable-key rule: a
// non-ASCII slug must fail at parse time (startup), never at runtime.
func TestParseTaxonomyRejectsNonASCII(t *testing.T) {
	bad := []byte("categories:\n  - name: 测试\n    slug: 视频\n    sequence: 1\n")
	if _, err := parseTaxonomy(bad); err == nil {
		t.Fatal("want error for non-ASCII category slug, got nil")
	}

	badTag := []byte("categories:\n  - name: 视频\n    slug: video\n    sequence: 0\n" +
		"tags:\n  - title: 测试\n    slug: 标签\n")
	if _, err := parseTaxonomy(badTag); err == nil {
		t.Fatal("want error for non-ASCII tag slug, got nil")
	}
}

// TestParseTaxonomyEmptyRejected guards against an accidentally empty taxonomy
// (which would silently seed nothing).
func TestParseTaxonomyEmptyRejected(t *testing.T) {
	if _, err := parseTaxonomy([]byte("categories: []\n")); err == nil {
		t.Fatal("want error for empty categories, got nil")
	}
}
