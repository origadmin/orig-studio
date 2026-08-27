package biz

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/auth/permission/dto"
)

var RoleDefaultPermissions = map[string][]string{
	"admin":  {"media:read", "media:write", "media:delete", "media:publish", "media:moderate", "comment:read", "comment:write", "comment:moderate", "user:manage", "system:config", "payment:read", "payment:write", "payment:delete", "permission:read", "permission:write", "permission:delete", "permission:manage"},
	"editor": {"media:read", "media:write", "comment:read", "comment:write", "comment:moderate"},
	"user":   {"media:read", "comment:read", "comment:write"},
}

var AllPermissions = []struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	ResourceType string `json:"resource_type"`
	Action       string `json:"action"`
}{
	{"media:read", "View Media", "media", "read"},
	{"media:write", "Edit Media", "media", "write"},
	{"media:delete", "Delete Media", "media", "delete"},
	{"media:publish", "Publish Media", "media", "publish"},
	{"media:moderate", "Moderate Media", "media", "moderate"},
	{"comment:read", "View Comments", "comment", "read"},
	{"comment:write", "Post Comments", "comment", "write"},
	{"comment:moderate", "Moderate Comments", "comment", "moderate"},
	{"user:manage", "Manage Users", "user", "manage"},
	{"system:config", "System Config", "system", "config"},
	{"payment:read", "View Payments", "payment", "read"},
	{"payment:write", "Manage Payments", "payment", "write"},
	{"payment:delete", "Delete Payments", "payment", "delete"},
	{"permission:read", "View Permissions", "permission", "read"},
	{"permission:write", "Manage Permissions", "permission", "write"},
	{"permission:delete", "Delete Permissions", "permission", "delete"},
	{"permission:manage", "Manage Members", "permission", "manage"},
}

func IsValidPermission(perm string) bool {
	for _, p := range AllPermissions {
		if p.Key == perm {
			return true
		}
	}
	return false
}

type GroupRepo interface {
	Create(ctx context.Context, name, description string, permissions, categoryScope []string, createdBy string) (*dto.GroupItem, error)
	Get(ctx context.Context, id string) (*dto.GroupItem, error)
	Update(ctx context.Context, id, name, description string, permissions, categoryScope []string) (*dto.GroupItem, error)
	Delete(ctx context.Context, id string) error
	Toggle(ctx context.Context, id string, isActive bool) error
	List(ctx context.Context, isActive *bool, page, pageSize int) ([]*dto.GroupItem, int, error)
	GetMemberIDs(ctx context.Context, groupID string) ([]string, error)
}

type MemberRepo interface {
	AddMembers(ctx context.Context, groupID string, userIDs []string) (added int, skipped int, err error)
	RemoveMember(ctx context.Context, groupID, userID string) error
	ListByGroup(ctx context.Context, groupID string, page, pageSize int) ([]*dto.MemberItem, int, error)
	ListByUser(ctx context.Context, userID string) ([]*dto.MemberItem, error)
}

type UserPermRepo interface {
	GetUserRole(ctx context.Context, userID string) (string, error)
}

type Checker interface {
	CheckPermission(ctx context.Context, userID string, permission string, categoryID string) (bool, error)
	InvalidateUserCache(ctx context.Context, userID string) error
	InvalidateGroupCache(ctx context.Context, groupID string) error
}

type permCacheEntry struct {
	permissions map[string]*dto.Source
	expiresAt   time.Time
}

type UseCase struct {
	groupRepo  GroupRepo
	memberRepo MemberRepo
	userRepo   UserPermRepo
	cache      sync.Map
	cacheTTL   time.Duration
	notifyCh   chan string
	logger     *log.Helper
}

func NewUseCase(
	groupRepo GroupRepo,
	memberRepo MemberRepo,
	userRepo UserPermRepo,
	logger log.Logger,
) *UseCase {
	uc := &UseCase{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		cacheTTL:   5 * time.Minute,
		notifyCh:   make(chan string, 100),
		logger:     log.NewHelper(log.With(logger, "module", "enterprise/permission.biz")),
	}
	go uc.processInvalidation()
	return uc
}

func (uc *UseCase) processInvalidation() {
	for userID := range uc.notifyCh {
		uc.cache.Delete(userID)
	}
}

func (uc *UseCase) CheckPermission(ctx context.Context, userID string, permission string, categoryID string) (bool, error) {
	perms, err := uc.getUserPermissionSet(ctx, userID)
	if err != nil {
		return false, err
	}
	src, ok := perms[permission]
	if !ok {
		return false, nil
	}
	if len(src.Scope) == 0 {
		return true, nil
	}
	if categoryID == "" {
		return false, nil
	}
	for _, scope := range src.Scope {
		if scope == categoryID {
			return true, nil
		}
	}
	return false, nil
}

func (uc *UseCase) ResolveUserPermissions(ctx context.Context, userID string) (map[string]*dto.Source, error) {
	role, err := uc.userRepo.GetUserRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}
	if role == "admin" {
		result := make(map[string]*dto.Source, len(AllPermissions))
		for _, p := range AllPermissions {
			result[p.Key] = &dto.Source{Sources: []string{"role:admin"}}
		}
		return result, nil
	}
	result := make(map[string]*dto.Source)
	if defaults, ok := RoleDefaultPermissions[role]; ok {
		for _, perm := range defaults {
			result[perm] = &dto.Source{Sources: []string{"role:" + role}}
		}
	}
	memberships, err := uc.memberRepo.ListByUser(ctx, userID)
	if err != nil {
		uc.logger.Warnf("failed to list group memberships for user %s: %v", userID, err)
		return result, nil
	}
	for _, m := range memberships {
		group, err := uc.groupRepo.Get(ctx, m.GroupID)
		if err != nil {
			uc.logger.Warnf("failed to get group %s: %v", m.GroupID, err)
			continue
		}
		if !group.IsActive {
			continue
		}
		sourceLabel := "group:" + group.Name
		for _, perm := range group.Permissions {
			if existing, ok := result[perm]; ok {
				found := false
				for _, s := range existing.Sources {
					if s == sourceLabel {
						found = true
						break
					}
				}
				if !found {
					existing.Sources = append(existing.Sources, sourceLabel)
				}
				if len(group.CategoryScope) > 0 {
					existing.Scope = mergeScopes(existing.Scope, group.CategoryScope)
				}
			} else {
				src := &dto.Source{Sources: []string{sourceLabel}}
				if len(group.CategoryScope) > 0 {
					src.Scope = make([]string, len(group.CategoryScope))
					copy(src.Scope, group.CategoryScope)
				}
				result[perm] = src
			}
		}
	}
	return result, nil
}

func mergeScopes(existing, newScopes []string) []string {
	scopeSet := make(map[string]struct{}, len(existing)+len(newScopes))
	result := make([]string, 0, len(existing)+len(newScopes))
	for _, s := range existing {
		if _, ok := scopeSet[s]; !ok {
			scopeSet[s] = struct{}{}
			result = append(result, s)
		}
	}
	for _, s := range newScopes {
		if _, ok := scopeSet[s]; !ok {
			scopeSet[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

func (uc *UseCase) getUserPermissionSet(ctx context.Context, userID string) (map[string]*dto.Source, error) {
	if val, ok := uc.cache.Load(userID); ok {
		entry := val.(*permCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.permissions, nil
		}
		uc.cache.Delete(userID)
	}
	perms, err := uc.ResolveUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	uc.cache.Store(userID, &permCacheEntry{
		permissions: perms,
		expiresAt:   time.Now().Add(uc.cacheTTL),
	})
	return perms, nil
}

func (uc *UseCase) InvalidateUserCache(ctx context.Context, userID string) error {
	uc.cache.Delete(userID)
	select {
	case uc.notifyCh <- userID:
	default:
		uc.logger.Warnf("notify channel full, skipping broadcast for user %s", userID)
	}
	return nil
}

func (uc *UseCase) InvalidateGroupCache(ctx context.Context, groupID string) error {
	memberIDs, err := uc.groupRepo.GetMemberIDs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group members for cache invalidation: %w", err)
	}
	for _, uid := range memberIDs {
		_ = uc.InvalidateUserCache(ctx, uid)
	}
	return nil
}

func (uc *UseCase) GetUserPermissions(ctx context.Context, userID string) (*dto.UserPermissionDetail, error) {
	perms, err := uc.ResolveUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	role, err := uc.userRepo.GetUserRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}
	memberships, err := uc.memberRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	groups := make([]dto.UserGroupInfo, 0, len(memberships))
	for _, m := range memberships {
		group, gErr := uc.groupRepo.Get(ctx, m.GroupID)
		if gErr != nil {
			uc.logger.Warnf("failed to get group %s: %v", m.GroupID, gErr)
			continue
		}
		groups = append(groups, dto.UserGroupInfo{
			ID: group.ID, Name: group.Name, IsActive: group.IsActive, JoinedAt: m.JoinedAt,
		})
	}
	return &dto.UserPermissionDetail{
		UserID: userID, Role: role, EffectivePermissions: perms, Groups: groups,
	}, nil
}

func (uc *UseCase) CreateGroup(ctx context.Context, name, description string, permissions, categoryScope []string, createdBy string) (*dto.GroupItem, error) {
	for _, perm := range permissions {
		if !IsValidPermission(perm) {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
	}
	return uc.groupRepo.Create(ctx, name, description, permissions, categoryScope, createdBy)
}

func (uc *UseCase) GetGroup(ctx context.Context, id string) (*dto.GroupItem, error) {
	return uc.groupRepo.Get(ctx, id)
}

func (uc *UseCase) UpdateGroup(ctx context.Context, id, name, description string, permissions, categoryScope []string) (*dto.GroupItem, error) {
	for _, perm := range permissions {
		if !IsValidPermission(perm) {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
	}
	result, err := uc.groupRepo.Update(ctx, id, name, description, permissions, categoryScope)
	if err != nil {
		return nil, err
	}
	_ = uc.InvalidateGroupCache(ctx, id)
	return result, nil
}

func (uc *UseCase) DeleteGroup(ctx context.Context, id string) error {
	_ = uc.InvalidateGroupCache(ctx, id)
	return uc.groupRepo.Delete(ctx, id)
}

func (uc *UseCase) ToggleGroup(ctx context.Context, id string, isActive bool) error {
	err := uc.groupRepo.Toggle(ctx, id, isActive)
	if err != nil {
		return err
	}
	_ = uc.InvalidateGroupCache(ctx, id)
	return nil
}

func (uc *UseCase) ListGroup(ctx context.Context, isActive *bool, page, pageSize int) ([]*dto.GroupItem, int, error) {
	return uc.groupRepo.List(ctx, isActive, page, pageSize)
}

func (uc *UseCase) AddMembers(ctx context.Context, groupID string, userIDs []string) (int, int, error) {
	added, skipped, err := uc.memberRepo.AddMembers(ctx, groupID, userIDs)
	if err != nil {
		return 0, 0, err
	}
	for _, uid := range userIDs {
		_ = uc.InvalidateUserCache(ctx, uid)
	}
	return added, skipped, nil
}

func (uc *UseCase) RemoveMember(ctx context.Context, groupID, userID string) error {
	err := uc.memberRepo.RemoveMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	_ = uc.InvalidateUserCache(ctx, userID)
	return nil
}

func (uc *UseCase) ListMembers(ctx context.Context, groupID string, page, pageSize int) ([]*dto.MemberItem, int, error) {
	return uc.memberRepo.ListByGroup(ctx, groupID, page, pageSize)
}