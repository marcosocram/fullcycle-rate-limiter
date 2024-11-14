package main

import (
	"github.com/marcosocram/fullcycle-rate-limiter/config"
	"github.com/marcosocram/fullcycle-rate-limiter/limiter"
	"github.com/marcosocram/fullcycle-rate-limiter/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()
	r := mux.NewRouter()

	redisLimiter := limiter.NewRedisLimiter(cfg.RedisAddr, cfg.IpRate, cfg.TokenRate, cfg.BanDuration)
	r.Use(middleware.RateLimitMiddleware(redisLimiter, cfg.IpRate, cfg.TokenRate))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Bem-vindo ao servidor com Rate Limiting!\n\n"))
	})

	log.Printf("Servidor iniciado na porta %s", cfg.ServerPort)
	http.ListenAndServe(":"+cfg.ServerPort, r)
}
