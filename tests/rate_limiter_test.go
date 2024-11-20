package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/marcosocram/fullcycle-rate-limiter/limiter"
	"github.com/marcosocram/fullcycle-rate-limiter/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Função auxiliar para configurar o Rate Limiter com miniredis
func setupRateLimiterWithMiniredis(ipLimit, tokenLimit int, banDuration time.Duration) (limiter.RateLimiter, *miniredis.Miniredis) {
	// Inicia um servidor Redis em memória
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	// Retorna o RateLimiter configurado e a instância do miniredis
	return limiter.NewRedisLimiter(mr.Addr(), ipLimit, tokenLimit, banDuration), mr
}

// Função auxiliar para criar o roteador com middleware de Rate Limiting
func createTestRouter(rateLimiter limiter.RateLimiter, ipLimit, tokenLimit int) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.RateLimitMiddleware(rateLimiter, ipLimit, tokenLimit))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Bem-vindo ao servidor com Rate Limiting!"))
	})
	return router
}

func TestRateLimiter_ByIP_LimitNotExceeded(t *testing.T) {
	rateLimiter, mr := setupRateLimiterWithMiniredis(5, 10, 1*time.Minute)
	defer mr.Close()

	router := createTestRouter(rateLimiter, 5, 10)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Bem-vindo")
	}
}

func TestRateLimiter_ByIP_LimitExceeded(t *testing.T) {
	rateLimiter, mr := setupRateLimiterWithMiniredis(5, 10, 1*time.Minute)
	defer mr.Close()

	router := createTestRouter(rateLimiter, 5, 10)

	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if i < 5 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.Contains(t, w.Body.String(), "you have reached the maximum number")
		}
	}
}

func TestRateLimiter_ByToken_LimitNotExceeded(t *testing.T) {
	rateLimiter, mr := setupRateLimiterWithMiniredis(5, 10, 1*time.Minute)
	defer mr.Close()

	router := createTestRouter(rateLimiter, 5, 10)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "token123")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Bem-vindo")
	}
}

func TestRateLimiter_ByToken_LimitExceeded(t *testing.T) {
	rateLimiter, mr := setupRateLimiterWithMiniredis(5, 10, 1*time.Minute)
	defer mr.Close()

	router := createTestRouter(rateLimiter, 5, 10)

	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "token123")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if i < 10 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.Contains(t, w.Body.String(), "you have reached the maximum number")
		}
	}
}

func TestRateLimiter_ByIP_BlockDuration(t *testing.T) {
	rateLimiter, mr := setupRateLimiterWithMiniredis(1, 10, 3*time.Second)
	defer mr.Close()

	router := createTestRouter(rateLimiter, 1, 10)

	// Primeira requisição deve passar
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Segunda requisição deve ser bloqueada
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Aguarda o tempo de bloqueio
	time.Sleep(3 * time.Second)
	mr.FlushDB()

	// Requisição após o tempo de bloqueio deve passar
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
