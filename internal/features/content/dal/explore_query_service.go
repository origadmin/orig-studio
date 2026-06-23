/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Explore query service encapsulates entity.Client usage for the service layer.
 */

package dal

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/media"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// ExploreQueryService provides explore query methods that return DTOs.
type ExploreQueryService struct {
	db *entity.Client
}

// NewExploreQueryService creates a new ExploreQueryService.
func NewExploreQueryService(db *entity.Client) *ExploreQueryService {
	return &ExploreQueryService{db: db}
}

// TrendingItem represents a trending media item.
type TrendingItem struct {
	ID          string `json:"id"`
	ShortToken  string `json:"short_token"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    int    `json:"duration"`
	ViewCount   int64  `json:"view_count"`
	LikeCount   int64  `json:"like_count"`
	PublishedAt string `json:"published_at"`
}

// GetTrending returns trending media items.
func (s *ExploreQueryService) GetTrending(ctx context.Context, limit int) ([]*TrendingItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	medias, err := s.db.Media.Query().
		Limit(limit).
		Order(entity.Desc(media.FieldViewCount)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get trending media: %w", err)
	}

	items := make([]*TrendingItem, 0, len(medias))
	for _, m := range medias {
		publishedAt := ""
		if !m.PublishedAt.IsZero() {
			publishedAt = m.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		items = append(items, &TrendingItem{
			ID:          m.ID,
			ShortToken:  m.ShortToken,
			Title:       m.Title,
			Description: m.Description,
			Thumbnail:   m.Thumbnail,
			Duration:    m.Duration,
			ViewCount:   m.ViewCount,
			LikeCount:   m.LikeCount,
			PublishedAt: publishedAt,
		})
	}

	return items, nil
}

// Ensure dto is used
var _ = (*dto.MediaEntityDTO)(nil)
