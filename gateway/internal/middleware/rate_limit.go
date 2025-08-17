package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)


type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}


func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}


func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	
	requests, exists := rl.requests[key]
	if !exists {
		requests = []time.Time{}
	}

	
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	
	if len(validRequests) >= rl.limit {
		return false
	}

	
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}


func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMinute, time.Minute)
	
	return func(c *gin.Context) {
		
		key := c.ClientIP()
		
		if !limiter.isAllowed(key) {
			log.Warn().
				Str("client_ip", key).
				Str("path", c.Request.URL.Path).
				Msg("Rate limit exceeded")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"retry_after": 60, // seconds
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}


