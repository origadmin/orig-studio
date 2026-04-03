// Package server provides HTTP handlers
package server

import (
	"github.com/gin-gonic/gin"
)

// Module interface for route registration
type Module interface {
	Register(group *gin.RouterGroup)
}

// RegisterRoutes registers all HTTP routes
func RegisterRoutes(router *gin.Engine, mods ...Module) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		for _, mod := range mods {
			mod.Register(v1)
		}
	}
}
