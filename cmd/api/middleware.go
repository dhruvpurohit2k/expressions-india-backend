package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/dhruvpurohit2k/expressions-india-backend/internal/pkg/utils"
	"github.com/gin-gonic/gin"
)

// MaxBodyBytes caps the size of the request body to prevent OOM from large uploads.
func MaxBodyBytes(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}

// Simple per-IP token bucket. Not distributed — fine for a single-instance deploy.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

type ipLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64
}

func newIPLimiter(rate float64, capacity float64) *ipLimiter {
	l := &ipLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		capacity: capacity,
	}
	go l.gc()
	return l
}

func (l *ipLimiter) gc() {
	for range time.Tick(10 * time.Minute) {
		l.mu.Lock()
		cutoff := time.Now().Add(-30 * time.Minute)
		for ip, b := range l.buckets {
			if b.lastSeen.Before(cutoff) {
				delete(l.buckets, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[ip]
	if !ok {
		l.buckets[ip] = &bucket{tokens: l.capacity - 1, lastSeen: now}
		return true
	}
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastSeen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// RateLimit enforces a per-IP token bucket. rate = tokens/sec, burst = max bucket size.
func RateLimit(rate float64, burst float64) gin.HandlerFunc {
	l := newIPLimiter(rate, burst)
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			utils.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
