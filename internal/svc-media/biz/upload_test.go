/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"origadmin/application/origcms/internal/svc-media/dto"
)

// MockRepo is a mock of UploadRepo
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateSession(ctx context.Context, session *UploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepo) GetSession(ctx context.Context, uploadID string) (*UploadSession, error) {
	args := m.Called(ctx, uploadID)
	return args.Get(0).(*UploadSession), args.Error(1)
}

func (m *MockRepo) UpdateSession(ctx context.Context, session *UploadSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepo) DeleteSession(ctx context.Context, uploadID string) error {
	return nil
}

func (m *MockRepo) ListSessions(
	ctx context.Context,
	userID int64,
	status string,
	page, pageSize int,
) ([]*UploadSession, int, error) {
	return nil, 0, nil
}

func (m *MockRepo) DeleteExpiredSessions(ctx context.Context, now time.Time) ([]string, error) {
	return nil, nil
}

// MockMediaRepo
type MockMediaRepo struct {
	mock.Mock
}

func (m *MockMediaRepo) Create(ctx context.Context, media *Media) (*Media, error) {
	args := m.Called(ctx, media)
	return args.Get(0).(*Media), args.Error(1)
}
func (m *MockMediaRepo) Get(ctx context.Context, id int64) (*Media, error) { return nil, nil }

func (m *MockMediaRepo) List(
	ctx context.Context,
	opts ...*dto.MediaQueryOption,
) ([]*Media, int32, error) {
	return nil, 0, nil
}

func (m *MockMediaRepo) Update(
	ctx context.Context,
	media *Media,
) (*Media, error) {
	return nil, nil
}
func (m *MockMediaRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *MockMediaRepo) IncrementViewCount(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

// MockEncodeProfileRepo
type MockEncodeProfileRepo struct {
	mock.Mock
}

func (m *MockEncodeProfileRepo) ListActive(ctx context.Context) ([]*EncodeProfile, error) {
	return nil, nil
}
func (m *MockEncodeProfileRepo) Get(ctx context.Context, id int) (*EncodeProfile, error) {
	return nil, nil
}

// MockEncodingTaskRepo
type MockEncodingTaskRepo struct {
	mock.Mock
}

func (m *MockEncodingTaskRepo) Create(ctx context.Context, task *EncodingTask) (*EncodingTask, error) {
	return task, nil
}
func (m *MockEncodingTaskRepo) Update(ctx context.Context, task *EncodingTask) (*EncodingTask, error) {
	return task, nil
}
func (m *MockEncodingTaskRepo) Get(ctx context.Context, id int) (*EncodingTask, error) {
	return nil, nil
}
func (m *MockEncodingTaskRepo) ListByMedia(ctx context.Context, mediaId int64) ([]*EncodingTask, error) {
	return nil, nil
}

func TestUploadWorkflow(t *testing.T) {
	// Setup
	tempDir, _ := os.MkdirTemp("", "upload-test-*")
	defer os.RemoveAll(tempDir)

	// We'll use a real localStorage for testing the file logic
	// but mocks for the DB repos
	logger := log.DefaultLogger
	storage := &testStorage{baseDir: tempDir}
	repo := new(MockRepo)
	mediaRepo := new(MockMediaRepo)
	profileRepo := new(MockEncodeProfileRepo)
	taskRepo := new(MockEncodingTaskRepo)
	mediaUC := NewMediaUseCase(mediaRepo, profileRepo, taskRepo, logger)

	uc := NewUploadUseCase(repo, mediaRepo, profileRepo, taskRepo, mediaUC, storage, logger)
	uc.chunkSize = 10 // Small chunk size for testing

	ctx := context.Background()
	filename := "test.txt"
	fileSize := int64(25) // Should result in 3 parts: 10, 10, 5
	userID := int64(123)

	// 1. Test Initiate
	repo.On("CreateSession", ctx, mock.AnythingOfType("*biz.UploadSession")).Return(nil)
	session, err := uc.InitiateMultipartUpload(
		ctx,
		filename,
		fileSize,
		"text/plain",
		"Title",
		"Desc",
		nil,
		nil,
		&userID,
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, session.TotalParts)
	uploadID := session.UploadID

	// 2. Test Upload Parts
	repo.On("GetSession", ctx, uploadID).Return(session, nil)
	repo.On("UpdateSession", ctx, session).Return(nil)

	etag1, err := uc.UploadPart(ctx, uploadID, 1, []byte("0123456789"))
	assert.NoError(t, err)
	assert.NotEmpty(t, etag1)

	_, err = uc.UploadPart(ctx, uploadID, 2, []byte("abcdefghij"))
	assert.NoError(t, err)

	_, err = uc.UploadPart(ctx, uploadID, 3, []byte("final"))
	assert.NoError(t, err)

	// 3. Test Complete
	// Using mock.Anything because Media is an alias to types.Media
	// and mock.AnythingOfType may have issues with type aliases in some environments.
	mediaRepo.On("Create", ctx, mock.Anything).Return(&Media{Id: 1, Title: "Title"}, nil)

	media, err := uc.CompleteMultipartUpload(ctx, uploadID, "hash")
	assert.NoError(t, err)
	assert.NotNil(t, media)

	// Verify final file exists and content is correct
	finalFilePath := filepath.Join(tempDir, "uploads", uploadID+".txt")
	content, _ := os.ReadFile(finalFilePath)
	assert.Equal(t, "0123456789abcdefghijfinal", string(content))
}

// Simple storage implementation for test to avoid dependency on data package
type testStorage struct {
	baseDir string
}

func (s *testStorage) StorePart(
	ctx context.Context,
	uploadID string,
	partNumber int,
	data []byte,
) (string, error) {
	path := filepath.Join(s.baseDir, uploadID, fmt.Sprintf("part_%d", partNumber))
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
	return "etag", nil
}

func (s *testStorage) MergeParts(
	ctx context.Context,
	uploadID string,
	totalParts int,
	finalPath string,
) error {
	dest := filepath.Join(s.baseDir, finalPath)
	_ = os.MkdirAll(filepath.Dir(dest), 0o755)
	f, _ := os.Create(dest)
	defer f.Close()
	for i := 1; i <= totalParts; i++ {
		p, _ := os.ReadFile(filepath.Join(s.baseDir, uploadID, fmt.Sprintf("part_%d", i)))
		f.Write(p)
	}
	return nil
}

func (s *testStorage) DeleteParts(ctx context.Context, uploadID string) error {
	return os.RemoveAll(filepath.Join(s.baseDir, uploadID))
}
