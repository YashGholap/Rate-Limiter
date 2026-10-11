package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/YashGholap/rate-limiter/internal/limiter"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type Handler struct {
	limiter RateLimiter
}

func NewHandler(l RateLimiter) *Handler {
	return &Handler{
		limiter: l,
	}
}

func (h *Handler) Resource(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")

	if apiKey == "" {
		http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
		return
	}

	allowed, err := h.limiter.Allow(r.Context(), apiKey)
	if err != nil {
		http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
		return
	}

	if !allowed {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "rate limit exceeded",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "request allowed",
	})
}

// compile-time check that the production limiter satisfies our interface
var _ RateLimiter = (limiter.Limiter)(nil)
