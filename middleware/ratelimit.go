package middleware

import (
	"context"
	"github.com/marcosocram/fullcycle-rate-limiter/limiter"
	"net/http"
	"strings"
)

func RateLimitMiddleware(l limiter.RateLimiter, ipLimit, tokenLimit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()
			ip := strings.Split(r.RemoteAddr, ":")[0]
			token := r.Header.Get("API_KEY")

			key := ip
			limit := ipLimit
			if token != "" {
				key = "token:" + token
				limit = tokenLimit
			}

			blocked, _ := l.IsBlocked(ctx, key)
			if blocked {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			allowed, _ := l.Allow(ctx, key, limit)
			if !allowed {
				l.Block(ctx, key)
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
