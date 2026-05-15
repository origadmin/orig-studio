/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/api/gen/v1/types" // Import the generated Media type
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/enums"
	"origadmin/application/origstudio/internal/helpers/ffmpeg"
	"origadmin/application/origstudio/internal/helpers/repo"
	"origadmin/application/origstudio/internal/infra/pubsub"
	"origadmin/application/origstudio/internal/features/media/dto"
)

// EncodingEvent represents a status change event for an encoding task.
type EncodingEvent struct {
	MediaId  string
	Task     *EncodingTask
	Progress int     `json:"progress,omitempty"`
	Speed    string  `json:"speed,omitempty"`
	Fps      float64 `json:"fps,omitempty"`
	Time     float64 `json:"time,omitempty"`
}

// Media is a wrapper for types.Media for biz layer.
type Media = types.Media

// MediaRepo defines the storage operations for media.
type MediaRepo = dto.MediaRepo

// EncodeProfileRepo defines the storage operations for encode profiles.
type EncodeProfileRepo = dto.EncodeProfileRepo

// EncodingTask represents a transcoding sub-task for a specific media and profile.
type EncodingTask = dto.EncodingTask

// EncodingTaskRepo defines the storage operations for encoding tasks.
type EncodingTaskRepo = dto.EncodingTaskRepo

// ReviewLog represents a single review log entry.
type ReviewLog struct {
	ID             string `json:"id"`
	MediaID        string `json:"media_id"`
	ReviewerID     string `json:"reviewer_id"`
	Action         string `json:"action"`
	Comment        string `json:"comment"`
	PreviousStatus string `json:"previous_status"`
	NewStatus      string `json:"new_status"`
	CreateTime      string `json:"create_time"`
}

// ReviewLogRepo defines the storage operations for review logs.
type ReviewLogRepo interface {
	Create(ctx context.Context, mediaID string, reviewerID string, action string, comment string, previousStatus string, newStatus string) (*ReviewLog, error)
	ListByMedia(ctx context.Context, mediaID string) ([]*ReviewLog, error)
}

type MediaUseCase struct {
	repo          MediaRepo
	profileRepo   EncodeProfileRepo
	encodingRepo  EncodingTaskRepo
	reviewLogRepo ReviewLogRepo
	storage       Storage
	publisher     message.Publisher
	log           *log.Helper
	spriteUC      *SpriteUseCase

	mu   sync.RWMutex
	subs map[string][]chan *EncodingEvent
}

func NewMediaUseCase(
	repo MediaRepo,
	profileRepo EncodeProfileRepo,
	encodingRepo EncodingTaskRepo,
	reviewLogRepo ReviewLogRepo,
	storage Storage,
	publisher message.Publisher,
	logger log.Logger,
	spriteUC *SpriteUseCase,
) *MediaUseCase {
	if logger == nil {
		logger = log.DefaultLogger
	}
	return &MediaUseCase{
		repo:          repo,
		profileRepo:   profileRepo,
		encodingRepo:  encodingRepo,
		reviewLogRepo: reviewLogRepo,
		storage:       storage,
		publisher:     publisher,
		log:           log.NewHelper(log.With(logger, "module", "media.biz")),
		spriteUC:      spriteUC,
		subs:          make(map[string][]chan *EncodingEvent),
	}
}

func (uc *MediaUseCase) GetMedia(ctx context.Context, id string) (*Media, error) {
	return uc.repo.Get(ctx, id)
}

// GetMediaWithEntity returns a single media with its loaded ent entity (including Edges).
// This avoids extra queries when edges (user, category) are needed by the server layer.
func (uc *MediaUseCase) GetMediaWithEntity(ctx context.Context, id string) (*entity.Media, *Media, error) {
	return uc.repo.GetWithEntity(ctx, id)
}

// GetByShortToken 通过 short_token 获取媒体（仅用于公开 API）
func (uc *MediaUseCase) GetByShortToken(ctx context.Context, shortToken string) (*Media, error) {
	return uc.repo.GetByShortToken(ctx, shortToken)
}

// GetByID 通过 UUID 获取媒体（仅用于 Admin/Authenticated API）
func (uc *MediaUseCase) GetByID(ctx context.Context, id string) (*Media, error) {
	return uc.repo.GetByID(ctx, id)
}

// ResolveToID 将 short_token 解析为内部 ID
func (uc *MediaUseCase) ResolveToID(ctx context.Context, shortToken string) (string, error) {
	return uc.repo.ResolveToID(ctx, shortToken)
}

// CheckMedia verifies that a media record exists. Returns an error if not found.
// Satisfies contentbiz.MediaUseCaseInterface without leaking *types.Media into the content layer.
func (uc *MediaUseCase) CheckMedia(ctx context.Context, id string) error {
	_, err := uc.repo.Get(ctx, id)
	return err
}

func (uc *MediaUseCase) ListMedias(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*Media, int32, error) {
	return uc.repo.List(ctx, opts...)
}

// ListMediasWithEntities returns media list with loaded ent entities (including Edges).
// This avoids N+1 queries when edges (user, category) are needed by the server layer.
func (uc *MediaUseCase) ListMediasWithEntities(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*entity.Media, []*Media, int32, error) {
	return uc.repo.ListWithEntities(ctx, opts...)
}

func (uc *MediaUseCase) CreateMedia(ctx context.Context, media *Media) (*Media, error) {
	// 验证媒体必须关联且仅关联一个分类
	if media.CategoryId == 0 {
		return nil, fmt.Errorf("media must have a category")
	}

	created, err := uc.repo.Create(ctx, media)
	if err != nil {
		return nil, err
	}

	// Trigger transcoding for videos
	if strings.HasPrefix(created.MimeType, "video/") && uc.publisher != nil {
		payload, _ := json.Marshal(struct {
			MediaID     string `json:"media_id"`
			MediaPath   string `json:"media_path"`
			ContentType string `json:"content_type"`
		}{
			MediaID:     created.Id,
			MediaPath:   created.Url,
			ContentType: created.MimeType,
		})
		msg := pubsub.NewMessage(payload)
		if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
			uc.log.Errorf("failed to publish encode request for media %s: %v", created.Id, err)
		}
	}

	return created, nil
}

func (uc *MediaUseCase) UpdateMedia(ctx context.Context, media *Media) (*Media, error) {
	// Note: CategoryId validation is intentionally NOT enforced on update.
	// The admin update handler (adminUpdateMedia) fetches the existing record first,
	// then merges partial fields. If category_id was not provided in the request,
	// the existing CategoryId from the database is preserved.
	// CategoryId is still required on CREATE (see CreateMedia).
	return uc.repo.Update(ctx, media)
}

func (uc *MediaUseCase) DeleteMedia(ctx context.Context, id string) error {
	m, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete encoding tasks first to avoid foreign key constraint
	if err := uc.encodingRepo.DeleteByMedia(ctx, id); err != nil {
		uc.log.Warnf("failed to delete encoding tasks for media %s: %v", id, err)
	}

	// Delete from DB
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Async cleanup of physical files
	if uc.storage != nil && m.Url != "" {
		go func() {
			bgCtx := context.Background()
			// Original file
			if err := uc.storage.Delete(bgCtx, m.Url); err != nil {
				uc.log.Warnf("failed to delete media file %s: %v", m.Url, err)
			}
			// Thumbnail
			if m.Thumbnail != "" {
				if err := uc.storage.Delete(bgCtx, m.Thumbnail); err != nil {
					uc.log.Warnf("failed to delete thumbnail %s: %v", m.Thumbnail, err)
				}
			}
			// Encoding tasks and their output files could also be cleaned here
		}()
	}

	return nil
}

func (uc *MediaUseCase) IncrementViewCount(ctx context.Context, id string) (int64, error) {
	return uc.repo.IncrementViewCount(ctx, id)
}

func (uc *MediaUseCase) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	return uc.repo.UpdateCommentCount(ctx, id, delta)
}

func (uc *MediaUseCase) UpdateLikeCount(ctx context.Context, id string, delta int) error {
	return uc.repo.UpdateLikeCount(ctx, id, delta)
}

func (uc *MediaUseCase) UpdateDislikeCount(ctx context.Context, id string, delta int) error {
	return uc.repo.UpdateDislikeCount(ctx, id, delta)
}

func (uc *MediaUseCase) UpdateFavoriteCount(ctx context.Context, id string, delta int) error {
	return uc.repo.UpdateFavoriteCount(ctx, id, delta)
}

func (uc *MediaUseCase) UpdateMediaState(ctx context.Context, id string, state string) error {
	m, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	m.State = state
	_, err = uc.repo.Update(ctx, m)
	return err
}

func (uc *MediaUseCase) ListEncodingTasks(
	ctx context.Context,
	mediaId string,
) ([]*EncodingTask, error) {
	return uc.encodingRepo.ListByMedia(ctx, mediaId)
}

// RetryTask resets a failed/partial encoding task back to "pending" so it can be re-processed.
// It re-publishes the media encode request with the specific task ID to trigger processing.
func (uc *MediaUseCase) RetryTask(ctx context.Context, taskID string) (*EncodingTask, error) {
	task, err := uc.encodingRepo.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task %s not found: %w", taskID, err)
	}

	// Only allow retrying failed or partial tasks
	if task.Status != "failed" && task.Status != "partial" {
		return nil, fmt.Errorf(
			"cannot retry task %s with status %q (only 'failed' or 'partial' can be retried)",
			taskID,
			task.Status,
		)
	}

	// Update task status and clear error message BEFORE publishing
	task.Status = "pending"
	task.ErrorMessage = ""
	if _, err := uc.encodingRepo.Update(ctx, task); err != nil {
		uc.log.Errorf("failed to reset task %s status in DB for retry: %v", taskID, err)
		return nil, fmt.Errorf("update task status: %w", err)
	}

	// Get media to re-publish encode request
	media, err := uc.repo.Get(ctx, task.MediaId)
	if err != nil {
		uc.log.Warnf("failed to get media %s for retry: %v", task.MediaId, err)
		return nil, err
	}

	// Re-publish encode request with specific task ID to trigger processing
	if uc.publisher != nil {
		payload, _ := json.Marshal(struct {
			MediaID     string `json:"media_id"`
			MediaPath   string `json:"media_path"`
			ContentType string `json:"content_type"`
			TaskID      string `json:"task_id"`
		}{
			MediaID:     media.Id,
			MediaPath:   media.Url,
			ContentType: media.MimeType,
			TaskID:      taskID,
		})
		msg := pubsub.NewMessage(payload)
		if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
			uc.log.Errorf("failed to re-publish encode request for media %s: %v", media.Id, err)
		} else {
			uc.log.Infof("re-published encode request for media %s with task ID %s", media.Id, taskID)
		}
	}

	uc.log.Infof(
		"task %s (media=%s profile=%d) queued for retry with specific task ID",
		taskID,
		task.MediaId,
		task.ProfileId,
	)
	return task, nil
}

// RetryAllFailedTasks resets all failed tasks for a given media back to "pending".
// Returns the count of tasks that were reset.
// It re-publishes the media encode request to trigger processing, but only the pending tasks will be processed.
func (uc *MediaUseCase) RetryAllFailedTasks(ctx context.Context, mediaID string) (int, error) {
	tasks, err := uc.encodingRepo.ListByMedia(ctx, mediaID)
	if err != nil {
		return 0, fmt.Errorf("failed to list tasks for media %s: %w", mediaID, err)
	}

	resetCount := 0
	for _, t := range tasks {
		if t.Status != "failed" && t.Status != "partial" {
			continue
		}
		t.Status = "pending"
		t.ErrorMessage = ""
		if _, err := uc.encodingRepo.Update(ctx, t); err != nil {
			uc.log.Warnf("failed to reset task %s during bulk retry: %v", t.Id, err)
			continue
		}
		resetCount++
	}

	if resetCount > 0 {
		// Get media to re-publish encode request
		media, err := uc.repo.Get(ctx, mediaID)
		if err != nil {
			uc.log.Warnf("failed to get media %s for retry: %v", mediaID, err)
			return resetCount, nil
		}

		// Re-publish encode request to trigger processing
		if uc.publisher != nil {
			payload, _ := json.Marshal(struct {
				MediaID     string `json:"media_id"`
				MediaPath   string `json:"media_path"`
				ContentType string `json:"content_type"`
			}{
				MediaID:     media.Id,
				MediaPath:   media.Url,
				ContentType: media.MimeType,
			})
			msg := pubsub.NewMessage(payload)
			if err := uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg); err != nil {
				uc.log.Errorf("failed to re-publish encode request for media %s: %v", media.Id, err)
			} else {
				uc.log.Infof("re-published encode request for media %s", media.Id)
			}
		}
	}

	uc.log.Infof("reset %d failed tasks for media %s to pending and re-queued", resetCount, mediaID)
	return resetCount, nil
}

// --- Transcoding Status ---

// VariantInfo holds aggregated info about a single encoding profile's result.
// Returned by GetMediaVariants for frontend display and player configuration.
type VariantInfo struct {
	TaskID       string                   `json:"task_id"`
	ProfileName  string                   `json:"profile_name"`
	ProfileID    int                      `json:"profile_id"`
	Resolution   string                   `json:"resolution"`  // e.g., "1280x720" or "720"
	Codec        string                   `json:"codec"`       // e.g., "h264", "h265"
	Status       enums.EncodingTaskStatus `json:"status"`      // pending, processing, success, failed, skipped
	OutputPath   string                   `json:"output_path"` // HLS playlist path or GIF path
	Bandwidth    int                      `json:"bandwidth"`   // bits per second (estimated)
	ErrorMessage string                   `json:"error_message,omitempty"`
}

// MediaVariantSummary is the aggregated transcoding status returned by media detail APIs.
// This is what the "media management" page displays — a compact view of all profile outcomes.
type MediaVariantSummary struct {
	MediaID              string        `json:"media_id"`
	EncodingStatus       string        `json:"encoding_status"` // overall: success/partial/failed/processing
	HlsFile              string        `json:"hls_file"`        // master.m3u8 path
	Thumbnail            string        `json:"thumbnail"`       // thumbnail path
	PreviewFile          string        `json:"preview_file"`    // GIF preview path
	VideoTotalCount      int           `json:"video_total_count"`
	VideoSuccessCount    int           `json:"video_success_count"`
	VideoFailedCount     int           `json:"video_failed_count"`
	VideoPendingCount    int           `json:"video_pending_count"`
	VideoProcessingCount int           `json:"video_processing_count"`
	Variants             []VariantInfo `json:"variants"` // all tasks with details
}

// TranscodingStatus holds aggregated counts.
type TranscodingStatus struct {
	ProcessingCount int `json:"processing_count"`
	PendingCount    int `json:"pending_count"`
	PartialCount    int `json:"partial_count"`
	FailedCount     int `json:"failed_count"`
	SuccessCount    int `json:"success_count"`
}

// FilterType defines the type of filter for transcoding tasks
type FilterType string

const (
	FilterTypeAll      FilterType = "all"
	FilterTypeActive   FilterType = "active"
	FilterTypeSpecific FilterType = "specific"
)

// TranscodingStatusFilter controls which media are returned in the items list.
type TranscodingStatusFilter struct {
	// FilterType: "all", "active", or "specific"
	FilterType FilterType
	// Status: specific status to filter by (only used when FilterType is "specific")
	Status string
	// Page is 1-based.  Defaults to 1.
	Page int
	// PageSize limits items returned.  Defaults to 20, max 100.
	PageSize int
	// OnlyStats: if true, returns only statistics without items list.
	OnlyStats bool
	// ProfileFilter: filter by profile name (partial match).
	ProfileFilter string
	// ChunkFilter: filter by chunk name (exact match).
	ChunkFilter string
	// SearchQuery: search query for media_id, profile_name, or status.
	SearchQuery string
}

// FlatTaskList holds a flat (non-grouped) list of encoding tasks with counts and pagination.
type FlatTaskList struct {
	ProcessingCount int            `json:"processing_count"`
	PendingCount    int            `json:"pending_count"`
	PartialCount    int            `json:"partial_count"`
	FailedCount     int            `json:"failed_count"`
	SuccessCount    int            `json:"success_count"`
	Total           int            `json:"total"`
	Page            int            `json:"page"`
	PageSize        int            `json:"page_size"`
	Items           []FlatTaskItem `json:"items"`
}

// FlatTaskItem is a single task row for the flat task list view.
type FlatTaskItem struct {
	Id           string                   `json:"id"`
	MediaId      string                   `json:"media_id"`
	MediaTitle   string                   `json:"media_title,omitempty"`
	Thumbnail    string                   `json:"thumbnail,omitempty"`
	ProfileId    int                      `json:"profile_id"`
	ProfileName  string                   `json:"profile_name,omitempty"`
	Status       enums.EncodingTaskStatus `json:"status"`
	OutputPath   string                   `json:"output_path,omitempty"`
	ErrorMessage string                   `json:"error_message,omitempty"`
	CreateTime   string                   `json:"create_time,omitempty"`
	UpdateTime   string                   `json:"update_time,omitempty"`
}

// StatusCounts holds per-media-status counts.
type StatusCounts = dto.StatusCounts

func (uc *MediaUseCase) GetTranscodingStatus(
	ctx context.Context,
	filter *TranscodingStatusFilter,
) (*TranscodingStatus, error) {
	// Use encoding_task table for accurate task-level counts
	counts, err := uc.encodingRepo.CountByStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &TranscodingStatus{
		ProcessingCount: counts.Processing,
		PendingCount:    counts.Pending,
		PartialCount:    counts.Partial,
		FailedCount:     counts.Failed,
		SuccessCount:    counts.Success,
	}, nil
}

// ListEncodingTasksFlat returns a flat, paginated list of encoding tasks (one row per task).
// Used by the TranscodingStatus page for a pure task-centric view.
func (uc *MediaUseCase) ListEncodingTasksFlat(
	ctx context.Context,
	filter *TranscodingStatusFilter,
	mediaID *string,
) (*FlatTaskList, error) {
	if filter == nil {
		filter = &TranscodingStatusFilter{Status: "active", Page: 1, PageSize: 25}
	}
	// Normalize pagination parameters using centralized validation
	filter.Page, filter.PageSize = repo.NormalizePagination(filter.Page, filter.PageSize)

	var status string
	switch filter.FilterType {
	case FilterTypeAll:
		status = "all"
	case FilterTypeActive:
		// Active filter: exclude success status
		status = "active"
	case FilterTypeSpecific:
		status = filter.Status
	default:
		// Default to all
		status = "all"
	}

	if mediaID != nil && *mediaID != "" {
		// When filtered to a specific media, default to "all" if not specified
		if status == "" {
			status = "all"
		}
	}

	offset := (filter.Page - 1) * filter.PageSize

	// 直接使用传入的 mediaID，无论是数字还是 UUID
	// 编码任务的 MediaId 字段存储的是 UUID，所以直接使用
	var mediaIDStr *string
	if mediaID != nil && *mediaID != "" {
		uc.log.Infof("Using media_id: %s", *mediaID)
		mediaIDStr = mediaID
	}

	// Get profile ID from profile name if provided
	var profileID int
	if filter.ProfileFilter != "" {
		profile, err := uc.profileRepo.GetByName(ctx, filter.ProfileFilter)
		if err == nil && profile != nil {
			profileID = profile.Id
			fmt.Printf("Profile filter: %s, Profile ID: %d\n", filter.ProfileFilter, profileID)
		} else {
			fmt.Printf("Profile not found: %s, Error: %v\n", filter.ProfileFilter, err)
		}
	}
	tasks, _, err := uc.encodingRepo.ListFlat(
		ctx,
		status,
		mediaIDStr,
		filter.ProfileFilter,
		profileID,
		filter.ChunkFilter,
		filter.SearchQuery,
		offset,
		filter.PageSize,
	)
	if err != nil {
		return nil, err
	}

	// Get counts: use complete stats if only_stats=true, else use filtered stats (including Status for total, excluding Status for filter buttons)
	var counts *StatusCounts
	var countErr error
	if filter.OnlyStats {
		// Card area: complete stats without any filter
		counts, countErr = uc.encodingRepo.CountByStatus(ctx)
	} else {
		// Filter buttons area: stats excluding Status filter but including other filters (Profile, Chunk, Search)
		// This ensures filter buttons show counts that exclude their own Status condition
		counts, countErr = uc.encodingRepo.CountByStatusWithFilter(
			ctx,
			"", // Exclude Status filter for filter buttons
			mediaIDStr,
			filter.ProfileFilter,
			profileID,
			filter.ChunkFilter,
			filter.SearchQuery,
		)
	}
	if countErr != nil {
		counts = &StatusCounts{}
	}

	// For the "All" filter button, we need the total count excluding Status filter
	// This ensures consistency between filter buttons and the "All" count
	var totalWithoutStatusFilter int
	if !filter.OnlyStats {
		totalWithoutStatusFilter = counts.Processing + counts.Pending + counts.Partial + counts.Failed + counts.Success
	}

	var items []FlatTaskItem
	if !filter.OnlyStats {
		// Collect unique media IDs for batch lookup
		mediaIDSet := make(map[string]struct{})
		for _, t := range tasks {
			if t.MediaId != "" {
				mediaIDSet[t.MediaId] = struct{}{}
			}
		}
		mediaCache := make(map[string]*types.Media)
		for mid := range mediaIDSet {
			m, err := uc.repo.Get(ctx, mid)
			if err != nil {
				uc.log.Warnf("ListEncodingTasksFlat: failed to get media %s: %v", mid, err)
				continue
			}
			if m != nil {
				mediaCache[mid] = m
			}
		}

		// Enrich with profile names and media info
		items = make([]FlatTaskItem, len(tasks))
		for i, t := range tasks {
			item := FlatTaskItem{
				Id:           t.Id,
				MediaId:      t.MediaId,
				ProfileId:    t.ProfileId,
				Status:       t.Status,
				OutputPath:   t.OutputPath,
				ErrorMessage: t.ErrorMessage,
				CreateTime:   t.CreateTime,
				UpdateTime:   t.UpdateTime,
			}
			// Look up profile name
			if profile, perr := uc.profileRepo.Get(ctx, t.ProfileId); perr == nil && profile != nil {
				item.ProfileName = profile.Name
			}
			// Look up media title and thumbnail
			if m, ok := mediaCache[t.MediaId]; ok {
				item.MediaTitle = m.Title
				item.Thumbnail = m.Thumbnail
			}
			items[i] = item
		}
	}

	return &FlatTaskList{
		ProcessingCount: counts.Processing,
		PendingCount:    counts.Pending,
		PartialCount:    counts.Partial,
		FailedCount:     counts.Failed,
		SuccessCount:    counts.Success,
		Total:           totalWithoutStatusFilter,
		Page:            filter.Page,
		PageSize:        filter.PageSize,
		Items:           items,
	}, nil
}

// GetMediaVariants returns a comprehensive variant summary for a single media.
// Used by media detail/management pages to show transcoding status and provide
// playback URLs to the frontend.
func (uc *MediaUseCase) GetMediaVariants(
	ctx context.Context,
	mediaID int64,
) (*MediaVariantSummary, error) {
	// 1. Load media
	mediaIDStr := strconv.FormatInt(mediaID, 10)
	return uc.getMediaVariantsByID(ctx, mediaIDStr, mediaID)
}

// getMediaVariantsByID is a helper function that takes a string media ID (UUID or numeric)
// and returns the media variant summary.
func (uc *MediaUseCase) getMediaVariantsByID(
	ctx context.Context,
	mediaIDStr string,
	mediaID int64,
) (*MediaVariantSummary, error) {
	media, err := uc.repo.Get(ctx, mediaIDStr)
	if err != nil {
		return nil, err
	}

	// 2. Load all encoding tasks for this media
	tasks, err := uc.encodingRepo.ListByMedia(ctx, mediaIDStr)
	if err != nil {
		// No tasks yet — return minimal info
		return &MediaVariantSummary{
			MediaID:        media.Id,
			EncodingStatus: media.EncodingStatus,
			HlsFile:        media.HlsFile,
			Thumbnail:      media.Thumbnail,
			PreviewFile:    media.PreviewFilePath,
			Variants:       []VariantInfo{},
		}, nil
	}

	// 3. Enrich tasks with profile info
	var variants []VariantInfo
	videoSuccessCount := 0
	videoFailedCount := 0
	videoPendingCount := 0
	videoProcessingCount := 0

	for _, t := range tasks {
		profile, perr := uc.profileRepo.Get(ctx, t.ProfileId)
		if perr != nil {
			uc.log.Warnf("profile %d not found for task %s", t.ProfileId, t.Id)
			continue
		}

		vi := VariantInfo{
			TaskID:       t.Id,
			ProfileName:  profile.Name,
			ProfileID:    t.ProfileId,
			Resolution:   profile.Resolution,
			Codec:        profile.VideoCodec,
			Status:       t.Status,
			OutputPath:   t.OutputPath,
			ErrorMessage: t.ErrorMessage,
		}

		// Estimate bandwidth from BentoParameters or resolution
		vi.Bandwidth = estimateProfileBandwidth(profile)

		variants = append(variants, vi)

		// Count video task outcomes by status (exclude preview)
		if !IsPreviewProfileFromName(profile.Name) {
			switch t.Status {
			case "success":
				videoSuccessCount++
			case "failed":
				videoFailedCount++
			case "pending":
				videoPendingCount++
			case "processing":
				videoProcessingCount++
			}
		}
	}

	videoTotalCount := videoSuccessCount + videoFailedCount + videoPendingCount + videoProcessingCount

	return &MediaVariantSummary{
		MediaID:              media.Id,
		EncodingStatus:       media.EncodingStatus,
		HlsFile:              media.HlsFile,
		Thumbnail:            media.Thumbnail,
		PreviewFile:          media.PreviewFilePath,
		VideoTotalCount:      videoTotalCount,
		VideoSuccessCount:    videoSuccessCount,
		VideoFailedCount:     videoFailedCount,
		VideoPendingCount:    videoPendingCount,
		VideoProcessingCount: videoProcessingCount,
		Variants:             variants,
	}, nil
}

// GetMediaVariantsByUUID returns a comprehensive variant summary for a single media by UUID.
// Used by gRPC service to handle UUID format media IDs.
func (uc *MediaUseCase) GetMediaVariantsByUUID(
	ctx context.Context,
	mediaIDStr string,
) (*MediaVariantSummary, error) {
	// For UUID format, we can use 0 as the numeric mediaID since it's not used in the response
	return uc.getMediaVariantsByID(ctx, mediaIDStr, 0)
}

// IsPreviewProfileFromName returns true if the profile name indicates it's a preview/GIF type.
func IsPreviewProfileFromName(name string) bool {
	return strings.EqualFold(name, "preview") || strings.EqualFold(name, "gif")
}

// estimateProfileBandwidth estimates bandwidth in bps from profile settings.
func estimateProfileBandwidth(p *dto.EncodeProfile) int {
	// Try parsing from BentoParameters first
	if p.BentoParameters != "" {
		fields := strings.Fields(p.BentoParameters)
		for i, f := range fields {
			if f == "--video-bitrate" && i+1 < len(fields) {
				bps := parseBitrateToBps(fields[i+1])
				if bps > 0 {
					return bps
				}
			}
		}
	}

	// Fallback by resolution
	switch p.Resolution {
	case "2160":
		return 20_000_000
	case "1440":
		return 12_000_000
	case "1080":
		return 8_000_000
	case "720":
		return 4_000_000
	case "480":
		return 2_000_000
	case "360":
		return 1_000_000
	case "240":
		return 500_000
	default:
		return 1_000_000
	}
}

// parseBitrateToBps converts human bitrate string (e.g., "400k", "1.5M") to bps.
func parseBitrateToBps(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) == 0 {
		return 0
	}

	multiplier := int(1000)
	if strings.HasSuffix(s, "m") {
		multiplier = 1_000_000
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "k") {
		multiplier = 1000
		s = s[:len(s)-1]
	}

	val, err := parseFloat64(s)
	if err != nil {
		return 0
	}
	return int(val * float64(multiplier))
}

// parseFloat64 parses a float64 with Sscanf fallback.
func parseFloat64(s string) (float64, error) {
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

func statusesForFilter(status string) []string {
	switch status {
	case "processing":
		return []string{"processing"}
	case "pending":
		return []string{"pending"}
	case "partial":
		return []string{"partial"}
	case "failed":
		return []string{"failed"}
	case "success":
		return []string{"success"}
	case "all":
		return []string{"pending", "processing", "partial", "failed", "success"}
	case "active":
		fallthrough
	default:
		return []string{"pending", "processing", "partial", "failed"}
	}
}

// --- SSE Pub/Sub ---

// Subscribe returns a channel that receives encoding events for a specific media.
// Pass mediaID="" to subscribe to all media events (dashboard use).
func (uc *MediaUseCase) Subscribe(
	ctx context.Context,
	mediaID string,
) (<-chan *EncodingEvent, func()) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	ch := make(chan *EncodingEvent, 32)
	uc.subs[mediaID] = append(uc.subs[mediaID], ch)

	cleanup := func() {
		uc.mu.Lock()
		defer uc.mu.Unlock()
		subs := uc.subs[mediaID]
		for i, s := range subs {
			if s == ch {
				uc.subs[mediaID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, cleanup
}

// Publish sends an encoding event to all subscribers of a media,
// and also broadcasts to the global subscriber (mediaID="").
func (uc *MediaUseCase) Publish(mediaID string, event *EncodingEvent) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// Send to per-media subscribers.
	for _, ch := range uc.subs[mediaID] {
		select {
		case ch <- event:
		default:
			// Buffer full, skip
		}
	}

	// Broadcast to global (dashboard) subscribers.
	for _, ch := range uc.subs[""] {
		select {
		case ch <- event:
		default:
		}
	}
}

// --- Encode Profiles ---

// ListEncodeProfiles returns all encoding profiles.
func (uc *MediaUseCase) ListEncodeProfiles(ctx context.Context) ([]*dto.EncodeProfile, error) {
	return uc.profileRepo.ListAll(ctx)
}

// GetEncodeProfile returns an encoding profile by ID.
func (uc *MediaUseCase) GetEncodeProfile(ctx context.Context, id int) (*dto.EncodeProfile, error) {
	return uc.profileRepo.Get(ctx, id)
}

// CreateEncodeProfile creates a new encoding profile.
func (uc *MediaUseCase) CreateEncodeProfile(
	ctx context.Context,
	profile *dto.EncodeProfile,
) (*dto.EncodeProfile, error) {
	return uc.profileRepo.Create(ctx, profile)
}

// UpdateEncodeProfile updates an existing encoding profile.
func (uc *MediaUseCase) UpdateEncodeProfile(
	ctx context.Context,
	profile *dto.EncodeProfile,
) (*dto.EncodeProfile, error) {
	return uc.profileRepo.Update(ctx, profile)
}

// DeleteEncodeProfile deletes an encoding profile.
func (uc *MediaUseCase) DeleteEncodeProfile(ctx context.Context, id int) error {
	return uc.profileRepo.Delete(ctx, id)
}

// ReviewMedia reviews a media item with state transition validation and audit logging.
func (uc *MediaUseCase) ReviewMedia(ctx context.Context, mediaID string, approve bool, comment string, reviewerID string) (*Media, error) {
	media, err := uc.repo.Get(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	previousStatus := media.ReviewStatus

	if previousStatus != "pending_review" && previousStatus != "rejected" {
		return nil, fmt.Errorf("invalid state transition: cannot review media with status %q, expected 'pending_review' or 'rejected'", previousStatus)
	}

	var newStatus string
	var action string
	if approve {
		newStatus = "reviewed"
		action = "approve"
	} else {
		newStatus = "rejected"
		action = "reject"
	}

	media.ReviewStatus = newStatus
	media.Listable = uc.ShouldBeListable(media)

	updated, err := uc.repo.Update(ctx, media)
	if err != nil {
		return nil, err
	}

	if uc.reviewLogRepo != nil {
		if _, logErr := uc.reviewLogRepo.Create(ctx, mediaID, reviewerID, action, comment, previousStatus, newStatus); logErr != nil {
			uc.log.Warnf("failed to create review log for media %s: %v", mediaID, logErr)
		}
	}

	uc.log.Infof("Media %s reviewed by %s: %s (previous: %s, new: %s)", mediaID, reviewerID, action, previousStatus, newStatus)
	return updated, nil
}

// ListReviewLogs returns the review log entries for a given media.
func (uc *MediaUseCase) ListReviewLogs(ctx context.Context, mediaID string) ([]*ReviewLog, error) {
	if uc.reviewLogRepo == nil {
		return []*ReviewLog{}, nil
	}
	return uc.reviewLogRepo.ListByMedia(ctx, mediaID)
}

// ShouldBeListable 计算媒体是否应该可见
func (uc *MediaUseCase) ShouldBeListable(media *Media) bool {
	return media.EncodingStatus == "success" && 
		   media.ReviewStatus == "reviewed" && 
		   media.State == "active"
}

// --- Entity-level data access for handler migration ---

// SpriteInfo holds sprite-related data for a media item.
// Used by handlers that need access to internal entity fields not exposed in types.Media.
type SpriteInfo struct {
	Type         string
	SpriteStatus string
	VttPath      string
	SpritePath   string
	UpdateTime   time.Time
}

// ThumbnailInfo holds thumbnail-related data for a media item.
// Used by handlers that need access to internal entity fields not exposed in types.Media.
type ThumbnailInfo struct {
	Thumbnail     string
	ThumbnailTime float64
}

// GetSpriteInfoByID returns sprite-related data for a media item by ID.
// Used by regenerateSprite handler to check type and sprite status before scheduling regeneration.
func (uc *MediaUseCase) GetSpriteInfoByID(ctx context.Context, id string) (*SpriteInfo, error) {
	ent, err := uc.repo.GetEntityByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SpriteInfo{
		Type:         ent.Type,
		SpriteStatus: ent.SpriteStatus,
		VttPath:      ent.VttPath,
		SpritePath:   ent.SpritePath,
		UpdateTime:   ent.UpdateTime,
	}, nil
}

// GetSpriteInfoByShortToken returns sprite-related data for a media item by short_token.
// Used by getSpriteVTT and getSpriteImage handlers.
func (uc *MediaUseCase) GetSpriteInfoByShortToken(ctx context.Context, shortToken string) (*SpriteInfo, error) {
	ent, err := uc.repo.GetEntityByShortToken(ctx, shortToken)
	if err != nil {
		return nil, err
	}
	return &SpriteInfo{
		Type:         ent.Type,
		SpriteStatus: ent.SpriteStatus,
		VttPath:      ent.VttPath,
		SpritePath:   ent.SpritePath,
		UpdateTime:   ent.UpdateTime,
	}, nil
}

// GetThumbnailInfoByID returns thumbnail-related data for a media item by ID.
// Used by regenerateThumbnail handler to return updated thumbnail info after regeneration.
func (uc *MediaUseCase) GetThumbnailInfoByID(ctx context.Context, id string) (*ThumbnailInfo, error) {
	ent, err := uc.repo.GetEntityByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ThumbnailInfo{
		Thumbnail:     ent.Thumbnail,
		ThumbnailTime: ent.ThumbnailTime,
	}, nil
}

func (uc *MediaUseCase) RegenerateSprite(ctx context.Context, mediaID string) error {
	if uc.spriteUC == nil {
		return fmt.Errorf("sprite use case not initialized")
	}
	return uc.spriteUC.GenerateSpriteAndVTT(ctx, mediaID)
}

func (uc *MediaUseCase) RegenerateThumbnail(ctx context.Context, mediaID string, timestamp float64) error {
	if uc.spriteUC == nil {
		return fmt.Errorf("sprite use case not initialized")
	}
	return uc.spriteUC.RegenerateThumbnail(ctx, mediaID, timestamp)
}

// GenerateCommandPreview generates a preview of the ffmpeg command that would be
// executed for the given encode profile. This does not execute the command, only
// produces the command string for display purposes.
func (uc *MediaUseCase) GenerateCommandPreview(ctx context.Context, profile *dto.EncodeProfile) string {
	inputPath := "<input_file>"
	switch {
	case IsVideoProfile(profile):
		outputDir := filepath.Join("<output_dir>", profile.Name)
		return ffmpeg.PreviewHLSCommand(
			inputPath, outputDir, profile.Name, profile.Resolution,
			profile.VideoCodec, profile.AudioCodec,
			profile.VideoBitrate, profile.AudioBitrate,
		)
	case IsPreviewProfile(profile):
		scale := extractScaleParam(profile.BentoParameters)
		outputPath := filepath.Join("<output_dir>", "previews", "<id>.gif")
		return ffmpeg.PreviewGIFCommand(inputPath, outputPath, scale)
	default:
		return ""
	}
}
