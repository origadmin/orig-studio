package dal

import (
	"context"

	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	permbiz "origadmin/application/origstudio/internal/enterprise/auth/permission/biz"
)

type CheckerAdapter struct {
	uc *permbiz.UseCase
}

func NewCheckerAdapter(uc *permbiz.UseCase) *CheckerAdapter {
	return &CheckerAdapter{uc: uc}
}

func (a *CheckerAdapter) CheckPermission(ctx context.Context, userID string, permission string, categoryID string) (bool, error) {
	return a.uc.CheckPermission(ctx, userID, permission, categoryID)
}

func (a *CheckerAdapter) InvalidateUserCache(ctx context.Context, userID string) error {
	return a.uc.InvalidateUserCache(ctx, userID)
}

func (a *CheckerAdapter) InvalidateGroupCache(ctx context.Context, groupID string) error {
	return a.uc.InvalidateGroupCache(ctx, groupID)
}

var _ authbiz.PermissionChecker = (*CheckerAdapter)(nil)
