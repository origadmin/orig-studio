/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/notification"
	"origadmin/application/origcms/internal/features/content/biz"
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

func (r *notificationRepo) ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*biz.Notification, int, error) {
	query := r.data.db.Notification.Query().Where(notification.UserIDEQ(userID))
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	ents, err := query.
		Limit(pageSize).
		Offset((page - 1) * pageSize).
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