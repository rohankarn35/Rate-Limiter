FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /rate-limiter ./cmd/server

FROM gcr.io/distroless/base-debian12
COPY --from=builder /rate-limiter /rate-limiter
COPY config/config.yaml /config/config.yaml

ENV CONFIG_PATH=/config/config.yaml
EXPOSE 8080 9093

ENTRYPOINT ["/rate-limiter"]
