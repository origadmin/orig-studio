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
	items, err := r.db.EncodingTask.Query().
		Where(encodingtask.MediaIDEQ(int(mediaId))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*biz.EncodingTask, len(items))
	for i, item := range items {
		result[i] = convertEncodingTaskToBiz(item)
	}
	return result, nil
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
