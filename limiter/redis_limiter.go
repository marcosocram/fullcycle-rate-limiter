package limiter

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisLimiter struct {
	client            *redis.Client
	ipRate, tokenRate int
	banDuration       time.Duration
}

func NewRedisLimiter(addr string, ipRate, tokenRate int, banDuration time.Duration) *RedisLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	client.FlushAll(context.Background())
	return &RedisLimiter{client, ipRate, tokenRate, banDuration}
}

func (r *RedisLimiter) Allow(ctx context.Context, key string, limit int) (bool, error) {
	current, err := r.client.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, err
	}

	if current >= limit {
		return false, nil
	}

	pipe := r.client.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Second)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisLimiter) Block(ctx context.Context, key string) error {
	return r.client.Set(ctx, fmt.Sprintf("ban:%s", key), "1", r.banDuration).Err()
}

func (r *RedisLimiter) IsBlocked(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, fmt.Sprintf("ban:%s", key)).Result()
	return exists > 0, err
}
