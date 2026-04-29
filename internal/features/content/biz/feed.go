/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// FeedRepo defines storage operations for feeds.
type FeedRepo interface {
	ListLatest(ctx context.Context, page, pageSize int) ([]*MediaInfo, int, error)
	ListTrending(ctx context.Context, page, pageSize int) ([]*MediaInfo, int, error)
	ListFeatured(ctx context.Context, page, pageSize int) ([]*MediaInfo, int, error)
}

// FeedUseCase handles high-level feed logic.
type FeedUseCase struct {
	repo FeedRepo
	log  *log.Helper
}

func NewFeedUseCase(repo FeedRepo, logger log.Logger) *FeedUseCase {
	return &FeedUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "feed.biz")),
	}
}

func (uc *FeedUseCase) GetHomeFeed(ctx context.Context, page, pageSize int) ([]*MediaInfo, int, error) {
	return uc.repo.ListLatest(ctx, page, pageSize)
}

func (uc *FeedUseCase) ListLatest(ctx context.Context, page, pageSize int) ([]*MediaInfo, int, error) {
	return uc.repo.ListLatest(ctx, page, pageSize)
}

func (uc *FeedUseCase) GetTrendingFeed(ctx context.Context, page, pageSize int) ([]*MediaInfo, int, error) {
	return uc.repo.ListTrending(ctx, page, pageSize)
}
