package biz

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

var RoleDefaultPermissions = map[string][]string{
	"admin":  {"media:read", "media:write", "media:delete", "media:publish", "media:moderate", "comment:read", "comment:write", "comment:moderate", "user:manage", "system:config"},
	"editor": {"media:read", "media:write", "comment:read", "comment:write", "comment:moderate"},
	"user":   {"media:read", "comment:read", "comment:write"},
}

var AllPermissions = []struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	ResourceType string `json:"resource_type"`
	Action       string `json:"action"`
}{
	{"media:read", "查看媒体", "media", "read"},
	{"media:write", "编辑媒体", "media", "write"},
	{"media:delete", "删除媒体", "media", "delete"},
	{"media:publish", "发布媒体", "media", "publish"},
	{"media:moderate", "审核媒体", "media", "moderate"},
	{"comment:read", "查看评论", "comment", "read"},
	{"comment:write", "发表评论", "comment", "write"},
	{"comment:moderate", "审核评论", "comment", "moderate"},
	{"user:manage", "管理用户", "user", "manage"},
	{"system:config", "系统配置", "system", "config"},
}

func IsValidPermission(perm string) bool {
	for _, p := range AllPermissions {
		if p.Key == perm {
			return true
		}
	}
	return false
}

type PermissionGroupItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Permissions   []string  `json:"permissions"`
	CategoryScope []string  `json:"category_scope,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedBy     string    `json:"created_by,omitempty"`
	MemberCount   int       `json:"member_count"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

type GroupMemberItem struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username,omitempty"`
	GroupID  string    `json:"group_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type UserPermissionDetail struct {
	UserID               string                       `json:"user_id"`
	Role                 string                       `json:"role"`
	EffectivePermissions map[string]*PermissionSource `json:"effective_permissions"`
	Groups               []UserGroupInfo              `json:"groups"`
}

type PermissionSource struct {
	Sources []string `json:"sources"`
	Scope   []string `json:"scope,omitempty"`
}

type UserGroupInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IsActive bool      `json:"is_active"`
	JoinedAt time.Time `json:"joined_at"`
}

type PermissionGroupRepo interface {
	Create(ctx context.Context, name, description string, permissions, categoryScope []string, createdBy string) (*PermissionGroupItem, error)
	Get(ctx context.Context, id string) (*PermissionGroupItem, error)
	Update(ctx context.Context, id, name, description string, permissions, categoryScope []string) (*PermissionGroupItem, error)
	Delete(ctx context.Context, id string) error
	Toggle(ctx context.Context, id string, isActive bool) error
	List(ctx context.Context, isActive *bool, page, pageSize int) ([]*PermissionGroupItem, int, error)
	GetMemberIDs(ctx context.Context, groupID string) ([]string, error)
}

type GroupMemberRepo interface {
	AddMembers(ctx context.Context, groupID string, userIDs []string) (added int, skipped int, err error)
	RemoveMember(ctx context.Context, groupID, userID string) error
	ListByGroup(ctx context.Context, groupID string, page, pageSize int) ([]*GroupMemberItem, int, error)
	ListByUser(ctx context.Context, userID string) ([]*GroupMemberItem, error)
}

type UserPermRepo interface {
	GetUserRole(ctx context.Context, userID string) (string, error)
}

type PermissionChecker interface {
	CheckPermission(ctx context.Context, userID string, permission string, categoryID string) (bool, error)
	InvalidateUserCache(ctx context.Context, userID string) error
	InvalidateGroupCache(ctx context.Context, groupID string) error
}

type permCacheEntry struct {
	permissions map[string]*PermissionSource
	expiresAt   time.Time
}

type PermissionUseCase struct {
	groupRepo  PermissionGroupRepo
	memberRepo GroupMemberRepo
	userRepo   UserPermRepo
	cache      sync.Map
	cacheTTL   time.Duration
	notifyCh   chan string
	logger     *log.Helper
}

func NewPermissionUseCase(
	groupRepo PermissionGroupRepo,
	memberRepo GroupMemberRepo,
	userRepo UserPermRepo,
	logger log.Logger,
) *PermissionUseCase {
	uc := &PermissionUseCase{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		cacheTTL:   5 * time.Minute,
		notifyCh:   make(chan string, 100),
		logger:     log.NewHelper(log.With(logger, "module", "svc-auth/permission.biz")),
	}
	go uc.processInvalidation()
	return uc
}

func (uc *PermissionUseCase) processInvalidation() {
	for userID := range uc.notifyCh {
		uc.cache.Delete(userID)
	}
}

func (uc *PermissionUseCase) CheckPermission(ctx context.Context, userID string, permission string, categoryID string) (bool, error) {
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

func (uc *PermissionUseCase) ResolveUserPermissions(ctx context.Context, userID string) (map[string]*PermissionSource, error) {
	role, err := uc.userRepo.GetUserRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	if role == "admin" {
		result := make(map[string]*PermissionSource, len(AllPermissions))
		for _, p := range AllPermissions {
			result[p.Key] = &PermissionSource{
				Sources: []string{"role:admin"},
			}
		}
		return result, nil
	}

	result := make(map[string]*PermissionSource)

	if defaults, ok := RoleDefaultPermissions[role]; ok {
		for _, perm := range defaults {
			result[perm] = &PermissionSource{
				Sources: []string{"role:" + role},
			}
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
				src := &PermissionSource{
					Sources: []string{sourceLabel},
				}
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

func (uc *PermissionUseCase) getUserPermissionSet(ctx context.Context, userID string) (map[string]*PermissionSource, error) {
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

func (uc *PermissionUseCase) InvalidateUserCache(ctx context.Context, userID string) error {
	uc.cache.Delete(userID)
	select {
	case uc.notifyCh <- userID:
	default:
		uc.logger.Warnf("notify channel full, skipping broadcast for user %s", userID)
	}
	return nil
}

func (uc *PermissionUseCase) InvalidateGroupCache(ctx context.Context, groupID string) error {
	memberIDs, err := uc.groupRepo.GetMemberIDs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group members for cache invalidation: %w", err)
	}
	for _, uid := range memberIDs {
		_ = uc.InvalidateUserCache(ctx, uid)
	}
	return nil
}

func (uc *PermissionUseCase) GetUserPermissions(ctx context.Context, userID string) (*UserPermissionDetail, error) {
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

	groups := make([]UserGroupInfo, 0, len(memberships))
	for _, m := range memberships {
		group, gErr := uc.groupRepo.Get(ctx, m.GroupID)
		if gErr != nil {
			uc.logger.Warnf("failed to get group %s: %v", m.GroupID, gErr)
			continue
		}
		groups = append(groups, UserGroupInfo{
			ID:       group.ID,
			Name:     group.Name,
			IsActive: group.IsActive,
			JoinedAt: m.JoinedAt,
		})
	}

	return &UserPermissionDetail{
		UserID:               userID,
		Role:                 role,
		EffectivePermissions: perms,
		Groups:               groups,
	}, nil
}

func (uc *PermissionUseCase) CreateGroup(ctx context.Context, name, description string, permissions, categoryScope []string, createdBy string) (*PermissionGroupItem, error) {
	for _, perm := range permissions {
		if !IsValidPermission(perm) {
			return nil, fmt.Errorf("invalid permission: %s", perm)
		}
	}
	return uc.groupRepo.Create(ctx, name, description, permissions, categoryScope, createdBy)
}

func (uc *PermissionUseCase) GetGroup(ctx context.Context, id string) (*PermissionGroupItem, error) {
	return uc.groupRepo.Get(ctx, id)
}

func (uc *PermissionUseCase) UpdateGroup(ctx context.Context, id, name, description string, permissions, categoryScope []string) (*PermissionGroupItem, error) {
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

func (uc *PermissionUseCase) DeleteGroup(ctx context.Context, id string) error {
	_ = uc.InvalidateGroupCache(ctx, id)
	return uc.groupRepo.Delete(ctx, id)
}

func (uc *PermissionUseCase) ToggleGroup(ctx context.Context, id string, isActive bool) error {
	err := uc.groupRepo.Toggle(ctx, id, isActive)
	if err != nil {
		return err
	}
	_ = uc.InvalidateGroupCache(ctx, id)
	return nil
}

func (uc *PermissionUseCase) ListGroup(ctx context.Context, isActive *bool, page, pageSize int) ([]*PermissionGroupItem, int, error) {
	return uc.groupRepo.List(ctx, isActive, page, pageSize)
}

func (uc *PermissionUseCase) AddMembers(ctx context.Context, groupID string, userIDs []string) (int, int, error) {
	added, skipped, err := uc.memberRepo.AddMembers(ctx, groupID, userIDs)
	if err != nil {
		return 0, 0, err
	}
	for _, uid := range userIDs {
		_ = uc.InvalidateUserCache(ctx, uid)
	}
	return added, skipped, nil
}

func (uc *PermissionUseCase) RemoveMember(ctx context.Context, groupID, userID string) error {
	err := uc.memberRepo.RemoveMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	_ = uc.InvalidateUserCache(ctx, userID)
	return nil
}

func (uc *PermissionUseCase) ListMembers(ctx context.Context, groupID string, page, pageSize int) ([]*GroupMemberItem, int, error) {
	return uc.memberRepo.ListByGroup(ctx, groupID, page, pageSize)
}
