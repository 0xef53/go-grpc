package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"runtime"

	"github.com/0xef53/go-grpc/server/interceptors"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var logger = log.StandardLogger().WithField("subsystem", "server")

// SetLogger sets the global logger used by the package's entities.
// It should be called during initialization, and it is strongly recommended
// not to change it afterward.
func SetLogger(entry *log.Entry) {
	logger = entry
}

// Server represents a gRPC server that handles incoming requests.
type Server struct {
	config    *Config
	tlsConfig *tls.Config

	grpcServer *grpc.Server

	buckets []string

	group *errgroup.Group
}

// NewServer creates and configures a new gRPC server instance with the provided configuration.
//
// On Linux systems, the cfg.GRPCSocketPath will be transform to abstract socket by prefixing it with '@'.
func NewServer(cfg *Config, tlsConfig *tls.Config, ui []grpc.UnaryServerInterceptor, si []grpc.StreamServerInterceptor) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		config:     cfg,
		tlsConfig:  tlsConfig,
		grpcServer: newServer(ui, si, tlsConfig),
		buckets:    []string{defaultServiceBucket},
		group:      new(errgroup.Group),
	}

	if runtime.GOOS == "linux" {
		s.config.GRPCSocketPath = "@" + s.config.GRPCSocketPath
	}

	return s, nil
}

// SetServiceBuckets sets a list of buckets used by the gRPC Gateway server to discover
// and register target services for serving.
//
// The list should not change after the server starts.
func (s *Server) SetServiceBuckets(names ...string) {
	s.buckets = names
}

// listenAndServe starts the gRPC server and serves services corresponding
// to the given list of buckets.
func (s *Server) listenAndServe(ctx context.Context) error {
	for _, svc := range Services(s.buckets...) {
		logger.Info("Registering service: ", svc.Name())

		svc.RegisterGRPC(s.grpcServer)
	}

	listeners, err := s.config.GetListeners()
	if err != nil {
		return err
	}

	// Default GRPC on Unix Socket
	if l, err := net.Listen("unix", s.config.GRPCSocketPath); err == nil {
		defer l.Close()

		listeners = append(listeners, l)
	} else {
		return err
	}

	group, groupCtx := errgroup.WithContext(ctx)

	idleConnsClosed := make(chan struct{})

	go func() {
		<-groupCtx.Done()

		s.grpcServer.GracefulStop()

		close(idleConnsClosed)
	}()

	for _, l := range listeners {
		listener := l

		group.Go(func() error {
			logger.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("Starting GRPC server")

			if err := s.grpcServer.Serve(listener); err != nil {
				// Error starting or closing listener
				return err
			}

			logger.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("GRPC server stopped")

			return nil
		})
	}

	<-idleConnsClosed

	if err := group.Wait(); err != nil {
		return fmt.Errorf("GRPC Server error: %s", err)
	}

	return nil
}

// Start starts the gRPC server in a separate goroutine and does not wait for it to complete.
//
// Use the Wait() method to wait for the server to complete and then read its exit code.
func (s *Server) Start(ctx context.Context) {
	s.group.Go(func() error { return s.listenAndServe(ctx) })
}

// Wait blocks until all goroutines started by the server have finished,
// then returns the first non-nil error (if any) from them.
func (s *Server) Wait() error {
	return s.group.Wait()
}

var DefaultUnaryInterceptors = []grpc.UnaryServerInterceptor{
	interceptors.TagsUnaryServerInterceptor(),
	interceptors.RequestIdentifierUnaryServerInterceptor(),
	grpc_logrus.UnaryServerInterceptor(logger),
}

var DefaultStreamInterceptors = []grpc.StreamServerInterceptor{
	interceptors.TagsStreamServerInterceptor(),
	interceptors.RequestIdentifierStreamServerInterceptor(),
	grpc_logrus.StreamServerInterceptor(logger),
}

// newServer returns a new grpc.Server instance with a preconfigured list of interceptors.
func newServer(ui []grpc.UnaryServerInterceptor, si []grpc.StreamServerInterceptor, tlsConfig *tls.Config) *grpc.Server {
	_ui := append(DefaultUnaryInterceptors, ui...)

	// Add after the "ui" to allow changes in "grpc_ctxtags"
	_ui = append(_ui, interceptors.LogRequestUnaryServerInterceptor())

	_si := append(DefaultStreamInterceptors, si...)

	// Add after the "si" to allow changes in "grpc_ctxtags"
	_si = append(_si, interceptors.LogRequestStreamServerInterceptor())

	opts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(_ui...),
		grpc_middleware.WithStreamServerChain(_si...),
	}

	if tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	return grpc.NewServer(opts...)
}
