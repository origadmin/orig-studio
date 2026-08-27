package server

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthStatus struct {
	Status    string                    `json:"status"`
	Timestamp string                    `json:"timestamp"`
	Checks   map[string]*ComponentCheck `json:"checks"`
}

type ComponentCheck struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) DetailedHealthHandler(c *gin.Context) {
	now := time.Now().UTC().Format(time.RFC3339)
	checks := make(map[string]*ComponentCheck)
	allHealthy := true

	dbCheck := s.checkDatabase(c.Request.Context())
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		allHealthy = false
	}

	storageCheck := s.checkStorage()
	checks["storage"] = storageCheck
	if storageCheck.Status != "healthy" {
		allHealthy = false
	}

	jwtCheck := s.checkJWT()
	checks["jwt"] = jwtCheck
	if jwtCheck.Status != "healthy" {
		allHealthy = false
	}

	overall := "healthy"
	statusCode := http.StatusOK
	if !allHealthy {
		overall = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthStatus{
		Status:    overall,
		Timestamp: now,
		Checks:    checks,
	})
}

func (s *Server) checkDatabase(ctx context.Context) *ComponentCheck {
	if s.sqlDB == nil {
		return &ComponentCheck{
			Status: "degraded",
			Error:  "sql.DB not available (detailed check skipped)",
		}
	}

	start := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.sqlDB.PingContext(pingCtx); err != nil {
		return &ComponentCheck{
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   err.Error(),
		}
	}

	return &ComponentCheck{
		Status:  "healthy",
		Latency: time.Since(start).String(),
	}
}

func (s *Server) checkStorage() *ComponentCheck {
	if s.paths == nil {
		return &ComponentCheck{
			Status: "unhealthy",
			Error:  "storage paths not initialized",
		}
	}

	basePath := s.paths.BasePath()
	info, err := os.Stat(basePath)
	if err != nil {
		return &ComponentCheck{
			Status: "unhealthy",
			Error:  "storage path inaccessible: " + err.Error(),
		}
	}
	if !info.IsDir() {
		return &ComponentCheck{
			Status: "unhealthy",
			Error:  "storage path is not a directory",
		}
	}

	testFile := basePath + string(os.PathSeparator) + ".health_check_write_test"
	if err := os.WriteFile(testFile, []byte("health"), 0600); err != nil {
		return &ComponentCheck{
			Status: "unhealthy",
			Error:  "storage path not writable: " + err.Error(),
		}
	}
	os.Remove(testFile)

	return &ComponentCheck{
		Status: "healthy",
	}
}

func (s *Server) checkJWT() *ComponentCheck {
	if s.jwtMgr == nil {
		return &ComponentCheck{
			Status: "unhealthy",
			Error:  "JWT manager not initialized",
		}
	}

	start := time.Now()
	token, err := s.jwtMgr.Generate("health-check", "system", "health")
	if err != nil {
		return &ComponentCheck{
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   "token generation failed: " + err.Error(),
		}
	}

	claims, err := s.jwtMgr.Parse(token)
	if err != nil {
		return &ComponentCheck{
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   "token parsing failed: " + err.Error(),
		}
	}

	if claims.GetUserID() != "health-check" {
		return &ComponentCheck{
			Status:  "unhealthy",
			Latency: time.Since(start).String(),
			Error:   "token claims mismatch",
		}
	}

	return &ComponentCheck{
		Status:  "healthy",
		Latency: time.Since(start).String(),
	}
}
