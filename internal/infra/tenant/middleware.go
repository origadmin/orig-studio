package tenant

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HeaderTenantID   = "X-Tenant-ID"
	HeaderTenantSlug = "X-Tenant-Slug"
	QueryTenantSlug  = "tenant"
)

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tc := resolveTenant(c)
		if tc != nil {
			ctx := WithContext(c.Request.Context(), tc)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

func resolveTenant(c *gin.Context) *Context {
	if slug := c.GetHeader(HeaderTenantSlug); slug != "" {
		return &Context{Slug: slug, Plan: "free"}
	}

	if id := c.GetHeader(HeaderTenantID); id != "" {
		return &Context{ID: id, Plan: "free"}
	}

	if slug := c.Query(QueryTenantSlug); slug != "" {
		return &Context{Slug: slug, Plan: "free"}
	}

	host := c.Request.Host
	if idx := strings.Index(host, "."); idx > 0 {
		subdomain := host[:idx]
		if subdomain != "www" && subdomain != "api" && subdomain != "admin" {
			return &Context{Slug: subdomain, Domain: host, Plan: "free"}
		}
	}

	return &Context{IsSystem: true, Slug: "default"}
}
