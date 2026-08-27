/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * BUG-143 regression: tags created through the portal endpoint carried only a
 * title, leaving slug NULL. Those rows 404'd on GET /api/v1/tags/{slug} and so
 * could not back the dedicated /tag/{slug} portal page.
 */

package biz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

// fakeTagRepo is an in-memory TagRepo good enough to exercise slug generation
// and collision probing.
type fakeTagRepo struct {
	bySlug  map[string]*Tag
	created []*Tag
	updated []*Tag
	nextID  int
}

func newFakeTagRepo(existing ...*Tag) *fakeTagRepo {
	r := &fakeTagRepo{bySlug: map[string]*Tag{}, nextID: 100}
	for _, t := range existing {
		r.bySlug[t.Slug] = t
	}
	return r
}

func (r *fakeTagRepo) Create(_ context.Context, t *Tag) (*Tag, error) {
	r.nextID++
	t.ID = r.nextID
	r.created = append(r.created, t)
	r.bySlug[t.Slug] = t
	return t, nil
}

func (r *fakeTagRepo) Get(_ context.Context, id int) (*Tag, error) {
	for _, t := range r.bySlug {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *fakeTagRepo) GetByName(_ context.Context, name string) (*Tag, error) {
	for _, t := range r.bySlug {
		if t.Title == name {
			return t, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *fakeTagRepo) GetBySlug(_ context.Context, slug string) (*Tag, error) {
	if t, ok := r.bySlug[slug]; ok {
		return t, nil
	}
	return nil, errors.New("not found")
}

func (r *fakeTagRepo) Update(_ context.Context, t *Tag) (*Tag, error) {
	r.updated = append(r.updated, t)
	r.bySlug[t.Slug] = t
	return t, nil
}

func (r *fakeTagRepo) Delete(_ context.Context, _ int) error { return nil }

func (r *fakeTagRepo) ListAll(_ context.Context, _, _ int) ([]*Tag, int, error) {
	return nil, 0, nil
}

func newTagSlugTestUseCase(repo TagRepo) *CategoryTagUseCase {
	return NewCategoryTagUseCase(nil, repo, log.DefaultLogger)
}

func TestCreateTagGeneratesSlugWhenMissing(t *testing.T) {
	cases := []struct {
		name  string
		title string
		want  string
	}{
		{"ascii word", "golang", "golang"},
		{"ascii digits", "3", "3"},
		{"ascii with spaces", "Portal Fix", "portal-fix"},
		{"ascii mixed case", "TAG1", "tag1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeTagRepo()
			uc := newTagSlugTestUseCase(repo)

			got, err := uc.CreateTag(context.Background(), &Tag{Title: tc.title})
			if err != nil {
				t.Fatalf("CreateTag returned error: %v", err)
			}
			if got.Slug != tc.want {
				t.Fatalf("slug = %q, want %q", got.Slug, tc.want)
			}
		})
	}
}

func TestCreateTagGeneratesNonEmptySlugForNonASCIITitle(t *testing.T) {
	repo := newFakeTagRepo()
	uc := newTagSlugTestUseCase(repo)

	got, err := uc.CreateTag(context.Background(), &Tag{Title: "视频"})
	if err != nil {
		t.Fatalf("CreateTag returned error: %v", err)
	}
	// Non-ASCII titles are Base58-encoded; we only require a usable, URL-safe slug.
	if got.Slug == "" {
		t.Fatal("slug is empty for a non-ASCII title")
	}
	if strings.ContainsAny(got.Slug, "/? #%") {
		t.Fatalf("slug %q contains characters that are unsafe in a URL path", got.Slug)
	}
}

func TestCreateTagRespectsExplicitSlug(t *testing.T) {
	repo := newFakeTagRepo()
	uc := newTagSlugTestUseCase(repo)

	got, err := uc.CreateTag(context.Background(), &Tag{Title: "Anything", Slug: "custom-slug"})
	if err != nil {
		t.Fatalf("CreateTag returned error: %v", err)
	}
	if got.Slug != "custom-slug" {
		t.Fatalf("slug = %q, want the caller-provided %q", got.Slug, "custom-slug")
	}
}

func TestCreateTagResolvesSlugCollision(t *testing.T) {
	repo := newFakeTagRepo(&Tag{ID: 1, Title: "Golang", Slug: "golang"})
	uc := newTagSlugTestUseCase(repo)

	got, err := uc.CreateTag(context.Background(), &Tag{Title: "golang"})
	if err != nil {
		t.Fatalf("CreateTag returned error: %v", err)
	}
	if got.Slug != "golang-2" {
		t.Fatalf("slug = %q, want %q (collision must be suffixed)", got.Slug, "golang-2")
	}
}

func TestUpdateTagBackfillsLegacyNullSlug(t *testing.T) {
	legacy := &Tag{ID: 18, Title: "3", Slug: ""}
	repo := newFakeTagRepo()
	uc := newTagSlugTestUseCase(repo)

	got, err := uc.UpdateTag(context.Background(), legacy)
	if err != nil {
		t.Fatalf("UpdateTag returned error: %v", err)
	}
	if got.Slug != "3" {
		t.Fatalf("slug = %q, want %q (legacy row must be backfilled)", got.Slug, "3")
	}
}

func TestUpdateTagKeepsOwnSlugWithoutSuffixing(t *testing.T) {
	// The tag already owns "golang"; re-saving it must not bump to "golang-2".
	self := &Tag{ID: 7, Title: "golang", Slug: ""}
	repo := newFakeTagRepo(&Tag{ID: 7, Title: "golang", Slug: "golang"})
	uc := newTagSlugTestUseCase(repo)

	got, err := uc.UpdateTag(context.Background(), self)
	if err != nil {
		t.Fatalf("UpdateTag returned error: %v", err)
	}
	if got.Slug != "golang" {
		t.Fatalf("slug = %q, want %q (owning tag must keep its slug)", got.Slug, "golang")
	}
}
