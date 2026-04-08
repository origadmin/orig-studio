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
	ID        int       `json:"id"`
	MediaID   int       `json:"media_id"`
	UserID    int       `json:"user_id"`
	LikeType  string    `json:"like_type"` // like or dislike
	CreatedAt time.Time `json:"created_at"`
}

// Favorite represents a user's favorite on a media.
type Favorite struct {
	ID        int       `json:"id"`
	MediaID   int       `json:"media_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
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
	Create(ctx context.Context, userID, mediaID int, likeType string) (*Like, error)
	Delete(ctx context.Context, userID, mediaID int) error
	GetStatus(ctx context.Context, userID, mediaID int) (string, error) // returns like, dislike or none
	CountByMedia(ctx context.Context, mediaID int, likeType string) (int64, error)
}

// FavoriteRepo defines storage operations for favorites.
type FavoriteRepo interface {
	Create(ctx context.Context, userID, mediaID int) (*Favorite, error)
	Delete(ctx context.Context, userID, mediaID int) error
	IsFavorited(ctx context.Context, userID, mediaID int) (bool, error)
	CountByMedia(ctx context.Context, mediaID int) (int64, error)
	ListByUser(ctx context.Context, userID int) ([]*Favorite, error)
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

func (uc *LikeFavoriteUseCase) ToggleLike(ctx context.Context, userID, mediaID int, likeType string) (*MediaStats, error) {
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
			_ = uc.mediaUC.UpdateLikeCount(ctx, int64(mediaID), -1)
		} else {
			_ = uc.mediaUC.UpdateDislikeCount(ctx, int64(mediaID), -1)
		}
	} else {
		// If changing from like to dislike or vice versa
		if currentStatus != "none" {
			err = uc.likeRepo.Delete(ctx, userID, mediaID)
			if err != nil {
				return nil, err
			}
			if currentStatus == "like" {
				_ = uc.mediaUC.UpdateLikeCount(ctx, int64(mediaID), -1)
			} else {
				_ = uc.mediaUC.UpdateDislikeCount(ctx, int64(mediaID), -1)
			}
		}

		// Create new like/dislike
		_, err = uc.likeRepo.Create(ctx, userID, mediaID, likeType)
		if err != nil {
			return nil, err
		}
		if likeType == "like" {
			_ = uc.mediaUC.UpdateLikeCount(ctx, int64(mediaID), 1)
		} else {
			_ = uc.mediaUC.UpdateDislikeCount(ctx, int64(mediaID), 1)
		}
	}

	return uc.GetMediaStats(ctx, userID, mediaID)
}

func (uc *LikeFavoriteUseCase) ToggleFavorite(ctx context.Context, userID, mediaID int) (*MediaStats, error) {
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
	_ = uc.mediaUC.UpdateFavoriteCount(ctx, int64(mediaID), delta)

	return uc.GetMediaStats(ctx, userID, mediaID)
}

func (uc *LikeFavoriteUseCase) GetMediaStats(ctx context.Context, userID, mediaID int) (*MediaStats, error) {
	likeCount, _ := uc.likeRepo.CountByMedia(ctx, mediaID, "like")
	dislikeCount, _ := uc.likeRepo.CountByMedia(ctx, mediaID, "dislike")
	favoriteCount, _ := uc.favoriteRepo.CountByMedia(ctx, mediaID)

	userLikeType := "none"
	var isFavorited bool
	if userID > 0 {
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

func (uc *LikeFavoriteUseCase) ListUserFavorites(ctx context.Context, userID int) ([]*Favorite, error) {
	return uc.favoriteRepo.ListByUser(ctx, userID)
}
