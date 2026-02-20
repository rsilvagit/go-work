package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rsilvagit/go-work/internal/model"
)

// Cache provides Redis-backed caching for scraped job results.
type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

// New connects to Redis at the given URL and returns a Cache.
// URL format: redis://localhost:6379
func New(redisURL string, ttl time.Duration) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache: invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: redis ping failed: %w", err)
	}

	return &Cache{client: client, ttl: ttl}, nil
}

// Get retrieves cached jobs for the given scraper/query/location combination.
// Returns the jobs and true if a valid cache entry exists, or nil and false otherwise.
func (c *Cache) Get(ctx context.Context, scraper, query, location string) ([]model.Job, bool) {
	key := buildKey(scraper, query, location)

	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var jobs []model.Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, false
	}

	return jobs, true
}

// Set stores jobs in the cache with the configured TTL.
func (c *Cache) Set(ctx context.Context, scraper, query, location string, jobs []model.Job) error {
	key := buildKey(scraper, query, location)

	data, err := json.Marshal(jobs)
	if err != nil {
		return fmt.Errorf("cache: marshal error: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Close closes the Redis connection.
func (c *Cache) Close() error {
	return c.client.Close()
}

func buildKey(scraper, query, location string) string {
	raw := strings.ToLower(scraper + ":" + query + ":" + location)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("gowork:%s:%x", strings.ToLower(scraper), hash[:8])
}
