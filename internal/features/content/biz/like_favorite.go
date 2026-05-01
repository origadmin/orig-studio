/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Like represents a user's like on a media.
type Like struct {
	ID        string    `json:"id"`
	MediaID   string    `json:"media_id"`
	UserID    string    `json:"user_id"`
	LikeType  string    `json:"like_type"` // like or dislike
	CreateTime time.Time `json:"create_time"`
}

// Favorite represents a user's favorite on a media.
type Favorite struct {
	ID        string    `json:"id"`
	MediaID   string    `json:"media_id"`
	UserID    string    `json:"user_id"`
	CreateTime time.Time `json:"create_time"`
}

// MediaStats holds counts for likes and favorites.
type MediaStats struct {
	LikeCount     int64  `json:"like_count"`
	DislikeCount  int64  `json:"dislike_count"`
	FavoriteCount int64  `json:"favorite_count"`
	UserLikeType  string `json:"user_like_type"` // none, like, dislike
	IsFavorited   bool   `json:"is_favorited"`
}

// LikeRepo defines storage operations for likes.
type LikeRepo interface {
	Create(ctx context.Context, userID, mediaID string, likeType string) (*Like, error)
	Delete(ctx context.Context, userID, mediaID string) error
	GetStatus(ctx context.Context, userID, mediaID string) (string, error) // returns like, dislike or none
	CountByMedia(ctx context.Context, mediaID string, likeType string) (int64, error)
	ListByUser(ctx context.Context, userID string) ([]*Like, error)
}

// FavoriteRepo defines storage operations for favorites.
type FavoriteRepo interface {
	Create(ctx context.Context, userID, mediaID string) (*Favorite, error)
	Delete(ctx context.Context, userID, mediaID string) error
	IsFavorited(ctx context.Context, userID, mediaID string) (bool, error)
	CountByMedia(ctx context.Context, mediaID string) (int64, error)
	ListByUser(ctx context.Context, userID string) ([]*Favorite, error)
}

// LikeFavoriteUseCase handles likes and favorites business logic.
type LikeFavoriteUseCase struct {
	likeRepo     LikeRepo
	favoriteRepo FavoriteRepo
	mediaUC      MediaUseCaseInterface
	log          *log.Helper
}

func NewLikeFavoriteUseCase(
	likeRepo LikeRepo,
	favoriteRepo FavoriteRepo,
	mediaUC MediaUseCaseInterface,
	logger log.Logger,
) *LikeFavoriteUseCase {
	return &LikeFavoriteUseCase{
		likeRepo:     likeRepo,
		favoriteRepo: favoriteRepo,
		mediaUC:      mediaUC,
		log:          log.NewHelper(log.With(logger, "module", "like_favorite.biz")),
	}
}

func (uc *LikeFavoriteUseCase) ToggleLike(ctx context.Context, userID, mediaID string, likeType string) (*MediaStats, error) {
	currentStatus, err := uc.likeRepo.GetStatus(ctx, userID, mediaID)
	if err != nil {
		return nil, err
	}

	if currentStatus == likeType {
		// Remove existing like/dislike
		err = uc.likeRepo.Delete(ctx, userID, mediaID)
		if err != nil {
			return nil, err
		}
		if likeType == "like" {
			_ = uc.mediaUC.UpdateLikeCount(ctx, mediaID, -1)
		} else {
			_ = uc.mediaUC.UpdateDislikeCount(ctx, mediaID, -1)
		}
	} else {
		// If changing from like to dislike or vice versa
		if currentStatus != "none" {
			err = uc.likeRepo.Delete(ctx, userID, mediaID)
			if err != nil {
				return nil, err
			}
			if currentStatus == "like" {
				_ = uc.mediaUC.UpdateLikeCount(ctx, mediaID, -1)
			} else {
				_ = uc.mediaUC.UpdateDislikeCount(ctx, mediaID, -1)
			}
		}

		// Create new like/dislike
		_, err = uc.likeRepo.Create(ctx, userID, mediaID, likeType)
		if err != nil {
			return nil, err
		}
		if likeType == "like" {
			_ = uc.mediaUC.UpdateLikeCount(ctx, mediaID, 1)
		} else {
			_ = uc.mediaUC.UpdateDislikeCount(ctx, mediaID, 1)
		}
	}

	return uc.GetMediaStats(ctx, userID, mediaID)
}

func (uc *LikeFavoriteUseCase) ToggleFavorite(ctx context.Context, userID, mediaID string) (*MediaStats, error) {
	favorited, err := uc.favoriteRepo.IsFavorited(ctx, userID, mediaID)
	if err != nil {
		return nil, err
	}

	var delta int
	if favorited {
		err = uc.favoriteRepo.Delete(ctx, userID, mediaID)
		delta = -1
	} else {
		_, err = uc.favoriteRepo.Create(ctx, userID, mediaID)
		delta = 1
	}

	if err != nil {
		return nil, err
	}

	// Update media favorite count
	_ = uc.mediaUC.UpdateFavoriteCount(ctx, mediaID, delta)

	return uc.GetMediaStats(ctx, userID, mediaID)
}

func (uc *LikeFavoriteUseCase) GetMediaStats(ctx context.Context, userID, mediaID string) (*MediaStats, error) {
	likeCount, _ := uc.likeRepo.CountByMedia(ctx, mediaID, "like")
	dislikeCount, _ := uc.likeRepo.CountByMedia(ctx, mediaID, "dislike")
	favoriteCount, _ := uc.favoriteRepo.CountByMedia(ctx, mediaID)

	userLikeType := "none"
	var isFavorited bool
	if userID != "" {
		userLikeType, _ = uc.likeRepo.GetStatus(ctx, userID, mediaID)
		isFavorited, _ = uc.favoriteRepo.IsFavorited(ctx, userID, mediaID)
	}

	return &MediaStats{
		LikeCount:     likeCount,
		DislikeCount:  dislikeCount,
		FavoriteCount: favoriteCount,
		UserLikeType:  userLikeType,
		IsFavorited:   isFavorited,
	}, nil
}

func (uc *LikeFavoriteUseCase) ListUserFavorites(ctx context.Context, userID string) ([]*Favorite, error) {
	return uc.favoriteRepo.ListByUser(ctx, userID)
}

func (uc *LikeFavoriteUseCase) ListUserLikes(ctx context.Context, userID string) ([]*Like, error) {
	return uc.likeRepo.ListByUser(ctx, userID)
}
