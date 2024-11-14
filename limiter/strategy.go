package limiter

import "context"

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int) (bool, error)
	Block(ctx context.Context, key string) error
	IsBlocked(ctx context.Context, key string) (bool, error)
}
