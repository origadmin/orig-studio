package data

import (
	"context"
	"fmt"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/encodingtask"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/svc-media/biz"
)

type encodingTaskRepo struct {
	db *entity.Client
}

// NewEncodingTaskRepo creates a new EncodingTask repository.
func NewEncodingTaskRepo(db *entity.Client) biz.EncodingTaskRepo {
	return &encodingTaskRepo{db: db}
}

func (r *encodingTaskRepo) Create(
	ctx context.Context,
	task *biz.EncodingTask,
) (*biz.EncodingTask, error) {
	m, err := r.db.EncodingTask.Create().
		SetMediaID(task.MediaId).
		SetProfileID(task.ProfileId).
		SetStatus(task.Status).
		SetOutputPath(task.OutputPath).
		SetErrorMessage(task.ErrorMessage).
		SetChunk(task.Chunk).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertEncodingTaskToBiz(m), nil
}

func (r *encodingTaskRepo) Update(
	ctx context.Context,
	task *biz.EncodingTask,
) (*biz.EncodingTask, error) {
	update := r.db.EncodingTask.UpdateOneID(task.Id).
		SetStatus(task.Status)

	if task.OutputPath != "" {
		update = update.SetOutputPath(task.OutputPath)
	}
	if task.ErrorMessage != "" {
		update = update.SetErrorMessage(task.ErrorMessage)
	}

	m, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertEncodingTaskToBiz(m), nil
}

func (r *encodingTaskRepo) Get(ctx context.Context, id string) (*biz.EncodingTask, error) {
	m, err := r.db.EncodingTask.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertEncodingTaskToBiz(m), nil
}

func (r *encodingTaskRepo) ListByMedia(
	ctx context.Context,
	mediaId string,
) ([]*biz.EncodingTask, error) {
	query := r.db.EncodingTask.Query()
	if mediaId != "" {
		query = query.Where(encodingtask.MediaIDEQ(mediaId))
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.EncodingTask, len(items))
	for i, item := range items {
		result[i] = convertEncodingTaskToBiz(item)
	}
	return result, nil
}

// DeleteByMedia deletes all encoding tasks for a given media ID.
func (r *encodingTaskRepo) DeleteByMedia(ctx context.Context, mediaID string) error {
	_, err := r.db.EncodingTask.Delete().
		Where(encodingtask.MediaIDEQ(mediaID)).
		Exec(ctx)
	return err
}

// ListFlat returns a paginated flat list of encoding tasks, optionally filtered by status, media_id, profile, chunk, and search query.
func (r *encodingTaskRepo) ListFlat(
	ctx context.Context,
	status string,
	mediaId *string,
	profileFilter string,
	chunkFilter string,
	searchQuery string,
	offset, limit int,
) ([]*biz.EncodingTask, int, error) {
	query := r.db.EncodingTask.Query()
	if status != "" && status != "all" {
		// Special handling for active status: exclude success
		if status == "active" {
			query = query.Where(
				encodingtask.StatusIn("pending", "processing", "partial", "failed", "skipped"),
			)
		} else if status == "failed" {
			// Special handling for failure statuses: "failed" includes both "failed" and "skipped"
			query = query.Where(encodingtask.StatusIn("failed", "skipped"))
		} else if status == "skipped" {
			// For compatibility, "skipped" still returns only skipped tasks
			query = query.Where(encodingtask.StatusEQ("skipped"))
		} else {
			query = query.Where(encodingtask.StatusEQ(enums.EncodingTaskStatus(status)))
		}
	}
	if mediaId != nil && *mediaId != "" {
		query = query.Where(encodingtask.MediaIDEQ(*mediaId))
	}

	// Profile filter (partial match on profile name)
	if profileFilter != "" {
		// TODO: Add profile filter support when available
	}

	// Chunk filter (boolean: true/false)
	if chunkFilter != "" {
		if chunkFilter == "true" {
			query = query.Where(encodingtask.ChunkEQ(true))
		} else if chunkFilter == "false" {
			query = query.Where(encodingtask.ChunkEQ(false))
		}
	}

	// Search query (search media_id, status, or profile name)
	if searchQuery != "" {
		// TODO: Add search functionality when available
	}

	// Count total matching
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Paginate
	items, err := query.
		Order(entity.Desc(encodingtask.FieldUpdatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*biz.EncodingTask, len(items))
	for i, item := range items {
		result[i] = convertEncodingTaskToBiz(item)
	}
	return result, total, nil
}

func convertEncodingTaskToBiz(m *entity.EncodingTask) *biz.EncodingTask {
	return &biz.EncodingTask{
		Id:           m.ID,
		MediaId:      m.MediaID,
		ProfileId:    m.ProfileID,
		Status:       enums.EncodingTaskStatus(m.Status),
		OutputPath:   m.OutputPath,
		ErrorMessage: m.ErrorMessage,
		Chunk:        m.Chunk,
	}
}

// CountByStatus returns per-status counts from the encoding_task table.
// This is the correct data source for task-level status counts (unlike
// mediaRepo.CountByEncodingStatus which queries the Media table).
func (r *encodingTaskRepo) CountByStatus(ctx context.Context) (*biz.StatusCounts, error) {
	type countRow struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	var rows []countRow
	err := r.db.EncodingTask.Query().
		GroupBy(encodingtask.FieldStatus).
		Aggregate(entity.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	counts := &biz.StatusCounts{}
	for _, row := range rows {
		switch row.Status {
		case "processing":
			counts.Processing = row.Count
		case "pending":
			counts.Pending = row.Count
		case "partial":
			counts.Partial = row.Count
		case "skipped", "failed":
			// "skipped" is our actual failure status; count it as Failed for UI display
			counts.Failed += row.Count
		case "success":
			counts.Success = row.Count
		}
	}
	return counts, nil
}

// CountByStatusWithFilter returns per-status counts filtered by status, media_id, profile, chunk, and search query.
func (r *encodingTaskRepo) CountByStatusWithFilter(
	ctx context.Context,
	status string,
	mediaId *string,
	profileFilter string,
	chunkFilter string,
	searchQuery string,
) (*biz.StatusCounts, error) {
	// Log the profileFilter to debug
	if profileFilter != "" {
		fmt.Printf("CountByStatusWithFilter: profileFilter=%s\n", profileFilter)
	}
	type countRow struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	// 先获取符合条件的任务ID
	baseQuery := r.db.EncodingTask.Query()
	if status != "" && status != "all" {
		// Special handling for active status: exclude success
		if status == "active" {
			baseQuery = baseQuery.Where(
				encodingtask.StatusIn("pending", "processing", "partial", "failed", "skipped"),
			)
		} else if status == "failed" {
			// Special handling for failure statuses: "failed" includes both "failed" and "skipped"
			baseQuery = baseQuery.Where(encodingtask.StatusIn("failed", "skipped"))
		} else if status == "skipped" {
			// For compatibility, "skipped" still returns only skipped tasks
			baseQuery = baseQuery.Where(encodingtask.StatusEQ("skipped"))
		} else {
			baseQuery = baseQuery.Where(encodingtask.StatusEQ(enums.EncodingTaskStatus(status)))
		}
	}
	if mediaId != nil && *mediaId != "" {
		baseQuery = baseQuery.Where(encodingtask.MediaIDEQ(*mediaId))
	}

	// Profile filter (partial match on profile name)
	if profileFilter != "" {
		// TODO: Add profile filter support when available
	}

	// Chunk filter (boolean: true/false)
	if chunkFilter != "" {
		if chunkFilter == "true" {
			baseQuery = baseQuery.Where(encodingtask.ChunkEQ(true))
		} else if chunkFilter == "false" {
			baseQuery = baseQuery.Where(encodingtask.ChunkEQ(false))
		}
	}

	// Search query (search media_id, status, or profile name)
	if searchQuery != "" {
		// TODO: Add search functionality when available
	}

	// 构建子查询获取符合条件的任务ID
	taskIDs, err := baseQuery.Select(encodingtask.FieldID).All(ctx)
	if err != nil {
		return nil, err
	}

	// 如果没有符合条件的任务，返回空计数
	if len(taskIDs) == 0 {
		return &biz.StatusCounts{}, nil
	}

	// 构建ID列表
	ids := make([]string, len(taskIDs))
	for i, task := range taskIDs {
		ids[i] = task.ID
	}

	// 使用子查询结果进行分组计数
	query := r.db.EncodingTask.Query().Where(encodingtask.IDIn(ids...))

	var rows []countRow
	err = query.
		GroupBy(encodingtask.FieldStatus).
		Aggregate(entity.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	counts := &biz.StatusCounts{}
	for _, row := range rows {
		switch row.Status {
		case "processing":
			counts.Processing = row.Count
		case "pending":
			counts.Pending = row.Count
		case "partial":
			counts.Partial = row.Count
		case "skipped", "failed":
			// "skipped" is our actual failure status; count it as Failed for UI display
			counts.Failed += row.Count
		case "success":
			counts.Success = row.Count
		}
	}
	return counts, nil
}
