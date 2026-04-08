package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const scholarshipListKey = "scholarships:list"
const scholarshipDetailKeyPrefix = "scholarships:detail:"

type ScholarshipCache interface {
	GetScholarshipList(ctx context.Context) ([]byte, bool, error)
	SetScholarshipList(ctx context.Context, value []byte) error
	GetScholarshipDetail(ctx context.Context, scholarshipID string) ([]byte, bool, error)
	SetScholarshipDetail(ctx context.Context, scholarshipID string, value []byte) error
	InvalidateScholarship(ctx context.Context, scholarshipID string) error
	InvalidateScholarshipList(ctx context.Context) error
}

type NoopScholarshipCache struct{}

func NewNoopScholarshipCache() *NoopScholarshipCache {
	return &NoopScholarshipCache{}
}

func (c *NoopScholarshipCache) GetScholarshipList(context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (c *NoopScholarshipCache) SetScholarshipList(context.Context, []byte) error {
	return nil
}

func (c *NoopScholarshipCache) GetScholarshipDetail(context.Context, string) ([]byte, bool, error) {
	return nil, false, nil
}

func (c *NoopScholarshipCache) SetScholarshipDetail(context.Context, string, []byte) error {
	return nil
}

func (c *NoopScholarshipCache) InvalidateScholarship(context.Context, string) error {
	return nil
}

func (c *NoopScholarshipCache) InvalidateScholarshipList(context.Context) error {
	return nil
}

type RedisScholarshipCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisScholarshipCache(client *redis.Client, ttl time.Duration) *RedisScholarshipCache {
	return &RedisScholarshipCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisScholarshipCache) GetScholarshipList(ctx context.Context) ([]byte, bool, error) {
	return c.get(ctx, scholarshipListKey)
}

func (c *RedisScholarshipCache) SetScholarshipList(ctx context.Context, value []byte) error {
	return c.set(ctx, scholarshipListKey, value)
}

func (c *RedisScholarshipCache) GetScholarshipDetail(ctx context.Context, scholarshipID string) ([]byte, bool, error) {
	return c.get(ctx, scholarshipDetailKey(scholarshipID))
}

func (c *RedisScholarshipCache) SetScholarshipDetail(ctx context.Context, scholarshipID string, value []byte) error {
	return c.set(ctx, scholarshipDetailKey(scholarshipID), value)
}

func (c *RedisScholarshipCache) InvalidateScholarship(ctx context.Context, scholarshipID string) error {
	return c.client.Del(ctx, scholarshipDetailKey(scholarshipID)).Err()
}

func (c *RedisScholarshipCache) InvalidateScholarshipList(ctx context.Context) error {
	return c.client.Del(ctx, scholarshipListKey).Err()
}

func (c *RedisScholarshipCache) get(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err == nil {
		return value, true, nil
	}

	if err == redis.Nil {
		return nil, false, nil
	}

	return nil, false, fmt.Errorf("redis get %s: %w", key, err)
}

func (c *RedisScholarshipCache) set(ctx context.Context, key string, value []byte) error {
	if err := c.client.Set(ctx, key, value, c.ttl).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}

	return nil
}

func scholarshipDetailKey(scholarshipID string) string {
	return scholarshipDetailKeyPrefix + scholarshipID
}
