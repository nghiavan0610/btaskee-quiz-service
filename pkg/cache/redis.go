package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type CacheKeyOption struct {
	Prefix    string
	UniqueKey string
	Suffix    string
	Module    string
	Value     interface{}
	TTL       time.Duration
}

type (
	Cache interface {
		SetGlobal(ctx context.Context, key string, value interface{}, ttl time.Duration) *exception.AppError
		Set(ctx context.Context, opt CacheKeyOption) *exception.AppError
		Get(ctx context.Context, opt CacheKeyOption) (string, *exception.AppError)
		GetObject(ctx context.Context, opt CacheKeyOption, dest interface{}) *exception.AppError
		Exists(ctx context.Context, opt CacheKeyOption) (bool, *exception.AppError)
		Clear(ctx context.Context, password string, systemPassword string) *exception.AppError
		Del(ctx context.Context, opt CacheKeyOption) *exception.AppError
		Close()
	}

	cache struct {
		client *redis.Client
		logger *logger.Logger
	}
)

var (
	cacheOnce     sync.Once
	cacheInstance Cache
	cacheError    error
)

func ProvideCache(cfg *config.Config, logger *logger.Logger) (Cache, error) {
	cacheOnce.Do(func() {
		cacheInstance, cacheError = initializeRedisCache(cfg, logger)
	})

	return cacheInstance, cacheError
}

func ProvideRedisClient(cfg *config.Config, logger *logger.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.CacheDB,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("[REDIS] failed to initialize Redis client: %w", err)
	}

	return rdb, nil
}

func initializeRedisCache(cfg *config.Config, logger *logger.Logger) (Cache, error) {
	rdb, err := ProvideRedisClient(cfg, logger)
	if err != nil {
		return nil, err
	}

	cacheImpl := &cache{
		client: rdb,
		logger: logger,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cacheImpl.client.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("[REDIS] failed to connect to Redis: %w", err)
	}

	logger.Info("[REDIS] Connected to Redis Cache DB",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port,
		"db", cfg.Redis.CacheDB,
	)

	return cacheImpl, nil
}

func buildCacheKey(opt CacheKeyOption) string {
	key := opt.UniqueKey
	if opt.Module != "" {
		key = opt.Module + ":" + key
	}
	if opt.Prefix != "" {
		key = opt.Prefix + ":" + key
	}
	if opt.Suffix != "" {
		key = key + ":" + opt.Suffix
	}
	return key
}

func (c *cache) SetGlobal(ctx context.Context, key string, value interface{}, ttl time.Duration) *exception.AppError {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	// Marshal to JSON for consistency with Set method
	jsonData, err := json.Marshal(value)
	if err != nil {
		return exception.InternalError(errors.CodeInternal, errors.ErrFailedToMarshalJSON).
			WithMetadata("key", key).
			WithMetadata("operation", "json_marshal_global")
	}

	if err := c.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrFailedToSetGlobalCache).
			WithMetadata("key", key).
			WithMetadata("operation", "set_global")
	}
	return nil
}

func (c *cache) Set(ctx context.Context, opt CacheKeyOption) *exception.AppError {
	ttl := opt.TTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	cacheKey := buildCacheKey(opt)

	jsonData, err := json.Marshal(opt.Value)
	if err != nil {
		return exception.InternalError(errors.CodeInternal, errors.ErrFailedToMarshalJSON).
			WithMetadata("key", cacheKey).
			WithMetadata("operation", "json_marshal")
	}

	if err := c.client.Set(ctx, cacheKey, jsonData, ttl).Err(); err != nil {
		return exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrFailedToSetCache).
			WithMetadata("key", cacheKey).
			WithMetadata("operation", "set")
	}
	return nil
}

func (c *cache) Get(ctx context.Context, opt CacheKeyOption) (string, *exception.AppError) {
	cacheKey := buildCacheKey(opt)
	result, err := c.client.Get(ctx, cacheKey).Result()

	if err != nil {
		if err == redis.Nil {
			return "", exception.NotFound(errors.CodeCacheNotFound, errors.ErrCacheKeyNotFound).
				WithMetadata("key", cacheKey)
		}
		return "", exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrFailedToGetCache).
			WithMetadata("key", cacheKey).
			WithMetadata("operation", "get")
	}
	return result, nil
}

func (c *cache) GetObject(ctx context.Context, opt CacheKeyOption, dest interface{}) *exception.AppError {
	cacheKey := buildCacheKey(opt)
	result, err := c.client.Get(ctx, cacheKey).Result()

	if err != nil {
		if err == redis.Nil {
			return exception.NotFound(errors.CodeCacheNotFound, errors.ErrCacheKeyNotFound).
				WithMetadata("key", cacheKey)
		}
		return exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrFailedToGetCache).
			WithMetadata("key", cacheKey).
			WithMetadata("operation", "get_object")
	}

	// Unmarshal JSON to destination object
	if err := json.Unmarshal([]byte(result), dest); err != nil {
		return exception.BadRequest(errors.CodeBadRequest, "Failed to deserialize cached value").
			WithDetails("Cached value is not a valid JSON object").
			WithMetadata("key", cacheKey).
			WithMetadata("error", err.Error())
	}

	return nil
}

func (c *cache) Exists(ctx context.Context, opt CacheKeyOption) (bool, *exception.AppError) {
	cacheKey := buildCacheKey(opt)
	result, err := c.client.Exists(ctx, cacheKey).Result()

	if err != nil {
		return false, exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrCacheKeyNotFound).
			WithMetadata("key", cacheKey).
			WithMetadata("operation", "exists")
	}

	return result > 0, nil
}

func (c *cache) Clear(ctx context.Context, password string, systemPassword string) *exception.AppError {
	if systemPassword == "" || password != systemPassword {
		return exception.Forbidden(errors.CodeForbidden, "Access denied for cache clear operation").
			WithDetails("Invalid password provided")
	}

	if err := c.client.FlushDB(ctx).Err(); err != nil {
		return exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrFailedToClearCache).
			WithMetadata("operation", "clear")
	}
	return nil
}

func (c *cache) Del(ctx context.Context, opt CacheKeyOption) *exception.AppError {
	cacheKey := buildCacheKey(opt)
	if err := c.client.Del(ctx, cacheKey).Err(); err != nil {
		return exception.ServiceUnavailable(errors.CodeCacheUnavailable, errors.ErrFailedToDeleteCache).
			WithMetadata("key", cacheKey).
			WithMetadata("operation", "delete")
	}
	return nil
}

func (c *cache) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
