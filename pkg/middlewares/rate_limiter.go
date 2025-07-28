package middlewares

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
)

type RateLimitConfig struct {
	RequestsPerSecond float64                 // Number of requests allowed per second
	BurstSize         int                     // Maximum burst size
	KeyGenerator      func(*fiber.Ctx) string // Function to generate unique keys
}

type RateLimitInfo struct {
	Allowed   bool
	Remaining int
	ResetTime time.Time
}

type tokenBucket struct {
	mu          sync.RWMutex
	buckets     map[string]*bucket
	refillRate  float64 // tokens per second
	capacity    int     // max tokens
	lastCleanup time.Time
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

func RateLimit(config RateLimitConfig) fiber.Handler {
	limiter := newTokenBucket(config.RequestsPerSecond, config.BurstSize)

	return func(c *fiber.Ctx) error {
		key := config.KeyGenerator(c)
		rateLimitInfo := limiter.checkRateLimit(key)

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", config.RequestsPerSecond))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rateLimitInfo.Remaining))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", rateLimitInfo.ResetTime.Unix()))

		if !rateLimitInfo.Allowed {
			retryAfter := int(time.Until(rateLimitInfo.ResetTime).Seconds()) + 1
			c.Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			appErr := exception.TooManyRequests(errors.CodeRateLimitExceeded, errors.ErrTooManyRequests).
				WithDetails("Rate limit exceeded. Please slow down and try again later").
				WithMetadata("requests_per_second", config.RequestsPerSecond).
				WithMetadata("burst_size", config.BurstSize).
				WithMetadata("path", c.Path()).
				WithMetadata("retry_after_seconds", retryAfter)

			return response.Error(c, appErr)
		}

		return c.Next()
	}
}

func DefaultKeyGenerator(prefix string) func(*fiber.Ctx) string {
	return func(c *fiber.Ctx) string {
		return fmt.Sprintf("%s:%s", prefix, c.IP())
	}
}

func UserKeyGenerator(prefix string) func(*fiber.Ctx) string {
	return func(c *fiber.Ctx) string {
		userID := c.Get("X-User-ID", "anonymous")
		return fmt.Sprintf("%s:%s:%s", prefix, userID, c.IP())
	}
}

func newTokenBucket(requestsPerSecond float64, burstSize int) *tokenBucket {
	tb := &tokenBucket{
		buckets:     make(map[string]*bucket),
		refillRate:  requestsPerSecond,
		capacity:    burstSize,
		lastCleanup: time.Now(),
	}

	// Cleanup old buckets every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			tb.cleanup()
		}
	}()

	return tb
}

func (tb *tokenBucket) checkRateLimit(key string) RateLimitInfo {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	b, exists := tb.buckets[key]

	if !exists {
		// New bucket starts full
		tb.buckets[key] = &bucket{
			tokens:     float64(tb.capacity) - 1, // consume 1 token immediately
			lastRefill: now,
			lastSeen:   now,
		}
		return RateLimitInfo{
			Allowed:   true,
			Remaining: tb.capacity - 1,
			ResetTime: now.Add(time.Second), // Next refill in 1 second
		}
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(b.lastRefill).Seconds()
	tokensToAdd := elapsed * tb.refillRate

	// Refill bucket
	b.tokens = utils.Min(b.tokens+tokensToAdd, float64(tb.capacity))
	b.lastRefill = now
	b.lastSeen = now

	// Calculate when next token will be available
	timeToNextToken := time.Duration(0)
	if b.tokens < 1 {
		tokensNeeded := 1 - b.tokens
		timeToNextToken = time.Duration(tokensNeeded/tb.refillRate) * time.Second
	}

	// Check to consume a token
	if b.tokens >= 1 {
		b.tokens--
		return RateLimitInfo{
			Allowed:   true,
			Remaining: int(b.tokens),
			ResetTime: now.Add(time.Second),
		}
	}

	return RateLimitInfo{
		Allowed:   false,
		Remaining: 0,
		ResetTime: now.Add(timeToNextToken),
	}
}

func (tb *tokenBucket) cleanup() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	cutoff := time.Now().Add(-time.Hour) // Remove buckets not seen for 1 hour
	for key, b := range tb.buckets {
		if b.lastSeen.Before(cutoff) {
			delete(tb.buckets, key)
		}
	}
}
