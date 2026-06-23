package server

import (
	"context"
	"strings"

	http2 "origadmin/application/origstudio/internal/pkg/http"
)

type AuditLogger interface {
	Log(ctx context.Context, userID, username, action, resource, ip, userAgent, result string) error
}

func Auditable(logger AuditLogger) http2.MiddlewareFunc {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx http2.Context) error {
			claims, _ := GetClaimsCtx(ctx)
			req := ctx.Request()

			action := inferAction(req.Method)
			resource := inferResource(req.URL.Path)

			err := next(ctx)

			result := "success"
			if err != nil {
				result = "failure"
			}

			userID := ""
			username := ""
			ip := req.RemoteAddr
			userAgent := req.UserAgent()

			if claims != nil {
				userID = claims.GetUserID()
				username = claims.Username
			}

			_ = logger.Log(ctx, userID, username, action, resource, ip, userAgent, result)

			return err
		}
	}
}

func inferAction(method string) string {
	switch strings.ToUpper(method) {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	case "GET":
		return "read"
	default:
		return method
	}
}

func inferResource(path string) string {
	path = strings.TrimPrefix(path, "/admin/")
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "system"
}