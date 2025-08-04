package interceptors

import (
	"github.com/0xef53/go-grpc/proto/message"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var extractor = func(fullMethod string, req interface{}) map[string]interface{} {
	if m, ok := req.(proto.Message); ok {
		return message.TagsFromMessage(m.ProtoReflect())
	}
	return nil
}

// TagsUnaryServerInterceptor returns custom [grpc_ctxtags.WithFieldExtractor] function which extracts
// tags of [proto.Message] according to the properties of [options.FieldLogging].
func TagsUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(extractor))
}

// TagsStreamServerInterceptor returns custom [grpc_ctxtags.WithFieldExtractor] function which extracts
// tags of [proto.Message] according to the properties of [options.FieldLogging].
func TagsStreamServerInterceptor() grpc.StreamServerInterceptor {
	return grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(extractor))
}
