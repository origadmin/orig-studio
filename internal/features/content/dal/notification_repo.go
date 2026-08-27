/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/groupmember"
	"origadmin/application/origstudio/internal/dal/entity/notification"
	"origadmin/application/origstudio/internal/dal/entity/permissiongroup"
	"origadmin/application/origstudio/internal/dal/entity/user"
	"origadmin/application/origstudio/internal/domain/types"
	"origadmin/application/origstudio/internal/features/content/biz"
)

type notificationRepo struct {
	data *Data
	log  *log.Helper
}

func NewNotificationRepo(data *Data, logger log.Logger) biz.NotificationRepo {
	return &notificationRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "notification.data")),
	}
}

func (r *notificationRepo) Create(ctx context.Context, n *biz.Notification) (*biz.Notification, error) {
	ent, err := r.data.db.Notification.Create().
		SetAction(n.Action).
		SetNotify(n.Notify).
		SetMethod(n.Method).
		SetUserID(n.UserID).
		SetTitle(n.Title).
		SetBody(n.Body).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapNotification(ent), nil
}

func (r *notificationRepo) Get(ctx context.Context, id int) (*biz.Notification, error) {
	ent, err := r.data.db.Notification.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapNotification(ent), nil
}

func (r *notificationRepo) Update(ctx context.Context, n *biz.Notification) (*biz.Notification, error) {
	ent, err := r.data.db.Notification.UpdateOneID(n.ID).
		SetAction(n.Action).
		SetNotify(n.Notify).
		SetMethod(n.Method).
		SetTitle(n.Title).
		SetBody(n.Body).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapNotification(ent), nil
}

func (r *notificationRepo) Delete(ctx context.Context, id int) error {
	return r.data.db.Notification.DeleteOneID(id).Exec(ctx)
}

func (r *notificationRepo) DeleteAllByUser(ctx context.Context, userID string) (int, error) {
	return r.data.db.Notification.Delete().
		Where(notification.UserIDEQ(userID)).
		Exec(ctx)
}

func (r *notificationRepo) DeleteReadByUser(ctx context.Context, userID string) (int, error) {
	return r.data.db.Notification.Delete().
		Where(notification.UserIDEQ(userID), notification.IsReadEQ(true)).
		Exec(ctx)
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*biz.Notification, int, error) {
	query := r.data.db.Notification.Query().Where(notification.UserIDEQ(userID))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset(types.CalcOffset(page, pageSize)).
		Order(entity.Desc(notification.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.Notification, len(ents))
	for i, ent := range ents {
		res[i] = mapNotification(ent)
	}
	return res, total, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, id int) error {
	return r.data.db.Notification.UpdateOneID(id).SetIsRead(true).Exec(ctx)
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	_, err := r.data.db.Notification.Update().
		Where(notification.UserIDEQ(userID)).
		SetIsRead(true).
		Save(ctx)
	return err
}

func (r *notificationRepo) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return r.data.db.Notification.Query().
		Where(notification.UserIDEQ(userID), notification.IsReadEQ(false)).
		Count(ctx)
}

func (r *notificationRepo) ListAll(ctx context.Context, page, pageSize int) ([]*biz.Notification, int, error) {
	query := r.data.db.Notification.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset(types.CalcOffset(page, pageSize)).
		Order(entity.Desc(notification.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*biz.Notification, len(ents))
	for i, ent := range ents {
		res[i] = mapNotification(ent)
	}
	return res, total, nil
}

func (r *notificationRepo) GetAllUserIDs(ctx context.Context) ([]string, error) {
	users, err := r.data.db.User.Query().Select("id").Strings(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *notificationRepo) BatchCreate(ctx context.Context, notifications []*biz.Notification) ([]*biz.Notification, error) {
	var created []*biz.Notification
	for _, n := range notifications {
		ent, err := r.data.db.Notification.Create().
			SetAction(n.Action).
			SetNotify(n.Notify).
			SetMethod(n.Method).
			SetUserID(n.UserID).
			SetTitle(n.Title).
			SetBody(n.Body).
			Save(ctx)
		if err != nil {
			r.log.Warnf("failed to create notification for user %s: %v", n.UserID, err)
			continue
		}
		created = append(created, mapNotification(ent))
	}
	return created, nil
}

// GetUserIDsByRole returns active user IDs matching any of the given roles.
func (r *notificationRepo) GetUserIDsByRole(ctx context.Context, roles []string) ([]string, error) {
	if len(roles) == 0 {
		return nil, nil
	}
	roleEnums := make([]user.Role, 0, len(roles))
	for _, role := range roles {
		roleEnums = append(roleEnums, user.Role(role))
	}
	users, err := r.data.db.User.Query().
		Where(user.RoleIn(roleEnums...)).
		Where(user.StatusEQ("ACTIVE")).
		Select("id").
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserIDsByGroup returns distinct user IDs that are members of the given permission groups.
func (r *notificationRepo) GetUserIDsByGroup(ctx context.Context, groupIDs []string) ([]string, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}
	members, err := r.data.db.GroupMember.Query().
		Where(groupmember.GroupIDIn(groupIDs...)).
		Select("user_id").
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	// deduplicate
	seen := make(map[string]struct{}, len(members))
	result := make([]string, 0, len(members))
	for _, m := range members {
		if _, ok := seen[m]; !ok {
			seen[m] = struct{}{}
			result = append(result, m)
		}
	}
	return result, nil
}

// ListPermissionGroups returns all permission groups with their computed member counts.
func (r *notificationRepo) ListPermissionGroups(ctx context.Context) ([]*biz.PermissionGroupInfo, error) {
	groups, err := r.data.db.PermissionGroup.Query().Order(permissiongroup.ByCreateTime()).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*biz.PermissionGroupInfo, 0, len(groups))
	for _, g := range groups {
		count, err := r.data.db.GroupMember.Query().Where(groupmember.GroupIDEQ(g.ID)).Count(ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, &biz.PermissionGroupInfo{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			MemberCount: count,
			IsActive:    g.IsActive,
			CreatedBy:   g.CreatedBy,
			CreateTime:  g.CreateTime,
			UpdateTime:  g.UpdateTime,
		})
	}
	return out, nil
}

// GetAllActiveUserIDs returns all active user IDs (for broadcast-to-all with proper state filter).
func (r *notificationRepo) GetAllActiveUserIDs(ctx context.Context) ([]string, error) {
	users, err := r.data.db.User.Query().
		Where(user.StatusEQ("ACTIVE")).
		Select("id").
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func mapNotification(ent *entity.Notification) *biz.Notification {
	return &biz.Notification{
		ID:         ent.ID,
		Action:     ent.Action,
		Notify:     ent.Notify,
		Method:     ent.Method,
		UserID:     ent.UserID,
		Title:      ent.Title,
		Body:       ent.Body,
		IsRead:     ent.IsRead,
		CreateTime: ent.CreateTime,
		UpdateTime: ent.UpdateTime,
	}
}