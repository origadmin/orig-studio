/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"

	"origadmin/application/origcms/internal/svc-media/dto"
)

// MockTranscodeWorker 模拟转码 worker
type MockTranscodeWorker struct {
	tasks []TranscodeJob
}

func NewMockTranscodeWorker() *MockTranscodeWorker {
	return &MockTranscodeWorker{
		tasks: make([]TranscodeJob, 0),
	}
}

func (w *MockTranscodeWorker) Submit(ctx context.Context, job TranscodeJob) error {
	w.tasks = append(w.tasks, job)
	// 模拟转码成功
	return nil
}

func (w *MockTranscodeWorker) Shutdown(ctx context.Context) error {
	// 模拟关闭
	return nil
}

func (w *MockTranscodeWorker) Status() WorkerPoolStatus {
	// 模拟状态
	return WorkerPoolStatus{
		MaxWorkers:    4,
		ActiveWorkers: 0,
		PendingJobs:   int32(len(w.tasks)),
	}
}

// MockMessagePublisher 模拟消息发布器
type MockMessagePublisher struct {
	messages []*message.Message
}

func NewMockMessagePublisher() *MockMessagePublisher {
	return &MockMessagePublisher{
		messages: make([]*message.Message, 0),
	}
}

func (p *MockMessagePublisher) Publish(topic string, messages ...*message.Message) error {
	p.messages = append(p.messages, messages...)
	return nil
}

func (p *MockMessagePublisher) Close() error {
	// 模拟关闭
	return nil
}

// TestTranscodeHandler_Handle 测试转码请求处理
func TestTranscodeHandler_Handle(t *testing.T) {
	// 创建模拟依赖
	mediaRepo := NewMockReviewRepo()
	profileRepo := NewMockEncodeProfileRepo()
	encodingRepo := NewMockEncodingTaskRepo()
	worker := NewMockTranscodeWorker()
	publisher := NewMockMessagePublisher()
	logger := log.NewStdLogger(os.Stdout)
	
	// 创建媒体用例
	mediaUC := NewMediaUseCase(mediaRepo, nil, nil, nil, nil, logger)
	
	// 创建转码处理器
	handler := NewTranscodeHandler(
		mediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		worker,
		publisher,
		logger,
		"./test-data",
		1*time.Second, // 短超时
	)
	
	// 创建测试媒体
	media := &Media{
		Id:             "media-123",
		Title:          "Test Video",
		Type:           "video",
		Url:            "uploads/test.mp4",
		UserId:         "user-123",
		State:          "active",
		EncodingStatus: "pending",
		MimeType:       "video/mp4",
		Size:           1024 * 1024,
		Extension:      "mp4",
		Privacy:        1,
	}
	
	// 保存媒体
	_, err := mediaRepo.Create(context.Background(), media)
	assert.NoError(t, err)
	
	// 创建测试编码配置
	profile1 := &dto.EncodeProfile{
		Id:         1,
		Name:       "720p",
		Resolution: "720",
		Extension:  "mp4",
		IsActive:   true,
	}
	_, err = profileRepo.Create(context.Background(), profile1)
	assert.NoError(t, err)
	
	// 创建转码请求
	req := MediaEncodeRequest{
		MediaID:     "media-123",
		MediaPath:   "uploads/test.mp4",
		ContentType: "video/mp4",
	}
	
	// 序列化请求
	payload, err := json.Marshal(req)
	assert.NoError(t, err)
	
	// 创建消息
	msg := message.NewMessage("msg-1", payload)
	
	// 处理转码请求
	err = handler.Handle(msg)
	// 预期会失败，因为缺少 ffmpeg 和输出文件
	assert.Error(t, err)
	
	// 验证媒体状态更新
	updatedMedia, err := mediaRepo.Get(context.Background(), "media-123")
	assert.NoError(t, err)
	assert.Equal(t, "processing", updatedMedia.EncodingStatus)
}

// TestTranscodeHandler_ProcessMedia 测试转码流程核心逻辑
func TestTranscodeHandler_ProcessMedia(t *testing.T) {
	// 创建模拟依赖
	mediaRepo := NewMockReviewRepo()
	profileRepo := NewMockEncodeProfileRepo()
	encodingRepo := NewMockEncodingTaskRepo()
	worker := NewMockTranscodeWorker()
	publisher := NewMockMessagePublisher()
	logger := log.NewStdLogger(os.Stdout)
	
	// 创建媒体用例
	mediaUC := NewMediaUseCase(mediaRepo, nil, nil, nil, nil, logger)
	
	// 创建转码处理器
	handler := NewTranscodeHandler(
		mediaUC,
		profileRepo,
		encodingRepo,
		mediaRepo,
		worker,
		publisher,
		logger,
		"./test-data",
		1*time.Second, // 短超时
	)
	
	// 创建测试媒体
	media := &Media{
		Id:             "media-123",
		Title:          "Test Video",
		Type:           "video",
		Url:            "uploads/test.mp4",
		UserId:         "user-123",
		State:          "active",
		EncodingStatus: "pending",
		MimeType:       "video/mp4",
		Size:           1024 * 1024,
		Extension:      "mp4",
		Privacy:        1,
	}
	
	// 保存媒体
	_, err := mediaRepo.Create(context.Background(), media)
	assert.NoError(t, err)
	
	// 创建测试编码配置
	profile1 := &dto.EncodeProfile{
		Id:         1,
		Name:       "720p",
		Resolution: "720",
		Extension:  "mp4",
		IsActive:   true,
	}
	_, err = profileRepo.Create(context.Background(), profile1)
	assert.NoError(t, err)
	
	// 创建转码请求
	req := &MediaEncodeRequest{
		MediaID:     "media-123",
		MediaPath:   "uploads/test.mp4",
		ContentType: "video/mp4",
	}
	
	// 处理转码请求
	err = handler.processMedia(context.Background(), req)
	// 预期会失败，因为缺少 ffmpeg 和输出文件
	assert.Error(t, err)
	
	// 验证任务创建
	tasks, err := encodingRepo.ListByMedia(context.Background(), "media-123")
	assert.NoError(t, err)
	assert.Greater(t, len(tasks), 0)
	
	// 验证媒体状态更新
	updatedMedia, err := mediaRepo.Get(context.Background(), "media-123")
	assert.NoError(t, err)
	assert.Equal(t, "processing", updatedMedia.EncodingStatus)
}
