package middleware

import (
	"context"
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/rohankarn35/rate_limiter_golang/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// UnaryRateLimitInterceptor applies rate limiting to unary RPCs.
func UnaryRateLimitInterceptor(manager *limiter.Manager, recorder MetricsRecorder) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if manager == nil {
			return handler(ctx, req)
		}

		httpReq := grpcRequestToHTTP(ctx, info.FullMethod)
		result, policy, matched, err := manager.Allow(ctx, httpReq)
		if err != nil {
			return nil, status.Error(codes.Internal, "rate limiter failure")
		}
		if matched && recorder != nil {
			recorder.Observe(policy, result.Allowed)
		}
		if matched && !result.Allowed {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func grpcRequestToHTTP(ctx context.Context, method string) *http.Request {
	reqURL := &url.URL{Path: method}
	httpReq := &http.Request{
		Method:     http.MethodPost,
		URL:        reqURL,
		Header:     http.Header{},
		RemoteAddr: "",
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key, values := range md {
			headerKey := textproto.CanonicalMIMEHeaderKey(key)
			for _, value := range values {
				httpReq.Header.Add(headerKey, value)
			}
		}
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		httpReq.RemoteAddr = p.Addr.String()
	}

	return httpReq
}
