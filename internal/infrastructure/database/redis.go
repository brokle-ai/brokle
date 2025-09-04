package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"brokle/internal/config"
)

// RedisDB represents Redis database connection
type RedisDB struct {
	Client *redis.Client
	config *config.Config
	logger *logrus.Logger
}

// NewRedisDB creates a new Redis database connection
func NewRedisDB(cfg *config.Config, logger *logrus.Logger) (*RedisDB, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configure connection settings
	opt.MaxRetries = 3
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	opt.PoolSize = 10
	opt.PoolTimeout = 30 * time.Second
	opt.MaxIdleConns = 5
	// opt.IdleCheckFrequency = 5 * time.Minute // Field not available

	// Create Redis client
	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Connected to Redis database")

	return &RedisDB{
		Client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (r *RedisDB) Close() error {
	r.logger.Info("Closing Redis connection")
	return r.Client.Close()
}

// Health checks Redis health
func (r *RedisDB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err()
}

// Set sets a key-value pair with optional expiration
func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Delete deletes keys
func (r *RedisDB) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func (r *RedisDB) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (r *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// HSet sets hash field
func (r *RedisDB) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet gets hash field
func (r *RedisDB) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields
func (r *RedisDB) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields
func (r *RedisDB) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// ZAdd adds members to sorted set
func (r *RedisDB) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.Client.ZAdd(ctx, key, members...).Err()
}

// ZRange gets sorted set members by range
func (r *RedisDB) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores gets sorted set members with scores
func (r *RedisDB) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return r.Client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// Increment increments a key
func (r *RedisDB) Increment(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// IncrementBy increments a key by value
func (r *RedisDB) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.IncrBy(ctx, key, value).Result()
}

// GetStats returns Redis connection pool statistics
func (r *RedisDB) GetStats() *redis.PoolStats {
	return r.Client.PoolStats()
}