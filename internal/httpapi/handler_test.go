package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeLimiter struct {
	allowed bool
	err     error
}

func (f *fakeLimiter) Allow(
	ctx context.Context,
	key string,
) (bool, error) {
	return f.allowed, f.err
}

func TestHandlerResource(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		allowed        bool
		limiterErr     error
		expectedStatus int
	}{
		{
			name:           "allowed request",
			apiKey:         "user-123",
			allowed:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			allowed:        true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "rate limit exceeded",
			apiKey:         "user-123",
			allowed:        false,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "limiter error",
			apiKey:         "user-123",
			limiterErr:     errors.New("Redis unavailable"),
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := &fakeLimiter{
				allowed: tt.allowed,
				err:     tt.limiterErr,
			}

			handler := NewHandler(limiter)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/resource",
				nil,
			)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			recorder := httptest.NewRecorder()

			handler.Resource(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d; body: %s",
					tt.expectedStatus,
					recorder.Code,
					recorder.Body.String(),
				)
			}
		})
	}
}
