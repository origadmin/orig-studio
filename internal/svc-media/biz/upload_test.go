/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/enums"
	"origadmin/application/origcms/internal/svc-media/dto"
)

// MockUploadRepo 模拟上传仓库
type MockUploadRepo struct {
	sessions map[string]*UploadSession
}

func NewMockUploadRepo() *MockUploadRepo {
	return &MockUploadRepo{
		sessions: make(map[string]*UploadSession),
	}
}

func (r *MockUploadRepo) CreateSession(ctx context.Context, session *UploadSession) error {
	r.sessions[session.UploadID] = session
	return nil
}

func (r *MockUploadRepo) GetSession(ctx context.Context, uploadID string) (*UploadSession, error) {
	session, ok := r.sessions[uploadID]
	if !ok {
		return nil, fmt.Errorf("entity: upload_session not found")
	}
	return session, nil
}

func (r *MockUploadRepo) GetSessionByID(ctx context.Context, uploadID string) (*UploadSession, error) {
	session, ok := r.sessions[uploadID]
	if !ok {
		return nil, fmt.Errorf("entity: upload_session not found")
	}
	return session, nil
}

func (r *MockUploadRepo) UpdateSession(ctx context.Context, session *UploadSession) error {
	r.sessions[session.UploadID] = session
	return nil
}

func (r *MockUploadRepo) DeleteSession(ctx context.Context, uploadID string) error {
	delete(r.sessions, uploadID)
	return nil
}

func (r *MockUploadRepo) ListSessions(ctx context.Context, userID string, status enums.UploadStatus, page, pageSize int) ([]*UploadSession, int, error) {
	var sessions []*UploadSession
	for _, session := range r.sessions {
		if (userID == "" || session.UserID != nil && *session.UserID == userID) &&
			(status == "" || session.Status == status) {
			sessions = append(sessions, session)
		}
	}
	return sessions, len(sessions), nil
}

func (r *MockUploadRepo) DeleteExpiredSessions(ctx context.Context, now time.Time) ([]string, error) {
	var deletedIDs []string
	for id, session := range r.sessions {
		if session.ExpiresAt.Before(now) {
			delete(r.sessions, id)
			deletedIDs = append(deletedIDs, id)
		}
	}
	return deletedIDs, nil
}

// MockReadCloser 是 bytes.Reader 的包装器，实现 io.ReadCloser 接口
type MockReadCloser struct {
	*bytes.Reader
}

// Close 实现 io.ReadCloser 接口的 Close 方法
func (m *MockReadCloser) Close() error {
	return nil
}

// MockStorage 模拟存储
type MockStorage struct {
	parts     map[string]map[int][]byte
	files     map[string][]byte
	deleteAll bool
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		parts: make(map[string]map[int][]byte),
		files: make(map[string][]byte),
	}
}

func (s *MockStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	s.files[key] = data
	return key, nil
}

func (s *MockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return &MockReadCloser{Reader: bytes.NewReader(s.files[key])}, nil
}

func (s *MockStorage) Delete(ctx context.Context, key string) error {
	delete(s.files, key)
	return nil
}

func (s *MockStorage) GetURL(ctx context.Context, key string) (string, error) {
	return "http://localhost:8080/" + key, nil
}

func (s *MockStorage) StorePart(ctx context.Context, uploadID string, partNumber int, data []byte) (string, error) {
	if _, ok := s.parts[uploadID]; !ok {
		s.parts[uploadID] = make(map[int][]byte)
	}
	s.parts[uploadID][partNumber] = data
	return "etag", nil
}

func (s *MockStorage) MergeParts(ctx context.Context, uploadID string, totalParts int, finalPath string) error {
	var merged []byte
	for i := 1; i <= totalParts; i++ {
		if part, ok := s.parts[uploadID][i]; ok {
			merged = append(merged, part...)
		}
	}
	s.files[finalPath] = merged
	return nil
}

func (s *MockStorage) DeleteParts(ctx context.Context, uploadID string) error {
	delete(s.parts, uploadID)
	s.deleteAll = true
	return nil
}

// MockMediaRepo 模拟媒体仓库
type MockMediaRepo struct {
	media map[string]*Media
}

func NewMockMediaRepo() *MockMediaRepo {
	return &MockMediaRepo{
		media: make(map[string]*Media),
	}
}

func (r *MockMediaRepo) Create(ctx context.Context, media *Media, opts ...*dto.MediaCreateOption) (*Media, error) {
	r.media[media.Id] = media
	return media, nil
}

func (r *MockMediaRepo) Get(ctx context.Context, id string, opts ...*dto.MediaQueryOption) (*Media, error) {
	return r.media[id], nil
}

func (r *MockMediaRepo) List(ctx context.Context, opts ...*dto.MediaQueryOption) ([]*Media, int32, error) {
	var mediaList []*Media
	for _, media := range r.media {
		mediaList = append(mediaList, media)
	}
	return mediaList, int32(len(mediaList)), nil
}

func (r *MockMediaRepo) Update(ctx context.Context, media *Media, opts ...*dto.MediaUpdateOption) (*Media, error) {
	r.media[media.Id] = media
	return media, nil
}

func (r *MockMediaRepo) Delete(ctx context.Context, id string) error {
	delete(r.media, id)
	return nil
}

func (r *MockMediaRepo) CreateWithEntity(ctx context.Context, media *Media) (*entity.Media, *Media, error) {
	r.media[media.Id] = media
	return &entity.Media{ID: media.Id}, media, nil
}

func (r *MockMediaRepo) ListCategories(ctx context.Context, opts ...*dto.CategoryQueryOption) ([]*types.Category, int32, error) {
	return nil, 0, nil
}

func (r *MockMediaRepo) GetCategory(ctx context.Context, id string) (*types.Category, error) {
	return nil, nil
}

func (r *MockMediaRepo) IncrementViewCount(ctx context.Context, id string) (int64, error) {
	return 0, nil
}

func (r *MockMediaRepo) UpdateCommentCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockMediaRepo) UpdateLikeCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockMediaRepo) UpdateDislikeCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockMediaRepo) UpdateFavoriteCount(ctx context.Context, id string, delta int) error {
	return nil
}

func (r *MockMediaRepo) ResetStaleProcessing(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *MockMediaRepo) CountByEncodingStatus(ctx context.Context) (*dto.StatusCounts, error) {
	return &dto.StatusCounts{}, nil
}

func (r *MockMediaRepo) ListFilteredByEncodingStatus(ctx context.Context, statuses []string, page, pageSize int) ([]*Media, int, error) {
	var mediaList []*Media
	for _, media := range r.media {
		mediaList = append(mediaList, media)
	}
	return mediaList, len(mediaList), nil
}

// MockEncodeProfileRepo 模拟编码配置仓库
type MockEncodeProfileRepo struct {
	profiles map[int]*dto.EncodeProfile
}

func NewMockEncodeProfileRepo() *MockEncodeProfileRepo {
	return &MockEncodeProfileRepo{
		profiles: make(map[int]*dto.EncodeProfile),
	}
}

func (r *MockEncodeProfileRepo) Create(ctx context.Context, profile *dto.EncodeProfile) (*dto.EncodeProfile, error) {
	r.profiles[profile.Id] = profile
	return profile, nil
}

func (r *MockEncodeProfileRepo) Get(ctx context.Context, id int) (*dto.EncodeProfile, error) {
	return r.profiles[id], nil
}

func (r *MockEncodeProfileRepo) Update(ctx context.Context, profile *dto.EncodeProfile) (*dto.EncodeProfile, error) {
	r.profiles[profile.Id] = profile
	return profile, nil
}

func (r *MockEncodeProfileRepo) Delete(ctx context.Context, id int) error {
	delete(r.profiles, id)
	return nil
}

func (r *MockEncodeProfileRepo) ListActive(ctx context.Context) ([]*dto.EncodeProfile, error) {
	var profiles []*dto.EncodeProfile
	for _, profile := range r.profiles {
		if profile.IsActive {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

func (r *MockEncodeProfileRepo) ListAll(ctx context.Context) ([]*dto.EncodeProfile, error) {
	var profiles []*dto.EncodeProfile
	for _, profile := range r.profiles {
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func (r *MockEncodeProfileRepo) GetByName(ctx context.Context, name string) (*dto.EncodeProfile, error) {
	for _, profile := range r.profiles {
		if profile.Name == name {
			return profile, nil
		}
	}
	return nil, nil
}

// MockEncodingTaskRepo 模拟编码任务仓库
type MockEncodingTaskRepo struct {
	tasks map[string]*dto.EncodingTask
}

func NewMockEncodingTaskRepo() *MockEncodingTaskRepo {
	return &MockEncodingTaskRepo{
		tasks: make(map[string]*dto.EncodingTask),
	}
}

func (r *MockEncodingTaskRepo) Create(ctx context.Context, task *dto.EncodingTask) (*dto.EncodingTask, error) {
	r.tasks[task.Id] = task
	return task, nil
}

func (r *MockEncodingTaskRepo) Get(ctx context.Context, id string) (*dto.EncodingTask, error) {
	return r.tasks[id], nil
}

func (r *MockEncodingTaskRepo) Update(ctx context.Context, task *dto.EncodingTask) (*dto.EncodingTask, error) {
	r.tasks[task.Id] = task
	return task, nil
}

func (r *MockEncodingTaskRepo) Delete(ctx context.Context, id string) error {
	delete(r.tasks, id)
	return nil
}

func (r *MockEncodingTaskRepo) ListByMedia(ctx context.Context, mediaID string) ([]*dto.EncodingTask, error) {
	var tasks []*dto.EncodingTask
	for _, task := range r.tasks {
		if task.MediaId == mediaID {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *MockEncodingTaskRepo) DeleteByMedia(ctx context.Context, mediaID string) error {
	for id, task := range r.tasks {
		if task.MediaId == mediaID {
			delete(r.tasks, id)
		}
	}
	return nil
}

func (r *MockEncodingTaskRepo) ListFlat(ctx context.Context, status string, mediaId *string, profileFilter string, profileID int, chunkFilter string, searchQuery string, offset, limit int) ([]*dto.EncodingTask, int, error) {
	var tasks []*dto.EncodingTask
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, len(tasks), nil
}

func (r *MockEncodingTaskRepo) CountByStatus(ctx context.Context) (*dto.StatusCounts, error) {
	return &dto.StatusCounts{}, nil
}

func (r *MockEncodingTaskRepo) CountByStatusWithFilter(ctx context.Context, status string, mediaId *string, profileFilter string, profileID int, chunkFilter string, searchQuery string) (*dto.StatusCounts, error) {
	return &dto.StatusCounts{}, nil
}

func TestUploadUseCase_InitiateMultipartUpload(t *testing.T) {
	repo := NewMockUploadRepo()
	mediaRepo := NewMockMediaRepo()
	profileRepo := NewMockEncodeProfileRepo()
	encodingRepo := NewMockEncodingTaskRepo()
	storage := NewMockStorage()
	logger := log.NewStdLogger(os.Stdout)
	
	uc := NewUploadUseCase(
		repo,
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		storage,
		5*1024*1024, // 5MB
		logger,
	)
	
	ctx := context.Background()
	session, err := uc.InitiateMultipartUpload(
		ctx,
		"test.mp4",
		10*1024*1024, // 10MB
		"video/mp4",
		"Test Video",
		"Test Description",
		nil,
		[]string{"test", "video"},
		"",
		nil,
	)
	
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "test.mp4", session.Filename)
	assert.Equal(t, int64(10*1024*1024), session.FileSize)
	assert.Equal(t, "video/mp4", session.ContentType)
	assert.Equal(t, 2, session.TotalParts) // 10MB / 5MB per part
	assert.Equal(t, StatusPending, session.Status)
}

func TestUploadUseCase_UploadPart(t *testing.T) {
	repo := NewMockUploadRepo()
	mediaRepo := NewMockMediaRepo()
	profileRepo := NewMockEncodeProfileRepo()
	encodingRepo := NewMockEncodingTaskRepo()
	storage := NewMockStorage()
	logger := log.NewStdLogger(os.Stdout)
	
	uc := NewUploadUseCase(
		repo,
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		storage,
		5*1024*1024, // 5MB
		logger,
	)
	
	ctx := context.Background()
	
	// 初始化上传
	session, err := uc.InitiateMultipartUpload(
		ctx,
		"test.mp4",
		10*1024*1024, // 10MB
		"video/mp4",
		"Test Video",
		"Test Description",
		nil,
		[]string{"test", "video"},
		"",
		nil,
	)
	assert.NoError(t, err)
	
	// 上传分片
	data := make([]byte, 5*1024*1024) // 5MB
	etag, err := uc.UploadPart(ctx, session.UploadID, 1, data)
	assert.NoError(t, err)
	assert.NotEmpty(t, etag)
	
	// 验证分片存储
	assert.Contains(t, storage.parts, session.UploadID)
	assert.Contains(t, storage.parts[session.UploadID], 1)
	
	// 验证会话更新
	updatedSession, err := repo.GetSession(ctx, session.UploadID)
	assert.NoError(t, err)
	assert.Equal(t, StatusUploading, updatedSession.Status)
	assert.Equal(t, int64(len(data)), updatedSession.UploadedSize)
}

func TestUploadUseCase_CompleteMultipartUpload(t *testing.T) {
	repo := NewMockUploadRepo()
	mediaRepo := NewMockMediaRepo()
	profileRepo := NewMockEncodeProfileRepo()
	encodingRepo := NewMockEncodingTaskRepo()
	storage := NewMockStorage()
	logger := log.NewStdLogger(os.Stdout)
	
	uc := NewUploadUseCase(
		repo,
		mediaRepo,
		profileRepo,
		encodingRepo,
		nil,
		storage,
		5*1024*1024, // 5MB
		logger,
	)
	
	ctx := context.Background()
	
	// 初始化上传
	session, err := uc.InitiateMultipartUpload(
		ctx,
		"test.mp4",
		10*1024*1024, // 10MB
		"video/mp4",
		"Test Video",
		"Test Description",
		nil,
		[]string{"test", "video"},
		"",
		nil,
	)
	assert.NoError(t, err)
	
	// 上传分片
	data := make([]byte, 5*1024*1024) // 5MB
	_, err = uc.UploadPart(ctx, session.UploadID, 1, data)
	assert.NoError(t, err)
	_, err = uc.UploadPart(ctx, session.UploadID, 2, data)
	assert.NoError(t, err)
	
	// 完成上传
	media, err := uc.CompleteMultipartUpload(
		ctx,
		session.UploadID,
		"sha256hash",
		"",
		"",
		nil,
		nil,
		"",
	)
	
	assert.NoError(t, err)
	assert.NotNil(t, media)
	assert.Equal(t, "Test Video", media.Title)
	assert.Equal(t, "Test Description", media.Description)
	assert.Equal(t, "video/mp4", media.MimeType)
	
	// 验证文件合并
	assert.Contains(t, storage.files, session.UploadID+".mp4")
	
	// 验证临时分片删除
	assert.True(t, storage.deleteAll)
	
	// 验证会话状态更新
	updatedSession, err := repo.GetSession(ctx, session.UploadID)
	assert.NoError(t, err)
	assert.Equal(t, StatusCompleted, updatedSession.Status)
	assert.Equal(t, "sha256hash", updatedSession.Sha256)
}
