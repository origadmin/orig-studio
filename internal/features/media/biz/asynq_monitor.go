package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/origadmin/runtime/log"
)

type AsynqMonitor struct {
	inspector *asynq.Inspector
	logger    *log.Helper
	addr      string
	server    *http.Server
}

func NewAsynqMonitor(config AsynqWorkerConfig, addr string, logger *log.Helper) *AsynqMonitor {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	if addr == "" {
		addr = ":8090"
	}

	return &AsynqMonitor{
		inspector: inspector,
		logger:    logger,
		addr:      addr,
	}
}

func (m *AsynqMonitor) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", m.handleHealth)
	mux.HandleFunc("/queue/stats", m.handleQueueStats)
	mux.HandleFunc("/queue/tasks", m.handleTasks)

	m.server = &http.Server{
		Addr:         m.addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		m.logger.Infof("asynq monitor listening on %s", m.addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Errorf("asynq monitor error: %v", err)
		}
	}()

	return nil
}

func (m *AsynqMonitor) Shutdown(ctx context.Context) error {
	if m.server != nil {
		return m.server.Shutdown(ctx)
	}
	return nil
}

func (m *AsynqMonitor) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func (m *AsynqMonitor) handleQueueStats(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queues, err := m.inspector.Queues()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	type queueInfo struct {
		Name     string `json:"name"`
		Paused   bool   `json:"paused"`
		Size     int    `json:"size"`
		Pending  int    `json:"pending"`
		Active   int    `json:"active"`
		Scheduled int   `json:"scheduled"`
		Retry    int    `json:"retry"`
		Archived int    `json:"archived"`
		Completed int   `json:"completed"`
	}

	result := make([]queueInfo, 0, len(queues))
	for _, q := range queues {
		info, err := m.inspector.GetQueueInfo(q)
		if err != nil {
			continue
		}
		result = append(result, queueInfo{
			Name:      q,
			Paused:    info.Paused,
			Size:      info.Size,
			Pending:   info.Pending,
			Active:    info.Active,
			Scheduled: info.Scheduled,
			Retry:     info.Retry,
			Archived:  info.Archived,
			Completed: info.Completed,
		})
	}

	data, _ := json.Marshal(result)
	w.Write(data)
}

func (m *AsynqMonitor) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queue := r.URL.Query().Get("queue")
	if queue == "" {
		queue = "default"
	}

	tasks, err := m.inspector.ListPendingTasks(queue, asynq.PageSize(50))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
		return
	}

	type taskInfo struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Queue   string `json:"queue"`
		Retried int    `json:"retried"`
		MaxRetry int   `json:"max_retry"`
	}

	result := make([]taskInfo, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, taskInfo{
			ID:       t.ID,
			Type:     t.Type,
			Queue:    t.Queue,
			Retried:  t.Retried,
			MaxRetry: t.MaxRetry,
		})
	}

	data, _ := json.Marshal(result)
	w.Write(data)
}
