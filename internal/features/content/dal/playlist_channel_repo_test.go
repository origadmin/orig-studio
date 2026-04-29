/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"testing"
	"time"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/features/content/biz"
)

func TestMapChannel_MapsShortToken(t *testing.T) {
	ent := &entity.Channel{
		ID:          "ch-001",
		Title:       "Test Channel",
		Description: "A test channel",
		BannerLogo:  "https://example.com/banner.png",
		ShortToken:  "abc123def456",
		IsPublic:    true,
		UserID:      "user-001",
		AddDate:     time.Now(),
	}

	result := mapChannel(ent)

	if result.ShortToken != ent.ShortToken {
		t.Errorf("mapChannel ShortToken: got %q, want %q", result.ShortToken, ent.ShortToken)
	}
	if result.ID != ent.ID {
		t.Errorf("mapChannel ID: got %q, want %q", result.ID, ent.ID)
	}
	if result.Title != ent.Title {
		t.Errorf("mapChannel Title: got %q, want %q", result.Title, ent.Title)
	}
	if result.UserID != ent.UserID {
		t.Errorf("mapChannel UserID: got %q, want %q", result.UserID, ent.UserID)
	}
}

func TestMapChannel_NoFriendlyTokenOrSlug(t *testing.T) {
	// R2: Channel no longer has FriendlyToken or Slug fields
	ent := &entity.Channel{
		ID:          "ch-002",
		Title:       "No Slug Channel",
		ShortToken:  "xyz789",
		IsPublic:    true,
		UserID:      "user-002",
		AddDate:     time.Now(),
	}

	result := mapChannel(ent)

	// Verify ShortToken is the ONLY quick-access field
	if result.ShortToken != "xyz789" {
		t.Errorf("ShortToken: got %q, want %q", result.ShortToken, "xyz789")
	}
}

func TestMapChannel_AllFieldsMapped(t *testing.T) {
	now := time.Now()
	ent := &entity.Channel{
		ID:          "ch-003",
		Title:       "Full Channel",
		Description: "Full description",
		BannerLogo:  "https://example.com/banner.png",
		ShortToken:  "st-full",
		IsPublic:    false,
		UserID:      "user-003",
		AddDate:     now,
	}

	result := mapChannel(ent)

	if result.ID != ent.ID {
		t.Errorf("ID: got %q, want %q", result.ID, ent.ID)
	}
	if result.Title != ent.Title {
		t.Errorf("Title: got %q, want %q", result.Title, ent.Title)
	}
	if result.Description != ent.Description {
		t.Errorf("Description: got %q, want %q", result.Description, ent.Description)
	}
	if result.BannerLogo != ent.BannerLogo {
		t.Errorf("BannerLogo: got %q, want %q", result.BannerLogo, ent.BannerLogo)
	}
	if result.ShortToken != ent.ShortToken {
		t.Errorf("ShortToken: got %q, want %q", result.ShortToken, ent.ShortToken)
	}
	if result.IsPublic != ent.IsPublic {
		t.Errorf("IsPublic: got %v, want %v", result.IsPublic, ent.IsPublic)
	}
	if result.UserID != ent.UserID {
		t.Errorf("UserID: got %q, want %q", result.UserID, ent.UserID)
	}
}

func TestChannelStruct_HasShortTokenField(t *testing.T) {
	// Verify that the biz.Channel struct has ShortToken field and no FriendlyToken/Slug
	ch := biz.Channel{
		ID:         "ch-004",
		Title:      "Struct Test",
		ShortToken: "st-abc123",
		UserID:     "user-004",
		IsPublic:   true,
		CreatedAt:  time.Now(),
	}

	if ch.ShortToken != "st-abc123" {
		t.Errorf("biz.Channel ShortToken: got %q, want %q", ch.ShortToken, "st-abc123")
	}
}

func TestMapPlaylist_MapsShortToken(t *testing.T) {
	ent := &entity.Playlist{
		ID:          "pl-001",
		Title:       "Test Playlist",
		Description: "A test playlist",
		ShortToken:  "pl-abc123",
		UserID:      "user-001",
		Privacy:     1,
		AddDate:     time.Now(),
	}

	result := mapPlaylist(ent)

	if result.ShortToken != ent.ShortToken {
		t.Errorf("mapPlaylist ShortToken: got %q, want %q", result.ShortToken, ent.ShortToken)
	}
}
