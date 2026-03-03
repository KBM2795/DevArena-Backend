package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// --------------------------------------------------------------------------
// visitor tracks a per-IP rate limiter and when it was last seen.
// --------------------------------------------------------------------------
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// --------------------------------------------------------------------------
// ipRateLimiter holds all visitors and the config for new ones.
// --------------------------------------------------------------------------
type ipRateLimiter struct {
	visitors sync.Map   // map[string]*visitor — IP address → visitor
	rps      rate.Limit // tokens added per second (e.g. 10)
	burst    int        // max tokens in the bucket (e.g. 20)
}

// newIPRateLimiter creates a limiter and starts a background cleanup goroutine.
func newIPRateLimiter(rps float64, burst int) *ipRateLimiter {
	rl := &ipRateLimiter{
		rps:   rate.Limit(rps),
		burst: burst,
	}

	// Background cleanup: every 3 minutes, remove IPs not seen for 3+ minutes.
	// This prevents the map from growing forever.
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			rl.visitors.Range(func(key, value any) bool {
				v := value.(*visitor)
				if time.Since(v.lastSeen) > 3*time.Minute {
					rl.visitors.Delete(key)
				}
				return true
			})
		}
	}()

	return rl
}

// getVisitor returns the rate limiter for an IP, creating one if it doesn't exist.
func (rl *ipRateLimiter) getVisitor(ip string) *rate.Limiter {
	// Fast path: visitor already exists
	if v, exists := rl.visitors.Load(ip); exists {
		vis := v.(*visitor)
		vis.lastSeen = time.Now()
		return vis.limiter
	}

	// Slow path: create a new limiter for this IP
	limiter := rate.NewLimiter(rl.rps, rl.burst)
	rl.visitors.Store(ip, &visitor{
		limiter:  limiter,
		lastSeen: time.Now(),
	})
	return limiter
}

// --------------------------------------------------------------------------
// Gin Middleware Functions
// --------------------------------------------------------------------------

// RateLimiter returns a Gin middleware that limits requests per IP.
//
//   - rps:   how many requests per second each IP is allowed (steady state)
//   - burst: how many requests can be made at once before throttling kicks in
//
// Example: RateLimiter(10, 20) → 10 req/s sustained, burst of 20.
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	rl := newIPRateLimiter(rps, burst)

	log.Printf("Rate limiter enabled: %.0f req/s, burst %d", rps, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP() // handles X-Forwarded-For behind proxies

		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			// No tokens left in the bucket → reject
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Please slow down.",
				"retry_after": "1s",
			})
			return
		}

		// Token consumed → allow request through
		c.Next()
	}
}

// StrictRateLimiter is the same as RateLimiter but intended for expensive
// endpoints (e.g. submission evaluation). Use lower values.
//
// Example: StrictRateLimiter(1, 3) → 1 req/s sustained, burst of 3.
func StrictRateLimiter(rps float64, burst int) gin.HandlerFunc {
	rl := newIPRateLimiter(rps, burst)

	log.Printf("Strict rate limiter enabled: %.0f req/s, burst %d", rps, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Submission rate limit exceeded. Please wait before trying again.",
				"retry_after": "1s",
			})
			return
		}

		c.Next()
	}
}
