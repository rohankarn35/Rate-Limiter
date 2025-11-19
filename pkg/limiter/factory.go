package limiter

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/pkg/config"
	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
)

// NewManagerFromConfig converts config policies into a Manager backed by the supplied store.
func NewManagerFromConfig(policies []config.Policy, store storage.Storage) (*Manager, error) {
	var parsed []*Policy
	for _, policyConfig := range policies {
		if policyConfig.Algorithm.Type == "" {
			return nil, fmt.Errorf("policy %s: algorithm.type is required", policyConfig.Name)
		}
		keyFunc, err := keyFuncFromConfig(policyConfig.Identity)
		if err != nil {
			return nil, fmt.Errorf("policy %s: %w", policyConfig.Name, err)
		}

		instance, err := limiterFromConfig(policyConfig.Algorithm, store, policyConfig.Name)
		if err != nil {
			return nil, fmt.Errorf("policy %s: %w", policyConfig.Name, err)
		}

		parsed = append(parsed, &Policy{
			Name:    policyConfig.Name,
			Routes:  policyConfig.Routes,
			Methods: policyConfig.Methods,
			Limiter: instance,
			KeyFunc: keyFunc,
		})
	}

	return NewManager(parsed), nil
}

func limiterFromConfig(cfg config.AlgorithmConfig, store storage.Storage, prefix string) (Limiter, error) {
	switch AlgorithmType(cfg.Type) {
	case AlgorithmTokenBucket:
		if cfg.Burst <= 0 {
			cfg.Burst = cfg.Limit
		}
		if cfg.Interval.Duration() <= 0 {
			cfg.Interval = config.Duration(time.Second)
		}
		if cfg.RefillRate <= 0 {
			cfg.RefillRate = cfg.Limit
		}
		return NewTokenBucketLimiter(store, cfg.Burst, cfg.RefillRate, cfg.Interval.Duration(), "tb:"+prefix), nil
	case AlgorithmLeakyBucket:
		if cfg.Limit <= 0 {
			return nil, fmt.Errorf("limit must be > 0")
		}
		if cfg.LeakRate <= 0 {
			return nil, fmt.Errorf("leak_rate must be > 0")
		}
		return NewLeakyBucketLimiter(store, cfg.Limit, cfg.LeakRate, "lb:"+prefix), nil
	case AlgorithmSlidingWindow:
		if cfg.Limit <= 0 {
			return nil, fmt.Errorf("limit must be > 0")
		}
		if cfg.Window.Duration() <= 0 {
			return nil, fmt.Errorf("window must be > 0")
		}
		return NewSlidingWindowLimiter(store, cfg.Limit, cfg.Window.Duration(), "sw:"+prefix), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", cfg.Type)
	}
}

func keyFuncFromConfig(identity config.IdentityConfig) (KeyFunc, error) {
	switch strings.ToLower(identity.Type) {
	case "", "ip":
		return func(r *http.Request) string {
			return extractIP(r)
		}, nil
	case "header", "api_key":
		headerName := identity.Key
		if headerName == "" {
			headerName = "X-API-Key"
		}
		fallback := strings.ToLower(identity.Fallback)
		return func(r *http.Request) string {
			value := r.Header.Get(headerName)
			if value == "" && fallback == "ip" {
				value = extractIP(r)
			}
			return value
		}, nil
	case "query":
		param := identity.Key
		if param == "" {
			return nil, fmt.Errorf("identity.key is required for query identity")
		}
		return func(r *http.Request) string {
			value := r.URL.Query().Get(param)
			if value == "" && strings.ToLower(identity.Fallback) == "ip" {
				value = extractIP(r)
			}
			return value
		}, nil
	default:
		return nil, fmt.Errorf("unsupported identity type %s", identity.Type)
	}
}

func extractIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
