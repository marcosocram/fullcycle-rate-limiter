package tests

import (
	"context"
	"github.com/marcosocram/fullcycle-rate-limiter/limiter"
	"testing"
	"time"
)

func TestRedisLimiter(t *testing.T) {
	ctx := context.Background()
	rl := limiter.NewRedisLimiter("localhost:6379", 5, 10, 300*time.Second)

	for i := 0; i < 5; i++ {
		allowed, _ := rl.Allow(ctx, "test_ip", 5)
		if !allowed {
			t.Errorf("Falha ao permitir requisição %d", i+1)
		}
	}

	allowed, _ := rl.Allow(ctx, "test_ip", 5)
	if allowed {
		t.Error("Rate limit não foi aplicado corretamente")
	}
}
