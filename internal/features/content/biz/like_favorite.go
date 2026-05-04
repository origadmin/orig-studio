/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
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
	ID         string      `json:"id"`
	MediaID    string      `json:"media_id"`
	UserID     string      `json:"user_id"`
	CreateTime time.Time   `json:"create_time"`
	Media      *FavoriteMedia `json:"media,omitempty"`
}

// FavoriteMedia holds the media details embedded in a favorite response.
type FavoriteMedia struct {
	ID          string              `json:"id"`
	ShortToken  string              `json:"short_token"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Thumbnail   string              `json:"thumbnail"`
	Duration    int64               `json:"duration"`
	ViewCount   int64               `json:"view_count"`
	Type        string              `json:"type"`
	UserID      string              `json:"user_id"`
	CreateTime  string              `json:"create_time"`
	Edges       *FavoriteMediaEdges `json:"edges,omitempty"`
}

// FavoriteMediaEdges holds the edge data for FavoriteMedia.
type FavoriteMediaEdges struct {
	User []FavoriteMediaUser `json:"user,omitempty"`
}

// FavoriteMediaUser holds user info for the media edge.
type FavoriteMediaUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
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
	DeleteByID(ctx context.Context, id string) error
	IsFavorited(ctx context.Context, userID, mediaID string) (bool, error)
	CountByMedia(ctx context.Context, mediaID string) (int64, error)
	ListByUser(ctx context.Context, userID string) ([]*Favorite, error)
	ListByUserPaginated(ctx context.Context, userID string, page, pageSize int) ([]*Favorite, int, error)
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

// resolveMediaID resolves a short_token to internal media ID.
// If resolution fails, returns the original idOrToken as-is.
func (uc *LikeFavoriteUseCase) resolveMediaID(ctx context.Context, idOrToken string) string {
	resolved, err := uc.mediaUC.ResolveToID(ctx, idOrToken)
	if err != nil {
		return idOrToken
	}
	return resolved
}

func (uc *LikeFavoriteUseCase) ToggleLike(ctx context.Context, userID, mediaID string, likeType string) (*MediaStats, error) {
	mediaID = uc.resolveMediaID(ctx, mediaID)
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
	mediaID = uc.resolveMediaID(ctx, mediaID)
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
	mediaID = uc.resolveMediaID(ctx, mediaID)
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

// ListUserFavoritesPaginated returns a paginated list of user favorites.
func (uc *LikeFavoriteUseCase) ListUserFavoritesPaginated(ctx context.Context, userID string, page, pageSize int) ([]*Favorite, int, error) {
	return uc.favoriteRepo.ListByUserPaginated(ctx, userID, page, pageSize)
}

// RemoveFavoriteByID removes a favorite by its ID directly.
// It also updates the media favorite count.
func (uc *LikeFavoriteUseCase) RemoveFavoriteByID(ctx context.Context, userID, favoriteID string) error {
	// First get the favorite to find the mediaID for count update
	favorites, err := uc.favoriteRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	var mediaID string
	found := false
	for _, fav := range favorites {
		if fav.ID == favoriteID {
			mediaID = fav.MediaID
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("favorite not found")
	}

	// Delete by ID directly
	if err := uc.favoriteRepo.DeleteByID(ctx, favoriteID); err != nil {
		return err
	}

	// Update media favorite count
	_ = uc.mediaUC.UpdateFavoriteCount(ctx, mediaID, -1)

	return nil
}

func (uc *LikeFavoriteUseCase) ListUserLikes(ctx context.Context, userID string) ([]*Like, error) {
	return uc.likeRepo.ListByUser(ctx, userID)
}
