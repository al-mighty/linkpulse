package repo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) CacheURL(ctx context.Context, code, url string) error {
	return r.client.Set(ctx, "link:"+code, url, 24*time.Hour).Err()
}

func (r *Redis) GetCachedURL(ctx context.Context, code string) (string, error) {
	return r.client.Get(ctx, "link:"+code).Result()
}