/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"
)

// EncodingTaskRepo defines the repository interface for encoding tasks.
type EncodingTaskRepo interface {
	// Create creates a new encoding task.
	Create(ctx context.Context, task *EncodingTask) (*EncodingTask, error)

	// Update updates an existing encoding task.
	Update(ctx context.Context, task *EncodingTask) (*EncodingTask, error)

	// Get retrieves an encoding task by ID.
	Get(ctx context.Context, id string) (*EncodingTask, error)

	// ListByMedia retrieves all encoding tasks for a specific media ID.
	ListByMedia(ctx context.Context, mediaId string) ([]*EncodingTask, error)

	// DeleteByMedia deletes all encoding tasks for a specific media ID.
	DeleteByMedia(ctx context.Context, mediaID string) error

	// ListFlat returns a paginated flat list of encoding tasks, optionally filtered by status, media_id, profile, chunk, and search query.
	ListFlat(
		ctx context.Context,
		status string,
		mediaId *string,
		profileFilter string,
		profileID int,
		chunkFilter string,
		searchQuery string,
		offset, limit int,
	) ([]*EncodingTask, int, error)

	// CountByStatus returns per-status counts from the encoding_task table.
	CountByStatus(ctx context.Context) (*StatusCounts, error)

	// CountByStatusWithFilter returns per-status counts filtered by status, media_id, profile, chunk, and search query.
	CountByStatusWithFilter(
		ctx context.Context,
		status string,
		mediaId *string,
		profileFilter string,
		profileID int,
		chunkFilter string,
		searchQuery string,
	) (*StatusCounts, error)
}
