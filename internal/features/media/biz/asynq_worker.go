package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hibiken/asynq"
	"github.com/origadmin/runtime/log"

	"origadmin/application/origstudio/internal/features/media/dto"
)

const (
	TaskTypeTranscode = "transcode:video"
)

type AsynqWorkerConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
	Queues        map[string]int
	RetryMax      int
	Timeout       time.Duration
}

func DefaultAsynqWorkerConfig() AsynqWorkerConfig {
	return AsynqWorkerConfig{
		RedisAddr:   "localhost:6379",
		Concurrency: 4,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		RetryMax: 3,
		Timeout:  2 * time.Hour,
	}
}

type AsynqWorker struct {
	client        *asynq.Client
	mux           *asynq.ServeMux
	server        *asynq.Server
	config        AsynqWorkerConfig
	pendingCount  atomic.Int32
	activeCount   atomic.Int32
	maxWorkers    int32
	logger        *log.Helper
	mediaUC       MediaEventPublisher
	encodingRepo  dto.EncodingTaskRepo
}

func NewAsynqWorker(config AsynqWorkerConfig, encodingRepo dto.EncodingTaskRepo, mediaUC MediaEventPublisher, logger *log.Helper) *AsynqWorker {
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	w := &AsynqWorker{
		client:       asynqClient,
		config:       config,
		maxWorkers:   int32(config.Concurrency),
		logger:       logger,
		mediaUC:      mediaUC,
		encodingRepo: encodingRepo,
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeTranscode, w.handleTranscodeTask)

	w.mux = mux

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		},
		asynq.Config{
			Concurrency: config.Concurrency,
			Queues:      config.Queues,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Errorf("asynq task error: type=%s payload=%s err=%v", task.Type(), string(task.Payload()), err)
			}),
		},
	)
	w.server = srv

	return w
}

type asynqTranscodePayload struct {
	MediaID   string           `json:"media_id"`
	TaskID    string           `json:"task_id"`
	Profile   *dto.EncodeProfile `json:"profile"`
	InputPath string           `json:"input_path"`
	OutputDir string           `json:"output_dir"`
}

func (w *AsynqWorker) Submit(ctx context.Context, job TranscodeJob) error {
	payload := asynqTranscodePayload{
		MediaID:   job.MediaID,
		TaskID:    job.TaskID,
		Profile:   job.Profile,
		InputPath: job.InputPath,
		OutputDir: job.OutputDir,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal transcode payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeTranscode, data,
		asynq.Queue("default"),
		asynq.MaxRetry(w.config.RetryMax),
		asynq.Timeout(w.config.Timeout),
		asynq.Retention(24*time.Hour),
	)

	info, err := w.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue transcode task: %w", err)
	}

	w.pendingCount.Add(1)

	w.logger.Infof("enqueued transcode task: media=%s profile=%s queue_id=%s",
		job.MediaID, job.Profile.Name, info.ID)

	return nil
}

func (w *AsynqWorker) handleTranscodeTask(ctx context.Context, t *asynq.Task) error {
	w.pendingCount.Add(-1)
	w.activeCount.Add(1)
	defer w.activeCount.Add(-1)

	var payload asynqTranscodePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job := TranscodeJob{
		MediaID:      payload.MediaID,
		TaskID:       payload.TaskID,
		Profile:      payload.Profile,
		InputPath:    payload.InputPath,
		OutputDir:    payload.OutputDir,
		EncodingRepo: w.encodingRepo,
		MediaUC:      w.mediaUC,
		Logger:       w.logger,
	}

	logger := w.logger
	logger.Infof("processing transcode task: media=%s profile=%s", job.MediaID, job.Profile.Name)

	if err := executeTranscodeJob(ctx, job, logger); err != nil {
		return fmt.Errorf("transcode job failed: %w", err)
	}

	return nil
}

func (w *AsynqWorker) Start() error {
	w.logger.Infof("starting asynq worker server: concurrency=%d redis=%s", w.config.Concurrency, w.config.RedisAddr)
	go func() {
		if err := w.server.Run(w.mux); err != nil {
			w.logger.Errorf("asynq server error: %v", err)
		}
	}()
	return nil
}

func (w *AsynqWorker) Status() WorkerPoolStatus {
	return WorkerPoolStatus{
		MaxWorkers:    w.maxWorkers,
		ActiveWorkers: w.activeCount.Load(),
		PendingJobs:   w.pendingCount.Load(),
	}
}

func (w *AsynqWorker) Shutdown(ctx context.Context) error {
	w.server.Shutdown()
	w.server.Stop()
	return w.client.Close()
}
