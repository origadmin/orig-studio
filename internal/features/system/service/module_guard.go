package service

import (
	http2 "origadmin/application/origstudio/internal/pkg/http"
	"origadmin/application/origstudio/internal/infra/auth"
	systembiz "origadmin/application/origstudio/internal/features/system/biz"
)

var moduleDefaults = map[string]bool{
	"module_articles": true,
	"module_videos":   true,
	"module_music":    false,
}

// ModuleGuardCtx returns a MiddlewareFunc that checks if a module is enabled.
// Admin/staff users bypass the check. For regular users, the setting value
// is looked up; if not set, the module default is used.
func ModuleGuardCtx(settingUC *systembiz.SettingUseCase, moduleKey string) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			// Admin/staff bypass
			if claims, ok := ctx.Get("claims"); ok {
				if cl, ok := claims.(*auth.Claims); ok {
					if cl.Role == "admin" || cl.IsAdmin() {
						return next(ctx)
					}
				}
			}

			if settingUC == nil {
				return next(ctx)
			}

			enabled := settingUC.GetBool(ctx.Request().Context(), moduleKey)
			if !enabled {
				val := settingUC.Get(ctx.Request().Context(), moduleKey)
				if val == "" {
					if def, ok := moduleDefaults[moduleKey]; ok {
						enabled = def
					} else {
						enabled = true
					}
				}
			}
			if !enabled {
				http2.Fail(ctx, http2.AppErrNotFound, "This content module is not available")
				return nil
			}
			return next(ctx)
		}
	}
}
