/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Comment query service encapsulates entity.Client usage for the service layer.
 */

package dal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/comment"
	"origadmin/application/origstudio/internal/data/entity/media"
	"origadmin/application/origstudio/internal/features/content/dto"
)

// CommentQueryService provides comment query methods that return DTOs
// instead of raw entities, isolating the service layer from entity imports.
type CommentQueryService struct {
	db *entity.Client
}

// NewCommentQueryService creates a new CommentQueryService.
func NewCommentQueryService(db *entity.Client) *CommentQueryService {
	return &CommentQueryService{db: db}
}

// ResolveMediaID resolves a media identifier (either UUID or short_token) to the internal UUID.
// If the input is already a valid UUID format, it returns it as-is.
// If it's a short_token, it looks up the media by short_token and returns its UUID.
// If lookup fails, returns the original idOrToken as-is.
func (s *CommentQueryService) ResolveMediaID(ctx context.Context, idOrToken string) string {
	// If it looks like a UUID (contains hyphens and is 36 chars), return as-is
	if strings.Contains(idOrToken, "-") && len(idOrToken) >= 36 {
		return idOrToken
	}

	// Try to look up by short_token
	m, err := s.db.Media.Query().
		Where(media.ShortTokenEQ(idOrToken)).
		Only(ctx)
	if err != nil {
		return idOrToken
	}
	return m.ID
}

// CheckMediaExists checks if a media exists by its UUID.
func (s *CommentQueryService) CheckMediaExists(ctx context.Context, mediaID string) error {
	_, err := s.db.Media.Get(ctx, mediaID)
	return err
}

// IncrementCommentCount increments the comment count for a media by delta.
func (s *CommentQueryService) IncrementCommentCount(ctx context.Context, mediaID string, delta int) error {
	_, err := s.db.Media.UpdateOneID(mediaID).
		AddCommentCount(int64(delta)).
		Save(ctx)
	return err
}

// CommentListParams holds parameters for listing comments.
type CommentListParams struct {
	MediaID  string
	UserID   string
	ParentID string
	RootOnly bool
	Status   string
	SortBy   string
	Order    string
	Page     int
	PageSize int
	IsAdmin  bool
}

// CommentListResult holds the result of listing comments.
type CommentListResult struct {
	Items []*CommentListItem
	Total int
}

// CommentListItem represents a comment with its user and parent data.
type CommentListItem struct {
	*dto.CommentDTO
	Username        string `json:"username,omitempty"`
	Avatar          string `json:"avatar,omitempty"`
	ParentID        string `json:"parent_id,omitempty"`
	ParentContent   string `json:"parent_content,omitempty"`
	ParentUsername  string `json:"parent_username,omitempty"`
	IsReply         bool   `json:"is_reply,omitempty"`
}

// ListComments lists comments with filtering, sorting, and pagination.
func (s *CommentQueryService) ListComments(ctx context.Context, params CommentListParams) (*CommentListResult, error) {
	query := s.db.Comment.Query()

	if params.MediaID != "" {
		query = query.Where(comment.MediaID(params.MediaID))
	}
	if params.UserID != "" {
		query = query.Where(comment.UserID(params.UserID))
	}
	if params.ParentID != "" {
		query = query.Where(comment.HasParentWith(comment.ID(params.ParentID)))
	}
	if params.RootOnly {
		// 仅返回根评论（无父级），用于首屏懒加载：点击根评论再按 parent_id 拉回复
		query = query.Where(comment.Not(comment.HasParent()))
	}

	// Status filtering
	if !params.IsAdmin {
		if params.UserID != "" {
			query = query.Where(comment.Or(
				comment.StatusEQ(comment.StatusAPPROVED),
				comment.UserID(params.UserID),
			))
		} else {
			query = query.Where(comment.StatusEQ(comment.StatusAPPROVED))
		}
	} else if params.Status != "" {
		query = query.Where(comment.StatusEQ(comment.Status(params.Status)))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count comments: %w", err)
	}

	// Sorting
	switch params.SortBy {
	case "like_count":
		if params.Order == "asc" {
			query = query.Order(entity.Asc(comment.FieldLikeCount))
		} else {
			query = query.Order(entity.Desc(comment.FieldLikeCount))
		}
	default:
		if params.Order == "asc" {
			query = query.Order(entity.Asc(comment.FieldAddDate))
		} else {
			query = query.Order(entity.Desc(comment.FieldAddDate))
		}
	}

	items, err := query.
		Limit(params.PageSize).
		Offset((params.Page - 1) * params.PageSize).
		WithUser().
		WithParent(func(pq *entity.CommentQuery) {
			pq.WithUser()
		}).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}

	result := &CommentListResult{
		Items: make([]*CommentListItem, len(items)),
		Total: total,
	}

	for i, item := range items {
		result.Items[i] = &CommentListItem{
			CommentDTO: &dto.CommentDTO{
				ID:         item.ID,
				Content:    item.Text,
				MediaID:    item.MediaID,
				UserID:     item.UserID,
				Status:     string(item.Status),
				LikeCount:  item.LikeCount,
				CreateTime: item.AddDate,
				UpdateTime: item.AddDate,
			},
			IsReply: item.Edges.Parent != nil,
		}

		if item.Edges.User != nil {
			result.Items[i].Username = item.Edges.User.Username
			result.Items[i].Avatar = item.Edges.User.Logo
		}

		if item.Edges.Parent != nil {
			result.Items[i].ParentID = item.Edges.Parent.ID
			result.Items[i].ParentContent = truncateText(item.Edges.Parent.Text, 100)
			if item.Edges.Parent.Edges.User != nil {
				result.Items[i].ParentUsername = item.Edges.Parent.Edges.User.Username
			}
		}
	}

	return result, nil
}

// GetComment gets a single comment by ID.
func (s *CommentQueryService) GetComment(ctx context.Context, id string) (*CommentListItem, error) {
	item, err := s.db.Comment.Query().
		Where(comment.ID(id)).
		WithUser().
		WithParent(func(pq *entity.CommentQuery) {
			pq.WithUser()
		}).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}

	result := &CommentListItem{
		CommentDTO: &dto.CommentDTO{
			ID:         item.ID,
			Content:    item.Text,
			MediaID:    item.MediaID,
			UserID:     item.UserID,
			Status:     string(item.Status),
			LikeCount:  item.LikeCount,
			CreateTime: item.AddDate,
			UpdateTime: item.AddDate,
		},
		IsReply: item.Edges.Parent != nil,
	}

	if item.Edges.User != nil {
		result.Username = item.Edges.User.Username
		result.Avatar = item.Edges.User.Logo
	}

	if item.Edges.Parent != nil {
		result.ParentID = item.Edges.Parent.ID
		result.ParentContent = truncateText(item.Edges.Parent.Text, 100)
		if item.Edges.Parent.Edges.User != nil {
			result.ParentUsername = item.Edges.Parent.Edges.User.Username
		}
	}

	return result, nil
}

// CreateCommentParams holds parameters for creating a comment.
type CreateCommentParams struct {
	Content  string
	UserID   string
	MediaID  string
	ParentID string
	Status   string
}

// CreateComment creates a new comment and returns it as a DTO.
func (s *CommentQueryService) CreateComment(ctx context.Context, params CreateCommentParams) (*CommentListItem, error) {
	builder := s.db.Comment.Create().
		SetText(params.Content).
		SetUserID(params.UserID).
		SetStatus(comment.Status(params.Status))

	if params.MediaID != "" {
		builder = builder.SetMediaID(params.MediaID)
	}
	if params.ParentID != "" {
		builder = builder.SetParentID(params.ParentID)
	}

	item, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	return &CommentListItem{
		CommentDTO: &dto.CommentDTO{
			ID:         item.ID,
			Content:    item.Text,
			MediaID:    item.MediaID,
			UserID:     item.UserID,
			Status:     string(item.Status),
			LikeCount:  item.LikeCount,
			CreateTime: item.AddDate,
			UpdateTime: item.AddDate,
		},
	}, nil
}

// UpdateCommentParams holds parameters for updating a comment.
type UpdateCommentParams struct {
	Content string
	Status  string
}

// UpdateComment updates a comment and returns it as a DTO.
func (s *CommentQueryService) UpdateComment(ctx context.Context, id string, params UpdateCommentParams) (*CommentListItem, error) {
	builder := s.db.Comment.UpdateOneID(id)
	if params.Content != "" {
		builder = builder.SetText(params.Content)
	}
	if params.Status != "" {
		builder = builder.SetStatus(comment.Status(params.Status))
	}

	item, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}

	return &CommentListItem{
		CommentDTO: &dto.CommentDTO{
			ID:         item.ID,
			Content:    item.Text,
			MediaID:    item.MediaID,
			UserID:     item.UserID,
			Status:     string(item.Status),
			LikeCount:  item.LikeCount,
			CreateTime: item.AddDate,
			UpdateTime: item.AddDate,
		},
	}, nil
}

// DeleteComment deletes a comment by ID.
func (s *CommentQueryService) DeleteComment(ctx context.Context, id string) error {
	return s.db.Comment.DeleteOneID(id).Exec(ctx)
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// Ensure time is used
var _ = time.Time{}
