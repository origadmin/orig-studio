// Package server provides HTTP handlers
package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler returns the health status of the server.
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
