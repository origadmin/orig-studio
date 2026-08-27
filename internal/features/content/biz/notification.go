/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Notification action type vocabulary（受控枚举，设计锚点：
// docs/modules/content/notification/12-NOTIFICATION_TYPES_AND_PREFS.md）。
// 发送侧一律使用常量，禁止自由字符串。
const (
	NotificationActionSystem           = "system"
	NotificationActionTest             = "test"
	NotificationActionTranscodeSuccess = "media.transcode_success"
	NotificationActionTranscodeFailed  = "media.transcode_failed"
	NotificationActionMediaPublished   = "media.published"
	NotificationActionMediaDeleted     = "media.deleted"
	NotificationActionChannelSubscribed = "channel.subscribed"
	NotificationActionChannelNewVideo   = "channel.new_video"
	NotificationActionCommentCreated    = "comment.created"
	NotificationActionCommentReplied    = "comment.replied"
	NotificationActionCommentMentioned  = "comment.mentioned"
	NotificationActionLikeCreated       = "like.created"
	NotificationActionFollowCreated     = "follow.created"
	NotificationActionMediaReviewResult = "media.review_result"
	NotificationActionReportResolved    = "report.resolved"
	NotificationActionSystemAnnouncement = "system.announcement"
	NotificationActionSystemMaintenance  = "system.maintenance"
)

// Notification type categories (分组，设计锚点 §2 按域分组)。
const (
	NotificationCategoryMedia        = "media"        // 媒体/转码
	NotificationCategorySubscription = "subscription" // 订阅/频道
	NotificationCategoryInteraction  = "interaction"  // 互动/社交
	NotificationCategoryReport       = "report"       // 举报/反馈
	NotificationCategorySystem       = "system"       // 系统/运营
	NotificationCategoryTest         = "test"         // 测试
)

// NotificationType describes one supported notification type for the admin UI.
// Status: "active" = 已有真实功能（自动触发或手动发送）；"planned" = 自动触发待实现（UI 标「即将推出」禁用）。
type NotificationType struct {
	Action           string `json:"action"`
	LabelKey         string `json:"label_key"`
	Category         string `json:"category"`
	CategoryLabelKey string `json:"category_label_key"`
	Status           string `json:"status"`
	DefaultEnabled   bool   `json:"default_enabled"`
}

// Notification type status values.
const (
	NotificationStatusActive  = "active"  // 已实现（自动触发 / 手动发送均有效）
	NotificationStatusPlanned = "planned" // 待实现（自动触发未接，UI 标「即将推出」）
)

// labelKeyOf converts an action value into a flat i18n key
// (dots → underscores to avoid i18next key separator nesting).
func labelKeyOf(action string) string {
	return "notificationType." + strings.ReplaceAll(action, ".", "_")
}

// NotificationTypes returns the supported notification type vocabulary, grouped by category.
func NotificationTypes() []NotificationType {
	types := []NotificationType{
		{Action: NotificationActionTranscodeSuccess, Category: NotificationCategoryMedia, Status: NotificationStatusActive, DefaultEnabled: true},
		{Action: NotificationActionTranscodeFailed, Category: NotificationCategoryMedia, Status: NotificationStatusActive, DefaultEnabled: true},
		{Action: NotificationActionMediaPublished, Category: NotificationCategoryMedia, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionMediaDeleted, Category: NotificationCategoryMedia, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionMediaReviewResult, Category: NotificationCategoryMedia, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionChannelSubscribed, Category: NotificationCategorySubscription, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionChannelNewVideo, Category: NotificationCategorySubscription, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionCommentCreated, Category: NotificationCategoryInteraction, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionCommentReplied, Category: NotificationCategoryInteraction, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionCommentMentioned, Category: NotificationCategoryInteraction, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionLikeCreated, Category: NotificationCategoryInteraction, Status: NotificationStatusPlanned, DefaultEnabled: false},
		{Action: NotificationActionFollowCreated, Category: NotificationCategoryInteraction, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionReportResolved, Category: NotificationCategoryReport, Status: NotificationStatusPlanned, DefaultEnabled: true},
		{Action: NotificationActionSystemAnnouncement, Category: NotificationCategorySystem, Status: NotificationStatusActive, DefaultEnabled: true},
		{Action: NotificationActionSystemMaintenance, Category: NotificationCategorySystem, Status: NotificationStatusActive, DefaultEnabled: true},
		{Action: NotificationActionSystem, Category: NotificationCategorySystem, Status: NotificationStatusActive, DefaultEnabled: true},
		{Action: NotificationActionTest, Category: NotificationCategoryTest, Status: NotificationStatusActive, DefaultEnabled: true},
	}
	for i := range types {
		types[i].LabelKey = labelKeyOf(types[i].Action)
		types[i].CategoryLabelKey = "notificationCategory." + types[i].Category
	}
	return types
}

// defaultEnabledFor returns the taxonomy default for an action (true if unknown).
func defaultEnabledFor(action string) bool {
	for _, t := range NotificationTypes() {
		if t.Action == action {
			return t.DefaultEnabled
		}
	}
	return true
}

type Notification struct {
	ID         int       `json:"id"`
	Action     string    `json:"action"`
	Notify     bool      `json:"notify"`
	Method     string    `json:"method"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	IsRead     bool      `json:"read"`
}

// PermissionGroupInfo is a lightweight projection of an auth permission group,
// returned to the admin notification audience selector so broadcasts can target
// by group. MemberCount is computed from the auth_group_members relation.
type PermissionGroupInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MemberCount int       `json:"member_count"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   string    `json:"created_by"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

type NotificationRepo interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	Get(ctx context.Context, id int) (*Notification, error)
	Update(ctx context.Context, n *Notification) (*Notification, error)
	Delete(ctx context.Context, id int) error
	DeleteAllByUser(ctx context.Context, userID string) (int, error)
	DeleteReadByUser(ctx context.Context, userID string) (int, error)
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Notification, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*Notification, int, error)
	MarkAsRead(ctx context.Context, id int) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	GetAllUserIDs(ctx context.Context) ([]string, error)
	GetAllActiveUserIDs(ctx context.Context) ([]string, error)
	GetUserIDsByRole(ctx context.Context, roles []string) ([]string, error)
	GetUserIDsByGroup(ctx context.Context, groupIDs []string) ([]string, error)
	ListPermissionGroups(ctx context.Context) ([]*PermissionGroupInfo, error)
	BatchCreate(ctx context.Context, notifications []*Notification) ([]*Notification, error)
}

// SettingReader abstracts the global settings read used for notification type switches.
type SettingReader interface {
	// Get 读取设置，可能命中进程内缓存（60s）。
	Get(ctx context.Context, key string) string
	// GetNoCache 直读 DB，绕过缓存。通知广播可能由不同微服务触发，
	// 而全局类型开关由管理端写入另一服务的缓存 → 必须实时读取，
	// 否则关闭类型后广播仍可能发送（BUG-265 回归）。
	GetNoCache(ctx context.Context, key string) string
}

type NotificationUseCase struct {
	repo     NotificationRepo
	settings SettingReader
	log      *log.Helper
}

func NewNotificationUseCase(repo NotificationRepo, settings SettingReader, logger log.Logger) *NotificationUseCase {
	return &NotificationUseCase{
		repo:     repo,
		settings: settings,
		log:      log.NewHelper(log.With(logger, "module", "notification.biz")),
	}
}

// isTypeEnabled 报告通知类型是否全局启用（设计锚点 §3.1）。
// 全局开关存于 settings `notification.type.{action}.enabled`；未设置（空）时回退 taxonomy 默认。
// 用 GetNoCache：开关由管理端写入 settings，广播可能在另一微服务进程执行，
// 进程内 60s 缓存会导致关闭后广播仍发送（BUG-265 回归实测 sent_count=10）。
func (uc *NotificationUseCase) isTypeEnabled(ctx context.Context, action string) bool {
	if uc.settings == nil {
		return true
	}
	v := uc.settings.GetNoCache(ctx, "notification.type."+action+".enabled")
	if v == "" {
		return defaultEnabledFor(action)
	}
	return v == "true" || v == "1"
}

// typeSettingKey 返回通知类型全局开关的 settings key。
func typeSettingKey(action string) string {
	return "notification.type." + action + ".enabled"
}

func (uc *NotificationUseCase) CreateNotification(ctx context.Context, n *Notification) (*Notification, error) {
	return uc.repo.Create(ctx, n)
}

// ListPermissionGroups returns all permission groups (with member counts) so the
// admin notification UI can offer them as a broadcast target.
func (uc *NotificationUseCase) ListPermissionGroups(ctx context.Context) ([]*PermissionGroupInfo, error) {
	return uc.repo.ListPermissionGroups(ctx)
}

func (uc *NotificationUseCase) ListUserNotifications(ctx context.Context, userID string, page, pageSize int) ([]*Notification, int, error) {
	return uc.repo.ListByUser(ctx, userID, page, pageSize)
}

func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, id int, userID string) error {
	n, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if n.UserID != userID {
		return nil
	}
	return uc.repo.MarkAsRead(ctx, id)
}

func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, userID string) error {
	return uc.repo.MarkAllAsRead(ctx, userID)
}

func (uc *NotificationUseCase) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return uc.repo.GetUnreadCount(ctx, userID)
}

func (uc *NotificationUseCase) DeleteNotification(ctx context.Context, id int, userID string) error {
	n, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if n.UserID != userID {
		return nil
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *NotificationUseCase) DeleteAllNotifications(ctx context.Context, userID string) (int, error) {
	return uc.repo.DeleteAllByUser(ctx, userID)
}

func (uc *NotificationUseCase) DeleteReadNotifications(ctx context.Context, userID string) (int, error) {
	return uc.repo.DeleteReadByUser(ctx, userID)
}

func (uc *NotificationUseCase) AdminDeleteNotification(ctx context.Context, id int) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *NotificationUseCase) ListAllNotifications(ctx context.Context, page, pageSize int) ([]*Notification, int, error) {
	return uc.repo.ListAll(ctx, page, pageSize)
}

func (uc *NotificationUseCase) BatchCreateNotifications(ctx context.Context, userIDs []string, n *Notification) ([]*Notification, error) {
	// 全局类型开关过滤（设计锚点 §3.1/§3.3）：类型关闭则不发送。
	if !uc.isTypeEnabled(ctx, n.Action) {
		uc.log.Infof("notification type %s globally disabled, skip send", n.Action)
		return []*Notification{}, nil
	}
	var notifications []*Notification
	for _, uid := range userIDs {
		notif := &Notification{
			Action: n.Action,
			Notify: n.Notify,
			Method: n.Method,
			UserID: uid,
			Title:  n.Title,
			Body:   n.Body,
		}
		notifications = append(notifications, notif)
	}
	return uc.repo.BatchCreate(ctx, notifications)
}

func (uc *NotificationUseCase) BroadcastToAll(ctx context.Context, n *Notification) ([]*Notification, error) {
	userIDs, err := uc.repo.GetAllActiveUserIDs(ctx)
	if err != nil {
		return nil, err
	}
	return uc.BatchCreateNotifications(ctx, userIDs, n)
}

// BroadcastByRole broadcasts a notification to all active users matching any of the given roles.
func (uc *NotificationUseCase) BroadcastByRole(ctx context.Context, roles []string, n *Notification) ([]*Notification, error) {
	userIDs, err := uc.repo.GetUserIDsByRole(ctx, roles)
	if err != nil {
		return nil, err
	}
	return uc.BatchCreateNotifications(ctx, userIDs, n)
}

// BroadcastByGroup broadcasts a notification to all members of the given permission groups.
func (uc *NotificationUseCase) BroadcastByGroup(ctx context.Context, groupIDs []string, n *Notification) ([]*Notification, error) {
	userIDs, err := uc.repo.GetUserIDsByGroup(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	return uc.BatchCreateNotifications(ctx, userIDs, n)
}

// ResolveTargetUsers resolves the final list of target user IDs based on multiple filter inputs.
// Priority: explicit user_ids > roles > groups > all active users.
func (uc *NotificationUseCase) ResolveTargetUsers(ctx context.Context, userIDs []string, roles []string, groupIDs []string) ([]string, error) {
	if len(userIDs) > 0 {
		return userIDs, nil
	}
	if len(roles) > 0 {
		return uc.repo.GetUserIDsByRole(ctx, roles)
	}
	if len(groupIDs) > 0 {
		return uc.repo.GetUserIDsByGroup(ctx, groupIDs)
	}
	return uc.repo.GetAllActiveUserIDs(ctx)
}

func (uc *NotificationUseCase) SendTestNotification(ctx context.Context, userID string) (*Notification, error) {
	n := &Notification{
		Action: "test",
		Notify: false,
		Method: "in_app",
		UserID: userID,
		Title:  "Test Notification",
		Body:   "This is a test notification from the admin panel. If you see this, notifications are working correctly.",
	}
	return uc.repo.Create(ctx, n)
}

// NotifyTranscodeStatus 在转码达到终态时通知上传者（设计锚点 §4）。
// status: "success"（含 partial，视为可播放）/ 其他（failed）。幂等由调用方保证
// （仅在首次达到终态的流程路径调用，redelivery 分支不重复通知）。
func (uc *NotificationUseCase) NotifyTranscodeStatus(ctx context.Context, userID, mediaTitle, status, errMsg string) error {
	if userID == "" {
		return nil
	}
	n := &Notification{
		Action: NotificationActionTranscodeSuccess,
		Notify: false,
		Method: "in_app",
		UserID: userID,
	}
	if status == "success" {
		n.Title = "转码完成"
		n.Body = fmt.Sprintf("视频《%s》转码成功，已可播放。", mediaTitle)
	} else {
		n.Action = NotificationActionTranscodeFailed
		n.Title = "转码失败"
		n.Body = fmt.Sprintf("视频《%s》转码失败，请检查后重试。", mediaTitle)
		if errMsg != "" {
			n.Body += " 原因：" + errMsg
		}
	}
	// 全局类型开关过滤（设计锚点 §3.1）：类型关闭则系统自动通知也不发。
	if !uc.isTypeEnabled(ctx, n.Action) {
		uc.log.Infof("notification type %s globally disabled, skip transcode notify", n.Action)
		return nil
	}
	_, err := uc.repo.Create(ctx, n)
	return err
}