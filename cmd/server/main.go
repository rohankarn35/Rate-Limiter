package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rohankarn35/rate_limiter_golang/internal/api/middleware"
	"github.com/rohankarn35/rate_limiter_golang/internal/server"
	"github.com/rohankarn35/rate_limiter_golang/pkg/config"
	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
	"github.com/rohankarn35/rate_limiter_golang/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := server.WaitForSignal(context.Background())

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	store, closer, err := buildStorage(cfg.Storage)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	if closer != nil {
		defer closer()
	}

	manager, err := limiter.NewManagerFromConfig(cfg.Policies, store)
	if err != nil {
		log.Fatalf("failed to build limiter manager: %v", err)
	}

	metrics := server.NewMetrics()
	httpServer := bootstrapHTTPServer(cfg, manager, metrics)
	grpcServer := bootstrapGRPCServer(cfg, manager, metrics)

	errCh := make(chan error, 2)

	go func() {
		log.Printf("HTTP server listening on %s", cfg.Server.Address)
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		if grpcServer == nil {
			return
		}
		log.Printf("gRPC server listening on %s", cfg.Server.GRPCAddress())
		if err := grpcServer.Start(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down servers...")
	case err := <-errCh:
		log.Printf("server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	if grpcServer != nil {
		grpcServer.Stop(shutdownCtx)
	}
}

func loadConfig() (*config.Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/config.yaml"
	}
	return config.Load(path)
}

func buildStorage(cfg config.StorageConfig) (storage.Storage, func(), error) {
	switch strings.ToLower(cfg.Driver) {
	case "redis":
		if cfg.Redis.Address == "" {
			return nil, nil, errors.New("redis address is required")
		}
		store := storage.NewRedisStorage(storage.RedisConfig{
			Addr:     cfg.Redis.Address,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		return store, func() {
			_ = store.Close()
		}, nil
	default:
		return storage.NewMemoryStorage(), nil, nil
	}
}

func bootstrapHTTPServer(cfg *config.Config, manager *limiter.Manager, metrics *server.Metrics) *server.HTTPServer {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/v1/payments", jsonResponder(map[string]any{"status": "ok"}))
	apiMux.HandleFunc("/api/v1/premium/resource", jsonResponder(map[string]any{"tier": "premium"}))

	mainMux := http.NewServeMux()
	mainMux.Handle("/api/", middleware.RateLimiter(manager, metrics)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiMux.ServeHTTP(w, r)
		}),
	))

	mainMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if cfg.Metrics.Enabled {
		mainMux.Handle(cfg.Metrics.Path, metrics.Handler())
	}

	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Token bucket rate limiter running",
		})
	})

	httpCfg := server.HTTPConfig{
		Address:      cfg.Server.Address,
		ReadTimeout:  cfg.Server.ReadTimeout.Duration(),
		WriteTimeout: cfg.Server.WriteTimeout.Duration(),
		IdleTimeout:  cfg.Server.IdleTimeout.Duration(),
	}

	return server.NewHTTPServer(httpCfg, mainMux)
}

func bootstrapGRPCServer(cfg *config.Config, manager *limiter.Manager, metrics *server.Metrics) *server.GRPCServer {
	address := cfg.Server.GRPCAddress()
	if address == "" {
		return nil
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.UnaryRateLimitInterceptor(manager, metrics)),
	}
	s := grpc.NewServer(opts...)
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)
	return server.NewGRPCServer(address, s)
}

func jsonResponder(payload any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}
}
