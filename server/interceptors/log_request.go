package interceptors

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"google.golang.org/grpc"
)

// LogRequestUnaryServerInterceptor returns a unary server interceptor that logs details of gRPC request.
func LogRequestUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctxlogrus.Extract(ctx).WithContext(ctx).Infof("GRPC Request: %s", info.FullMethod)

		return handler(ctx, req)
	}
}

// LogRequestStreamServerInterceptor returns a stream server interceptor that logs details of gRPC request.
func LogRequestStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctxlogrus.Extract(ss.Context()).WithContext(ss.Context()).Infof("GRPC Request: %s", info.FullMethod)

		return handler(srv, ss)
	}
}
