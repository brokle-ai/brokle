package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
	"brokle/pkg/response"
)

// RateLimitMiddleware handles Redis-based rate limiting
type RateLimitMiddleware struct {
	redis  *redis.Client
	config *config.AuthConfig
	logger *logrus.Logger
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(
	redis *redis.Client,
	config *config.AuthConfig,
	logger *logrus.Logger,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redis:  redis,
		config: config,
		logger: logger,
	}
}

// RateLimitByIP implements IP-based rate limiting using Redis sliding window
func (m *RateLimitMiddleware) RateLimitByIP() gin.HandlerFunc {
	if !m.config.RateLimitEnabled {
		// Rate limiting disabled, pass through
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:ip:%s", clientIP)
		
		allowed, err := m.checkRateLimit(c.Request.Context(), key, m.config.RateLimitPerIP, m.config.RateLimitWindow)
		if err != nil {
			m.logger.WithError(err).WithField("ip", clientIP).Error("Rate limit check failed")
			// On error, allow request to continue
			c.Next()
			return
		}

		if !allowed {
			m.logger.WithField("ip", clientIP).Warn("Rate limit exceeded for IP")
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUser implements user-based rate limiting using Redis sliding window
func (m *RateLimitMiddleware) RateLimitByUser() gin.HandlerFunc {
	if !m.config.RateLimitEnabled {
		// Rate limiting disabled, pass through
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Get user ID from auth context
		userID, exists := c.Get(UserIDKey)
		if !exists {
			// No user context, skip user-based rate limiting
			c.Next()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		key := fmt.Sprintf("rate_limit:user:%s", userIDStr)
		
		allowed, err := m.checkRateLimit(c.Request.Context(), key, m.config.RateLimitPerUser, m.config.RateLimitWindow)
		if err != nil {
			m.logger.WithError(err).WithField("user_id", userIDStr).Error("User rate limit check failed")
			// On error, allow request to continue
			c.Next()
			return
		}

		if !allowed {
			m.logger.WithField("user_id", userIDStr).Warn("Rate limit exceeded for user")
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByAPIKey implements API key-based rate limiting
func (m *RateLimitMiddleware) RateLimitByAPIKey() gin.HandlerFunc {
	if !m.config.RateLimitEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Get auth context (set by RequireKeyPair middleware)
		authCtx, exists := GetAuthContext(c)
		if !exists || authCtx == nil || authCtx.KeyPairID == nil {
			c.Next()
			return
		}

		// Use key pair ID for rate limiting
		keyPairStr := authCtx.KeyPairID.String()
		key := fmt.Sprintf("rate_limit:keypair:%s", keyPairStr)

		// Key pairs typically have higher limits
		keyPairLimit := m.config.RateLimitPerUser * 5 // 5x higher limit for key pairs

		allowed, err := m.checkRateLimit(c.Request.Context(), key, keyPairLimit, m.config.RateLimitWindow)
		if err != nil {
			m.logger.WithError(err).WithField("key_pair_id", keyPairStr).Error("Key pair rate limit check failed")
			c.Next()
			return
		}

		if !allowed {
			m.logger.WithField("key_pair_id", keyPairStr).Warn("Rate limit exceeded for key pair")
			response.TooManyRequests(c, "Key pair rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit implements sliding window rate limiting using Redis
func (m *RateLimitMiddleware) checkRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	pipe := m.redis.TxPipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.Unix(), 10))
	
	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)
	
	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.Unix()),
		Member: fmt.Sprintf("%d-%d", now.Unix(), now.Nanosecond()),
	})
	
	// Set expiry for the key
	pipe.Expire(ctx, key, window)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis pipeline failed: %w", err)
	}

	// Check if limit exceeded
	count := countCmd.Val()
	return count < int64(limit), nil
}

// GetRemainingRequests returns the number of remaining requests for a key
func (m *RateLimitMiddleware) GetRemainingRequests(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	if !m.config.RateLimitEnabled {
		return limit, nil
	}

	now := time.Now()
	windowStart := now.Add(-window)

	// Remove expired entries and count current
	pipe := m.redis.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.Unix(), 10))
	countCmd := pipe.ZCard(ctx, key)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get remaining requests: %w", err)
	}

	current := int(countCmd.Val())
	remaining := limit - current
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// SetRateLimitHeaders sets rate limit headers in the response
func (m *RateLimitMiddleware) SetRateLimitHeaders(clientIP, userID string) gin.HandlerFunc {
	if !m.config.RateLimitEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Get IP-based rate limit info
		ipKey := fmt.Sprintf("rate_limit:ip:%s", clientIP)
		ipRemaining, _ := m.GetRemainingRequests(c.Request.Context(), ipKey, m.config.RateLimitPerIP, m.config.RateLimitWindow)
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(m.config.RateLimitPerIP))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(ipRemaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(m.config.RateLimitWindow).Unix(), 10))

		// Add user-specific headers if user is authenticated
		if userID != "" {
			userKey := fmt.Sprintf("rate_limit:user:%s", userID)
			userRemaining, _ := m.GetRemainingRequests(c.Request.Context(), userKey, m.config.RateLimitPerUser, m.config.RateLimitWindow)
			c.Header("X-RateLimit-User-Limit", strconv.Itoa(m.config.RateLimitPerUser))
			c.Header("X-RateLimit-User-Remaining", strconv.Itoa(userRemaining))
		}

		c.Next()
	}
}