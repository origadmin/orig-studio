/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/api/gen/v1/types" // Import the generated Media type
	"origadmin/application/origcms/internal/svc-media/dto"
)

// EncodingEvent represents a status change event for an encoding task.
type EncodingEvent struct {
	MediaId int64
	Task    *EncodingTask
}

// Media is a wrapper for types.Media for biz layer.
type Media = types.Media

// MediaRepo defines the storage operations for media.
type MediaRepo interface {
	Create(ctx context.Context, media *Media) (*Media, error)
	Get(ctx context.Context, id int64) (*Media, error)
	List(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error)
	Update(ctx context.Context, media *Media) (*Media, error)
	Delete(ctx context.Context, id int64) error
	IncrementViewCount(ctx context.Context, id int64) (int64, error)
	// ResetStaleProcessing resets media stuck in "processing" state back to "pending"
	// and deletes their orphaned encoding tasks (which were interrupted by the restart).
	// Called at startup to recover from service restarts.
	ResetStaleProcessing(ctx context.Context) (int, error)
	// CountByEncodingStatus returns per-status media counts using a single GROUP BY query.
	CountByEncodingStatus(ctx context.Context) (*StatusCounts, error)
	// ListFilteredByEncodingStatus returns a paginated list of media matching the given encoding statuses.
	ListFilteredByEncodingStatus(
		ctx context.Context,
		statuses []string,
		page, pageSize int,
	) ([]*Media, int, error)
}

// EncodeProfile represents an encoding preset.
type EncodeProfile struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Extension    string `json:"extension"`
	Resolution   string `json:"resolution"`
	VideoCodec   string `json:"video_codec"`
	VideoBitrate string `json:"video_bitrate"`
	AudioCodec   string `json:"audio_codec"`
	AudioBitrate string `json:"audio_bitrate"`
	IsActive     bool   `json:"is_active"`
	// BentoParameters stores additional arguments for Bento4 tools (e.g., mp4hls)
	BentoParameters string `json:"bento_parameters"`
}

// EncodeProfileRepo defines the storage operations for encode profiles.
type EncodeProfileRepo interface {
	ListActive(ctx context.Context) ([]*EncodeProfile, error)
	ListAll(ctx context.Context) ([]*EncodeProfile, error)
	Get(ctx context.Context, id int) (*EncodeProfile, error)
	Create(ctx context.Context, profile *EncodeProfile) (*EncodeProfile, error)
	Update(ctx context.Context, profile *EncodeProfile) (*EncodeProfile, error)
	Delete(ctx context.Context, id int) error
}

// EncodingTask represents a transcoding sub-task for a specific media and profile.
type EncodingTask struct {
	Id           int    `json:"id"`
	MediaId      int64  `json:"media_id"`
	ProfileId    int    `json:"profile_id"`
	Status       string `json:"status"` // pending, processing, success, failed
	Progress     int    `json:"progress"`
	OutputPath   string `json:"output_path"`
	ErrorMessage string `json:"error_message"`
}

// EncodingTaskRepo defines the storage operations for encoding tasks.
type EncodingTaskRepo interface {
	Create(ctx context.Context, task *EncodingTask) (*EncodingTask, error)
	Update(ctx context.Context, task *EncodingTask) (*EncodingTask, error)
	Get(ctx context.Context, id int) (*EncodingTask, error)
	ListByMedia(ctx context.Context, mediaId int64) ([]*EncodingTask, error)
	// DeleteByMedia deletes all encoding tasks for a given media ID.
	DeleteByMedia(ctx context.Context, mediaID int64) error
	// ListFlat returns a paginated flat list of tasks filtered by status/media_id.
	ListFlat(
		ctx context.Context,
		status string,
		mediaId *int64,
		offset, limit int,
	) ([]*EncodingTask, int, error)
	// CountByStatus returns per-status counts from the encoding_task table (NOT the media table).
	CountByStatus(ctx context.Context) (*StatusCounts, error)
}

type MediaUseCase struct {
	repo         MediaRepo
	profileRepo  EncodeProfileRepo
	encodingRepo EncodingTaskRepo
	log          *log.Helper

	mu   sync.RWMutex
	subs map[int64][]chan *EncodingEvent
}

func NewMediaUseCase(
	repo MediaRepo,
	profileRepo EncodeProfileRepo,
	encodingRepo EncodingTaskRepo,
	logger log.Logger,
) *MediaUseCase {
	return &MediaUseCase{
		repo:         repo,
		profileRepo:  profileRepo,
		encodingRepo: encodingRepo,
		log:          log.NewHelper(log.With(logger, "module", "media.biz")),
		subs:         make(map[int64][]chan *EncodingEvent),
	}
}

func (uc *MediaUseCase) GetMedia(ctx context.Context, id int64) (*Media, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *MediaUseCase) ListMedias(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*Media, int32, error) {
	return uc.repo.List(ctx, opts...)
}

func (uc *MediaUseCase) CreateMedia(ctx context.Context, media *Media) (*Media, error) {
	return uc.repo.Create(ctx, media)
}

func (uc *MediaUseCase) UpdateMedia(ctx context.Context, media *Media) (*Media, error) {
	return uc.repo.Update(ctx, media)
}

func (uc *MediaUseCase) DeleteMedia(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *MediaUseCase) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	return uc.repo.IncrementViewCount(ctx, id)
}

func (uc *MediaUseCase) ListEncodingTasks(
	ctx context.Context,
	mediaId int64,
) ([]*EncodingTask, error) {
	return uc.encodingRepo.ListByMedia(ctx, mediaId)
}

// RetryTask resets a failed/partial encoding task back to "pending" so it can be re-processed.
// This only resets the task state; the actual re-processing must be triggered by
// publishing a new encode request or by the caller invoking the TranscodeHandler.
func (uc *MediaUseCase) RetryTask(ctx context.Context, taskID int) (*EncodingTask, error) {
	task, err := uc.encodingRepo.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task %d not found: %w", taskID, err)
	}

	// Only allow retrying failed or partial tasks
	if task.Status != "failed" && task.Status != "partial" {
		return nil, fmt.Errorf(
			"cannot retry task %d with status %q (only 'failed' can be retried)",
			taskID,
			task.Status,
		)
	}

	task.Status = "pending"
	task.Progress = 0
	task.ErrorMessage = ""

	updated, err := uc.encodingRepo.Update(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to reset task %d: %w", taskID, err)
	}

	uc.log.Infof(
		"task %d (media=%d profile=%d) reset to pending for retry",
		taskID,
		updated.MediaId,
		updated.ProfileId,
	)
	return updated, nil
}

// RetryAllFailedTasks resets all failed tasks for a given media back to "pending".
// Returns the count of tasks that were reset.
func (uc *MediaUseCase) RetryAllFailedTasks(ctx context.Context, mediaID int64) (int, error) {
	tasks, err := uc.encodingRepo.ListByMedia(ctx, mediaID)
	if err != nil {
		return 0, fmt.Errorf("failed to list tasks for media %d: %w", mediaID, err)
	}

	resetCount := 0
	for _, t := range tasks {
		if t.Status != "failed" && t.Status != "partial" {
			continue
		}
		t.Status = "pending"
		t.Progress = 0
		t.ErrorMessage = ""
		if _, err := uc.encodingRepo.Update(ctx, t); err != nil {
			uc.log.Warnf("failed to reset task %d during bulk retry: %v", t.Id, err)
			continue
		}
		resetCount++
	}

	uc.log.Infof("reset %d failed tasks for media %d to pending", resetCount, mediaID)
	return resetCount, nil
}

// --- Transcoding Status ---

// VariantInfo holds aggregated info about a single encoding profile's result.
// Returned by GetMediaVariants for frontend display and player configuration.
type VariantInfo struct {
	TaskID       int    `json:"task_id"`
	ProfileName  string `json:"profile_name"`
	ProfileID    int    `json:"profile_id"`
	Resolution   string `json:"resolution"`  // e.g., "1280x720" or "720"
	Codec        string `json:"codec"`       // e.g., "h264", "h265"
	Status       string `json:"status"`      // pending, processing, success, failed, skipped
	OutputPath   string `json:"output_path"` // HLS playlist path or GIF path
	Bandwidth    int    `json:"bandwidth"`   // bits per second (estimated)
	ErrorMessage string `json:"error_message,omitempty"`
}

// MediaVariantSummary is the aggregated transcoding status returned by media detail APIs.
// This is what the "media management" page displays — a compact view of all profile outcomes.
type MediaVariantSummary struct {
	MediaID              int64         `json:"media_id"`
	UUID                 string        `json:"uuid"`
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

// TranscodingMediaItem represents a media with its encoding tasks (for task-list view).
type TranscodingMediaItem struct {
	Media *Media          `json:"media"`
	Tasks []*EncodingTask `json:"tasks"`
}

// TranscodingStatus holds aggregated counts and a (filtered) page of items.
type TranscodingStatus struct {
	ProcessingCount int                     `json:"processing_count"`
	PendingCount    int                     `json:"pending_count"`
	PartialCount    int                     `json:"partial_count"`
	FailedCount     int                     `json:"failed_count"`
	SuccessCount    int                     `json:"success_count"`
	Items           []*TranscodingMediaItem `json:"items"`
	TotalFiltered   int                     `json:"total_filtered"`
	Page            int                     `json:"page"`
	PageSize        int                     `json:"page_size"`
}

// TranscodingStatusFilter controls which media are returned in the items list.
type TranscodingStatusFilter struct {
	// Status filter: "active" (default), "processing", "pending", "failed", "success", "all".
	Status string
	// Page is 1-based.  Defaults to 1.
	Page int
	// PageSize limits items returned.  Defaults to 20, max 100.
	PageSize int
}

// FlatTaskList holds a flat (non-grouped) list of encoding tasks with counts and pagination.
type FlatTaskList struct {
	ProcessingCount int            `json:"processing_count"`
	PendingCount    int            `json:"pending_count"`
	PartialCount    int            `json:"partial_count"`
	FailedCount     int            `json:"failed_count"`
	SuccessCount    int            `json:"success_count"`
	TotalFiltered   int            `json:"total_filtered"`
	Page            int            `json:"page"`
	PageSize        int            `json:"page_size"`
	Items           []FlatTaskItem `json:"items"`
}

// FlatTaskItem is a single task row for the flat task list view.
type FlatTaskItem struct {
	Id           int    `json:"id"`
	MediaId      int64  `json:"media_id"`
	MediaTitle   string `json:"media_title,omitempty"`
	Thumbnail    string `json:"thumbnail,omitempty"`
	ProfileId    int    `json:"profile_id"`
	ProfileName  string `json:"profile_name,omitempty"`
	Status       string `json:"status"`
	Progress     int    `json:"progress"`
	OutputPath   string `json:"output_path,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreateTime   string `json:"created_at,omitempty"`
	UpdateTime   string `json:"update_time,omitempty"`
}

// StatusCounts holds per-media-status counts.
type StatusCounts struct {
	Processing int `json:"processing"`
	Pending    int `json:"pending"`
	Partial    int `json:"partial"`
	Failed     int `json:"failed"`
	Success    int `json:"success"`
}

func (uc *MediaUseCase) GetTranscodingStatus(
	ctx context.Context,
	filter *TranscodingStatusFilter,
) (*TranscodingStatus, error) {
	if filter == nil {
		filter = &TranscodingStatusFilter{Status: "active", Page: 1, PageSize: 20}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	// 1. Aggregate counts by media encoding_status.
	counts, err := uc.repo.CountByEncodingStatus(ctx)
	if err != nil {
		return nil, err
	}

	status := &TranscodingStatus{
		ProcessingCount: counts.Processing,
		PendingCount:    counts.Pending,
		PartialCount:    counts.Partial,
		FailedCount:     counts.Failed,
		SuccessCount:    counts.Success,
		Page:            filter.Page,
		PageSize:        filter.PageSize,
	}

	// 2. Determine which media statuses to include.
	listStatuses := statusesForFilter(filter.Status)

	// 3. Fetch filtered media list with pagination (media-level, not task-level).
	mediaList, total, err := uc.repo.ListFilteredByEncodingStatus(
		ctx,
		listStatuses,
		filter.Page,
		filter.PageSize,
	)
	if err != nil {
		return nil, err
	}
	status.TotalFiltered = total

	// 4. For each media, fetch its encoding tasks.
	status.Items = make([]*TranscodingMediaItem, 0, len(mediaList))
	for _, m := range mediaList {
		tasks, err := uc.encodingRepo.ListByMedia(ctx, m.Id)
		if err != nil {
			uc.log.Errorf("failed to fetch tasks for media %d: %v", m.Id, err)
			tasks = nil
		}
		status.Items = append(status.Items, &TranscodingMediaItem{
			Media: m,
			Tasks: tasks,
		})
	}

	return status, nil
}

// ListEncodingTasksFlat returns a flat, paginated list of encoding tasks (one row per task).
// Used by the TranscodingStatus page for a pure task-centric view.
func (uc *MediaUseCase) ListEncodingTasksFlat(
	ctx context.Context,
	filter *TranscodingStatusFilter,
	mediaID *int64,
) (*FlatTaskList, error) {
	if filter == nil {
		filter = &TranscodingStatusFilter{Status: "active", Page: 1, PageSize: 25}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 25
	}

	status := filter.Status
	if mediaID != nil && *mediaID > 0 {
		// When filtered to a specific media, default to "all" if not specified
		if status == "" || status == "active" {
			status = "all"
		}
	}

	offset := (filter.Page - 1) * filter.PageSize

	tasks, total, err := uc.encodingRepo.ListFlat(ctx, status, mediaID, offset, filter.PageSize)
	if err != nil {
		return nil, err
	}

	// Get counts from the encoding_task table (NOT the Media table)
	counts, countErr := uc.encodingRepo.CountByStatus(ctx)
	if countErr != nil {
		counts = &StatusCounts{}
	}

	// Enrich with profile names
	items := make([]FlatTaskItem, len(tasks))
	for i, t := range tasks {
		item := FlatTaskItem{
			Id:           t.Id,
			MediaId:      t.MediaId,
			ProfileId:    t.ProfileId,
			Status:       t.Status,
			Progress:     t.Progress,
			OutputPath:   t.OutputPath,
			ErrorMessage: t.ErrorMessage,
		}
		// Look up profile name
		if profile, perr := uc.profileRepo.Get(ctx, t.ProfileId); perr == nil && profile != nil {
			item.ProfileName = profile.Name
		}
		items[i] = item
	}

	return &FlatTaskList{
		ProcessingCount: counts.Processing,
		PendingCount:    counts.Pending,
		PartialCount:    counts.Partial,
		FailedCount:     counts.Failed,
		SuccessCount:    counts.Success,
		TotalFiltered:   total,
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
	media, err := uc.repo.Get(ctx, mediaID)
	if err != nil {
		return nil, err
	}

	// 2. Load all encoding tasks for this media
	tasks, err := uc.encodingRepo.ListByMedia(ctx, mediaID)
	if err != nil {
		// No tasks yet — return minimal info
		return &MediaVariantSummary{
			MediaID:        mediaID,
			UUID:           media.Uuid,
			EncodingStatus: media.EncodingStatus,
			HlsFile:        media.HlsFile,
			Thumbnail:      media.Thumbnail,
			PreviewFile:    media.PreviewFile,
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
			uc.log.Warnf("profile %d not found for task %d", t.ProfileId, t.Id)
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
			case "failed", "skipped":
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
		MediaID:              mediaID,
		UUID:                 media.Uuid,
		EncodingStatus:       media.EncodingStatus,
		HlsFile:              media.HlsFile,
		Thumbnail:            media.Thumbnail,
		PreviewFile:          media.PreviewFile,
		VideoTotalCount:      videoTotalCount,
		VideoSuccessCount:    videoSuccessCount,
		VideoFailedCount:     videoFailedCount,
		VideoPendingCount:    videoPendingCount,
		VideoProcessingCount: videoProcessingCount,
		Variants:             variants,
	}, nil
}

// IsPreviewProfileFromName returns true if the profile name indicates it's a preview/GIF type.
func IsPreviewProfileFromName(name string) bool {
	return strings.EqualFold(name, "preview") || strings.EqualFold(name, "gif")
}

// estimateProfileBandwidth estimates bandwidth in bps from profile settings.
func estimateProfileBandwidth(p *EncodeProfile) int {
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
// Pass mediaID=0 to subscribe to all media events (dashboard use).
func (uc *MediaUseCase) Subscribe(
	ctx context.Context,
	mediaID int64,
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
// and also broadcasts to the global subscriber (mediaID=0).
func (uc *MediaUseCase) Publish(mediaID int64, event *EncodingEvent) {
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
	for _, ch := range uc.subs[0] {
		select {
		case ch <- event:
		default:
		}
	}
}

// --- Encode Profiles ---

// ListEncodeProfiles returns all encoding profiles.
func (uc *MediaUseCase) ListEncodeProfiles(ctx context.Context) ([]*EncodeProfile, error) {
	return uc.profileRepo.ListAll(ctx)
}

// GetEncodeProfile returns an encoding profile by ID.
func (uc *MediaUseCase) GetEncodeProfile(ctx context.Context, id int) (*EncodeProfile, error) {
	return uc.profileRepo.Get(ctx, id)
}

// CreateEncodeProfile creates a new encoding profile.
func (uc *MediaUseCase) CreateEncodeProfile(
	ctx context.Context,
	profile *EncodeProfile,
) (*EncodeProfile, error) {
	return uc.profileRepo.Create(ctx, profile)
}

// UpdateEncodeProfile updates an existing encoding profile.
func (uc *MediaUseCase) UpdateEncodeProfile(
	ctx context.Context,
	profile *EncodeProfile,
) (*EncodeProfile, error) {
	return uc.profileRepo.Update(ctx, profile)
}

// DeleteEncodeProfile deletes an encoding profile.
func (uc *MediaUseCase) DeleteEncodeProfile(ctx context.Context, id int) error {
	return uc.profileRepo.Delete(ctx, id)
}
