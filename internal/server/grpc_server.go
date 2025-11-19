package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// GRPCServer wraps grpc.Server to coordinate lifecycle.
type GRPCServer struct {
	server  *grpc.Server
	address string
	lis     net.Listener
}

// NewGRPCServer builds a new gRPC server bound to address.
func NewGRPCServer(address string, server *grpc.Server) *GRPCServer {
	return &GRPCServer{
		server:  server,
		address: address,
	}
}

// Start begins serving on the configured address.
func (g *GRPCServer) Start() error {
	if g.server == nil || g.address == "" {
		return fmt.Errorf("grpc server misconfigured")
	}
	lis, err := net.Listen("tcp", g.address)
	if err != nil {
		return err
	}
	g.lis = lis
	return g.server.Serve(lis)
}

// Stop gracefully stops the server.
func (g *GRPCServer) Stop(ctx context.Context) {
	if g.server == nil {
		return
	}
	done := make(chan struct{})
	go func() {
		g.server.GracefulStop()
		close(done)
	}()
	select {
	case <-ctx.Done():
		g.server.Stop()
	case <-done:
	}
	if g.lis != nil {
		_ = g.lis.Close()
	}
}
