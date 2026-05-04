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

// History represents a watch history record.
type History struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	ContentID       string    `json:"content_id"`
	ContentType     string    `json:"content_type"` // video, article, audio
	ProgressSeconds int       `json:"progress_seconds"`
	DurationSeconds int       `json:"duration_seconds"`
	IsFinished      bool      `json:"is_finished"`
	LastWatchedAt   time.Time `json:"last_watched_at"`
	CreateTime      time.Time `json:"create_time"`
	UpdateTime      time.Time `json:"update_time"`

	Title      string `json:"title,omitempty"`
	Thumbnail  string `json:"thumbnail,omitempty"`
	ShortToken string `json:"short_token,omitempty"`
	Deleted    bool   `json:"deleted,omitempty"`
}

// HistoryRepo defines storage operations for watch history.
type HistoryRepo interface {
	// Upsert creates or updates a history record.
	// If a record exists for (user_id, content_id, content_type), it updates
	// progress to max(old, new) and last_watched_at to now.
	Upsert(ctx context.Context, h *History) (*History, error)

	// List retrieves paginated history for a user.
	List(ctx context.Context, userID string, contentType string, page, pageSize int) ([]*History, int, error)

	// GetByUserContent retrieves a specific history record.
	GetByUserContent(ctx context.Context, userID, contentID, contentType string) (*History, error)

	// Delete removes a single history record by ID.
	Delete(ctx context.Context, id string) error

	// DeleteAll removes all history records for a user.
	DeleteAll(ctx context.Context, userID string) (int, error)

	// Sync merges local history items with server records.
	// For each item: upsert with merge logic (max progress, latest time).
	// Returns the complete merged list.
	Sync(ctx context.Context, userID string, items []*History) ([]*History, int, error)

	// CountByUser returns the total number of history records for a user.
	CountByUser(ctx context.Context, userID string) (int, error)

	// DeleteOldest removes the oldest N history records for a user.
	DeleteOldest(ctx context.Context, userID string, n int) error
}

// HistoryUseCase handles watch history business logic.
type HistoryUseCase struct {
	repo     HistoryRepo
	log      *log.Helper
	maxItems int // Max history items per user (default: 500)
}

// NewHistoryUseCase creates a new HistoryUseCase.
func NewHistoryUseCase(repo HistoryRepo, logger log.Logger) *HistoryUseCase {
	return &HistoryUseCase{
		repo:     repo,
		log:      log.NewHelper(log.With(logger, "module", "history.biz")),
		maxItems: 500,
	}
}

// Upsert creates or updates a history record.
// Enforces max items limit by deleting oldest records when exceeded.
func (uc *HistoryUseCase) Upsert(ctx context.Context, h *History) (*History, error) {
	// Validate content_type
	if h.ContentType != "video" && h.ContentType != "article" && h.ContentType != "audio" {
		return nil, fmt.Errorf("invalid content_type: %s", h.ContentType)
	}

	// Calculate is_finished
	if h.DurationSeconds > 0 {
		threshold := 0.9
		if h.ContentType == "article" {
			threshold = 0.95
		}
		h.IsFinished = float64(h.ProgressSeconds) >= float64(h.DurationSeconds)*threshold
	}

	result, err := uc.repo.Upsert(ctx, h)
	if err != nil {
		return nil, err
	}

	// Enforce max items limit
	count, _ := uc.repo.CountByUser(ctx, h.UserID)
	if count > uc.maxItems {
		overflow := count - uc.maxItems
		_ = uc.repo.DeleteOldest(ctx, h.UserID, overflow)
	}

	return result, nil
}

// List retrieves paginated history for a user.
func (uc *HistoryUseCase) List(ctx context.Context, userID, contentType string, page, pageSize int) ([]*History, int, error) {
	return uc.repo.List(ctx, userID, contentType, page, pageSize)
}

// ClearAll removes all history for a user.
func (uc *HistoryUseCase) ClearAll(ctx context.Context, userID string) (int, error) {
	return uc.repo.DeleteAll(ctx, userID)
}

// Remove removes a single history record.
func (uc *HistoryUseCase) Remove(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// Sync merges local history items with server records.
func (uc *HistoryUseCase) Sync(ctx context.Context, userID string, items []*History) ([]*History, int, error) {
	return uc.repo.Sync(ctx, userID, items)
}
