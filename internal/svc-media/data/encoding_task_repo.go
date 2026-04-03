package data

import (
	"context"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/encodingtask"
	"origadmin/application/origcms/internal/svc-media/biz"
)

type encodingTaskRepo struct {
	db *entity.Client
}

// NewEncodingTaskRepo creates a new EncodingTask repository.
func NewEncodingTaskRepo(db *entity.Client) biz.EncodingTaskRepo {
	return &encodingTaskRepo{db: db}
}

func (r *encodingTaskRepo) Create(ctx context.Context, task *biz.EncodingTask) (*biz.EncodingTask, error) {
	m, err := r.db.EncodingTask.Create().
		SetMediaID(int(task.MediaId)).
		SetProfileID(task.ProfileId).
		SetStatus(task.Status).
		SetProgress(task.Progress).
		SetOutputPath(task.OutputPath).
		SetErrorMessage(task.ErrorMessage).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertEncodingTaskToBiz(m), nil
}

func (r *encodingTaskRepo) Update(ctx context.Context, task *biz.EncodingTask) (*biz.EncodingTask, error) {
	update := r.db.EncodingTask.UpdateOneID(task.Id).
		SetStatus(task.Status).
		SetProgress(task.Progress)

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

func (r *encodingTaskRepo) Get(ctx context.Context, id int) (*biz.EncodingTask, error) {
	m, err := r.db.EncodingTask.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertEncodingTaskToBiz(m), nil
}

func (r *encodingTaskRepo) ListByMedia(ctx context.Context, mediaId int64) ([]*biz.EncodingTask, error) {
	query := r.db.EncodingTask.Query()
	if mediaId > 0 {
		query = query.Where(encodingtask.MediaIDEQ(int(mediaId)))
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
func (r *encodingTaskRepo) DeleteByMedia(ctx context.Context, mediaID int64) error {
	_, err := r.db.EncodingTask.Delete().
		Where(encodingtask.MediaIDEQ(int(mediaID))).
		Exec(ctx)
	return err
}

// ListFlat returns a paginated flat list of encoding tasks, optionally filtered by status and/or media_id.
func (r *encodingTaskRepo) ListFlat(ctx context.Context, status string, mediaId *int64, offset, limit int) ([]*biz.EncodingTask, int, error) {
	query := r.db.EncodingTask.Query()
	if status != "" && status != "all" && status != "active" {
		query = query.Where(encodingtask.StatusEQ(status))
	} else if status == "active" {
		query = query.Where(
			encodingtask.StatusIn("pending", "processing", "partial", "failed"),
		)
	}
	if mediaId != nil && *mediaId > 0 {
		query = query.Where(encodingtask.MediaIDEQ(int(*mediaId)))
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
		MediaId:      int64(m.MediaID),
		ProfileId:    m.ProfileID,
		Status:       m.Status,
		Progress:     m.Progress,
		OutputPath:   m.OutputPath,
		ErrorMessage: m.ErrorMessage,
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
