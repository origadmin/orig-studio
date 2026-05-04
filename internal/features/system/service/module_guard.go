package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"origadmin/application/origcms/internal/infra/auth"
	systembiz "origadmin/application/origcms/internal/features/system/biz"
)

var moduleDefaults = map[string]bool{
	"module_articles": true,
	"module_videos":   true,
	"module_music":    false,
}

func ModuleGuard(settingUC *systembiz.SettingUseCase, moduleKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, exists := c.Get("claims"); exists {
			if cl, ok := claims.(*auth.Claims); ok {
				if cl.Role == "admin" || cl.IsStaff {
					c.Next()
					return
				}
			}
		}

		if settingUC == nil {
			c.Next()
			return
		}

		enabled := settingUC.GetBool(c.Request.Context(), moduleKey)
		if !enabled {
			val := settingUC.Get(c.Request.Context(), moduleKey)
			if val == "" {
				if def, ok := moduleDefaults[moduleKey]; ok {
					enabled = def
				} else {
					enabled = true
				}
			}
		}
		if !enabled {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "This content module is not available",
			})
			return
		}
		c.Next()
	}
}
