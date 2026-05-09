package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(rpm int) *RateLimiter {
	if rpm <= 0 {
		rpm = 60
	}
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     rate.Every(time.Minute / time.Duration(rpm)),
		burst:    rpm,
	}
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}
	return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			rl.mu.Lock()
			rl.visitors = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getVisitor(ip)
		if !limiter.Allow() {
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
