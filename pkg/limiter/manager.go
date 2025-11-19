package limiter

import (
	"context"
	"net/http"
	"path"
	"strings"
)

// KeyFunc extracts a logical identity for a request.
type KeyFunc func(r *http.Request) string

// Policy wraps a limiter with routing metadata.
type Policy struct {
	Name    string
	Routes  []string
	Methods []string
	Limiter Limiter
	KeyFunc KeyFunc
}

// Manager selects the proper policy per request.
type Manager struct {
	policies []*Policy
}

// NewManager builds a Manager from policies (evaluated in-order).
func NewManager(policies []*Policy) *Manager {
	return &Manager{policies: policies}
}

// Allow evaluates a request against configured policies.
func (m *Manager) Allow(ctx context.Context, r *http.Request) (Result, string, bool, error) {
	for _, policy := range m.policies {
		if !policy.matches(r) {
			continue
		}
		if policy.KeyFunc == nil {
			continue
		}
		key := policy.KeyFunc(r)
		if key == "" {
			continue
		}
		result, err := policy.Limiter.Allow(ctx, key)
		return result, policy.Name, true, err
	}

	return Result{Allowed: true}, "", false, nil
}

func (p *Policy) matches(r *http.Request) bool {
	if len(p.Methods) > 0 {
		methodMatch := false
		for _, method := range p.Methods {
			if strings.EqualFold(method, r.Method) {
				methodMatch = true
				break
			}
		}
		if !methodMatch {
			return false
		}
	}

	if len(p.Routes) == 0 {
		return true
	}

	for _, pattern := range p.Routes {
		if matchRoute(pattern, r.URL.Path) {
			return true
		}
	}
	return false
}

func matchRoute(pattern, actual string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(actual, strings.TrimSuffix(pattern, "*"))
	}
	ok, err := path.Match(pattern, actual)
	if err != nil {
		return actual == pattern
	}
	return ok
}
