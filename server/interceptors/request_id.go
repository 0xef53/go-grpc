package interceptors

import (
	"context"

	"github.com/0xef53/go-grpc/utils"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	grpc_metadata "google.golang.org/grpc/metadata"
)

// RequestIdentifierUnaryServerInterceptor returns a unary server interceptor which appends a request ID
// to the context.
func RequestIdentifierUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		reqID := utils.ExtractRequestID(ctx)

		// For logging using ctxlogrus
		grpc_ctxtags.Extract(ctx).Set("request.uid", reqID)

		// Add to OutgoingMetadata so that the client interceptor can access the request ID.
		ctx = grpc_metadata.AppendToOutgoingContext(ctx, "request-id", reqID)

		return handler(ctx, req)
	}
}

type wrappedStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (s *wrappedStream) Context() context.Context {
	return s.ctx
}

// RequestIdentifierStreamServerInterceptor returns a stream server interceptor which appends a request ID
// to the context.
func RequestIdentifierStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		reqID := utils.ExtractRequestID(ctx)

		// For logging using ctxlogrus
		grpc_ctxtags.Extract(ctx).Set("request.uid", reqID)

		// Add to OutgoingMetadata so that the client interceptor can access the request ID.
		ctx = grpc_metadata.AppendToOutgoingContext(ctx, "request-id", reqID)

		return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
	}
}
