package server

import (
	"context"
	"crypto/tls"
	"fmt"

	grpcgateway "github.com/0xef53/go-grpc/gateway"
	grpcserver "github.com/0xef53/go-grpc/server"

	"google.golang.org/grpc"

	"golang.org/x/sync/errgroup"
)

// Server represents a composite server that combines both the gRPC server
// and the HTTP-to-gRPC Gateway server.
type Server struct {
	grpcServer *grpcserver.Server
	gwServer   *grpcgateway.Server

	group *errgroup.Group
}

// NewServer creates and configures a new composite server instance with the provided configuration.
func NewServer(cfg *grpcserver.Config, tlsConfig *tls.Config, ui []grpc.UnaryServerInterceptor, si []grpc.StreamServerInterceptor) (*Server, error) {
	grpcServer, err := grpcserver.NewServer(cfg, tlsConfig, ui, si)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new gRPC server: %w", err)
	}

	gwServer, err := grpcgateway.NewServer(cfg, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new gRPC Gateway server: %w", err)
	}

	return &Server{
		grpcServer: grpcServer,
		gwServer:   gwServer,
		group:      new(errgroup.Group),
	}, nil
}

// SetServiceBuckets sets a list of buckets used by composite server to discover
// and register target services for serving.
//
// The list should not change after the server starts.
func (s *Server) SetServiceBuckets(names ...string) {
	s.grpcServer.SetServiceBuckets(names...)
	s.gwServer.SetServiceBuckets(names...)
}

// Start starts the composite server but does not wait for it to complete.
//
// Use the Wait() method to wait for the server to complete and then read its exit code.
func (s *Server) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	s.group.Go(func() error {
		s.grpcServer.Start(ctx)

		if err := s.grpcServer.Wait(); err != nil {
			cancel()

			return err
		}

		return nil
	})

	s.group.Go(func() error {
		s.gwServer.Start(ctx)

		if err := s.gwServer.Wait(); err != nil {
			cancel()

			return err
		}

		return nil
	})
}

// Wait blocks until the composite server has finished.
func (s *Server) Wait() error {
	return s.group.Wait()
}
