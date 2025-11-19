# 🚦 Distributed Rate Limiter in Go

Version 2 turns the simple demo into a **production-ready, multi-algorithm rate limiter** that can be embedded into HTTP and gRPC services, deployed with Docker/Kubernetes, and monitored with Prometheus.

---

## ✨ Highlights

- **Config-driven policies** – declare routes, methods, identities (IP/header/query), and limits in `config/config.yaml`.
- **Multiple algorithms** – Token Bucket, Leaky Bucket, and Sliding Window backed by shared storage.
- **Per-IP / per-API key controls** – key extractors support IP fallback, arbitrary headers, or query params.
- **Pluggable storage** – in-memory engine for local testing and Redis adapter for distributed deployments.
- **HTTP & gRPC middleware** – attach the limiter manager to REST handlers or unary RPC interceptors.
- **Observability** – Prometheus counters exposed at `/metrics`, ready for scraping.
- **Batteries included ops** – Dockerfile, docker-compose stack (with Redis), and Kubernetes manifests.

---

## 📁 Project Layout

```
.
├── cmd/server               # Application entrypoint
├── config/                  # Sample configuration
├── deployments/             # Docker + Kubernetes assets
├── internal/
│   ├── api/middleware       # HTTP & gRPC middlewares
│   └── server               # HTTP/gRPC bootstrapping + metrics
├── pkg/
│   ├── config               # YAML loader & duration helpers
│   ├── limiter              # Algorithms, manager, policy builder
│   └── storage              # Memory & Redis backends
├── test/                    # Unit + benchmark suites
└── README.md
```

---

## ⚙️ Configuration

Policies live in `config/config.yaml` (see file for full example):

```yaml
server:
  address: ":8080"
  grpc_address: ":9090"
storage:
  driver: redis
  redis:
    address: "redis:6379"
policies:
  - name: public-ip
    routes: ["/api/v1/*"]
    methods: ["GET", "POST"]
    identity:
      type: ip
    algorithm:
      type: token_bucket
      limit: 30
      burst: 30
      refill_rate: 5
      interval: 1s
  - name: api-key-tier
    routes: ["/api/v1/premium/*"]
    methods: ["GET", "POST"]
    identity:
      type: header
      key: X-API-Key
      fallback: ip
    algorithm:
      type: sliding_window
      limit: 100
      window: 1m
```

Override the config path with `CONFIG_PATH=/path/to/config.yaml`.

---

## 🚀 Getting Started

```bash
# Run locally (memory storage)
go run ./cmd/server

# Run tests & benchmarks
go test ./...
go test -bench=. -benchmem ./test

# Docker (includes Redis via docker-compose)
docker compose -f deployments/docker-compose.yaml up --build
```

Once running:

- REST API demo: `curl http://localhost:8080/api/v1/payments`
- Metrics: `curl http://localhost:8080/metrics`
- gRPC health check: `grpcurl localhost:9090 grpc.health.v1.Health/Check`

---

## 📦 Deployments

- **Dockerfile** – multi-stage build producing a distroless image.
- **docker-compose** – spins up the limiter + Redis for local integration tests.
- **Kubernetes manifests** – Deployment, Service, and ConfigMap under `deployments/k8s/`.

---

## 🧪 Tests

- Algorithm unit tests verify each limiter with shared storage.
- Concurrency tests ensure thread safety under high contention.
- Benchmarks (`test/benchmark_test.go`) help gauge algorithm throughput.

---

## 📄 License

Apache License 2.0 – see `LICENSE`.

---

## 👨‍💻 Author

**Rohan Karn** — Backend & Go Developer
