package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/0xef53/go-grpc/proto/message"

	"google.golang.org/grpc"
	grpc_metadata "google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	log "github.com/sirupsen/logrus"
)

// WithRequestLogging returns an unary client interceptor that logs details about the request and response:
// start/end time, target server, full method name, request metadata, fields allowed to be displayed
// (see github.com/0xef53/go-grpc/options) and errors.
func WithRequestLogging(logger *log.Entry) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		fields := propertiesAsFields(ctx, req, nil)

		fields["server"] = cc.Target()
		fields["method"] = method

		logger.WithFields(fields).WithField("duration", time.Since(start)).Info("Invoked RPC method")

		// Call the invoker to execute RPC
		err := invoker(ctx, method, req, reply, cc, opts...)

		if err == nil {
			fields := propertiesAsFields(ctx, nil, reply)

			logger.WithFields(fields).WithField("duration", time.Since(start)).Info("Completed RPC method")
		} else {
			logger.WithError(err).WithField("duration", time.Since(start)).Error("Failed RPC method")
		}

		return err
	}
}

// WithStreamRequestLogging returns a stream client interceptor that logs details about the request and response:
// start/end time, target server, full method name, request metadata, fields allowed to be displayed
// (see github.com/0xef53/go-grpc/options) and errors.
func WithStreamRequestLogging(logger *log.Entry) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()

		fields := propertiesAsFields(ctx, nil, nil)

		fields["server"] = cc.Target()
		fields["method"] = method

		logger.WithFields(fields).WithField("duration", time.Since(start)).Info("Invoked RPC stream")

		// Call the streamer
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			logger.WithError(err).WithField("duration", time.Since(start)).Error("Failed RPC stream")

			return nil, err
		}

		logger.WithField("duration", time.Since(start)).Info("Completed RPC stream")

		return stream, nil
	}
}

func propertiesAsFields(ctx context.Context, req, reply interface{}) log.Fields {
	fields := make(log.Fields)

	if ctx != nil {
		if md, ok := grpc_metadata.FromOutgoingContext(ctx); ok {
			// Request metadata
			for k, v := range md {
				if k == "request-id" {
					k = "request.uid"
				}

				fields[k] = strings.Join(v, ":")
			}
		}
	}

	if req != nil {
		if msg, ok := req.(proto.Message); ok {
			tags := message.TagsFromMessage(msg.ProtoReflect())

			for k, v := range tags {
				fields[k] = v
			}
		}
	}

	if reply != nil {
		if msg, ok := reply.(proto.Message); ok {
			tags := message.TagsFromMessage(msg.ProtoReflect())

			for k, v := range tags {
				fields["grpc.response."+k] = v
			}
		}
	}

	return fields
}
