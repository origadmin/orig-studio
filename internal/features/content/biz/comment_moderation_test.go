package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockCommentModerationRepo struct {
	getFunc              func(ctx context.Context, id string) (*CommentModerationItem, error)
	updateStatusFunc     func(ctx context.Context, commentID string, status string, moderatedBy string) error
	resetReportCountFunc func(ctx context.Context, commentID string) error
}

func (m *mockCommentModerationRepo) Get(ctx context.Context, id string) (*CommentModerationItem, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockCommentModerationRepo) UpdateStatus(ctx context.Context, commentID string, status string, moderatedBy string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, commentID, status, moderatedBy)
	}
	return nil
}

func (m *mockCommentModerationRepo) BatchUpdateStatus(ctx context.Context, commentIDs []string, status string, moderatedBy string) (int, error) {
	return 0, nil
}

func (m *mockCommentModerationRepo) IncrementReportCount(ctx context.Context, commentID string) (int, error) {
	return 0, nil
}

func (m *mockCommentModerationRepo) ResetReportCount(ctx context.Context, commentID string) error {
	if m.resetReportCountFunc != nil {
		return m.resetReportCountFunc(ctx, commentID)
	}
	return nil
}

func (m *mockCommentModerationRepo) ListByMedia(ctx context.Context, mediaID string, status string, page, pageSize int) ([]*CommentModerationItem, int, error) {
	return nil, 0, nil
}

func (m *mockCommentModerationRepo) ListAdminComments(ctx context.Context, mediaID string, status string, reportStatus string, tree bool, page, pageSize int) ([]*CommentModerationItem, int, error) {
	return nil, 0, nil
}

func (m *mockCommentModerationRepo) ListPending(ctx context.Context, mediaID string, page, pageSize int) ([]*CommentModerationItem, int, error) {
	return nil, 0, nil
}

func (m *mockCommentModerationRepo) CountByStatus(ctx context.Context, mediaID string) (*CommentStatusCounts, error) {
	return nil, nil
}

func (m *mockCommentModerationRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockCommentReportRepo struct {
	updateStatusByCommentFunc func(ctx context.Context, commentID string, fromStatus string, toStatus string) (int, error)
	countPendingByCommentFunc func(ctx context.Context, commentID string) (int, error)
}

func (m *mockCommentReportRepo) Create(ctx context.Context, commentID string, reporterID string, reason string, description string) (*CommentReportItem, error) {
	return nil, nil
}

func (m *mockCommentReportRepo) Exists(ctx context.Context, reporterID, commentID string) (bool, error) {
	return false, nil
}

func (m *mockCommentReportRepo) ListByComment(ctx context.Context, commentID string) ([]*CommentReportItem, error) {
	return nil, nil
}

func (m *mockCommentReportRepo) UpdateStatusByComment(ctx context.Context, commentID string, fromStatus string, toStatus string) (int, error) {
	if m.updateStatusByCommentFunc != nil {
		return m.updateStatusByCommentFunc(ctx, commentID, fromStatus, toStatus)
	}
	return 0, nil
}

func (m *mockCommentReportRepo) CountPendingByComment(ctx context.Context, commentID string) (int, error) {
	if m.countPendingByCommentFunc != nil {
		return m.countPendingByCommentFunc(ctx, commentID)
	}
	return 0, nil
}

type mockConfigProvider struct{}

func (m *mockConfigProvider) Get(ctx context.Context, key string) string            { return "" }
func (m *mockConfigProvider) GetBool(ctx context.Context, key string) bool           { return true }
func (m *mockConfigProvider) GetInt(ctx context.Context, key string) int             { return 3 }
func (m *mockConfigProvider) GetAll(ctx context.Context) map[string]string           { return nil }

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newTestUseCase(commentRepo CommentModerationRepo, reportRepo CommentReportRepo) *CommentModerationUseCase {
	return NewCommentModerationUseCase(commentRepo, reportRepo, &mockConfigProvider{}, log.DefaultLogger)
}

// ---------------------------------------------------------------------------
// Tests: BlockComment
// ---------------------------------------------------------------------------

func TestBlockComment_Success(t *testing.T) {
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return &CommentModerationItem{ID: id, Status: "APPROVED"}, nil
		},
		updateStatusFunc: func(ctx context.Context, commentID string, status string, moderatedBy string) error {
			if status != "BLOCKED" {
				t.Errorf("expected status BLOCKED, got %s", status)
			}
			return nil
		},
	}
	reportRepo := &mockCommentReportRepo{
		updateStatusByCommentFunc: func(ctx context.Context, commentID string, fromStatus string, toStatus string) (int, error) {
			if fromStatus != "PENDING" || toStatus != "REVIEWED" {
				t.Errorf("expected PENDING->REVIEWED, got %s->%s", fromStatus, toStatus)
			}
			return 2, nil
		},
	}

	uc := newTestUseCase(commentRepo, reportRepo)
	result, err := uc.BlockComment(context.Background(), "comment-1", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result // result may be nil if Get after update fails, which is acceptable
}

func TestBlockComment_InvalidTransition(t *testing.T) {
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return &CommentModerationItem{ID: id, Status: "PENDING"}, nil
		},
	}
	reportRepo := &mockCommentReportRepo{}

	uc := newTestUseCase(commentRepo, reportRepo)
	_, err := uc.BlockComment(context.Background(), "comment-1", "admin-1")
	if err == nil {
		t.Fatal("expected error for invalid transition PENDING->BLOCKED, got nil")
	}
}

func TestBlockComment_BlockedToBlocked(t *testing.T) {
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return &CommentModerationItem{ID: id, Status: "BLOCKED"}, nil
		},
	}
	reportRepo := &mockCommentReportRepo{}

	uc := newTestUseCase(commentRepo, reportRepo)
	_, err := uc.BlockComment(context.Background(), "comment-1", "admin-1")
	if err == nil {
		t.Fatal("expected error for BLOCKED->BLOCKED transition, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: UnblockComment
// ---------------------------------------------------------------------------

func TestUnblockComment_Success(t *testing.T) {
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return &CommentModerationItem{ID: id, Status: "BLOCKED"}, nil
		},
		updateStatusFunc: func(ctx context.Context, commentID string, status string, moderatedBy string) error {
			if status != "APPROVED" {
				t.Errorf("expected status APPROVED, got %s", status)
			}
			return nil
		},
		resetReportCountFunc: func(ctx context.Context, commentID string) error {
			return nil
		},
	}
	reportRepo := &mockCommentReportRepo{}

	uc := newTestUseCase(commentRepo, reportRepo)
	_, err := uc.UnblockComment(context.Background(), "comment-1", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnblockComment_InvalidTransition(t *testing.T) {
	// APPROVED->APPROVED is not a valid transition, so UnblockComment on an already APPROVED comment should fail
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return &CommentModerationItem{ID: id, Status: "APPROVED"}, nil
		},
	}
	reportRepo := &mockCommentReportRepo{}

	uc := newTestUseCase(commentRepo, reportRepo)
	_, err := uc.UnblockComment(context.Background(), "comment-1", "admin-1")
	if err == nil {
		t.Fatal("expected error for invalid transition APPROVED->APPROVED via UnblockComment, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: DismissReports
// ---------------------------------------------------------------------------

func TestDismissReports_Success(t *testing.T) {
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return &CommentModerationItem{ID: id, Status: "APPROVED", ReportCount: 3}, nil
		},
		resetReportCountFunc: func(ctx context.Context, commentID string) error {
			return nil
		},
	}
	reportRepo := &mockCommentReportRepo{
		updateStatusByCommentFunc: func(ctx context.Context, commentID string, fromStatus string, toStatus string) (int, error) {
			if fromStatus != "PENDING" || toStatus != "DISMISSED" {
				t.Errorf("expected PENDING->DISMISSED, got %s->%s", fromStatus, toStatus)
			}
			return 3, nil
		},
	}

	uc := newTestUseCase(commentRepo, reportRepo)
	result, err := uc.DismissReports(context.Background(), "comment-1", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DismissedCount != 3 {
		t.Errorf("expected dismissed count 3, got %d", result.DismissedCount)
	}
	if result.ReportCount != 0 {
		t.Errorf("expected report count 0, got %d", result.ReportCount)
	}
}

func TestDismissReports_CommentNotFound(t *testing.T) {
	commentRepo := &mockCommentModerationRepo{
		getFunc: func(ctx context.Context, id string) (*CommentModerationItem, error) {
			return nil, errors.New("comment not found")
		},
	}
	reportRepo := &mockCommentReportRepo{}

	uc := newTestUseCase(commentRepo, reportRepo)
	_, err := uc.DismissReports(context.Background(), "nonexistent", "admin-1")
	if err == nil {
		t.Fatal("expected error for nonexistent comment, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests: Status Transitions
// ---------------------------------------------------------------------------

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		isValid bool
	}{
		// PENDING transitions
		{"PENDING->APPROVED", "PENDING", "APPROVED", true},
		{"PENDING->REJECTED", "PENDING", "REJECTED", true},
		{"PENDING->BLOCKED", "PENDING", "BLOCKED", false},
		{"PENDING->PENDING", "PENDING", "PENDING", false},

		// APPROVED transitions
		{"APPROVED->REJECTED", "APPROVED", "REJECTED", true},
		{"APPROVED->BLOCKED", "APPROVED", "BLOCKED", true},
		{"APPROVED->PENDING", "APPROVED", "PENDING", true},
		{"APPROVED->APPROVED", "APPROVED", "APPROVED", false},

		// REJECTED transitions
		{"REJECTED->APPROVED", "REJECTED", "APPROVED", true},
		{"REJECTED->BLOCKED", "REJECTED", "BLOCKED", false},
		{"REJECTED->PENDING", "REJECTED", "PENDING", false},

		// BLOCKED transitions
		{"BLOCKED->APPROVED", "BLOCKED", "APPROVED", true},
		{"BLOCKED->REJECTED", "BLOCKED", "REJECTED", false},
		{"BLOCKED->PENDING", "BLOCKED", "PENDING", false},
		{"BLOCKED->BLOCKED", "BLOCKED", "BLOCKED", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, ok := validStatusTransitions[tt.from]
			isValid := ok && allowed[tt.to]
			if isValid != tt.isValid {
				t.Errorf("transition %s->%s: expected valid=%v, got valid=%v", tt.from, tt.to, tt.isValid, isValid)
			}
		})
	}
}
