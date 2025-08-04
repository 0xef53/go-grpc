package client

import (
	"crypto/tls"

	"github.com/0xef53/go-grpc/client/interceptors"
	"github.com/0xef53/go-grpc/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	log "github.com/sirupsen/logrus"
)

var logger = log.StandardLogger().WithField("subsystem", "client")

// SetLogger sets the global logger used by the package's entities.
// It should be called during initialization, and it is strongly recommended
// not to change it afterward.
func SetLogger(entry *log.Entry) {
	logger = entry
}

// newConnection creates and configures a new gRPC client connection to the specified host:port
// according to the passed arguments.
func newConnection(hostport string, tlsConfig *tls.Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	hostport = utils.NormalizeHostport(hostport)

	dialOpts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			interceptors.WithRequestIdentifier(),
			interceptors.WithRequestLogging(logger),
		),
		grpc.WithChainStreamInterceptor(
			interceptors.WithStreamRequestIdentifier(),
			interceptors.WithStreamRequestLogging(logger),
		),
	}

	if tlsConfig == nil {
		// Insecure connection
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Secure connection
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	dialOpts = append(dialOpts, opts...)

	return grpc.Dial(hostport, dialOpts...)
}

// NewSecureConnection returns a secure gRPC client connection to the specified host:port.
//
// When configuring the connection, two mandatory unary and stream interceptors are used
// to handle the request ID and log the request parameters.
// Additional dial options can be provided using arguments.
func NewSecureConnection(hostport string, tlsConfig *tls.Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return newConnection(hostport, tlsConfig, opts...)
}

// NewInsecureConnection returns an insecure gRPC client connection to the specified host:port.
//
// When configuring the connection, two mandatory unary and stream interceptors are used
// to handle the request ID and log the request parameters.
// Additional dial options can be provided using arguments.
func NewInsecureConnection(hostport string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return newConnection(hostport, nil, opts...)
}
