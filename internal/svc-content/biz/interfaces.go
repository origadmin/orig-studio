/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import "context"

// MediaUseCaseInterface defines the minimal interface needed from MediaUseCase.
type MediaUseCaseInterface interface {
	CheckMedia(ctx context.Context, id int64) error
	UpdateCommentCount(ctx context.Context, id int64, delta int) error
	UpdateLikeCount(ctx context.Context, id int64, delta int) error
	UpdateDislikeCount(ctx context.Context, id int64, delta int) error
	UpdateFavoriteCount(ctx context.Context, id int64, delta int) error
}

// MediaInfo is a minimal media record for dependency injection and feeds.
type MediaInfo struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    int    `json:"duration"`
	ViewCount   int64  `json:"view_count"`
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Type        string `json:"type"`
	URL         string `json:"url"`
}
