/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package dal provides the data access layer for svc-media.
//
// tag_sync.go implements the write-time merge for the unified tag model
// (BUG-131 / BUG-132, strategy γ). The authoritative relationship lives in
// the M2M pivot content_media_tags; content_media.tags (jsonb) is only the
// API serialization projection. Both must be kept in lock-step whenever a
// media is created or updated.
package dal

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/mediatag"
	"origadmin/application/origstudio/internal/data/entity/tag"
	"origadmin/application/origstudio/internal/pkg/hashtag"
)

// resolveTagID returns the content_tags row id for a canonical tag term,
// creating the row on first sight (write-time merge into the unified
// vocabulary). The lookup is case-insensitive on content_tags.title so that
// "4k" reuses the seeded "4K" row instead of creating a duplicate; slugs are
// display URLs only and never used for matching.
func (r *mediaRepo) resolveTagID(ctx context.Context, term string) (int, error) {
	existing, err := r.db.Tag.Query().Where(tag.TitleEqualFold(term)).Only(ctx)
	if err == nil {
		return existing.ID, nil
	}
	if !entity.IsNotFound(err) {
		return 0, fmt.Errorf("lookup tag %q: %w", term, err)
	}

	// BUG-184: give auto-created rows the canonical slug so GET /tags/{slug}
	// resolves (slug is Optional with no default; without it the row only ever
	// matches via the detail-page's raw-title fallback).
	created, err := r.db.Tag.Create().
		SetTitle(term).
		SetSlug(hashtag.GenerateTagSlug(term)).
		SetStatus(tag.StatusACTIVE).
		Save(ctx)
	if err != nil {
		// Raced with another writer: fall back to a re-query.
		existing, err2 := r.db.Tag.Query().Where(tag.TitleEqualFold(term)).Only(ctx)
		if err2 == nil {
			return existing.ID, nil
		}
		return 0, fmt.Errorf("create tag %q: %w", term, err)
	}
	return created.ID, nil
}

// SyncMediaTags reconciles the authoritative M2M rows (content_media_tags)
// to exactly match the canonical merged tag set of a media. Call this AFTER
// the media row itself has been saved (the media id must already exist).
// The edge set is replaced wholesale: stale rows are dropped, new terms get
// their content_tags row created on the fly.
//
// NOTE: the pivot is manipulated directly (MediaTag delete + bulk create)
// instead of Media.UpdateOneID().ClearTagsRel().AddTagsRelIDs(): ent's edge
// mutation rejects re-associating existing MediaTag rows ("already connected
// to a different media_id") when both clear and add run in one mutation.
func (r *mediaRepo) SyncMediaTags(ctx context.Context, mediaID string, merged []string) error {
	ids := make([]int, 0, len(merged))
	for _, term := range merged {
		id, err := r.resolveTagID(ctx, term)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}

	// Drop the old edge rows for this media.
	if _, err := r.db.MediaTag.Delete().Where(mediatag.MediaID(mediaID)).Exec(ctx); err != nil {
		return fmt.Errorf("clear media_tags for %s: %w", mediaID, err)
	}
	if len(ids) == 0 {
		return nil
	}

	// Re-insert the full set. unique(media_id, tag_id) guards duplicates.
	builders := make([]*entity.MediaTagCreate, 0, len(ids))
	for _, id := range ids {
		builders = append(builders, r.db.MediaTag.Create().SetMediaID(mediaID).SetTagID(id))
	}
	if _, err := r.db.MediaTag.CreateBulk(builders...).Save(ctx); err != nil {
		return fmt.Errorf("insert media_tags for %s: %w", mediaID, err)
	}
	return nil
}
