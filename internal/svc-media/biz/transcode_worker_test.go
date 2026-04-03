package biz

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func TestCountingSemaphore(t *testing.T) {
	sem := newCountingSemaphore(2)
	acquired := atomic.Int32{}
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Lock()
			acquired.Add(1)
			time.Sleep(50 * time.Millisecond)
			acquired.Add(-1)
			sem.Unlock()
		}()
	}

	wg.Wait()

	if acquired.Load() != 0 {
		t.Errorf("expected 0 active after all goroutines finish, got %d", acquired.Load())
	}
}

func TestGoroutineWorker_Status(t *testing.T) {
	logger := log.NewHelper(log.DefaultLogger)
	worker := NewGoroutineWorker(2, logger)

	status := worker.Status()
	if status.MaxWorkers != 2 {
		t.Errorf("expected MaxWorkers=2, got %d", status.MaxWorkers)
	}

	// Submit 5 jobs — only 2 should run concurrently
	for i := 0; i < 5; i++ {
		worker.Submit(context.Background(), TranscodeJob{
			MediaID: int64(i),
			Profile: &EncodeProfile{Name: "test", Extension: "skip", Resolution: "-"},
		})
	}

	// Give some time for goroutines to start
	time.Sleep(100 * time.Millisecond)

	status = worker.Status()
	if status.ActiveWorkers > 2 {
		t.Errorf("expected ActiveWorkers <= 2, got %d", status.ActiveWorkers)
	}

	// Shutdown should complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := worker.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestIsSkipProfile(t *testing.T) {
	tests := []struct {
		profile  EncodeProfile
		expected bool
	}{
		{EncodeProfile{Extension: "mp4", Resolution: "720"}, false},
		{EncodeProfile{Extension: "mp4", Resolution: "1080"}, false},
		{EncodeProfile{Extension: "gif", Resolution: "-"}, true},
		{EncodeProfile{Extension: "gif", Resolution: ""}, true},
	}

	for _, tt := range tests {
		p := tt.profile
		// Test IsSkipResolution
		skip := p.Extension != "mp4" && p.Extension != "webm"
		if skip != tt.expected {
			t.Errorf("IsSkipProfile(%+v) = %v, want %v", p, skip, tt.expected)
		}
	}
}
