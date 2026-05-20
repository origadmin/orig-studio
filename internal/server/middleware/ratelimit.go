package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors      map[string]*visitor
	mu            sync.RWMutex
	rate          rate.Limit
	burst         int
	stopCh        chan struct{}
	excludePrefix []string
	disabled      bool
}

func NewRateLimiter(rpm int, excludePrefixes ...string) *RateLimiter {
	disabled := false
	if rpm <= 0 {
		disabled = true
		rpm = 60
	}
	return &RateLimiter{
		visitors:      make(map[string]*visitor),
		rate:          rate.Every(time.Minute / time.Duration(rpm)),
		burst:         rpm,
		stopCh:        make(chan struct{}),
		excludePrefix: excludePrefixes,
		disabled:      disabled,
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) isExcluded(path string) bool {
	for _, prefix := range rl.excludePrefix {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		threshold := 30 * time.Minute
		for {
			select {
			case <-ticker.C:
				rl.mu.Lock()
				for ip, v := range rl.visitors {
					if time.Since(v.lastSeen) > threshold {
						delete(rl.visitors, ip)
					}
				}
				rl.mu.Unlock()
			case <-rl.stopCh:
				return
			}
		}
	}()

	return func(c *gin.Context) {
		if rl.disabled {
			c.Next()
			return
		}

		if rl.isExcluded(c.Request.URL.Path) {
			c.Next()
			return
		}

		ip := c.ClientIP()
		limiter := rl.getVisitor(ip)
		if !limiter.Allow() {
			c.Writer.Header().Set("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
