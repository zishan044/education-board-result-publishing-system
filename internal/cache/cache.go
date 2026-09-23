package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb redis.Cmdable
}

func New(rdb redis.Cmdable) *Cache { return &Cache{rdb: rdb} }

func (c *Cache) IsIPBlocked(ctx context.Context, ip string) (bool, error) {
	n, err := c.rdb.Exists(ctx, "blocked:"+ip).Result()
	return n > 0, err
}

func (c *Cache) BlockIP(ctx context.Context, ip string, ttl time.Duration) error {
	return c.rdb.Set(ctx, "blocked:"+ip, "1", ttl).Err()
}

func (c *Cache) SetStats(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, "stats:"+key, b, ttl).Err()
}

func (c *Cache) GetStats(ctx context.Context, key string, dest any) (bool, error) {
	b, err := c.rdb.Get(ctx, "stats:"+key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(b, dest)
}