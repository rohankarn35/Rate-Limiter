package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPServer wraps the stdlib server with convenience methods.
type HTTPServer struct {
	server *http.Server
}

// HTTPConfig describes the listener settings.
type HTTPConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewHTTPServer builds an HTTP server with the provided handler.
func NewHTTPServer(cfg HTTPConfig, handler http.Handler) *HTTPServer {
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return &HTTPServer{server: srv}
}

// Start begins listening in a separate goroutine.
func (s *HTTPServer) Start() error {
	if s.server == nil {
		return fmt.Errorf("http server not configured")
	}
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts the server down.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
