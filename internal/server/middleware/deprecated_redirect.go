package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeprecatedRedirects() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		newPath := ""
		switch {
		case matchPrefix(path, "/api/v1/users/username/"):
			username := trimPrefix(path, "/api/v1/users/username/")
			newPath = "/api/v1/users"
			if query != "" {
				query += "&username=" + username
			} else {
				query = "username=" + username
			}

		case matchPrefix(path, "/api/v1/users/slug/"):
			slug := trimPrefix(path, "/api/v1/users/slug/")
			newPath = "/api/v1/users/" + slug

		case matchPrefix(path, "/api/v1/articles/slug/"):
			slug := trimPrefix(path, "/api/v1/articles/slug/")
			newPath = "/api/v1/articles/" + slug

		case matchPrefix(path, "/api/v1/ads/placement/"):
			slug := trimPrefix(path, "/api/v1/ads/placement/")
			newPath = "/api/v1/ads"
			if query != "" {
				query += "&placement=" + slug
			} else {
				query = "placement=" + slug
			}

		case matchPrefix(path, "/api/v1/resolve/@"):
			handle := trimPrefix(path, "/api/v1/resolve/@")
			newPath = "/api/v1/resolve"
			if query != "" {
				query += "&handle=" + handle
			} else {
				query = "handle=" + handle
			}
		}

		if newPath != "" {
			redirectURL := newPath
			if query != "" {
				redirectURL += "?" + query
			}
			c.Redirect(http.StatusMovedPermanently, redirectURL)
			c.Abort()
			return
		}

		c.Next()
	}
}

func matchPrefix(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}

func trimPrefix(path, prefix string) string {
	if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return path
}
