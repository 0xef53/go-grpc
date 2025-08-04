package interceptors

import (
	"context"

	"github.com/0xef53/go-grpc/utils"

	"google.golang.org/grpc"
	grpc_metadata "google.golang.org/grpc/metadata"
)

// WithRequestIdentifier returns an unary client interceptor that appends an unique ID
// to the context for the outgoing request.
// If "request-id" is already present in the metadata, its value will be supplemented with a random string.
func WithRequestIdentifier() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = withRequestID(ctx)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// WithStreamRequestIdentifier returns a stream client interceptor that appends an unique ID
// to the context for the outgoing request.
// If "request-id" is already present in the metadata, its value will be supplemented with a random string.
func WithStreamRequestIdentifier() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = withRequestID(ctx)

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func withRequestID(ctx context.Context) context.Context {
	/*
			TODO:
				Perhaps it would be more correct to use this code:

		return metadata.AppendToOutgoingContext(ctx, "request-id", utils.NewRequestID())
	*/

	reqID := utils.NewRequestID()

	// When creating a new outgoing ID, we rely on the fact that the original incoming request ID
	// was added to OutgoingMetadata by the server interceptor:
	// https://gitlab.netangels.ru/golang/grpc/-/blob/master/server/interceptors/request_id.go
	if md, ok := grpc_metadata.FromOutgoingContext(ctx); ok {
		if v, ok := md["request-id"]; ok {
			reqID = v[0] + ":" + reqID
		}

		md.Set("request-id", reqID)

		ctx = grpc_metadata.NewOutgoingContext(ctx, md)
	} else {
		ctx = grpc_metadata.AppendToOutgoingContext(ctx, "request-id", reqID)
	}

	return ctx
}
