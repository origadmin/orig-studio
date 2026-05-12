/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

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
	IsRead     bool      `json:"is_read"`
}

type NotificationRepo interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	Get(ctx context.Context, id int) (*Notification, error)
	Update(ctx context.Context, n *Notification) (*Notification, error)
	Delete(ctx context.Context, id int) error
	ListByUser(ctx context.Context, userID string, page, pageSize int) ([]*Notification, int, error)
	MarkAsRead(ctx context.Context, id int) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type NotificationUseCase struct {
	repo NotificationRepo
	log  *log.Helper
}

func NewNotificationUseCase(repo NotificationRepo, logger log.Logger) *NotificationUseCase {
	return &NotificationUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "notification.biz")),
	}
}

func (uc *NotificationUseCase) CreateNotification(ctx context.Context, n *Notification) (*Notification, error) {
	return uc.repo.Create(ctx, n)
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