package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache provides Redis caching functionality
type Cache struct {
	redisClient *redis.Client
	defaultTTL  time.Duration
}

// NewCache creates a new cache instance
func NewCache(redisClient *redis.Client) *Cache {
	return &Cache{
		redisClient: redisClient,
		defaultTTL:  1 * time.Hour,
	}
}

// Set stores a value in cache
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	return c.redisClient.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Delete removes a key from cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.redisClient.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.redisClient.Exists(ctx, key).Result()
	return n > 0, err
}

// Increment increments a counter
func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.redisClient.Incr(ctx, key).Result()
}

// Decrement decrements a counter
func (c *Cache) Decrement(ctx context.Context, key string) (int64, error) {
	return c.redisClient.Decr(ctx, key).Result()
}

// SetNX sets a value only if key doesn't exist
func (c *Cache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return c.redisClient.SetNX(ctx, key, data, ttl).Result()
}

// GetMulti retrieves multiple keys at once
func (c *Cache) GetMulti(ctx context.Context, keys []string) ([]interface{}, error) {
	results, err := c.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var values []interface{}
	for _, result := range results {
		if result == nil {
			values = append(values, nil)
			continue
		}

		str, ok := result.(string)
		if !ok {
			values = append(values, nil)
			continue
		}

		values = append(values, str)
	}

	return values, nil
}

// SetMulti sets multiple keys at once
func (c *Cache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	pipe := c.redisClient.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			continue
		}

		if ttl == 0 {
			ttl = c.defaultTTL
		}

		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Clear deletes all keys matching a pattern
func (c *Cache) Clear(ctx context.Context, pattern string) error {
	keys, err := c.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.redisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// GetStats returns cache statistics
func (c *Cache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.redisClient.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["info"] = info

	// Get key count
	keys, _ := c.redisClient.Keys(ctx, "crawl:*").Result()
	stats["total_keys"] = len(keys)

	return stats, nil
}

// URLFilter stores URL filters for crawling
type URLFilter struct {
	Domain        string   `json:"domain"`
	Include       []string `json:"include"`
	Exclude       []string `json:"exclude"`
	MaxDepth      int      `json:"max_depth"`
	AllowExternal bool     `json:"allow_external"`
}

// SaveURLFilter saves a URL filter configuration
func (c *Cache) SaveURLFilter(ctx context.Context, jobID string, filter URLFilter) error {
	key := fmt.Sprintf("crawl:filter:%s", jobID)
	return c.Set(ctx, key, filter, 24*time.Hour)
}

// GetURLFilter retrieves a URL filter configuration
func (c *Cache) GetURLFilter(ctx context.Context, jobID string) (*URLFilter, error) {
	key := fmt.Sprintf("crawl:filter:%s", jobID)
	var filter URLFilter
	err := c.Get(ctx, key, &filter)
	if err != nil {
		return nil, err
	}
	return &filter, nil
}
