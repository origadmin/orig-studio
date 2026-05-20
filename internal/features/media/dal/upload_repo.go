package dal

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/uploadsession"
	"origadmin/application/origstudio/internal/dal/enums"
	"origadmin/application/origstudio/internal/features/media/dto"
)

type uploadRepo struct {
	data *entity.Client
	log  *log.Helper
}

func NewUploadRepo(data *entity.Client, logger log.Logger) dto.UploadRepo {
	return &uploadRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data/upload")),
	}
}

func (r *uploadRepo) CreateSession(ctx context.Context, session *dto.UploadSession) error {
	builder := r.data.UploadSession.Create().
		SetUploadID(session.UploadID).
		SetFilename(session.Filename).
		SetFileSize(session.FileSize).
		SetContentType(session.ContentType).
		SetTotalParts(session.TotalParts).
		SetChunkSize(session.ChunkSize).
		SetUploadedSize(session.UploadedSize).
		SetTitle(session.Title).
		SetDescription(session.Description).
		SetTags(session.Tags).
		SetStatus(string(session.Status)).
		SetParts(session.Parts).
		SetSha256(session.Sha256).
		SetStoragePath(session.StoragePath).
		SetTempDir(session.TempDir).
		SetExpiresAt(session.ExpiresAt)

	if session.CategoryID != nil && *session.CategoryID != 0 {
		builder.SetCategoryID(*session.CategoryID)
	}
	if session.UserID != nil && *session.UserID != "" {
		builder.SetUserID(*session.UserID)
	}

	_, err := builder.Save(ctx)
	return err
}

func (r *uploadRepo) GetSession(ctx context.Context, uploadID string) (*dto.UploadSession, error) {
	entSession, err := r.data.UploadSession.Query().
		Where(uploadsession.UploadID(uploadID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return r.entToBiz(entSession), nil
}

func (r *uploadRepo) UpdateSession(ctx context.Context, session *dto.UploadSession) error {
	return r.data.UploadSession.Update().
		Where(uploadsession.UploadID(session.UploadID)).
		SetUploadedSize(session.UploadedSize).
		SetStatus(string(session.Status)).
		SetParts(session.Parts).
		SetSha256(session.Sha256).
		SetStoragePath(session.StoragePath).
		SetUpdateTime(time.Now()).
		Exec(ctx)
}

func (r *uploadRepo) DeleteSession(ctx context.Context, uploadID string) error {
	_, err := r.data.UploadSession.Delete().
		Where(uploadsession.UploadID(uploadID)).
		Exec(ctx)
	return err
}

func (r *uploadRepo) ListSessions(ctx context.Context, userID string, status enums.UploadStatus, page, pageSize int) ([]*dto.UploadSession, int, error) {
	query := r.data.UploadSession.Query()
	if userID != "" {
		query = query.Where(uploadsession.UserIDEQ(userID))
	}
	if status != "" {
		query = query.Where(uploadsession.Status(string(status)))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	list, err := query.Offset(offset).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*dto.UploadSession, len(list))
	for i, s := range list {
		res[i] = r.entToBiz(s)
	}
	return res, total, nil
}

func (r *uploadRepo) DeleteExpiredSessions(ctx context.Context, now time.Time) ([]string, error) {
	expired, err := r.data.UploadSession.Query().
		Where(
			uploadsession.ExpiresAtLT(now),
			uploadsession.StatusNEQ(string(enums.UploadStatusCompleted)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(expired) == 0 {
		return nil, nil
	}

	ids := make([]string, len(expired))
	for i, s := range expired {
		ids[i] = s.UploadID
	}

	_, err = r.data.UploadSession.Delete().
		Where(uploadsession.UploadIDIn(ids...)).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *uploadRepo) entToBiz(s *entity.UploadSession) *dto.UploadSession {
	var categoryID *int64
	if s.CategoryID != 0 {
		categoryID = &s.CategoryID
	}

	var userID *string
	if s.UserID != "" {
		userID = &s.UserID
	}

	return &dto.UploadSession{
		UploadID:     s.UploadID,
		Filename:     s.Filename,
		FileSize:     s.FileSize,
		ContentType:  s.ContentType,
		TotalParts:   s.TotalParts,
		ChunkSize:    s.ChunkSize,
		UploadedSize: s.UploadedSize,
		Title:        s.Title,
		Description:  s.Description,
		CategoryID:   categoryID,
		Tags:         s.Tags,
		UserID:       userID,
		Status:       enums.UploadStatus(s.Status),
		Parts:        s.Parts,
		Sha256:       s.Sha256,
		StoragePath:  s.StoragePath,
		TempDir:      s.TempDir,
		ExpiresAt:    s.ExpiresAt,
		CreateTime:    s.CreateTime,
		UpdateTime:    s.UpdateTime,
	}
}
