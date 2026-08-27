/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

type mockNotificationRepo struct {
	usersByRole       map[string][]string
	usersByGroup      map[string][]string
	allActiveUsers    []string
	createErr         error
	batchCreateErr    error
	getErr            error
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *Notification) (*Notification, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	n.ID = 1
	return n, nil
}

func (m *mockNotificationRepo) Get(ctx context.Context, id int) (*Notification, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return &Notification{ID: id}, nil
}

func (m *mockNotificationRepo) Update(ctx context.Context, n *Notification) (*Notification, error) {
	return n, nil
}

func (m *mockNotificationRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *mockNotificationRepo) DeleteAllByUser(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockNotificationRepo) DeleteReadByUser(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockNotificationRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Notification, int, error) {
	return nil, 0, nil
}

func (m *mockNotificationRepo) ListAll(ctx context.Context, page, pageSize int) ([]*Notification, int, error) {
	return nil, 0, nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, id int) error {
	return nil
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	return nil
}

func (m *mockNotificationRepo) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *mockNotificationRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
	return m.allActiveUsers, nil
}

func (m *mockNotificationRepo) GetAllActiveUserIDs(ctx context.Context) ([]string, error) {
	return m.allActiveUsers, nil
}

func (m *mockNotificationRepo) GetUserIDsByRole(ctx context.Context, roles []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)
	for _, role := range roles {
		for _, uid := range m.usersByRole[role] {
			if !seen[uid] {
				seen[uid] = true
				result = append(result, uid)
			}
		}
	}
	return result, nil
}

func (m *mockNotificationRepo) GetUserIDsByGroup(ctx context.Context, groupIDs []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)
	for _, gid := range groupIDs {
		for _, uid := range m.usersByGroup[gid] {
			if !seen[uid] {
				seen[uid] = true
				result = append(result, uid)
			}
		}
	}
	return result, nil
}

func (m *mockNotificationRepo) BatchCreate(ctx context.Context, notifications []*Notification) ([]*Notification, error) {
	if m.batchCreateErr != nil {
		return nil, m.batchCreateErr
	}
	for i, n := range notifications {
		n.ID = i + 1
	}
	return notifications, nil
}

func (m *mockNotificationRepo) ListPermissionGroups(ctx context.Context) ([]*PermissionGroupInfo, error) {
	return nil, nil
}

func newNotificationUseCase(repo NotificationRepo) *NotificationUseCase {
	return &NotificationUseCase{repo: repo}
}

func sortStrings(s []string) []string {
	sorted := make([]string, len(s))
	copy(sorted, s)
	sort.Strings(sorted)
	return sorted
}

func TestResolveTargetUsers_ExplicitIDs(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole:    map[string][]string{"admin": {"u1", "u2"}},
		usersByGroup:   map[string][]string{"g1": {"u3", "u4"}},
		allActiveUsers: []string{"u1", "u2", "u3", "u4", "u5"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	result, err := uc.ResolveTargetUsers(ctx, []string{"u99", "u100"}, []string{"admin"}, []string{"g1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"u99", "u100"}
	if !reflect.DeepEqual(sortStrings(result), sortStrings(expected)) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveTargetUsers_ByRole(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole: map[string][]string{
			"admin":  {"u1", "u2"},
			"user":   {"u3", "u4"},
			"editor": {"u5"},
		},
		usersByGroup:   map[string][]string{},
		allActiveUsers: []string{"u1", "u2", "u3", "u4", "u5"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	result, err := uc.ResolveTargetUsers(ctx, nil, []string{"admin", "user"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"u1", "u2", "u3", "u4"}
	if !reflect.DeepEqual(sortStrings(result), sortStrings(expected)) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveTargetUsers_ByGroup(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole: map[string][]string{},
		usersByGroup: map[string][]string{
			"g1": {"u1", "u2"},
			"g2": {"u2", "u3"},
			"g3": {"u4"},
		},
		allActiveUsers: []string{"u1", "u2", "u3", "u4", "u5"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	result, err := uc.ResolveTargetUsers(ctx, nil, nil, []string{"g1", "g2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"u1", "u2", "u3"}
	if !reflect.DeepEqual(sortStrings(result), sortStrings(expected)) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveTargetUsers_AllActive(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole:    map[string][]string{},
		usersByGroup:   map[string][]string{},
		allActiveUsers: []string{"u1", "u2", "u3"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	result, err := uc.ResolveTargetUsers(ctx, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"u1", "u2", "u3"}
	if !reflect.DeepEqual(sortStrings(result), sortStrings(expected)) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestResolveTargetUsers_PriorityOrder(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole:    map[string][]string{"admin": {"u1"}},
		usersByGroup:   map[string][]string{"g1": {"u2"}},
		allActiveUsers: []string{"u3"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	result, err := uc.ResolveTargetUsers(ctx, []string{"u99"}, []string{"admin"}, []string{"g1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "u99" {
		t.Errorf("explicit user_ids should take priority, got %v", result)
	}

	result, err = uc.ResolveTargetUsers(ctx, nil, []string{"admin"}, []string{"g1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "u1" {
		t.Errorf("roles should take priority over groups, got %v", result)
	}
}

func TestResolveTargetUsers_EmptyRolesReturnsEmpty(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole:    map[string][]string{"admin": {}},
		usersByGroup:   map[string][]string{},
		allActiveUsers: []string{"u1", "u2"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	result, err := uc.ResolveTargetUsers(ctx, nil, []string{"admin"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for role with no users, got %v", result)
	}
}

func TestBroadcastByRole(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByRole: map[string][]string{
			"admin": {"u1", "u2"},
		},
		allActiveUsers: []string{"u1", "u2", "u3"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	n := &Notification{Action: "system", Title: "Hello", Body: "Test"}
	results, err := uc.BroadcastByRole(ctx, []string{"admin"}, n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(results))
	}
}

func TestBroadcastByGroup(t *testing.T) {
	repo := &mockNotificationRepo{
		usersByGroup: map[string][]string{
			"g1": {"u1", "u2", "u3"},
		},
		allActiveUsers: []string{},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	n := &Notification{Action: "system", Title: "Hello", Body: "Test"}
	results, err := uc.BroadcastByGroup(ctx, []string{"g1"}, n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 notifications, got %d", len(results))
	}
}

func TestBroadcastToAll(t *testing.T) {
	repo := &mockNotificationRepo{
		allActiveUsers: []string{"u1", "u2", "u3", "u4"},
	}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	n := &Notification{Action: "system", Title: "Hello", Body: "Test"}
	results, err := uc.BroadcastToAll(ctx, n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 notifications, got %d", len(results))
	}
}

func TestBatchCreateNotifications_EmptyList(t *testing.T) {
	repo := &mockNotificationRepo{}
	uc := newNotificationUseCase(repo)
	ctx := context.Background()

	n := &Notification{Action: "system", Title: "Hello", Body: "Test"}
	results, err := uc.BatchCreateNotifications(ctx, nil, n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 notifications for empty user list, got %d", len(results))
	}
}
