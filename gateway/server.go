package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/0xef53/go-grpc/client/interceptors"
	"github.com/0xef53/go-grpc/gateway/utils"
	grpcserver "github.com/0xef53/go-grpc/server"

	grpc_runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var logger = log.StandardLogger().WithField("subsystem", "gateway")

// SetLogger sets the global logger used by the package's entities.
// It should be called during initialization, and it is strongly recommended
// not to change it afterward.
func SetLogger(entry *log.Entry) {
	logger = entry
}

// Server represents a gRPC-Gateway server that bridges gRPC services with HTTP/REST clients.
type Server struct {
	config    *grpcserver.Config
	tlsConfig *tls.Config

	httpServer *http.Server
	mux        *grpc_runtime.ServeMux
	dialOpts   []grpc.DialOption

	buckets []string

	group *errgroup.Group
}

// NewServer creates and configures a new gRPC Gateway server.
//
// Communication with the gRPC server takes place via a unix socket
// (GRPCSocketPath field in grpcserver.Config structure).
//
// When configuring the connection, аn unary client logging interceptor are used
// (see ... for details).
func NewServer(cfg *grpcserver.Config, tlsConfig *tls.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		config:     cfg,
		tlsConfig:  tlsConfig,
		httpServer: new(http.Server),
		mux:        utils.NewGatewayMux(),
		dialOpts:   make([]grpc.DialOption, 0, 2),
		group:      new(errgroup.Group),
	}

	if tlsConfig == nil {
		s.dialOpts = append(s.dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		s.dialOpts = append(s.dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	s.dialOpts = append(s.dialOpts, grpc.WithChainUnaryInterceptor(interceptors.WithRequestLogging(logger)))

	s.SetHTTPHandler(func(m *grpc_runtime.ServeMux) http.Handler {
		mux := http.NewServeMux()

		mux.Handle("/", m)

		return mux
	})

	return s, nil
}

// SetServiceBuckets sets a list of buckets used by the gRPC Gateway server to discover
// and register target services for serving.
//
// The list should not change after the server starts.
func (s *Server) SetServiceBuckets(names ...string) {
	s.buckets = names
}

// SetHTTPHandler allows replacing the default HTTP handler with a custom one.
//
// The handler should not change after the server starts.
func (s *Server) SetHTTPHandler(fn func(m *grpc_runtime.ServeMux) http.Handler) {
	s.httpServer.Handler = fn(s.mux)
}

// listenAndServe starts the gRPC Gateway server and serves services corresponding
// to the given list of buckets.
func (s *Server) listenAndServe(ctx context.Context) error {
	for _, svc := range grpcserver.Services(s.buckets...) {
		logger.Info("Registering GW service: ", svc.Name())

		svc.RegisterGW(s.mux, fmt.Sprintf("unix:%s", s.config.GRPCSocketPath), s.dialOpts)
	}

	listeners, err := s.config.GetGatewayListeners()
	if err != nil {
		return err
	}

	group, groupCtx := errgroup.WithContext(ctx)

	idleConnsClosed := make(chan struct{})

	go func() {
		<-groupCtx.Done()

		/*
			TODO:
				Probably use a "timeout" context here
		*/
		s.httpServer.Shutdown(context.Background())

		close(idleConnsClosed)
	}()

	for _, l := range listeners {
		listener := l

		group.Go(func() error {
			logger.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("Starting GRPC Gateway server")

			if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				// Error starting or closing listener
				return err
			}

			logger.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("GRPC Gateway server stopped")

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("GRPC Gateway server error: %s", err)
	}

	return nil
}

// Start starts the gRPC Gateway server in a separate goroutine and does not wait for it to complete.
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
