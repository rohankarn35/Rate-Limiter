package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/config"
	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

func TestManagerMatchesRoutesAndMethods(t *testing.T) {
	store := storage.NewMemoryStorage()
	cfg := []config.Policy{{
		Name:    "per-ip",
		Routes:  []string{"/api/v1/*"},
		Methods: []string{http.MethodGet},
		Identity: config.IdentityConfig{
			Type: "ip",
		},
		Algorithm: config.AlgorithmConfig{
			Type:     string(limiter.AlgorithmTokenBucket),
			Limit:    1,
			Burst:    1,
			Interval: config.Duration(100 * time.Millisecond),
		},
	}}

	manager, err := limiter.NewManagerFromConfig(cfg, store)
	if err != nil {
		t.Fatalf("failed to build manager: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments", nil)
	res, _, matched, err := manager.Allow(context.Background(), req)
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if !matched || !res.Allowed {
		t.Fatalf("expected request to match and be allowed")
	}

	res, _, matched, err = manager.Allow(context.Background(), req)
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if !matched || res.Allowed {
		t.Fatalf("second request should be rate limited")
	}

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/payments", nil)
	res, _, matched, err = manager.Allow(context.Background(), postReq)
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if matched {
		t.Fatalf("POST should not match GET-only policy")
	}
}
