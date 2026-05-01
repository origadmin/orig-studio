package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type CommentLike struct {
	ID        string    `json:"id"`
	CommentID string    `json:"comment_id"`
	UserID    string    `json:"user_id"`
	LikeType  string    `json:"like_type"`
	CreateTime time.Time `json:"create_time"`
}

type CommentLikeStats struct {
	LikeCount    int64 `json:"like_count"`
	DislikeCount int64 `json:"dislike_count"`
	IsLiked      bool  `json:"is_liked"`
	IsDisliked   bool  `json:"is_disliked"`
}

type CommentLikeRepo interface {
	Create(ctx context.Context, userID, commentID string, likeType string) (*CommentLike, error)
	Delete(ctx context.Context, userID, commentID string) error
	GetStatus(ctx context.Context, userID, commentID string) (string, error)
	CountByComment(ctx context.Context, commentID string, likeType string) (int64, error)
}

type CommentLikeUseCase struct {
	repo CommentLikeRepo
	log  *log.Helper
}

func NewCommentLikeUseCase(
	repo CommentLikeRepo,
	logger log.Logger,
) *CommentLikeUseCase {
	return &CommentLikeUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "comment_like.biz")),
	}
}

func (uc *CommentLikeUseCase) ToggleLike(ctx context.Context, userID, commentID string) (*CommentLikeStats, error) {
	currentStatus, err := uc.repo.GetStatus(ctx, userID, commentID)
	if err != nil {
		return nil, err
	}

	if currentStatus == "like" {
		err = uc.repo.Delete(ctx, userID, commentID)
		if err != nil {
			return nil, err
		}
	} else {
		if currentStatus == "dislike" {
			err = uc.repo.Delete(ctx, userID, commentID)
			if err != nil {
				return nil, err
			}
		}
		_, err = uc.repo.Create(ctx, userID, commentID, "like")
		if err != nil {
			return nil, err
		}
	}

	return uc.GetStats(ctx, userID, commentID)
}

func (uc *CommentLikeUseCase) ToggleDislike(ctx context.Context, userID, commentID string) (*CommentLikeStats, error) {
	currentStatus, err := uc.repo.GetStatus(ctx, userID, commentID)
	if err != nil {
		return nil, err
	}

	if currentStatus == "dislike" {
		err = uc.repo.Delete(ctx, userID, commentID)
		if err != nil {
			return nil, err
		}
	} else {
		if currentStatus == "like" {
			err = uc.repo.Delete(ctx, userID, commentID)
			if err != nil {
				return nil, err
			}
		}
		_, err = uc.repo.Create(ctx, userID, commentID, "dislike")
		if err != nil {
			return nil, err
		}
	}

	return uc.GetStats(ctx, userID, commentID)
}

func (uc *CommentLikeUseCase) GetStats(ctx context.Context, userID, commentID string) (*CommentLikeStats, error) {
	likeCount, _ := uc.repo.CountByComment(ctx, commentID, "like")
	dislikeCount, _ := uc.repo.CountByComment(ctx, commentID, "dislike")

	var isLiked, isDisliked bool
	if userID != "" {
		status, _ := uc.repo.GetStatus(ctx, userID, commentID)
		isLiked = status == "like"
		isDisliked = status == "dislike"
	}

	return &CommentLikeStats{
		LikeCount:    likeCount,
		DislikeCount: dislikeCount,
		IsLiked:      isLiked,
		IsDisliked:   isDisliked,
	}, nil
}
