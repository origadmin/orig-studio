/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"time"

	"origadmin/application/origcms/internal/data/entity"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Comment represents a comment on a media (biz layer).
type Comment struct {
	ID        int       `json:"id"`
	UID       uuid.UUID `json:"uid"`
	Text      string    `json:"text"`
	MediaID   int       `json:"media_id"`
	UserID    int       `json:"user_id"`
	ParentID  *int      `json:"parent_id,omitempty"`
	AddDate   time.Time `json:"add_date"`
	UpdatedAt time.Time `json:"updated_at"`

	// Edges
	User    *entity.User `json:"user,omitempty"`
	Replies []*Comment   `json:"replies,omitempty"`
}

// CommentRepo defines the storage operations for comments.
type CommentRepo interface {
	Create(ctx context.Context, comment *Comment) (*Comment, error)
	Get(ctx context.Context, id int) (*Comment, error)
	Update(ctx context.Context, comment *Comment) (*Comment, error)
	Delete(ctx context.Context, id int) error
	ListByMedia(ctx context.Context, mediaID int, page, pageSize int) ([]*Comment, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*Comment, int, error)
}

// CommentUseCase handles comment business logic.
type CommentUseCase struct {
	repo    CommentRepo
	mediaUC MediaUseCaseInterface // Use interface to avoid circular dependency
	log     *log.Helper
}

func NewCommentUseCase(repo CommentRepo, mediaUC MediaUseCaseInterface, logger log.Logger) *CommentUseCase {
	return &CommentUseCase{
		repo:    repo,
		mediaUC: mediaUC,
		log:     log.NewHelper(log.With(logger, "module", "comment.biz")),
	}
}

func (uc *CommentUseCase) CreateComment(ctx context.Context, c *Comment) (*Comment, error) {
	if c.Text == "" {
		return nil, fmt.Errorf("comment text cannot be empty")
	}

	// Verify media exists
	if err := uc.mediaUC.CheckMedia(ctx, int64(c.MediaID)); err != nil {
		return nil, fmt.Errorf("media not found: %w", err)
	}

	created, err := uc.repo.Create(ctx, c)
	if err != nil {
		return nil, err
	}

	// Update media comment count
	_ = uc.mediaUC.UpdateCommentCount(ctx, int64(c.MediaID), 1)

	return created, nil
}

func (uc *CommentUseCase) UpdateComment(ctx context.Context, id int, userID int, isAdmin bool, text string) (*Comment, error) {
	comment, err := uc.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if comment.UserID != userID && !isAdmin {
		return nil, fmt.Errorf("permission denied")
	}

	comment.Text = text
	return uc.repo.Update(ctx, comment)
}

func (uc *CommentUseCase) DeleteComment(ctx context.Context, id int, userID int, isAdmin bool) error {
	comment, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if comment.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}

	err = uc.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Update media comment count
	_ = uc.mediaUC.UpdateCommentCount(ctx, int64(comment.MediaID), -1)

	return nil
}

func (uc *CommentUseCase) ListMediaComments(ctx context.Context, mediaID int, page, pageSize int) ([]*Comment, int, error) {
	return uc.repo.ListByMedia(ctx, mediaID, page, pageSize)
}

func (uc *CommentUseCase) GetComment(ctx context.Context, id int) (*Comment, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *CommentUseCase) ListAll(ctx context.Context, page, pageSize int) ([]*Comment, int, error) {
	return uc.repo.ListAll(ctx, page, pageSize)
}
